package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type costBudgetDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func NewCostBudgetDataSource() datasource.DataSource {
	return &costBudgetDataSource{}
}

type costBudgetDataSourceModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	MetricsQuery types.String  `tfsdk:"metrics_query"`
	StartMonth   types.Int64   `tfsdk:"start_month"`
	EndMonth     types.Int64   `tfsdk:"end_month"`
	TotalAmount  types.Float64 `tfsdk:"total_amount"`
	Entries      types.List    `tfsdk:"entries"`
}

func (d *costBudgetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "cost_budget"
}

func (d *costBudgetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog cost budget.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the budget.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the budget.",
				Computed:    true,
			},
			"metrics_query": schema.StringAttribute{
				Description: "The cost query used to track against the budget.",
				Computed:    true,
			},
			"start_month": schema.Int64Attribute{
				Description: "The month when the budget starts (YYYYMM).",
				Computed:    true,
			},
			"end_month": schema.Int64Attribute{
				Description: "The month when the budget ends (YYYYMM).",
				Computed:    true,
			},
			"total_amount": schema.Float64Attribute{
				Description: "The sum of all budget entries' amounts.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"entries": schema.ListNestedBlock{
				Description: "The entries of the budget.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"amount": schema.Float64Attribute{
							Computed: true,
						},
						"month": schema.Int64Attribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"tag_filters": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key":   schema.StringAttribute{Computed: true},
									"tag_value": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *costBudgetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *costBudgetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state costBudgetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, _, err := d.Api.GetBudget(d.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading budget", err.Error())
		return
	}

	setDataSourceModelFromBudgetWithEntries(ctx, &state, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setDataSourceModelFromBudgetWithEntries(ctx context.Context, model *costBudgetDataSourceModel, apiResp datadogV2.BudgetWithEntries) {
	if apiResp.Data == nil || apiResp.Data.Attributes == nil {
		return
	}
	data := apiResp.Data
	attr := data.Attributes

	// Set top-level fields
	if data.Id != nil {
		model.ID = types.StringValue(*data.Id)
	}
	if attr.Name != nil {
		model.Name = types.StringValue(*attr.Name)
	}
	if attr.MetricsQuery != nil {
		model.MetricsQuery = types.StringValue(*attr.MetricsQuery)
	}
	if attr.StartMonth != nil {
		model.StartMonth = types.Int64Value(*attr.StartMonth)
	}
	if attr.EndMonth != nil {
		model.EndMonth = types.Int64Value(*attr.EndMonth)
	}
	if attr.TotalAmount != nil {
		model.TotalAmount = types.Float64Value(*attr.TotalAmount)
	}

	// Set entries
	var entries []budgetEntry
	for _, apiEntry := range attr.Entries {
		var tagFilters []tagFilter
		for _, tf := range apiEntry.TagFilters {
			var tagKey, tagValue types.String
			if tf.TagKey != nil {
				tagKey = types.StringValue(*tf.TagKey)
			} else {
				tagKey = types.StringNull()
			}
			if tf.TagValue != nil {
				tagValue = types.StringValue(*tf.TagValue)
			} else {
				tagValue = types.StringNull()
			}
			tagFilters = append(tagFilters, tagFilter{
				TagKey:   tagKey,
				TagValue: tagValue,
			})
		}

		var amount types.Float64
		if apiEntry.Amount != nil {
			amount = types.Float64Value(*apiEntry.Amount)
		} else {
			amount = types.Float64Null()
		}
		var month types.Int64
		if apiEntry.Month != nil {
			month = types.Int64Value(*apiEntry.Month)
		} else {
			month = types.Int64Null()
		}

		// Convert []tagFilter to types.List
		tagFiltersList, _ := types.ListValueFrom(
			ctx,
			types.ObjectType{AttrTypes: tagFilterAttrTypes()},
			tagFilters,
		)

		entries = append(entries, budgetEntry{
			Amount:     amount,
			Month:      month,
			TagFilters: tagFiltersList,
		})
	}

	// Convert []budgetEntry to types.List
	model.Entries, _ = types.ListValueFrom(
		ctx,
		types.ObjectType{AttrTypes: budgetEntryAttrTypes()},
		entries,
	)
}
