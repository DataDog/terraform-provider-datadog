package fwprovider

import (
	"context"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithModifyPlan = &costBudgetResource{}
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
	Entries      types.List    `tfsdk:"entries"`     // Deprecated: use BudgetLine
	BudgetLine   types.Set     `tfsdk:"budget_line"` // New grouped schema (unordered)
}

type budgetEntry struct {
	Amount     types.Float64 `tfsdk:"amount"`
	Month      types.Int64   `tfsdk:"month"`
	TagFilters types.List    `tfsdk:"tag_filters"`
}

type tagFilter struct {
	TagKey   types.String `tfsdk:"tag_key"`
	TagValue types.String `tfsdk:"tag_value"`
}

// New structs for budget_line (grouped schema)
type budgetLine struct {
	Amounts          types.Map  `tfsdk:"amounts"`            // map[month]amount
	TagFilters       types.List `tfsdk:"tag_filters"`        // For non-hierarchical budgets
	ParentTagFilters types.List `tfsdk:"parent_tag_filters"` // For hierarchical budgets (parent tag)
	ChildTagFilters  types.List `tfsdk:"child_tag_filters"`  // For hierarchical budgets (child tag)
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
				DeprecationMessage: "Use budget_line instead. This field will be removed in a future version.",
				Description:        "The entries of the budget. **Note:** You must provide entries for all months in the budget period. For hierarchical budgets, each unique tag combination must have entries for all months.",
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
			"budget_line": schema.SetNestedBlock{
				Description: "Budget lines that group monthly amounts by tag combination. Use this instead of `entries` for a more convenient schema. **Note:** The order of budget_line blocks does not matter.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"amounts": schema.MapAttribute{
							Required:    true,
							ElementType: types.Float64Type,
							Description: "Map of month (YYYYMM) to budget amount. Example: {\"202601\": 1000.0, \"202602\": 1200.0}",
						},
					},
					Blocks: map[string]schema.Block{
						"tag_filters": schema.ListNestedBlock{
							Description: "Tag filters for non-hierarchical budgets. **Note:** Cannot be used with parent_tag_filters/child_tag_filters.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key": schema.StringAttribute{
										Required:    true,
										Description: "Must be one of the tags from the `metrics_query`.",
									},
									"tag_value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"parent_tag_filters": schema.ListNestedBlock{
							Description: "Parent tag filters for hierarchical budgets. **Note:** Must be used with child_tag_filters. Cannot be used with tag_filters.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key": schema.StringAttribute{
										Required:    true,
										Description: "Must be one of the tags from the `metrics_query`.",
									},
									"tag_value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
						"child_tag_filters": schema.ListNestedBlock{
							Description: "Child tag filters for hierarchical budgets. **Note:** Must be used with parent_tag_filters. Cannot be used with tag_filters.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key": schema.StringAttribute{
										Required:    true,
										Description: "Must be one of the tags from the `metrics_query`.",
									},
									"tag_value": schema.StringAttribute{
										Required: true,
									},
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

	apiReq := buildBudgetWithEntriesFromModel(ctx, plan)
	apiResp, response, err := r.Api.UpsertBudget(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating budget", utils.TranslateClientError(err, response, "").Error())
		return
	}

	setModelFromBudgetWithEntries(ctx, &plan, apiResp)

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

	setModelFromBudgetWithEntries(ctx, &state, apiResp)

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

	apiReq := buildBudgetWithEntriesFromModel(ctx, plan)
	apiResp, response, err := r.Api.UpsertBudget(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating budget", utils.TranslateClientError(err, response, "").Error())
		return
	}

	setModelFromBudgetWithEntries(ctx, &plan, apiResp)

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

// ModifyPlan validates the budget by calling the backend /validate API endpoint
// This ensures validation errors are caught during terraform plan
func (r *costBudgetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {

	// Skip validation for resource destroy
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan costBudgetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure entries and budget_line are mutually exclusive
	hasEntries := !plan.Entries.IsNull() && !plan.Entries.IsUnknown()
	hasBudgetLine := !plan.BudgetLine.IsNull() && !plan.BudgetLine.IsUnknown()

	if hasEntries && hasBudgetLine {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot use both 'entries' and 'budget_line' simultaneously. Please use 'budget_line' (entries is deprecated).",
		)
		return
	}

	if !hasEntries && !hasBudgetLine {
		resp.Diagnostics.AddError(
			"Missing Configuration",
			"Either 'entries' or 'budget_line' must be specified.",
		)
		return
	}

	// Skip validation if required fields are unknown
	if plan.MetricsQuery.IsUnknown() || plan.StartMonth.IsUnknown() || plan.EndMonth.IsUnknown() {
		return
	}

	// Also skip if the schema fields are unknown
	if (hasEntries && plan.Entries.IsUnknown()) || (hasBudgetLine && plan.BudgetLine.IsUnknown()) {
		return
	}

	// Build the budget request from the plan
	budgetWithEntries := buildBudgetWithEntriesFromModel(ctx, plan)

	// Convert BudgetWithEntries to BudgetValidationRequest for the /validate endpoint
	// BudgetValidationRequestData uses BudgetWithEntriesDataAttributes, so we need to convert
	validationDataAttrs := datadogV2.BudgetWithEntriesDataAttributes{
		Name:         budgetWithEntries.Data.Attributes.Name,
		MetricsQuery: budgetWithEntries.Data.Attributes.MetricsQuery,
		StartMonth:   budgetWithEntries.Data.Attributes.StartMonth,
		EndMonth:     budgetWithEntries.Data.Attributes.EndMonth,
		Entries:      budgetWithEntries.Data.Attributes.Entries,
	}

	budgetTypeEnum := datadogV2.BUDGETWITHENTRIESDATATYPE_BUDGET
	validationRequest := datadogV2.BudgetValidationRequest{
		Data: &datadogV2.BudgetValidationRequestData{
			Attributes: &validationDataAttrs,
			Id:         budgetWithEntries.Data.Id,
			Type:       budgetTypeEnum,
		},
	}

	// Call the /validate API endpoint to catch errors during terraform plan
	_, _, err := r.Api.ValidateBudget(r.Auth, validationRequest)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating budget"))
		return
	}
}

// --- Helper functions ---

// tagFilterAttrTypes returns the attribute type definition for tagFilter
// This is used for converting between []tagFilter and types.List
func tagFilterAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tag_key":   types.StringType,
		"tag_value": types.StringType,
	}
}

// budgetEntryAttrTypes returns the attribute type definition for budgetEntry
// This is used for converting between []budgetEntry and types.List
func budgetEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"amount": types.Float64Type,
		"month":  types.Int64Type,
		"tag_filters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: tagFilterAttrTypes(),
			},
		},
	}
}

func budgetLineAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"amounts": types.MapType{ElemType: types.Float64Type},
		"tag_filters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: tagFilterAttrTypes(),
			},
		},
		"parent_tag_filters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: tagFilterAttrTypes(),
			},
		},
		"child_tag_filters": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: tagFilterAttrTypes(),
			},
		},
	}
}

// --- Helper functions to map between model and API types go here ---

// convertBudgetLineToFlatEntries converts budget_line (grouped schema) to flat API entries
func convertBudgetLineToFlatEntries(ctx context.Context, budgetLines []budgetLine) []budgetEntry {
	var flatEntries []budgetEntry

	for _, line := range budgetLines {
		// Extract the amounts map
		amounts := make(map[string]float64)
		line.Amounts.ElementsAs(ctx, &amounts, false)

		// Extract tag filters (for non-hierarchical budgets)
		var tagFilters []tagFilter
		if !line.TagFilters.IsNull() && !line.TagFilters.IsUnknown() {
			line.TagFilters.ElementsAs(ctx, &tagFilters, false)
		}

		// Extract parent and child tag filters (for hierarchical budgets)
		var parentTagFilters []tagFilter
		if !line.ParentTagFilters.IsNull() && !line.ParentTagFilters.IsUnknown() {
			line.ParentTagFilters.ElementsAs(ctx, &parentTagFilters, false)
		}

		var childTagFilters []tagFilter
		if !line.ChildTagFilters.IsNull() && !line.ChildTagFilters.IsUnknown() {
			line.ChildTagFilters.ElementsAs(ctx, &childTagFilters, false)
		}

		// Combine all tag filters
		var allTagFilters []tagFilter
		allTagFilters = append(allTagFilters, tagFilters...)
		allTagFilters = append(allTagFilters, parentTagFilters...)
		allTagFilters = append(allTagFilters, childTagFilters...)

		// Create an entry for each month in the amounts map
		for monthStr, amount := range amounts {
			// Convert month string to int64
			month, err := strconv.ParseInt(monthStr, 10, 64)
			if err != nil {
				continue // Skip invalid months
			}

			// Convert tag filters to types.List
			tagFiltersList, _ := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: tagFilterAttrTypes(),
			}, allTagFilters)

			flatEntries = append(flatEntries, budgetEntry{
				Month:      types.Int64Value(month),
				Amount:     types.Float64Value(amount),
				TagFilters: tagFiltersList,
			})
		}
	}

	return flatEntries
}

// convertFlatEntriesToBudgetLine converts flat API entries to budget_line (grouped schema)
func convertFlatEntriesToBudgetLine(ctx context.Context, flatEntries []budgetEntry) []budgetLine {
	type tagGroup struct {
		tags   []tagFilter
		months map[string]float64
	}

	groups := make(map[string]*tagGroup)

	for _, entry := range flatEntries {
		var tags []tagFilter
		if !entry.TagFilters.IsNull() && !entry.TagFilters.IsUnknown() {
			entry.TagFilters.ElementsAs(ctx, &tags, false)
		}

		key := tagSignature(tags)
		if groups[key] == nil {
			groups[key] = &tagGroup{tags: tags, months: make(map[string]float64)}
		}
		groups[key].months[strconv.FormatInt(entry.Month.ValueInt64(), 10)] = entry.Amount.ValueFloat64()
	}

	budgetLines := make([]budgetLine, 0, len(groups))
	tagObjType := types.ObjectType{AttrTypes: tagFilterAttrTypes()}

	for _, group := range groups {
		amountsMap, _ := types.MapValueFrom(ctx, types.Float64Type, group.months)
		line := budgetLine{Amounts: amountsMap}

		if len(group.tags) == 2 {
			line.ParentTagFilters, _ = types.ListValueFrom(ctx, tagObjType, []tagFilter{group.tags[0]})
			line.ChildTagFilters, _ = types.ListValueFrom(ctx, tagObjType, []tagFilter{group.tags[1]})
			line.TagFilters = types.ListNull(tagObjType)
		} else {
			line.ParentTagFilters = types.ListNull(tagObjType)
			line.ChildTagFilters = types.ListNull(tagObjType)
			if len(group.tags) > 0 {
				line.TagFilters, _ = types.ListValueFrom(ctx, tagObjType, group.tags)
			} else {
				line.TagFilters = types.ListNull(tagObjType)
			}
		}

		budgetLines = append(budgetLines, line)
	}

	return budgetLines
}

// tagSignature creates a unique identifier for grouping entries by tag combination
func tagSignature(tags []tagFilter) string {
	if len(tags) == 0 {
		return ""
	}
	var parts []string
	for _, t := range tags {
		parts = append(parts, t.TagKey.ValueString()+"="+t.TagValue.ValueString())
	}
	return strings.Join(parts, ",")
}

func buildBudgetWithEntriesFromModel(ctx context.Context, plan costBudgetModel) datadogV2.BudgetWithEntries {
	var planEntries []budgetEntry

	// Check if budget_line is used (new schema)
	if !plan.BudgetLine.IsNull() && !plan.BudgetLine.IsUnknown() {
		// Convert budget_line to flat entries
		var budgetLines []budgetLine
		plan.BudgetLine.ElementsAs(ctx, &budgetLines, false)
		planEntries = convertBudgetLineToFlatEntries(ctx, budgetLines)
	} else {
		// Use legacy entries schema
		plan.Entries.ElementsAs(ctx, &planEntries, false)
	}

	// Convert entries to API format
	var entries []datadogV2.BudgetWithEntriesDataAttributesEntriesItems
	for _, e := range planEntries {
		// Convert tag_filters from types.List to []tagFilter
		var entryTagFilters []tagFilter
		e.TagFilters.ElementsAs(ctx, &entryTagFilters, false)

		var tagFilters []datadogV2.BudgetWithEntriesDataAttributesEntriesItemsTagFiltersItems
		for _, tf := range entryTagFilters {
			tagFilters = append(tagFilters, datadogV2.BudgetWithEntriesDataAttributesEntriesItemsTagFiltersItems{
				TagKey:   tf.TagKey.ValueStringPointer(),
				TagValue: tf.TagValue.ValueStringPointer(),
			})
		}
		entries = append(entries, datadogV2.BudgetWithEntriesDataAttributesEntriesItems{
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

func setModelFromBudgetWithEntries(ctx context.Context, model *costBudgetModel, apiResp datadogV2.BudgetWithEntries) {
	if apiResp.Data == nil || apiResp.Data.Attributes == nil {
		return
	}
	data, attr := apiResp.Data, apiResp.Data.Attributes

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

	entries := apiEntriesToBudgetEntries(ctx, attr.Entries)
	usingBudgetLine := !model.BudgetLine.IsNull() && !model.BudgetLine.IsUnknown()

	if usingBudgetLine {
		model.BudgetLine, _ = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: budgetLineAttrTypes()}, convertFlatEntriesToBudgetLine(ctx, entries))
		model.Entries = types.ListNull(types.ObjectType{AttrTypes: budgetEntryAttrTypes()})
	} else {
		model.Entries, _ = types.ListValueFrom(ctx, types.ObjectType{AttrTypes: budgetEntryAttrTypes()}, entries)
		model.BudgetLine = types.SetNull(types.ObjectType{AttrTypes: budgetLineAttrTypes()})
	}
}

// apiEntriesToBudgetEntries converts API entries to internal budgetEntry model
func apiEntriesToBudgetEntries(ctx context.Context, apiEntries []datadogV2.BudgetWithEntriesDataAttributesEntriesItems) []budgetEntry {
	entries := make([]budgetEntry, 0, len(apiEntries))
	tagObjType := types.ObjectType{AttrTypes: tagFilterAttrTypes()}

	for _, apiEntry := range apiEntries {
		tagFilters := make([]tagFilter, 0, len(apiEntry.TagFilters))
		for _, tf := range apiEntry.TagFilters {
			tagFilters = append(tagFilters, tagFilter{
				TagKey:   types.StringPointerValue(tf.TagKey),
				TagValue: types.StringPointerValue(tf.TagValue),
			})
		}

		tagFiltersList, _ := types.ListValueFrom(ctx, tagObjType, tagFilters)
		entries = append(entries, budgetEntry{
			Amount:     types.Float64PointerValue(apiEntry.Amount),
			Month:      types.Int64PointerValue(apiEntry.Month),
			TagFilters: tagFiltersList,
		})
	}
	return entries
}
