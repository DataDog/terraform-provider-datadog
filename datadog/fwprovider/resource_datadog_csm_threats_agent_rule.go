package fwprovider

import (
	"context"
	"strings"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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
	Id          types.String           `tfsdk:"id"`
	PolicyId    types.String           `tfsdk:"policy_id"`
	Name        types.String           `tfsdk:"name"`
	Description types.String           `tfsdk:"description"`
	Enabled     types.Bool             `tfsdk:"enabled"`
	Expression  types.String           `tfsdk:"expression"`
	ProductTags types.Set              `tfsdk:"product_tags"`
	Actions     []agentRuleActionModel `tfsdk:"actions"`
}

type agentRuleActionModel struct {
	Filter   types.String                  `tfsdk:"filter"`
	Set      *agentRuleActionSetModel      `tfsdk:"set"`
	Kill     *agentRuleActionKillModel     `tfsdk:"kill"`
	Metadata *agentRuleActionMetadataModel `tfsdk:"metadata"`
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

type agentRuleActionKillModel struct {
	Signal types.String `tfsdk:"signal"`
}

type agentRuleActionMetadataModel struct {
	ShortImage types.String `tfsdk:"short_image"`
	ImageTag   types.String `tfsdk:"image_tag"`
	Service    types.String `tfsdk:"service"`
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
			"actions": schema.ListNestedAttribute{
				Optional:    true,
				Description: "The list of actions the rule can perform",
				Attributes: map[string]schema.Attribute{
					"filter": schema.StringAttribute{
						Optional:    true,
						Description: "SECL expression used to target the container",
					},
					"set": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Set action",
						Attributes: map[string]schema.Attribute{
							"name":   schema.StringAttribute{Required: true},
							"field":  schema.StringAttribute{Optional: true},
							"value":  schema.StringAttribute{Optional: true},
							"append": schema.BoolAttribute{Optional: true},
							"size":   schema.Int64Attribute{Optional: true},
							"ttl":    schema.Int64Attribute{Optional: true},
							"scope":  schema.StringAttribute{Optional: true},
						},
					},
					"kill": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Kill system call action",
						Attributes: map[string]schema.Attribute{
							"signal": schema.StringAttribute{Optional: true},
						},
					},
					"metadata": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "Metadata action",
						Attributes: map[string]schema.Attribute{
							"short_image": schema.StringAttribute{Optional: true},
							"image_tag":   schema.StringAttribute{Optional: true},
							"service":     schema.StringAttribute{Optional: true},
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

func (r *csmThreatsAgentRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate that at least one of field or value is set
	if !state.Actions.IsNull() && !state.Actions.IsUnknown() {
		for i, a := range state.Actions {
			if a.Set != nil {
				hasValue := !a.Set.Value.IsNull() && !a.Set.Value.IsUnknown()
				hasField := !a.Set.Field.IsNull() && !a.Set.Field.IsUnknown()
				if !hasValue && !hasField {
					response.Diagnostics.AddAttributeError(
						path.Root("actions").Index(i).AtName("set"),
						"Missing Required Field",
						"One of `value` or `field` must be specified.",
					)
					return
				}

				// Check that only one of field or value is set
				if hasField && hasValue {
					response.Diagnostics.AddAttributeError(
						path.Root("actions").Index(i).AtName("set"),
						"Invalid Configuration",
						"Only one of `field` or `value` can be specified.",
					)
					return
				}
			
			if !hasField && !hasValue {
				response.Diagnostics.AddError(
					"Missing Required Field",
					"At least one of 'field' or 'value' must be set in the action set configuration.",
				)
				return
			}

			// Check that only one of field or value is set
			if hasField && hasValue {
				response.Diagnostics.AddError(
					"Invalid Configuration",
					"Only one of 'field' or 'value' can be set in the action set configuration.",
				)
				return
			}
		}
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

	// Validate that at least one of field or value is set
	if !state.Actions.IsNull() && !state.Actions.IsUnknown() {
		for _, actionElement := range state.Actions.Elements() {
			if actionObj, ok := actionElement.(types.Object); ok {
				actionAttrs := actionObj.Attributes()
				if setObj, ok := actionAttrs["set"].(types.Object); ok {
					setAttrs := setObj.Attributes()

					// Check if name is set
					if name, ok := setAttrs["name"].(types.String); !ok || name.IsNull() || name.IsUnknown() {
						response.Diagnostics.AddError(
							"Missing Required Field",
							"The 'name' field is required in the action set configuration.",
						)
						return
					}

					// Check if at least one of field or value is set
					field, fieldOk := setAttrs["field"].(types.String)
					value, valueOk := setAttrs["value"].(types.String)

					hasField := fieldOk && !field.IsNull() && !field.IsUnknown()
					hasValue := valueOk && !value.IsNull() && !value.IsUnknown()

					if !hasField && !hasValue {
						response.Diagnostics.AddError(
							"Missing Required Field",
							"At least one of 'field' or 'value' must be set in the action set configuration.",
						)
						return
					}

					// Check that only one of field or value is set
					if hasField && hasValue {
						response.Diagnostics.AddError(
							"Invalid Configuration",
							"Only one of 'field' or 'value' can be set in the action set configuration.",
						)
						return
					}
				}
			}
		}
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

	var outActions []datadogV2.CloudWorkloadSecurityAgentRuleAction
	for _, a := range state.Actions {
		if a.Set == nil {
			continue
		}
		s := a.Set
		sa := datadogV2.CloudWorkloadSecurityAgentRuleActionSet{
			Name: &s.Name.ValueString(),
		}
		if !s.Field.IsNull() {
			f := s.Field.ValueString()
			sa.Field = &f
		}
		if !s.Value.IsNull() {
			v := s.Value.ValueString()
			sa.Value = &v
		}
		if !s.Append.IsNull() {
			b := s.Append.ValueBool()
			sa.Append = &b
		}
		if !s.Size.IsNull() {
			i := s.Size.ValueInt64()
			sa.Size = &i
		}
		if !s.Ttl.IsNull() {
			t := s.Ttl.ValueInt64()
			sa.Ttl = &t
		}
		if !s.Scope.IsNull() {
			sc := s.Scope.ValueString()
			sa.Scope = &sc
		}
		outActions = append(outActions, datadogV2.CloudWorkloadSecurityAgentRuleAction{Set: &sa})
	}

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

	var outActions []datadogV2.CloudWorkloadSecurityAgentRuleAction

	for _, a := range state.Actions {
		if a.Set == nil {
			continue
		}
		s := a.Set
		sa := datadogV2.CloudWorkloadSecurityAgentRuleActionSet{
			Name: &s.Name.ValueString(),
		}
		if !s.Field.IsNull() {
			f := s.Field.ValueString()
			sa.Field = &f
		}
		if !s.Value.IsNull() {
			v := s.Value.ValueString()
			sa.Value = &v
		}
		if !s.Append.IsNull() {
			b := s.Append.ValueBool()
			sa.Append = &b
		}
		if !s.Size.IsNull() {
			i := s.Size.ValueInt64()
			sa.Size = &i
		}
		if !s.Ttl.IsNull() {
			t := s.Ttl.ValueInt64()
			sa.Ttl = &t
		}
		if !s.Scope.IsNull() {
			sc := s.Scope.ValueString()
			sa.Scope = &sc
		}
		outActions = append(outActions, datadogV2.CloudWorkloadSecurityAgentRuleAction{Set: &sa})
	}

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

	var stActions []agentRuleActionModel
	for _, act := range res.Data.Attributes.GetActions() {
		if act.GetSet() == nil {
			continue
		}
		set := act.GetSet()
		stActions = append(stActions, agentRuleActionModel{
			Set: &agentRuleActionSetModel{
				Name:   types.StringValue(set.GetName()),
				Field:  types.StringValue(set.GetField()),
				Value:  types.StringValue(set.GetValue()),
				Append: types.BoolValue(set.GetAppend()),
				Size:   types.Int64Value(set.GetSize()),
				Ttl:    types.Int64Value(set.GetTtl()),
				Scope:  types.StringValue(set.GetScope()),
			},
		})
	}
	state.Actions = stActions
}
