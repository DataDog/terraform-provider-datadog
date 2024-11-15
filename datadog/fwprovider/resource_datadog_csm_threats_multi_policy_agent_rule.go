package fwprovider

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type csmThreatsMultiPolicyAgentRuleResource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsMultiPolicyAgentRuleModel struct {
	Id          types.String `tfsdk:"id"`
	PolicyId    types.String `tfsdk:"policy_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Expression  types.String `tfsdk:"expression"`
}

func NewCSMThreatsMultiPolicyAgentRuleResource() resource.Resource {
	return &csmThreatsMultiPolicyAgentRuleResource{}
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "csm_threats_multi_policy_agent_rule"
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog CSM Threats Agent Rule API resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"policy_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the agent policy in which the rule is saved",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Agent rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description for the Agent rule.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Indicates Whether the Agent rule is enabled.",
			},
			"expression": schema.StringAttribute{
				Required:    true,
				Description: "The SECL expression of the Agent rule",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *csmThreatsMultiPolicyAgentRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving policy_id or rule_id from given ID", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state csmThreatsMultiPolicyAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	agentRulePayload, err := r.buildCreateCSMThreatsAgentRulePayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
	}

	res, _, err := r.api.CreateCSMThreatsAgentRule(r.auth, *agentRulePayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating agent rule"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsMultiPolicyAgentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	agentRuleId := state.Id.ValueString()
	policyId := state.PolicyId.ValueString()
	res, httpResponse, err := r.api.GetCSMThreatsAgentRule(r.auth, agentRuleId, *datadogV2.NewGetCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching agent rule"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state csmThreatsMultiPolicyAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	agentRulePayload, err := r.buildUpdateCSMThreatsAgentRulePayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
	}

	res, _, err := r.api.UpdateCSMThreatsAgentRule(r.auth, state.Id.ValueString(), *agentRulePayload)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating agent rule"))
		return
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsMultiPolicyAgentRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state csmThreatsMultiPolicyAgentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	id := state.Id.ValueString()
	policyId := state.PolicyId.ValueString()
	httpResp, err := r.api.DeleteCSMThreatsAgentRule(r.auth, id, *datadogV2.NewDeleteCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting agent rule"))
		return
	}
}

func (r *csmThreatsMultiPolicyAgentRuleResource) buildCreateCSMThreatsAgentRulePayload(state *csmThreatsMultiPolicyAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest, error) {
	_, policyId, name, description, enabled, expression := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleCreateAttributes{}
	attributes.Expression = expression
	attributes.Name = name
	attributes.Description = description
	attributes.Enabled = &enabled
	attributes.PolicyId = &policyId

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleCreateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	return datadogV2.NewCloudWorkloadSecurityAgentRuleCreateRequest(*data), nil
}

func (r *csmThreatsMultiPolicyAgentRuleResource) buildUpdateCSMThreatsAgentRulePayload(state *csmThreatsMultiPolicyAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest, error) {
	agentRuleId, policyId, _, description, enabled, _ := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleUpdateAttributes{}
	attributes.Description = description
	attributes.Enabled = &enabled
	attributes.PolicyId = &policyId

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	data.Id = &agentRuleId
	return datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateRequest(*data), nil
}

func (r *csmThreatsMultiPolicyAgentRuleResource) extractAgentRuleAttributesFromResource(state *csmThreatsMultiPolicyAgentRuleModel) (string, string, string, *string, bool, string) {
	// Mandatory fields
	id := state.Id.ValueString()
	policyId := state.PolicyId.ValueString()
	name := state.Name.ValueString()
	enabled := state.Enabled.ValueBool()
	expression := state.Expression.ValueString()
	description := state.Description.ValueStringPointer()

	return id, policyId, name, description, enabled, expression
}

func (r *csmThreatsMultiPolicyAgentRuleResource) updateStateFromResponse(ctx context.Context, state *csmThreatsMultiPolicyAgentRuleModel, res *datadogV2.CloudWorkloadSecurityAgentRuleResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Name = types.StringValue(attributes.GetName())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Expression = types.StringValue(attributes.GetExpression())
}
