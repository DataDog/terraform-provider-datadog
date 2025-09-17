package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OpentelemetrySourceModel represents the Terraform model for OpenTelemetry source configuration
type OpentelemetrySourceModel struct {
	Id  types.String `tfsdk:"id"`
	Tls *tlsModel    `tfsdk:"tls"`
}

// ExpandOpentelemetrySource converts the Terraform model to the Datadog API model
func ExpandOpentelemetrySource(src *OpentelemetrySourceModel) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineOpentelemetrySourceWithDefaults()
	s.SetId(src.Id.ValueString())

	if src.Tls != nil {
		s.Tls = ExpandTls(src.Tls)
	}

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineOpentelemetrySource: s,
	}
}

// FlattenOpentelemetrySource converts the Datadog API model to the Terraform model
func FlattenOpentelemetrySource(src *datadogV2.ObservabilityPipelineOpentelemetrySource) *OpentelemetrySourceModel {
	if src == nil {
		return nil
	}

	out := &OpentelemetrySourceModel{
		Id: types.StringValue(src.GetId()),
	}

	if src.Tls != nil {
		tls := FlattenTls(src.Tls)
		out.Tls = &tls
	}

	return out
}

// OpentelemetrySourceSchema returns the schema for OpenTelemetry source
func OpentelemetrySourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `opentelemetry` source receives OpenTelemetry data through gRPC or HTTP.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:    true,
					Description: "The unique identifier for this component. Used to reference this component in other parts of the pipeline (e.g., as input to downstream components).",
				},
			},
			Blocks: map[string]schema.Block{
				"tls": TlsSchema(),
			},
		},
	}
}
