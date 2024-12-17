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
	_ datasource.DataSource = &awsAvailableNamespacesDataSource{}
)

func NewAwsAvailableNamespacesDataSource() datasource.DataSource {
	return &awsAvailableNamespacesDataSource{}
}

type awsAvailableNamespacesDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	AvailableNamespaces []string     `tfsdk:"aws_namespaces"`
}

type awsAvailableNamespacesDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (r *awsAvailableNamespacesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (d *awsAvailableNamespacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_available_namespaces"
}

func (d *awsAvailableNamespacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all available AWS namespaces. This is the list of allowed values for `metrics_config.namespace_filters` `include_only` or `exclude_only` in [`datadog_integration_aws_account` resource](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/integration_aws_account).",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"aws_namespaces": schema.ListAttribute{
				Description: "List of available AWS namespaces.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *awsAvailableNamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsAvailableNamespacesDataSourceModel
	if resp.Diagnostics.HasError() {
		return
	}

	awsAvailableNamespacesResp, httpResp, err := d.Api.ListAWSNamespaces(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS Namespaces"), ""))
		return
	}

	state.ID = types.StringValue("integration-aws-available-namespaces")

	d.updateState(&state, &awsAvailableNamespacesResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (d *awsAvailableNamespacesDataSource) updateState(state *awsAvailableNamespacesDataSourceModel, resp *datadogV2.AWSNamespacesResponse) {
	namespacesDd := resp.Data.GetAttributes().Namespaces
	state.AvailableNamespaces = append(state.AvailableNamespaces, namespacesDd...)
}
