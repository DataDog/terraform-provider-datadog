package fwprovider

// TODO: This resource requires the following API methods and types to be added to the datadog-api-client-go library:
// - AWSIntegrationApi.GetAWSAccountCCMConfig
// - AWSIntegrationApi.CreateAWSAccountCCMConfig
// - AWSIntegrationApi.UpdateAWSAccountCCMConfig
// - AWSIntegrationApi.DeleteAWSAccountCCMConfig
// - datadogV2.AWSCCMConfigResponse
// - datadogV2.AWSCCMConfigRequest
// - datadogV2.AWSCCMConfigRequestAttributes
// - datadogV2.AWSCCMConfigRequestData
// - datadogV2.AWSCCMConfig
// - datadogV2.DataExportConfig
//
// These types are defined in the aws.yaml OpenAPI spec at:
// /api/v2/integration/aws/accounts/{aws_account_config_id}/ccm_config

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationAwsAccountCcmConfigResource{}
	_ resource.ResourceWithImportState = &integrationAwsAccountCcmConfigResource{}
)

type integrationAwsAccountCcmConfigResource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

type integrationAwsAccountCcmConfigModel struct {
	ID                 types.String       `tfsdk:"id"`
	AwsAccountConfigId types.String       `tfsdk:"aws_account_config_id"`
	CcmConfig          *awsCcmConfigModel `tfsdk:"ccm_config"`
}

type awsCcmConfigModel struct {
	DataExportConfigs []*awsDataExportConfigModel `tfsdk:"data_export_configs"`
}

type awsDataExportConfigModel struct {
	ReportName   types.String `tfsdk:"report_name"`
	ReportPrefix types.String `tfsdk:"report_prefix"`
	ReportType   types.String `tfsdk:"report_type"`
	BucketName   types.String `tfsdk:"bucket_name"`
	BucketRegion types.String `tfsdk:"bucket_region"`
}

func NewIntegrationAwsAccountCcmConfigResource() resource.Resource {
	return &integrationAwsAccountCcmConfigResource{}
}

func (r *integrationAwsAccountCcmConfigResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationAwsAccountCcmConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws_account_ccm_config"
}

func (r *integrationAwsAccountCcmConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationAwsAccountCcmConfig resource. This can be used to create and manage Cloud Cost Management configuration for an AWS Account Integration.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"aws_account_config_id": schema.StringAttribute{
				Required:    true,
				Description: "Unique Datadog ID of the AWS Account Integration Config.",
			},
		},
		Blocks: map[string]schema.Block{
			"ccm_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"data_export_configs": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"report_name": schema.StringAttribute{
									Optional:    true,
									Description: "Name of the Cost and Usage Report.",
								},
								"report_prefix": schema.StringAttribute{
									Optional:    true,
									Description: "S3 prefix where the Cost and Usage Report is stored.",
								},
								"report_type": schema.StringAttribute{
									Optional:    true,
									Description: "Type of the Cost and Usage Report.",
								},
								"bucket_name": schema.StringAttribute{
									Optional:    true,
									Description: "Name of the S3 bucket where the Cost and Usage Report is stored.",
								},
								"bucket_region": schema.StringAttribute{
									Optional:    true,
									Description: "AWS region of the S3 bucket.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *integrationAwsAccountCcmConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("aws_account_config_id"), request, response)
}

func (r *integrationAwsAccountCcmConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAwsAccountCcmConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	awsAccountConfigId := state.AwsAccountConfigId.ValueString()

	// TODO: Uncomment when API methods are available in datadog-api-client-go
	resp, httpResp, err := r.Api.GetAWSAccountCCMConfig(r.Auth, awsAccountConfigId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationAwsAccountCcmConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAwsAccountCcmConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAwsAccountCcmConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountConfigId := state.AwsAccountConfigId.ValueString()

	body, diags := r.buildIntegrationAwsAccountCcmConfigRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// TODO: Uncomment when API methods are available in datadog-api-client-go
	resp, _, err := r.Api.CreateAWSAccountCCMConfig(r.Auth, awsAccountConfigId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating IntegrationAwsAccountCcmConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAwsAccountCcmConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationAwsAccountCcmConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountConfigId := state.AwsAccountConfigId.ValueString()

	body, diags := r.buildIntegrationAwsAccountCcmConfigRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// TODO: Uncomment when API methods are available in datadog-api-client-go
	resp, _, err := r.Api.UpdateAWSAccountCCMConfig(r.Auth, awsAccountConfigId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating IntegrationAwsAccountCcmConfig"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAwsAccountCcmConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAwsAccountCcmConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountConfigId := state.AwsAccountConfigId.ValueString()

	// TODO: Uncomment when API methods are available in datadog-api-client-go
	httpResp, err := r.Api.DeleteAWSAccountCCMConfig(r.Auth, awsAccountConfigId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_aws_account_ccm_config"))
		return
	}
}

// TODO: Update the type when AWSCCMConfigResponse is available in datadog-api-client-go
func (r *integrationAwsAccountCcmConfigResource) updateState(ctx context.Context, state *integrationAwsAccountCcmConfigModel, resp *datadogV2.AWSCCMConfigResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if dataExportConfigs, ok := attributes.GetDataExportConfigsOk(); ok && len(*dataExportConfigs) > 0 {
		if state.CcmConfig == nil {
			state.CcmConfig = &awsCcmConfigModel{}
		}
		state.CcmConfig.DataExportConfigs = []*awsDataExportConfigModel{}
		for _, dataExportConfigsDd := range *dataExportConfigs {
			dataExportConfigsTf := &awsDataExportConfigModel{}
			if reportName, ok := dataExportConfigsDd.GetReportNameOk(); ok {
				dataExportConfigsTf.ReportName = types.StringValue(*reportName)
			}
			if reportPrefix, ok := dataExportConfigsDd.GetReportPrefixOk(); ok {
				dataExportConfigsTf.ReportPrefix = types.StringValue(*reportPrefix)
			}
			if reportType, ok := dataExportConfigsDd.GetReportTypeOk(); ok {
				dataExportConfigsTf.ReportType = types.StringValue(*reportType)
			}
			if bucketName, ok := dataExportConfigsDd.GetBucketNameOk(); ok {
				dataExportConfigsTf.BucketName = types.StringValue(*bucketName)
			}
			if bucketRegion, ok := dataExportConfigsDd.GetBucketRegionOk(); ok {
				dataExportConfigsTf.BucketRegion = types.StringValue(*bucketRegion)
			}
			state.CcmConfig.DataExportConfigs = append(state.CcmConfig.DataExportConfigs, dataExportConfigsTf)
		}
	}
}

// TODO: Update the return type when AWSCCMConfigRequest is available in datadog-api-client-go
func (r *integrationAwsAccountCcmConfigResource) buildIntegrationAwsAccountCcmConfigRequestBody(ctx context.Context, state *integrationAwsAccountCcmConfigModel) (*datadogV2.AWSCCMConfigRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewAWSCCMConfigRequestAttributesWithDefaults()

	if state.CcmConfig != nil {
		var ccmConfig datadogV2.AWSCCMConfig

		if state.CcmConfig.DataExportConfigs != nil {
			var dataExportConfigs []datadogV2.DataExportConfig
			for _, dataExportConfigsTFItem := range state.CcmConfig.DataExportConfigs {
				dataExportConfigsDDItem := datadogV2.NewDataExportConfigWithDefaults()

				if !dataExportConfigsTFItem.ReportName.IsNull() {
					dataExportConfigsDDItem.SetReportName(dataExportConfigsTFItem.ReportName.ValueString())
				}
				if !dataExportConfigsTFItem.ReportPrefix.IsNull() {
					dataExportConfigsDDItem.SetReportPrefix(dataExportConfigsTFItem.ReportPrefix.ValueString())
				}
				if !dataExportConfigsTFItem.ReportType.IsNull() {
					dataExportConfigsDDItem.SetReportType(dataExportConfigsTFItem.ReportType.ValueString())
				}
				if !dataExportConfigsTFItem.BucketName.IsNull() {
					dataExportConfigsDDItem.SetBucketName(dataExportConfigsTFItem.BucketName.ValueString())
				}
				if !dataExportConfigsTFItem.BucketRegion.IsNull() {
					dataExportConfigsDDItem.SetBucketRegion(dataExportConfigsTFItem.BucketRegion.ValueString())
				}
				dataExportConfigs = append(dataExportConfigs, *dataExportConfigsDDItem)
			}
			ccmConfig.SetDataExportConfigs(dataExportConfigs)
		}
		attributes.CcmConfig = &ccmConfig
	}

	req := datadogV2.NewAWSCCMConfigRequestWithDefaults()
	req.Data = *datadogV2.NewAWSCCMConfigRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
