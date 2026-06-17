package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

type securityFindingsMuteRulesOrderModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	RuleIDs types.List   `tfsdk:"rule_ids"`
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
	response.Schema = schema.Schema{
		Description: "Provides a Datadog security findings automation mute rules order resource. This is used to manage the evaluation order of mute rules for an organization. " +
			"This resource claims full ownership of the mute rules ordering: rules created outside Terraform are appended to the end of the order (and reported as a warning). " +
			"To control their position, list every mute rule ID in `rule_ids` (including rules created in the UI).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "A unique identifier for the order resource. This field has no server-side equivalent; it is recommended to match the resource name.",
				Required:    true,
			},
			"rule_ids": schema.ListAttribute{
				Description: "The ordered list of mute rule IDs. The order of IDs in this attribute defines the evaluation order of the mute rules.",
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *securityFindingsMuteRulesOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("name"), request, response)
}

func (r *securityFindingsMuteRulesOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityFindingsMuteRulesOrderModel
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
	var state securityFindingsMuteRulesOrderModel
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
	var state securityFindingsMuteRulesOrderModel
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

func (r *securityFindingsMuteRulesOrderResource) applyOrder(ctx context.Context, state *securityFindingsMuteRulesOrderModel) diag.Diagnostics {
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
		r.Api.ReorderSecurityFindingsAutomationMuteRules,
	)...)
	return diags
}
