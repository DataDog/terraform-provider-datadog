package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RenameMetricTagsProcessorModel struct {
	Tags []RenameMetricTagsProcessorTagModel `tfsdk:"tag"`
}

type RenameMetricTagsProcessorTagModel struct {
	Tag      types.String `tfsdk:"tag"`
	RenameTo types.String `tfsdk:"rename_to"`
}

func RenameMetricTagsProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `rename_metric_tags` processor changes the keys of tags on metrics.",
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{},
			Blocks: map[string]schema.Block{
				"tag": schema.ListNestedBlock{
					Description: "A list of rename rules. Up to 15 tags may be defined.",
					Validators: []validator.List{
						listvalidator.IsRequired(),
						listvalidator.SizeAtMost(15),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"tag": schema.StringAttribute{
								Required:    true,
								Description: "The original tag key on the metric event.",
							},
							"rename_to": schema.StringAttribute{
								Required:    true,
								Description: "The new tag key to assign in place of the original.",
							},
						},
					},
				},
			},
		},
	}
}

func ExpandRenameMetricTagsProcessor(common BaseProcessorFields, src *RenameMetricTagsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineRenameMetricTagsProcessorWithDefaults()
	common.ApplyTo(proc)

	tags := make([]datadogV2.ObservabilityPipelineRenameMetricTagsProcessorTag, 0, len(src.Tags))
	for _, t := range src.Tags {
		tags = append(tags, datadogV2.ObservabilityPipelineRenameMetricTagsProcessorTag{
			Tag:      t.Tag.ValueString(),
			RenameTo: t.RenameTo.ValueString(),
		})
	}
	proc.SetTags(tags)

	return datadogV2.ObservabilityPipelineRenameMetricTagsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

func FlattenRenameMetricTagsProcessor(src *datadogV2.ObservabilityPipelineRenameMetricTagsProcessor) *RenameMetricTagsProcessorModel {
	if src == nil {
		return nil
	}
	var tags []RenameMetricTagsProcessorTagModel
	for _, t := range src.GetTags() {
		tags = append(tags, RenameMetricTagsProcessorTagModel{
			Tag:      types.StringValue(t.GetTag()),
			RenameTo: types.StringValue(t.GetRenameTo()),
		})
	}
	return &RenameMetricTagsProcessorModel{Tags: tags}
}
