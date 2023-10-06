package fwprovider

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure   = &ApmRetentionFiltersOrderResource{}
	_ resource.ResourceWithImportState = &ApmRetentionFiltersOrderResource{}
)

type ApmRetentionFiltersOrderResource struct {
	Api  *datadogV2.APMRetentionFiltersApi
	Auth context.Context
}
type ApmRetentionFiltersOrderModel struct {
	ID        types.String   `tfsdk:"id"`
	FilterIds []types.String `tfsdk:"filter_ids"`
}

func NewApmRetentionFiltersOrderResource() resource.Resource {
	return &ApmRetentionFiltersOrderResource{}
}
func (r *ApmRetentionFiltersOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetApmRetentionFiltersApiV2()
	r.Auth = providerData.Auth
}

func (r *ApmRetentionFiltersOrderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "apm_retention_filter_order"
}
func (d *ApmRetentionFiltersOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog [APM Retention Filters API](https://docs.datadoghq.com/api/v2/apm-retention-filters/) resource, which is used to manage Datadog APM retention filters order.",
		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				Description: "The filter IDs list. The order of filters IDs in this attribute defines the overall APM retention filters order.. If `filter_ids` is empty or not specified, it will import the actual order, and create the resource. Otherwise, it will try to update the order.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"id": utils.ResourceIDAttribute(),
		}}
}

func (r *ApmRetentionFiltersOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *ApmRetentionFiltersOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ApmRetentionFiltersOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.ListApmRetentionFilters(r.Auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving retention filter"))
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

func (r *ApmRetentionFiltersOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state ApmRetentionFiltersOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if len(state.FilterIds) > 0 {
		body, diags := r.buildRetentionFiltersOrderRequestBody(ctx, &state)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		_, err := r.Api.ReorderApmRetentionFilters(r.Auth, *body)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error re-ordering retention filters"))
			return

		}
	}
	listData, httpResponse, err := r.Api.ListApmRetentionFilters(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving spans metric"))
		return
	}
	if err := utils.CheckForUnparsed(httpResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &listData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ApmRetentionFiltersOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state ApmRetentionFiltersOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	body, diags := r.buildRetentionFiltersOrderRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.ReorderApmRetentionFilters(r.Auth, *body)
	listData, httpResponse, err := r.Api.ListApmRetentionFilters(r.Auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 400 {
			if err != nil || httpResponse.StatusCode >= 400 {
				response.Diagnostics.AddError("response contains unparsedObject", err.Error())
				return
			}
			r.updateState(ctx, &state, &listData)
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating order, the current order is saved into the state"))
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating retention filters order"))
		return
	}

	r.updateState(ctx, &state, &listData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
func (r *ApmRetentionFiltersOrderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}

func GetApmFilterIds(apmRetentionFiltersOrder datadogV2.RetentionFiltersResponse) []types.String {
	filterIds := make([]types.String, len(apmRetentionFiltersOrder.Data))
	for i, rf := range apmRetentionFiltersOrder.Data {
		filterIds[i] = types.StringValue(rf.Id)
	}
	return filterIds
}

func (r *ApmRetentionFiltersOrderResource) updateState(ctx context.Context, state *ApmRetentionFiltersOrderModel, resp *datadogV2.RetentionFiltersResponse) {
	filterIds := GetApmFilterIds(*resp)
	state.ID = types.StringValue("filtersOrderID")
	state.FilterIds = filterIds
}

func (r *ApmRetentionFiltersOrderResource) buildRetentionFiltersOrderRequestBody(ctx context.Context, state *ApmRetentionFiltersOrderModel) (*datadogV2.ReorderRetentionFiltersRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	filtersOrderReq := datadogV2.NewReorderRetentionFiltersRequestWithDefaults()
	filtersOrderReq.SetData(getFilterIdList(state))
	return filtersOrderReq, diags
}

func getFilterIdList(state *ApmRetentionFiltersOrderModel) []datadogV2.RetentionFilterWithoutAttributes {
	rfList := make([]datadogV2.RetentionFilterWithoutAttributes, len(state.FilterIds))
	for i, id := range state.FilterIds {
		rfList[i] = datadogV2.RetentionFilterWithoutAttributes{
			Id:   id.ValueString(),
			Type: "apm_retention_filter",
		}
	}
	return rfList
}
