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
	_ datasource.DataSource = &datadogGcpUcConfigDataSource{}
)

type datadogGcpUcConfigDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type datadogGcpUcConfigDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	CloudAccountId types.Int64 `tfsdk:"cloud_account_id"`

	// Computed values
	AccountId         types.String `tfsdk:"account_id"`
	BucketName        types.String `tfsdk:"bucket_name"`
	CreatedAt         types.String `tfsdk:"created_at"`
	Dataset           types.String `tfsdk:"dataset"`
	ExportPrefix      types.String `tfsdk:"export_prefix"`
	ExportProjectName types.String `tfsdk:"export_project_name"`
	Months            types.Int64  `tfsdk:"months"`
	ProjectId         types.String `tfsdk:"project_id"`
	ServiceAccount    types.String `tfsdk:"service_account"`
	Status            types.String `tfsdk:"status"`
	StatusUpdatedAt   types.String `tfsdk:"status_updated_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	ErrorMessages     types.List   `tfsdk:"error_messages"`
}

func NewDatadogGcpUcConfigDataSource() datasource.DataSource {
	return &datadogGcpUcConfigDataSource{}
}

func (d *datadogGcpUcConfigDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogGcpUcConfigDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "gcp_uc_config"
}

func (d *datadogGcpUcConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific Datadog GCP Usage Cost configuration. This allows you to fetch details about an existing Cloud Cost Management configuration for GCP billing data access.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"cloud_account_id": schema.Int64Attribute{
				Required:    true,
				Description: "The Datadog cloud account ID for the GCP Usage Cost configuration you want to retrieve information about.",
			},
			// Computed values
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal account identifier for this GCP Usage Cost configuration.",
			},
			"bucket_name": schema.StringAttribute{
				Computed:    true,
				Description: "The Google Cloud Storage bucket name where Usage Cost export files are stored.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the GCP Usage Cost configuration was created.",
			},
			"dataset": schema.StringAttribute{
				Computed:    true,
				Description: "The resolved BigQuery dataset name used for the Usage Cost export.",
			},
			"export_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "The prefix path within the storage bucket where Usage Cost export files are organized.",
			},
			"export_project_name": schema.StringAttribute{
				Computed:    true,
				Description: "The Google Cloud Project ID where the Usage Cost export is configured.",
			},
			"months": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of months of historical cost data available for analysis.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resolved Google Cloud Project ID for the Usage Cost export.",
			},
			"service_account": schema.StringAttribute{
				Computed:    true,
				Description: "The Google Cloud service account email that Datadog uses to access the Usage Cost export data.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the GCP Usage Cost configuration (e.g., active, archived).",
			},
			"status_updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the configuration status was last updated.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the GCP Usage Cost configuration was last modified.",
			},
			"error_messages": schema.ListAttribute{
				Computed:    true,
				Description: "List of error messages if the GCP Usage Cost configuration encountered any issues during setup or data processing.",
				ElementType: types.StringType,
			},
		},
	}
}

func (d *datadogGcpUcConfigDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogGcpUcConfigDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	cloudAccountId := state.CloudAccountId.ValueInt64()
	ddResp, _, err := d.Api.GetCostGCPUsageCostConfig(d.Auth, cloudAccountId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog gcpUcConfig"))
		return
	}

	d.updateState(ctx, &state, &ddResp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogGcpUcConfigDataSource) updateState(ctx context.Context, state *datadogGcpUcConfigDataSourceModel, gcpUcConfigResponse *datadogV2.GcpUcConfigResponse) {
	responseData := gcpUcConfigResponse.GetData()
	state.ID = types.StringValue(responseData.GetId())
	// CloudAccountId is input parameter, don't overwrite it

	if attributes, ok := responseData.GetAttributesOk(); ok {
		state.AccountId = types.StringValue(attributes.GetAccountId())
		state.BucketName = types.StringValue(attributes.GetBucketName())
		state.Dataset = types.StringValue(attributes.GetDataset())
		state.ExportPrefix = types.StringValue(attributes.GetExportPrefix())
		state.ExportProjectName = types.StringValue(attributes.GetExportProjectName())
		state.ServiceAccount = types.StringValue(attributes.GetServiceAccount())
		state.Status = types.StringValue(attributes.GetStatus())
		state.CreatedAt = types.StringValue(attributes.GetCreatedAt())
		state.Months = types.Int64Value(int64(attributes.GetMonths()))
		state.ProjectId = types.StringValue(attributes.GetProjectId())
		state.StatusUpdatedAt = types.StringValue(attributes.GetStatusUpdatedAt())
		state.UpdatedAt = types.StringValue(attributes.GetUpdatedAt())
		state.ErrorMessages, _ = types.ListValueFrom(ctx, types.StringType, attributes.GetErrorMessages())
	}
}
