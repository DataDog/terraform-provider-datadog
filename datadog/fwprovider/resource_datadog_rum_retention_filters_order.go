package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &rumRetentionFiltersOrderResource{}
	_ resource.ResourceWithImportState = &rumRetentionFiltersOrderResource{}
)

type rumRetentionFiltersOrderResource struct {
	Api  *datadogV2.RumRetentionFiltersApi
	Auth context.Context
}

type rumRetentionFiltersOrderResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	ApplicationID     types.String   `tfsdk:"application_id"`
	RetentionFilterID []types.String `tfsdk:"retention_filter_ids"`
}

func NewRumRetentionFiltersOrderResource() resource.Resource {
	return &rumRetentionFiltersOrderResource{}
}

func (r *rumRetentionFiltersOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRumRetentionFiltersApiV2()
	r.Auth = providerData.Auth
}

func (r *rumRetentionFiltersOrderResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_retention_filters_order"
}

func (r *rumRetentionFiltersOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RumRetentionFiltersOrder resource. This is used to manage the order of Datadog RUM retention filters. " +
			"Please note that retention_filter_ids should contain all IDs of retention filters, including the default ones created internally for a given RUM application.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"application_id": schema.StringAttribute{
				Description: "RUM application ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"retention_filter_ids": schema.ListAttribute{
				Description: "RUM retention filter ID list. The order of IDs in this attribute defines the order of RUM retention filters.",
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *rumRetentionFiltersOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("application_id"), request.ID)...)
}

func (r *rumRetentionFiltersOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumRetentionFiltersOrderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.ListRetentionFilters(r.Auth, state.ApplicationID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing RumRetentionFilters"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	retentionFilterIds := make([]types.String, len(resp.GetData()))
	for i, retentionFilter := range resp.GetData() {
		retentionFilterIds[i] = types.StringValue(*retentionFilter.Id)
	}

	r.updateState(&state, retentionFilterIds)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionFiltersOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumRetentionFiltersOrderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.orderRetentionFiltersAndUpdateState(&state, &response.Diagnostics)

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *rumRetentionFiltersOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumRetentionFiltersOrderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.orderRetentionFiltersAndUpdateState(&state, &response.Diagnostics)

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *rumRetentionFiltersOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *rumRetentionFiltersOrderResource) orderRetentionFiltersAndUpdateState(state *rumRetentionFiltersOrderResourceModel, diagnostics *diag.Diagnostics) {
	body, diags := r.buildRetentionFiltersOrderRequestBody(state)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}
	state.ID = state.ApplicationID

	resp, _, err := r.Api.OrderRetentionFilters(r.Auth, state.ApplicationID.ValueString(), *body)
	if err != nil {
		diagnostics.Append(utils.FrameworkErrorDiag(err, "error re-ordering retention filters"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	retentionFilterIds := make([]types.String, len(resp.GetData()))
	for i, retentionFilter := range resp.GetData() {
		retentionFilterIds[i] = types.StringValue(retentionFilter.Id)
	}

	r.updateState(state, retentionFilterIds)
}

func (r *rumRetentionFiltersOrderResource) updateState(state *rumRetentionFiltersOrderResourceModel, retentionFilterIds []types.String) {
	state.RetentionFilterID = retentionFilterIds
}

func (r *rumRetentionFiltersOrderResource) buildRetentionFiltersOrderRequestBody(state *rumRetentionFiltersOrderResourceModel) (*datadogV2.RumRetentionFiltersOrderRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	orderReq := datadogV2.NewRumRetentionFiltersOrderRequestWithDefaults()

	dataList := make([]datadogV2.RumRetentionFiltersOrderData, len(state.RetentionFilterID))
	for i, id := range state.RetentionFilterID {
		dataList[i] = datadogV2.RumRetentionFiltersOrderData{
			Id:   id.ValueString(),
			Type: "retention_filters",
		}
	}

	orderReq.SetData(dataList)
	return orderReq, diags
}
