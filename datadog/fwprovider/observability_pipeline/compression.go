package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// compressionModel represents compression configuration
type compressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	Level     types.Int64  `tfsdk:"level"`
}

// ExpandCompression converts the Terraform compression model to the Datadog API model
func ExpandCompression(compressionTF []compressionModel) *datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompression {
	if len(compressionTF) == 0 {
		return nil
	}

	compression := datadogV2.NewObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompressionWithDefaults()

	if !compressionTF[0].Algorithm.IsNull() {
		compression.SetAlgorithm(datadogV2.ObservabilityPipelineCrowdStrikeNextGenSiemDestinationCompressionAlgorithm(compressionTF[0].Algorithm.ValueString()))
	}

	if !compressionTF[0].Level.IsNull() {
		compression.SetLevel(compressionTF[0].Level.ValueInt64())
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
func CompressionSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Compression configuration for log events.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"algorithm": schema.StringAttribute{
					Required:    true,
					Description: "Compression algorithm for log events.",
				},
				"level": schema.Int64Attribute{
					Optional:    true,
					Description: "Compression level.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// ArchiveCompressionModel represents the compression configuration for the Datadog Archives
// destinations (amazon_s3, google_cloud_storage). It is a tagged union keyed on `algorithm`
// with variants none/gzip/zstd.
type ArchiveCompressionModel struct {
	Algorithm types.String `tfsdk:"algorithm"`
	Level     types.Int64  `tfsdk:"level"`
}

// ArchiveCompressionSchema returns the schema for the archive destinations' compression block.
// The `compression` block is optional; when omitted, the worker keeps its existing default
// (gzip) so behavior is unchanged for existing pipelines.
func ArchiveCompressionSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "Compression configuration for archived logs. When omitted, the worker default (gzip) is used.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"algorithm": schema.StringAttribute{
					Required:    true,
					Description: "Compression algorithm. One of `none`, `gzip`, or `zstd`.",
					Validators: []validator.String{
						stringvalidator.OneOf("none", "gzip", "zstd"),
					},
				},
				"level": schema.Int64Attribute{
					Optional:    true,
					Description: "Compression level. Applies to `gzip` (1-9) and `zstd` (1-21); ignored for `none`.",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
