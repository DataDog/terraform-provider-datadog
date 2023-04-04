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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &IntegrationFastlyServiceResource{}
	_ resource.ResourceWithImportState = &IntegrationFastlyServiceResource{}
)

type IntegrationFastlyServiceResource struct {
	Api  *datadogV2.FastlyIntegrationApi
	Auth context.Context
}

type IntegrationFastlyServiceModel struct {
	ID        types.String `tfsdk:"id"`
	AccountId types.String `tfsdk:"account_id"`
	Tags      types.Set    `tfsdk:"tags"`
}

func NewIntegrationFastlyServiceResource() resource.Resource {
	return &IntegrationFastlyServiceResource{}
}

func (r *IntegrationFastlyServiceResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError("Unexpected Resource Configure Type", "")
		return
	}

	r.Api = providerData.DatadogApiInstances.GetFastlyIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *IntegrationFastlyServiceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "integration_fastly_service"
}

func (r *IntegrationFastlyServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationFastlyService resource. This can be used to create and manage Datadog integration_fastly_service.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Optional:    true,
				Description: "Fastly Account id.",
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Description: "A list of tags for the Fastly service.",
				ElementType: types.StringType,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the Fastly service.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *IntegrationFastlyServiceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *IntegrationFastlyServiceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state IntegrationFastlyServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	accountId := state.AccountId.ValueString()

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetFastlyService(r.Auth, accountId, id)
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

func (r *IntegrationFastlyServiceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state IntegrationFastlyServiceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountId := state.AccountId.ValueString()

	body, diags := r.buildIntegrationFastlyServiceRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateFastlyService(r.Auth, accountId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationFastlyService"))
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

func (r *IntegrationFastlyServiceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state IntegrationFastlyServiceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountId := state.AccountId.ValueString()

	id := state.ID.ValueString()

	body, diags := r.buildIntegrationFastlyServiceRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateFastlyService(r.Auth, accountId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationFastlyService"))
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

func (r *IntegrationFastlyServiceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state IntegrationFastlyServiceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	accountId := state.AccountId.ValueString()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteFastlyService(r.Auth, accountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_fastly_service"))
		return
	}
}

func (r *IntegrationFastlyServiceResource) updateState(ctx context.Context, state *IntegrationFastlyServiceModel, resp *datadogV2.FastlyServiceResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		state.Tags, _ = types.SetValueFrom(ctx, types.StringType, *tags)
	}
}

func (r *IntegrationFastlyServiceResource) buildIntegrationFastlyServiceRequestBody(ctx context.Context, state *IntegrationFastlyServiceModel) (*datadogV2.FastlyServiceRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewFastlyServiceAttributesWithDefaults()

	if !state.Tags.IsNull() {
		var tags []string
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	req := datadogV2.NewFastlyServiceRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyServiceDataWithDefaults()
	req.Data.SetId(state.ID.ValueString())
	req.Data.SetAttributes(*attributes)

	return req, diags
}
