package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CustomProcessorProcessorModel represents the Terraform model for remap VRL processor configuration
type CustomProcessorProcessorModel struct {
	Id     types.String                         `tfsdk:"id"`
	Inputs types.List                           `tfsdk:"inputs"`
	Remaps []CustomProcessorProcessorRemapModel `tfsdk:"remaps"`
}

// CustomProcessorProcessorRemapModel represents a single VRL remap rule
type CustomProcessorProcessorRemapModel struct {
	Include     types.String `tfsdk:"include"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Source      types.String `tfsdk:"source"`
	DropOnError types.Bool   `tfsdk:"drop_on_error"`
}

// ExpandCustomProcessorProcessor converts the Terraform model to the Datadog API model
func ExpandCustomProcessorProcessor(ctx context.Context, src *CustomProcessorProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineCustomProcessorProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var remaps []datadogV2.ObservabilityPipelineCustomProcessorProcessorRemap
	for _, remap := range src.Remaps {
		apiRemap := datadogV2.ObservabilityPipelineCustomProcessorProcessorRemap{
			Include: remap.Include.ValueString(),
			Name:    remap.Name.ValueString(),
			Source:  remap.Source.ValueString(),
		}
		if !remap.Enabled.IsNull() {
			enabled := remap.Enabled.ValueBool()
			apiRemap.Enabled = &enabled
		}
		if !remap.DropOnError.IsNull() {
			dropOnError := remap.DropOnError.ValueBool()
			apiRemap.DropOnError = &dropOnError
		}
		remaps = append(remaps, apiRemap)
	}
	proc.SetRemaps(remaps)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineCustomProcessorProcessor: proc,
	}
}

// FlattenCustomProcessorProcessor converts the Datadog API model to the Terraform model
func FlattenCustomProcessorProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineCustomProcessorProcessor) *CustomProcessorProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	var remaps []CustomProcessorProcessorRemapModel
	for _, remap := range src.GetRemaps() {
		remapModel := CustomProcessorProcessorRemapModel{
			Include: types.StringValue(remap.GetInclude()),
			Name:    types.StringValue(remap.GetName()),
			Source:  types.StringValue(remap.GetSource()),
		}
		if remap.Enabled != nil {
			remapModel.Enabled = types.BoolValue(remap.GetEnabled())
		}
		if remap.DropOnError != nil {
			remapModel.DropOnError = types.BoolValue(remap.GetDropOnError())
		}
		remaps = append(remaps, remapModel)
	}

	return &CustomProcessorProcessorModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
		Remaps: remaps,
	}
}

// CustomProcessorProcessorSchema returns the schema for remap VRL processor
func CustomProcessorProcessorSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `custom_processor` processor transforms events using Vector Remap Language (VRL) scripts with advanced filtering capabilities.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:    true,
					Description: "The unique identifier for this processor.",
				},
				"inputs": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "A list of component IDs whose output is used as the input for this processor.",
				},
			},
			Blocks: map[string]schema.Block{
				"remaps": schema.ListNestedBlock{
					Description: "Array of VRL remap configurations. Each remap defines a transformation rule with its own filter and VRL script.",
					Validators: []validator.List{
						listvalidator.SizeAtLeast(1),
						listvalidator.SizeAtMost(15),
					},
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"include": schema.StringAttribute{
								Required:    true,
								Description: "A Datadog search query used to filter events for this specific remap rule.",
							},
							"name": schema.StringAttribute{
								Required:    true,
								Description: "A descriptive name for this remap rule.",
							},
							"enabled": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether this remap rule is enabled.",
							},
							"source": schema.StringAttribute{
								Required:    true,
								Description: "The VRL script source code that defines the transformation logic. Must not exceed 1000 characters and cannot contain forbidden functions.",
								Validators: []validator.String{
									stringvalidator.LengthAtMost(1000),
								},
							},
							"drop_on_error": schema.BoolAttribute{
								Optional:    true,
								Description: "Whether to drop events that cause errors during transformation.",
							},
						},
					},
				},
			},
		},
	}
}
