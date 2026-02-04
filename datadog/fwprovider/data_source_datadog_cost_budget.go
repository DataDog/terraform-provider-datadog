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
	BudgetLine   types.Set     `tfsdk:"budget_line"`
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
				DeprecationMessage: "Use budget_line instead. The entries block will be removed in a future version.",
				Description:        "The flat list of budget entries (deprecated - use budget_line instead).",
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
			"budget_line": schema.SetNestedBlock{
				Description: "Budget entries grouped by tag combination with amounts map (month -> amount).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"amounts": schema.MapAttribute{
							Description: "Map of month (YYYYMM as string) to budget amount.",
							Computed:    true,
							ElementType: types.Float64Type,
						},
					},
					Blocks: map[string]schema.Block{
						"tag_filters": schema.ListNestedBlock{
							Description: "Tag filters for non-hierarchical budgets (single tag or no tags).",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key":   schema.StringAttribute{Computed: true},
									"tag_value": schema.StringAttribute{Computed: true},
								},
							},
						},
						"parent_tag_filters": schema.ListNestedBlock{
							Description: "Parent tag filters for hierarchical budgets (first tag in 'by' clause).",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key":   schema.StringAttribute{Computed: true},
									"tag_value": schema.StringAttribute{Computed: true},
								},
							},
						},
						"child_tag_filters": schema.ListNestedBlock{
							Description: "Child tag filters for hierarchical budgets (second tag in 'by' clause).",
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
	data, attr := apiResp.Data, apiResp.Data.Attributes

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

	// Convert API entries to internal model
	entries := apiEntriesToBudgetEntries(ctx, attr.Entries)

	// Populate both schemas for backward compatibility
	model.Entries, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: budgetEntryAttrTypes()}, entries)
	model.BudgetLine, _ = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: budgetLineAttrTypes()}, convertFlatEntriesToBudgetLine(ctx, entries))
}
