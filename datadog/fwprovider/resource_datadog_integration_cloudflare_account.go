package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationCloudflareAccountResource{}
	_ resource.ResourceWithImportState = &integrationCloudflareAccountResource{}
)

type integrationCloudflareAccountResource struct {
	Api  *datadogV2.CloudflareIntegrationApi
	Auth context.Context
}

type integrationCloudflareAccountModel struct {
	ID              types.String `tfsdk:"id"`
	ApiKey          types.String `tfsdk:"api_key"`
	ApiKeyWo        types.String `tfsdk:"api_key_wo"`
	ApiKeyWoVersion types.String `tfsdk:"api_key_wo_version"`
	Email           types.String `tfsdk:"email"`
	Name            types.String `tfsdk:"name"`
	Resources       types.Set    `tfsdk:"resources"`
}

// Write-only secret configuration for Cloudflare API key
var cloudflareApiKeyConfig = utils.WriteOnlySecretConfig{
	OriginalAttr:         "api_key",
	WriteOnlyAttr:        "api_key_wo",
	TriggerAttr:          "api_key_wo_version",
	OriginalDescription:  "The API key (or token) for the Cloudflare account.",
	WriteOnlyDescription: "Write-only API key (or token) for the Cloudflare account.",
	TriggerDescription:   "Version associated with api_key_wo. Changing this triggers an update. Can be any string (e.g., '1', 'v2.1', '2024-Q1').",
}

func NewIntegrationCloudflareAccountResource() resource.Resource {
	return &integrationCloudflareAccountResource{}
}

func (r *integrationCloudflareAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudflareIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationCloudflareAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_cloudflare_account"
}

func (r *integrationCloudflareAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	// Generate write-only secret attributes using helper
	writeOnlyAttrs := utils.CreateWriteOnlySecretAttributes(cloudflareApiKeyConfig)

	// Combine with other resource-specific attributes
	allAttributes := map[string]schema.Attribute{
		"email": schema.StringAttribute{
			Optional:    true,
			Description: "The email associated with the Cloudflare account. If an API key is provided (and not a token), this field is also required.",
		},
		"name": schema.StringAttribute{
			Required:    true,
			Description: "The name of the Cloudflare account.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"id": utils.ResourceIDAttribute(),
		"resources": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Computed:    true,
			Description: "An allowlist of resources to pull metrics for. Includes `web`, `dns`, `lb` (load balancer), and `worker`).",
		},
	}

	// Merge write-only attributes with resource-specific ones
	for key, attr := range writeOnlyAttrs {
		allAttributes[key] = attr
	}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationCloudflareAccount resource. This can be used to create and manage Datadog integration_cloudflare_account.",
		Attributes:  allAttributes,
	}
}

func (r *integrationCloudflareAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationCloudflareAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetCloudflareAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving API Key"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationCloudflareAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildIntegrationCloudflareAccountRequestBody(ctx, &state, &request.Config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateCloudflareAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationCloudflareAccount"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationCloudflareAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan integrationCloudflareAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var prior integrationCloudflareAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &prior)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	body, diags := r.buildIntegrationCloudflareAccountUpdateRequestBody(ctx, &plan, &prior, &request.Config, &request)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateCloudflareAccount(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationCloudflareAccount"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &plan, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *integrationCloudflareAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteCloudflareAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_cloudflare_account"))
		return
	}
}

func (r *integrationCloudflareAccountResource) updateState(ctx context.Context, state *integrationCloudflareAccountModel, resp *datadogV2.CloudflareAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if email, ok := attributes.GetEmailOk(); ok && *email != "" {
		state.Email = types.StringValue(*email)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if resources, ok := attributes.GetResourcesOk(); ok {
		state.Resources, _ = types.SetValueFrom(ctx, types.StringType, resources)
	}
}

func (r *integrationCloudflareAccountResource) buildIntegrationCloudflareAccountRequestBody(ctx context.Context, state *integrationCloudflareAccountModel, config *tfsdk.Config) (*datadogV2.CloudflareAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCloudflareAccountCreateRequestAttributesWithDefaults()

	// Use helper to get secret for creation
	handler := utils.WriteOnlySecretHandler{Config: cloudflareApiKeyConfig}
	secret, useWriteOnly, secretDiags := handler.GetSecretForCreate(ctx, state, config)
	diags.Append(secretDiags...)

	if useWriteOnly {
		attributes.SetApiKey(secret)
	} else if !state.ApiKey.IsNull() && !state.ApiKey.IsUnknown() {
		attributes.SetApiKey(state.ApiKey.ValueString())
	}
	if !state.Email.IsNull() {
		attributes.SetEmail(state.Email.ValueString())
	}
	attributes.SetName(state.Name.ValueString())

	if !state.Resources.IsNull() && !state.Resources.IsUnknown() {
		var resources []string
		diags.Append(state.Resources.ElementsAs(ctx, &resources, false)...)
		attributes.SetResources(resources)
	}

	req := datadogV2.NewCloudflareAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewCloudflareAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *integrationCloudflareAccountResource) buildIntegrationCloudflareAccountUpdateRequestBody(ctx context.Context, plan *integrationCloudflareAccountModel, prior *integrationCloudflareAccountModel, config *tfsdk.Config, request *resource.UpdateRequest) (*datadogV2.CloudflareAccountUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCloudflareAccountUpdateRequestAttributesWithDefaults()

	// Use helper to determine if secret should be updated
	handler := utils.WriteOnlySecretHandler{Config: cloudflareApiKeyConfig}
	secret, shouldUpdate, secretDiags := handler.GetSecretForUpdate(ctx, config, request)
	diags.Append(secretDiags...)

	if shouldUpdate {
		attributes.SetApiKey(secret)
	} else if !plan.ApiKey.IsNull() && !plan.ApiKey.IsUnknown() {
		// Plaintext mode: always update
		attributes.SetApiKey(plan.ApiKey.ValueString())
	}

	if !plan.Email.IsNull() {
		attributes.SetEmail(plan.Email.ValueString())
	}

	if !plan.Resources.IsNull() && !plan.Resources.IsUnknown() {
		var resources []string
		diags.Append(plan.Resources.ElementsAs(ctx, &resources, false)...)
		attributes.SetResources(resources)
	}

	req := datadogV2.NewCloudflareAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewCloudflareAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
