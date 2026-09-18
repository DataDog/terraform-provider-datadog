package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &securityFindingsInboxRuleResource{}
	_ resource.ResourceWithImportState = &securityFindingsInboxRuleResource{}
)

type securityFindingsInboxRuleResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityFindingsInboxRuleModel struct {
	ID      types.String              `tfsdk:"id"`
	Name    types.String              `tfsdk:"name"`
	Enabled types.Bool                `tfsdk:"enabled"`
	Rule    *automationRuleScopeModel `tfsdk:"rule"`
	Action  *inboxRuleActionModel     `tfsdk:"action"`
}

// inboxRuleActionModel maps the inbox rule action: pushing matching findings into the Security
// Inbox triage view, with an optional description. The same wire shape (`InboxRuleAction`) is
// shared by the default inbox rules.
type inboxRuleActionModel struct {
	Description types.String `tfsdk:"description"`
}

func NewSecurityFindingsInboxRuleResource() resource.Resource {
	return &securityFindingsInboxRuleResource{}
}

func (r *securityFindingsInboxRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsInboxRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_inbox_rule"
}

func (r *securityFindingsInboxRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings automation inbox rule resource. This can be used to create and manage inbox rules that push matching security findings into the Security Inbox triage view. Use the `datadog_security_findings_inbox_rules_order` resource to manage the evaluation order of inbox rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the inbox rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the inbox rule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"rule": securityFindingsAutomationRuleScopeAttribute(),
			"action": schema.SingleNestedAttribute{
				Description: "The action to take when the inbox rule matches a finding. Matching findings are pushed into the Security Inbox triage view.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						Description: "An optional description providing more context for the rule.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *securityFindingsInboxRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *securityFindingsInboxRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsInboxRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid inbox rule ID", err.Error())
		return
	}

	resp, httpResp, err := r.Api.GetSecurityFindingsAutomationInboxRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving inbox rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsInboxRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityFindingsInboxRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewInboxRuleCreateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.CreateSecurityFindingsAutomationInboxRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating inbox rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsInboxRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityFindingsInboxRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid inbox rule ID", err.Error())
		return
	}

	data, diags := r.buildRuleUpdateData(ctx, &state, id)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewInboxRuleUpdateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.UpdateSecurityFindingsAutomationInboxRule(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating inbox rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsInboxRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityFindingsInboxRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid inbox rule ID", err.Error())
		return
	}

	httpResp, err := r.Api.DeleteSecurityFindingsAutomationInboxRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting inbox rule"))
		return
	}
}

func (r *securityFindingsInboxRuleResource) updateState(ctx context.Context, state *securityFindingsInboxRuleModel, resp *datadogV2.InboxRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.ID = types.StringValue(data.GetId().String())
	state.Name = types.StringValue(attributes.GetName())
	state.Enabled = types.BoolValue(attributes.GetEnabled())

	scope, d := flattenAutomationRuleScope(ctx, attributes.GetRule())
	diags.Append(d...)
	state.Rule = scope

	state.Action = flattenInboxRuleAction(attributes.GetAction())

	return diags
}

// buildRuleAttributes builds the attributes object shared by the create and update payloads.
func (r *securityFindingsInboxRuleResource) buildRuleAttributes(ctx context.Context, state *securityFindingsInboxRuleModel) (*datadogV2.InboxRuleAttributesCreate, diag.Diagnostics) {
	var diags diag.Diagnostics

	scope, d := buildAutomationRuleScope(ctx, state.Rule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	// The API requires the action attribute to be present (even when it only carries an
	// optional description), so it is always sent.
	action := datadogV2.NewInboxRuleActionWithDefaults()
	if !state.Action.Description.IsNull() && !state.Action.Description.IsUnknown() {
		action.SetDescription(state.Action.Description.ValueString())
	}

	attributes := datadogV2.NewInboxRuleAttributesCreateWithDefaults()
	attributes.SetName(state.Name.ValueString())
	attributes.SetEnabled(state.Enabled.ValueBool())
	attributes.SetRule(*scope)
	attributes.SetAction(*action)
	return attributes, diags
}

// buildRuleData builds the JSON:API data object for a create request.
func (r *securityFindingsInboxRuleResource) buildRuleData(ctx context.Context, state *securityFindingsInboxRuleModel) (*datadogV2.InboxRuleDataCreate, diag.Diagnostics) {
	attributes, diags := r.buildRuleAttributes(ctx, state)
	if diags.HasError() {
		return nil, diags
	}

	data := datadogV2.NewInboxRuleDataCreateWithDefaults()
	data.SetType(datadogV2.INBOXRULETYPE_INBOX_RULES)
	data.SetAttributes(*attributes)
	return data, diags
}

// buildRuleUpdateData builds the JSON:API data object for an update request.
// The API requires the resource id in the update body, matching the rule_id
// path parameter.
func (r *securityFindingsInboxRuleResource) buildRuleUpdateData(ctx context.Context, state *securityFindingsInboxRuleModel, id uuid.UUID) (*datadogV2.InboxRuleDataUpdate, diag.Diagnostics) {
	attributes, diags := r.buildRuleAttributes(ctx, state)
	if diags.HasError() {
		return nil, diags
	}

	data := datadogV2.NewInboxRuleDataUpdateWithDefaults()
	data.SetId(id)
	data.SetType(datadogV2.INBOXRULETYPE_INBOX_RULES)
	data.SetAttributes(*attributes)
	return data, diags
}
