package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type costBudgetResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func NewCostBudgetResource() resource.Resource {
	return &costBudgetResource{}
}

type costBudgetModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	MetricsQuery types.String  `tfsdk:"metrics_query"`
	StartMonth   types.Int64   `tfsdk:"start_month"`
	EndMonth     types.Int64   `tfsdk:"end_month"`
	TotalAmount  types.Float64 `tfsdk:"total_amount"`
	Entries      []budgetEntry `tfsdk:"entries"`
}

type budgetEntry struct {
	Amount     types.Float64 `tfsdk:"amount"`
	Month      types.Int64   `tfsdk:"month"`
	TagFilters []tagFilter   `tfsdk:"tag_filters"`
}

type tagFilter struct {
	TagKey   types.String `tfsdk:"tag_key"`
	TagValue types.String `tfsdk:"tag_value"`
}

func (r *costBudgetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "cost_budget"
}

func (r *costBudgetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Cost Budget resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the budget.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the budget.",
			},
			"metrics_query": schema.StringAttribute{
				Required:    true,
				Description: "The cost query used to track against the budget. **Note:** For hierarchical budgets using `by {tag1,tag2}`, the order of tags determines the UI hierarchy (parent, child).",
			},
			"start_month": schema.Int64Attribute{
				Required:    true,
				Description: "The month when the budget starts (YYYYMM).",
			},
			"end_month": schema.Int64Attribute{
				Required:    true,
				Description: "The month when the budget ends (YYYYMM).",
			},
			"total_amount": schema.Float64Attribute{
				Computed:    true,
				Description: "The sum of all budget entries' amounts.",
			},
		},
		Blocks: map[string]schema.Block{
			"entries": schema.ListNestedBlock{
				Description: "The entries of the budget. **Note:** You must provide entries for all months in the budget period. For hierarchical budgets, each unique tag combination must have entries for all months.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"amount": schema.Float64Attribute{
							Required: true,
						},
						"month": schema.Int64Attribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"tag_filters": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key": schema.StringAttribute{
										Required:    true,
										Description: "**Note:** Must be one of the tags from the `metrics_query`.",
									},
									"tag_value": schema.StringAttribute{Required: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *costBudgetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

// --- CRUD ---

func (r *costBudgetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan costBudgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildBudgetWithEntriesFromModel(plan)
	apiResp, response, err := r.Api.UpsertBudget(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating budget", utils.TranslateClientError(err, response, "").Error())
		return
	}

	setModelFromBudgetWithEntries(&plan, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *costBudgetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state costBudgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, response, err := r.Api.GetBudget(r.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading budget", utils.TranslateClientError(err, response, "").Error())
		return
	}

	setModelFromBudgetWithEntries(&state, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *costBudgetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan costBudgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// we need to retrieve the ID from the current state and copy it to the plan
	// otherwise the API will create a new budget instead of updating
	var state costBudgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	apiReq := buildBudgetWithEntriesFromModel(plan)
	apiResp, response, err := r.Api.UpsertBudget(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating budget", utils.TranslateClientError(err, response, "").Error())
		return
	}

	setModelFromBudgetWithEntries(&plan, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *costBudgetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state costBudgetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.Api.DeleteBudget(r.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting budget", err.Error())
		return
	}
}

func (r *costBudgetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

// --- Helper functions to map between model and API types go here ---
func buildBudgetWithEntriesFromModel(plan costBudgetModel) datadogV2.BudgetWithEntries {
	// Convert entries
	var entries []datadogV2.BudgetEntry
	for _, e := range plan.Entries {
		var tagFilters []datadogV2.TagFilter
		for _, tf := range e.TagFilters {
			tagFilters = append(tagFilters, datadogV2.TagFilter{
				TagKey:   tf.TagKey.ValueStringPointer(),
				TagValue: tf.TagValue.ValueStringPointer(),
			})
		}
		entries = append(entries, datadogV2.BudgetEntry{
			Amount:     e.Amount.ValueFloat64Pointer(),
			Month:      e.Month.ValueInt64Pointer(),
			TagFilters: tagFilters,
		})
	}

	// Build attributes
	attributes := datadogV2.BudgetAttributes{
		Name:         plan.Name.ValueStringPointer(),
		MetricsQuery: plan.MetricsQuery.ValueStringPointer(),
		StartMonth:   plan.StartMonth.ValueInt64Pointer(),
		EndMonth:     plan.EndMonth.ValueInt64Pointer(),
		Entries:      entries,
		// total_amount is computed by the API, not sent in the request
	}

	// Build data
	budgetType := "budget"
	data := datadogV2.BudgetWithEntriesData{
		Attributes: &attributes,
		Type:       &budgetType,
	}
	// If updating, you may need to set ID
	if !plan.ID.IsNull() && plan.ID.ValueString() != "" {
		data.Id = plan.ID.ValueStringPointer()
	}

	// Build and return the top-level object
	return datadogV2.BudgetWithEntries{
		Data: &data,
	}
}

func setModelFromBudgetWithEntries(model *costBudgetModel, apiResp datadogV2.BudgetWithEntries) {
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

		entries = append(entries, budgetEntry{
			Amount:     amount,
			Month:      month,
			TagFilters: tagFilters,
		})
	}
	model.Entries = entries
}
