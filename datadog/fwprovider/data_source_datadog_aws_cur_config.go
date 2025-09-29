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
		Description: "Use this data source to retrieve information about a specific Datadog AWS CUR (Cost and Usage Report) configuration.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"cloud_account_id": schema.Int64Attribute{
				Required:    true,
				Description: "The cloud account ID of the AWS CUR config to retrieve.",
			},
			// Computed values
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `account_id`.",
			},
			"bucket_name": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `bucket_name`.",
			},
			"bucket_region": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `bucket_region`.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `created_at`.",
			},
			"report_name": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `report_name`.",
			},
			"report_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `report_prefix`.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `status`.",
			},
			"status_updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `status_updated_at`.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `updated_at`.",
			},
			"error_messages": schema.ListAttribute{
				Computed:    true,
				Description: "The `attributes` `error_messages`.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			// Computed values
			"account_filters": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"include_new_accounts": schema.BoolAttribute{
						Computed:    true,
						Description: "The `account_filters` `include_new_accounts`.",
					},
					"excluded_accounts": schema.ListAttribute{
						Computed:    true,
						Description: "The `account_filters` `excluded_accounts`.",
						ElementType: types.StringType,
					},
					"included_accounts": schema.ListAttribute{
						Computed:    true,
						Description: "The `account_filters` `included_accounts`.",
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
