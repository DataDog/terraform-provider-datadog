package fwprovider

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	webhookMutex sync.Mutex                       = sync.Mutex{}
	_            resource.ResourceWithConfigure   = &webhookResource{}
	_            resource.ResourceWithImportState = &webhookResource{}
)

type webhookResource struct {
	Api  *datadogV1.WebhooksIntegrationApi
	Auth context.Context
}

type webhookModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	URL           types.String `tfsdk:"url"`
	Payload       types.String `tfsdk:"payload"`
	CustomHeaders types.String `tfsdk:"custom_headers"`
	EncodeAs      types.String `tfsdk:"encode_as"`
}

func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

func (r *webhookResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetWebhooksIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *webhookResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog webhook resource. This can be used to create and manage Datadog webhooks.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the webhook. It corresponds with `<WEBHOOK_NAME>`.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The URL of the webhook.",
				Required:    true,
			},
			"payload": schema.StringAttribute{
				Description: "The payload of the webhook.",
				Optional:    true,
				Computed:    true,
			},
			"custom_headers": schema.StringAttribute{
				Description: "The headers attached to the webhook.",
				Optional:    true,
			},
			"encode_as": schema.StringAttribute{
				Description: "Encoding type.",
				Optional:    true,
				Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV1.NewWebhooksIntegrationEncodingFromValue)},
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
		},
	}
}
func (r *webhookResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *webhookResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state webhookModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Name.ValueString()
	// handle import case
	if state.Name.IsNull() {
		id = state.ID.ValueString()
	}

	resp, httpResp, err := r.Api.GetWebhooksIntegration(r.Auth, id)

	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting webhook"))
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

func (r *webhookResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state webhookModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildWebhookRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	webhookMutex.Lock()
	defer webhookMutex.Unlock()
	resp, _, err := r.Api.CreateWebhooksIntegration(r.Auth, *body)
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

func (r *webhookResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state webhookModel
	var prev_state webhookModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	response.Diagnostics.Append(request.State.Get(ctx, &prev_state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := prev_state.Name.ValueString()

	body, diags := r.buildWebhookUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	webhookMutex.Lock()
	defer webhookMutex.Unlock()
	resp, _, err := r.Api.UpdateWebhooksIntegration(r.Auth, id, *body)

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

func (r *webhookResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state webhookModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Name.ValueString()

	webhookMutex.Lock()
	defer webhookMutex.Unlock()
	httpResp, err := r.Api.DeleteWebhooksIntegration(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting webhook"))
		return
	}
}

func (r *webhookResource) updateState(ctx context.Context, state *webhookModel, resp *datadogV1.WebhooksIntegration) {
	state.ID = types.StringValue(resp.GetName())
	if name, ok := resp.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if url, ok := resp.GetUrlOk(); ok {
		state.URL = types.StringValue(*url)
	}

	if payload, ok := resp.GetPayloadOk(); ok && payload != nil {
		state.Payload = types.StringValue(*payload)
	}

	if customHeaders, ok := resp.GetCustomHeadersOk(); ok && customHeaders != nil {
		state.CustomHeaders = types.StringValue(*customHeaders)
	}

	if encode_as, ok := resp.GetEncodeAsOk(); ok && encode_as != nil {
		state.EncodeAs = types.StringValue(string(*encode_as))
	}

}

func (r *webhookResource) buildWebhookRequestBody(ctx context.Context, state *webhookModel) (*datadogV1.WebhooksIntegration, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV1.WebhooksIntegration{}

	attributes.SetName(state.Name.ValueString())
	attributes.SetUrl(state.URL.ValueString())

	if !state.Payload.IsNull() && state.Payload.ValueString() != "" {
		attributes.SetPayload(state.Payload.ValueString())
	}
	if !state.CustomHeaders.IsNull() {
		attributes.SetCustomHeaders(state.CustomHeaders.ValueString())
	}
	if !state.EncodeAs.IsNull() && !state.EncodeAs.IsUnknown() {
		encoding, _ := datadogV1.NewWebhooksIntegrationEncodingFromValue(state.EncodeAs.ValueString())
		attributes.SetEncodeAs(*encoding)
	}

	return &attributes, diags
}

func (r *webhookResource) buildWebhookUpdateRequestBody(ctx context.Context, state *webhookModel) (*datadogV1.WebhooksIntegrationUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV1.WebhooksIntegrationUpdateRequest{}

	attributes.SetName(state.Name.ValueString())
	attributes.SetUrl(state.URL.ValueString())

	if !state.Payload.IsNull() && state.Payload.ValueString() != "" {
		attributes.SetPayload(state.Payload.ValueString())
	}
	if !state.CustomHeaders.IsNull() {
		attributes.SetCustomHeaders(state.CustomHeaders.ValueString())
	}

	if !state.EncodeAs.IsNull() {
		encoding, _ := datadogV1.NewWebhooksIntegrationEncodingFromValue(state.EncodeAs.ValueString())
		attributes.SetEncodeAs(*encoding)
	}
	return &attributes, diags
}
