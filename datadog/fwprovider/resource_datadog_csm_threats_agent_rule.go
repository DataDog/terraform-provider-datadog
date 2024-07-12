package fwprovider

import (
	"context"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	csmThreatsMutex sync.Mutex
	_               resource.ResourceWithConfigure   = &csmThreatsAgentRuleResource{}
	_               resource.ResourceWithImportState = &csmThreatsAgentRuleResource{}
)

type csmThreatsAgentRuleModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Expression  types.String `tfsdk:"expression"`
}

type csmThreatsAgentRuleResource struct {
	api  *datadogV2.CSMThreatsApi
	auth context.Context
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
				Default:     stringdefault.StaticString(""),
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

func (r *csmThreatsAgentRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
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

	r.updateStateFromResponse(ctx, &state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *csmThreatsAgentRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state csmThreatsAgentRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	agentRuleId := state.Id.ValueString()
	res, httpResponse, err := r.api.GetCSMThreatsAgentRule(r.auth, agentRuleId)
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

	httpResp, err := r.api.DeleteCSMThreatsAgentRule(r.auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting agent rule"))
		return
	}
}

func (r *csmThreatsAgentRuleResource) buildCreateCSMThreatsAgentRulePayload(state *csmThreatsAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleCreateRequest, error) {
	_, name, description, enabled, expression := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleCreateAttributes{}
	attributes.Expression = expression
	attributes.Name = name
	attributes.Description = description
	attributes.Enabled = &enabled

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleCreateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	return datadogV2.NewCloudWorkloadSecurityAgentRuleCreateRequest(*data), nil
}

func (r *csmThreatsAgentRuleResource) buildUpdateCSMThreatsAgentRulePayload(state *csmThreatsAgentRuleModel) (*datadogV2.CloudWorkloadSecurityAgentRuleUpdateRequest, error) {
	agentRuleId, _, description, enabled, _ := r.extractAgentRuleAttributesFromResource(state)

	attributes := datadogV2.CloudWorkloadSecurityAgentRuleUpdateAttributes{}
	attributes.Description = description
	attributes.Enabled = &enabled

	data := datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateData(attributes, datadogV2.CLOUDWORKLOADSECURITYAGENTRULETYPE_AGENT_RULE)
	data.Id = &agentRuleId
	return datadogV2.NewCloudWorkloadSecurityAgentRuleUpdateRequest(*data), nil
}

func (r *csmThreatsAgentRuleResource) extractAgentRuleAttributesFromResource(state *csmThreatsAgentRuleModel) (string, string, *string, bool, string) {
	// Mandatory fields
	id := state.Id.ValueString()
	name := state.Name.ValueString()
	enabled := state.Enabled.ValueBool()
	expression := state.Expression.ValueString()
	description := state.Description.ValueStringPointer()

	return id, name, description, enabled, expression
}

func (r *csmThreatsAgentRuleResource) updateStateFromResponse(ctx context.Context, state *csmThreatsAgentRuleModel, res *datadogV2.CloudWorkloadSecurityAgentRuleResponse) {
	state.Id = types.StringValue(res.Data.GetId())

	attributes := res.Data.Attributes

	state.Name = types.StringValue(attributes.GetName())
	state.Description = types.StringValue(attributes.GetDescription())
	state.Enabled = types.BoolValue(attributes.GetEnabled())
	state.Expression = types.StringValue(attributes.GetExpression())
}
