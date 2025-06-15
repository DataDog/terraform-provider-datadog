package fwprovider

import (
	"context"
	"regexp"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &awsAccountUuidDataSource{}
)

func NewAwsAccountUuidDataSource() datasource.DataSource {
	return &awsAccountUuidDataSource{}
}

type awsAccountUuidDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	AwsAccountId types.String `tfsdk:"aws_account_id"`
}

type awsAccountUuidDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (r *awsAccountUuidDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (d *awsAccountUuidDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_account_uuid"
}

func (d *awsAccountUuidDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the Datadog AWS Account Config ID associated with your integrated account.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"aws_account_id": schema.StringAttribute{
				Required:    true,
				Description: "Your AWS Account ID without dashes.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9]{12}$`), "invalid aws_account_id"),
				},
			},
		},
	}
}

func (d *awsAccountUuidDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsAccountUuidDataSourceModel
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountId := state.AwsAccountId.String()

	params := datadogV2.ListAWSAccountsOptionalParameters{
		AwsAccountId: &awsAccountId,
	}

	awsAccountResp, httpResp, err := d.Api.ListAWSAccounts(ctx, params)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying for AWS Account"), ""))
		return
	}

	if len(awsAccountResp.Data) > 1 {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "found multiple AWS Account Integrations matching ID"), ""))
		return
	}

	if len(awsAccountResp.Data) < 1 {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "no AWS Account Integration matching ID was found"), ""))
		return
	}

	state.ID = types.StringValue("integration-aws-account-uuid")

	d.updateState(&state, &awsAccountResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (d *awsAccountUuidDataSource) updateState(state *awsAccountUuidDataSourceModel, resp *datadogV2.AWSAccountsResponse) {
	state.ID = types.StringValue(resp.Data[0].GetId())
}
