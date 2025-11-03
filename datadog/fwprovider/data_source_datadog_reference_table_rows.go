package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogReferenceTableRowsDataSource{}
)

type datadogReferenceTableRowsDataSource struct {
	Api  *datadogV2.ReferenceTablesApi
	Auth context.Context
}

type datadogReferenceTableRowsDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	RowIds types.List `tfsdk:"row_ids"`

	// Computed values
	Data []*dataModel `tfsdk:"data"`
}

type dataModel struct {
	Id         types.String     `tfsdk:"id"`
	Type       types.String     `tfsdk:"type"`
	Attributes *attributesModel `tfsdk:"attributes"`
}
type attributesModel struct {
	Values map[string]interface{} `tfsdk:"values"`
}

func NewDatadogReferenceTableRowsDataSource() datasource.DataSource {
	return &datadogReferenceTableRowsDataSource{}
}

func (d *datadogReferenceTableRowsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetReferenceTablesApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogReferenceTableRowsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "reference_table_rows"
}

func (d *datadogReferenceTableRowsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog reference_table_rows.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			"row_ids": schema.ListAttribute{
				Optional:    false,
				Description: "Array of row IDs to retrieve.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			// Computed values
			"data": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the row.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "Row resource type.",
						},
					},
					Blocks: map[string]schema.Block{
						"attributes": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{},
							Blocks: map[string]schema.Block{
								"values": schema.SingleNestedBlock{
									Attributes: map[string]schema.Attribute{},
								},
							},
						},
					},
				},
			},
		},
	}
}

// TODO: how are we supposed to handle multiple rows? Should we return a list of rows or a single row?
func (d *datadogReferenceTableRowsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogReferenceTableRowsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		tableId := state.ID.ValueString()
		// parse through data to get rows ids
		rowsIds := make([]string, len(state.RowIds.Elements()))
		for i, rowId := range state.RowIds.Elements() {
			rowsIds[i] = rowId.String()
		}
		ddResp, _, err := d.Api.GetRowsByID(d.Auth, tableId, rowsIds)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog referenceTableRows"))
			return
		}

		if len(ddResp.Data) == 0 {
			response.Diagnostics.AddError("query returned no results, check the row_ids parameter", "")
			return
		} else if len(ddResp.Data) > 1 {
			response.Diagnostics.AddError("query returned more than one result, check the row_ids parameter", "")
			return
		}

		d.updateStateFromListResponse(ctx, &state, ddResp.Data)
	}
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// TODO: how are we supposed to handle multiple rows? Should we return a list of rows or a single row?
func (d *datadogReferenceTableRowsDataSource) updateStateFromListResponse(ctx context.Context, state *datadogReferenceTableRowsDataSourceModel, referenceTableRowsData []datadogV2.TableRowResourceData) {
	state.Data = make([]*dataModel, len(referenceTableRowsData))
	for i, row := range referenceTableRowsData {
		state.Data[i] = &dataModel{
			Id:   types.StringValue(row.GetId()),
			Type: types.StringValue(string(datadogV2.TABLEROWRESOURCEDATATYPE_ROW)),
			Attributes: &attributesModel{
				Values: row.GetAttributes().Values,
			},
		}
	}
}
