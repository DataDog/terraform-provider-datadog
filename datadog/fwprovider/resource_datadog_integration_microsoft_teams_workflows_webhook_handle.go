package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const maskedSecret = "*****"

var (
	_ resource.ResourceWithConfigure   = &workflowsWebhookHandleResource{}
	_ resource.ResourceWithImportState = &workflowsWebhookHandleResource{}
)

type workflowsWebhookHandleResource struct {
	Api  *datadogV2.MicrosoftTeamsIntegrationApi
	Auth context.Context
}

type workflowsWebhookHandleModel struct {
	ID   types.String `tfsdk:"id"`
	URL  types.String `tfsdk:"url"`
	Name types.String `tfsdk:"name"`
}

func NewWorkflowsWebhookHandleResource() resource.Resource {
	return &workflowsWebhookHandleResource{}
}

func (r *workflowsWebhookHandleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMicrosoftTeamsIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *workflowsWebhookHandleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_ms_teams_workflows_webhook_handle"
}

func (r *workflowsWebhookHandleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Resource for interacting with Datadog Microsoft Teams integration Microsoft Workflows webhook handles.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Your Microsoft Workflows webhook handle name.",
			},
			"url": schema.StringAttribute{
				Description: "Your Microsoft Workflows webhook URL.",
				Required:    true,
				Sensitive:   true,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *workflowsWebhookHandleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *workflowsWebhookHandleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state workflowsWebhookHandleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	// Check if handle exists
	resp, httpResp, err := r.Api.GetWorkflowsWebhookHandle(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Microsoft Workflows webhook handle"))
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

func (r *workflowsWebhookHandleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state workflowsWebhookHandleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildWorkflowsWebhookHandleRequestBody(ctx, &state)

	resp, _, err := r.Api.CreateWorkflowsWebhookHandle(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Microsoft Workflows webhook handle"))
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

func (r *workflowsWebhookHandleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state workflowsWebhookHandleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body := r.buildWorkflowsWebhookHandleUpdateRequestBody(ctx, &state)

	resp, _, err := r.Api.UpdateWorkflowsWebhookHandle(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Microsoft Workflows webhook handle"))
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

func (r *workflowsWebhookHandleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state workflowsWebhookHandleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteWorkflowsWebhookHandle(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting Microsoft Workflows webhook handle"))
		return
	}
}

func (r *workflowsWebhookHandleResource) updateState(ctx context.Context, state *workflowsWebhookHandleModel, resp *datadogV2.MicrosoftTeamsWorkflowsWebhookHandleResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	attributes := resp.Data.Attributes

	if name, ok := attributes.GetNameOk(); ok && name != nil {
		state.Name = types.StringValue(*name)
	}
}

func (r *workflowsWebhookHandleResource) buildWorkflowsWebhookHandleRequestBody(ctx context.Context, state *workflowsWebhookHandleModel) *datadogV2.MicrosoftTeamsCreateWorkflowsWebhookHandleRequest {
	attributes := datadogV2.NewMicrosoftTeamsWorkflowsWebhookHandleRequestAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	attributes.SetUrl(state.URL.ValueString())

	req := datadogV2.NewMicrosoftTeamsCreateWorkflowsWebhookHandleRequestWithDefaults()
	req.Data = *datadogV2.NewMicrosoftTeamsWorkflowsWebhookHandleRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func (r *workflowsWebhookHandleResource) buildWorkflowsWebhookHandleUpdateRequestBody(ctx context.Context, state *workflowsWebhookHandleModel) *datadogV2.MicrosoftTeamsUpdateWorkflowsWebhookHandleRequest {
	attributes := datadogV2.NewMicrosoftTeamsWorkflowsWebhookHandleAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	attributes.SetUrl(state.URL.ValueString())

	req := datadogV2.NewMicrosoftTeamsUpdateWorkflowsWebhookHandleRequestWithDefaults()
	req.Data = *datadogV2.NewMicrosoftTeamsUpdateWorkflowsWebhookHandleRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}
