package fwprovider

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSourceWithConfigure = &csmThreatsPoliciesDataSource{}
)

type csmThreatsPoliciesDataSource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsPoliciesDataSourceModel struct {
	Id        types.String            `tfsdk:"id"`
	PolicyIds types.List              `tfsdk:"policy_ids"`
	Policies  []csmThreatsPolicyModel `tfsdk:"policies"`
}

func NewCSMThreatsPoliciesDataSource() datasource.DataSource {
	return &csmThreatsPoliciesDataSource{}
}

func (r *csmThreatsPoliciesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (*csmThreatsPoliciesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "csm_threats_policies"
}

func (r *csmThreatsPoliciesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state csmThreatsPoliciesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, _, err := r.api.ListCSMThreatsAgentPolicies(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching agent rules"))
		return
	}

	data := res.GetData()
	policyIds := make([]string, len(data))
	policies := make([]csmThreatsPolicyModel, len(data))

	for idx, policy := range res.GetData() {
		var policyModel csmThreatsPolicyModel
		policyModel.Id = types.StringValue(policy.GetId())
		attributes := policy.Attributes
		policyModel.Name = types.StringValue(attributes.GetName())
		policyModel.Description = types.StringValue(attributes.GetDescription())
		policyModel.Enabled = types.BoolValue(attributes.GetEnabled())
		policyModel.Tags, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetHostTags())
		policyIds[idx] = policy.GetId()
		policies[idx] = policyModel
	}

	stateId := strings.Join(policyIds, "--")
	state.Id = types.StringValue(computeDataSourceID(&stateId))
	tfAgentRuleIds, diags := types.ListValueFrom(ctx, types.StringType, policyIds)
	response.Diagnostics.Append(diags...)
	state.PolicyIds = tfAgentRuleIds
	state.Policies = policies

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (*csmThreatsPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing policies.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"policy_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of IDs for the policies.",
				ElementType: types.StringType,
			},
			"policies": schema.ListAttribute{
				Computed:    true,
				Description: "List of policies",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"tags":        types.SetType{ElemType: types.StringType},
						"name":        types.StringType,
						"description": types.StringType,
						"enabled":     types.BoolType,
					},
				},
			},
		},
	}
}
