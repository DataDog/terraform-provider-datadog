package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type costCustomForecastDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func NewCostCustomForecastDataSource() datasource.DataSource {
	return &costCustomForecastDataSource{}
}

type costCustomForecastDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	BudgetUid types.String `tfsdk:"budget_uid"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.Int64  `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Entries   types.Set    `tfsdk:"entries"`
}

func (d *costCustomForecastDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "cost_custom_forecast"
}

func (d *costCustomForecastDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the custom forecast for an existing Datadog cost budget.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the custom forecast set.",
				Computed:    true,
			},
			"budget_uid": schema.StringAttribute{
				Description: "The UUID of the budget that this custom forecast belongs to.",
				Required:    true,
			},
			"created_at": schema.Int64Attribute{
				Description: "Timestamp the custom forecast was created, in Unix milliseconds.",
				Computed:    true,
			},
			"created_by": schema.StringAttribute{
				Description: "The ID of the user that created the custom forecast.",
				Computed:    true,
			},
			"updated_at": schema.Int64Attribute{
				Description: "Timestamp the custom forecast was last updated, in Unix milliseconds.",
				Computed:    true,
			},
			"updated_by": schema.StringAttribute{
				Description: "The ID of the user that last updated the custom forecast.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"entries": schema.SetNestedBlock{
				Description: "Monthly custom forecast entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"month": schema.Int64Attribute{
							Description: "The month the entry applies to, in `YYYYMM` format.",
							Computed:    true,
						},
						"amount": schema.Float64Attribute{
							Description: "The forecast override amount for the month.",
							Computed:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"tag_filters": schema.ListNestedBlock{
							Description: "Tag filters that scope this entry to a specific budget entry tag combination.",
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

func (d *costCustomForecastDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *costCustomForecastDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state costCustomForecastDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, httpResp, err := d.Api.GetCustomForecast(d.Auth, state.BudgetUid.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom forecast", utils.TranslateClientError(err, httpResp, "").Error())
		return
	}

	resp.Diagnostics.Append(setCostCustomForecastDataSourceModelFromResponse(ctx, &state, apiResp)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setCostCustomForecastDataSourceModelFromResponse(ctx context.Context, model *costCustomForecastDataSourceModel, apiResp datadogV2.CustomForecastResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	data := apiResp.Data
	attrs := data.Attributes

	model.ID = types.StringValue(data.Id)
	model.BudgetUid = types.StringValue(attrs.BudgetUid)
	model.CreatedAt = types.Int64Value(attrs.CreatedAt)
	model.CreatedBy = types.StringValue(attrs.CreatedBy)
	model.UpdatedAt = types.Int64Value(attrs.UpdatedAt)
	model.UpdatedBy = types.StringValue(attrs.UpdatedBy)

	tagObjType := types.ObjectType{AttrTypes: customForecastTagFilterAttrTypes()}
	entries := make([]customForecastEntryModel, 0, len(attrs.Entries))
	for _, e := range attrs.Entries {
		tagFilters := make([]customForecastTagFilterModel, 0, len(e.TagFilters))
		for _, tf := range e.TagFilters {
			tagFilters = append(tagFilters, customForecastTagFilterModel{
				TagKey:   types.StringValue(tf.TagKey),
				TagValue: types.StringValue(tf.TagValue),
			})
		}

		tagFiltersList, d := types.ListValueFrom(ctx, tagObjType, tagFilters)
		diags.Append(d...)
		entries = append(entries, customForecastEntryModel{
			Month:      types.Int64Value(e.Month),
			Amount:     types.Float64Value(e.Amount),
			TagFilters: tagFiltersList,
		})
	}

	entrySet, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: customForecastEntryAttrTypes()}, entries)
	diags.Append(d...)
	model.Entries = entrySet

	return diags
}
