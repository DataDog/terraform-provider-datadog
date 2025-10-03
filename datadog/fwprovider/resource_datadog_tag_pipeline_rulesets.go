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
	_ resource.ResourceWithConfigure   = &tagPipelineRulesetOrderResource{}
	_ resource.ResourceWithImportState = &tagPipelineRulesetOrderResource{}
)

func NewTagPipelineRulesetOrderResource() resource.Resource {
	return &tagPipelineRulesetOrderResource{}
}

type tagPipelineRulesetOrderModel struct {
	ID         types.String `tfsdk:"id"`
	RulesetIDs types.List   `tfsdk:"ruleset_ids"`
}

type tagPipelineRulesetOrderResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func (r *tagPipelineRulesetOrderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *tagPipelineRulesetOrderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "tag_pipeline_rulesets"
}

func (r *tagPipelineRulesetOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Tag Pipeline Ruleset Order resource that can be used to manage the order of Tag Pipeline Rulesets.",
		Attributes: map[string]schema.Attribute{
			"ruleset_ids": schema.ListAttribute{
				Description: "The list of Tag Pipeline Ruleset IDs, in order. Rulesets are executed in the order specified in this list.",
				ElementType: types.StringType,
				Required:    true,
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *tagPipelineRulesetOrderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state tagPipelineRulesetOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetOrderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state tagPipelineRulesetOrderModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the current list of rulesets to read their order
	resp, httpResponse, err := r.Api.ListRulesets(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reading tag pipeline rulesets. http response: %v", httpResponse)))
		return
	}

	var rulesets []datadogV2.RulesetRespData
	if respData, ok := resp.GetDataOk(); ok {
		rulesets = *respData
	}

	// Create a slice of structs to sort by position
	type rulesetWithPosition struct {
		id       string
		position int32
	}

	rulesetPositions := make([]rulesetWithPosition, 0, len(rulesets))
	for _, ruleset := range rulesets {
		if rulesetID, ok := ruleset.GetIdOk(); ok {
			position := int32(0)
			if rulesetAttrs, ok := ruleset.GetAttributesOk(); ok {
				position = rulesetAttrs.GetPosition()
			}
			rulesetPositions = append(rulesetPositions, rulesetWithPosition{
				id:       *rulesetID,
				position: position,
			})
		}
	}

	// Sort by position
	for i := 0; i < len(rulesetPositions); i++ {
		for j := i + 1; j < len(rulesetPositions); j++ {
			if rulesetPositions[i].position > rulesetPositions[j].position {
				rulesetPositions[i], rulesetPositions[j] = rulesetPositions[j], rulesetPositions[i]
			}
		}
	}

	// Extract ordered IDs
	orderedList := make([]string, 0, len(rulesetPositions))
	for _, rp := range rulesetPositions {
		orderedList = append(orderedList, rp.id)
	}

	state.RulesetIDs, _ = types.ListValueFrom(ctx, types.StringType, orderedList)
	state.ID = types.StringValue("order") // Static ID like other order resources

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetOrderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state tagPipelineRulesetOrderModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.updateOrder(&state, &response.Diagnostics)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetOrderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// No-op: Deleting the order resource doesn't change the actual order of rulesets
	// This follows the same pattern as other order resources in the provider
}

func (r *tagPipelineRulesetOrderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *tagPipelineRulesetOrderResource) updateOrder(state *tagPipelineRulesetOrderModel, diag *diag.Diagnostics) {
	// Set the ID immediately to prevent "unknown value" errors
	state.ID = types.StringValue("order")

	// Convert the Terraform list to strings for the API call
	var desiredRulesetIDs []string
	for _, tfID := range state.RulesetIDs.Elements() {
		rulesetID := tfID.(types.String).ValueString()
		desiredRulesetIDs = append(desiredRulesetIDs, rulesetID)
	}

	// Strict validation: only allow reordering if ALL existing rulesets are managed by Terraform
	// This ensures complete infrastructure control and prevents configuration drift
	r.updateOrderWithAllRulesets(state, diag, desiredRulesetIDs)
}

// Validates that all existing rulesets are managed by Terraform before reordering
func (r *tagPipelineRulesetOrderResource) updateOrderWithAllRulesets(state *tagPipelineRulesetOrderModel, diag *diag.Diagnostics, desiredOrder []string) {
	// Get all existing rulesets
	resp, httpResponse, err := r.Api.ListRulesets(r.Auth)
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
	for _, ruleset := range existingRulesets {
		if rulesetID, ok := ruleset.GetIdOk(); ok {
			existingIDs[*rulesetID] = true
		}
	}

	// Validate desired rulesets exist
	for _, rulesetID := range desiredOrder {
		if !existingIDs[rulesetID] {
			diag.AddError("Invalid ruleset ID", fmt.Sprintf("ruleset ID %s does not exist", rulesetID))
			return
		}
	}

	// Strict validation: Check if there are unmanaged rulesets
	if len(existingRulesets) != len(desiredOrder) {
		// Find unmanaged rulesets
		unmanagedRulesets := make([]string, 0)
		desiredIDsSet := make(map[string]bool)
		for _, id := range desiredOrder {
			desiredIDsSet[id] = true
		}

		for _, ruleset := range existingRulesets {
			if rulesetID, ok := ruleset.GetIdOk(); ok {
				if !desiredIDsSet[*rulesetID] {
					unmanagedRulesets = append(unmanagedRulesets, *rulesetID)
				}
			}
		}

		diag.AddError(
			"Unmanaged rulesets detected",
			fmt.Sprintf("Found %d rulesets in Datadog that are not managed by this Terraform configuration: %v. "+
				"All rulesets must be managed by Terraform. Please either:\n"+
				"1. Import existing rulesets using 'terraform import datadog_tag_pipeline_ruleset.<name> <ruleset_id>'\n"+
				"2. Add the missing rulesets to your Terraform configuration\n"+
				"3. Delete unmanaged rulesets from Datadog if they're no longer needed\n\n"+
				"This ensures complete infrastructure control and prevents configuration drift.",
				len(unmanagedRulesets), unmanagedRulesets),
		)
		return
	}

	finalOrder := desiredOrder

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

	httpResponse, err = r.Api.ReorderRulesets(r.Auth, reorderRequest)
	if err != nil {
		diag.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error reordering tag pipeline rulesets with all rulesets: %v", httpResponse)))
		return
	}
}
