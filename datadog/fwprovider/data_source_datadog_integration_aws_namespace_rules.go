package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &datadogIntegrationAWSNamespaceRulesDatasource{}

func NewDatadogIntegrationAWSNamespaceRulesDatasource() datasource.DataSource {
	return &datadogIntegrationAWSNamespaceRulesDatasource{}
}

type datadogIntegrationAWSNamespaceRulesDatasourceModel struct {
	ID             types.String `tfsdk:"id"`
	NamespaceRules types.List   `tfsdk:"namespace_rules"`
}

type datadogIntegrationAWSNamespaceRulesDatasource struct {
	Api  *datadogV1.AWSIntegrationApi
	Auth context.Context
}

func (d *datadogIntegrationAWSNamespaceRulesDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogIntegrationAWSNamespaceRulesDatasource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "integration_aws_namespace_rules"
}

func (d *datadogIntegrationAWSNamespaceRulesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AWS Integration Namespace Rules data source. This can be used to retrieve all available namespace rules for a Datadog-AWS integration.",
		Attributes: map[string]schema.Attribute{
			"namespace_rules": schema.ListAttribute{
				Description: "The list of available namespace rules for a Datadog-AWS integration.",
				ElementType: types.StringType,
				Computed:    true,
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (d *datadogIntegrationAWSNamespaceRulesDatasource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogIntegrationAWSNamespaceRulesDatasourceModel
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResponse, err := d.Api.ListAvailableAWSNamespaces(d.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading available namespace rules. http response: %v", httpResponse)))
		return
	}

	state.NamespaceRules, _ = types.ListValueFrom(ctx, types.StringType, resp)
	state.ID = types.StringValue("integration-aws-namespace-rules")
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
