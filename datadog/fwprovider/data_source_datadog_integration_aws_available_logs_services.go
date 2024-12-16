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
	_ datasource.DataSource = &awsLogsServicesDataSource{}
)

func NewAwsLogsServicesDataSource() datasource.DataSource {
	return &awsLogsServicesDataSource{}
}

type awsLogsServicesDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	LogsServices []string     `tfsdk:"aws_logs_services"`
}

type awsLogsServicesDataSource struct {
	Api  *datadogV2.AWSLogsIntegrationApi
	Auth context.Context
}

func (r *awsLogsServicesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSLogsIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (d *awsLogsServicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_available_logs_services"
}

func (d *awsLogsServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all AWS log ready services. This is the list of allowed values for `logs_config.lambda_forwarder.sources` in [`datadog_integration_aws_account` resource](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/integration_aws_account).",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"aws_logs_services": schema.ListAttribute{
				Description: "List of AWS log ready services.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *awsLogsServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsLogsServicesDataSourceModel
	if resp.Diagnostics.HasError() {
		return
	}

	awsLogsServicesResp, httpResp, err := d.Api.ListAWSLogsServices(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS Logs Services"), ""))
		return
	}

	state.ID = types.StringValue("integration-aws-available-logs-services")

	d.updateState(&state, &awsLogsServicesResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (d *awsLogsServicesDataSource) updateState(state *awsLogsServicesDataSourceModel, resp *datadogV2.AWSLogsServicesResponse) {
	logsServicesDd := resp.Data.GetAttributes().LogsServices
	state.LogsServices = append(state.LogsServices, logsServicesDd...)
}
