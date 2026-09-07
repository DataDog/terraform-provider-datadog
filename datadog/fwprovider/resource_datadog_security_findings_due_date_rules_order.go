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
	_ resource.ResourceWithConfigure   = &securityFindingsDueDateRulesOrderResource{}
	_ resource.ResourceWithImportState = &securityFindingsDueDateRulesOrderResource{}
)

type securityFindingsDueDateRulesOrderResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

func NewSecurityFindingsDueDateRulesOrderResource() resource.Resource {
	return &securityFindingsDueDateRulesOrderResource{}
}

func (r *securityFindingsDueDateRulesOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsDueDateRulesOrderResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_due_date_rules_order"
}

func (r *securityFindingsDueDateRulesOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = securityFindingsRulesOrderSchema("due date")
}

func (r *securityFindingsDueDateRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsDueDateRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readRulesOrder(ctx, request, response,
		func() (datadogV2.DueDateRulesResponse, *http.Response, error) {
			return r.Api.ListSecurityFindingsAutomationDueDateRules(r.Auth)
		},
		func(resp datadogV2.DueDateRulesResponse) []datadogV2.DueDateRuleDataResponse { return resp.GetData() },
		func(rule datadogV2.DueDateRuleDataResponse) string { return rule.GetId().String() },
		"error listing due date rules",
	)
}

func (r *securityFindingsDueDateRulesOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

func (r *securityFindingsDueDateRulesOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

// Delete is a no-op: an ordering cannot be deleted. Removing this resource from configuration
// simply stops Terraform from managing the evaluation order.
func (r *securityFindingsDueDateRulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *securityFindingsDueDateRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsRulesOrderModel, diags *diag.Diagnostics) {
	var ruleIDs []string
	diags.Append(state.RuleIDs.ElementsAs(ctx, &ruleIDs, false)...)
	if diags.HasError() {
		return
	}
	orderedIDs, d := reorderSecurityFindingsAutomationRules(
		r.Auth,
		ruleIDs,
		func(id uuid.UUID) datadogV2.DueDateRuleReorderItem {
			return *datadogV2.NewDueDateRuleReorderItem(id, datadogV2.DUEDATERULETYPE_DUE_DATE_RULES)
		},
		func(items []datadogV2.DueDateRuleReorderItem) datadogV2.DueDateRuleReorderRequest {
			return *datadogV2.NewDueDateRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationDueDateRules,
		func(resp datadogV2.DueDateRuleReorderRequest) []datadogV2.DueDateRuleReorderItem {
			return resp.GetData()
		},
		func(item datadogV2.DueDateRuleReorderItem) string { return item.GetId().String() },
	)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	setOrderState(ctx, state, orderedIDs, diags)
}
