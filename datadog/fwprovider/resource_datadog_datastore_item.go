package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &datastoreItemResource{}
	_ resource.ResourceWithImportState = &datastoreItemResource{}
)

type datastoreItemResource struct {
	Api  *datadogV2.ActionsDatastoresApi
	Auth context.Context
}

type datastoreItemModel struct {
	ID          types.String `tfsdk:"id"`
	DatastoreID types.String `tfsdk:"datastore_id"`
	ItemKey     types.String `tfsdk:"item_key"`
	Value       types.Map    `tfsdk:"value"`
}

func NewDatastoreItemResource() resource.Resource {
	return &datastoreItemResource{}
}

func (r *datastoreItemResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetActionsDatastoresApiV2()
	r.Auth = providerData.Auth
}

func (r *datastoreItemResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "datastore_item"
}

func (r *datastoreItemResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Datastore Item resource. This can be used to create and manage items in a Datadog datastore.",
		Attributes: map[string]schema.Attribute{
			"datastore_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the datastore containing this item.",
			},
			"item_key": schema.StringAttribute{
				Required:    true,
				Description: "The primary key value that identifies this item. Cannot exceed 256 characters.",
			},
			"value": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The data content (as key-value pairs) of the datastore item.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *datastoreItemResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Import format: datastore_id:item_key
	parts := strings.SplitN(request.ID, ":", 2)
	if len(parts) != 2 {
		response.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'datastore_id:item_key', got: %s", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("datastore_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("item_key"), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("id"), request.ID)...)
}

func (r *datastoreItemResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state datastoreItemModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	datastoreID := state.DatastoreID.ValueString()
	itemKey := state.ItemKey.ValueString()

	// Use ListDatastoreItems with item_key query parameter to get specific item
	optionalParams := datadogV2.NewListDatastoreItemsOptionalParameters()
	optionalParams.ItemKey = &itemKey

	resp, httpResp, err := r.Api.ListDatastoreItems(r.Auth, datastoreID, *optionalParams)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Datastore Item"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Check if item was found
	items := resp.GetData()
	if len(items) == 0 {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(ctx, &state, &items[0])

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datastoreItemResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state datastoreItemModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	datastoreID := state.DatastoreID.ValueString()
	itemKey := state.ItemKey.ValueString()

	body, diags := r.buildDatastoreItemRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.BulkWriteDatastoreItems(r.Auth, datastoreID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating Datastore Item"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Set the composite ID using the plan values (not from API response)
	// All fields are user-provided, so we don't need to read back from the API
	state.ID = types.StringValue(fmt.Sprintf("%s:%s", datastoreID, itemKey))

	// Save data into Terraform state using the plan values directly
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datastoreItemResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state datastoreItemModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	datastoreID := state.DatastoreID.ValueString()

	body, diags := r.buildDatastoreItemUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateDatastoreItem(r.Auth, datastoreID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating Datastore Item"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	itemData := resp.GetData()
	r.updateState(ctx, &state, &itemData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datastoreItemResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state datastoreItemModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	datastoreID := state.DatastoreID.ValueString()
	itemKey := state.ItemKey.ValueString()

	// Build delete request
	body := datadogV2.NewDeleteAppsDatastoreItemRequestWithDefaults()
	body.Data = datadogV2.NewDeleteAppsDatastoreItemRequestDataWithDefaults()
	attributes := datadogV2.NewDeleteAppsDatastoreItemRequestDataAttributesWithDefaults()
	attributes.SetItemKey(itemKey)
	body.Data.SetAttributes(*attributes)

	_, httpResp, err := r.Api.DeleteDatastoreItem(r.Auth, datastoreID, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting datastore_item"))
		return
	}
}

func (r *datastoreItemResource) updateState(ctx context.Context, state *datastoreItemModel, resp *datadogV2.ItemApiPayloadData) {
	// Preserve datastoreID and itemKey from the existing state
	// These are Required fields and should not be overwritten from the API response
	datastoreID := state.DatastoreID.ValueString()
	itemKey := state.ItemKey.ValueString()

	// Set composite ID using preserved values
	state.ID = types.StringValue(fmt.Sprintf("%s:%s", datastoreID, itemKey))

	// Note: We don't update state.Value from the API response because it's a Required field
	// that comes from the user's configuration. The API should return the same values we sent.
}

func (r *datastoreItemResource) buildDatastoreItemRequestBody(ctx context.Context, state *datastoreItemModel) (*datadogV2.BulkPutAppsDatastoreItemsRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Convert the value map to a format suitable for the API
	// First convert to map[string]string, then to map[string]interface{}
	stringElements := make(map[string]string)
	diags.Append(state.Value.ElementsAs(ctx, &stringElements, false)...)
	if diags.HasError() {
		return nil, diags
	}
	valueElements := make(map[string]interface{})
	for k, v := range stringElements {
		valueElements[k] = v
	}

	values := []map[string]interface{}{valueElements}

	req := datadogV2.NewBulkPutAppsDatastoreItemsRequestWithDefaults()
	req.Data = datadogV2.NewBulkPutAppsDatastoreItemsRequestDataWithDefaults()
	attributes := datadogV2.NewBulkPutAppsDatastoreItemsRequestDataAttributesWithDefaults()
	attributes.SetValues(values)
	// Use fail_on_conflict for create to prevent accidental overwrites
	attributes.SetConflictMode(datadogV2.DATASTOREITEMCONFLICTMODE_FAIL_ON_CONFLICT)
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *datastoreItemResource) buildDatastoreItemUpdateRequestBody(ctx context.Context, state *datastoreItemModel) (*datadogV2.UpdateAppsDatastoreItemRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Convert the value map to ops_set format
	// First convert to map[string]string, then to map[string]interface{}
	stringElements := make(map[string]string)
	diags.Append(state.Value.ElementsAs(ctx, &stringElements, false)...)
	if diags.HasError() {
		return nil, diags
	}
	valueElements := make(map[string]interface{})
	for k, v := range stringElements {
		valueElements[k] = v
	}

	req := datadogV2.NewUpdateAppsDatastoreItemRequestWithDefaults()
	req.Data = datadogV2.NewUpdateAppsDatastoreItemRequestDataWithDefaults()
	attributes := datadogV2.NewUpdateAppsDatastoreItemRequestDataAttributesWithDefaults()
	attributes.SetItemKey(state.ItemKey.ValueString())

	itemChanges := datadogV2.NewUpdateAppsDatastoreItemRequestDataAttributesItemChangesWithDefaults()
	itemChanges.SetOpsSet(valueElements)
	attributes.SetItemChanges(*itemChanges)

	req.Data.SetAttributes(*attributes)

	return req, diags
}
