package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &tagIndexingRuleResource{}
	_ resource.ResourceWithImportState = &tagIndexingRuleResource{}
)

type tagIndexingRuleResource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

type tagIndexingRuleModel struct {
	ID                       types.String                 `tfsdk:"id"`
	Name                     types.String                 `tfsdk:"name"`
	MetricNameMatches        types.List                   `tfsdk:"metric_name_matches"`
	IgnoredMetricNameMatches types.List                   `tfsdk:"ignored_metric_name_matches"`
	Tags                     types.List                   `tfsdk:"tags"`
	ExcludeTagsMode          types.Bool                   `tfsdk:"exclude_tags_mode"`
	Options                  *tagIndexingRuleOptionsModel `tfsdk:"options"`
	RuleOrder                types.Int64                  `tfsdk:"rule_order"`
	CreatedAt                types.String                 `tfsdk:"created_at"`
	ModifiedAt               types.String                 `tfsdk:"modified_at"`
	CreatedByHandle          types.String                 `tfsdk:"created_by_handle"`
	ModifiedByHandle         types.String                 `tfsdk:"modified_by_handle"`
}

type tagIndexingRuleOptionsModel struct {
	Version types.Int64                      `tfsdk:"version"`
	Data    *tagIndexingRuleOptionsDataModel `tfsdk:"data"`
}

type tagIndexingRuleOptionsDataModel struct {
	OverridePreviousRules    types.Bool                       `tfsdk:"override_previous_rules"`
	ManagePreexistingMetrics types.Bool                       `tfsdk:"manage_preexisting_metrics"`
	DynamicTags              *tagIndexingRuleDynamicTagsModel `tfsdk:"dynamic_tags"`
	MetricMatch              *tagIndexingRuleMetricMatchModel `tfsdk:"metric_match"`
}

type tagIndexingRuleDynamicTagsModel struct {
	QueriedTagsWindowSeconds types.Int64 `tfsdk:"queried_tags_window_seconds"`
	RelatedAssetTags         types.Bool  `tfsdk:"related_asset_tags"`
}

type tagIndexingRuleMetricMatchModel struct {
	IsQueried            types.Bool  `tfsdk:"is_queried"`
	NotQueried           types.Bool  `tfsdk:"not_queried"`
	NotUsedInAssets      types.Bool  `tfsdk:"not_used_in_assets"`
	QueriedWindowSeconds types.Int64 `tfsdk:"queried_window_seconds"`
	UsedInAssets         types.Bool  `tfsdk:"used_in_assets"`
}

func NewTagIndexingRuleResource() resource.Resource {
	return &tagIndexingRuleResource{}
}

func (r *tagIndexingRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	r.Auth = providerData.Auth
}

func (r *tagIndexingRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "tag_indexing_rule"
}

func (r *tagIndexingRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Tag Indexing Rule resource. Tag indexing rules control which tag keys are indexed for metrics, reducing cardinality costs while preserving queryability.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the rule.",
			},
			"metric_name_matches": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Metric name prefixes (glob patterns) this rule applies to.",
			},
			"ignored_metric_name_matches": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Metric name prefixes excluded from the rule's scope.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tag keys managed by this rule.",
			},
			"exclude_tags_mode": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, the rule excludes the listed tags and indexes all others. When false (default), the rule includes only the listed tags.",
			},
			"rule_order": schema.Int64Attribute{
				Computed:    true,
				Description: "Evaluation order within the org. Lower values are evaluated first. Server-assigned on create; use `datadog_tag_indexing_rule_order` to control ordering.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the rule was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the rule was last modified.",
			},
			"created_by_handle": schema.StringAttribute{
				Computed:    true,
				Description: "Handle of the user who created the rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified_by_handle": schema.StringAttribute{
				Computed:    true,
				Description: "Handle of the user who last modified the rule.",
			},
			"options": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Versioned configuration options for the rule.",
				Attributes: map[string]schema.Attribute{
					"version": schema.Int64Attribute{
						Required:    true,
						Description: "Options schema version. Only `1` is supported.",
					},
					"data": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Options data payload.",
						Attributes: map[string]schema.Attribute{
							"override_previous_rules": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(false),
								Description: "When true, this rule's tag list overrides tags configured by earlier rules for the same metric.",
							},
							"manage_preexisting_metrics": schema.BoolAttribute{
								Optional:    true,
								Computed:    true,
								Default:     booldefault.StaticBool(true),
								Description: "When true, the rule applies to metrics ingested before the rule was created.",
							},
							"dynamic_tags": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Configuration for including dynamically queried tags.",
								Attributes: map[string]schema.Attribute{
									"queried_tags_window_seconds": schema.Int64Attribute{
										Optional:    true,
										Description: "Window in seconds for evaluating queried tags.",
									},
									"related_asset_tags": schema.BoolAttribute{
										Optional:    true,
										Description: "When true, tags from related assets are included.",
									},
								},
							},
							"metric_match": schema.SingleNestedAttribute{
								Optional:    true,
								Description: "Criteria for matching metrics based on query state.",
								Attributes: map[string]schema.Attribute{
									"is_queried": schema.BoolAttribute{
										Optional:    true,
										Description: "Match metrics that are being queried.",
									},
									"not_queried": schema.BoolAttribute{
										Optional:    true,
										Description: "Match metrics that are not being queried.",
									},
									"not_used_in_assets": schema.BoolAttribute{
										Optional:    true,
										Description: "Match metrics not used in any dashboards or monitors.",
									},
									"queried_window_seconds": schema.Int64Attribute{
										Optional:    true,
										Description: "Window in seconds for evaluating query state.",
									},
									"used_in_assets": schema.BoolAttribute{
										Optional:    true,
										Description: "Match metrics used in dashboards or monitors.",
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

func (r *tagIndexingRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *tagIndexingRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tagIndexingRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetTagIndexingRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving tag indexing rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagIndexingRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state tagIndexingRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildCreateRequest(ctx, &state)

	resp, _, err := r.Api.CreateTagIndexingRule(r.Auth, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating tag indexing rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagIndexingRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state tagIndexingRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	body := r.buildUpdateRequest(ctx, &state)

	resp, _, err := r.Api.UpdateTagIndexingRule(r.Auth, id, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating tag indexing rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagIndexingRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state tagIndexingRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	httpResp, err := r.Api.DeleteTagIndexingRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return // idempotent
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting tag indexing rule"))
	}
}

func (r *tagIndexingRuleResource) updateState(_ context.Context, state *tagIndexingRuleModel, resp *datadogV2.TagIndexingRuleResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attrs := data.GetAttributes()
	state.Name = types.StringValue(attrs.GetName())
	state.RuleOrder = types.Int64Value(int64(attrs.GetRuleOrder()))
	state.ExcludeTagsMode = types.BoolValue(attrs.GetExcludeTagsMode())

	matches := make([]types.String, 0, len(attrs.GetMetricNameMatches()))
	for _, m := range attrs.GetMetricNameMatches() {
		matches = append(matches, types.StringValue(m))
	}
	state.MetricNameMatches, _ = types.ListValueFrom(context.Background(), types.StringType, matches)

	ignored := make([]types.String, 0, len(attrs.GetIgnoredMetricNameMatches()))
	for _, m := range attrs.GetIgnoredMetricNameMatches() {
		ignored = append(ignored, types.StringValue(m))
	}
	state.IgnoredMetricNameMatches, _ = types.ListValueFrom(context.Background(), types.StringType, ignored)

	tags := make([]types.String, 0, len(attrs.GetTags()))
	for _, t := range attrs.GetTags() {
		tags = append(tags, types.StringValue(t))
	}
	state.Tags, _ = types.ListValueFrom(context.Background(), types.StringType, tags)

	if opts, ok := attrs.GetOptionsOk(); ok && opts != nil {
		optModel := &tagIndexingRuleOptionsModel{
			Version: types.Int64Value(int64(opts.GetVersion())),
		}
		if d, ok := opts.GetDataOk(); ok && d != nil {
			dataModel := &tagIndexingRuleOptionsDataModel{
				OverridePreviousRules:    types.BoolValue(d.GetOverridePreviousRules()),
				ManagePreexistingMetrics: types.BoolValue(d.GetManagePreexistingMetrics()),
			}
			if dt, ok := d.GetDynamicTagsOk(); ok && dt != nil {
				dataModel.DynamicTags = &tagIndexingRuleDynamicTagsModel{
					QueriedTagsWindowSeconds: types.Int64Value(int64(dt.GetQueriedTagsWindowSeconds())),
					RelatedAssetTags:         types.BoolValue(dt.GetRelatedAssetTags()),
				}
			}
			if mm, ok := d.GetMetricMatchOk(); ok && mm != nil {
				dataModel.MetricMatch = &tagIndexingRuleMetricMatchModel{
					IsQueried:            types.BoolValue(mm.GetIsQueried()),
					NotQueried:           types.BoolValue(mm.GetNotQueried()),
					NotUsedInAssets:      types.BoolValue(mm.GetNotUsedInAssets()),
					QueriedWindowSeconds: types.Int64Value(int64(mm.GetQueriedWindowSeconds())),
					UsedInAssets:         types.BoolValue(mm.GetUsedInAssets()),
				}
			}
			optModel.Data = dataModel
		}
		state.Options = optModel
	}

	if v, ok := attrs.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(v.String())
	}
	if v, ok := attrs.GetModifiedAtOk(); ok {
		state.ModifiedAt = types.StringValue(v.String())
	}
	if v, ok := attrs.GetCreatedByHandleOk(); ok && v != nil {
		state.CreatedByHandle = types.StringValue(*v)
	}
	if v, ok := attrs.GetModifiedByHandleOk(); ok && v != nil {
		state.ModifiedByHandle = types.StringValue(*v)
	}
}

func (r *tagIndexingRuleResource) buildCreateRequest(_ context.Context, state *tagIndexingRuleModel) datadogV2.TagIndexingRuleCreateRequest {
	attrs := datadogV2.NewTagIndexingRuleCreateAttributesWithDefaults()
	attrs.SetName(state.Name.ValueString())

	var matches []string
	state.MetricNameMatches.ElementsAs(context.Background(), &matches, false)
	attrs.SetMetricNameMatches(matches)

	if !state.ExcludeTagsMode.IsNull() && !state.ExcludeTagsMode.IsUnknown() {
		attrs.SetExcludeTagsMode(state.ExcludeTagsMode.ValueBool())
	}

	var ignored []string
	if !state.IgnoredMetricNameMatches.IsNull() {
		state.IgnoredMetricNameMatches.ElementsAs(context.Background(), &ignored, false)
		attrs.SetIgnoredMetricNameMatches(ignored)
	}

	var tags []string
	if !state.Tags.IsNull() {
		state.Tags.ElementsAs(context.Background(), &tags, false)
		attrs.SetTags(tags)
	}

	if state.Options != nil {
		opts := buildOptionsFromModel(state.Options)
		attrs.SetOptions(opts)
	}

	data := datadogV2.NewTagIndexingRuleCreateDataWithDefaults()
	data.SetAttributes(*attrs)

	req := datadogV2.NewTagIndexingRuleCreateRequestWithDefaults()
	req.SetData(*data)
	return *req
}

func (r *tagIndexingRuleResource) buildUpdateRequest(_ context.Context, state *tagIndexingRuleModel) datadogV2.TagIndexingRuleUpdateRequest {
	attrs := datadogV2.NewTagIndexingRuleUpdateAttributesWithDefaults()
	attrs.SetName(state.Name.ValueString())

	var matches []string
	state.MetricNameMatches.ElementsAs(context.Background(), &matches, false)
	attrs.SetMetricNameMatches(matches)

	if !state.ExcludeTagsMode.IsNull() && !state.ExcludeTagsMode.IsUnknown() {
		attrs.SetExcludeTagsMode(state.ExcludeTagsMode.ValueBool())
	}

	var ignored []string
	if !state.IgnoredMetricNameMatches.IsNull() {
		state.IgnoredMetricNameMatches.ElementsAs(context.Background(), &ignored, false)
		attrs.SetIgnoredMetricNameMatches(ignored)
	}

	var tags []string
	if !state.Tags.IsNull() {
		state.Tags.ElementsAs(context.Background(), &tags, false)
		attrs.SetTags(tags)
	}

	if state.Options != nil {
		opts := buildOptionsFromModel(state.Options)
		attrs.SetOptions(opts)
	}

	data := datadogV2.NewTagIndexingRuleUpdateDataWithDefaults()
	data.SetAttributes(*attrs)

	req := datadogV2.NewTagIndexingRuleUpdateRequestWithDefaults()
	req.SetData(*data)
	return *req
}

func buildOptionsFromModel(m *tagIndexingRuleOptionsModel) datadogV2.TagIndexingRuleOptions {
	opts := datadogV2.NewTagIndexingRuleOptionsWithDefaults()
	opts.SetVersion(m.Version.ValueInt64())

	if m.Data != nil {
		d := datadogV2.NewTagIndexingRuleOptionsDataWithDefaults()
		d.SetOverridePreviousRules(m.Data.OverridePreviousRules.ValueBool())
		d.SetManagePreexistingMetrics(m.Data.ManagePreexistingMetrics.ValueBool())

		if m.Data.DynamicTags != nil {
			dt := datadogV2.NewTagIndexingRuleDynamicTagsWithDefaults()
			if !m.Data.DynamicTags.QueriedTagsWindowSeconds.IsNull() {
				dt.SetQueriedTagsWindowSeconds(m.Data.DynamicTags.QueriedTagsWindowSeconds.ValueInt64())
			}
			if !m.Data.DynamicTags.RelatedAssetTags.IsNull() {
				dt.SetRelatedAssetTags(m.Data.DynamicTags.RelatedAssetTags.ValueBool())
			}
			d.SetDynamicTags(*dt)
		}

		if m.Data.MetricMatch != nil {
			mm := datadogV2.NewTagIndexingRuleMetricMatchWithDefaults()
			if !m.Data.MetricMatch.IsQueried.IsNull() {
				mm.SetIsQueried(m.Data.MetricMatch.IsQueried.ValueBool())
			}
			if !m.Data.MetricMatch.NotQueried.IsNull() {
				mm.SetNotQueried(m.Data.MetricMatch.NotQueried.ValueBool())
			}
			if !m.Data.MetricMatch.NotUsedInAssets.IsNull() {
				mm.SetNotUsedInAssets(m.Data.MetricMatch.NotUsedInAssets.ValueBool())
			}
			if !m.Data.MetricMatch.QueriedWindowSeconds.IsNull() {
				mm.SetQueriedWindowSeconds(m.Data.MetricMatch.QueriedWindowSeconds.ValueInt64())
			}
			if !m.Data.MetricMatch.UsedInAssets.IsNull() {
				mm.SetUsedInAssets(m.Data.MetricMatch.UsedInAssets.ValueBool())
			}
			d.SetMetricMatch(*mm)
		}

		opts.SetData(*d)
	}
	return *opts
}
