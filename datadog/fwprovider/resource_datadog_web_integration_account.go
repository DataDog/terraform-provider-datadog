package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &webIntegrationAccountResource{}
	_ resource.ResourceWithImportState = &webIntegrationAccountResource{}
)

// API request/response types for Web Integrations (matches spec/v2/web_integrations.yaml)
type webIntegrationAccountCreateAttributes struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
	Secrets  map[string]interface{} `json:"secrets"`
}

type webIntegrationAccountUpdateAttributes struct {
	Name     string                 `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Secrets  map[string]interface{} `json:"secrets,omitempty"`
}

type webIntegrationAccountCreateRequest struct {
	Data struct {
		Type       string                            `json:"type"`
		Attributes webIntegrationAccountCreateAttributes `json:"attributes"`
	} `json:"data"`
}

type webIntegrationAccountUpdateRequest struct {
	Data struct {
		Type       string                            `json:"type"`
		Attributes webIntegrationAccountUpdateAttributes `json:"attributes"`
	} `json:"data"`
}

type webIntegrationAccountAttributes struct {
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
	// Secrets are never returned by the API (write-only)
}

type webIntegrationAccountResponse struct {
	IntegrationName string `json:"integration_name"`
	Data            struct {
		ID         string                        `json:"id"`
		Type       string                        `json:"type"`
		Attributes webIntegrationAccountAttributes `json:"attributes"`
	} `json:"data"`
}

type webIntegrationAccountResource struct {
	Client *datadog.APIClient
	Auth   context.Context
}

type webIntegrationAccountModel struct {
	ID              types.String       `tfsdk:"id"`
	IntegrationName types.String       `tfsdk:"integration_name"`
	Name            types.String       `tfsdk:"name"`
	SettingsJson    jsontypes.Normalized `tfsdk:"settings_json"`
	SecretsJson     jsontypes.Normalized `tfsdk:"secrets_json"`
}

func NewWebIntegrationAccountResource() resource.Resource {
	return &webIntegrationAccountResource{}
}

func (r *webIntegrationAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Client = providerData.DatadogApiInstances.HttpClient
	r.Auth = providerData.Auth
}

func (r *webIntegrationAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "web_integration_account"
}

func (r *webIntegrationAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Web Integration Account resource. This can be used to create and manage accounts for third-party web integrations (e.g., Twilio, Snowflake, Databricks). The account name must be unique within each integration.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"integration_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the integration (e.g., twilio, snowflake-web, databricks). Changing this forces recreation of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the account. Must be unique among accounts of the same integration. Used for display and drift detection.",
			},
			"settings_json": schema.StringAttribute{
				Required:    true,
				Description: "Integration-specific settings as JSON. Structure varies by integration; use GET /api/v2/web-integrations/{integration_name}/accounts/schema to retrieve the schema.",
				CustomType:  jsontypes.NormalizedType{},
			},
			"secrets_json": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Sensitive credentials as JSON. Structure varies by integration. Values are write-only and never returned; changes outside Terraform will not be drift-detected.",
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *webIntegrationAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.SplitN(request.ID, ":", 2)
	if len(parts) != 2 {
		response.Diagnostics.AddError(
			"Invalid import ID",
			`Expected ID in format "integration_name:account_id". Example: terraform import datadog_web_integration_account.example "twilio:abc123def456"`,
		)
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("integration_name"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("id"), parts[1])...)
}

func (r *webIntegrationAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state webIntegrationAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/v2/web-integrations/%s/accounts/%s",
		state.IntegrationName.ValueString(),
		state.ID.ValueString())

	body, httpResp, err := utils.SendRequest(r.Auth, r.Client, "GET", path, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Web Integration Account"))
		if apiErr, ok := err.(utils.CustomRequestAPIError); ok && len(apiErr.Body()) > 0 {
			response.Diagnostics.AddError("API response", string(apiErr.Body()))
		}
		return
	}

	var apiResp webIntegrationAccountResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Diagnostics.AddError("error parsing API response", err.Error())
		return
	}

	state.ID = types.StringValue(apiResp.Data.ID)
	state.IntegrationName = types.StringValue(apiResp.IntegrationName)
	state.Name = types.StringValue(apiResp.Data.Attributes.Name)

	settingsBytes, _ := json.Marshal(apiResp.Data.Attributes.Settings)
	state.SettingsJson = jsontypes.NewNormalizedValue(string(settingsBytes))

	// Secrets are never returned by the API; already in state, no need to overwrite

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *webIntegrationAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan webIntegrationAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var settings, secrets map[string]interface{}
	if err := json.Unmarshal([]byte(plan.SettingsJson.ValueString()), &settings); err != nil {
		response.Diagnostics.AddError("invalid settings_json", err.Error())
		return
	}
	if err := json.Unmarshal([]byte(plan.SecretsJson.ValueString()), &secrets); err != nil {
		response.Diagnostics.AddError("invalid secrets_json", err.Error())
		return
	}

	reqBody := webIntegrationAccountCreateRequest{}
	reqBody.Data.Type = "Account"
	reqBody.Data.Attributes = webIntegrationAccountCreateAttributes{
		Name:     plan.Name.ValueString(),
		Settings: settings,
		Secrets:  secrets,
	}

	path := fmt.Sprintf("/api/v2/web-integrations/%s/accounts", plan.IntegrationName.ValueString())
	body, httpResp, err := utils.SendRequest(r.Auth, r.Client, "POST", path, reqBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Web Integration Account"))
		if apiErr, ok := err.(utils.CustomRequestAPIError); ok && len(apiErr.Body()) > 0 {
			response.Diagnostics.AddError("API response", string(apiErr.Body()))
		}
		return
	}

	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError("unexpected status", fmt.Sprintf("expected 201, got %d", httpResp.StatusCode))
		return
	}

	var apiResp webIntegrationAccountResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Diagnostics.AddError("error parsing API response", err.Error())
		return
	}

	plan.ID = types.StringValue(apiResp.Data.ID)
	plan.IntegrationName = types.StringValue(apiResp.IntegrationName)
	plan.Name = types.StringValue(apiResp.Data.Attributes.Name)
	settingsBytes, _ := json.Marshal(apiResp.Data.Attributes.Settings)
	plan.SettingsJson = jsontypes.NewNormalizedValue(string(settingsBytes))
	// Secrets: keep from plan (never returned)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *webIntegrationAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, priorState webIntegrationAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &priorState)...)
	if response.Diagnostics.HasError() {
		return
	}

	attrs := webIntegrationAccountUpdateAttributes{}
	attrs.Name = plan.Name.ValueString()

	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(plan.SettingsJson.ValueString()), &settings); err != nil {
		response.Diagnostics.AddError("invalid settings_json", err.Error())
		return
	}
	attrs.Settings = settings

	// Preserve secrets from prior state if plan has UseStateForUnknown (e.g. no change)
	secretsJson := plan.SecretsJson.ValueString()
	if secretsJson == "" || plan.SecretsJson.IsNull() {
		secretsJson = priorState.SecretsJson.ValueString()
	}
	var secrets map[string]interface{}
	if err := json.Unmarshal([]byte(secretsJson), &secrets); err != nil {
		response.Diagnostics.AddError("invalid secrets_json", err.Error())
		return
	}
	attrs.Secrets = secrets

	reqBody := webIntegrationAccountUpdateRequest{}
	reqBody.Data.Type = "Account"
	reqBody.Data.Attributes = attrs

	path := fmt.Sprintf("/api/v2/web-integrations/%s/accounts/%s",
		plan.IntegrationName.ValueString(),
		plan.ID.ValueString())

	body, _, err := utils.SendRequest(r.Auth, r.Client, "PATCH", path, reqBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Web Integration Account"))
		if apiErr, ok := err.(utils.CustomRequestAPIError); ok && len(apiErr.Body()) > 0 {
			response.Diagnostics.AddError("API response", string(apiErr.Body()))
		}
		return
	}

	var apiResp webIntegrationAccountResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		response.Diagnostics.AddError("error parsing API response", err.Error())
		return
	}

	plan.Name = types.StringValue(apiResp.Data.Attributes.Name)
	settingsBytes, _ := json.Marshal(apiResp.Data.Attributes.Settings)
	plan.SettingsJson = jsontypes.NewNormalizedValue(string(settingsBytes))
	plan.SecretsJson = priorState.SecretsJson // preserve; never returned

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *webIntegrationAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state webIntegrationAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/v2/web-integrations/%s/accounts/%s",
		state.IntegrationName.ValueString(),
		state.ID.ValueString())

	_, httpResp, err := utils.SendRequest(r.Auth, r.Client, "DELETE", path, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return // already deleted
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting Web Integration Account"))
		if apiErr, ok := err.(utils.CustomRequestAPIError); ok && len(apiErr.Body()) > 0 {
			response.Diagnostics.AddError("API response", string(apiErr.Body()))
		}
		return
	}
}
