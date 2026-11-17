package fwprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &statusPageEmailAccessPolicyResource{}
	_ resource.ResourceWithImportState = &statusPageEmailAccessPolicyResource{}
)

type statusPageEmailAccessPolicyResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

type statusPageEmailAccessPolicyModel struct {
	ID        types.String `tfsdk:"id"`
	PageID    types.String `tfsdk:"page_id"`
	Pattern   types.String `tfsdk:"pattern"`
	Effect    types.String `tfsdk:"effect"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewStatusPageEmailAccessPolicyResource() resource.Resource {
	return &statusPageEmailAccessPolicyResource{}
}

func (r *statusPageEmailAccessPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "status_page_email_access_policy"
}

func (r *statusPageEmailAccessPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// The API has no update operation for email access policies, so every
	// configurable attribute requires replacement.
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog status page email access policy resource. This can be used to control which email addresses can access a private status page. Deny policies take precedence over allow policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the email access policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the private status page this email access policy belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pattern": schema.StringAttribute{
				Description: "The email address pattern matched by the policy. Supports an exact address (`user@example.com`), all addresses in a domain (`@example.com`), or all addresses in subdomains (`@*.example.com`).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"effect": schema.StringAttribute{
				Description: "The effect applied when an email address matches the policy. Deny policies take precedence over allow policies.",
				Required:    true,
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewAccessPolicyEffectFromValue),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the email access policy was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *statusPageEmailAccessPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	r.Auth = providerData.Auth
}

func (r *statusPageEmailAccessPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan statusPageEmailAccessPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(plan.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	body := datadogV2.CreateStatusPageEmailAccessPolicyRequest{
		Data: datadogV2.CreateStatusPageEmailAccessPolicyRequestData{
			Type: datadogV2.STATUSPAGEEMAILACCESSPOLICYTYPE_EMAIL_ACCESS_POLICIES,
			Attributes: datadogV2.CreateStatusPageEmailAccessPolicyRequestDataAttributes{
				Effect:  datadogV2.AccessPolicyEffect(plan.Effect.ValueString()),
				Pattern: plan.Pattern.ValueString(),
			},
		},
	}

	resp, httpResp, err := r.Api.CreateEmailAccessPolicy(r.Auth, pageID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating email access policy",
			fmt.Sprintf("Could not create email access policy, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating email access policy",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state statusPageEmailAccessPolicyModel
	state.PageID = plan.PageID
	data := resp.GetData()
	r.updateStateFromData(&state, &data)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageEmailAccessPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageEmailAccessPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	policyID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse email access policy ID: "+err.Error(),
		)
		return
	}

	// The API has no endpoint to get a single email access policy, so look it
	// up in the status page's list of policies.
	resp, httpResp, err := r.Api.ListEmailAccessPolicies(r.Auth, pageID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading email access policy",
			"Could not read email access policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	for _, policy := range resp.GetData() {
		if policy.GetId() == policyID {
			r.updateStateFromData(&state, &policy)
			response.Diagnostics.Append(response.State.Set(ctx, &state)...)
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (r *statusPageEmailAccessPolicyResource) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError(
		"Update Not Supported",
		"Status page email access policies do not support updates. Changes require resource recreation.",
	)
}

func (r *statusPageEmailAccessPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageEmailAccessPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	policyID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse email access policy ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteEmailAccessPolicy(r.Auth, pageID, policyID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting email access policy",
			"Could not delete email access policy ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *statusPageEmailAccessPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("unexpected import format", "expected '<page_id>:<policy_id>'")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("page_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
}

func (r *statusPageEmailAccessPolicyResource) updateStateFromData(state *statusPageEmailAccessPolicyModel, data *datadogV2.StatusPageEmailAccessPolicyData) {
	state.ID = types.StringValue(data.GetId().String())

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if statusPage, ok := relationships.GetStatusPageOk(); ok && statusPage != nil {
			if pageData, ok := statusPage.GetDataOk(); ok && pageData != nil && pageData.GetId() != uuid.Nil {
				state.PageID = types.StringValue(pageData.GetId().String())
			}
		}
	}

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		state.Pattern = types.StringValue(attributes.GetPattern())
		state.Effect = types.StringValue(string(attributes.GetEffect()))
		state.CreatedAt = types.StringValue(attributes.GetCreatedAt().UTC().Format(time.RFC3339))
	}
}
