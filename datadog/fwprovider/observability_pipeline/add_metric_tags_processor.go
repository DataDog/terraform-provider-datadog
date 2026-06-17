package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AddMetricTagsProcessorModel struct {
	Tags []AddMetricTagsProcessorTagModel `tfsdk:"tag"`
}

type AddMetricTagsProcessorTagModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func AddMetricTagsProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `add_metric_tags` processor adds static tags to metrics.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{},
			Blocks: map[string]schema.Block{
				"tag": schema.ListNestedBlock{
					Description: "A list of static tags to add to each metric. Up to 15 tags may be defined.",
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(15),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Required:    true,
								Description: "The tag name.",
							},
							"value": schema.StringAttribute{
								Required:    true,
								Description: "The tag value.",
							},
						},
					},
				},
			},
		},
	}
}

func ExpandAddMetricTagsProcessor(common BaseProcessorFields, src *AddMetricTagsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineAddMetricTagsProcessorWithDefaults()
	common.ApplyTo(proc)

	tags := make([]datadogV2.ObservabilityPipelineFieldValue, 0, len(src.Tags))
	for _, t := range src.Tags {
		tags = append(tags, datadogV2.ObservabilityPipelineFieldValue{
			Name:  t.Name.ValueString(),
			Value: t.Value.ValueString(),
		})
	}
	proc.SetTags(tags)

	return datadogV2.ObservabilityPipelineAddMetricTagsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenAddMetricTagsProcessor(src *datadogV2.ObservabilityPipelineAddMetricTagsProcessor) *AddMetricTagsProcessorModel {
	if src == nil {
		return nil
	}
	var tags []AddMetricTagsProcessorTagModel
	for _, t := range src.GetTags() {
		tags = append(tags, AddMetricTagsProcessorTagModel{
			Name:  types.StringValue(t.GetName()),
			Value: types.StringValue(t.GetValue()),
		})
	}
	return &AddMetricTagsProcessorModel{Tags: tags}
}
