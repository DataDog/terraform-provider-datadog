package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

// automationRuleScopeModel maps the shared `rule` block used by every security findings
// automation rule (mute, due date, ticket creation). It defines which findings a rule applies to.
type automationRuleScopeModel struct {
	FindingTypes types.List   `tfsdk:"finding_types"`
	Query        types.String `tfsdk:"query"`
}

// securityFindingsAutomationRuleScopeBlock returns the schema for the shared `rule` block.
func securityFindingsAutomationRuleScopeBlock() schema.Block {
	return schema.SingleNestedBlock{
		Description: "The scope of findings to which the rule applies.",
		Attributes: map[string]schema.Attribute{
			"finding_types": schema.ListAttribute{
				Description: "The list of security finding types the rule applies to. Valid values are `api_security`, `attack_path`, `host_and_container_vulnerability`, `iac_misconfiguration`, `identity_risk`, `library_vulnerability`, `misconfiguration`, `runtime_code_vulnerability`, `secret`, `static_code_vulnerability`, `workload_activity`.",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(validators.NewEnumValidator[validator.String](datadogV2.NewSecurityFindingTypeFromValue)),
				},
			},
			"query": schema.StringAttribute{
				Description: "A search query to further filter the findings matched by this rule. The `@workflow.*` namespace, and the `@is_in_security_inbox` and `@status` fields, are not permitted.",
				Optional:    true,
			},
		},
		Validators: []validator.Object{objectvalidator.IsRequired()},
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

// ----------------------------------------------------------------------------
// Shared helpers for the `*_order` resources.
//
// The reorder API requires the complete, exact set of rule IDs that currently exist server-side;
// a partial list is rejected. To make the resource lifecycle work (and to coexist with rules
// created outside Terraform), the order resources reconcile the user-declared list against the
// live set before every reorder call.
// ----------------------------------------------------------------------------

// reconcileOrder merges the user-declared ordered IDs with the live set of rule IDs from the
// server. Declared IDs keep their relative order; any rule that exists server-side but is not
// declared is appended at the end, so the reorder request always contains the complete set the
// API requires. It returns the full ordered list to submit and the list of undeclared IDs.
func reconcileOrder(declared []string, serverOrder []string) (submit []string, unmanaged []string) {
	declaredSet := make(map[string]struct{}, len(declared))
	for _, id := range declared {
		declaredSet[id] = struct{}{}
	}

	submit = make([]string, 0, len(serverOrder)+len(declared))
	submit = append(submit, declared...)
	for _, id := range serverOrder {
		if _, ok := declaredSet[id]; !ok {
			submit = append(submit, id)
			unmanaged = append(unmanaged, id)
		}
	}
	return submit, unmanaged
}

// trackedOrder computes the value to persist in state for an order resource's `rule_ids`. On a
// fresh import (adoptAll), it adopts the full server order. Otherwise it returns the managed
// (declared) rule IDs in their current server-side order: it walks the live server order and keeps
// only the declared IDs. This means rules created outside Terraform stay invisible (no perpetual
// drift), while reordering or deleting a managed rule out-of-band surfaces as drift, so the next
// apply re-asserts the configured order.
func trackedOrder(declared []string, serverOrder []string, adoptAll bool) []string {
	if adoptAll {
		out := make([]string, len(serverOrder))
		copy(out, serverOrder)
		return out
	}

	declaredSet := make(map[string]struct{}, len(declared))
	for _, id := range declared {
		declaredSet[id] = struct{}{}
	}
	out := make([]string, 0, len(declared))
	for _, id := range serverOrder {
		if _, ok := declaredSet[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

// applySecurityFindingsAutomationRulesOrder reconciles the declared order against the live rule
// set and submits a complete reorder request. ruleType is the JSON:API type (for example
// "mute_rules"), used both for the reorder items and in user-facing messages. reorderFn is the
// type-specific client method (passed as a method value).
func applySecurityFindingsAutomationRulesOrder(
	auth context.Context,
	declared []string,
	serverOrder []string,
	ruleType string,
	reorderFn func(context.Context, datadogV2.SecurityAutomationRuleReorderRequest) (datadogV2.SecurityAutomationRuleReorderRequest, *http.Response, error),
) diag.Diagnostics {
	var diags diag.Diagnostics

	submit, unmanaged := reconcileOrder(declared, serverOrder)
	if len(unmanaged) > 0 {
		diags.AddWarning(
			"Unmanaged rules appended to ordering",
			fmt.Sprintf("The following %s exist but are not listed in rule_ids, so they were appended to the end of the evaluation order: %s. "+
				"This order resource claims full ownership of the evaluation order; list every rule ID (including rules created in the UI) "+
				"to control their position and silence this warning.", ruleType, strings.Join(unmanaged, ", ")),
		)
	}

	items := make([]datadogV2.SecurityAutomationRuleReorderItem, len(submit))
	for i, id := range submit {
		ruleUUID, err := uuid.Parse(id)
		if err != nil {
			diags.AddError(fmt.Sprintf("invalid %s ID", ruleType), err.Error())
			return diags
		}
		items[i] = datadogV2.SecurityAutomationRuleReorderItem{
			Id:   ruleUUID,
			Type: ruleType,
		}
	}

	body := datadogV2.NewSecurityAutomationRuleReorderRequest(items)
	resp, httpResp, err := reorderFn(auth, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			diags.AddError(fmt.Sprintf("one or more %s IDs not found", ruleType),
				"the reorder request referenced a rule that does not exist; ensure all rule_ids exist before setting order")
			return diags
		}
		diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering %s", ruleType)))
		return diags
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsedObject", err.Error())
		return diags
	}

	return diags
}
