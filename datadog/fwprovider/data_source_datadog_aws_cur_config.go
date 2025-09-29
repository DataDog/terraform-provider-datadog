package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogAwsCurConfigDataSource{}
)

type datadogAwsCurConfigDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type datadogAwsCurConfigDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	CloudAccountId types.Int64 `tfsdk:"cloud_account_id"`

	// Computed values
	AccountId       types.String         `tfsdk:"account_id"`
	BucketName      types.String         `tfsdk:"bucket_name"`
	BucketRegion    types.String         `tfsdk:"bucket_region"`
	CreatedAt       types.String         `tfsdk:"created_at"`
	ReportName      types.String         `tfsdk:"report_name"`
	ReportPrefix    types.String         `tfsdk:"report_prefix"`
	Status          types.String         `tfsdk:"status"`
	StatusUpdatedAt types.String         `tfsdk:"status_updated_at"`
	UpdatedAt       types.String         `tfsdk:"updated_at"`
	ErrorMessages   types.List           `tfsdk:"error_messages"`
	AccountFilters  *accountFiltersModel `tfsdk:"account_filters"`
}

func NewDatadogAwsCurConfigDataSource() datasource.DataSource {
	return &datadogAwsCurConfigDataSource{}
}

func (d *datadogAwsCurConfigDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogAwsCurConfigDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cur_config"
}

func (d *datadogAwsCurConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific Datadog AWS CUR (Cost and Usage Report) configuration. This allows you to fetch details about an existing Cloud Cost Management configuration for AWS billing data access.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"cloud_account_id": schema.Int64Attribute{
				Required:    true,
				Description: "The Datadog cloud account ID for the AWS CUR configuration you want to retrieve information about.",
			},
			// Computed values
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The AWS account ID associated with this CUR configuration.",
			},
			"bucket_name": schema.StringAttribute{
				Computed:    true,
				Description: "The S3 bucket name where Cost and Usage Report files are stored.",
			},
			"bucket_region": schema.StringAttribute{
				Computed:    true,
				Description: "The AWS region where the S3 bucket is located.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the CUR configuration was created.",
			},
			"report_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the AWS Cost and Usage Report.",
			},
			"report_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "The S3 key prefix where CUR files are stored within the bucket.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the CUR configuration (e.g., active, archived).",
			},
			"status_updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the CUR configuration status was last updated.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the CUR configuration was last updated.",
			},
			"error_messages": schema.ListAttribute{
				Computed:    true,
				Description: "List of error messages if the CUR configuration encountered any issues.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			// Computed values
			"account_filters": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"include_new_accounts": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether new member accounts are automatically included in cost analysis.",
					},
					"excluded_accounts": schema.ListAttribute{
						Computed:    true,
						Description: "List of AWS account IDs excluded from cost analysis.",
						ElementType: types.StringType,
					},
					"included_accounts": schema.ListAttribute{
						Computed:    true,
						Description: "List of AWS account IDs included in cost analysis.",
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogAwsCurConfigDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogAwsCurConfigDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	cloudAccountId := state.CloudAccountId.ValueInt64()
	ddResp, _, err := d.Api.GetCostAWSCURConfig(d.Auth, cloudAccountId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog awsCurConfig"))
		return
	}

	responseData := ddResp.GetData()
	d.updateState(ctx, &state, &responseData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogAwsCurConfigDataSource) updateState(ctx context.Context, state *datadogAwsCurConfigDataSourceModel, awsCurConfigData *datadogV2.AwsCurConfigResponseData) {
	state.ID = types.StringValue(awsCurConfigData.GetId())
	// CloudAccountId is input parameter, don't overwrite it

	if attributes, ok := awsCurConfigData.GetAttributesOk(); ok {
		state.AccountId = types.StringValue(attributes.GetAccountId())
		state.BucketName = types.StringValue(attributes.GetBucketName())
		state.BucketRegion = types.StringValue(attributes.GetBucketRegion())
		state.CreatedAt = types.StringValue(attributes.GetCreatedAt())
		state.ReportName = types.StringValue(attributes.GetReportName())
		state.ReportPrefix = types.StringValue(attributes.GetReportPrefix())
		state.Status = types.StringValue(attributes.GetStatus())
		state.StatusUpdatedAt = types.StringValue(attributes.GetStatusUpdatedAt())
		state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt())
		state.ErrorMessages, _ = types.ListValueFrom(ctx, types.StringType, attributes.GetErrorMessages())
		if accountFilters, ok := attributes.GetAccountFiltersOk(); ok {
			state.AccountFilters = mapAccountFiltersFromResponseData(ctx, accountFilters)
		}
	}
}
