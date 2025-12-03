package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

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
	// Query Parameters
	TableId types.String `tfsdk:"table_id"`
	RowIds  types.List   `tfsdk:"row_ids"`

	// Computed values (list of rows)
	Rows []*rowModel `tfsdk:"rows"`
}

type rowModel struct {
	Id     types.String `tfsdk:"id"`
	Values types.Map    `tfsdk:"values"`
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
		Description: "Use this data source to retrieve specific rows from a Datadog reference table by their primary key values. Works with all reference table source types.",
		Attributes: map[string]schema.Attribute{
			"table_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the reference table to query rows from.",
			},
			"row_ids": schema.ListAttribute{
				Required:    true,
				Description: "List of primary key values (row IDs) to retrieve. These are the values of the table's primary key field(s). Maximum 250 IDs per request.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"rows": schema.ListNestedBlock{
				Description: "List of retrieved rows. Each row contains its ID and field values.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The primary key value of the row.",
						},
						"values": schema.MapAttribute{
							Computed:    true,
							Description: "Map of field names to values for this row. All values are returned as strings.",
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *datadogReferenceTableRowsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogReferenceTableRowsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tableId := state.TableId.ValueString()

	// Extract row IDs from the list
	var rowIds []string
	response.Diagnostics.Append(state.RowIds.ElementsAs(ctx, &rowIds, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate row ID count (API limit is 250)
	if len(rowIds) > 250 {
		response.Diagnostics.AddError(
			"Too many row IDs",
			fmt.Sprintf("Maximum 250 row IDs allowed per request, got %d", len(rowIds)),
		)
		return
	}

	// Call API to get rows by ID with retry logic using exponential backoff
	// Rows are written asynchronously, so we need to retry if the table hasn't synced yet
	// Uses the Terraform SDK retry helper which implements exponential backoff
	var ddResp datadogV2.TableRowResourceArray
	var httpResp *http.Response
	var err error

	// Retry for up to 1 minute with exponential backoff (managed by the SDK)
	retryErr := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
		ddResp, httpResp, err = d.Api.GetRowsByID(d.Auth, tableId, rowIds)
		if err != nil {
			// If we get a 404, the table might not have synced yet - retry
			if httpResp != nil && httpResp.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("rows not found (table may not have synced yet): %v", err))
			}
			// For other errors, don't retry
			return retry.NonRetryableError(fmt.Errorf("error getting reference table rows: %v", err))
		}
		// Success - check if we got the expected number of rows
		if len(ddResp.Data) == len(rowIds) {
			return nil
		}
		// If we got some rows but not all, the table might still be syncing - retry
		if len(ddResp.Data) > 0 && len(ddResp.Data) < len(rowIds) {
			return retry.RetryableError(fmt.Errorf("only %d of %d rows found (table may still be syncing)", len(ddResp.Data), len(rowIds)))
		}
		// If we got no rows, retry
		return retry.RetryableError(fmt.Errorf("no rows found (table may not have synced yet)"))
	})

	if retryErr != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(retryErr, "error getting reference table rows"))
		return
	}

	// Convert API response to state
	state.Rows = make([]*rowModel, len(ddResp.Data))
	for i, row := range ddResp.Data {
		rowTf := &rowModel{
			Id: types.StringValue(row.GetId()),
		}

		// Convert values map to types.Map with string values
		if attrs, ok := row.GetAttributesOk(); ok && attrs.Values != nil {
			// Type assert Values to map[string]interface{}
			if valuesMap, ok := attrs.Values.(map[string]interface{}); ok {
				// Convert all values to strings for the map
				stringValues := make(map[string]string)
				for k, v := range valuesMap {
					// Convert value to string representation
					stringValues[k] = fmt.Sprintf("%v", v)
				}
				rowTf.Values, _ = types.MapValueFrom(ctx, types.StringType, stringValues)
			}
		}

		state.Rows[i] = rowTf
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
