package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AmazonS3DestinationModel represents the Terraform model for the AmazonS3Destination
type AmazonS3DestinationModel struct {
	Bucket       types.String              `tfsdk:"bucket"`
	Region       types.String              `tfsdk:"region"`
	KeyPrefix    types.String              `tfsdk:"key_prefix"`
	StorageClass types.String              `tfsdk:"storage_class"`
	Auth         []AwsAuthModel            `tfsdk:"auth"`
	Buffer       []BufferOptionsModel      `tfsdk:"buffer"`
	Compression  []ArchiveCompressionModel `tfsdk:"compression"`
}

// AmazonS3DestinationSchema returns the schema for the AmazonS3Destination
func AmazonS3DestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `amazon_s3` destination sends your logs in Datadog-rehydratable format to an Amazon S3 bucket for archiving.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"bucket": schema.StringAttribute{
					Required:    true,
					Description: "S3 bucket name.",
				},
				"region": schema.StringAttribute{
					Required:    true,
					Description: "AWS region of the S3 bucket.",
				},
				"key_prefix": schema.StringAttribute{
					Required:    true,
					Description: "Prefix for object keys.",
				},
				"storage_class": schema.StringAttribute{
					Required:    true,
					Description: "S3 storage class.",
					Validators: []validator.String{
						stringvalidator.OneOf("STANDARD", "REDUCED_REDUNDANCY", "INTELLIGENT_TIERING", "STANDARD_IA", "EXPRESS_ONEZONE", "ONEZONE_IA", "GLACIER", "GLACIER_IR", "DEEP_ARCHIVE"),
					},
				},
			},
			Blocks: map[string]schema.Block{
				"auth":        AwsAuthSchema(),
				"buffer":      BufferOptionsSchema(),
				"compression": ArchiveCompressionSchema(),
			},
		},
	}
}

// ExpandAmazonS3Destination converts the Terraform model to the API model
func ExpandAmazonS3Destination(ctx context.Context, id string, inputs types.List, src *AmazonS3DestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAmazonS3DestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	dest.SetBucket(src.Bucket.ValueString())
	dest.SetRegion(src.Region.ValueString())
	dest.SetKeyPrefix(src.KeyPrefix.ValueString())
	dest.SetStorageClass(datadogV2.ObservabilityPipelineAmazonS3DestinationStorageClass(src.StorageClass.ValueString()))

	if len(src.Auth) > 0 {
		dest.SetAuth(ExpandAwsAuth(src.Auth[0]))
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	if len(src.Compression) > 0 {
		dest.SetCompression(expandAmazonS3Compression(src.Compression[0]))
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonS3Destination: dest,
	}
}

// expandAmazonS3Compression converts the Terraform archive compression model to the API oneOf.
func expandAmazonS3Compression(m ArchiveCompressionModel) datadogV2.ObservabilityPipelineAmazonS3DestinationCompression {
	switch m.Algorithm.ValueString() {
	case "gzip":
		c := datadogV2.NewObservabilityPipelineAmazonS3DestinationCompressionGzipWithDefaults()
		if !m.Level.IsNull() {
			c.SetLevel(m.Level.ValueInt64())
		}
		return datadogV2.ObservabilityPipelineAmazonS3DestinationCompressionGzipAsObservabilityPipelineAmazonS3DestinationCompression(c)
	case "zstd":
		c := datadogV2.NewObservabilityPipelineAmazonS3DestinationCompressionZstdWithDefaults()
		if !m.Level.IsNull() {
			c.SetLevel(m.Level.ValueInt64())
		}
		return datadogV2.ObservabilityPipelineAmazonS3DestinationCompressionZstdAsObservabilityPipelineAmazonS3DestinationCompression(c)
	default: // "none"
		return datadogV2.ObservabilityPipelineAmazonS3DestinationCompressionNoneAsObservabilityPipelineAmazonS3DestinationCompression(
			datadogV2.NewObservabilityPipelineAmazonS3DestinationCompressionNoneWithDefaults(),
		)
	}
}

// FlattenAmazonS3Destination converts the API model to the Terraform model
func FlattenAmazonS3Destination(ctx context.Context, src *datadogV2.ObservabilityPipelineAmazonS3Destination) *AmazonS3DestinationModel {
	if src == nil {
		return nil
	}

	model := &AmazonS3DestinationModel{
		Bucket:       types.StringValue(src.GetBucket()),
		Region:       types.StringValue(src.GetRegion()),
		KeyPrefix:    types.StringValue(src.GetKeyPrefix()),
		StorageClass: types.StringValue(string(src.GetStorageClass())),
	}

	if auth, ok := src.GetAuthOk(); ok {
		model.Auth = FlattenAwsAuth(auth)
	}

	if buffer, ok := src.GetBufferOk(); ok {
		outBuffer := FlattenBufferOptions(buffer)
		if outBuffer != nil {
			model.Buffer = []BufferOptionsModel{*outBuffer}
		}
	}

	if compression, ok := src.GetCompressionOk(); ok {
		model.Compression = flattenAmazonS3Compression(compression)
	}

	return model
}

// flattenAmazonS3Compression converts the API archive compression oneOf to the Terraform model.
func flattenAmazonS3Compression(src *datadogV2.ObservabilityPipelineAmazonS3DestinationCompression) []ArchiveCompressionModel {
	if src == nil {
		return nil
	}
	switch {
	case src.ObservabilityPipelineAmazonS3DestinationCompressionGzip != nil:
		return []ArchiveCompressionModel{{
			Algorithm: types.StringValue("gzip"),
			Level:     types.Int64Value(src.ObservabilityPipelineAmazonS3DestinationCompressionGzip.GetLevel()),
		}}
	case src.ObservabilityPipelineAmazonS3DestinationCompressionZstd != nil:
		return []ArchiveCompressionModel{{
			Algorithm: types.StringValue("zstd"),
			Level:     types.Int64Value(src.ObservabilityPipelineAmazonS3DestinationCompressionZstd.GetLevel()),
		}}
	case src.ObservabilityPipelineAmazonS3DestinationCompressionNone != nil:
		return []ArchiveCompressionModel{{
			Algorithm: types.StringValue("none"),
			Level:     types.Int64Null(),
		}}
	default:
		return nil
	}
}
