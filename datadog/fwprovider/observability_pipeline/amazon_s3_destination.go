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
	Id           types.String  `tfsdk:"id"`
	Inputs       types.List    `tfsdk:"inputs"`
	Bucket       types.String  `tfsdk:"bucket"`
	Region       types.String  `tfsdk:"region"`
	KeyPrefix    types.String  `tfsdk:"key_prefix"`
	StorageClass types.String  `tfsdk:"storage_class"`
	Auth         *AwsAuthModel `tfsdk:"auth"`
}

// AmazonS3DestinationSchema returns the schema for the AmazonS3Destination
func AmazonS3DestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `amazon_s3` destination sends your logs in Datadog-rehydratable format to an Amazon S3 bucket for archiving.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Required:    true,
					Description: "Unique identifier for the destination component.",
				},
				"inputs": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
					Description: "A list of component IDs whose output is used as the `input` for this component.",
				},
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
				"auth": AwsAuthSchema(),
			},
		},
	}
}

// ExpandAmazonS3Destination converts the Terraform model to the API model
func ExpandAmazonS3Destination(ctx context.Context, src *AmazonS3DestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAmazonS3DestinationWithDefaults()
	dest.SetId(src.Id.ValueString())

	var inputs []string
	src.Inputs.ElementsAs(ctx, &inputs, false)
	dest.SetInputs(inputs)

	dest.SetBucket(src.Bucket.ValueString())
	dest.SetRegion(src.Region.ValueString())
	dest.SetKeyPrefix(src.KeyPrefix.ValueString())
	dest.SetStorageClass(datadogV2.ObservabilityPipelineAmazonS3DestinationStorageClass(src.StorageClass.ValueString()))

	if src.Auth != nil {
		auth := ExpandAwsAuth(src.Auth)
		if auth != nil {
			dest.SetAuth(*auth)
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

	inputs, _ := types.ListValueFrom(ctx, types.StringType, src.GetInputs())

	model := &AmazonS3DestinationModel{
		Id:           types.StringValue(src.GetId()),
		Inputs:       inputs,
		Bucket:       types.StringValue(src.GetBucket()),
		Region:       types.StringValue(src.GetRegion()),
		KeyPrefix:    types.StringValue(src.GetKeyPrefix()),
		StorageClass: types.StringValue(string(src.GetStorageClass())),
	}

	if auth, ok := src.GetAuthOk(); ok {
		model.Auth = FlattenAwsAuth(auth)
	}

	return model
}
