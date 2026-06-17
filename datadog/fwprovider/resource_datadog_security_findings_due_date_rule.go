package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	_ resource.ResourceWithConfigure   = &securityFindingsDueDateRuleResource{}
	_ resource.ResourceWithImportState = &securityFindingsDueDateRuleResource{}
)

type securityFindingsDueDateRuleResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityFindingsDueDateRuleModel struct {
	ID      types.String              `tfsdk:"id"`
	Name    types.String              `tfsdk:"name"`
	Enabled types.Bool                `tfsdk:"enabled"`
	Rule    *automationRuleScopeModel `tfsdk:"rule"`
	Action  *dueDateRuleActionModel   `tfsdk:"action"`
}

type dueDateRuleActionModel struct {
	DueFrom            types.String                  `tfsdk:"due_from"`
	ReasonDescription  types.String                  `tfsdk:"reason_description"`
	DueDaysPerSeverity []dueDatePerSeverityItemModel `tfsdk:"due_days_per_severity"`
}

type dueDatePerSeverityItemModel struct {
	Severity  types.String `tfsdk:"severity"`
	DueInDays types.Int64  `tfsdk:"due_in_days"`
}

func NewSecurityFindingsDueDateRuleResource() resource.Resource {
	return &securityFindingsDueDateRuleResource{}
}

func (r *securityFindingsDueDateRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsDueDateRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_due_date_rule"
}

func (r *securityFindingsDueDateRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings automation due date rule resource. This can be used to create and manage due date rules that automatically assign remediation deadlines to matching security findings. Use the `datadog_security_findings_due_date_rules_order` resource to manage the evaluation order of due date rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the due date rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the due date rule is enabled.",
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
					"due_from": schema.StringAttribute{
						Description: "The reference point from which the due date is calculated. When `fix_available` is selected but not applicable to the finding type, `first_seen` is used instead.",
						Required:    true,
						Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewDueDateFromFromValue)},
					},
					"reason_description": schema.StringAttribute{
						Description: "An optional description providing more context for the due date assignment.",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"due_days_per_severity": schema.ListNestedBlock{
						Description: "The number of days until a finding is due, configured per severity. Each severity may appear at most once.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"severity": schema.StringAttribute{
									Description: "The severity level.",
									Required:    true,
									Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewDueDateSeverityFromValue)},
								},
								"due_in_days": schema.Int64Attribute{
									Description: "The number of days from the reference point until the finding is due.",
									Required:    true,
								},
							},
						},
						Validators: []validator.List{listvalidator.SizeAtLeast(1)},
					},
				},
				Validators: []validator.Object{objectvalidator.IsRequired()},
			},
		},
	}
}

func (r *securityFindingsDueDateRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *securityFindingsDueDateRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsDueDateRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid due date rule ID", err.Error())
		return
	}

	resp, httpResp, err := r.Api.GetSecurityFindingsAutomationDueDateRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving due date rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsDueDateRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityFindingsDueDateRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateSecurityFindingsAutomationDueDateRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating due date rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsDueDateRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityFindingsDueDateRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid due date rule ID", err.Error())
		return
	}

	body, diags := r.buildUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateSecurityFindingsAutomationDueDateRule(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating due date rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsDueDateRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityFindingsDueDateRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid due date rule ID", err.Error())
		return
	}

	httpResp, err := r.Api.DeleteSecurityFindingsAutomationDueDateRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting due date rule"))
		return
	}
}

func (r *securityFindingsDueDateRuleResource) updateState(ctx context.Context, state *securityFindingsDueDateRuleModel, resp *datadogV2.DueDateRuleResponse) diag.Diagnostics {
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
	actionModel := &dueDateRuleActionModel{
		DueFrom: types.StringValue(string(action.GetDueFrom())),
	}
	if action.HasReasonDescription() {
		actionModel.ReasonDescription = types.StringValue(action.GetReasonDescription())
	} else {
		actionModel.ReasonDescription = types.StringNull()
	}
	items := action.GetDueDaysPerSeverity()
	actionModel.DueDaysPerSeverity = make([]dueDatePerSeverityItemModel, len(items))
	for i, item := range items {
		actionModel.DueDaysPerSeverity[i] = dueDatePerSeverityItemModel{
			Severity:  types.StringValue(string(item.GetSeverity())),
			DueInDays: types.Int64Value(item.GetDueInDays()),
		}
	}
	state.Action = actionModel

	return diags
}

func (r *securityFindingsDueDateRuleResource) buildAttributes(ctx context.Context, state *securityFindingsDueDateRuleModel) (*datadogV2.DueDateRuleAttributesCreate, diag.Diagnostics) {
	var diags diag.Diagnostics

	scope, d := buildAutomationRuleScope(ctx, state.Rule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	items := make([]datadogV2.DueDatePerSeverityItem, len(state.Action.DueDaysPerSeverity))
	for i, item := range state.Action.DueDaysPerSeverity {
		items[i] = *datadogV2.NewDueDatePerSeverityItem(
			item.DueInDays.ValueInt64(),
			datadogV2.DueDateSeverity(item.Severity.ValueString()),
		)
	}

	action := datadogV2.NewDueDateRuleActionWithDefaults()
	action.SetDueDaysPerSeverity(items)
	action.SetDueFrom(datadogV2.DueDateFrom(state.Action.DueFrom.ValueString()))
	if !state.Action.ReasonDescription.IsNull() && !state.Action.ReasonDescription.IsUnknown() {
		action.SetReasonDescription(state.Action.ReasonDescription.ValueString())
	}

	attributes := datadogV2.NewDueDateRuleAttributesCreateWithDefaults()
	attributes.SetName(state.Name.ValueString())
	attributes.SetEnabled(state.Enabled.ValueBool())
	attributes.SetRule(*scope)
	attributes.SetAction(*action)

	return attributes, diags
}

func (r *securityFindingsDueDateRuleResource) buildCreateRequestBody(ctx context.Context, state *securityFindingsDueDateRuleModel) (*datadogV2.DueDateRuleCreateRequest, diag.Diagnostics) {
	attributes, diags := r.buildAttributes(ctx, state)
	if diags.HasError() {
		return nil, diags
	}

	data := datadogV2.NewDueDateRuleDataCreateWithDefaults()
	data.SetType(datadogV2.DUEDATERULETYPE_DUE_DATE_RULES)
	data.SetAttributes(*attributes)

	req := datadogV2.NewDueDateRuleCreateRequestWithDefaults()
	req.SetData(*data)
	return req, diags
}

func (r *securityFindingsDueDateRuleResource) buildUpdateRequestBody(ctx context.Context, state *securityFindingsDueDateRuleModel) (*datadogV2.DueDateRuleUpdateRequest, diag.Diagnostics) {
	attributes, diags := r.buildAttributes(ctx, state)
	if diags.HasError() {
		return nil, diags
	}

	data := datadogV2.NewDueDateRuleDataCreateWithDefaults()
	data.SetType(datadogV2.DUEDATERULETYPE_DUE_DATE_RULES)
	data.SetAttributes(*attributes)

	req := datadogV2.NewDueDateRuleUpdateRequestWithDefaults()
	req.SetData(*data)
	return req, diags
}
