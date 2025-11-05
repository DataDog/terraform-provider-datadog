package fwprovider

import (
	"context"
	"fmt"
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
	_ resource.ResourceWithConfigure   = &tagPipelineRulesetsResource{}
	_ resource.ResourceWithImportState = &tagPipelineRulesetsResource{}
	_ resource.ResourceWithModifyPlan  = &tagPipelineRulesetsResource{}
)

func NewTagPipelineRulesetsResource() resource.Resource {
	return &tagPipelineRulesetsResource{}
}

type tagPipelineRulesetsModel struct {
	ID                         types.String `tfsdk:"id"`
	RulesetIDs                 types.List   `tfsdk:"ruleset_ids"`
	OverrideUIDefinedResources types.Bool   `tfsdk:"override_ui_defined_resources"`
}

type tagPipelineRulesetsResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func (r *tagPipelineRulesetsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *tagPipelineRulesetsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "tag_pipeline_rulesets"
}

func (r *tagPipelineRulesetsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Tag Pipeline Ruleset Order resource that can be used to manage the order of Tag Pipeline Rulesets.",
		Attributes: map[string]schema.Attribute{
			"ruleset_ids": schema.ListAttribute{
				Description: "The list of Tag Pipeline Ruleset IDs, in order. Rulesets are executed in the order specified in this list.",
				ElementType: types.StringType,
				Required:    true,
			},
			"override_ui_defined_resources": schema.BoolAttribute{
				Description: "Whether to override UI-defined rulesets. When set to true, any rulesets created via the UI that are not defined in Terraform will be deleted and perform reorder based on the rulesets from the terraform. Default is false",
				Optional:    true,
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *tagPipelineRulesetsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state tagPipelineRulesetsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tagPipelineRulesetsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the managed ruleset IDs from the current state
	var managedRulesetIDs []string
	for _, tfID := range state.RulesetIDs.Elements() {
		rulesetID := tfID.(types.String).ValueString()
		managedRulesetIDs = append(managedRulesetIDs, rulesetID)
	}

	// Check override setting to determine how to build state
	override := false
	if !state.OverrideUIDefinedResources.IsNull() {
		override = state.OverrideUIDefinedResources.ValueBool()
	}

	// Get the current list of rulesets to read their order
	resp, httpResponse, err := r.Api.ListTagPipelinesRulesets(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading tag pipeline rulesets. http response: %v", httpResponse)))
		return
	}

	var rulesets []datadogV2.RulesetRespData
	if respData, ok := resp.GetDataOk(); ok {
		rulesets = *respData
	}

	// Create a map for quick lookup of managed IDs
	managedIDsSet := make(map[string]bool)
	for _, id := range managedRulesetIDs {
		managedIDsSet[id] = true
	}

	// Get rulesets with positions
	// During import (managedRulesetIDs is empty): get ALL rulesets
	// When override=false: only managed rulesets
	// When override=true: ALL rulesets (so Terraform can detect difference and trigger Update to delete unmanaged)
	isImport := len(managedRulesetIDs) == 0
	rulesetPositions := getRulesetsWithPositions(rulesets, managedIDsSet, !override && !isImport)

	// Verify all managed rulesets still exist
	managedCount := 0
	for _, rp := range rulesetPositions {
		if managedIDsSet[rp.ID] {
			managedCount++
		}
	}

	if managedCount != len(managedRulesetIDs) {
		// Some managed rulesets were deleted
		missingIDs := []string{}
		foundIDs := make(map[string]bool)
		for _, rp := range rulesetPositions {
			foundIDs[rp.ID] = true
		}
		for _, id := range managedRulesetIDs {
			if !foundIDs[id] {
				missingIDs = append(missingIDs, id)
			}
		}
		response.Diagnostics.AddWarning(
			"Managed rulesets deleted outside Terraform",
			fmt.Sprintf("The following managed ruleset(s) no longer exist in Datadog: %v. "+
				"They may have been deleted outside of Terraform. "+
				"Run 'terraform apply' to update the state.",
				missingIDs),
		)
	}

	// Extract ordered IDs
	orderedList := make([]string, 0, len(rulesetPositions))
	for _, rp := range rulesetPositions {
		orderedList = append(orderedList, rp.ID)
	}

	state.RulesetIDs, _ = types.ListValueFrom(ctx, types.StringType, orderedList)
	state.ID = types.StringValue("order") // Static ID like other order resources

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state tagPipelineRulesetsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// No-op: Deleting the order resource doesn't change the actual order of rulesets
	// This follows the same pattern as other order resources in the provider
}

func (r *tagPipelineRulesetsResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// Show plan warnings during create/update operations
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var plan tagPipelineRulesetsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Check if override is enabled
	override := false
	if !plan.OverrideUIDefinedResources.IsNull() {
		override = plan.OverrideUIDefinedResources.ValueBool()
	}

	// Get all existing rulesets to check for unmanaged ones
	resp, httpResponse, err := r.Api.ListTagPipelinesRulesets(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing tag pipeline rulesets during plan: %v", httpResponse)))
		return
	}

	var existingRulesets []datadogV2.RulesetRespData
	if respData, ok := resp.GetDataOk(); ok {
		existingRulesets = *respData
	}

	// Convert the Terraform list to strings
	var desiredRulesetIDs []string
	for _, tfID := range plan.RulesetIDs.Elements() {
		rulesetID := tfID.(types.String).ValueString()
		desiredRulesetIDs = append(desiredRulesetIDs, rulesetID)
	}

	// Create a map of desired IDs for checking unmanaged rulesets
	desiredIDsSet := make(map[string]bool)
	for _, id := range desiredRulesetIDs {
		desiredIDsSet[id] = true
	}

	// Get all rulesets with positions sorted
	allRulesetPositions := getRulesetsWithPositions(existingRulesets, desiredIDsSet, false)

	// Find unmanaged rulesets
	unmanagedInfo := findUnmanagedRulesets(allRulesetPositions, desiredIDsSet)

	if len(unmanagedInfo.Rulesets) > 0 {
		// Format the list nicely
		unmanagedDetails := formatUnmanagedDetails(unmanagedInfo.Rulesets, false)
		detailsList := "  • " + strings.Join(unmanagedDetails, "\n  • ") + "\n"

		if override {
			// With override=true, warn about deletion (unmanaged can be anywhere)
			response.Diagnostics.AddWarning(
				"UI-defined rulesets will be deleted",
				fmt.Sprintf("The following %d ruleset(s) will be deleted because override_ui_defined_resources is set to true:\n\n%s\n"+
					"These rulesets exist in Datadog but are not defined in your Terraform configuration. "+
					"When you run 'terraform apply', they will be permanently deleted.",
					len(unmanagedInfo.Rulesets),
					detailsList),
			)
		} else {
			// With override=false, need to check position constraints
			if !unmanagedInfo.AllAtEnd {
				// Unmanaged in middle - ERROR
				response.Diagnostics.AddError(
					"Unmanaged rulesets detected in the middle of order",
					fmt.Sprintf("Found %d rulesets in Datadog that are not managed by this Terraform configuration and are not at the end of the order.\n\n"+
						"Current order: %v\n"+
						"Unmanaged rulesets: %v at positions: %v\n\n"+
						"To fix this, either:\n"+
						"1. Set override_ui_defined_resources=true to automatically delete unmanaged rulesets\n"+
						"2. Import the unmanaged rulesets into Terraform\n"+
						"3. Manually reorder or delete the unmanaged rulesets in Datadog UI",
						len(unmanagedInfo.Rulesets),
						allRulesetPositions, unmanagedInfo.Rulesets, unmanagedInfo.Positions),
				)
			} else {
				// Unmanaged at end - WARNING
				response.Diagnostics.AddWarning(
					"Unmanaged rulesets detected at the end of order",
					fmt.Sprintf("Found %d unmanaged ruleset(s) at the end of the order:\n\n%s\n"+
						"These rulesets are not managed by Terraform. Consider:\n"+
						"1. Importing them: terraform import datadog_tag_pipeline_ruleset.<name> <ruleset_id>\n"+
						"2. Deleting them from Datadog if not needed\n"+
						"3. Setting override_ui_defined_resources=true to automatically delete them",
						len(unmanagedInfo.Rulesets),
						detailsList),
				)
			}
		}
	}
}

func (r *tagPipelineRulesetsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

// rulesetWithPosition is a helper struct to track ruleset ID, position, and name
type rulesetWithPosition struct {
	ID       string `json:"id"`
	Position int32  `json:"position"`
	Name     string `json:"name"`
}

// unmanagedRulesetInfo contains information about unmanaged rulesets and their positions
type unmanagedRulesetInfo struct {
	Rulesets  []rulesetWithPosition
	Positions []int
	AllAtEnd  bool
}

// extractRulesetFields extracts ID, name, and position from a ruleset in a single pass,
// handling the case where the API client failed to deserialize and put the data in UnparsedObject
// Returns: (id, name, position, ok) where ok indicates if ID was successfully extracted
func extractRulesetFields(ruleset datadogV2.RulesetRespData) (id string, name string, position int32, ok bool) {
	// Try to get data from normal fields first
	if rulesetID, idOk := ruleset.GetIdOk(); idOk && rulesetID != nil && *rulesetID != "" {
		id = *rulesetID
		ok = true

		if attrs, attrsOk := ruleset.GetAttributesOk(); attrsOk {
			name = attrs.GetName()
			position = attrs.GetPosition()
			return
		}
	}

	// If normal fields failed, try UnparsedObject
	// This happens when the API returns fields the generated client doesn't know about
	if ruleset.UnparsedObject != nil {
		// Extract ID
		if idVal, idOk := ruleset.UnparsedObject["id"].(string); idOk && idVal != "" {
			id = idVal
			ok = true
		}

		// Extract attributes (name and position)
		if attributesRaw, attrsOk := ruleset.UnparsedObject["attributes"].(map[string]interface{}); attrsOk {
			// Extract name
			if nameVal, nameOk := attributesRaw["name"].(string); nameOk {
				name = nameVal
			}

			// Extract position (might be float64, int, or int32)
			if posFloat, posOk := attributesRaw["position"].(float64); posOk {
				position = int32(posFloat)
			} else if posInt, posOk := attributesRaw["position"].(int); posOk {
				position = int32(posInt)
			} else if posInt32, posOk := attributesRaw["position"].(int32); posOk {
				position = posInt32
			}
		}
	}

	return
}

// extractRulesetID extracts the ID from a ruleset (convenience wrapper around extractRulesetFields)
func extractRulesetID(ruleset datadogV2.RulesetRespData) (string, bool) {
	id, _, _, ok := extractRulesetFields(ruleset)
	return id, ok
}

// extractRulesetName extracts the name from a ruleset (convenience wrapper around extractRulesetFields)
func extractRulesetName(ruleset datadogV2.RulesetRespData) string {
	_, name, _, _ := extractRulesetFields(ruleset)
	return name
}

// extractRulesetPosition extracts the position from a ruleset (convenience wrapper around extractRulesetFields)
func extractRulesetPosition(ruleset datadogV2.RulesetRespData) int32 {
	_, _, position, _ := extractRulesetFields(ruleset)
	return position
}

// getRulesetsWithPositions extracts all rulesets with their positions and sorts them by position
// If managedOnly is true, only include rulesets in the managedIDsSet
func getRulesetsWithPositions(rulesets []datadogV2.RulesetRespData, managedIDsSet map[string]bool, managedOnly bool) []rulesetWithPosition {
	result := make([]rulesetWithPosition, 0, len(rulesets))

	for _, ruleset := range rulesets {
		rulesetID, name, position, ok := extractRulesetFields(ruleset)
		if !ok {
			continue
		}

		// Skip unmanaged if managedOnly is true
		if managedOnly && !managedIDsSet[rulesetID] {
			continue
		}

		result = append(result, rulesetWithPosition{
			ID:       rulesetID,
			Position: position,
			Name:     name,
		})
	}

	// Sort by position
	sortRulesetsByPosition(result)

	return result
}

// sortRulesetsByPosition sorts rulesets by their position field (in-place)
func sortRulesetsByPosition(rulesets []rulesetWithPosition) {
	for i := 0; i < len(rulesets); i++ {
		for j := i + 1; j < len(rulesets); j++ {
			if rulesets[i].Position > rulesets[j].Position {
				rulesets[i], rulesets[j] = rulesets[j], rulesets[i]
			}
		}
	}
}

// findUnmanagedRulesets identifies unmanaged rulesets and checks if they are all at the end
func findUnmanagedRulesets(allRulesets []rulesetWithPosition, managedIDsSet map[string]bool) unmanagedRulesetInfo {
	var unmanagedRulesets []rulesetWithPosition
	unmanagedPositions := make([]int, 0)

	for i, rp := range allRulesets {
		if !managedIDsSet[rp.ID] {
			unmanagedRulesets = append(unmanagedRulesets, rp)
			unmanagedPositions = append(unmanagedPositions, i)
		}
	}

	// Check if all unmanaged are at the end
	allAtEnd := false
	if len(unmanagedRulesets) > 0 {
		firstUnmanagedPos := unmanagedPositions[0]
		expectedStartPos := len(allRulesets) - len(unmanagedRulesets)

		allAtEnd = firstUnmanagedPos == expectedStartPos
		if allAtEnd {
			for i, pos := range unmanagedPositions {
				if pos != expectedStartPos+i {
					allAtEnd = false
					break
				}
			}
		}
	}

	return unmanagedRulesetInfo{
		Rulesets:  unmanagedRulesets,
		Positions: unmanagedPositions,
		AllAtEnd:  allAtEnd,
	}
}

// formatUnmanagedDetails creates a formatted list of unmanaged rulesets for display
func formatUnmanagedDetails(unmanagedRulesets []rulesetWithPosition, includePosition bool) []string {
	details := make([]string, 0, len(unmanagedRulesets))
	for _, urs := range unmanagedRulesets {
		if urs.Name != "" {
			if includePosition {
				details = append(details, fmt.Sprintf("'%s' (ID: %s, Position: %d)", urs.Name, urs.ID, urs.Position))
			} else {
				details = append(details, fmt.Sprintf("'%s' (%s)", urs.Name, urs.ID))
			}
		} else {
			if includePosition {
				details = append(details, fmt.Sprintf("ID: %s (Position: %d)", urs.ID, urs.Position))
			} else {
				details = append(details, urs.ID)
			}
		}
	}
	return details
}

func (r *tagPipelineRulesetsResource) updateOrder(state *tagPipelineRulesetsModel, diag *diag.Diagnostics) {
	// Set the ID immediately to prevent "unknown value" errors
	state.ID = types.StringValue("order")

	// Convert the Terraform list to strings for the API call
	var desiredRulesetIDs []string
	for _, tfID := range state.RulesetIDs.Elements() {
		rulesetID := tfID.(types.String).ValueString()
		desiredRulesetIDs = append(desiredRulesetIDs, rulesetID)
	}

	// Check override parameter - default to false for current behavior
	override := false
	if !state.OverrideUIDefinedResources.IsNull() {
		override = state.OverrideUIDefinedResources.ValueBool()
	}

	if override {
		r.updateOrderWithDeletion(state, diag, desiredRulesetIDs)
	} else {
		r.updateOrderWithAllRulesets(state, diag, desiredRulesetIDs)
	}
}

// Deletes unmanaged rulesets and reorders remaining ones when override is enabled
func (r *tagPipelineRulesetsResource) updateOrderWithDeletion(state *tagPipelineRulesetsModel, diag *diag.Diagnostics, desiredOrder []string) {
	// Get all existing rulesets
	resp, httpResponse, err := r.Api.ListTagPipelinesRulesets(r.Auth)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing tag pipeline rulesets: %v", httpResponse)))
		return
	}

	var existingRulesets []datadogV2.RulesetRespData
	if respData, ok := resp.GetDataOk(); ok {
		existingRulesets = *respData
	}

	// Create a map of existing ruleset IDs for validation
	existingIDs := make(map[string]bool)
	var existingIDList []string

	// Extract IDs from all rulesets, handling UnparsedObject for problematic rulesets
	for _, ruleset := range existingRulesets {
		rulesetID, ok := extractRulesetID(ruleset)
		if ok {
			existingIDs[rulesetID] = true
			existingIDList = append(existingIDList, rulesetID)
		}
	}

	// Check if we extracted fewer IDs than total rulesets (API client deserialization bug)
	if len(existingIDList) < len(existingRulesets) {
		diag.AddError(
			"Unable to extract IDs from all rulesets",
			fmt.Sprintf("The API returned %d rulesets, but we could only extract IDs from %d of them. "+
				"This is likely due to a deserialization issue in the API client library. "+
				"To work around this issue, please:\n"+
				"1. Go to the Datadog UI and manually delete any problematic rulesets\n"+
				"2. Or set override_ui_defined_resources=false to preserve unmanaged rulesets\n"+
				"3. Then try again\n\n"+
				"Successfully extracted IDs: %v\n"+
				"Total rulesets from API: %d",
				len(existingRulesets), len(existingIDList), existingIDList, len(existingRulesets)),
		)
		return
	}

	// Validate desired rulesets exist
	for _, rulesetID := range desiredOrder {
		if !existingIDs[rulesetID] {
			diag.AddError("Invalid ruleset ID", fmt.Sprintf("ruleset ID %s does not exist", rulesetID))
			return
		}
	}

	// Create a map of desired IDs for checking unmanaged rulesets
	desiredIDsSet := make(map[string]bool)
	for _, id := range desiredOrder {
		desiredIDsSet[id] = true
	}

	// Find and delete unmanaged rulesets
	var unmanagedRulesets []string
	for _, ruleset := range existingRulesets {
		rulesetID, ok := extractRulesetID(ruleset)
		if !ok {
			continue
		}

		if !desiredIDsSet[rulesetID] {
			unmanagedRulesets = append(unmanagedRulesets, rulesetID)
		}
	}

	// Delete unmanaged rulesets
	for _, rulesetID := range unmanagedRulesets {
		httpResp, err := r.Api.DeleteTagPipelinesRuleset(r.Auth, rulesetID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				// Resource already deleted - continue
				continue
			}
			diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error deleting unmanaged ruleset %s: %v", rulesetID, httpResp)))
			return
		}
	}

	// Now reorder the remaining rulesets
	rulesetData := make([]datadogV2.ReorderRulesetResourceData, len(desiredOrder))
	for i, rulesetID := range desiredOrder {
		rulesetData[i] = datadogV2.ReorderRulesetResourceData{
			Id:   &rulesetID,
			Type: datadogV2.REORDERRULESETRESOURCEDATATYPE_RULESET,
		}
	}

	reorderRequest := datadogV2.ReorderRulesetResourceArray{
		Data: rulesetData,
	}

	httpResponse, err = r.Api.ReorderTagPipelinesRulesets(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering tag pipeline rulesets: %v", httpResponse)))
		return
	}
}

func (r *tagPipelineRulesetsResource) updateOrderWithAllRulesets(state *tagPipelineRulesetsModel, diag *diag.Diagnostics, desiredOrder []string) {
	// Get all existing rulesets
	resp, httpResponse, err := r.Api.ListTagPipelinesRulesets(r.Auth)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error listing tag pipeline rulesets: %v", httpResponse)))
		return
	}

	var existingRulesets []datadogV2.RulesetRespData
	if respData, ok := resp.GetDataOk(); ok {
		existingRulesets = *respData
	}

	// Create a map of desired IDs for quick lookup
	desiredIDsSet := make(map[string]bool)
	for _, id := range desiredOrder {
		desiredIDsSet[id] = true
	}

	// Get all rulesets with positions (sorted)
	rulesetPositions := getRulesetsWithPositions(existingRulesets, desiredIDsSet, false)

	// Validate desired rulesets exist
	existingIDs := make(map[string]bool)
	for _, rp := range rulesetPositions {
		existingIDs[rp.ID] = true
	}

	for _, rulesetID := range desiredOrder {
		if !existingIDs[rulesetID] {
			diag.AddError("Invalid ruleset ID", fmt.Sprintf("ruleset ID %s does not exist", rulesetID))
			return
		}
	}

	// Check if there are unmanaged rulesets in the middle of the order
	if len(existingRulesets) > len(desiredOrder) {
		// Find unmanaged rulesets and check positions
		unmanagedInfo := findUnmanagedRulesets(rulesetPositions, desiredIDsSet)

		if len(unmanagedInfo.Rulesets) > 0 {
			if !unmanagedInfo.AllAtEnd {
				diag.AddError(
					"Unmanaged rulesets detected in the middle of order",
					fmt.Sprintf("Found %d rulesets in Datadog that are not managed by this Terraform configuration and are not all at the end of the order: %v. "+
						"Unmanaged rulesets must be at the end of the order to allow reordering. Please either:\n"+
						"1. Add the unmanaged rulesets to your Terraform configuration and import them using 'terraform import datadog_tag_pipeline_ruleset.<name> <ruleset_id>'\n"+
						"2. Delete unmanaged rulesets from Datadog if they're no longer needed\n"+
						"3. Set override_ui_defined_resources=true to automatically delete unmanaged rulesets\n"+
						"4. Manually reorder the rulesets in Datadog UI so all unmanaged rulesets are at the end\n\n"+
						"Current order positions: %v\n"+
						"Unmanaged rulesets: %v at positions: %v\n"+
						"Expected position for unmanaged rulesets to start: %d",
						len(unmanagedInfo.Rulesets), unmanagedInfo.Rulesets,
						rulesetPositions, unmanagedInfo.Rulesets, unmanagedInfo.Positions, len(rulesetPositions)-len(unmanagedInfo.Rulesets)),
				)
				return
			}

			// Warning: Notify user about unmanaged rulesets even though they're at the end
			unmanagedDetails := formatUnmanagedDetails(unmanagedInfo.Rulesets, true)

			diag.AddWarning(
				"Unmanaged rulesets detected",
				fmt.Sprintf("Found %d ruleset(s) in Datadog that are not managed by this Terraform configuration:\n\n"+
					"%s\n\n"+
					"These rulesets are currently at the end of the order, so they won't block this operation. However, "+
					"to ensure complete infrastructure management and prevent configuration drift, consider:\n"+
					"1. Importing them into Terraform: 'terraform import datadog_tag_pipeline_ruleset.<name> <ruleset_id>'\n"+
					"2. Deleting them from Datadog if they're no longer needed\n"+
					"3. Setting override_ui_defined_resources=true to automatically delete unmanaged rulesets\n\n"+
					"Your managed rulesets will be placed first, followed by these unmanaged rulesets.",
					len(unmanagedInfo.Rulesets),
					"  • "+strings.Join(unmanagedDetails, "\n  • ")),
			)
		}
	}

	// Create final order: desired rulesets first, then remaining rulesets in their current order
	finalOrder := make([]string, 0, len(rulesetPositions))

	// Add desired rulesets in the specified order
	finalOrder = append(finalOrder, desiredOrder...)

	// Add remaining unmanaged rulesets in their current order
	for _, rp := range rulesetPositions {
		if !desiredIDsSet[rp.ID] {
			finalOrder = append(finalOrder, rp.ID)
		}
	}

	// Build reorder request with ALL rulesets
	rulesetData := make([]datadogV2.ReorderRulesetResourceData, len(finalOrder))
	for i, rulesetID := range finalOrder {
		rulesetData[i] = datadogV2.ReorderRulesetResourceData{
			Id:   &rulesetID,
			Type: datadogV2.REORDERRULESETRESOURCEDATATYPE_RULESET,
		}
	}

	reorderRequest := datadogV2.ReorderRulesetResourceArray{
		Data: rulesetData,
	}

	httpResponse, err = r.Api.ReorderTagPipelinesRulesets(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering tag pipeline rulesets with all rulesets: %v", httpResponse)))
		return
	}
}
