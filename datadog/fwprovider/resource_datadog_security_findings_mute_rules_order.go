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
	_ resource.ResourceWithConfigure   = &securityFindingsMuteRulesOrderResource{}
	_ resource.ResourceWithImportState = &securityFindingsMuteRulesOrderResource{}
)

type securityFindingsMuteRulesOrderResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

func NewSecurityFindingsMuteRulesOrderResource() resource.Resource {
	return &securityFindingsMuteRulesOrderResource{}
}

func (r *securityFindingsMuteRulesOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsMuteRulesOrderResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_mute_rules_order"
}

func (r *securityFindingsMuteRulesOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = securityFindingsRulesOrderSchema("mute")
}

func (r *securityFindingsMuteRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsMuteRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readRulesOrder(ctx, request, response,
		func() (datadogV2.MuteRulesResponse, *http.Response, error) {
			return r.Api.ListSecurityFindingsAutomationMuteRules(r.Auth)
		},
		func(resp datadogV2.MuteRulesResponse) []datadogV2.MuteRuleDataResponse { return resp.GetData() },
		func(rule datadogV2.MuteRuleDataResponse) string { return rule.GetId().String() },
		"error listing mute rules",
	)
}

func (r *securityFindingsMuteRulesOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

func (r *securityFindingsMuteRulesOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

// Delete is a no-op: an ordering cannot be deleted. Removing this resource from configuration
// simply stops Terraform from managing the evaluation order.
func (r *securityFindingsMuteRulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *securityFindingsMuteRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsRulesOrderModel, diags *diag.Diagnostics) {
	var ruleIDs []string
	diags.Append(state.RuleIDs.ElementsAs(ctx, &ruleIDs, false)...)
	if diags.HasError() {
		return
	}
	orderedIDs, d := reorderSecurityFindingsAutomationRules(
		r.Auth,
		ruleIDs,
		func(id uuid.UUID) datadogV2.MuteRuleReorderItem {
			return *datadogV2.NewMuteRuleReorderItem(id, datadogV2.MUTERULETYPE_MUTE_RULES)
		},
		func(items []datadogV2.MuteRuleReorderItem) datadogV2.MuteRuleReorderRequest {
			return *datadogV2.NewMuteRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationMuteRules,
		func(resp datadogV2.MuteRuleReorderRequest) []datadogV2.MuteRuleReorderItem { return resp.GetData() },
		func(item datadogV2.MuteRuleReorderItem) string { return item.GetId().String() },
	)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	setOrderState(ctx, state, orderedIDs, diags)
}
