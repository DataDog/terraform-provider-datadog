package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &securityFindingsMuteRuleResource{}
	_ resource.ResourceWithImportState = &securityFindingsMuteRuleResource{}
)

type securityFindingsMuteRuleResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityFindingsMuteRuleModel struct {
	ID      types.String              `tfsdk:"id"`
	Name    types.String              `tfsdk:"name"`
	Enabled types.Bool                `tfsdk:"enabled"`
	Rule    *automationRuleScopeModel `tfsdk:"rule"`
	Action  *muteRuleActionModel      `tfsdk:"action"`
}

type muteRuleActionModel struct {
	Reason            types.String `tfsdk:"reason"`
	ReasonDescription types.String `tfsdk:"reason_description"`
	ExpireAt          types.Int64  `tfsdk:"expire_at"`
}

func NewSecurityFindingsMuteRuleResource() resource.Resource {
	return &securityFindingsMuteRuleResource{}
}

func (r *securityFindingsMuteRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsMuteRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_mute_rule"
}

func (r *securityFindingsMuteRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings automation mute rule resource. This can be used to create and manage mute rules that automatically suppress matching security findings. Use the `datadog_security_findings_mute_rules_order` resource to manage the evaluation order of mute rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the mute rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the mute rule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"rule": securityFindingsAutomationRuleScopeBlock(),
			"action": schema.SingleNestedBlock{
				Description: "The action taken when the rule matches a finding.",
				Attributes: map[string]schema.Attribute{
					"reason": schema.StringAttribute{
						Description: "The reason for muting the matched findings.",
						Required:    true,
						Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewMuteReasonFromValue)},
					},
					"reason_description": schema.StringAttribute{
						Description: "An optional description providing more context for the mute reason.",
						Optional:    true,
					},
					"expire_at": schema.Int64Attribute{
						Description: "The Unix timestamp in milliseconds at which the mute expires. If omitted, the mute does not expire.",
						Optional:    true,
					},
				},
				Validators: []validator.Object{objectvalidator.IsRequired()},
			},
		},
	}
}

func (r *securityFindingsMuteRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *securityFindingsMuteRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsMuteRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid mute rule ID", err.Error())
		return
	}

	resp, httpResp, err := r.Api.GetSecurityFindingsAutomationMuteRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving mute rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsMuteRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityFindingsMuteRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewMuteRuleCreateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.CreateSecurityFindingsAutomationMuteRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating mute rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsMuteRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityFindingsMuteRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid mute rule ID", err.Error())
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewMuteRuleUpdateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.UpdateSecurityFindingsAutomationMuteRule(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating mute rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsMuteRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityFindingsMuteRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid mute rule ID", err.Error())
		return
	}

	httpResp, err := r.Api.DeleteSecurityFindingsAutomationMuteRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting mute rule"))
		return
	}
}

func (r *securityFindingsMuteRuleResource) updateState(ctx context.Context, state *securityFindingsMuteRuleModel, resp *datadogV2.MuteRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId().String())
	state.Name = types.StringValue(attributes.GetName())
	state.Enabled = types.BoolValue(attributes.GetEnabled())

	scope, d := flattenAutomationRuleScope(ctx, attributes.GetRule())
	diags.Append(d...)
	state.Rule = scope

	action := attributes.GetAction()
	actionModel := &muteRuleActionModel{
		Reason: types.StringValue(string(action.GetReason())),
	}
	if action.HasReasonDescription() {
		actionModel.ReasonDescription = types.StringValue(action.GetReasonDescription())
	} else {
		actionModel.ReasonDescription = types.StringNull()
	}
	if action.HasExpireAt() {
		actionModel.ExpireAt = types.Int64Value(action.GetExpireAt())
	} else {
		actionModel.ExpireAt = types.Int64Null()
	}
	state.Action = actionModel

	return diags
}

// buildRuleData builds the JSON:API data object shared by the create and update requests.
func (r *securityFindingsMuteRuleResource) buildRuleData(ctx context.Context, state *securityFindingsMuteRuleModel) (*datadogV2.MuteRuleDataCreate, diag.Diagnostics) {
	var diags diag.Diagnostics

	scope, d := buildAutomationRuleScope(ctx, state.Rule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	action := datadogV2.NewMuteRuleActionWithDefaults()
	action.SetReason(datadogV2.MuteReason(state.Action.Reason.ValueString()))
	if !state.Action.ReasonDescription.IsNull() && !state.Action.ReasonDescription.IsUnknown() {
		action.SetReasonDescription(state.Action.ReasonDescription.ValueString())
	}
	if !state.Action.ExpireAt.IsNull() && !state.Action.ExpireAt.IsUnknown() {
		action.SetExpireAt(state.Action.ExpireAt.ValueInt64())
	}

	attributes := datadogV2.NewMuteRuleAttributesCreateWithDefaults()
	attributes.SetName(state.Name.ValueString())
	attributes.SetEnabled(state.Enabled.ValueBool())
	attributes.SetRule(*scope)
	attributes.SetAction(*action)

	data := datadogV2.NewMuteRuleDataCreateWithDefaults()
	data.SetType(datadogV2.MUTERULETYPE_MUTE_RULES)
	data.SetAttributes(*attributes)
	return data, diags
}
