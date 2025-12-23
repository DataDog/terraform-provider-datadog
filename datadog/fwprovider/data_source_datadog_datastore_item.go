package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datastoreItemDataSource{}
)

type datastoreItemDataSource struct {
	Api  *datadogV2.ActionsDatastoresApi
	Auth context.Context
}

type datastoreItemDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	DatastoreID types.String `tfsdk:"datastore_id"`
	ItemKey     types.String `tfsdk:"item_key"`

	// Computed values
	Value      types.Map    `tfsdk:"value"`
	CreatedAt  types.String `tfsdk:"created_at"`
	ModifiedAt types.String `tfsdk:"modified_at"`
	OrgID      types.Int64  `tfsdk:"org_id"`
	StoreID    types.String `tfsdk:"store_id"`
	Signature  types.String `tfsdk:"signature"`
}

func NewDatastoreItemDataSource() datasource.DataSource {
	return &datastoreItemDataSource{}
}

func (d *datastoreItemDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetActionsDatastoresApiV2()
	d.Auth = providerData.Auth
}

func (d *datastoreItemDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "datadog_datastore_item"
}

func (d *datastoreItemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog datastore item.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"datastore_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the datastore containing the item.",
			},
			"item_key": schema.StringAttribute{
				Required:    true,
				Description: "The primary key value that identifies the item to retrieve.",
			},
			// Computed values
			"value": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The data content (as key-value pairs) of the datastore item.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the item was first created.",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the item was last modified.",
			},
			"org_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the organization that owns this item.",
			},
			"store_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the datastore containing this item.",
			},
			"signature": schema.StringAttribute{
				Computed:    true,
				Description: "A unique signature identifying this item version.",
			},
		},
	}
}

func (d *datastoreItemDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datastoreItemDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	datastoreID := state.DatastoreID.ValueString()
	itemKey := state.ItemKey.ValueString()

	// Use ListDatastoreItems with item_key query parameter to get specific item
	optionalParams := datadogV2.NewListDatastoreItemsOptionalParameters()
	optionalParams.ItemKey = &itemKey

	resp, _, err := d.Api.ListDatastoreItems(d.Auth, datastoreID, *optionalParams)
	if err != nil {
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
		response.Diagnostics.AddError(
			"Item not found",
			fmt.Sprintf("No item found with key '%s' in datastore '%s'", itemKey, datastoreID),
		)
		return
	}

	d.updateState(ctx, &state, &items[0])

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datastoreItemDataSource) updateState(ctx context.Context, state *datastoreItemDataSourceModel, resp *datadogV2.ItemApiPayloadData) {
	datastoreID := state.DatastoreID.ValueString()
	itemKey := state.ItemKey.ValueString()

	// Set composite ID
	state.ID = types.StringValue(fmt.Sprintf("%s:%s", datastoreID, itemKey))

	attributes := resp.GetAttributes()

	// Set computed values
	if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
		state.ModifiedAt = types.StringValue(modifiedAt.String())
	}

	if orgID, ok := attributes.GetOrgIdOk(); ok && orgID != nil {
		state.OrgID = types.Int64Value(*orgID)
	}

	if storeID, ok := attributes.GetStoreIdOk(); ok && storeID != nil {
		state.StoreID = types.StringValue(*storeID)
	}

	if signature, ok := attributes.GetSignatureOk(); ok && signature != nil {
		state.Signature = types.StringValue(*signature)
	}

	// Convert value map to types.Map
	if value, ok := attributes.GetValueOk(); ok && value != nil {
		valueMap := make(map[string]string)
		for k, v := range *value {
			valueMap[k] = fmt.Sprintf("%v", v)
		}
		mapValue, diags := types.MapValueFrom(ctx, types.StringType, valueMap)
		if !diags.HasError() {
			state.Value = mapValue
		}
	}
}
