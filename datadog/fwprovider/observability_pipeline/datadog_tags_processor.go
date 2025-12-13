package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatadogTagsProcessorModel represents the Terraform model for the DatadogTagsProcessor
type DatadogTagsProcessorModel struct {
	Mode   types.String   `tfsdk:"mode"`
	Action types.String   `tfsdk:"action"`
	Keys   []types.String `tfsdk:"keys"`
}

// DatadogTagsProcessorSchema returns the schema for the DatadogTagsProcessor
func DatadogTagsProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"mode":   types.StringType,
					"action": types.StringType,
					"keys":   types.ListType{ElemType: types.StringType},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"mode": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf("filter"),
					},
				},
				"action": schema.StringAttribute{
					Required: true,
					Validators: []validator.String{
						stringvalidator.OneOf("include", "exclude"),
					},
				},
				"keys": schema.ListAttribute{
					ElementType: types.StringType,
					Required:    true,
				},
			},
		},
	}
}

// ExpandDatadogTagsProcessor converts the Terraform model to the API model
func ExpandDatadogTagsProcessor(src *DatadogTagsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineDatadogTagsProcessorWithDefaults()
	proc.SetMode(datadogV2.ObservabilityPipelineDatadogTagsProcessorMode(src.Mode.ValueString()))
	proc.SetAction(datadogV2.ObservabilityPipelineDatadogTagsProcessorAction(src.Action.ValueString()))

	var keys []string
	for _, key := range src.Keys {
		keys = append(keys, key.ValueString())
	}
	proc.SetKeys(keys)

	return datadogV2.ObservabilityPipelineDatadogTagsProcessorAsObservabilityPipelineConfigProcessorItem(proc)
}

// FlattenDatadogTagsProcessor converts the API model to the Terraform model
func FlattenDatadogTagsProcessor(src *datadogV2.ObservabilityPipelineDatadogTagsProcessor) *DatadogTagsProcessorModel {
	if src == nil {
		return nil
	}

	var keys []types.String
	for _, key := range src.GetKeys() {
		keys = append(keys, types.StringValue(key))
	}

	return &DatadogTagsProcessorModel{
		Mode:   types.StringValue(string(src.GetMode())),
		Action: types.StringValue(string(src.GetAction())),
		Keys:   keys,
	}
}
