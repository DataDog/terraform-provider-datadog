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
	_ datasource.DataSource = &awsIntegrationExternalIDDataSource{}
)

func NewAwsIntegrationExternalIDDataSource() datasource.DataSource {
	return &awsIntegrationExternalIDDataSource{}
}

type awsIntegrationExternalIDDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	AwsAccountId types.String `tfsdk:"aws_account_id"`
	ExternalId   types.String `tfsdk:"external_id"`
}

type awsIntegrationExternalIDDataSource struct {
	Api  *datadogV2.AWSIntegrationApi
	Auth context.Context
}

func (d *awsIntegrationExternalIDDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV2()
	d.Auth = providerData.Auth
}

func (d *awsIntegrationExternalIDDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "integration_aws_external_id"
}

func (d *awsIntegrationExternalIDDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the external ID from an existing AWS integration. This can be used to reference the external ID value from an existing AWS account integration.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			"aws_account_id": schema.StringAttribute{
				Description: "The AWS account ID of the integration to retrieve the external ID from.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9]{12}$`), "invalid aws_account_id"),
				},
			},
			"external_id": schema.StringAttribute{
				Description: "The external ID associated with the AWS integration.",
				Computed:    true,
			},
		},
	}
}

func (d *awsIntegrationExternalIDDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state awsIntegrationExternalIDDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountId := state.AwsAccountId.ValueString()

	// Retrieve AWS account integration for the specific account ID
	optionalParams := datadogV2.NewListAWSAccountsOptionalParameters().WithAwsAccountId(awsAccountId)
	accountsResp, httpResp, err := d.Api.ListAWSAccounts(d.Auth, *optionalParams)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error querying AWS Account Integrations"), ""))
		return
	}
	if err := utils.CheckForUnparsed(accountsResp); err != nil {
		resp.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	// Check if we got exactly one account back
	accounts := accountsResp.GetData()
	if len(accounts) == 0 {
		resp.Diagnostics.AddError(
			"AWS Integration not found",
			fmt.Sprintf("No AWS integration found for account ID: %s", awsAccountId),
		)
		return
	}
	if len(accounts) > 1 {
		resp.Diagnostics.AddError(
			"Multiple AWS Integrations found",
			fmt.Sprintf("Unexpected: Multiple integrations found for account ID %s", awsAccountId),
		)
		return
	}

	// Get the single account
	target := accounts[0]
	attributes := target.GetAttributes()
	authConfig, ok := attributes.GetAuthConfigOk()
	if !ok {
		resp.Diagnostics.AddError(
			"No auth config found",
			fmt.Sprintf("The AWS integration for account %s does not have an auth config", awsAccountId),
		)
		return
	}
	if authConfig.AWSAuthConfigRole == nil {
		resp.Diagnostics.AddError(
			"No role-based auth config found",
			fmt.Sprintf("The AWS integration for account %s does not use role-based authentication with external ID", awsAccountId),
		)
		return
	}

	externalId := authConfig.AWSAuthConfigRole.GetExternalId()

	// Set state
	state.ID = types.StringValue(fmt.Sprintf("integration-aws-external-id-%s", awsAccountId))
	state.AwsAccountId = types.StringValue(awsAccountId)
	state.ExternalId = types.StringValue(externalId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
