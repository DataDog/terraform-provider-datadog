package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CustomProcessorModel represents the Terraform model for remap VRL processor configuration
type CustomProcessorModel struct {
	Id     types.String                `tfsdk:"id"`
	Inputs types.List                  `tfsdk:"inputs"`
	Remaps []CustomProcessorRemapModel `tfsdk:"remaps"`
}

// CustomProcessorRemapModel represents a single VRL remap rule
type CustomProcessorRemapModel struct {
	Include     types.String `tfsdk:"include"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Source      types.String `tfsdk:"source"`
	DropOnError types.Bool   `tfsdk:"drop_on_error"`
}

// ExpandCustomProcessor converts the Terraform model to the Datadog API model
func ExpandCustomProcessor(ctx context.Context, src *CustomProcessorModel) datadogV2.ObservabilityPipelineConfigProcessorItem {
	proc := datadogV2.NewObservabilityPipelineCustomProcessorWithDefaults()
	proc.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	proc.SetInputs(inputs)

	var remaps []datadogV2.ObservabilityPipelineCustomProcessorRemap
	for _, remap := range src.Remaps {
		remaps = append(remaps, datadogV2.ObservabilityPipelineCustomProcessorRemap{
			Include:     remap.Include.ValueString(),
			Name:        remap.Name.ValueString(),
			Source:      remap.Source.ValueString(),
			Enabled:     remap.Enabled.ValueBool(),
			DropOnError: remap.DropOnError.ValueBool(),
		})
	}
	proc.SetRemaps(remaps)

	return datadogV2.ObservabilityPipelineConfigProcessorItem{
		ObservabilityPipelineCustomProcessor: proc,
	}
}

// FlattenCustomProcessor converts the Datadog API model to the Terraform model
func FlattenCustomProcessor(ctx context.Context, src *datadogV2.ObservabilityPipelineCustomProcessor) *CustomProcessorModel {
	if src == nil {
		return nil
	}

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	var remaps []CustomProcessorRemapModel
	for _, remap := range src.GetRemaps() {
		remaps = append(remaps, CustomProcessorRemapModel{
			Include:     types.StringValue(remap.GetInclude()),
			Name:        types.StringValue(remap.GetName()),
			Source:      types.StringValue(remap.GetSource()),
			Enabled:     types.BoolValue(remap.GetEnabled()),
			DropOnError: types.BoolValue(remap.GetDropOnError()),
		})
	}

	return &CustomProcessorModel{
		Id:     types.StringValue(src.GetId()),
		Inputs: inputs,
		Remaps: remaps,
	}
}

// CustomProcessorSchema returns the schema for remap VRL processor
func CustomProcessorSchema() schema.ListNestedBlock {
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
								Required:    true,
								Description: "Whether this remap rule is enabled.",
							},
							"source": schema.StringAttribute{
								Required:    true,
								Description: "The VRL script source code that defines the transformation logic.",
							},
							"drop_on_error": schema.BoolAttribute{
								Required:    true,
								Description: "Whether to drop events that cause errors during transformation.",
							},
						},
					},
				},
			},
		},
	}
}
