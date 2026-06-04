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

// SplunkHECMetricsDestinationModel represents the Terraform model for Splunk HEC metrics destination configuration
type SplunkHECMetricsDestinationModel struct {
	EndpointUrlKey   types.String         `tfsdk:"endpoint_url_key"`
	TokenKey         types.String         `tfsdk:"token_key"`
	DefaultNamespace types.String         `tfsdk:"default_namespace"`
	Index            types.String         `tfsdk:"index"`
	Source           types.String         `tfsdk:"source"`
	Sourcetype       types.String         `tfsdk:"sourcetype"`
	Compression      types.String         `tfsdk:"compression"`
	Tls              []TlsModel           `tfsdk:"tls"`
	Buffer           []BufferOptionsModel `tfsdk:"buffer"`
}

// ExpandSplunkHECMetricsDestination converts the Terraform model to the Datadog API model
func ExpandSplunkHECMetricsDestination(ctx context.Context, id string, inputs types.List, src *SplunkHECMetricsDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineSplunkHecMetricsDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	if !src.EndpointUrlKey.IsNull() {
		dest.SetEndpointUrlKey(src.EndpointUrlKey.ValueString())
	}
	if !src.TokenKey.IsNull() {
		dest.SetTokenKey(src.TokenKey.ValueString())
	}
	if !src.DefaultNamespace.IsNull() {
		dest.SetDefaultNamespace(src.DefaultNamespace.ValueString())
	}
	if !src.Index.IsNull() {
		dest.SetIndex(src.Index.ValueString())
	}
	if !src.Source.IsNull() {
		dest.SetSource(src.Source.ValueString())
	}
	if !src.Sourcetype.IsNull() {
		dest.SetSourcetype(src.Sourcetype.ValueString())
	}
	if !src.Compression.IsNull() {
		dest.SetCompression(datadogV2.ObservabilityPipelineSplunkHecMetricsDestinationCompression(src.Compression.ValueString()))
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
		ObservabilityPipelineSplunkHecMetricsDestination: dest,
	}
}

// FlattenSplunkHECMetricsDestination converts the Datadog API model to the Terraform model
func FlattenSplunkHECMetricsDestination(src *datadogV2.ObservabilityPipelineSplunkHecMetricsDestination) *SplunkHECMetricsDestinationModel {
	if src == nil {
		return nil
	}

	out := &SplunkHECMetricsDestinationModel{}

	if v, ok := src.GetEndpointUrlKeyOk(); ok {
		out.EndpointUrlKey = types.StringValue(*v)
	}
	if v, ok := src.GetTokenKeyOk(); ok {
		out.TokenKey = types.StringValue(*v)
	}
	if v, ok := src.GetDefaultNamespaceOk(); ok {
		out.DefaultNamespace = types.StringValue(*v)
	}
	if v, ok := src.GetIndexOk(); ok {
		out.Index = types.StringValue(*v)
	}
	if v, ok := src.GetSourceOk(); ok {
		out.Source = types.StringValue(*v)
	}
	if v, ok := src.GetSourcetypeOk(); ok {
		out.Sourcetype = types.StringValue(*v)
	}
	if v, ok := src.GetCompressionOk(); ok {
		out.Compression = types.StringValue(string(*v))
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

// SplunkHECMetricsDestinationSchema returns the schema for the Splunk HEC metrics destination
func SplunkHECMetricsDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `splunk_hec_metrics` destination forwards metrics to Splunk using the HTTP Event Collector (HEC).",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"endpoint_url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the Splunk HEC endpoint URL.",
				},
				"token_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the Splunk HEC token.",
				},
				"default_namespace": schema.StringAttribute{
					Optional:    true,
					Description: "Optional default namespace for metrics sent to Splunk HEC.",
				},
				"index": schema.StringAttribute{
					Optional:    true,
					Description: "Optional name of the Splunk index where metrics are written.",
				},
				"source": schema.StringAttribute{
					Optional:    true,
					Description: "The Splunk source field value for metric events.",
				},
				"sourcetype": schema.StringAttribute{
					Optional:    true,
					Description: "The Splunk sourcetype to assign to metric events.",
				},
				"compression": schema.StringAttribute{
					Optional:    true,
					Description: "Compression algorithm applied when sending metrics to Splunk HEC.",
					Validators: []validator.String{
						stringvalidator.OneOf("none", "gzip"),
					},
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
