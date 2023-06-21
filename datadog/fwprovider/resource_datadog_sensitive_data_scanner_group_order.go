package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &sensitiveDataScannerGroupOrder{}
	_ resource.ResourceWithImportState = &sensitiveDataScannerGroupOrder{}
)

func NewSensitiveDataScannerGroupOrder() resource.Resource {
	return &sensitiveDataScannerGroupOrder{}
}

type sensitiveDataScannerGroupOrderModel struct {
	ID       types.String `tfsdk:"id"`
	GroupIDs types.List   `tfsdk:"group_ids"`
}

type sensitiveDataScannerGroupOrder struct {
	Api  *datadogV2.SensitiveDataScannerApi
	Auth context.Context
}

func (r *sensitiveDataScannerGroupOrder) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSensitiveDataScannerApiV2()
	r.Auth = providerData.Auth
}

func (r *sensitiveDataScannerGroupOrder) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "sensitive_data_scanner_group_order"
}

func (r *sensitiveDataScannerGroupOrder) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Sensitive Data Scanner Group Order API resource. This can be used to manage the order of Datadog Sensitive Data Scanner Groups.",
		Attributes: map[string]schema.Attribute{
			"group_ids": schema.ListAttribute{
				Description: "The list of Sensitive Data Scanner group IDs, in order. Logs are tested against the query filter of each index one by one following the order of the list.",
				ElementType: types.StringType,
				Required:    true,
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *sensitiveDataScannerGroupOrder) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state sensitiveDataScannerGroupOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *sensitiveDataScannerGroupOrder) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state sensitiveDataScannerGroupOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResponse, err := r.Api.ListScanningGroups(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading SDS groups. http response: %v", httpResponse)))
		return
	}
	var groups []datadogV2.SensitiveDataScannerGroupItem
	var groupID string
	if respData, ok := resp.GetDataOk(); ok {
		if respRelationships, ok := respData.GetRelationshipsOk(); ok {
			if respGroups, ok := respRelationships.GetGroupsOk(); ok {
				groups = respGroups.GetData()
			}
		}
		groupID = respData.GetId()
	}
	tfList := make([]string, len(groups))
	for i, ddGroup := range groups {
		tfList[i] = ddGroup.GetId()
	}

	state.GroupIDs, _ = types.ListValueFrom(ctx, types.StringType, tfList)
	state.ID = types.StringValue(groupID)
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *sensitiveDataScannerGroupOrder) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state sensitiveDataScannerGroupOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *sensitiveDataScannerGroupOrder) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
}

func (r *sensitiveDataScannerGroupOrder) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *sensitiveDataScannerGroupOrder) updateOrder(state *sensitiveDataScannerGroupOrderModel, diag *diag.Diagnostics) {
	ddList := make([]datadogV2.SensitiveDataScannerGroupItem, len(state.GroupIDs.Elements()))
	for i, tfName := range state.GroupIDs.Elements() {
		ddList[i] = *datadogV2.NewSensitiveDataScannerGroupItemWithDefaults()
		ddList[i].SetId(tfName.(types.String).ValueString())
	}

	ddSDSGroupsList, httpResponse, err := r.Api.ListScanningGroups(r.Auth)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error getting Sensitive Data Scanner groups list: %v", httpResponse)))
	}

	SDSGroupOrderRequest := datadogV2.NewSensitiveDataScannerConfigRequestWithDefaults()
	SDSGroupOrderRequestConfig := datadogV2.NewSensitiveDataScannerReorderConfigWithDefaults()
	SDSGroupOrderRequestRelationships := datadogV2.NewSensitiveDataScannerConfigurationRelationshipsWithDefaults()
	SDSGroupOrderRequestGroups := datadogV2.NewSensitiveDataScannerGroupListWithDefaults()
	SDSGroupOrderRequestGroups.SetData(ddList)
	SDSGroupOrderRequestRelationships.SetGroups(*SDSGroupOrderRequestGroups)
	SDSGroupOrderRequestConfig.SetRelationships(*SDSGroupOrderRequestRelationships)
	SDSGroupOrderRequestConfig.SetId(ddSDSGroupsList.Data.GetId())
	SDSGroupOrderRequest.SetData(*SDSGroupOrderRequestConfig)

	updatedOrder, httpResponse, err := r.Api.ReorderScanningGroups(r.Auth, *SDSGroupOrderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error updating Sensitive Data Scanner groups list: %v", httpResponse)))
	}
	if err := utils.CheckForUnparsed(updatedOrder); err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, ""))
	}
	state.ID = types.StringValue(ddSDSGroupsList.Data.GetId())
}
