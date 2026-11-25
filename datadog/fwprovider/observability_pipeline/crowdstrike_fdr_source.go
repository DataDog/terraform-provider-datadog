package observability_pipeline

import (
	datadogV2 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CrowdStrikeFdrSourceModel represents the Terraform model for the CrowdStrike FDR source configuration
type CrowdStrikeFdrSourceModel struct {
	Region                types.String `tfsdk:"region"`
	UrlKey                types.String `tfsdk:"url_key"`
	AccessKeyIdKey        types.String `tfsdk:"access_key_id_key"`
	SecretAccessKeyKey    types.String `tfsdk:"secret_access_key_key"`
	VisibilityTimeoutSecs types.Int64  `tfsdk:"visibility_timeout_secs"`
	ClientConcurrency     types.Int64  `tfsdk:"client_concurrency"`
	MaxNumberOfMessages   types.Int64  `tfsdk:"max_number_of_messages"`
	Tls                   []TlsModel   `tfsdk:"tls"`
}

// ExpandCrowdStrikeFdrSource converts the Terraform model to the Datadog API model
func ExpandCrowdStrikeFdrSource(src *CrowdStrikeFdrSourceModel, id string) datadogV2.ObservabilityPipelineConfigSourceItem {
	s := datadogV2.NewObservabilityPipelineCrowdStrikeFdrSourceWithDefaults()
	s.SetId(id)
	s.SetRegion(src.Region.ValueString())
	if !src.UrlKey.IsNull() {
		s.SetUrlKey(src.UrlKey.ValueString())
	}
	if !src.AccessKeyIdKey.IsNull() {
		s.SetAccessKeyIdKey(src.AccessKeyIdKey.ValueString())
	}
	if !src.SecretAccessKeyKey.IsNull() {
		s.SetSecretAccessKeyKey(src.SecretAccessKeyKey.ValueString())
	}
	if !src.VisibilityTimeoutSecs.IsNull() {
		s.SetVisibilityTimeoutSecs(src.VisibilityTimeoutSecs.ValueInt64())
	}
	if !src.ClientConcurrency.IsNull() {
		s.SetClientConcurrency(src.ClientConcurrency.ValueInt64())
	}
	if !src.MaxNumberOfMessages.IsNull() {
		s.SetMaxNumberOfMessages(src.MaxNumberOfMessages.ValueInt64())
	}
	s.Tls = ExpandTls(src.Tls)

	return datadogV2.ObservabilityPipelineConfigSourceItem{
		ObservabilityPipelineCrowdStrikeFdrSource: s,
	}
}

// FlattenCrowdStrikeFdrSource converts the Datadog API model to the Terraform model
func FlattenCrowdStrikeFdrSource(src *datadogV2.ObservabilityPipelineCrowdStrikeFdrSource) *CrowdStrikeFdrSourceModel {
	if src == nil {
		return nil
	}

	out := &CrowdStrikeFdrSourceModel{
		Region:                types.StringValue(src.GetRegion()),
		UrlKey:                types.StringPointerValue(src.UrlKey),
		AccessKeyIdKey:        types.StringPointerValue(src.AccessKeyIdKey),
		SecretAccessKeyKey:    types.StringPointerValue(src.SecretAccessKeyKey),
		VisibilityTimeoutSecs: types.Int64PointerValue(src.VisibilityTimeoutSecs),
		ClientConcurrency:     types.Int64PointerValue(src.ClientConcurrency),
		MaxNumberOfMessages:   types.Int64PointerValue(src.MaxNumberOfMessages),
	}
	if src.Tls != nil {
		out.Tls = FlattenTls(src.Tls)
	}

	return out
}

// CrowdStrikeFdrSourceSchema returns the schema for the CrowdStrike FDR source
func CrowdStrikeFdrSourceSchema() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		Description: "The `crowdstrike_fdr` source ingests CrowdStrike Falcon Data Replicator (FDR) data. It reads file notifications from the CrowdStrike-managed SQS queue for your FDR feed and downloads the referenced objects from the CrowdStrike-managed S3 bucket.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"region": schema.StringAttribute{
					Required:    true,
					Description: "AWS region of the FDR feed's SQS queue and S3 bucket.",
				},
				"url_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the URL of the FDR feed's SQS queue.",
				},
				"access_key_id_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the AWS access key ID that CrowdStrike generated for the FDR feed.",
				},
				"secret_access_key_key": schema.StringAttribute{
					Optional:    true,
					Description: "Name of the environment variable or secret that holds the AWS secret access key that CrowdStrike generated for the FDR feed.",
				},
				"visibility_timeout_secs": schema.Int64Attribute{
					Optional:    true,
					Description: "Number of seconds a received SQS message is hidden from other consumers. Set it higher than the time needed to download and process the largest FDR file. Defaults to 300.",
				},
				"client_concurrency": schema.Int64Attribute{
					Optional:    true,
					Description: "Number of concurrent tasks polling the SQS queue. Defaults to the number of Worker threads.",
				},
				"max_number_of_messages": schema.Int64Attribute{
					Optional:    true,
					Description: "Maximum number of SQS messages to receive per poll. Each message can reference multiple files. Defaults to 1.",
				},
			},
			Blocks: map[string]schema.Block{
				"tls": TlsSchema(),
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
