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

// AmazonS3GenericDestinationModel represents the Terraform model for the AmazonS3GenericDestination
type AmazonS3GenericDestinationModel struct {
	Bucket        types.String                        `tfsdk:"bucket"`
	Region        types.String                        `tfsdk:"region"`
	KeyPrefix     types.String                        `tfsdk:"key_prefix"`
	StorageClass  types.String                        `tfsdk:"storage_class"`
	Encoding      []AmazonS3GenericEncodingModel      `tfsdk:"encoding"`
	Compression   []AmazonS3GenericCompressionModel   `tfsdk:"compression"`
	Auth          []AwsAuthModel                      `tfsdk:"auth"`
	BatchSettings []AmazonS3GenericBatchSettingsModel `tfsdk:"batch_settings"`
}

// AmazonS3GenericEncodingModel represents the encoding format for the destination.
type AmazonS3GenericEncodingModel struct {
	Type types.String `tfsdk:"type"`
}

// AmazonS3GenericCompressionModel represents the compression algorithm applied to encoded logs.
type AmazonS3GenericCompressionModel struct {
	Type  types.String `tfsdk:"type"`
	Level types.Int64  `tfsdk:"level"`
}

// AmazonS3GenericBatchSettingsModel represents event batching settings.
type AmazonS3GenericBatchSettingsModel struct {
	BatchSize   types.Int64 `tfsdk:"batch_size"`
	TimeoutSecs types.Int64 `tfsdk:"timeout_secs"`
}

// AmazonS3GenericDestinationSchema returns the schema for the AmazonS3GenericDestination
func AmazonS3GenericDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `amazon_s3_generic` destination sends your logs to an Amazon S3 bucket.",
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
					Optional:    true,
					Description: "Optional prefix for object keys.",
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
				"encoding": schema.ListNestedBlock{
					Description: "Encoding format for the destination.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required:    true,
								Description: "The encoding type.",
								Validators: []validator.String{
									stringvalidator.OneOf("json", "parquet"),
								},
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.IsRequired(),
					},
				},
				"compression": schema.ListNestedBlock{
					Description: "Compression algorithm applied to encoded logs.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required:    true,
								Description: "The compression type. Use `gzip` or `zstd` with a `level`, or `snappy` without.",
								Validators: []validator.String{
									stringvalidator.OneOf("gzip", "zstd", "snappy"),
								},
							},
							"level": schema.Int64Attribute{
								Optional:    true,
								Description: "Compression level. Required for `gzip` and `zstd`; not used for `snappy`.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
						listvalidator.IsRequired(),
					},
				},
				"auth":          AwsAuthSchema(),
				"batch_settings": schema.ListNestedBlock{
					Description: "Event batching settings.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"batch_size": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum batch size in bytes.",
							},
							"timeout_secs": schema.Int64Attribute{
								Optional:    true,
								Description: "Maximum number of seconds to wait before flushing the batch.",
							},
						},
					},
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

// ExpandAmazonS3GenericDestination converts the Terraform model to the API model
func ExpandAmazonS3GenericDestination(ctx context.Context, id string, inputs types.List, src *AmazonS3GenericDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAmazonS3GenericDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	dest.SetBucket(src.Bucket.ValueString())
	dest.SetRegion(src.Region.ValueString())
	dest.SetStorageClass(datadogV2.ObservabilityPipelineAmazonS3DestinationStorageClass(src.StorageClass.ValueString()))

	if !src.KeyPrefix.IsNull() {
		dest.SetKeyPrefix(src.KeyPrefix.ValueString())
	}

	if len(src.Encoding) > 0 {
		dest.SetEncoding(expandS3GenericEncoding(src.Encoding[0]))
	}

	if len(src.Compression) > 0 {
		dest.SetCompression(expandS3GenericCompression(src.Compression[0]))
	}

	if len(src.Auth) > 0 {
		dest.SetAuth(ExpandAwsAuth(src.Auth[0]))
	}

	if len(src.BatchSettings) > 0 {
		dest.SetBatchSettings(expandS3GenericBatchSettings(src.BatchSettings[0]))
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonS3GenericDestination: dest,
	}
}

func expandS3GenericEncoding(m AmazonS3GenericEncodingModel) datadogV2.ObservabilityPipelineAmazonS3GenericEncoding {
	switch m.Type.ValueString() {
	case "parquet":
		return datadogV2.ObservabilityPipelineAmazonS3GenericEncodingParquetAsObservabilityPipelineAmazonS3GenericEncoding(
			datadogV2.NewObservabilityPipelineAmazonS3GenericEncodingParquetWithDefaults(),
		)
	default: // "json"
		return datadogV2.ObservabilityPipelineAmazonS3GenericEncodingJsonAsObservabilityPipelineAmazonS3GenericEncoding(
			datadogV2.NewObservabilityPipelineAmazonS3GenericEncodingJsonWithDefaults(),
		)
	}
}

func expandS3GenericCompression(m AmazonS3GenericCompressionModel) datadogV2.ObservabilityPipelineAmazonS3GenericCompression {
	switch m.Type.ValueString() {
	case "gzip":
		c := datadogV2.NewObservabilityPipelineAmazonS3GenericCompressionGzipWithDefaults()
		if !m.Level.IsNull() {
			c.SetLevel(m.Level.ValueInt64())
		}
		return datadogV2.ObservabilityPipelineAmazonS3GenericCompressionGzipAsObservabilityPipelineAmazonS3GenericCompression(c)
	case "zstd":
		c := datadogV2.NewObservabilityPipelineAmazonS3GenericCompressionZstdWithDefaults()
		if !m.Level.IsNull() {
			c.SetLevel(m.Level.ValueInt64())
		}
		return datadogV2.ObservabilityPipelineAmazonS3GenericCompressionZstdAsObservabilityPipelineAmazonS3GenericCompression(c)
	default: // "snappy"
		return datadogV2.ObservabilityPipelineAmazonS3GenericCompressionSnappyAsObservabilityPipelineAmazonS3GenericCompression(
			datadogV2.NewObservabilityPipelineAmazonS3GenericCompressionSnappyWithDefaults(),
		)
	}
}

func expandS3GenericBatchSettings(m AmazonS3GenericBatchSettingsModel) datadogV2.ObservabilityPipelineAmazonS3GenericBatchSettings {
	bs := datadogV2.NewObservabilityPipelineAmazonS3GenericBatchSettingsWithDefaults()
	if !m.BatchSize.IsNull() {
		bs.SetBatchSize(m.BatchSize.ValueInt64())
	}
	if !m.TimeoutSecs.IsNull() {
		bs.SetTimeoutSecs(m.TimeoutSecs.ValueInt64())
	}
	return *bs
}

// FlattenAmazonS3GenericDestination converts the API model to the Terraform model
func FlattenAmazonS3GenericDestination(src *datadogV2.ObservabilityPipelineAmazonS3GenericDestination) *AmazonS3GenericDestinationModel {
	if src == nil {
		return nil
	}

	model := &AmazonS3GenericDestinationModel{
		Bucket:       types.StringValue(src.GetBucket()),
		Region:       types.StringValue(src.GetRegion()),
		StorageClass: types.StringValue(string(src.GetStorageClass())),
		Encoding:     flattenS3GenericEncoding(src.GetEncoding()),
		Compression:  flattenS3GenericCompression(src.GetCompression()),
	}

	if v, ok := src.GetKeyPrefixOk(); ok {
		model.KeyPrefix = types.StringValue(*v)
	}

	if auth, ok := src.GetAuthOk(); ok {
		model.Auth = FlattenAwsAuth(auth)
	}

	if bs, ok := src.GetBatchSettingsOk(); ok {
		model.BatchSettings = []AmazonS3GenericBatchSettingsModel{flattenS3GenericBatchSettings(bs)}
	}

	return model
}

func flattenS3GenericEncoding(src datadogV2.ObservabilityPipelineAmazonS3GenericEncoding) []AmazonS3GenericEncodingModel {
	switch {
	case src.ObservabilityPipelineAmazonS3GenericEncodingParquet != nil:
		return []AmazonS3GenericEncodingModel{{Type: types.StringValue("parquet")}}
	default:
		return []AmazonS3GenericEncodingModel{{Type: types.StringValue("json")}}
	}
}

func flattenS3GenericCompression(src datadogV2.ObservabilityPipelineAmazonS3GenericCompression) []AmazonS3GenericCompressionModel {
	switch {
	case src.ObservabilityPipelineAmazonS3GenericCompressionGzip != nil:
		return []AmazonS3GenericCompressionModel{{
			Type:  types.StringValue("gzip"),
			Level: types.Int64Value(src.ObservabilityPipelineAmazonS3GenericCompressionGzip.GetLevel()),
		}}
	case src.ObservabilityPipelineAmazonS3GenericCompressionZstd != nil:
		return []AmazonS3GenericCompressionModel{{
			Type:  types.StringValue("zstd"),
			Level: types.Int64Value(src.ObservabilityPipelineAmazonS3GenericCompressionZstd.GetLevel()),
		}}
	default: // snappy
		return []AmazonS3GenericCompressionModel{{
			Type:  types.StringValue("snappy"),
			Level: types.Int64Null(),
		}}
	}
}

func flattenS3GenericBatchSettings(src *datadogV2.ObservabilityPipelineAmazonS3GenericBatchSettings) AmazonS3GenericBatchSettingsModel {
	m := AmazonS3GenericBatchSettingsModel{
		BatchSize:   types.Int64Null(),
		TimeoutSecs: types.Int64Null(),
	}
	if v, ok := src.GetBatchSizeOk(); ok {
		m.BatchSize = types.Int64Value(*v)
	}
	if v, ok := src.GetTimeoutSecsOk(); ok {
		m.TimeoutSecs = types.Int64Value(*v)
	}
	return m
}
