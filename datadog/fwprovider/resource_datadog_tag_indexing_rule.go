package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &tagIndexingRuleResource{}
	_ resource.ResourceWithImportState    = &tagIndexingRuleResource{}
	_ resource.ResourceWithValidateConfig = &tagIndexingRuleResource{}
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
}

type tagIndexingRuleDynamicTagsModel struct {
	ExcludeNotQueriedWindowSeconds types.Int64 `tfsdk:"exclude_not_queried_window_seconds"`
	ExcludeNotUsedInAssets         types.Bool  `tfsdk:"exclude_not_used_in_assets"`
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
				Description: "Tag keys this rule includes or excludes, depending on exclude_tags_mode.",
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
				Description: "Versioned configuration options for the rule.",
				Attributes: map[string]schema.Attribute{
					"version": schema.Int64Attribute{
						Required:    true,
						Description: "Options schema version. Only `1` is supported.",
					},
					"data": schema.SingleNestedAttribute{
						Required:    true,
						Description: "Behavioral options for how the rule applies to metrics, including backfill and override behavior.",
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
								Description: "Configuration for excluding tags based on dynamic usage signals. Only applies when `exclude_tags_mode` is `true`.",
								Attributes: map[string]schema.Attribute{
									"exclude_not_queried_window_seconds": schema.Int64Attribute{
										Optional:    true,
										Description: "Lookback window, in seconds, for excluding tags that were not queried in that period. Requires `exclude_tags_mode` to be `true`.",
										Validators: []validator.Int64{
											int64validator.Between(1, 7776000),
										},
									},
									"exclude_not_used_in_assets": schema.BoolAttribute{
										Optional:    true,
										Description: "When true, excludes tags not used in any dashboards or monitors. Requires `exclude_tags_mode` to be `true`.",
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

// ValidateConfig mirrors the backend's update-time check (apiv2handler.go:1312-1326) at plan time:
// the exclude_not_* usage fields on dynamic_tags only take effect (and are only accepted by the API)
// when exclude_tags_mode is true.
func (r *tagIndexingRuleResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config tagIndexingRuleModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	if config.Options == nil || config.Options.Data == nil || config.Options.Data.DynamicTags == nil {
		return
	}

	if config.ExcludeTagsMode.IsUnknown() {
		return
	}

	dt := config.Options.Data.DynamicTags
	usageFieldSet := (!dt.ExcludeNotQueriedWindowSeconds.IsNull() && !dt.ExcludeNotQueriedWindowSeconds.IsUnknown()) ||
		(!dt.ExcludeNotUsedInAssets.IsNull() && !dt.ExcludeNotUsedInAssets.IsUnknown())

	if usageFieldSet && !config.ExcludeTagsMode.ValueBool() {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("exclude_tags_mode"),
			"Invalid dynamic_tags usage configuration",
			"options.data.dynamic_tags.exclude_not_queried_window_seconds and exclude_not_used_in_assets require exclude_tags_mode to be true.",
		)
	}
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

	// Populate options if they were configured (state.Options != nil) OR if this is
	// an import read (CreatedAt not yet set, so prior state only has the ID).
	isImport := state.CreatedAt.IsNull()
	if opts, ok := attrs.GetOptionsOk(); ok && opts != nil && (state.Options != nil || isImport) {
		optModel := &tagIndexingRuleOptionsModel{
			Version: types.Int64Value(int64(opts.GetVersion())),
		}
		if d, ok := opts.GetDataOk(); ok && d != nil {
			dataModel := &tagIndexingRuleOptionsDataModel{
				OverridePreviousRules:    types.BoolValue(d.GetOverridePreviousRules()),
				ManagePreexistingMetrics: types.BoolValue(d.GetManagePreexistingMetrics()),
			}
			if dt, ok := d.GetDynamicTagsOk(); ok && dt != nil {
				dtModel := &tagIndexingRuleDynamicTagsModel{
					ExcludeNotQueriedWindowSeconds: types.Int64Null(),
					ExcludeNotUsedInAssets:         types.BoolNull(),
				}
				if v, ok := dt.GetExcludeNotQueriedWindowSecondsOk(); ok && v != nil {
					dtModel.ExcludeNotQueriedWindowSeconds = types.Int64Value(*v)
				}
				if v, ok := dt.GetExcludeNotUsedInAssetsOk(); ok && v != nil {
					dtModel.ExcludeNotUsedInAssets = types.BoolValue(*v)
				}
				dataModel.DynamicTags = dtModel
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

	// exclude_tags_mode is Optional+Computed+Default(false), so state.ExcludeTagsMode is always
	// known by the time we get here. Do not make this conditional: the backend 400s an update that
	// touches exclude_not_* fields unless exclude_tags_mode is explicitly present in the body
	// (apiv2handler.go:1312).
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
			if !m.Data.DynamicTags.ExcludeNotQueriedWindowSeconds.IsNull() {
				dt.SetExcludeNotQueriedWindowSeconds(m.Data.DynamicTags.ExcludeNotQueriedWindowSeconds.ValueInt64())
			}
			if !m.Data.DynamicTags.ExcludeNotUsedInAssets.IsNull() {
				dt.SetExcludeNotUsedInAssets(m.Data.DynamicTags.ExcludeNotUsedInAssets.ValueBool())
			}
			d.SetDynamicTags(*dt)
		}

		opts.SetData(*d)
	}
	return *opts
}
