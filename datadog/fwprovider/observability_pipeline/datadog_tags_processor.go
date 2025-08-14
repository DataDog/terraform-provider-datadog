package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatadogTagsProcessorModel represents the Terraform model for the DatadogTagsProcessor
type DatadogTagsProcessorModel struct {
	Id      types.String   `tfsdk:"id"`
	Include types.String   `tfsdk:"include"`
	Inputs  types.List     `tfsdk:"inputs"`
	Mode    types.String   `tfsdk:"mode"`
	Action  types.String   `tfsdk:"action"`
	Keys    []types.String `tfsdk:"keys"`
}

// DatadogTagsProcessorSchema returns the schema for the DatadogTagsProcessor
func DatadogTagsProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":      types.StringType,
					"include": types.StringType,
					"inputs":  types.ListType{ElemType: types.StringType},
					"mode":    types.StringType,
					"action":  types.StringType,
					"keys":    types.ListType{ElemType: types.StringType},
				},
			},
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required: true,
				},
				"include": schema.StringAttribute{
					Required: true,
				},
				"inputs": schema.ListAttribute{
					ElementType: types.StringType,
					Required:    true,
				},
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
func ExpandDatadogTagsProcessor(ctx context.Context, src *DatadogTagsProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineDatadogTagsProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())
	proc.SetInclude(src.Include.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	proc.SetMode(datadogV2.ObservabilityPipelineDatadogTagsProcessorMode(src.Mode.ValueString()))
	proc.SetAction(datadogV2.ObservabilityPipelineDatadogTagsProcessorAction(src.Action.ValueString()))

	var keys []string
	for _, key := range src.Keys {
		keys = append(keys, key.ValueString())
	}
	proc.SetKeys(keys)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineDatadogTagsProcessor: proc,
	}
}

// FlattenDatadogTagsProcessor converts the API model to the Terraform model
func FlattenDatadogTagsProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineDatadogTagsProcessor) *DatadogTagsProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.Inputs)
	var keys []types.String
	for _, key := range src.Keys {
		keys = append(keys, types.StringValue(key))
	}

	return &DatadogTagsProcessorModel{
		Id:      types.StringValue(src.Id),
		Include: types.StringValue(src.Include),
		Inputs:  inputs,
		Mode:    types.StringValue(string(src.Mode)),
		Action:  types.StringValue(string(src.Action)),
		Keys:    keys,
	}
}
