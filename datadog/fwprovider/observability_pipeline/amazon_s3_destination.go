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
	Bucket               types.String         `tfsdk:"bucket"`
	Region               types.String         `tfsdk:"region"`
	KeyPrefix            types.String         `tfsdk:"key_prefix"`
	StorageClass         types.String         `tfsdk:"storage_class"`
	ServerSideEncryption types.String         `tfsdk:"server_side_encryption"`
	SseKmsKeyId          types.String         `tfsdk:"ssekms_key_id"`
	Auth                 []AwsAuthModel       `tfsdk:"auth"`
	Buffer               []BufferOptionsModel `tfsdk:"buffer"`
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
				"server_side_encryption": schema.StringAttribute{
					Optional:    true,
					Description: "The server-side encryption algorithm used when storing objects in S3. Valid values: `aws:kms`, `AES256`.",
					Validators: []validator.String{
						stringvalidator.OneOf("aws:kms", "AES256"),
					},
				},
				"ssekms_key_id": schema.StringAttribute{
					Optional:    true,
					Description: "ID of the AWS KMS key to use for SSE-KMS encryption. Only applies when `server_side_encryption` is `aws:kms`.",
				},
			},
			Blocks: map[string]schema.Block{
				"auth":   AwsAuthSchema(),
				"buffer": BufferOptionsSchema(),
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

	// TODO(OPA-5637): the client's ObservabilityPipelineAmazonS3Destination model has no typed
	// ServerSideEncryption/SsekmsKeyId fields yet (unlike the generic destination), so bridge through
	// AdditionalProperties, which round-trips via Marshal/Unmarshal. Replace with typed setters once
	// the client is regenerated from the api-spec change.
	if !src.ServerSideEncryption.IsNull() || !src.SseKmsKeyId.IsNull() {
		if dest.AdditionalProperties == nil {
			dest.AdditionalProperties = map[string]interface{}{}
		}
		if !src.ServerSideEncryption.IsNull() {
			dest.AdditionalProperties["server_side_encryption"] = src.ServerSideEncryption.ValueString()
		}
		if !src.SseKmsKeyId.IsNull() {
			dest.AdditionalProperties["ssekms_key_id"] = src.SseKmsKeyId.ValueString()
		}
	}

	if len(src.Auth) > 0 {
		dest.SetAuth(ExpandAwsAuth(src.Auth[0]))
	}

	if len(src.Buffer) > 0 {
		buffer := ExpandBufferOptions(src.Buffer[0])
		if buffer != nil {
			dest.SetBuffer(*buffer)
		}
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonS3Destination: dest,
	}
}

// FlattenAmazonS3Destination converts the API model to the Terraform model
func FlattenAmazonS3Destination(ctx context.Context, src *datadogV2.ObservabilityPipelineAmazonS3Destination) *AmazonS3DestinationModel {
	if src == nil {
		return nil
	}

	model := &AmazonS3DestinationModel{
		Bucket:               types.StringValue(src.GetBucket()),
		Region:               types.StringValue(src.GetRegion()),
		KeyPrefix:            types.StringValue(src.GetKeyPrefix()),
		StorageClass:         types.StringValue(string(src.GetStorageClass())),
		ServerSideEncryption: types.StringNull(),
		SseKmsKeyId:          types.StringNull(),
	}

	// TODO(OPA-5637): read via AdditionalProperties until the client exposes typed getters (see Expand).
	if v, ok := src.AdditionalProperties["server_side_encryption"].(string); ok && v != "" {
		model.ServerSideEncryption = types.StringValue(v)
	}
	if v, ok := src.AdditionalProperties["ssekms_key_id"].(string); ok && v != "" {
		model.SseKmsKeyId = types.StringValue(v)
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

	return model
}
