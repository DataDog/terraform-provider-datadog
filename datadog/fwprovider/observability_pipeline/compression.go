package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// compressionModel represents compression configuration
type compressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	Level     types.Int64  `tfsdk:"level"`
}

// ExpandCompression converts the Terraform compression model to the Datadog API model
func ExpandCompression(compressionTF *compressionModel) *datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompression {
	if compressionTF == nil {
		return nil
	}

	compression := datadogV2.NewObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompressionWithDefaults()

	if !compressionTF.Algorithm.IsNull() {
		compression.SetAlgorithm(datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompressionAlgorithm(compressionTF.Algorithm.ValueString()))
	}

	if !compressionTF.Level.IsNull() {
		compression.SetLevel(compressionTF.Level.ValueInt64())
	}
	return compression
}

// FlattenCompression converts the Datadog API compression model to the Terraform model
func FlattenCompression(src *datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompression) compressionModel {
	if src == nil {
		return compressionModel{}
	}

	return compressionModel{
		Algorithm: types.StringValue(string(src.GetAlgorithm())),
		Level:     types.Int64Value(src.GetLevel()),
	}
}

// CompressionSchema returns the schema for compression configuration
func CompressionSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: "Compression configuration for log events.",
		Attributes: map[string]schema.Attribute{
			"algorithm": schema.StringAttribute{
				Optional:    true, // must be optional to make the block optional
				Description: "Compression algorithm for log events.",
			},
			"level": schema.Int64Attribute{
				Optional:    true,
				Description: "Compression level.",
			},
		},
	}
}
