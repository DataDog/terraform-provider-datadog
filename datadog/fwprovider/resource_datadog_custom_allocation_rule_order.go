package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &customAllocationRuleOrderResource{}
	_ resource.ResourceWithImportState = &customAllocationRuleOrderResource{}
)

func NewCustomAllocationRuleOrderResource() resource.Resource {
	return &customAllocationRuleOrderResource{}
}

type customAllocationRuleOrderModel struct {
	ID      types.String `tfsdk:"id"`
	RuleIDs types.List   `tfsdk:"rule_ids"`
}

type customAllocationRuleOrderResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func (r *customAllocationRuleOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *customAllocationRuleOrderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "custom_allocation_rule_order"
}

func (r *customAllocationRuleOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (r *customAllocationRuleOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state customAllocationRuleOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *customAllocationRuleOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state customAllocationRuleOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the current list of rules to read their order
	resp, httpResponse, err := r.Api.ListArbitraryCostRules(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading custom allocation rules. http response: %v", httpResponse)))
		return
	}

	var rules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		rules = *respData
	}

	// Sort rules by order_id and extract IDs
	tfList := make([]string, len(rules))
	for _, rule := range rules {
		if ruleAttrs, ok := rule.GetAttributesOk(); ok {
			// Find the correct position for this rule
			orderId := int(ruleAttrs.GetOrderId())
			if orderId < len(tfList) {
				if ruleID, ok := rule.GetIdOk(); ok {
					tfList[orderId] = *ruleID
				}
			}
		}
	}

	// Remove any empty slots and create a clean ordered list
	orderedList := make([]string, 0, len(tfList))
	for _, id := range tfList {
		if id != "" {
			orderedList = append(orderedList, id)
		}
	}

	state.RuleIDs, _ = types.ListValueFrom(ctx, types.StringType, orderedList)
	state.ID = types.StringValue("order") // Static ID like other order resources

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *customAllocationRuleOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state customAllocationRuleOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *customAllocationRuleOrderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// No-op: Deleting the order resource doesn't change the actual order of rules
	// This follows the same pattern as other order resources in the provider
}

func (r *customAllocationRuleOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *customAllocationRuleOrderResource) updateOrder(state *customAllocationRuleOrderModel, diag *diag.Diagnostics) {
	// Set the ID immediately to prevent "unknown value" errors
	state.ID = types.StringValue("order")

	// Convert the Terraform list to strings for the API call
	var desiredRuleIDs []string
	for _, tfID := range state.RuleIDs.Elements() {
		ruleID := tfID.(types.String).ValueString()
		desiredRuleIDs = append(desiredRuleIDs, ruleID)
	}

	// The custom allocation rule API requires ALL existing rules to be included in the reorder call
	// So we always use the comprehensive approach
	r.updateOrderWithAllRules(state, diag, desiredRuleIDs)
}

// Fallback method that includes all existing rules
func (r *customAllocationRuleOrderResource) updateOrderWithAllRules(state *customAllocationRuleOrderModel, diag *diag.Diagnostics, desiredOrder []string) {
	// Get all existing rules
	resp, httpResponse, err := r.Api.ListArbitraryCostRules(r.Auth)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing custom allocation rules: %v", httpResponse)))
		return
	}

	var existingRules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		existingRules = *respData
	}

	// Create a map of existing rule IDs for validation
	existingIDs := make(map[string]bool)
	for _, rule := range existingRules {
		if ruleID, ok := rule.GetIdOk(); ok {
			existingIDs[*ruleID] = true
		}
	}

	// Validate desired rules exist
	for _, ruleID := range desiredOrder {
		if !existingIDs[ruleID] {
			diag.AddError("Invalid rule ID", fmt.Sprintf("rule ID %s does not exist", ruleID))
			return
		}
	}

	// Create a slice of structs to sort by current order_id
	type ruleWithOrderId struct {
		id      string
		orderId int64
	}

	ruleOrderIds := make([]ruleWithOrderId, 0, len(existingRules))
	for _, rule := range existingRules {
		if ruleID, ok := rule.GetIdOk(); ok {
			orderId := int64(0)
			if ruleAttrs, ok := rule.GetAttributesOk(); ok {
				orderId = ruleAttrs.GetOrderId()
			}
			ruleOrderIds = append(ruleOrderIds, ruleWithOrderId{
				id:      *ruleID,
				orderId: orderId,
			})
		}
	}

	// Sort by order_id to get current order
	for i := 0; i < len(ruleOrderIds); i++ {
		for j := i + 1; j < len(ruleOrderIds); j++ {
			if ruleOrderIds[i].orderId > ruleOrderIds[j].orderId {
				ruleOrderIds[i], ruleOrderIds[j] = ruleOrderIds[j], ruleOrderIds[i]
			}
		}
	}

	// Create final order: desired order first, then remaining rules
	finalOrder := make([]string, 0, len(ruleOrderIds))
	desiredIDsMap := make(map[string]bool)

	// Add desired rules in specified order
	for _, id := range desiredOrder {
		finalOrder = append(finalOrder, id)
		desiredIDsMap[id] = true
	}

	// Add remaining rules in their current order
	for _, ro := range ruleOrderIds {
		if !desiredIDsMap[ro.id] {
			finalOrder = append(finalOrder, ro.id)
		}
	}

	// Convert to API format
	ruleData := make([]datadogV2.ReorderRuleResourceData, len(finalOrder))
	for i, ruleID := range finalOrder {
		ruleData[i] = datadogV2.ReorderRuleResourceData{
			Id:   &ruleID,
			Type: datadogV2.REORDERRULERESOURCEDATATYPE_ARBITRARY_RULE,
		}
	}

	// Create the reorder request
	reorderRequest := datadogV2.ReorderRuleResourceArray{
		Data: ruleData,
	}

	// Call the reorder API
	httpResponse, err = r.Api.ReorderArbitraryCostRules(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering custom allocation rules with all rules: %v", httpResponse)))
		return
	}
}
