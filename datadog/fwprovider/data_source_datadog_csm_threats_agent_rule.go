package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSourceWithConfigure = &csmThreatsAgentRulesDataSource{}
)

type csmThreatsAgentRulesDataSource struct {
	api  *datadogV2.CloudWorkloadSecurityApi
	auth context.Context
}

type csmThreatsAgentRuleDataSourceModel struct {
	AgentRules []csmThreatsAgentRuleModel `tfsdk:"agent_rules"`
}

func NewCSMThreatsAgentRulesDataSource() datasource.DataSource {
	return &csmThreatsAgentRulesDataSource{}
}

func (r *csmThreatsAgentRulesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCloudWorkloadSecurityApiV2()
	r.auth = providerData.Auth
}

func (*csmThreatsAgentRulesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "agent_rule"
}

func (r *csmThreatsAgentRulesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state csmThreatsAgentRuleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, _, err := r.api.ListCSMThreatsAgentRules(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching suppressions"))
		return
	}

	data := res.GetData()
	agent_rules := make([]csmThreatsAgentRuleModel, len(data))

	for _, agentRule := range res.GetData() {
		var agentRuleModel csmThreatsAgentRuleModel
		agentRuleModel.Id = types.StringValue(agentRule.GetId())
		attributes := agentRule.Attributes
		agentRuleModel.Name = types.StringValue(attributes.GetName())
		agentRuleModel.Description = types.StringValue(attributes.GetDescription())
		agentRuleModel.Enabled = types.BoolValue(attributes.GetEnabled())
		agentRuleModel.Expression = types.StringValue(*attributes.Expression)
		agent_rules = append(agent_rules, agentRuleModel)
	}

	state.AgentRules = agent_rules

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (*csmThreatsAgentRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing agent rules, and use them in other resources.",
		Attributes: map[string]schema.Attribute{
			"agent_rules": schema.ListAttribute{
				Computed:    true,
				Description: "List of agent_rules",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"name":        types.StringType,
						"description": types.StringType,
						"enabled":     types.BoolType,
						"expression":  types.StringType,
					},
				},
			},
		},
	}
}
