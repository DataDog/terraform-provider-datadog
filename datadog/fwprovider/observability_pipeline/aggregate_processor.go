package observability_pipeline

import (
	"encoding/json"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AggregateProcessorModel struct {
	IntervalSecs      types.Int64                        `tfsdk:"interval_secs"`
	Mode              types.String                       `tfsdk:"mode"`
	AggregationTiming []*AggregateAggregationTimingModel `tfsdk:"aggregation_timing"`
}

type AggregateAggregationTimingModel struct {
	Type                types.String `tfsdk:"type"`
	AllowedLatenessSecs types.Int64  `tfsdk:"allowed_lateness_secs"`
}

func AggregateProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `aggregate` processor combines metrics that share the same name and tags into a single metric over a configurable interval.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"interval_secs": schema.Int64Attribute{
					Required:    true,
					Description: "The interval, in seconds, over which metrics are aggregated. Must be between 1 and 60.",
					Validators: []validator.Int64{
						int64validator.Between(1, 60),
					},
				},
				"mode": schema.StringAttribute{
					Required:    true,
					Description: "The aggregation mode. One of `auto`, `sum`, `latest`, `count`, `max`, `min`, `mean`.",
					Validators: []validator.String{
						stringvalidator.OneOf("auto", "sum", "latest", "count", "max", "min", "mean"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"aggregation_timing": schema.ListNestedBlock{
					Description: "Configures how metrics are assigned to aggregation windows. When omitted, metrics are grouped using system time.",
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required:    true,
								Description: "Determines whether metrics are assigned to aggregation windows based on when they are processed (`system_time`) or their timestamps (`event_time`).",
								Validators: []validator.String{
									stringvalidator.OneOf("system_time", "event_time"),
								},
							},
							"allowed_lateness_secs": schema.Int64Attribute{
								Optional:    true,
								Description: "Grace period, in seconds, for late-arriving metrics when using event time. Defaults to 10 seconds when omitted.",
								Validators: []validator.Int64{
									int64validator.Between(0, 3600),
								},
							},
						},
					},
				},
			},
		},
	}
}

func ExpandAggregateProcessor(common BaseProcessorFields, src *AggregateProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAggregateProcessorWithDefaults()
	common.ApplyTo(proc)
	proc.SetIntervalSecs(src.IntervalSecs.ValueInt64())
	proc.SetMode(datadogV2.ObservabilityPipelineAggregateProcessorMode(src.Mode.ValueString()))
	if len(src.AggregationTiming) > 0 {
		timing := src.AggregationTiming[0]
		serializedTiming := map[string]interface{}{
			"type": timing.Type.ValueString(),
		}
		if !timing.AllowedLatenessSecs.IsNull() && !timing.AllowedLatenessSecs.IsUnknown() {
			serializedTiming["allowed_lateness_secs"] = timing.AllowedLatenessSecs.ValueInt64()
		}
		// The generated API client does not expose aggregation_timing until the
		// corresponding API spec change is released. AdditionalProperties keeps the
		// provider compatible with the current client while emitting the same wire shape.
		if proc.AdditionalProperties == nil {
			proc.AdditionalProperties = make(map[string]interface{})
		}
		proc.AdditionalProperties["aggregation_timing"] = serializedTiming
	}
	return datadogV2.ObservabilityPipelineAggregateProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenAggregateProcessor(src *datadogV2.ObservabilityPipelineAggregateProcessor) *AggregateProcessorModel {
	if src == nil {
		return nil
	}
	return &AggregateProcessorModel{
		IntervalSecs:      types.Int64Value(src.GetIntervalSecs()),
		Mode:              types.StringValue(string(src.GetMode())),
		AggregationTiming: flattenAggregateAggregationTiming(src),
	}
}

func flattenAggregateAggregationTiming(src *datadogV2.ObservabilityPipelineAggregateProcessor) []*AggregateAggregationTimingModel {
	payload, err := json.Marshal(src)
	if err != nil {
		return nil
	}

	var serialized struct {
		AggregationTiming *struct {
			Type                string `json:"type"`
			AllowedLatenessSecs *int64 `json:"allowed_lateness_secs,omitempty"`
		} `json:"aggregation_timing,omitempty"`
	}
	if err := json.Unmarshal(payload, &serialized); err != nil || serialized.AggregationTiming == nil {
		return nil
	}

	allowedLatenessSecs := types.Int64Null()
	if serialized.AggregationTiming.AllowedLatenessSecs != nil {
		allowedLatenessSecs = types.Int64Value(*serialized.AggregationTiming.AllowedLatenessSecs)
	}

	return []*AggregateAggregationTimingModel{{
		Type:                types.StringValue(serialized.AggregationTiming.Type),
		AllowedLatenessSecs: allowedLatenessSecs,
	}}
}
