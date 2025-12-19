package fwprovider

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_                resource.ResourceWithValidateConfig = &costBudgetResource{}
	metricQueryRegex                                     = regexp.MustCompile(`by\s*\{(.+)\}`)
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

// ValidateConfig performs client-side validation during terraform plan
// Note: This duplicates the API's validation logic in BudgetWithEntries.validate() in dd-source
// Will be replaced by API-based validation (dry-run or /validate endpoint) in a future release
func (r *costBudgetResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data costBudgetModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.MetricsQuery.IsUnknown() || data.StartMonth.IsUnknown() || data.EndMonth.IsUnknown() {
		return
	}

	// Extract tags from metrics_query
	tags := extractTagsFromQuery(data.MetricsQuery.ValueString())

	// Validate tags length
	if len(tags) > 2 {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("metrics_query"),
			"Invalid metrics_query",
			"tags must have 0, 1 or 2 elements",
		)
	}

	// Validate tags are unique
	if len(tags) == 2 && tags[0] == tags[1] {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("metrics_query"),
			"Invalid metrics_query",
			"tags must be unique",
		)
	}

	startMonth := data.StartMonth.ValueInt64()
	endMonth := data.EndMonth.ValueInt64()

	// Validate start_month
	if startMonth <= 0 {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("start_month"),
			"Invalid start_month",
			"start_month must be greater than 0 and of the format YYYYMM",
		)
	}

	// Validate end_month
	if endMonth <= 0 {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("end_month"),
			"Invalid end_month",
			"end_month must be greater than 0 and of the format YYYYMM",
		)
	}

	// Validate end_month >= start_month
	if startMonth > endMonth {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("end_month"),
			"Invalid end_month",
			"end_month must be greater than or equal to start_month",
		)
	}

	// Track which months exist for each unique tag combination
	// Example: {"ASE\tstaging": {202501: true, 202502: true}} means team=ASE,account=staging has entries for 202501 & 202502
	entriesMap := make(map[string]map[int64]bool)

	// Validate entries
	for i, entry := range data.Entries {
		month := entry.Month.ValueInt64()
		amount := entry.Amount.ValueFloat64()

		// Validate entry month in range
		if month < startMonth || month > endMonth {
			resp.Diagnostics.AddAttributeError(
				frameworkPath.Root("entries").AtListIndex(i).AtName("month"),
				"Invalid month",
				"entry month must be between start_month and end_month",
			)
		}

		// Validate entry amount >= 0
		if amount < 0 {
			resp.Diagnostics.AddAttributeError(
				frameworkPath.Root("entries").AtListIndex(i).AtName("amount"),
				"Invalid amount",
				"entry amount must be greater than or equal to 0",
			)
		}

		// Validate tag_filters count
		if len(entry.TagFilters) != len(tags) {
			resp.Diagnostics.AddAttributeError(
				frameworkPath.Root("entries").AtListIndex(i).AtName("tag_filters"),
				"Invalid tag_filters",
				"entry tag_filters must include all group by tags",
			)
			continue
		}

		// Validate tag_key and collect tag values
		tagValues := make([]string, len(entry.TagFilters))
		for j, tf := range entry.TagFilters {
			tagKey := tf.TagKey.ValueString()

			if !slices.Contains(tags, tagKey) {
				resp.Diagnostics.AddAttributeError(
					frameworkPath.Root("entries").AtListIndex(i).AtName("tag_filters").AtListIndex(j).AtName("tag_key"),
					"Invalid tag_key",
					"tag_key must be one of the values inside the tags array",
				)
			}

			tagValues[j] = tf.TagValue.ValueString()
		}

		// Build unique key for this tag combination (e.g., "ASE\tstaging")
		// We sort to ensure same combination regardless of order: {team:ASE,account:staging} = {account:staging,team:ASE}
		sort.Strings(tagValues)
		tagCombination := strings.Join(tagValues, "\t")
		if entriesMap[tagCombination] == nil {
			entriesMap[tagCombination] = make(map[int64]bool)
		}
		entriesMap[tagCombination][month] = true
	}

	// Validate entries exist
	if len(entriesMap) == 0 {
		resp.Diagnostics.AddAttributeError(
			frameworkPath.Root("entries"),
			"Missing entries",
			"entries are required",
		)
		return
	}

	// Validate all tag combinations have entries for all months
	expectedMonthCount := calculateMonthCount(startMonth, endMonth)
	for tagCombination, months := range entriesMap {
		if len(months) != expectedMonthCount {
			resp.Diagnostics.AddError(
				"Missing entries for tag combination",
				fmt.Sprintf("missing entries for tag value pair: %v", tagCombination),
			)
		}
	}
}

// --- Validation helper functions ---

// extractTagsFromQuery extracts tags from "by {tag1,tag2}" in metrics_query
// Copied from dd-source: domains/cloud_cost_management/libs/costplanningdb/tables.go
func extractTagsFromQuery(query string) []string {
	subGroups := metricQueryRegex.FindStringSubmatch(query)
	if len(subGroups) != 2 {
		return []string{}
	}
	tags := strings.Split(subGroups[1], ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}

// calculateMonthCount returns the number of months between start and end (inclusive)
// Copied from dd-source: domains/cloud_cost_management/libs/costplanningdb/tables.go (GetBudgetDuration)
func calculateMonthCount(start, end int64) int {
	startYear := start / 100
	endYear := end / 100
	startMonth := start % 100
	endMonth := end % 100
	return int((endYear-startYear)*12 + endMonth - startMonth + 1)
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
