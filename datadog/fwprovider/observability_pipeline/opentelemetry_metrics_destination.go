package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// OpentelemetryMetricsDestinationModel represents the Terraform model for OpenTelemetry metrics destination configuration
type OpentelemetryMetricsDestinationModel struct {
	HttpClientUriKey types.String         `tfsdk:"http_client_uri_key"`
	Tls              []TlsModel           `tfsdk:"tls"`
	Buffer           []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandOpentelemetryMetricsDestination converts the Terraform model to the Datadog API model
func ExpandOpentelemetryMetricsDestination(ctx context.Context, id string, inputs types.List, src *OpentelemetryMetricsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineOpentelemetryMetricsDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	if !src.HttpClientUriKey.IsNull() {
		dest.SetHttpClientUriKey(src.HttpClientUriKey.ValueString())
	}
	if len(src.Tls) > 0 {
		dest.Tls = ExpandTls(src.Tls)
	}
	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineOpentelemetryMetricsDestination: dest,
	}
}

// FlattenOpentelemetryMetricsDestination converts the Datadog API model to the Terraform model
func FlattenOpentelemetryMetricsDestination(src *datadogV2.ObservabilityPipelineOpentelemetryMetricsDestination) *OpentelemetryMetricsDestinationModel {
	if src == nil {
		return nil
	}

	out := &OpentelemetryMetricsDestinationModel{}

	if v, ok := src.GetHttpClientUriKeyOk(); ok {
		out.HttpClientUriKey = types.StringValue(*v)
	}
	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}
	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			out.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	return out
}

// OpentelemetryMetricsDestinationSchema returns the schema for the OpenTelemetry metrics destination
func OpentelemetryMetricsDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `opentelemetry` destination forwards metrics using the OpenTelemetry Protocol (OTLP) over HTTP.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"http_client_uri_key": schema.StringAttribute{
					Optional:    true,
					Description: "Environment variable name containing the URI of the OTLP HTTP endpoint to send metrics to.",
				},
			},
			Blocks: map[string]schema.Block{
				"tls":    TlsSchema(),
				"buffer": BufferOptionsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
