package fwprovider

import (
	"context"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &tagIndexingRuleOrderResource{}
	_ resource.ResourceWithImportState = &tagIndexingRuleOrderResource{}
)

type tagIndexingRuleOrderResource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

type tagIndexingRuleOrderModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	RuleIDs types.List   `tfsdk:"rule_ids"`
}

func NewTagIndexingRuleOrderResource() resource.Resource {
	return &tagIndexingRuleOrderResource{}
}

func (r *tagIndexingRuleOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	r.Auth = providerData.Auth
}

func (r *tagIndexingRuleOrderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "tag_indexing_rule_order"
}

func (r *tagIndexingRuleOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Tag Indexing Rule Order resource. Manages the evaluation order of tag indexing rules for an org. Only one instance of this resource should exist per org; the `name` field is a user-chosen identifier with no server-side equivalent.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A unique name for the order resource. Recommended to match the resource name. No corresponding field exists in the API.",
			},
			"rule_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Ordered list of ALL tag indexing rule UUIDs. The server assigns each rule a rule_order value (1, 2, 3, ...) corresponding to its position in this list. This resource claims full ownership of evaluation order: rules created outside Terraform (e.g. via the UI) will appear as configuration drift on the next plan. All rules must be listed here; omitting a rule ID will result in a 404 error from the API.",
			},
		},
	}
}

func (r *tagIndexingRuleOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

// Create delegates to Update since the order resource is a singleton — there's nothing to
// create server-side, just a reorder call.
func (r *tagIndexingRuleOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state tagIndexingRuleOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if diags := r.applyOrder(ctx, &state); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state.ID = state.Name
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagIndexingRuleOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tagIndexingRuleOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Read current order by listing all rules and sorting by rule_order.
	resp, _, err := r.Api.ListTagIndexingRules(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing tag indexing rules for order read"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	rules := resp.GetData()
	sort.Slice(rules, func(i, j int) bool {
		ai, aj := rules[i].GetAttributes(), rules[j].GetAttributes()
		return ai.GetRuleOrder() < aj.GetRuleOrder()
	})

	ids := make([]types.String, 0, len(rules))
	for _, rule := range rules {
		ids = append(ids, types.StringValue(rule.GetId()))
	}
	state.RuleIDs, _ = types.ListValueFrom(ctx, types.StringType, ids)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagIndexingRuleOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state tagIndexingRuleOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if diags := r.applyOrder(ctx, &state); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	state.ID = state.Name
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Delete is a no-op: ordering cannot be "deleted" from the API.
// Removing this resource from config just stops Terraform from managing the order.
func (r *tagIndexingRuleOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *tagIndexingRuleOrderResource) applyOrder(ctx context.Context, state *tagIndexingRuleOrderModel) diag.Diagnostics {
	var diags diag.Diagnostics

	var ruleIDs []string
	if d := state.RuleIDs.ElementsAs(ctx, &ruleIDs, false); d.HasError() {
		diags.Append(d...)
		return diags
	}

	attrs := datadogV2.NewTagIndexingRuleOrderAttributesWithDefaults()
	attrs.SetRuleIds(ruleIDs)

	data := datadogV2.NewTagIndexingRuleOrderDataWithDefaults()
	data.SetAttributes(*attrs)

	body := datadogV2.NewTagIndexingRuleOrderRequestWithDefaults()
	body.SetData(*data)

	httpResp, err := r.Api.ReorderTagIndexingRules(r.Auth, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			diags.AddError("one or more rule IDs not found",
				"ReorderTagIndexingRules returned 404: ensure all rule_ids exist before setting order")
			return diags
		}
		diags.Append(utils.FrameworkErrorDiag(err, "error reordering tag indexing rules"))
	}
	return diags
}
