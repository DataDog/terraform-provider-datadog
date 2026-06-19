package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
	var state securityFindingsRulesOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	resp, _, err := r.Api.ListSecurityFindingsAutomationDueDateRules(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing due date rules"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	ruleIDs := make([]string, 0, len(resp.GetData()))
	for _, rule := range resp.GetData() {
		ruleIDs = append(ruleIDs, rule.GetId().String())
	}
	setOrderState(ctx, &state, ruleIDs, &response.Diagnostics)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
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
	diags.Append(reorderSecurityFindingsAutomationRules(
		r.Auth,
		ruleIDs,
		func(id uuid.UUID) datadogV2.DueDateRuleReorderItem {
			return *datadogV2.NewDueDateRuleReorderItem(id, datadogV2.DUEDATERULETYPE_DUE_DATE_RULES)
		},
		func(items []datadogV2.DueDateRuleReorderItem) datadogV2.DueDateRuleReorderRequest {
			return *datadogV2.NewDueDateRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationDueDateRules,
	)...)
	if diags.HasError() {
		return
	}
	setOrderState(ctx, state, ruleIDs, diags)
}
