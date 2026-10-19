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
	_ resource.ResourceWithConfigure   = &securityFindingsInboxRulesOrderResource{}
	_ resource.ResourceWithImportState = &securityFindingsInboxRulesOrderResource{}
)

type securityFindingsInboxRulesOrderResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

func NewSecurityFindingsInboxRulesOrderResource() resource.Resource {
	return &securityFindingsInboxRulesOrderResource{}
}

func (r *securityFindingsInboxRulesOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsInboxRulesOrderResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_inbox_rules_order"
}

func (r *securityFindingsInboxRulesOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = securityFindingsRulesOrderSchema("inbox")
}

func (r *securityFindingsInboxRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsInboxRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readRulesOrder(ctx, request, response,
		func() (datadogV2.InboxRulesResponse, *http.Response, error) {
			return r.Api.ListSecurityFindingsAutomationInboxRules(r.Auth)
		},
		func(resp datadogV2.InboxRulesResponse) []datadogV2.InboxRuleDataResponse { return resp.GetData() },
		func(rule datadogV2.InboxRuleDataResponse) string { return rule.GetId().String() },
		"error listing inbox rules",
	)
}

func (r *securityFindingsInboxRulesOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

func (r *securityFindingsInboxRulesOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

// Delete is a no-op: an ordering cannot be deleted. Removing this resource from configuration
// simply stops Terraform from managing the evaluation order.
func (r *securityFindingsInboxRulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *securityFindingsInboxRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsRulesOrderModel, diags *diag.Diagnostics) {
	var ruleIDs []string
	diags.Append(state.RuleIDs.ElementsAs(ctx, &ruleIDs, false)...)
	if diags.HasError() {
		return
	}
	orderedIDs, d := reorderSecurityFindingsAutomationRules(
		r.Auth,
		ruleIDs,
		func(id uuid.UUID) datadogV2.InboxRuleReorderItem {
			return *datadogV2.NewInboxRuleReorderItem(id, datadogV2.INBOXRULETYPE_INBOX_RULES)
		},
		func(items []datadogV2.InboxRuleReorderItem) datadogV2.InboxRuleReorderRequest {
			return *datadogV2.NewInboxRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationInboxRules,
		func(resp datadogV2.InboxRuleReorderResponse) []datadogV2.InboxRuleReorderItem { return resp.GetData() },
		func(item datadogV2.InboxRuleReorderItem) string { return item.GetId().String() },
	)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	setOrderState(ctx, state, orderedIDs, diags)
}
