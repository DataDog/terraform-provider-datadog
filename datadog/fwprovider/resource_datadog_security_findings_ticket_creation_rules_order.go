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
	_ resource.ResourceWithConfigure   = &securityFindingsTicketCreationRulesOrderResource{}
	_ resource.ResourceWithImportState = &securityFindingsTicketCreationRulesOrderResource{}
)

type securityFindingsTicketCreationRulesOrderResource struct {
	Api  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

func NewSecurityFindingsTicketCreationRulesOrderResource() resource.Resource {
	return &securityFindingsTicketCreationRulesOrderResource{}
}

func (r *securityFindingsTicketCreationRulesOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *securityFindingsTicketCreationRulesOrderResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_findings_ticket_creation_rules_order"
}

func (r *securityFindingsTicketCreationRulesOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = securityFindingsRulesOrderSchema("ticket creation")
}

func (r *securityFindingsTicketCreationRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsTicketCreationRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readRulesOrder(ctx, request, response,
		func() (datadogV2.TicketCreationRulesResponse, *http.Response, error) {
			return r.Api.ListSecurityFindingsAutomationTicketCreationRules(r.Auth)
		},
		func(resp datadogV2.TicketCreationRulesResponse) []datadogV2.TicketCreationRuleDataResponse {
			return resp.GetData()
		},
		func(rule datadogV2.TicketCreationRuleDataResponse) string { return rule.GetId().String() },
		"error listing ticket creation rules",
	)
}

func (r *securityFindingsTicketCreationRulesOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

func (r *securityFindingsTicketCreationRulesOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	upsertRulesOrder(ctx, request.Plan, &response.State, &response.Diagnostics, r.applyOrder)
}

// Delete is a no-op: an ordering cannot be deleted. Removing this resource from configuration
// simply stops Terraform from managing the evaluation order.
func (r *securityFindingsTicketCreationRulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *securityFindingsTicketCreationRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsRulesOrderModel, diags *diag.Diagnostics) {
	var ruleIDs []string
	diags.Append(state.RuleIDs.ElementsAs(ctx, &ruleIDs, false)...)
	if diags.HasError() {
		return
	}
	diags.Append(reorderSecurityFindingsAutomationRules(
		r.Auth,
		ruleIDs,
		func(id uuid.UUID) datadogV2.TicketCreationRuleReorderItem {
			return *datadogV2.NewTicketCreationRuleReorderItem(id, datadogV2.TICKETCREATIONRULETYPE_TICKET_CREATION_RULES)
		},
		func(items []datadogV2.TicketCreationRuleReorderItem) datadogV2.TicketCreationRuleReorderRequest {
			return *datadogV2.NewTicketCreationRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationTicketCreationRules,
	)...)
	if diags.HasError() {
		return
	}
	setOrderState(ctx, state, ruleIDs, diags)
}
