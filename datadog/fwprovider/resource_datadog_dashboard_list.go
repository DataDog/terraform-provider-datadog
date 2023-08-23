package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &dashboardListResource{}
	_ resource.ResourceWithImportState = &dashboardListResource{}
)

func NewDashboardListResource() resource.Resource {
	return &dashboardListResource{}
}

type dashboardListResource struct {
	ApiV1 *datadogV1.DashboardListsApi
	ApiV2 *datadogV2.DashboardListsApi
	Auth  context.Context
}

type dashboardListResourceModel struct {
	ID       types.String     `tfsdk:"id"`
	Name     types.String     `tfsdk:"name"`
	DashItem []*dashItemModel `tfsdk:"dash_item"`
}

type dashItemModel struct {
	Type   types.String `tfsdk:"type"`
	DashId types.String `tfsdk:"dash_id"`
}

func (r *dashboardListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	r.ApiV1 = providerData.DatadogApiInstances.GetDashboardListsApiV1()
	r.ApiV2 = providerData.DatadogApiInstances.GetDashboardListsApiV2()
	r.Auth = providerData.Auth
}

func (r *dashboardListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

func (r *dashboardListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "dashboard_list"
}

func (r *dashboardListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog dashboard_list resource. This can be used to create and manage Datadog Dashboard Lists and the individual dashboards within them.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the Dashboard List",
				Required:    true,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"dash_item": schema.SetNestedBlock{
				Description: "A set of dashboard items that belong to this list",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The type of this dashboard.",
							Required:    true,
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewDashboardTypeFromValue),
							},
						},
						"dash_id": schema.StringAttribute{
							Description: "The ID of the dashboard to add",
							Required:    true,
						},
					},
				},
			},
		},
	}

}

func (r *dashboardListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state dashboardListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardListPayload, err := buildDatadogDashboardList(&state)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse resource configuration: ", err.Error())
		return
	}

	dashboardList, httpresp, err := r.ApiV1.CreateDashboardList(r.Auth, *dashboardListPayload)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error creating dashboard lists"))
		return
	}
	if err := utils.CheckForUnparsed(dashboardList); err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
		return
	}

	id := dashboardList.GetId()
	state.ID = types.StringValue(strconv.Itoa(int(id)))

	// Add all the dash list items into the List
	if len(state.DashItem) > 0 {
		dashboardListV2Items, err := buildDatadogDashboardListUpdateItemsV2(&state)
		if err != nil {
			resp.Diagnostics.AddError("failed to parse resource configuration: ", err.Error())
			return
		}
		dashboardListUpdateItemsResponse, _, err := r.ApiV2.UpdateDashboardListItems(r.Auth, id, *dashboardListV2Items)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error updating dashboard list item"))
			return
		}
		dashboards := dashboardListUpdateItemsResponse.GetDashboards()
		r.updateStateFromResponse(ctx, &state, dashboards)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state dashboardListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse resource id: ", err.Error())
		return
	}

	// Make any necessary updates to the Overall Dashboard List Object
	dashList, err := buildDatadogDashboardList(&state)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse resource configuration: ", err.Error())
		return
	}

	dashList.SetName(state.Name.ValueString())

	_, httpresp, err := r.ApiV1.UpdateDashboardList(r.Auth, id, *dashList)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error updating dashboard list"))
		return
	}

	// Delete all elements from the dash list and add back only the ones in the config
	completeDashListV2, httpresp, err := r.ApiV2.GetDashboardListItems(r.Auth, id)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error getting dashboard list item"))
		return
	}
	if err := utils.CheckForUnparsed(completeDashListV2); err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}

	completeDashListDeleteV2, err := buildDatadogDashboardListDeleteItemsV2(&completeDashListV2)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error creating dashboard list delete item"))
		return
	}
	_, httpresp, err = r.ApiV2.DeleteDashboardListItems(r.Auth, id, *completeDashListDeleteV2)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error deleting dashboard list item"))
		return
	}

	if len(state.DashItem) > 0 {
		dashboardListV2Items, err := buildDatadogDashboardListUpdateItemsV2(&state)
		if err != nil {
			resp.Diagnostics.AddError("failed to parse resource configuration: ", err.Error())
			return
		}
		dashboardListUpdateItemsResponse, httpresp, err := r.ApiV2.UpdateDashboardListItems(r.Auth, id, *dashboardListV2Items)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error updating dashboard list item"))
			return
		}
		r.updateStateFromResponse(ctx, &state, dashboardListUpdateItemsResponse.GetDashboards())
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse resource id: ", err.Error())
		return
	}

	//Read the overall Dashboard List object
	dashList, httpresp, err := r.ApiV1.GetDashboardList(r.Auth, id)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error getting dashboard list"))
		return
	}
	state.Name = types.StringValue(dashList.GetName())

	// Read and set all the dashboard list elements
	completeItemListV2, _, err := r.ApiV2.GetDashboardListItems(r.Auth, id)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error getting dashboard list item"))
		return
	}
	if err := utils.CheckForUnparsed(completeItemListV2); err != nil {
		resp.Diagnostics.AddError("", err.Error())
		return
	}
	r.updateStateFromDashItem(ctx, &state, completeItemListV2.GetDashboards())
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, _ := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	_, httpresp, err := r.ApiV1.DeleteDashboardList(r.Auth, id)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error deleting dashboard list"))
		return
	}
}

func buildDatadogDashboardList(state *dashboardListResourceModel) (*datadogV1.DashboardList, error) {
	var dashboardList datadogV1.DashboardList
	dashboardList.SetName(state.Name.ValueString())
	return &dashboardList, nil
}

func buildDatadogDashboardListUpdateItemsV2(state *dashboardListResourceModel) (*datadogV2.DashboardListUpdateItemsRequest, error) {
	dashboardListV2ItemsArr := make([]datadogV2.DashboardListItemRequest, 0)
	for _, dashItem := range state.DashItem {
		dashType := datadogV2.DashboardType(dashItem.Type.ValueString())
		dashItem := datadogV2.NewDashboardListItemRequest(dashItem.DashId.ValueString(), dashType)
		dashboardListV2ItemsArr = append(dashboardListV2ItemsArr, *dashItem)
	}
	dashboardListV2Items := datadogV2.NewDashboardListUpdateItemsRequest()
	dashboardListV2Items.SetDashboards(dashboardListV2ItemsArr)
	return dashboardListV2Items, nil
}

func buildDatadogDashboardListDeleteItemsV2(dashboardListItems *datadogV2.DashboardListItems) (*datadogV2.DashboardListDeleteItemsRequest, error) {
	dashboardListV2ItemsArr := make([]datadogV2.DashboardListItemRequest, 0)
	for _, dashItem := range dashboardListItems.GetDashboards() {
		dashType := dashItem.GetType()
		dashID := dashItem.GetId()
		dashItem := datadogV2.NewDashboardListItemRequest(dashID, dashType)
		dashboardListV2ItemsArr = append(dashboardListV2ItemsArr, *dashItem)
	}
	dashboardListV2Items := datadogV2.NewDashboardListDeleteItemsRequest()
	dashboardListV2Items.SetDashboards(dashboardListV2ItemsArr)
	return dashboardListV2Items, nil
}

func (r *dashboardListResource) updateStateFromResponse(ctx context.Context, state *dashboardListResourceModel, dashboards []datadogV2.DashboardListItemResponse) {
	dashItemList := []*dashItemModel{}
	for _, dashboard := range dashboards {
		dashboardItem := dashItemModel{}
		dashboardItem.DashId = types.StringValue(dashboard.GetId())
		found := false
		for _, dashItem := range state.DashItem {
			if dashItem.DashId.ValueString() == dashboard.GetId() {
				found = true
				dashboardItem.Type = dashItem.Type
			}
		}
		if !found {
			dashboardItem.Type = types.StringValue(string(dashboard.GetType()))
		}
		dashItemList = append(dashItemList, &dashboardItem)
	}
	state.DashItem = dashItemList
}

func (r *dashboardListResource) updateStateFromDashItem(ctx context.Context, state *dashboardListResourceModel, dashboards []datadogV2.DashboardListItem) {
	dashItemList := []*dashItemModel{}
	for _, dashboard := range dashboards {
		dashboardItem := dashItemModel{}
		dashboardItem.DashId = types.StringValue(dashboard.GetId())
		found := false
		for _, dashItem := range state.DashItem {
			if dashItem.DashId.ValueString() == dashboard.GetId() {
				found = true
				dashboardItem.Type = dashItem.Type
			}
		}
		if !found {
			dashboardItem.Type = types.StringValue(string(dashboard.GetType()))
		}
		dashItemList = append(dashItemList, &dashboardItem)
	}
	state.DashItem = dashItemList
}
