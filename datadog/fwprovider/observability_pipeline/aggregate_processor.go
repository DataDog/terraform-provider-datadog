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

type AggregateProcessorModel struct {
	IntervalSecs types.Int64  `tfsdk:"interval_secs"`
	Mode         types.String `tfsdk:"mode"`
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
		},
	}
}

func ExpandAggregateProcessor(common BaseProcessorFields, src *AggregateProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAggregateProcessorWithDefaults()
	common.ApplyTo(proc)
	proc.SetIntervalSecs(src.IntervalSecs.ValueInt64())
	proc.SetMode(datadogV2.ObservabilityPipelineAggregateProcessorMode(src.Mode.ValueString()))
	return datadogV2.ObservabilityPipelineAggregateProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenAggregateProcessor(src *datadogV2.ObservabilityPipelineAggregateProcessor) *AggregateProcessorModel {
	if src == nil {
		return nil
	}
	return &AggregateProcessorModel{
		IntervalSecs: types.Int64Value(src.GetIntervalSecs()),
		Mode:         types.StringValue(string(src.GetMode())),
	}
}
