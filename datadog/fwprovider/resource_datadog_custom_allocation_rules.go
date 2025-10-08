package fwprovider

import (
	"context"
	"fmt"
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
	_ resource.ResourceWithConfigure   = &customAllocationRulesResource{}
	_ resource.ResourceWithImportState = &customAllocationRulesResource{}
)

func NewCustomAllocationRulesResource() resource.Resource {
	return &customAllocationRulesResource{}
}

type customAllocationRulesModel struct {
	ID      types.String `tfsdk:"id"`
	RuleIDs types.List   `tfsdk:"rule_ids"`
}

type customAllocationRulesResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func (r *customAllocationRulesResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *customAllocationRulesResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "custom_allocation_rules"
}

func (r *customAllocationRulesResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Custom Allocation Rule Order API resource. This can be used to manage the order of Datadog Custom Allocation Rules.",
		Attributes: map[string]schema.Attribute{
			"rule_ids": schema.ListAttribute{
				Description: "The list of Custom Allocation Rule IDs, in order. Rules are executed in the order specified in this list.",
				ElementType: types.StringType,
				Required:    true,
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *customAllocationRulesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state customAllocationRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *customAllocationRulesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state customAllocationRulesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the current list of rules from API to read their order
	resp, httpResponse, err := r.Api.ListArbitraryCostRules(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading custom allocation rules. http response: %v", httpResponse)))
		return
	}

	var rules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		rules = *respData
	}

	// Create a slice of structs to sort by order_id
	type ruleWithOrder struct {
		id      string
		orderId int64
	}

	ruleOrderIds := make([]ruleWithOrder, 0, len(rules))
	for _, rule := range rules {
		if ruleID, ok := rule.GetIdOk(); ok {
			orderId := int64(0)
			if ruleAttrs, ok := rule.GetAttributesOk(); ok {
				orderId = ruleAttrs.GetOrderId()
			}
			ruleOrderIds = append(ruleOrderIds, ruleWithOrder{
				id:      *ruleID,
				orderId: orderId,
			})
		}
	}

	// Sort by order_id
	sort.Slice(ruleOrderIds, func(i, j int) bool {
		return ruleOrderIds[i].orderId < ruleOrderIds[j].orderId
	})

	// Extract ordered IDs
	orderedList := make([]string, 0, len(ruleOrderIds))
	for _, ro := range ruleOrderIds {
		orderedList = append(orderedList, ro.id)
	}

	state.RuleIDs, _ = types.ListValueFrom(ctx, types.StringType, orderedList)
	state.ID = types.StringValue("order")

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *customAllocationRulesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state customAllocationRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *customAllocationRulesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// No-op: deleting this resource only removes it from Terraform state
}

func (r *customAllocationRulesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *customAllocationRulesResource) updateOrder(state *customAllocationRulesModel, diag *diag.Diagnostics) {
	// Set the ID immediately to prevent "unknown value" errors
	state.ID = types.StringValue("order")

	// Convert the Terraform list to strings for the API call
	var desiredRuleIDs []string
	for _, tfID := range state.RuleIDs.Elements() {
		ruleID := tfID.(types.String).ValueString()
		desiredRuleIDs = append(desiredRuleIDs, ruleID)
	}

	// Validate that all existing rules in Datadog are managed by Terraform
	resp, httpResponse, err := r.Api.ListArbitraryCostRules(r.Auth)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing custom allocation rules: %v", httpResponse)))
		return
	}

	var existingRules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		existingRules = *respData
	}

	// Build a map of existing rule IDs
	existingIDs := make(map[string]bool)
	for _, rule := range existingRules {
		if ruleID, ok := rule.GetIdOk(); ok {
			existingIDs[*ruleID] = true
		}
	}

	// Validate that all specified rule IDs exist
	for _, ruleID := range desiredRuleIDs {
		if !existingIDs[ruleID] {
			diag.AddError("Invalid rule ID", fmt.Sprintf("rule ID %s does not exist", ruleID))
			return
		}
	}

	// Ensure all rules in Datadog are included in this configuration
	if len(desiredRuleIDs) != len(existingIDs) {
		desiredIDsMap := make(map[string]bool)
		for _, id := range desiredRuleIDs {
			desiredIDsMap[id] = true
		}

		var unmanagedRules []string
		for id := range existingIDs {
			if !desiredIDsMap[id] {
				unmanagedRules = append(unmanagedRules, id)
			}
		}

		diag.AddError(
			"Unmanaged rules detected",
			fmt.Sprintf("Found %d custom allocation rules in Datadog that are not managed by this Terraform configuration: %v. "+
				"All custom allocation rules must be managed by Terraform. Please either:\n"+
				"1. Import existing rules using 'terraform import datadog_custom_allocation_rule.<name> <rule_id>'\n"+
				"2. Add the missing rules to your Terraform configuration\n"+
				"3. Delete unmanaged rules from Datadog if they're no longer needed\n\n"+
				"This ensures complete infrastructure control and prevents configuration drift.",
				len(unmanagedRules), unmanagedRules))
		return
	}

	// Build the reorder request with rules in the specified order
	ruleData := make([]datadogV2.ReorderRuleResourceData, len(desiredRuleIDs))
	for i, ruleID := range desiredRuleIDs {
		ruleData[i] = datadogV2.ReorderRuleResourceData{
			Id:   &ruleID,
			Type: datadogV2.REORDERRULERESOURCEDATATYPE_ARBITRARY_RULE,
		}
	}

	reorderRequest := datadogV2.ReorderRuleResourceArray{
		Data: ruleData,
	}
	httpResponse, err = r.Api.ReorderArbitraryCostRules(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering custom allocation rules: %v", httpResponse)))
		return
	}
}
