package fwprovider

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.ResourceWithConfigure   = &securityFindingsSeverityModifierRulesOrderResource{}
	_ resource.ResourceWithImportState = &securityFindingsSeverityModifierRulesOrderResource{}
)

type securityFindingsSeverityModifierRulesOrderResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

func NewSecurityFindingsSeverityModifierRulesOrderResource() resource.Resource {
	return &securityFindingsSeverityModifierRulesOrderResource{}
}

func (r *securityFindingsSeverityModifierRulesOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsSeverityModifierRulesOrderResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_severity_modifier_rules_order"
}

func (r *securityFindingsSeverityModifierRulesOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = securityFindingsRulesOrderSchema("severity modifier")
}

func (r *securityFindingsSeverityModifierRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsSeverityModifierRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readRulesOrder(ctx, request, response,
		func() (datadogV2.SeverityModifierRulesResponse, *http.Response, error) {
			return r.Api.ListSecurityFindingsAutomationSeverityModifierRules(r.Auth)
		},
		func(resp datadogV2.SeverityModifierRulesResponse) []datadogV2.SeverityModifierRuleDataResponse {
			return resp.GetData()
		},
		func(rule datadogV2.SeverityModifierRuleDataResponse) string { return rule.GetId().String() },
		"error listing severity modifier rules",
	)
}

func (r *securityFindingsSeverityModifierRulesOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

func (r *securityFindingsSeverityModifierRulesOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

// Delete is a no-op: an ordering cannot be deleted. Removing this resource from configuration
// simply stops Terraform from managing the evaluation order.
func (r *securityFindingsSeverityModifierRulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *securityFindingsSeverityModifierRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsRulesOrderModel, diags *diag.Diagnostics) {
	var ruleIDs []string
	diags.Append(state.RuleIDs.ElementsAs(ctx, &ruleIDs, false)...)
	if diags.HasError() {
		return
	}
	orderedIDs, d := reorderSecurityFindingsAutomationRules(
		r.Auth,
		ruleIDs,
		func(id uuid.UUID) datadogV2.SeverityModifierRuleReorderItem {
			return *datadogV2.NewSeverityModifierRuleReorderItem(id, datadogV2.SEVERITYMODIFIERRULETYPE_SEVERITY_MODIFIER_RULES)
		},
		func(items []datadogV2.SeverityModifierRuleReorderItem) datadogV2.SeverityModifierRuleReorderRequest {
			return *datadogV2.NewSeverityModifierRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationSeverityModifierRules,
		func(resp datadogV2.SeverityModifierRuleReorderResponse) []datadogV2.SeverityModifierRuleReorderItem {
			return resp.GetData()
		},
		func(item datadogV2.SeverityModifierRuleReorderItem) string { return item.GetId().String() },
	)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	setOrderState(ctx, state, orderedIDs, diags)
}
