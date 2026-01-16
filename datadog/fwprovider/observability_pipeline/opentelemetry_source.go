package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// OpentelemetrySourceModel represents the Terraform model for opentelemetry source configuration
type OpentelemetrySourceModel struct {
	Tls []TlsModel `tfsdk:"tls"`
}

// ExpandOpentelemetrySource converts the Terraform model to the Datadog API model
func ExpandOpentelemetrySource(src *OpentelemetrySourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineOpentelemetrySourceWithDefaults()
	s.SetId(id)

	if len(src.Tls) > 0 {
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

	out := &OpentelemetrySourceModel{}

	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}

	return out
}

// OpentelemetrySourceSchema returns the schema for opentelemetry source
func OpentelemetrySourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `opentelemetry` source receives telemetry data using the OpenTelemetry Protocol (OTLP) over gRPC and HTTP.",
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"tls": TlsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
