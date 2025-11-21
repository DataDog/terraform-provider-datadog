package fwprovider

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	_ resource.ResourceWithModifyPlan  = &customAllocationRulesResource{}
)

func NewCustomAllocationRulesResource() resource.Resource {
	return &customAllocationRulesResource{}
}

type customAllocationRulesModel struct {
	ID                         types.String `tfsdk:"id"`
	RuleIDs                    types.List   `tfsdk:"rule_ids"`
	OverrideUIDefinedResources types.Bool   `tfsdk:"override_ui_defined_resources"`
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
				Description: "The list of Custom Allocation Rule IDs, in order. Rules are executed in the order specified in this list. Comes from the `id` field on a `datadog_custom_allocation_rule` resource.",
				ElementType: types.StringType,
				Required:    true,
			},
			"override_ui_defined_resources": schema.BoolAttribute{
				Description: "Whether to override UI-defined rules. When set to true, any rules created via the UI that are not defined in Terraform will be deleted and Terraform will be used as the source of truth for rules and their ordering. When set to false, any rules created via the UI that are at the end of order will be kept but will be warned, otherwise an error will be thrown in terraform plan phase. Default is false",
				Optional:    true,
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

	// Create a map for quick lookup of managed IDs
	managedIDsSet := make(map[string]bool)
	for _, tfID := range state.RuleIDs.Elements() {
		ruleID := tfID.(types.String).ValueString()
		managedIDsSet[ruleID] = true
	}

	// Check override setting to determine how to build state
	override := false
	if !state.OverrideUIDefinedResources.IsNull() {
		override = state.OverrideUIDefinedResources.ValueBool()
	}

	// Get the current list of rules from API to read their order
	resp, httpResponse, err := r.Api.ListCustomAllocationRules(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading custom allocation rules. http response: %v", httpResponse)))
		return
	}

	var rules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		rules = *respData
	}

	// Get rules with positions
	// During import (managedIDsSet is empty): get ALL rules
	// When override=false: only managed rules
	// When override=true: ALL rules (so Terraform can detect difference and trigger Update to delete unmanaged)
	isImport := len(managedIDsSet) == 0
	rulePositions := getRulesWithPositions(rules, managedIDsSet, !override && !isImport)

	// Verify all managed rules still exist
	managedCount := 0
	for _, rp := range rulePositions {
		if managedIDsSet[rp.ID] {
			managedCount++
		}
	}

	if managedCount != len(managedIDsSet) {
		// Some managed rules were deleted
		missingIDs := []string{}
		foundIDs := make(map[string]bool)
		for _, rp := range rulePositions {
			foundIDs[rp.ID] = true
		}
		for id := range managedIDsSet {
			if !foundIDs[id] {
				missingIDs = append(missingIDs, id)
			}
		}
		response.Diagnostics.AddWarning(
			"Managed rules deleted outside Terraform",
			fmt.Sprintf("The following managed rule(s) no longer exist in Datadog: %v. "+
				"They may have been deleted outside of Terraform. "+
				"Run 'terraform apply' to update the state.",
				missingIDs),
		)
	}

	// Extract ordered IDs
	orderedList := make([]string, 0, len(rulePositions))
	for _, rp := range rulePositions {
		orderedList = append(orderedList, rp.ID)
	}

	state.RuleIDs, _ = types.ListValueFrom(ctx, types.StringType, orderedList)
	state.ID = types.StringValue("order") // Static ID like other order resources

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

func (r *customAllocationRulesResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// Show plan warnings during create/update operations
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var plan customAllocationRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if override is enabled
	override := false
	if !plan.OverrideUIDefinedResources.IsNull() {
		override = plan.OverrideUIDefinedResources.ValueBool()
	}

	// Get all existing rules to check for unmanaged ones
	resp, httpResponse, err := r.Api.ListCustomAllocationRules(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing custom allocation rules during plan: %v", httpResponse)))
		return
	}

	var existingRules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		existingRules = *respData
	}

	// Create a map of desired IDs for checking unmanaged rules
	desiredIDsSet := make(map[string]bool)
	for _, tfID := range plan.RuleIDs.Elements() {
		ruleID := tfID.(types.String).ValueString()
		desiredIDsSet[ruleID] = true
	}

	// Get all rules with positions sorted
	allRulePositions := getRulesWithPositions(existingRules, desiredIDsSet, false)

	// Find unmanaged rules
	unmanagedInfo := findUnmanagedRules(allRulePositions, desiredIDsSet)

	if len(unmanagedInfo.Rules) > 0 {
		// Format the list nicely
		unmanagedDetails := formatUnmanagedRuleDetails(unmanagedInfo.Rules, false)
		detailsList := "  • " + strings.Join(unmanagedDetails, "\n  • ") + "\n"

		if override {
			// With override=true, warn about deletion (unmanaged can be anywhere)
			response.Diagnostics.AddWarning(
				"UI-defined rules will be deleted",
				fmt.Sprintf("The following %d rule(s) will be deleted because override_ui_defined_resources is set to true:\n\n%s\n"+
					"These rules exist in Datadog but are not defined in your Terraform configuration. "+
					"When you run 'terraform apply', they will be permanently deleted.",
					len(unmanagedInfo.Rules),
					detailsList),
			)
		} else {
			// With override=false, need to check position constraints
			if !unmanagedInfo.AllAtEnd {
				// Unmanaged in middle - ERROR
				response.Diagnostics.AddError(
					"Unmanaged rules detected in the middle of order",
					fmt.Sprintf("Found %d rules in Datadog that are not managed by this Terraform configuration and are not at the end of the order.\n\n"+
						"Current order: %v\n"+
						"Unmanaged rules: %v at positions: %v\n\n"+
						"To fix this, either:\n"+
						"1. Set override_ui_defined_resources=true to automatically delete unmanaged rules\n"+
						"2. Import the unmanaged rules into Terraform\n"+
						"3. Manually reorder or delete the unmanaged rules in Datadog UI",
						len(unmanagedInfo.Rules),
						allRulePositions, unmanagedInfo.Rules, unmanagedInfo.Positions),
				)
			} else {
				// Unmanaged at end - WARNING
				response.Diagnostics.AddWarning(
					"Unmanaged rules detected at the end of order",
					fmt.Sprintf("Found %d unmanaged rule(s) at the end of the order:\n\n%s\n"+
						"These rules are not managed by Terraform. Consider:\n"+
						"1. Importing them: terraform import datadog_custom_allocation_rule.<name> <rule_id>\n"+
						"2. Deleting them from Datadog if not needed\n"+
						"3. Setting override_ui_defined_resources=true to automatically delete them",
						len(unmanagedInfo.Rules),
						detailsList),
				)
			}
		}
	}
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

	// Check override parameter - default to false for current behavior
	override := false
	if !state.OverrideUIDefinedResources.IsNull() {
		override = state.OverrideUIDefinedResources.ValueBool()
	}

	if override {
		r.updateOrderWithDeletion(state, diag, desiredRuleIDs)
	} else {
		r.updateOrderWithAllRules(state, diag, desiredRuleIDs)
	}
}

// ruleWithPosition is a helper struct to track rule ID, position, and name
type ruleWithPosition struct {
	ID       string `json:"id"`
	Position int64  `json:"position"`
	Name     string `json:"name"`
}

// unmanagedRuleInfo contains information about unmanaged rules and their positions
type unmanagedRuleInfo struct {
	Rules     []ruleWithPosition
	Positions []int
	AllAtEnd  bool
}

// extractRuleFields extracts ID, name, and order_id from a rule in a single pass
func extractRuleFields(rule datadogV2.ArbitraryRuleResponseData) (id string, name string, orderId int64, ok bool) {
	// Try to get data from normal fields first
	if ruleID, idOk := rule.GetIdOk(); idOk && ruleID != nil && *ruleID != "" {
		id = *ruleID
		ok = true

		if attrs, attrsOk := rule.GetAttributesOk(); attrsOk {
			if ruleName, nameOk := attrs.GetRuleNameOk(); nameOk {
				name = *ruleName
			}
			if orderIdPtr, orderOk := attrs.GetOrderIdOk(); orderOk {
				orderId = *orderIdPtr
			}
			return
		}
	}

	// If normal fields failed, try UnparsedObject
	// This happens when the API returns fields the generated client doesn't know about
	if rule.UnparsedObject != nil {
		// Extract ID
		if idVal, idOk := rule.UnparsedObject["id"].(string); idOk && idVal != "" {
			id = idVal
			ok = true
		}

		// Extract attributes (name and order_id)
		if attributesRaw, attrsOk := rule.UnparsedObject["attributes"].(map[string]interface{}); attrsOk {
			// Extract rule_name
			if nameVal, nameOk := attributesRaw["rule_name"].(string); nameOk {
				name = nameVal
			}

			// Extract order_id (might be float64, int, or int64)
			if orderFloat, orderOk := attributesRaw["order_id"].(float64); orderOk {
				orderId = int64(orderFloat)
			} else if orderInt, orderOk := attributesRaw["order_id"].(int); orderOk {
				orderId = int64(orderInt)
			} else if orderInt64, orderOk := attributesRaw["order_id"].(int64); orderOk {
				orderId = orderInt64
			}
		}
	}

	return
}

// extractRuleID extracts the ID from a rule (convenience wrapper around extractRuleFields)
func extractRuleID(rule datadogV2.ArbitraryRuleResponseData) (string, bool) {
	id, _, _, ok := extractRuleFields(rule)
	return id, ok
}

// getRulesWithPositions extracts all rules with their positions and sorts them by order_id
// If managedOnly is true, only include rules in the managedIDsSet
func getRulesWithPositions(rules []datadogV2.ArbitraryRuleResponseData, managedIDsSet map[string]bool, managedOnly bool) []ruleWithPosition {
	result := make([]ruleWithPosition, 0, len(rules))

	for _, rule := range rules {
		ruleID, name, orderId, ok := extractRuleFields(rule)
		if !ok {
			continue
		}

		// Skip unmanaged if managedOnly is true
		if managedOnly && !managedIDsSet[ruleID] {
			continue
		}

		result = append(result, ruleWithPosition{
			ID:       ruleID,
			Position: orderId,
			Name:     name,
		})
	}

	// Sort by position
	sortRulesByPosition(result)

	return result
}

// sortRulesByPosition sorts rules by their order_id field (in-place)
func sortRulesByPosition(rules []ruleWithPosition) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Position < rules[j].Position
	})
}

// findUnmanagedRules identifies unmanaged rules and checks if they are all at the end
func findUnmanagedRules(allRules []ruleWithPosition, managedIDsSet map[string]bool) unmanagedRuleInfo {
	var unmanagedRules []ruleWithPosition
	unmanagedPositions := make([]int, 0)

	for i, rp := range allRules {
		if !managedIDsSet[rp.ID] {
			unmanagedRules = append(unmanagedRules, rp)
			unmanagedPositions = append(unmanagedPositions, i)
		}
	}

	// Check if all unmanaged are at the end
	// Since allRules is sorted by position, if unmanaged rules are all at the end,
	// they should form a contiguous block. We only need to check if the first unmanaged
	// rule starts at the expected position.
	allAtEnd := false
	if len(unmanagedRules) > 0 {
		firstUnmanagedPos := unmanagedPositions[0]
		expectedStartPos := len(allRules) - len(unmanagedRules)
		allAtEnd = firstUnmanagedPos == expectedStartPos
	}

	return unmanagedRuleInfo{
		Rules:     unmanagedRules,
		Positions: unmanagedPositions,
		AllAtEnd:  allAtEnd,
	}
}

// formatUnmanagedRuleDetails creates a formatted list of unmanaged rules for display
func formatUnmanagedRuleDetails(unmanagedRules []ruleWithPosition, includePosition bool) []string {
	details := make([]string, 0, len(unmanagedRules))
	for _, ur := range unmanagedRules {
		if ur.Name != "" {
			if includePosition {
				details = append(details, fmt.Sprintf("'%s' (ID: %s, Position: %d)", ur.Name, ur.ID, ur.Position))
			} else {
				details = append(details, fmt.Sprintf("'%s' (%s)", ur.Name, ur.ID))
			}
		} else {
			if includePosition {
				details = append(details, fmt.Sprintf("ID: %s (Position: %d)", ur.ID, ur.Position))
			} else {
				details = append(details, ur.ID)
			}
		}
	}
	return details
}

// Deletes unmanaged rules and reorders remaining ones when override is enabled
func (r *customAllocationRulesResource) updateOrderWithDeletion(state *customAllocationRulesModel, diag *diag.Diagnostics, desiredOrder []string) {
	// Get all existing rules
	resp, httpResponse, err := r.Api.ListCustomAllocationRules(r.Auth)
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
	var existingIDList []string

	// Extract IDs from all rules, handling UnparsedObject for problematic rules
	for _, rule := range existingRules {
		ruleID, ok := extractRuleID(rule)
		if ok {
			existingIDs[ruleID] = true
			existingIDList = append(existingIDList, ruleID)
		}
	}

	// Validate desired rules exist
	for _, ruleID := range desiredOrder {
		if !existingIDs[ruleID] {
			diag.AddError("Invalid rule ID", fmt.Sprintf("rule ID %s does not exist", ruleID))
			return
		}
	}

	// Create a map of desired IDs for checking unmanaged rules
	desiredIDsSet := make(map[string]bool)
	for _, id := range desiredOrder {
		desiredIDsSet[id] = true
	}

	// Find and delete unmanaged rules
	var unmanagedRules []string
	for _, rule := range existingRules {
		ruleID, ok := extractRuleID(rule)
		if !ok {
			continue
		}

		if !desiredIDsSet[ruleID] {
			unmanagedRules = append(unmanagedRules, ruleID)
		}
	}

	// Delete unmanaged rules
	for _, ruleID := range unmanagedRules {
		id, parseErr := strconv.ParseInt(ruleID, 10, 64)
		if parseErr != nil {
			diag.AddError("Invalid rule ID", fmt.Sprintf("rule ID %s is not a valid integer: %v", ruleID, parseErr))
			return
		}
		httpResp, err := r.Api.DeleteCustomAllocationRule(r.Auth, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				// Resource already deleted - continue
				continue
			}
			diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error deleting unmanaged rule %s: %v", ruleID, httpResp)))
			return
		}
	}

	// Now reorder the remaining rules
	ruleData := make([]datadogV2.ReorderRuleResourceData, len(desiredOrder))
	for i, ruleID := range desiredOrder {
		ruleData[i] = datadogV2.ReorderRuleResourceData{
			Id:   &ruleID,
			Type: datadogV2.REORDERRULERESOURCEDATATYPE_ARBITRARY_RULE,
		}
	}

	reorderRequest := datadogV2.ReorderRuleResourceArray{
		Data: ruleData,
	}

	httpResponse, err = r.Api.ReorderCustomAllocationRules(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering custom allocation rules: %v", httpResponse)))
		return
	}
}

func (r *customAllocationRulesResource) updateOrderWithAllRules(state *customAllocationRulesModel, diag *diag.Diagnostics, desiredOrder []string) {
	// Get all existing rules
	resp, httpResponse, err := r.Api.ListCustomAllocationRules(r.Auth)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing custom allocation rules: %v", httpResponse)))
		return
	}

	var existingRules []datadogV2.ArbitraryRuleResponseData
	if respData, ok := resp.GetDataOk(); ok {
		existingRules = *respData
	}

	// Create a map of desired IDs for quick lookup
	desiredIDsSet := make(map[string]bool)
	for _, id := range desiredOrder {
		desiredIDsSet[id] = true
	}

	// Get all rules with positions (sorted)
	rulePositions := getRulesWithPositions(existingRules, desiredIDsSet, false)

	// Validate desired rules exist
	existingIDs := make(map[string]bool)
	for _, rp := range rulePositions {
		existingIDs[rp.ID] = true
	}

	for _, ruleID := range desiredOrder {
		if !existingIDs[ruleID] {
			diag.AddError("Invalid rule ID", fmt.Sprintf("rule ID %s does not exist", ruleID))
			return
		}
	}

	// Check if there are unmanaged rules in the middle of the order
	if len(existingRules) > len(desiredOrder) {
		// Find unmanaged rules and check positions
		unmanagedInfo := findUnmanagedRules(rulePositions, desiredIDsSet)

		if len(unmanagedInfo.Rules) > 0 {
			if !unmanagedInfo.AllAtEnd {
				diag.AddError(
					"Unmanaged rules detected in the middle of order",
					fmt.Sprintf("Found %d rules in Datadog that are not managed by this Terraform configuration and are not all at the end of the order: %v. "+
						"Unmanaged rules must be at the end of the order to allow reordering. Please either:\n"+
						"1. Add the unmanaged rules to your Terraform configuration and import them using 'terraform import datadog_custom_allocation_rule.<name> <rule_id>'\n"+
						"2. Delete unmanaged rules from Datadog if they're no longer needed\n"+
						"3. Set override_ui_defined_resources=true to automatically delete unmanaged rules\n"+
						"4. Manually reorder the rules in Datadog UI so all unmanaged rules are at the end\n\n"+
						"Current order positions: %v\n"+
						"Unmanaged rules: %v at positions: %v\n"+
						"Expected position for unmanaged rules to start: %d",
						len(unmanagedInfo.Rules), unmanagedInfo.Rules,
						rulePositions, unmanagedInfo.Rules, unmanagedInfo.Positions, len(rulePositions)-len(unmanagedInfo.Rules)),
				)
				return
			}

			// Warning: Notify user about unmanaged rules even though they're at the end
			unmanagedDetails := formatUnmanagedRuleDetails(unmanagedInfo.Rules, true)

			diag.AddWarning(
				"Unmanaged rules detected",
				fmt.Sprintf("Found %d rule(s) in Datadog that are not managed by this Terraform configuration:\n\n"+
					"%s\n\n"+
					"These rules are currently at the end of the order, so they won't block this operation. However, "+
					"to ensure complete infrastructure management and prevent configuration drift, consider:\n"+
					"1. Importing them into Terraform: 'terraform import datadog_custom_allocation_rule.<name> <rule_id>'\n"+
					"2. Deleting them from Datadog if they're no longer needed\n"+
					"3. Setting override_ui_defined_resources=true to automatically delete unmanaged rules\n\n"+
					"Your managed rules will be placed first, followed by these unmanaged rules.",
					len(unmanagedInfo.Rules),
					"  • "+strings.Join(unmanagedDetails, "\n  • ")),
			)
		}
	}

	// Create final order: desired rules first, then remaining rules in their current order
	finalOrder := make([]string, 0, len(rulePositions))

	// Add desired rules in the specified order
	finalOrder = append(finalOrder, desiredOrder...)

	// Add remaining unmanaged rules in their current order
	for _, rp := range rulePositions {
		if !desiredIDsSet[rp.ID] {
			finalOrder = append(finalOrder, rp.ID)
		}
	}

	// Build reorder request with ALL rules
	ruleData := make([]datadogV2.ReorderRuleResourceData, len(finalOrder))
	for i, ruleID := range finalOrder {
		ruleData[i] = datadogV2.ReorderRuleResourceData{
			Id:   &ruleID,
			Type: datadogV2.REORDERRULERESOURCEDATATYPE_ARBITRARY_RULE,
		}
	}

	reorderRequest := datadogV2.ReorderRuleResourceArray{
		Data: ruleData,
	}

	httpResponse, err = r.Api.ReorderCustomAllocationRules(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering custom allocation rules with all rules: %v", httpResponse)))
		return
	}
}
