package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &webhookCustomVariableResource{}
	_ resource.ResourceWithImportState = &webhookCustomVariableResource{}
)

type webhookCustomVariableResource struct {
	Api  *datadogV1.WebhooksIntegrationApi
	Auth context.Context
}

type webhookCustomVariableModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Value    types.String `tfsdk:"value"`
	IsSecret types.Bool   `tfsdk:"is_secret"`
}

func NewWebhookCustomVariableResource() resource.Resource {
	return &webhookCustomVariableResource{}
}

func (r *webhookCustomVariableResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetWebhooksIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *webhookCustomVariableResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "webhook_custom_variable"
}

func (r *webhookCustomVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog webhooks custom variable resource. This can be used to create and manage Datadog webhooks custom variables.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the variable. It corresponds with `<CUSTOM_VARIABLE_NAME>`.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the custom variable.",
				Required:    true,
				Sensitive:   true,
			},
			"is_secret": schema.BoolAttribute{
				Description: "Whether the custom variable is secret or not.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
		},
	}
}

func (r *webhookCustomVariableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *webhookCustomVariableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state webhookCustomVariableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Name.ValueString()
	resp, httpResp, err := r.Api.GetWebhooksIntegrationCustomVariable(r.Auth, id)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting webhooks custom variable"))
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

func (r *webhookCustomVariableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state webhookCustomVariableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildWebhookCustomVariableRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	webhookMutex.Lock()
	defer webhookMutex.Unlock()
	resp, _, err := r.Api.CreateWebhooksIntegrationCustomVariable(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating webhooks custom variable"))
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

func (r *webhookCustomVariableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state webhookCustomVariableModel
	var prev_state webhookCustomVariableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	response.Diagnostics.Append(request.State.Get(ctx, &prev_state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := prev_state.Name.ValueString()

	body, diags := r.buildWebhookCustomVariableUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	webhookMutex.Lock()
	defer webhookMutex.Unlock()
	resp, _, err := r.Api.UpdateWebhooksIntegrationCustomVariable(r.Auth, id, *body)

	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating webhook"))
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

func (r *webhookCustomVariableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state webhookCustomVariableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Name.ValueString()

	webhookMutex.Lock()
	defer webhookMutex.Unlock()
	httpResp, err := r.Api.DeleteWebhooksIntegrationCustomVariable(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting webhooks custom variable"))
		return
	}
}

func (r *webhookCustomVariableResource) updateState(ctx context.Context, state *webhookCustomVariableModel, resp *datadogV1.WebhooksIntegrationCustomVariableResponse) {
	state.ID = types.StringValue(resp.GetName())
	if name, ok := resp.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if value, ok := resp.GetValueOk(); ok {
		state.Value = types.StringValue(*value)
	}

	if is_secret, ok := resp.GetIsSecretOk(); ok {
		state.IsSecret = types.BoolValue(*is_secret)
	}

}

func (r *webhookCustomVariableResource) buildWebhookCustomVariableRequestBody(ctx context.Context, state *webhookCustomVariableModel) (*datadogV1.WebhooksIntegrationCustomVariable, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV1.WebhooksIntegrationCustomVariable{}

	attributes.SetName(state.Name.ValueString())
	attributes.SetValue(state.Value.ValueString())
	attributes.SetIsSecret(state.IsSecret.ValueBool())

	return &attributes, diags
}

func (r *webhookCustomVariableResource) buildWebhookCustomVariableUpdateRequestBody(ctx context.Context, state *webhookCustomVariableModel) (*datadogV1.WebhooksIntegrationCustomVariableUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV1.WebhooksIntegrationCustomVariableUpdateRequest{}

	attributes.SetName(state.Name.ValueString())
	attributes.SetValue(state.Value.ValueString())
	attributes.SetIsSecret(state.IsSecret.ValueBool())

	return &attributes, diags
}
