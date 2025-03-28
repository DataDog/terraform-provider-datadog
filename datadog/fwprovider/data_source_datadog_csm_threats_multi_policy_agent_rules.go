package fwprovider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSourceWithConfigure = &csmThreatsMultiPolicyAgentRulesDataSource{}
)

type csmThreatsMultiPolicyAgentRulesDataSource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsMultiPolicyAgentRulesDataSourceModel struct {
	PolicyId      types.String                                    `tfsdk:"policy_id"`
	Id            types.String                                    `tfsdk:"id"`
	AgentRulesIds types.List                                      `tfsdk:"agent_rules_ids"`
	AgentRules    []csmThreatsMultiPolicyAgentRuleDataSourceModel `tfsdk:"agent_rules"`
}

type csmThreatsMultiPolicyAgentRuleDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Expression  types.String `tfsdk:"expression"`
	ProductTags types.Set    `tfsdk:"product_tags"`
}

func NewCSMThreatsMultiPolicyAgentRulesDataSource() datasource.DataSource {
	return &csmThreatsMultiPolicyAgentRulesDataSource{}
}

func (r *csmThreatsMultiPolicyAgentRulesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (r *csmThreatsMultiPolicyAgentRulesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "csm_threats_multi_policy_agent_rules"
}

func (r *csmThreatsMultiPolicyAgentRulesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state csmThreatsMultiPolicyAgentRulesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := datadogV2.NewListCSMThreatsAgentRulesOptionalParameters()
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		policyId := state.PolicyId.ValueString()
		params.WithPolicyId(policyId)
	}

	res, _, err := r.api.ListCSMThreatsAgentRules(r.auth, *params)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching agent rules"))
		return
	}

	data := res.GetData()
	agentRuleIds := make([]string, len(data))
	agentRules := make([]csmThreatsMultiPolicyAgentRuleDataSourceModel, len(data))

	for idx, agentRule := range res.GetData() {
		var agentRuleModel csmThreatsMultiPolicyAgentRuleDataSourceModel
		agentRuleModel.Id = types.StringValue(agentRule.GetId())
		attributes := agentRule.Attributes
		agentRuleModel.Name = types.StringValue(attributes.GetName())
		agentRuleModel.Description = types.StringValue(attributes.GetDescription())
		agentRuleModel.Enabled = types.BoolValue(attributes.GetEnabled())
		agentRuleModel.Expression = types.StringValue(*attributes.Expression)
		agentRuleModel.ProductTags, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetProductTags())
		agentRuleIds[idx] = agentRule.GetId()
		agentRules[idx] = agentRuleModel
	}

	stateId := strings.Join(agentRuleIds, "--")
	state.Id = types.StringValue(computeMultiPolicyAgentRulesID(&stateId))
	tfAgentRuleIds, diags := types.ListValueFrom(ctx, types.StringType, agentRuleIds)
	response.Diagnostics.Append(diags...)
	state.AgentRulesIds = tfAgentRuleIds
	state.AgentRules = agentRules

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func computeMultiPolicyAgentRulesID(ids *string) string {
	// Key for hashing
	var b strings.Builder
	if ids != nil {
		b.WriteString(*ids)
	}
	keyStr := b.String()
	h := sha256.New()
	h.Write([]byte(keyStr))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (*csmThreatsMultiPolicyAgentRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Agent rules.",
		Attributes: map[string]schema.Attribute{
			// Input
			"policy_id": schema.StringAttribute{
				Description: "Listing only the rules in the policy with this field as the ID",
				Optional:    true,
			},
			// Output
			"id": utils.ResourceIDAttribute(),
			"agent_rules_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of IDs for the Agent rules.",
				ElementType: types.StringType,
			},
			"agent_rules": schema.ListAttribute{
				Computed:    true,
				Description: "List of Agent rules",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":           types.StringType,
						"name":         types.StringType,
						"description":  types.StringType,
						"enabled":      types.BoolType,
						"expression":   types.StringType,
						"product_tags": types.SetType{ElemType: types.StringType},
					},
				},
			},
		},
	}
}
