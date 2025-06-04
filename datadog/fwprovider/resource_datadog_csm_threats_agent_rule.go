package fwprovider

import (
	"context"
	"strings"
	"strings"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"net/http"

	"net/http"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	csmThreatsMutex sync.Mutex
	_               resource.ResourceWithConfigure   = &csmThreatsAgentRuleResource{}
	_               resource.ResourceWithImportState = &csmThreatsAgentRuleResource{}
)

type csmThreatsAgentRuleResource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsAgentRuleResource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
}

type csmThreatsAgentRuleModel struct {
	Id          types.String `tfsdk:"id"`
	PolicyId    types.String `tfsdk:"policy_id"`
	PolicyId    types.String `tfsdk:"policy_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Expression  types.String `tfsdk:"expression"`
	ProductTags types.Set    `tfsdk:"product_tags"`
	Actions     types.List   `tfsdk:"actions"`
}

type agentRuleActionSetModel struct {
	Name   types.String `tfsdk:"name"`
	Field  types.String `tfsdk:"field"`
	Value  types.String `tfsdk:"value"`
	Append types.Bool   `tfsdk:"append"`
	Size   types.Int64  `tfsdk:"size"`
	Ttl    types.Int64  `tfsdk:"ttl"`
	Scope  types.String `tfsdk:"scope"`
}

type agentRuleActionModel struct {
	Set agentRuleActionSetModel `tfsdk:"set"`
}

func NewCSMThreatsAgentRuleResource() resource.Resource {
	return &csmThreatsAgentRuleResource{}
}

func (r *csmThreatsAgentRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "csm_threats_agent_rule"
}

func (r *csmThreatsAgentRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (r *csmThreatsAgentRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog CSM Threats Agent Rule API resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"policy_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the agent policy in which the rule is saved",
			},
			"policy_id": schema.StringAttribute{
				Optional:    true,
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
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates whether the Agent rule is enabled. Must not be used without policy_id.",
			},
			"expression": schema.StringAttribute{
				Required:    true,
				Description: "The SECL expression of the Agent rule",
			},
			"actions": schema.ListAttribute{
				Optional:    true,
				Description: "The list of actions the rule can perform if triggered",
				ElementType: types.ObjectType{AttrTypes: map[string]attr.Type{
					"set": types.ObjectType{AttrTypes: map[string]attr.Type{
						"name":   types.StringType,
						"field":  types.StringType,
						"value":  types.StringType,
						"append": types.BoolType,
						"size":   types.Int64Type,
						"ttl":    types.Int64Type,
						"scope":  types.StringType,
					}},
				}},
			},
			"product_tags": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "The list of product tags associated with the rule",
			},
		},
	}
}

func (r *csmThreatsAgentRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)

	if len(result) == 2 {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_id"), result[0])...)
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
	} else if len(result) == 1 {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[0])...)
	} else {
		response.Diagnostics.AddError("unexpected import format", "expected '<policy_id>:<rule_id>' or '<rule_id>'")
	}
	result := strings.SplitN(request.ID, ":", 2)

	if len(result) == 2 {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_id"), result[0])...)
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
	} else if len(result) == 1 {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[0])...)
	} else {
		response.Diagnostics.AddError("unexpected import format", "expected '<policy_id>:<rule_id>' or '<rule_id>'")
	}
}

func (r *csmThreatsAgentRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state csmThreatsAgentRuleModel
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

	// Update essential fields from API response
	state.Id = types.StringValue(res.Data.GetId())
	attributes := res.Data.Attributes
	state.Name = types.StringValue(attributes.GetName())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Expression = types.StringValue(attributes.GetExpression())

	// Handle description - preserve from plan if it was null and API returns empty
	var planState csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &planState)...)
	if !response.Diagnostics.HasError() {
		description := attributes.GetDescription()
		if description == "" && planState.Description.IsNull() {
			state.Description = types.StringNull()
		} else {
			state.Description = types.StringValue(description)
		}

		// Preserve actions from the plan since API may return stale data or missing optional fields
		state.Actions = planState.Actions
	}

	// Handle product tags
	tags := attributes.GetProductTags()
	if len(tags) == 0 && state.ProductTags.IsNull() {
		state.ProductTags = types.SetNull(types.StringType)
	} else {
		state.ProductTags, _ = types.SetValueFrom(ctx, types.StringType, tags)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsAgentRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	agentRuleId := state.Id.ValueString()

	var res datadogV2.CloudWorkloadSecurityAgentRuleResponse
	var httpResp *http.Response
	var err error
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		policyId := state.PolicyId.ValueString()
		res, httpResp, err = r.api.GetCSMThreatsAgentRule(r.auth, agentRuleId, *datadogV2.NewGetCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
	} else {
		res, httpResp, err = r.api.GetCSMThreatsAgentRule(r.auth, agentRuleId)
	}


	var res datadogV2.CloudWorkloadSecurityAgentRuleResponse
	var httpResp *http.Response
	var err error
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		policyId := state.PolicyId.ValueString()
		res, httpResp, err = r.api.GetCSMThreatsAgentRule(r.auth, agentRuleId, *datadogV2.NewGetCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
	} else {
		res, httpResp, err = r.api.GetCSMThreatsAgentRule(r.auth, agentRuleId)
	}

	if err != nil {
		if httpResp.StatusCode == 404 {
		if httpResp.StatusCode == 404 {
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

func (r *csmThreatsAgentRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	agentRulePayload, err := r.buildUpdateCSMThreatsAgentRulePayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
		return
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

	// Update essential fields from API response
	attributes := res.Data.Attributes
	state.Name = types.StringValue(attributes.GetName())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Expression = types.StringValue(attributes.GetExpression())

	// Handle description - preserve from plan if it was null and API returns empty
	var planState csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &planState)...)
	if !response.Diagnostics.HasError() {
		description := attributes.GetDescription()
		if description == "" && planState.Description.IsNull() {
			state.Description = types.StringNull()
		} else {
			state.Description = types.StringValue(description)
		}

		// Preserve actions from the plan since API may return stale data or missing optional fields
		state.Actions = planState.Actions
	}

	// Handle product tags
	tags := attributes.GetProductTags()
	if len(tags) == 0 && state.ProductTags.IsNull() {
		state.ProductTags = types.SetNull(types.StringType)
	} else {
		state.ProductTags, _ = types.SetValueFrom(ctx, types.StringType, tags)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsAgentRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	id := state.Id.ValueString()

	var httpResp *http.Response
	var err error
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		policyId := state.PolicyId.ValueString()
		httpResp, err = r.api.DeleteCSMThreatsAgentRule(r.auth, id, *datadogV2.NewDeleteCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
	} else {
		httpResp, err = r.api.DeleteCSMThreatsAgentRule(r.auth, id)
	}

	var httpResp *http.Response
	var err error
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		policyId := state.PolicyId.ValueString()
		httpResp, err = r.api.DeleteCSMThreatsAgentRule(r.auth, id, *datadogV2.NewDeleteCSMThreatsAgentRuleOptionalParameters().WithPolicyId(policyId))
	} else {
		httpResp, err = r.api.DeleteCSMThreatsAgentRule(r.auth, id)
	}

	if err != nil {
		if httpResp.StatusCode == 404 {
		if httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting agent rule"))
		return
	}
}

func (r *csmThreatsAgentRuleResource) buildCreateCSMThreatsAgentRulePayload(state *csmThreatsAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest, error) {
	_, policyId, name, description, enabled, expression, productTags := r.extractAgentRuleAttributesFromResource(state)
	_, policyId, name, description, enabled, expression, productTags := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleCreateAttributes{}
	attributes.Expression = expression
	attributes.Name = name
	attributes.Description = description
	attributes.Enabled = &enabled
	attributes.PolicyId = policyId
	attributes.ProductTags = productTags

	if !state.Actions.IsNull() && !state.Actions.IsUnknown() {
		var actions []agentRuleActionModel
		state.Actions.ElementsAs(context.Background(), &actions, false)
		if len(actions) > 0 {
			apiActions := make([]datadogV2.CloudWorkloadSecurityAgentRuleAction, len(actions))
			for i, action := range actions {
				setAction := datadogV2.NewCloudWorkloadSecurityAgentRuleActionSet()
				setAction.SetName(action.Set.Name.ValueString())

				// Handle optional fields - only set if not null and not empty
				if !action.Set.Field.IsNull() && action.Set.Field.ValueString() != "" {
					setAction.SetField(action.Set.Field.ValueString())
				}
				if !action.Set.Value.IsNull() && action.Set.Value.ValueString() != "" {
					setAction.SetValue(action.Set.Value.ValueString())
				}
				if !action.Set.Append.IsNull() {
					setAction.SetAppend(action.Set.Append.ValueBool())
				}
				if !action.Set.Size.IsNull() && action.Set.Size.ValueInt64() != 0 {
					setAction.SetSize(action.Set.Size.ValueInt64())
				}
				if !action.Set.Ttl.IsNull() && action.Set.Ttl.ValueInt64() != 0 {
					setAction.SetTtl(action.Set.Ttl.ValueInt64())
				}
				if !action.Set.Scope.IsNull() && action.Set.Scope.ValueString() != "" {
					setAction.SetScope(action.Set.Scope.ValueString())
				}
				ruleAction := datadogV2.NewCloudWorkloadSecurityAgentRuleAction()
				ruleAction.SetSet(*setAction)
				apiActions[i] = *ruleAction
			}
			attributes.SetActions(apiActions)
		}
	}

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleCreateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	return datadogV2.NewCloudWorkloadSecurityAgentRuleCreateRequest(*data), nil
}

func (r *csmThreatsAgentRuleResource) buildUpdateCSMThreatsAgentRulePayload(state *csmThreatsAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest, error) {
	agentRuleId, policyId, _, description, enabled, expression, productTags := r.extractAgentRuleAttributesFromResource(state)
	agentRuleId, policyId, _, description, enabled, expression, productTags := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleUpdateAttributes{}
	attributes.Expression = &expression
	attributes.Expression = &expression
	attributes.Description = description
	attributes.Enabled = &enabled
	attributes.PolicyId = policyId
	attributes.ProductTags = productTags

	// Always process actions to ensure they are properly sent to the API
	var actions []agentRuleActionModel
	if !state.Actions.IsNull() && !state.Actions.IsUnknown() {
		state.Actions.ElementsAs(context.Background(), &actions, false)
	}

	apiActions := make([]datadogV2.CloudWorkloadSecurityAgentRuleAction, len(actions))
	for i, action := range actions {
		setAction := datadogV2.NewCloudWorkloadSecurityAgentRuleActionSet()
		if !action.Set.Name.IsNull() {
			setAction.SetName(action.Set.Name.ValueString())
		}

		// Handle optional fields - only set if not null and not empty
		if !action.Set.Field.IsNull() && action.Set.Field.ValueString() != "" {
			setAction.SetField(action.Set.Field.ValueString())
		}
		if !action.Set.Value.IsNull() && action.Set.Value.ValueString() != "" {
			setAction.SetValue(action.Set.Value.ValueString())
		}
		if !action.Set.Append.IsNull() {
			setAction.SetAppend(action.Set.Append.ValueBool())
		}
		if !action.Set.Size.IsNull() && action.Set.Size.ValueInt64() != 0 {
			setAction.SetSize(action.Set.Size.ValueInt64())
		}
		if !action.Set.Ttl.IsNull() && action.Set.Ttl.ValueInt64() != 0 {
			setAction.SetTtl(action.Set.Ttl.ValueInt64())
		}
		if !action.Set.Scope.IsNull() && action.Set.Scope.ValueString() != "" {
			setAction.SetScope(action.Set.Scope.ValueString())
		}
		ruleAction := datadogV2.NewCloudWorkloadSecurityAgentRuleAction()
		ruleAction.SetSet(*setAction)
		apiActions[i] = *ruleAction
	}
	attributes.Actions = apiActions

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	data.Id = &agentRuleId
	return datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateRequest(*data), nil
}

func (r *csmThreatsAgentRuleResource) extractAgentRuleAttributesFromResource(state *csmThreatsAgentRuleModel) (string, *string, string, *string, bool, string, []string) {
func (r *csmThreatsAgentRuleResource) extractAgentRuleAttributesFromResource(state *csmThreatsAgentRuleModel) (string, *string, string, *string, bool, string, []string) {
	// Mandatory fields
	id := state.Id.ValueString()
	name := state.Name.ValueString()

	// Optional fields
	var policyId *string
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		val := state.PolicyId.ValueString()
		policyId = &val
	}

	// Optional fields
	var policyId *string
	if !state.PolicyId.IsNull() && !state.PolicyId.IsUnknown() {
		val := state.PolicyId.ValueString()
		policyId = &val
	}
	enabled := state.Enabled.ValueBool()
	expression := state.Expression.ValueString()
	description := state.Description.ValueStringPointer()
	var productTags []string
	if !state.ProductTags.IsNull() && !state.ProductTags.IsUnknown() {
		for _, tag := range state.ProductTags.Elements() {
			tagStr, ok := tag.(types.String)
			if !ok {
				return "", nil, "", nil, false, "", nil
			}
			productTags = append(productTags, tagStr.ValueString())
		}
	}
	var productTags []string
	if !state.ProductTags.IsNull() && !state.ProductTags.IsUnknown() {
		for _, tag := range state.ProductTags.Elements() {
			tagStr, ok := tag.(types.String)
			if !ok {
				return "", nil, "", nil, false, "", nil
			}
			productTags = append(productTags, tagStr.ValueString())
		}
	}

	return id, policyId, name, description, enabled, expression, productTags
	return id, policyId, name, description, enabled, expression, productTags
}

func (r *csmThreatsAgentRuleResource) updateStateFromResponse(ctx context.Context, state *csmThreatsAgentRuleModel, res *datadogV2.CloudWorkloadSecurityAgentRuleResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Name = types.StringValue(attributes.GetName())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Expression = types.StringValue(attributes.GetExpression())
	tags := attributes.GetProductTags()
	if len(tags) == 0 && state.ProductTags.IsNull() {
		state.ProductTags = types.SetNull(types.StringType)
	} else {
		state.ProductTags, _ = types.SetValueFrom(ctx, types.StringType, tags)
	}

	actions := attributes.GetActions()
	actionObjects := make([]agentRuleActionModel, 0, len(actions))
	for _, action := range actions {
		if action.Set != nil {
			setModel := agentRuleActionSetModel{
				Name: types.StringValue(action.Set.GetName()),
			}

			// Handle optional fields
			if field, ok := action.Set.GetFieldOk(); ok && field != nil {
				setModel.Field = types.StringValue(*field)
			} else {
				setModel.Field = types.StringNull()
			}

			if value, ok := action.Set.GetValueOk(); ok && value != nil {
				setModel.Value = types.StringValue(*value)
			} else {
				setModel.Value = types.StringNull()
			}

			if append, ok := action.Set.GetAppendOk(); ok && append != nil {
				setModel.Append = types.BoolValue(*append)
			} else {
				setModel.Append = types.BoolNull()
			}

			if size, ok := action.Set.GetSizeOk(); ok && size != nil {
				setModel.Size = types.Int64Value(*size)
			} else {
				setModel.Size = types.Int64Null()
			}

			if ttl, ok := action.Set.GetTtlOk(); ok && ttl != nil {
				setModel.Ttl = types.Int64Value(*ttl)
			} else {
				setModel.Ttl = types.Int64Null()
			}

			if scope, ok := action.Set.GetScopeOk(); ok && scope != nil {
				setModel.Scope = types.StringValue(*scope)
			} else {
				setModel.Scope = types.StringNull()
			}

			actionObjects = append(actionObjects, agentRuleActionModel{
				Set: setModel,
			})
		}
	}

	if len(actionObjects) > 0 {
		state.Actions, _ = types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"set": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":   types.StringType,
						"field":  types.StringType,
						"value":  types.StringType,
						"append": types.BoolType,
						"size":   types.Int64Type,
						"ttl":    types.Int64Type,
						"scope":  types.StringType,
					},
				},
			},
		}, actionObjects)
	} else {
		state.Actions = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"set": types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":   types.StringType,
						"field":  types.StringType,
						"value":  types.StringType,
						"append": types.BoolType,
						"size":   types.Int64Type,
						"ttl":    types.Int64Type,
						"scope":  types.StringType,
					},
				},
			},
		})
	}
}
