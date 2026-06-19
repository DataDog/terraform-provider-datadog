package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
	response.Schema = securityFindingsRulesOrderSchema("mute rule")
}

func (r *securityFindingsMuteRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsMuteRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsRulesOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	serverOrder, diags := r.listServerOrder()
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	adoptAll := state.RuleIDs.IsNull()
	var declared []string
	if !adoptAll {
		response.Diagnostics.Append(state.RuleIDs.ElementsAs(ctx, &declared, false)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	list, d := types.ListValueFrom(ctx, types.StringType, trackedOrder(declared, serverOrder, adoptAll))
	response.Diagnostics.Append(d...)
	state.RuleIDs = list
	if state.ID.IsNull() {
		state.ID = state.Name
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsMuteRulesOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state securityFindingsRulesOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.applyOrder(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	state.ID = state.Name

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *securityFindingsMuteRulesOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state securityFindingsRulesOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.applyOrder(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	state.ID = state.Name

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Delete is a no-op: an ordering cannot be deleted from the API. Removing this resource from
// configuration simply stops Terraform from managing the mute rules evaluation order.
func (r *securityFindingsMuteRulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// listServerOrder returns the IDs of all mute rules in their current server-side order.
func (r *securityFindingsMuteRulesOrderResource) listServerOrder() ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics

	resp, _, err := r.Api.ListSecurityFindingsAutomationMuteRules(r.Auth)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing mute rules"))
		return nil, diags
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsedObject", err.Error())
		return nil, diags
	}

	ids := make([]string, 0, len(resp.GetData()))
	for _, rule := range resp.GetData() {
		ids = append(ids, rule.GetId().String())
	}
	return ids, diags
}

func (r *securityFindingsMuteRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsRulesOrderModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var declared []string
	diags.Append(state.RuleIDs.ElementsAs(ctx, &declared, false)...)
	if diags.HasError() {
		return diags
	}

	serverOrder, d := r.listServerOrder()
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(applySecurityFindingsAutomationRulesOrder(
		r.Auth,
		declared,
		serverOrder,
		string(datadogV2.MUTERULETYPE_MUTE_RULES),
		func(id uuid.UUID) datadogV2.MuteRuleReorderItem {
			return *datadogV2.NewMuteRuleReorderItem(id, datadogV2.MUTERULETYPE_MUTE_RULES)
		},
		func(items []datadogV2.MuteRuleReorderItem) datadogV2.MuteRuleReorderRequest {
			return *datadogV2.NewMuteRuleReorderRequest(items)
		},
		r.Api.ReorderSecurityFindingsAutomationMuteRules,
	)...)
	return diags
}
