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
	_ datasource.DataSource = &awsIntegrationIAMPermissionsStandardDataSource{}
)

func NewAwsIntegrationIAMPermissionsStandardDataSource() datasource.DataSource {
	return &awsIntegrationIAMPermissionsStandardDataSource{}
}

type awsIntegrationIAMPermissionsStandardDataSourceModel struct {
	ID                     types.String `tfsdk:"id"`
	IAMPermissionsStandard types.List   `tfsdk:"iam_permissions"`
}

type awsIntegrationIAMPermissionsStandardDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (r *awsIntegrationIAMPermissionsStandardDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (d *awsIntegrationIAMPermissionsStandardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_iam_permissions_standard"
}

func (d *awsIntegrationIAMPermissionsStandardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the standard IAM permissions required for the AWS integration. This provides the minimum list of IAM actions that should be included in the AWS role policy for Datadog integration.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"iam_permissions": schema.ListAttribute{
				Description: "The list of standard IAM actions required for the AWS integration.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *awsIntegrationIAMPermissionsStandardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsIntegrationIAMPermissionsStandardDataSourceModel
	if resp.Diagnostics.HasError() {
		return
	}

	IAMPermissionsStandardResp, httpResp, err := d.Api.GetAWSIntegrationIAMPermissionsStandard(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS IAM Permissions"), ""))
		return
	}

	state.ID = types.StringValue("integration-aws-iam-permissions")

	d.updateState(&state, &IAMPermissionsStandardResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *awsIntegrationIAMPermissionsStandardDataSource) updateState(state *awsIntegrationIAMPermissionsStandardDataSourceModel, resp *datadogV2.AWSIntegrationIamPermissionsResponse) {
	permissions := resp.Data.Attributes.Permissions
	state.IAMPermissionsStandard, _ = types.ListValueFrom(context.Background(), types.StringType, permissions)
}
