package fwprovider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

type csmThreatsAgentRuleModel struct {
	Id          types.String  `tfsdk:"id"`
	PolicyId    types.String  `tfsdk:"policy_id"`
	Name        types.String  `tfsdk:"name"`
	Description types.String  `tfsdk:"description"`
	Enabled     types.Bool    `tfsdk:"enabled"`
	Expression  types.String  `tfsdk:"expression"`
	ProductTags types.Set     `tfsdk:"product_tags"`
	Actions     []ActionModel `tfsdk:"actions"`
}

type ActionModel struct {
	Set  *SetActionModel  `tfsdk:"set"`
	Hash *HashActionModel `tfsdk:"hash"`
}

type SetActionModel struct {
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	Field  types.String `tfsdk:"field"`
	Append types.Bool   `tfsdk:"append"`
	Size   types.Int64  `tfsdk:"size"`
	Ttl    types.Int64  `tfsdk:"ttl"`
	Scope  types.String `tfsdk:"scope"`
}

type HashActionModel struct {
	// empty on purpose, has no attributes
}

func NewCSMThreatsAgentRuleResource() resource.Resource {
	return &csmThreatsAgentRuleResource{}
}

func (r *csmThreatsAgentRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "csm_threats_agent_rule"
}

func (r *csmThreatsAgentRuleResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetCSMThreatsApiV2()
	r.auth = providerData.Auth
}

func (r *csmThreatsAgentRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Workload Protection (CSM Threats) Agent Rule API resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
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
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates whether the Agent rule is enabled. Must not be used without policy_id.",
				Computed:    true,
			},
			"expression": schema.StringAttribute{
				Required:    true,
				Description: "The SECL expression of the Agent rule",
			},
			"product_tags": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "The list of product tags associated with the rule",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"actions": schema.ListNestedBlock{
				Description: "The list of actions the rule can perform",
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"set": schema.SingleNestedBlock{
							Description: "Set action configuration",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required:    true,
									Description: "The name of the set action",
								},
								"value": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "The value to set",
								},
								"field": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "The field to get the value from",
								},
								"append": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Whether to append to the set",
								},
								"size": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Description: "The maximum size of the set",
								},
								"ttl": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Description: "The time to live for the set in nanoseconds",
								},
								"scope": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "The scope of the set action (process, container, cgroup, or empty)",
								},
							},
						},
						"hash": schema.SingleNestedBlock{
							Description: "Hash action configuration",
							Attributes:  map[string]schema.Attribute{
								// empty on purpose, has no attributes
							},
						},
					},
				},
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
}

// validateActions validates the actions list
func (r *csmThreatsAgentRuleResource) validateActions(_ context.Context, actions []ActionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for i, action := range actions {
		// Check that exactly one action type is set
		hasSet := action.Set != nil
		hasHash := action.Hash != nil

		if !hasSet && !hasHash {
			diags.AddError(
				"Missing Action Type",
				fmt.Sprintf("Action %d: At least one action type (set or hash) must be specified.", i),
			)
			continue
		}

		// Validate set action if present
		if hasSet {
			// Check that set name is provided and not empty
			if action.Set.Name.IsNull() || action.Set.Name.IsUnknown() || action.Set.Name.ValueString() == "" {
				diags.AddError(
					"Missing Required Field",
					fmt.Sprintf("Action %d: 'name' is required in the set action configuration.", i),
				)
				continue
			}

			// Check that exactly one of value, field is set
			hasValue := !action.Set.Value.IsNull() && !action.Set.Value.IsUnknown() && action.Set.Value.ValueString() != ""
			hasField := !action.Set.Field.IsNull() && !action.Set.Field.IsUnknown() && action.Set.Field.ValueString() != ""

			if !hasValue && !hasField {
				diags.AddError(
					"Missing Required Field",
					fmt.Sprintf("Action %d: One of 'value' or 'field' must be set in the set action configuration.", i),
				)
				continue
			}

			if hasValue && hasField {
				diags.AddError(
					"Invalid Configuration",
					fmt.Sprintf("Action %d: Only one of 'value' or 'field' can be set in the set action configuration.", i),
				)
				continue
			}

			// Validate scope if set
			if !action.Set.Scope.IsNull() && !action.Set.Scope.IsUnknown() {
				scope := action.Set.Scope.ValueString()
				if scope != "" && scope != "process" && scope != "container" && scope != "cgroup" {
					diags.AddError(
						"Invalid Configuration",
						fmt.Sprintf("Action %d: 'scope' must be one of: 'process', 'container', 'cgroup', or empty.", i),
					)
					continue
				}
			}
		}

		// Hash action validation (currently no specific validation needed)
		// if hasHash {
		//     // Add hash-specific validation here if needed in the future
		// }
	}

	return diags
}

func (r *csmThreatsAgentRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate actions
	response.Diagnostics.Append(r.validateActions(ctx, state.Actions)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	agentRulePayload, err := r.buildCreateCSMThreatsAgentRulePayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
		return
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

func (r *csmThreatsAgentRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

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

	if err != nil {
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

	// Validate actions
	response.Diagnostics.Append(r.validateActions(ctx, state.Actions)...)
	if response.Diagnostics.HasError() {
		return
	}

	csmThreatsMutex.Lock()
	defer csmThreatsMutex.Unlock()

	agentRulePayload, err := r.buildUpdateCSMThreatsAgentRulePayload(&state)
	if err != nil {
		response.Diagnostics.AddError("error while parsing resource", err.Error())
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

	r.updateStateFromResponse(ctx, &state, &res)
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

	if err != nil {
		if httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting agent rule"))
		return
	}
}

func (r *csmThreatsAgentRuleResource) buildCreateCSMThreatsAgentRulePayload(state *csmThreatsAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest, error) {
	_, policyId, name, description, enabled, expression, productTags := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleCreateAttributes{}
	attributes.Expression = expression
	attributes.Name = name
	attributes.Description = description
	attributes.Enabled = &enabled
	attributes.PolicyId = policyId
	attributes.ProductTags = productTags

	// Initialize empty actions array - this ensures we always send the actions field
	outActions := make([]datadogV2.CloudWorkloadSecurityAgentRuleAction, 0)

	// Only populate actions if there are any configured
	if state.Actions != nil {
		for _, a := range state.Actions {
			action := datadogV2.CloudWorkloadSecurityAgentRuleAction{}

			if a.Set != nil {
				sa := datadogV2.CloudWorkloadSecurityAgentRuleActionSet{}

				if !a.Set.Name.IsNull() && !a.Set.Name.IsUnknown() {
					name := a.Set.Name.ValueString()
					sa.Name = &name
				}
				if !a.Set.Field.IsNull() && !a.Set.Field.IsUnknown() {
					field := a.Set.Field.ValueString()
					sa.Field = &field
				}
				if !a.Set.Value.IsNull() && !a.Set.Value.IsUnknown() {
					value := a.Set.Value.ValueString()
					sa.Value = &value
				}
				if !a.Set.Append.IsNull() && !a.Set.Append.IsUnknown() {
					append := a.Set.Append.ValueBool()
					sa.Append = &append
				}
				if !a.Set.Size.IsNull() && !a.Set.Size.IsUnknown() {
					size := a.Set.Size.ValueInt64()
					sa.Size = &size
				}
				if !a.Set.Ttl.IsNull() && !a.Set.Ttl.IsUnknown() {
					ttl := a.Set.Ttl.ValueInt64()
					sa.Ttl = &ttl
				}
				if !a.Set.Scope.IsNull() && !a.Set.Scope.IsUnknown() {
					scope := a.Set.Scope.ValueString()
					sa.Scope = &scope
				}
				action.Set = &sa
			}

			if a.Hash != nil {
				ha := make(map[string]interface{})
				action.Hash = ha
			}

			outActions = append(outActions, action)
		}
	}

	// Always set actions field - empty slice will remove all actions from API
	attributes.Actions = outActions

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleCreateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	return datadogV2.NewCloudWorkloadSecurityAgentRuleCreateRequest(*data), nil
}

func (r *csmThreatsAgentRuleResource) buildUpdateCSMThreatsAgentRulePayload(state *csmThreatsAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest, error) {
	agentRuleId, policyId, _, description, enabled, expression, productTags := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleUpdateAttributes{}
	attributes.Expression = &expression
	attributes.Description = description
	attributes.Enabled = &enabled
	attributes.PolicyId = policyId
	attributes.ProductTags = productTags

	// Initialize empty actions array - this ensures we always send the actions field
	outActions := make([]datadogV2.CloudWorkloadSecurityAgentRuleAction, 0)

	// Only populate actions if there are any configured
	if state.Actions != nil {
		for _, a := range state.Actions {
			action := datadogV2.CloudWorkloadSecurityAgentRuleAction{}

			if a.Set != nil {
				sa := datadogV2.CloudWorkloadSecurityAgentRuleActionSet{}

				if !a.Set.Name.IsNull() && !a.Set.Name.IsUnknown() {
					name := a.Set.Name.ValueString()
					sa.Name = &name
				}
				if !a.Set.Field.IsNull() && !a.Set.Field.IsUnknown() {
					field := a.Set.Field.ValueString()
					sa.Field = &field
				}
				if !a.Set.Value.IsNull() && !a.Set.Value.IsUnknown() {
					value := a.Set.Value.ValueString()
					sa.Value = &value
				}
				if !a.Set.Append.IsNull() && !a.Set.Append.IsUnknown() {
					append := a.Set.Append.ValueBool()
					sa.Append = &append
				}
				if !a.Set.Size.IsNull() && !a.Set.Size.IsUnknown() {
					size := a.Set.Size.ValueInt64()
					sa.Size = &size
				}
				if !a.Set.Ttl.IsNull() && !a.Set.Ttl.IsUnknown() {
					ttl := a.Set.Ttl.ValueInt64()
					sa.Ttl = &ttl
				}
				if !a.Set.Scope.IsNull() && !a.Set.Scope.IsUnknown() {
					scope := a.Set.Scope.ValueString()
					sa.Scope = &scope
				}
				action.Set = &sa
			}

			if a.Hash != nil {
				ha := make(map[string]interface{})
				action.Hash = ha
			}

			outActions = append(outActions, action)
		}
	}

	// Always set actions field - empty slice will remove all actions from API
	attributes.Actions = outActions

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	data.Id = &agentRuleId
	return datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateRequest(*data), nil
}

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

	return id, policyId, name, description, enabled, expression, productTags
}

func (r *csmThreatsAgentRuleResource) updateStateFromResponse(ctx context.Context, state *csmThreatsAgentRuleModel, res *datadogV2.CloudWorkloadSecurityAgentRuleResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Name = types.StringValue(attributes.GetName())
	if attributes.Description != nil {
		state.Description = types.StringValue(*attributes.Description)
	} else {
		state.Description = types.StringNull()
	}
	if attributes.Enabled != nil {
		state.Enabled = types.BoolValue(*attributes.Enabled)
	} else {
		state.Enabled = types.BoolNull()
	}
	state.Expression = types.StringValue(attributes.GetExpression())

	tags := attributes.GetProductTags()
	if len(tags) > 0 {
		state.ProductTags, _ = types.SetValueFrom(ctx, types.StringType, tags)
	} else {
		state.ProductTags = types.SetNull(types.StringType)
	}

	var actions []ActionModel
	for _, act := range res.Data.Attributes.GetActions() {
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
			// Handle append with proper default when not returned by API
			if s.Append != nil {
				setAction.Append = types.BoolValue(*s.Append)
			} else {
				// Use false as default when API doesn't return append value
				setAction.Append = types.BoolValue(false)
			}
			if s.Size != nil {
				setAction.Size = types.Int64Value(*s.Size)
			} else {
				// Use 0 as default when API doesn't return size value
				setAction.Size = types.Int64Value(0)
			}
			if s.Ttl != nil {
				setAction.Ttl = types.Int64Value(*s.Ttl)
			} else {
				// Use 0 as default when API doesn't return ttl value
				setAction.Ttl = types.Int64Value(0)
			}
			if s.Scope != nil {
				setAction.Scope = types.StringValue(*s.Scope)
			} else {
				// Use empty string as default when API returns null for scope
				setAction.Scope = types.StringValue("")
			}
			action.Set = setAction
		}

		if act.Hash != nil {
			action.Hash = &HashActionModel{}
		}

		actions = append(actions, action)
	}

	state.Actions = actions
}
