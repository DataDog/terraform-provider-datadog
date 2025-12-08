package observability_pipeline

import (
	"context"

	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AmazonSecurityLakeDestinationModel represents the Terraform model for the AmazonSecurityLakeDestination
type AmazonSecurityLakeDestinationModel struct {
	Bucket           types.String  `tfsdk:"bucket"`
	Region           types.String  `tfsdk:"region"`
	CustomSourceName types.String  `tfsdk:"custom_source_name"`
	Tls              *tlsModel     `tfsdk:"tls"`
	Auth             *AwsAuthModel `tfsdk:"auth"`
}

func AmazonSecurityLakeDestinationSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `amazon_security_lake` destination sends your logs to Amazon Security Lake.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"bucket": schema.StringAttribute{
					Required:    true,
					Description: "Name of the Amazon S3 bucket in Security Lake (3-63 characters).",
				},
				"region": schema.StringAttribute{
					Required:    true,
					Description: "AWS region of the Security Lake bucket.",
				},
				"custom_source_name": schema.StringAttribute{
					Required:    true,
					Description: "Custom source name for the logs in Security Lake.",
				},
			},
			Blocks: map[string]schema.Block{
				"tls":  TlsSchema(),
				"auth": AwsAuthSchema(),
			},
		},
	}
}

func ExpandObservabilityPipelinesAmazonSecurityLakeDestination(ctx context.Context, id string, inputs types.List, src *AmazonSecurityLakeDestinationModel) datadogV2.ObservabilityPipelineConfigDestinationItem {
	dest := datadogV2.NewObservabilityPipelineAmazonSecurityLakeDestinationWithDefaults()
	dest.SetId(id)

	var inputsList []string
	inputs.ElementsAs(ctx, &inputsList, false)
	dest.SetInputs(inputsList)

	if !src.Bucket.IsNull() {
		dest.SetBucket(src.Bucket.ValueString())
	}
	if !src.Region.IsNull() {
		dest.SetRegion(src.Region.ValueString())
	}
	if !src.CustomSourceName.IsNull() {
		dest.SetCustomSourceName(src.CustomSourceName.ValueString())
	}
	if src.Tls != nil {
		dest.Tls = ExpandTls(src.Tls)
	}
	if src.Auth != nil {
		dest.Auth = ExpandAwsAuth(src.Auth)
	}

	return datadogV2.ObservabilityPipelineConfigDestinationItem{
		ObservabilityPipelineAmazonSecurityLakeDestination: dest,
	}
}

func FlattenObservabilityPipelinesAmazonSecurityLakeDestination(ctx context.Context, src *datadogV2.ObservabilityPipelineAmazonSecurityLakeDestination) *AmazonSecurityLakeDestinationModel {
	if src == nil {
		return nil
	}

	model := &AmazonSecurityLakeDestinationModel{
		Bucket:           types.StringValue(src.GetBucket()),
		Region:           types.StringValue(src.GetRegion()),
		CustomSourceName: types.StringValue(src.GetCustomSourceName()),
	}

	if src.Tls != nil {
		tls := FlattenTls(src.Tls)
		model.Tls = &tls
	}
	if src.Auth != nil {
		auth := FlattenAwsAuth(src.Auth)
		model.Auth = auth
	}

	return model
}
