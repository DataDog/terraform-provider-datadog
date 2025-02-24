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
	_ datasource.DataSource = &awsPermissionsDataSource{}
)

func NewAwsPermissionsDataSource() datasource.DataSource {
	return &awsPermissionsDataSource{}
}

type awsPermissionsDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Permissions []string     `tfsdk:"aws_permissions"`
}

type awsPermissionsDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (r *awsPermissionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (d *awsPermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_permissions"
}

func (d *awsPermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the IAM permissions required by the AWS Integration (https://docs.datadoghq.com/integrations/amazon_web_services/?tab=manual#aws-iam-permissions).",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"aws_permissions": schema.ListAttribute{
				Description: "List of AWS Integration permissions.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *awsPermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsPermissionsDataSourceModel
	if resp.Diagnostics.HasError() {
		return
	}

	// pending implementation/generation of go api client
	awsPermissionsResp, httpResp, err := d.Api.ListAWSPermissions(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS Integration permissions"), ""))
		return
	}

	state.ID = types.StringValue("integration-aws-permissiosn")

	d.updateState(&state, &awsPermissionsResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (d *awsPermissionsDataSource) updateState(state *awsPermissionsDataSourceModel, resp *datadogV2.AWSPermissionsResponse) {
	permissionsDd := resp.Data.GetAttributes().Permissions
	state.Permissions = append(state.Permissions, permissionsDd...)
}
