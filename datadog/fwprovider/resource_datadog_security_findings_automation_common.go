package fwprovider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

// automationRuleScopeModel maps the shared `rule` attribute used by every security findings
// automation rule (mute, due date, ticket creation). It defines which findings a rule applies to.
type automationRuleScopeModel struct {
	FindingTypes types.List   `tfsdk:"finding_types"`
	Query        types.String `tfsdk:"query"`
}

// securityFindingsAutomationRuleScopeAttribute returns the schema for the shared, required `rule`
// nested attribute.
func securityFindingsAutomationRuleScopeAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Defines the scope of findings to which the automation rule applies.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"finding_types": schema.ListAttribute{
				Description: "The list of security finding types that the automation rule applies to. Valid values are `api_security`, `attack_path`, `host_and_container_vulnerability`, `iac_misconfiguration`, `identity_risk`, `library_vulnerability`, `misconfiguration`, `runtime_code_vulnerability`, `secret`, `static_code_vulnerability`, `workload_activity`.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(validators.NewEnumValidator[validator.String](datadogV2.NewSecurityFindingTypeFromValue)),
				},
			},
			"query": schema.StringAttribute{
				Description: "A search query to further filter the findings matched by this rule. The `@workflow.*` namespace and `@status` fields are not permitted. For a reference of available fields, see the [Security Findings schema documentation](https://docs.datadoghq.com/security/guide/findings-schema/).",
				Optional:    true,
			},
		},
	}
}

// buildAutomationRuleScope converts the `rule` block model into the API scope object.
func buildAutomationRuleScope(ctx context.Context, m *automationRuleScopeModel) (*datadogV2.AutomationRuleScope, diag.Diagnostics) {
	var diags diag.Diagnostics

	var findingTypes []string
	diags.Append(m.FindingTypes.ElementsAs(ctx, &findingTypes, false)...)
	if diags.HasError() {
		return nil, diags
	}

	scope := datadogV2.NewAutomationRuleScopeWithDefaults()
	apiFindingTypes := make([]datadogV2.SecurityFindingType, len(findingTypes))
	for i, ft := range findingTypes {
		apiFindingTypes[i] = datadogV2.SecurityFindingType(ft)
	}
	scope.SetFindingTypes(apiFindingTypes)

	if !m.Query.IsNull() && !m.Query.IsUnknown() {
		scope.SetQuery(m.Query.ValueString())
	}

	return scope, diags
}

// flattenAutomationRuleScope converts the API scope object back into the `rule` block model.
func flattenAutomationRuleScope(ctx context.Context, scope datadogV2.AutomationRuleScope) (*automationRuleScopeModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	m := &automationRuleScopeModel{}

	findingTypes := scope.GetFindingTypes()
	ftStrings := make([]string, len(findingTypes))
	for i, ft := range findingTypes {
		ftStrings[i] = string(ft)
	}
	list, d := types.ListValueFrom(ctx, types.StringType, ftStrings)
	diags.Append(d...)
	m.FindingTypes = list

	if scope.HasQuery() {
		m.Query = types.StringValue(scope.GetQuery())
	} else {
		m.Query = types.StringNull()
	}

	return m, diags
}

// securityFindingsRulesOrderModel is the shared state model for the order resources.
type securityFindingsRulesOrderModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	RuleIDs types.List   `tfsdk:"rule_ids"`
}

// securityFindingsRulesOrderSchema builds the schema shared by the `*_rules_order` resources.
// ruleType is the automation rule type, for example "mute" or "due date".
func securityFindingsRulesOrderSchema(ruleType string) schema.Schema {
	return schema.Schema{
		Description: fmt.Sprintf(
			"Provides a resource that manages the evaluation order of %[1]s rules for the security findings in an organization. "+
				"The `rule_ids` list must contain every %[1]s rule ID; %[1]s rules created outside Terraform appear as drift. "+
				"**Note:** the %[1]s rule order is a single, organization-wide setting, so only one resource of this type should be declared per organization.", ruleType),
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "A unique identifier for the order resource. This field has no server-side equivalent; Datadog recommends matching the resource name.",
				Required:    true,
			},
			"rule_ids": schema.ListAttribute{
				Description: fmt.Sprintf("The ordered list of all %[1]s rule IDs. The order of IDs in this attribute defines the evaluation order of the %[1]s rules.", ruleType),
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

// upsertRulesOrder is the shared Create/Update body for the order resources.
func upsertRulesOrder(
	ctx context.Context,
	plan tfsdk.Plan,
	state *tfsdk.State,
	diags *diag.Diagnostics,
	apply func(context.Context, *securityFindingsRulesOrderModel, *diag.Diagnostics),
) {
	var model securityFindingsRulesOrderModel
	diags.Append(plan.Get(ctx, &model)...)
	if diags.HasError() {
		return
	}
	apply(ctx, &model, diags)
	if diags.HasError() {
		return
	}
	diags.Append(state.Set(ctx, &model)...)
}

// setOrderState writes the rule IDs and computed id into an order resource's state.
func setOrderState(ctx context.Context, state *securityFindingsRulesOrderModel, ruleIDs []string, diags *diag.Diagnostics) {
	list, d := types.ListValueFrom(ctx, types.StringType, ruleIDs)
	diags.Append(d...)
	state.RuleIDs = list
	state.ID = state.Name
}

// collectRuleIDs maps a slice of rule items to their string IDs.
func collectRuleIDs[T any](items []T, getID func(T) string) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = getID(item)
	}
	return ids
}

// readRulesOrder is the shared Read body for the order resources.
func readRulesOrder[Resp any, Rule any](
	ctx context.Context,
	request resource.ReadRequest,
	response *resource.ReadResponse,
	list func() (Resp, *http.Response, error),
	getData func(Resp) []Rule,
	getID func(Rule) string,
	errSummary string,
) {
	var state securityFindingsRulesOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	resp, _, err := list()
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, errSummary))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	setOrderState(ctx, &state, collectRuleIDs(getData(resp), getID), &response.Diagnostics)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// reorderSecurityFindingsAutomationRules submits a reorder request for the given rule IDs.
func reorderSecurityFindingsAutomationRules[I any, Req any](
	auth context.Context,
	ruleIDs []string,
	makeItem func(uuid.UUID) I,
	makeRequest func([]I) Req,
	reorderFn func(context.Context, Req) (Req, *http.Response, error),
	getRespItems func(Req) []I,
	getID func(I) string,
) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	items := make([]I, len(ruleIDs))
	for i, id := range ruleIDs {
		ruleUUID, err := uuid.Parse(id)
		if err != nil {
			diags.AddError("invalid rule ID", fmt.Sprintf("%q is not a valid rule ID: %s", id, err))
			return nil, diags
		}
		items[i] = makeItem(ruleUUID)
	}

	resp, _, err := reorderFn(auth, makeRequest(items))
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error reordering rules"))
		return nil, diags
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsedObject", err.Error())
		return nil, diags
	}
	return collectRuleIDs(getRespItems(resp), getID), diags
}
