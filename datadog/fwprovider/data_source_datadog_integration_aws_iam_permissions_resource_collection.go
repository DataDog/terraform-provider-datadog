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
	_ datasource.DataSource = &awsIntegrationIAMPermissionsResourceCollectionDataSource{}
)

func NewAwsIntegrationIAMPermissionsResourceCollectionDataSource() datasource.DataSource {
	return &awsIntegrationIAMPermissionsResourceCollectionDataSource{}
}

type awsIntegrationIAMPermissionsResourceCollectionDataSourceModel struct {
	ID                               types.String `tfsdk:"id"`
	IAMPermissionsResourceCollection types.List   `tfsdk:"iam_permissions"`
}

type awsIntegrationIAMPermissionsResourceCollectionDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (r *awsIntegrationIAMPermissionsResourceCollectionDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (d *awsIntegrationIAMPermissionsResourceCollectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_iam_permissions_resource_collection"
}

func (d *awsIntegrationIAMPermissionsResourceCollectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the IAM permissions required for AWS integration resource collection.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"iam_permissions": schema.ListAttribute{
				Description: "The list of IAM actions required for AWS integration resource collection.",
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *awsIntegrationIAMPermissionsResourceCollectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsIntegrationIAMPermissionsResourceCollectionDataSourceModel
	if resp.Diagnostics.HasError() {
		return
	}

	IAMPermissionsResourceCollectionResp, httpResp, err := d.Api.GetAWSIntegrationIAMPermissionsResourceCollection(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS IAM Permissions"), ""))
		return
	}

	state.ID = types.StringValue("integration-aws-iam-permissions")

	d.updateState(&state, &IAMPermissionsResourceCollectionResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *awsIntegrationIAMPermissionsResourceCollectionDataSource) updateState(state *awsIntegrationIAMPermissionsResourceCollectionDataSourceModel, resp *datadogV2.AWSIntegrationIamPermissionsResponse) {
	permissions := resp.Data.Attributes.Permissions
	state.IAMPermissionsResourceCollection, _ = types.ListValueFrom(context.Background(), types.StringType, permissions)
}
