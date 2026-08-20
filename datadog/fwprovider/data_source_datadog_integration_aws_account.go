package fwprovider

import (
	"context"
	"fmt"
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
	_ datasource.DataSource = &awsIntegrationAccountDataSource{}
)

func NewAwsIntegrationAccountDataSource() datasource.DataSource {
	return &awsIntegrationAccountDataSource{}
}

type awsIntegrationAccountDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	AwsAccountId types.String `tfsdk:"aws_account_id"`
}

type awsIntegrationAccountDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (d *awsIntegrationAccountDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	d.Auth = providerData.Auth
}

func (d *awsIntegrationAccountDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "integration_aws_account"
}

func (d *awsIntegrationAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve the Datadog AWS account integration config ID for a given AWS Account ID. This is the ID used to import an existing AWS integration into the [`datadog_integration_aws_account` resource](https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/integration_aws_account).",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"aws_account_id": schema.StringAttribute{
				Required:    true,
				Description: "The AWS Account ID of the integration config to look up.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9]{12}$`), "must be a 12-digit AWS Account ID without dashes"),
				},
			},
		},
	}
}

func (d *awsIntegrationAccountDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state awsIntegrationAccountDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	awsAccountId := state.AwsAccountId.ValueString()
	awsAccountsResp, httpResp, err := d.Api.ListAWSAccounts(d.Auth,
		*datadogV2.NewListAWSAccountsOptionalParameters().WithAwsAccountId(awsAccountId))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS Accounts"), ""))
		return
	}

	accounts := awsAccountsResp.GetData()
	if len(accounts) == 0 {
		response.Diagnostics.AddError(
			"AWS account integration config not found",
			fmt.Sprintf("no AWS account integration config found for AWS Account ID %s", awsAccountId),
		)
		return
	}
	if len(accounts) > 1 {
		response.Diagnostics.AddError(
			"multiple AWS account integration configs found",
			fmt.Sprintf("found %d AWS account integration configs for AWS Account ID %s, expected exactly one", len(accounts), awsAccountId),
		)
		return
	}

	state.ID = types.StringValue(accounts[0].GetId())

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
