package fwprovider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithValidateConfig = &costCustomForecastResource{}
)

type costCustomForecastResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func NewCostCustomForecastResource() resource.Resource {
	return &costCustomForecastResource{}
}

type costCustomForecastModel struct {
	ID        types.String `tfsdk:"id"`
	BudgetUid types.String `tfsdk:"budget_uid"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
	CreatedBy types.String `tfsdk:"created_by"`
	UpdatedAt types.Int64  `tfsdk:"updated_at"`
	UpdatedBy types.String `tfsdk:"updated_by"`
	Entries   types.Set    `tfsdk:"entries"`
}

type customForecastEntryModel struct {
	Month      types.Int64   `tfsdk:"month"`
	Amount     types.Float64 `tfsdk:"amount"`
	TagFilters types.List    `tfsdk:"tag_filters"`
}

type customForecastTagFilterModel struct {
	TagKey   types.String `tfsdk:"tag_key"`
	TagValue types.String `tfsdk:"tag_value"`
}

func customForecastTagFilterAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tag_key":   types.StringType,
		"tag_value": types.StringType,
	}
}

func customForecastEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"month":  types.Int64Type,
		"amount": types.Float64Type,
		"tag_filters": types.ListType{
			ElemType: types.ObjectType{AttrTypes: customForecastTagFilterAttrTypes()},
		},
	}
}

func (r *costCustomForecastResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "cost_custom_forecast"
}

func (r *costCustomForecastResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Cost Custom Forecast resource. This resource manages the custom forecast override entries for a `datadog_cost_budget`. **Note:** each entry's `(month, tag_filters)` combination must correspond to an existing entry on the referenced budget, and the budget must exist before this resource is created.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the custom forecast set.",
			},
			"budget_uid": schema.StringAttribute{
				Required:      true,
				Description:   "The UUID of the budget that this custom forecast belongs to. Changing this value forces a new resource to be created.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Timestamp the custom forecast was created, in Unix milliseconds.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the user that created the custom forecast.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Timestamp the custom forecast was last updated, in Unix milliseconds.",
			},
			"updated_by": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the user that last updated the custom forecast.",
			},
		},
		Blocks: map[string]schema.Block{
			"entries": schema.SetNestedBlock{
				Description: "Monthly custom forecast entries. Each entry overrides the forecast for one `(month, tag_filters)` combination that must already exist as a budget entry. To remove all custom forecast entries, destroy this resource rather than setting an empty `entries` set.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"month": schema.Int64Attribute{
							Required:    true,
							Description: "The month the entry applies to, in `YYYYMM` format.",
						},
						"amount": schema.Float64Attribute{
							Required:    true,
							Description: "The forecast override amount for the month.",
							Validators: []validator.Float64{
								float64validator.AtLeast(0),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"tag_filters": schema.ListNestedBlock{
							Description: "Tag filters that scope this entry to a specific budget entry tag combination.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"tag_key":   schema.StringAttribute{Required: true},
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

func (r *costCustomForecastResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

// ValidateConfig catches duplicate (month, tag_filters) entries at `terraform plan`, before any API call.
// Per-entry amount>=0 is enforced by the amount attribute's float64validator.
func (r *costCustomForecastResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config costCustomForecastModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Entries.IsNull() || config.Entries.IsUnknown() {
		return
	}

	var entries []customForecastEntryModel
	if diags := config.Entries.ElementsAs(ctx, &entries, false); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	seen := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if e.Month.IsUnknown() {
			continue
		}

		var tagFilters []customForecastTagFilterModel
		if !e.TagFilters.IsNull() && !e.TagFilters.IsUnknown() {
			if diags := e.TagFilters.ElementsAs(ctx, &tagFilters, false); diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
		}

		key := customForecastEntryKey(e.Month.ValueInt64(), tagFilters)
		if _, ok := seen[key]; ok {
			resp.Diagnostics.AddAttributeError(
				frameworkPath.Root("entries"),
				"Duplicate custom forecast entry",
				"more than one entry targets the same month and tag_filters combination",
			)
			return
		}
		seen[key] = struct{}{}
	}
}

// customForecastEntryKey builds a deterministic key from (month, tag_filters), sorting the tag
// k=v pairs so callers can match regardless of block order. Mirrors the dd-source validation key.
func customForecastEntryKey(month int64, tagFilters []customForecastTagFilterModel) string {
	parts := make([]string, 0, len(tagFilters))
	for _, tf := range tagFilters {
		parts = append(parts, tf.TagKey.ValueString()+"="+tf.TagValue.ValueString())
	}
	sort.Strings(parts)
	return fmt.Sprintf("%d\x00%s", month, strings.Join(parts, "\x00"))
}

// --- CRUD ---

func (r *costCustomForecastResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan costCustomForecastModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildCustomForecastUpsertRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, httpResp, err := r.Api.UpsertCustomForecast(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating custom forecast", utils.TranslateClientError(err, httpResp, "").Error())
		return
	}

	resp.Diagnostics.Append(setCostCustomForecastModelFromResponse(ctx, &plan, apiResp)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *costCustomForecastResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state costCustomForecastModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, httpResp, err := r.Api.GetCustomForecast(r.Auth, state.BudgetUid.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading custom forecast", utils.TranslateClientError(err, httpResp, "").Error())
		return
	}

	resp.Diagnostics.Append(setCostCustomForecastModelFromResponse(ctx, &state, apiResp)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *costCustomForecastResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan costCustomForecastModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq, diags := buildCustomForecastUpsertRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, httpResp, err := r.Api.UpsertCustomForecast(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating custom forecast", utils.TranslateClientError(err, httpResp, "").Error())
		return
	}

	resp.Diagnostics.Append(setCostCustomForecastModelFromResponse(ctx, &plan, apiResp)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *costCustomForecastResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state costCustomForecastModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteCustomForecast(r.Auth, state.BudgetUid.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Already gone - e.g. the parent budget was deleted, which cascades.
			return
		}
		resp.Diagnostics.AddError("Error deleting custom forecast", utils.TranslateClientError(err, httpResp, "").Error())
		return
	}
}

// ImportState imports by budget_uid rather than the computed custom-forecast-set id,
// since budget_uid is the resource's effective primary key.
func (r *costCustomForecastResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("budget_uid"), req, resp)
}

// --- helpers ---

func buildCustomForecastUpsertRequest(ctx context.Context, plan costCustomForecastModel) (datadogV2.CustomForecastUpsertRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var planEntries []customForecastEntryModel
	if !plan.Entries.IsNull() && !plan.Entries.IsUnknown() {
		diags.Append(plan.Entries.ElementsAs(ctx, &planEntries, false)...)
	}

	entries := make([]datadogV2.CustomForecastEntry, 0, len(planEntries))
	for _, e := range planEntries {
		var tagFilterModels []customForecastTagFilterModel
		if !e.TagFilters.IsNull() && !e.TagFilters.IsUnknown() {
			diags.Append(e.TagFilters.ElementsAs(ctx, &tagFilterModels, false)...)
		}

		tagFilters := make([]datadogV2.CustomForecastEntryTagFilter, 0, len(tagFilterModels))
		for _, tf := range tagFilterModels {
			tagFilters = append(tagFilters, datadogV2.CustomForecastEntryTagFilter{
				TagKey:   tf.TagKey.ValueString(),
				TagValue: tf.TagValue.ValueString(),
			})
		}

		entries = append(entries, datadogV2.CustomForecastEntry{
			Month:      e.Month.ValueInt64(),
			Amount:     e.Amount.ValueFloat64(),
			TagFilters: tagFilters,
		})
	}

	attributes := datadogV2.NewCustomForecastUpsertRequestDataAttributes(plan.BudgetUid.ValueString(), entries)
	data := datadogV2.NewCustomForecastUpsertRequestData(*attributes, datadogV2.CUSTOMFORECASTTYPE_CUSTOM_FORECAST)
	return *datadogV2.NewCustomForecastUpsertRequest(*data), diags
}

func setCostCustomForecastModelFromResponse(ctx context.Context, model *costCustomForecastModel, apiResp datadogV2.CustomForecastResponse) diag.Diagnostics {
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
