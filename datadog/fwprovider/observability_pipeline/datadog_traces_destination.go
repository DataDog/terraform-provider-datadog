package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DatadogTracesDestinationModel represents the Terraform model for the Datadog traces destination configuration
type DatadogTracesDestinationModel struct {
	Buffer []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandDatadogTracesDestination converts the Terraform model to the Datadog API model
func ExpandDatadogTracesDestination(ctx context.Context, id string, inputs types.List, src *DatadogTracesDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineDatadogTracesDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineDatadogTracesDestination: dest,
	}
}

// FlattenDatadogTracesDestination converts the Datadog API model to the Terraform model
func FlattenDatadogTracesDestination(src *datadogV2.ObservabilityPipelineDatadogTracesDestination) *DatadogTracesDestinationModel {
	if src == nil {
		return nil
	}

	out := &DatadogTracesDestinationModel{}

	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return out
}

// DatadogTracesDestinationSchema returns the schema for the Datadog traces destination
func DatadogTracesDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `datadog_traces` destination forwards traces to Datadog.",
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
