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
	_ resource.ResourceWithConfigure   = &securityFindingsSeverityModifierRuleResource{}
	_ resource.ResourceWithImportState = &securityFindingsSeverityModifierRuleResource{}
)

type securityFindingsSeverityModifierRuleResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityFindingsSeverityModifierRuleModel struct {
	ID      types.String                     `tfsdk:"id"`
	Name    types.String                     `tfsdk:"name"`
	Enabled types.Bool                       `tfsdk:"enabled"`
	Rule    *automationRuleScopeModel        `tfsdk:"rule"`
	Action  *severityModifierRuleActionModel `tfsdk:"action"`
}

type severityModifierRuleActionModel struct {
	Set   *severityModifierRuleSetActionModel   `tfsdk:"set"`
	Shift *severityModifierRuleShiftActionModel `tfsdk:"shift"`
}

type severityModifierRuleSetActionModel struct {
	Severity    types.String `tfsdk:"severity"`
	Description types.String `tfsdk:"description"`
}

type severityModifierRuleShiftActionModel struct {
	SeverityDelta types.String `tfsdk:"severity_delta"`
	Description   types.String `tfsdk:"description"`
}

func NewSecurityFindingsSeverityModifierRuleResource() resource.Resource {
	return &securityFindingsSeverityModifierRuleResource{}
}

func (r *securityFindingsSeverityModifierRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsSeverityModifierRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_severity_modifier_rule"
}

func (r *securityFindingsSeverityModifierRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings automation severity modifier rule resource. This can be used to create and manage rules that automatically adjust the severity of matching security findings. Use the `datadog_security_findings_severity_modifier_rules_order` resource to manage the evaluation order of severity modifier rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the severity modifier rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the severity modifier rule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"rule": securityFindingsAutomationRuleScopeAttribute(),
			"action": schema.SingleNestedAttribute{
				Description: "The action to take when the severity modifier rule matches a finding. Exactly one of `set` or `shift` must be provided.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"set": schema.SingleNestedAttribute{
						Description: "Sets matched findings to a fixed severity.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"severity": schema.StringAttribute{
								Description: "The severity to assign to matched findings. `info_none` is not supported for the `iac_misconfiguration`, `runtime_code_vulnerability`, `secret`, or `static_code_vulnerability` finding types.",
								Required:    true,
								Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewSeverityModifierSeverityFromValue)},
							},
							"description": schema.StringAttribute{
								Description: "An optional free-form explanation for the severity change.",
								Optional:    true,
							},
						},
						Validators: []validator.Object{
							objectvalidator.ExactlyOneOf(frameworkPath.MatchRelative().AtParent().AtName("shift")),
						},
					},
					"shift": schema.SingleNestedAttribute{
						Description: "Shifts matched findings up or down by one severity rank.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"severity_delta": schema.StringAttribute{
								Description: "The direction in which to shift the severity of matched findings by one rank.",
								Required:    true,
								Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewSeverityModifierSeverityDeltaFromValue)},
							},
							"description": schema.StringAttribute{
								Description: "An optional free-form explanation for the severity change.",
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *securityFindingsSeverityModifierRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *securityFindingsSeverityModifierRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsSeverityModifierRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid severity modifier rule ID", err.Error())
		return
	}

	resp, httpResp, err := r.Api.GetSecurityFindingsAutomationSeverityModifierRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving severity modifier rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsSeverityModifierRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityFindingsSeverityModifierRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewSeverityModifierRuleCreateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.CreateSecurityFindingsAutomationSeverityModifierRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating severity modifier rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsSeverityModifierRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityFindingsSeverityModifierRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid severity modifier rule ID", err.Error())
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewSeverityModifierRuleUpdateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.UpdateSecurityFindingsAutomationSeverityModifierRule(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating severity modifier rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsSeverityModifierRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityFindingsSeverityModifierRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid severity modifier rule ID", err.Error())
		return
	}

	httpResp, err := r.Api.DeleteSecurityFindingsAutomationSeverityModifierRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting severity modifier rule"))
		return
	}
}

func (r *securityFindingsSeverityModifierRuleResource) updateState(ctx context.Context, state *securityFindingsSeverityModifierRuleModel, resp *datadogV2.SeverityModifierRuleResponse) diag.Diagnostics {
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
	actionModel := &severityModifierRuleActionModel{}
	if setAction := action.SeverityModifierRuleSetAction; setAction != nil {
		setModel := &severityModifierRuleSetActionModel{
			Severity: types.StringValue(string(setAction.GetSeverity())),
		}
		if setAction.HasDescription() {
			setModel.Description = types.StringValue(setAction.GetDescription())
		} else {
			setModel.Description = types.StringNull()
		}
		actionModel.Set = setModel
	} else if shiftAction := action.SeverityModifierRuleShiftAction; shiftAction != nil {
		shiftModel := &severityModifierRuleShiftActionModel{
			SeverityDelta: types.StringValue(string(shiftAction.GetSeverityDelta())),
		}
		if shiftAction.HasDescription() {
			shiftModel.Description = types.StringValue(shiftAction.GetDescription())
		} else {
			shiftModel.Description = types.StringNull()
		}
		actionModel.Shift = shiftModel
	}
	state.Action = actionModel

	return diags
}

// buildRuleData builds the JSON:API data object shared by the create and update requests.
func (r *securityFindingsSeverityModifierRuleResource) buildRuleData(ctx context.Context, state *securityFindingsSeverityModifierRuleModel) (*datadogV2.SeverityModifierRuleDataCreate, diag.Diagnostics) {
	var diags diag.Diagnostics

	scope, d := buildAutomationRuleScope(ctx, state.Rule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	var action datadogV2.SeverityModifierRuleAction
	if state.Action.Set != nil {
		setAction := datadogV2.NewSeverityModifierRuleSetAction(
			datadogV2.SeverityModifierSeverity(state.Action.Set.Severity.ValueString()),
			datadogV2.SEVERITYMODIFIERRULESETACTIONTYPE_SET,
		)
		if !state.Action.Set.Description.IsNull() && !state.Action.Set.Description.IsUnknown() {
			setAction.SetDescription(state.Action.Set.Description.ValueString())
		}
		action = datadogV2.SeverityModifierRuleSetActionAsSeverityModifierRuleAction(setAction)
	} else {
		shiftAction := datadogV2.NewSeverityModifierRuleShiftAction(
			datadogV2.SeverityModifierSeverityDelta(state.Action.Shift.SeverityDelta.ValueString()),
			datadogV2.SEVERITYMODIFIERRULESHIFTACTIONTYPE_SHIFT,
		)
		if !state.Action.Shift.Description.IsNull() && !state.Action.Shift.Description.IsUnknown() {
			shiftAction.SetDescription(state.Action.Shift.Description.ValueString())
		}
		action = datadogV2.SeverityModifierRuleShiftActionAsSeverityModifierRuleAction(shiftAction)
	}

	attributes := datadogV2.NewSeverityModifierRuleAttributesCreate(action, state.Name.ValueString(), *scope)
	attributes.SetEnabled(state.Enabled.ValueBool())

	data := datadogV2.NewSeverityModifierRuleDataCreateWithDefaults()
	data.SetType(datadogV2.SEVERITYMODIFIERRULETYPE_SEVERITY_MODIFIER_RULES)
	data.SetAttributes(*attributes)
	return data, diags
}
