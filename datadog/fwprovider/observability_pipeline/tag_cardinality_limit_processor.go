package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TagCardinalityLimitProcessorModel struct {
	LimitExceededAction types.String                                      `tfsdk:"limit_exceeded_action"`
	ValueLimit          types.Int64                                       `tfsdk:"value_limit"`
	TrackingMode        []TagCardinalityLimitProcessorTrackingModeModel   `tfsdk:"tracking_mode"`
	PerMetricLimits     []TagCardinalityLimitProcessorPerMetricLimitModel `tfsdk:"per_metric_limit"`
}

type TagCardinalityLimitProcessorTrackingModeModel struct {
	Mode types.String `tfsdk:"mode"`
}

type TagCardinalityLimitProcessorPerMetricLimitModel struct {
	MetricName          types.String                                   `tfsdk:"metric_name"`
	OverrideType        types.String                                   `tfsdk:"override_type"`
	LimitExceededAction types.String                                   `tfsdk:"limit_exceeded_action"`
	ValueLimit          types.Int64                                    `tfsdk:"value_limit"`
	PerTagLimits        []TagCardinalityLimitProcessorPerTagLimitModel `tfsdk:"per_tag_limit"`
}

type TagCardinalityLimitProcessorPerTagLimitModel struct {
	TagKey       types.String `tfsdk:"tag_key"`
	OverrideType types.String `tfsdk:"override_type"`
	ValueLimit   types.Int64  `tfsdk:"value_limit"`
}

func TagCardinalityLimitProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `tag_cardinality_limit` processor caps the number of distinct tag value combinations on metrics, dropping tags or events once the limit is exceeded.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"limit_exceeded_action": schema.StringAttribute{
					Required:    true,
					Description: "The default action to take when the cardinality limit is exceeded. One of `drop_tag`, `drop_event`.",
					Validators: []validator.String{
						stringvalidator.OneOf("drop_tag", "drop_event"),
					},
				},
				"value_limit": schema.Int64Attribute{
					Required:    true,
					Description: "The default maximum number of distinct tag value combinations allowed per metric. Between 0 and 1000000.",
					Validators: []validator.Int64{
						int64validator.Between(0, 1000000),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"tracking_mode": schema.ListNestedBlock{
					Description: "Controls whether the processor uses exact or probabilistic tag tracking.",
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Required:    true,
								Description: "The cardinality tracking algorithm to use. One of `exact_fingerprint`, `probabilistic`.",
								Validators: []validator.String{
									stringvalidator.OneOf("exact_fingerprint", "probabilistic"),
								},
							},
						},
					},
				},
				"per_metric_limit": schema.ListNestedBlock{
					Description: "Per-metric cardinality overrides that take precedence over the default `value_limit`.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(100),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"metric_name": schema.StringAttribute{
								Required:    true,
								Description: "The metric name this override applies to.",
							},
							"override_type": schema.StringAttribute{
								Required:    true,
								Description: "How the per-metric override is applied. One of `limit_override`, `excluded`.",
								Validators: []validator.String{
									stringvalidator.OneOf("limit_override", "excluded"),
								},
							},
							"limit_exceeded_action": schema.StringAttribute{
								Optional:    true,
								Description: "The action to take on this metric when the limit is exceeded. Required when `override_type` is `limit_override`; must be omitted when `override_type` is `excluded`.",
								Validators: []validator.String{
									stringvalidator.OneOf("drop_tag", "drop_event"),
								},
							},
							"value_limit": schema.Int64Attribute{
								Optional:    true,
								Description: "The cardinality cap for this metric. Required when `override_type` is `limit_override`; must be omitted when `override_type` is `excluded`.",
								Validators: []validator.Int64{
									int64validator.Between(0, 1000000),
								},
							},
						},
						Blocks: map[string]schema.Block{
							"per_tag_limit": schema.ListNestedBlock{
								Description: "Per-tag cardinality overrides that apply within this metric. Must be omitted when `override_type` is `excluded`.",
								Validators: []validator.List{
									listvalidator.SizeAtMost(50),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"tag_key": schema.StringAttribute{
											Required:    true,
											Description: "The tag key this override applies to.",
										},
										"override_type": schema.StringAttribute{
											Required:    true,
											Description: "How the per-tag override is applied. One of `limit_override`, `excluded`.",
											Validators: []validator.String{
												stringvalidator.OneOf("limit_override", "excluded"),
											},
										},
										"value_limit": schema.Int64Attribute{
											Optional:    true,
											Description: "The cardinality cap for this tag. Required when `override_type` is `limit_override`; must be omitted when `override_type` is `excluded`.",
											Validators: []validator.Int64{
												int64validator.Between(0, 1000000),
											},
										},
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

func ExpandTagCardinalityLimitProcessor(common BaseProcessorFields, src *TagCardinalityLimitProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineTagCardinalityLimitProcessorWithDefaults()
	common.ApplyTo(proc)

	proc.SetLimitExceededAction(datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorAction(src.LimitExceededAction.ValueString()))
	proc.SetValueLimit(src.ValueLimit.ValueInt64())
	if len(src.TrackingMode) > 0 {
		trackingMode := datadogV2.NewObservabilityPipelineTagCardinalityLimitProcessorTrackingModeWithDefaults()
		trackingMode.SetMode(datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorTrackingModeMode(src.TrackingMode[0].Mode.ValueString()))
		proc.SetTrackingMode(*trackingMode)
	}

	if len(src.PerMetricLimits) > 0 {
		perMetric := make([]datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorPerMetricLimit, 0, len(src.PerMetricLimits))
		for _, pm := range src.PerMetricLimits {
			item := datadogV2.NewObservabilityPipelineTagCardinalityLimitProcessorPerMetricLimitWithDefaults()
			item.SetMetricName(pm.MetricName.ValueString())
			item.SetOverrideType(datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorOverrideType(pm.OverrideType.ValueString()))
			if !pm.LimitExceededAction.IsNull() && !pm.LimitExceededAction.IsUnknown() {
				item.SetLimitExceededAction(datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorAction(pm.LimitExceededAction.ValueString()))
			}
			if !pm.ValueLimit.IsNull() && !pm.ValueLimit.IsUnknown() {
				item.SetValueLimit(pm.ValueLimit.ValueInt64())
			}
			if len(pm.PerTagLimits) > 0 {
				perTag := make([]datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorPerTagLimit, 0, len(pm.PerTagLimits))
				for _, pt := range pm.PerTagLimits {
					ti := datadogV2.NewObservabilityPipelineTagCardinalityLimitProcessorPerTagLimitWithDefaults()
					ti.SetTagKey(pt.TagKey.ValueString())
					ti.SetOverrideType(datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorOverrideType(pt.OverrideType.ValueString()))
					if !pt.ValueLimit.IsNull() && !pt.ValueLimit.IsUnknown() {
						ti.SetValueLimit(pt.ValueLimit.ValueInt64())
					}
					perTag = append(perTag, *ti)
				}
				item.SetPerTagLimits(perTag)
			}
			perMetric = append(perMetric, *item)
		}
		proc.SetPerMetricLimits(perMetric)
	}

	return datadogV2.ObservabilityPipelineTagCardinalityLimitProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenTagCardinalityLimitProcessor(src *datadogV2.ObservabilityPipelineTagCardinalityLimitProcessor) *TagCardinalityLimitProcessorModel {
	if src == nil {
		return nil
	}
	model := &TagCardinalityLimitProcessorModel{
		LimitExceededAction: types.StringValue(string(src.GetLimitExceededAction())),
		ValueLimit:          types.Int64Value(src.GetValueLimit()),
	}
	if trackingMode, ok := src.GetTrackingModeOk(); ok && trackingMode != nil {
		model.TrackingMode = append(model.TrackingMode, TagCardinalityLimitProcessorTrackingModeModel{
			Mode: types.StringValue(string(trackingMode.GetMode())),
		})
	}
	for _, pm := range src.GetPerMetricLimits() {
		pmModel := TagCardinalityLimitProcessorPerMetricLimitModel{
			MetricName:   types.StringValue(pm.GetMetricName()),
			OverrideType: types.StringValue(string(pm.GetOverrideType())),
		}
		if v, ok := pm.GetLimitExceededActionOk(); ok && v != nil {
			pmModel.LimitExceededAction = types.StringValue(string(*v))
		} else {
			pmModel.LimitExceededAction = types.StringNull()
		}
		if v, ok := pm.GetValueLimitOk(); ok && v != nil {
			pmModel.ValueLimit = types.Int64Value(*v)
		} else {
			pmModel.ValueLimit = types.Int64Null()
		}
		for _, pt := range pm.GetPerTagLimits() {
			ptModel := TagCardinalityLimitProcessorPerTagLimitModel{
				TagKey:       types.StringValue(pt.GetTagKey()),
				OverrideType: types.StringValue(string(pt.GetOverrideType())),
			}
			if v, ok := pt.GetValueLimitOk(); ok && v != nil {
				ptModel.ValueLimit = types.Int64Value(*v)
			} else {
				ptModel.ValueLimit = types.Int64Null()
			}
			pmModel.PerTagLimits = append(pmModel.PerTagLimits, ptModel)
		}
		model.PerMetricLimits = append(model.PerMetricLimits, pmModel)
	}
	return model
}
