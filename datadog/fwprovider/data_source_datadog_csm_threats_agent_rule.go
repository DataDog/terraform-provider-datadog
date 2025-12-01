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
	_ datasource.DataSourceWithConfigure = &csmThreatsAgentRulesDataSource{}
)

type csmThreatsAgentRulesDataSource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsAgentRulesDataSourceModel struct {
	PolicyId      types.String                         `tfsdk:"policy_id"`
	Id            types.String                         `tfsdk:"id"`
	AgentRulesIds types.List                           `tfsdk:"agent_rules_ids"`
	AgentRules    []csmThreatsAgentRuleDataSourceModel `tfsdk:"agent_rules"`
}

type csmThreatsAgentRuleDataSourceModel struct {
	Id          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	Enabled     types.Bool    `tfsdk:"enabled"`
	Expression  types.String  `tfsdk:"expression"`
	ProductTags types.Set     `tfsdk:"product_tags"`
	Actions     []ActionModel `tfsdk:"actions"`
}

func NewCSMThreatsAgentRulesDataSource() datasource.DataSource {
	return &csmThreatsAgentRulesDataSource{}
}

func (r *csmThreatsAgentRulesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (*csmThreatsAgentRulesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "csm_threats_agent_rules"
}

func (r *csmThreatsAgentRulesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state csmThreatsAgentRulesDataSourceModel
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
	agentRules := make([]csmThreatsAgentRuleDataSourceModel, len(data))

	for idx, agentRule := range res.GetData() {
		var agentRuleModel csmThreatsAgentRuleDataSourceModel
		agentRuleModel.Id = types.StringValue(agentRule.GetId())
		attributes := agentRule.Attributes
		agentRuleModel.Name = types.StringValue(attributes.GetName())
		agentRuleModel.Description = types.StringValue(attributes.GetDescription())
		agentRuleModel.Enabled = types.BoolValue(attributes.GetEnabled())
		agentRuleModel.Expression = types.StringValue(*attributes.Expression)
		tags := attributes.GetProductTags()
		tagSet := make(map[string]struct{})
		for _, tag := range tags {
			tagSet[tag] = struct{}{}
		}
		uniqueTags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			uniqueTags = append(uniqueTags, tag)
		}

		productTags, diags := types.SetValueFrom(ctx, types.StringType, uniqueTags)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			continue
		}
		agentRuleModel.ProductTags = productTags

		// Handle actions
		var actions []ActionModel
		for _, act := range attributes.GetActions() {
			action := ActionModel{}

			if act.Set != nil {
				setAction := &SetActionModel{}
				s := act.Set

				if s.Name != nil {
					setAction.Name = types.StringValue(*s.Name)
				} else {
					setAction.Name = types.StringNull()
				}
				if s.Field != nil {
					setAction.Field = types.StringValue(*s.Field)
				} else {
					setAction.Field = types.StringNull()
				}
				if s.Value != nil {
					setAction.Value = types.StringValue(*s.Value)
				} else {
					setAction.Value = types.StringNull()
				}
				if s.Append != nil {
					setAction.Append = types.BoolValue(*s.Append)
				} else {
					setAction.Append = types.BoolValue(false)
				}
				if s.Size != nil {
					setAction.Size = types.Int64Value(*s.Size)
				} else {
					setAction.Size = types.Int64Value(0)
				}
				if s.Ttl != nil {
					setAction.Ttl = types.Int64Value(*s.Ttl)
				} else {
					setAction.Ttl = types.Int64Value(0)
				}
				if s.Scope != nil {
					setAction.Scope = types.StringValue(*s.Scope)
				} else {
					setAction.Scope = types.StringValue("")
				}
				if s.Expression != nil {
					setAction.Expression = types.StringValue(*s.Expression)
				} else {
					setAction.Expression = types.StringNull()
				}
				if s.Inherited != nil {
					setAction.Inherited = types.BoolValue(*s.Inherited)
				} else {
					setAction.Inherited = types.BoolValue(false)
				}
				if s.DefaultValue != nil {
					setAction.DefaultValue = types.StringValue(*s.DefaultValue)
				} else {
					setAction.DefaultValue = types.StringNull()
				}
				action.Set = setAction
			}

			if act.Hash != nil {
				action.Hash = &HashActionModel{}
			}

			actions = append(actions, action)
		}
		agentRuleModel.Actions = actions

		agentRuleIds[idx] = agentRule.GetId()
		agentRules[idx] = agentRuleModel
	}

	stateId := strings.Join(agentRuleIds, "--")
	state.Id = types.StringValue(computeDataSourceID(&stateId))
	tfAgentRuleIds, diags := types.ListValueFrom(ctx, types.StringType, agentRuleIds)
	response.Diagnostics.Append(diags...)
	state.AgentRulesIds = tfAgentRuleIds
	state.AgentRules = agentRules

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (*csmThreatsAgentRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Agent rules.",
		Attributes: map[string]schema.Attribute{
			// Input
			"policy_id": schema.StringAttribute{
				Description: "Listing only the rules in the policy with this field as the ID",
				Optional:    true,
			},
			// Output
			"id": schema.StringAttribute{
				Description: "The ID of the data source",
				Computed:    true,
			},
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
						"actions": types.ListType{
							ElemType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"set": types.ObjectType{
										AttrTypes: map[string]attr.Type{
											"name":          types.StringType,
											"value":         types.StringType,
											"field":         types.StringType,
											"append":        types.BoolType,
											"size":          types.Int64Type,
											"ttl":           types.Int64Type,
											"scope":         types.StringType,
											"expression":    types.StringType,
											"inherited":     types.BoolType,
											"default_value": types.StringType,
										},
									},
									"hash": types.ObjectType{
										AttrTypes: map[string]attr.Type{},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func computeDataSourceID(ids *string) string {
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
