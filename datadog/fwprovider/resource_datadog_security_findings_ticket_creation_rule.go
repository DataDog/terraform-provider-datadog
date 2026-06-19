package fwprovider

import (
	"context"
	"encoding/json"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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
	_ resource.ResourceWithConfigure   = &securityFindingsTicketCreationRuleResource{}
	_ resource.ResourceWithImportState = &securityFindingsTicketCreationRuleResource{}
)

type securityFindingsTicketCreationRuleResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type securityFindingsTicketCreationRuleModel struct {
	ID      types.String                   `tfsdk:"id"`
	Name    types.String                   `tfsdk:"name"`
	Enabled types.Bool                     `tfsdk:"enabled"`
	Rule    *automationRuleScopeModel      `tfsdk:"rule"`
	Action  *ticketCreationRuleActionModel `tfsdk:"action"`
}

type ticketCreationRuleActionModel struct {
	ProjectID          types.String         `tfsdk:"project_id"`
	Target             types.String         `tfsdk:"target"`
	AssigneeID         types.String         `tfsdk:"assignee_id"`
	Fields             jsontypes.Normalized `tfsdk:"fields"`
	MaxTicketsPerDay   types.Int64          `tfsdk:"max_tickets_per_day"`
	AutoDisabledReason types.String         `tfsdk:"auto_disabled_reason"`
}

func NewSecurityFindingsTicketCreationRuleResource() resource.Resource {
	return &securityFindingsTicketCreationRuleResource{}
}

func (r *securityFindingsTicketCreationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsTicketCreationRuleResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_ticket_creation_rule"
}

func (r *securityFindingsTicketCreationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings automation ticket creation rule resource. This can be used to create and manage rules that automatically open tickets for matching security findings. Use the `datadog_security_findings_ticket_creation_rules_order` resource to manage the evaluation order of ticket creation rules.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of the ticket creation rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the ticket creation rule is enabled.",
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
					"project_id": schema.StringAttribute{
						Description: "The UUID of the case management project in which tickets are created.",
						Required:    true,
					},
					"target": schema.StringAttribute{
						Description: "The ticketing system in which to create tickets.",
						Required:    true,
						Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewTicketCreationTargetFromValue)},
					},
					"assignee_id": schema.StringAttribute{
						Description: "The UUID of the default assignee for created tickets.",
						Optional:    true,
					},
					"fields": schema.StringAttribute{
						Description: "A JSON-encoded object of custom fields for the created Jira issue. See the [Jira documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issues/#api-rest-api-2-issue-createmeta-projectidorkey-issuetypes-issuetypeid-get) for the list of available fields.",
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"max_tickets_per_day": schema.Int64Attribute{
						Description: "The maximum number of tickets the rule may create per day. If exceeded, a final ticket is created explaining the limit was hit. Must be between 1 and 500.",
						Required:    true,
					},
					"auto_disabled_reason": schema.StringAttribute{
						Description: "The reason the rule was automatically disabled by the system due to a ticketing integration error. This field is read-only.",
						Computed:    true,
					},
				},
				Validators: []validator.Object{objectvalidator.IsRequired()},
			},
		},
	}
}

func (r *securityFindingsTicketCreationRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *securityFindingsTicketCreationRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsTicketCreationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid ticket creation rule ID", err.Error())
		return
	}

	resp, httpResp, err := r.Api.GetSecurityFindingsAutomationTicketCreationRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ticket creation rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsTicketCreationRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityFindingsTicketCreationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewTicketCreationRuleCreateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.CreateSecurityFindingsAutomationTicketCreationRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating ticket creation rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsTicketCreationRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityFindingsTicketCreationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid ticket creation rule ID", err.Error())
		return
	}

	data, diags := r.buildRuleData(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewTicketCreationRuleUpdateRequestWithDefaults()
	body.SetData(*data)

	resp, _, err := r.Api.UpdateSecurityFindingsAutomationTicketCreationRule(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ticket creation rule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsTicketCreationRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state securityFindingsTicketCreationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("invalid ticket creation rule ID", err.Error())
		return
	}

	httpResp, err := r.Api.DeleteSecurityFindingsAutomationTicketCreationRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting ticket creation rule"))
		return
	}
}

func (r *securityFindingsTicketCreationRuleResource) updateState(ctx context.Context, state *securityFindingsTicketCreationRuleModel, resp *datadogV2.TicketCreationRuleResponse) diag.Diagnostics {
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
	actionModel := &ticketCreationRuleActionModel{
		ProjectID:        types.StringValue(action.GetProjectId().String()),
		Target:           types.StringValue(string(action.GetTarget())),
		MaxTicketsPerDay: types.Int64Value(action.GetMaxTicketsPerDay()),
	}
	if action.HasAssigneeId() {
		actionModel.AssigneeID = types.StringValue(action.GetAssigneeId().String())
	} else {
		actionModel.AssigneeID = types.StringNull()
	}
	if action.HasFields() {
		encoded, err := json.Marshal(action.GetFields())
		if err != nil {
			diags.AddError("error encoding ticket creation rule fields", err.Error())
		} else {
			actionModel.Fields = jsontypes.NewNormalizedValue(string(encoded))
		}
	} else {
		actionModel.Fields = jsontypes.NewNormalizedNull()
	}
	if action.HasAutoDisabledReason() {
		actionModel.AutoDisabledReason = types.StringValue(action.GetAutoDisabledReason())
	} else {
		actionModel.AutoDisabledReason = types.StringNull()
	}
	state.Action = actionModel

	return diags
}

// buildRuleData builds the JSON:API data object shared by the create and update requests.
func (r *securityFindingsTicketCreationRuleResource) buildRuleData(ctx context.Context, state *securityFindingsTicketCreationRuleModel) (*datadogV2.TicketCreationRuleDataCreate, diag.Diagnostics) {
	var diags diag.Diagnostics

	scope, d := buildAutomationRuleScope(ctx, state.Rule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	projectID, err := uuid.Parse(state.Action.ProjectID.ValueString())
	if err != nil {
		diags.AddError("invalid project_id", err.Error())
		return nil, diags
	}

	action := datadogV2.NewTicketCreationRuleActionWithDefaults()
	action.SetProjectId(projectID)
	action.SetTarget(datadogV2.TicketCreationTarget(state.Action.Target.ValueString()))
	action.SetMaxTicketsPerDay(state.Action.MaxTicketsPerDay.ValueInt64())

	if !state.Action.AssigneeID.IsNull() && !state.Action.AssigneeID.IsUnknown() {
		assigneeID, err := uuid.Parse(state.Action.AssigneeID.ValueString())
		if err != nil {
			diags.AddError("invalid assignee_id", err.Error())
			return nil, diags
		}
		action.SetAssigneeId(assigneeID)
	}

	if !state.Action.Fields.IsNull() && !state.Action.Fields.IsUnknown() {
		var fields interface{}
		if err := json.Unmarshal([]byte(state.Action.Fields.ValueString()), &fields); err != nil {
			diags.AddError("invalid fields", "the fields attribute must be a valid JSON object: "+err.Error())
			return nil, diags
		}
		action.SetFields(fields)
	}

	attributes := datadogV2.NewTicketCreationRuleAttributesCreateWithDefaults()
	attributes.SetName(state.Name.ValueString())
	attributes.SetEnabled(state.Enabled.ValueBool())
	attributes.SetRule(*scope)
	attributes.SetAction(*action)

	data := datadogV2.NewTicketCreationRuleDataCreateWithDefaults()
	data.SetType(datadogV2.TICKETCREATIONRULETYPE_TICKET_CREATION_RULES)
	data.SetAttributes(*attributes)
	return data, diags
}
