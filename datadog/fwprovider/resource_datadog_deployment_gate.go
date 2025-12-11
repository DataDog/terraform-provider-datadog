package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &deploymentGateResource{}
	_ resource.ResourceWithImportState = &deploymentGateResource{}
)

type deploymentGateResource struct {
	Api  *datadogV2.DeploymentGatesApi
	Auth context.Context
}

type deploymentGateModel struct {
	ID         types.String              `tfsdk:"id"`
	DryRun     types.Bool                `tfsdk:"dry_run"`
	Env        types.String              `tfsdk:"env"`
	Identifier types.String              `tfsdk:"identifier"`
	Service    types.String              `tfsdk:"service"`
	CreatedAt  types.String              `tfsdk:"created_at"`
	UpdatedAt  types.String              `tfsdk:"updated_at"`
	Rules      []deploymentGateRuleModel `tfsdk:"rule"`
}

type deploymentGateRuleModel struct {
	ID      types.String                    `tfsdk:"id"`
	Name    types.String                    `tfsdk:"name"`
	Type    types.String                    `tfsdk:"type"`
	DryRun  types.Bool                      `tfsdk:"dry_run"`
	Options *deploymentGateRuleOptionsModel `tfsdk:"options"`
}

type deploymentGateRuleOptionsModel struct {
	ExcludedResources types.List   `tfsdk:"excluded_resources"`
	Duration          types.Int64  `tfsdk:"duration"`
	Query             types.String `tfsdk:"query"`
}

func NewDeploymentGateResource() resource.Resource {
	return &deploymentGateResource{}
}

func (r *deploymentGateResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetDeploymentGatesApiV2()
	r.Auth = providerData.Auth
}

func (r *deploymentGateResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "deployment_gate"
}

func (r *deploymentGateResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog DeploymentGate resource. This can be used to create and manage Datadog deployment_gate.",
		Attributes: map[string]schema.Attribute{
			"dry_run": schema.BoolAttribute{
				Optional:    true,
				Description: "The `attributes` `dry_run`.",
				Computed:    true,
			},
			"env": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `env`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identifier": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The `attributes` `identifier`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `service`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp of the deployment gate.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp of the deployment gate.",
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				Description: "Deployment rules for this gate.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The rule ID.",
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The rule name. Must be unique within the deployment gate.",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The rule type (e.g., 'faulty_deployment_detection', 'monitor').",
						},
						"dry_run": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the rule is in dry run mode.",
						},
					},
					Blocks: map[string]schema.Block{
						"options": schema.SingleNestedBlock{
							Description: "Options for the deployment rule.",
							Attributes: map[string]schema.Attribute{
								"duration": schema.Int64Attribute{
									Description: "The duration for the rule.",
									Required:    true,
								},
								"query": schema.StringAttribute{
									Optional:    true,
									Description: "The query for monitor rules.",
								},
								"excluded_resources": schema.ListAttribute{
									Optional:    true,
									Description: "Resources to exclude from faulty deployment detection.",
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *deploymentGateResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *deploymentGateResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetDeploymentGate(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DeploymentGate"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddWarning("response contains unparsedObject", err.Error())
	}

	r.updateState(ctx, &state, &resp)

	response.Diagnostics.Append(r.readAndReconcileRules(ctx, id, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *deploymentGateResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.validateRules(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildDeploymentGateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateDeploymentGate(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating DeploymentGate"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddWarning("response contains unparsedObject", err.Error())
	}
	r.updateState(ctx, &state, &resp)

	gateID := state.ID.ValueString()
	response.Diagnostics.Append(r.createRules(ctx, gateID, &state)...)
	if response.Diagnostics.HasError() {
		// Rollback: delete the gate if rule creation failed
		_, err = r.Api.DeleteDeploymentGate(r.Auth, gateID)
		if err != nil {
			response.Diagnostics.AddError("Error rolling back rule creation", err.Error())
		}
		return
	}

	gateResp, _, err := r.Api.GetDeploymentGate(r.Auth, gateID)
	if err == nil {
		r.updateState(ctx, &state, &gateResp)
	}

	response.Diagnostics.Append(r.readAndReconcileRules(ctx, gateID, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *deploymentGateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state deploymentGateModel
	var priorState deploymentGateModel
	request.State.Get(ctx, &priorState)

	var plan deploymentGateModel
	request.Plan.Get(ctx, &plan)

	priorRulesByName := make(map[string]deploymentGateRuleModel)
	for _, rule := range priorState.Rules {
		key := rule.Name.ValueString()
		priorRulesByName[key] = rule
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Preserve computed fields from prior state
	state.ID = priorState.ID
	state.CreatedAt = priorState.CreatedAt
	state.UpdatedAt = priorState.UpdatedAt

	// Match state rules with prior rules by name to preserve IDs
	// Only preserve ID if the type hasn't changed (type changes require recreation)
	for i := range state.Rules {
		key := state.Rules[i].Name.ValueString()
		if priorRule, exists := priorRulesByName[key]; exists {
			// Only preserve ID if type is the same
			// If type changed, treat it as a new rule (old one will be deleted, new one created)
			if state.Rules[i].Type.ValueString() == priorRule.Type.ValueString() {
				state.Rules[i].ID = priorRule.ID
			}
		}
	}

	response.Diagnostics.Append(r.validateRules(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildDeploymentGateUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateDeploymentGate(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating DeploymentGate"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddWarning("response contains unparsedObject", err.Error())
	}
	r.updateState(ctx, &state, &resp)
	gateResp, _, err := r.Api.GetDeploymentGate(r.Auth, id)
	if err == nil {
		r.updateState(ctx, &state, &gateResp)
	}

	response.Diagnostics.Append(r.syncRules(ctx, id, &state)...)
	response.Diagnostics.Append(
		r.readAndReconcileRules(ctx, id, &state)...,
	)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *deploymentGateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	response.Diagnostics.Append(r.deleteAllRules(ctx, id)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteDeploymentGate(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting deployment_gate"))
		return
	}
}

func (r *deploymentGateResource) updateState(ctx context.Context, state *deploymentGateModel, resp *datadogV2.DeploymentGateResponse) {
	data := resp.GetData()
	state.ID = types.StringValue(data.GetId())

	attributes := data.GetAttributes()
	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(createdAt.String())
	}

	if dryRun, ok := attributes.GetDryRunOk(); ok {
		state.DryRun = types.BoolValue(*dryRun)
	}

	if env, ok := attributes.GetEnvOk(); ok {
		state.Env = types.StringValue(*env)
	}

	if identifier, ok := attributes.GetIdentifierOk(); ok {
		state.Identifier = types.StringValue(*identifier)
	}

	if service, ok := attributes.GetServiceOk(); ok {
		state.Service = types.StringValue(*service)
	}

	if updatedAt, ok := attributes.GetUpdatedAtOk(); ok {
		state.UpdatedAt = types.StringValue(updatedAt.String())
	}
}

func (r *deploymentGateResource) buildDeploymentGateRequestBody(ctx context.Context, state *deploymentGateModel) (*datadogV2.CreateDeploymentGateParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateDeploymentGateParamsDataAttributesWithDefaults()

	if !state.DryRun.IsNull() {
		attributes.SetDryRun(state.DryRun.ValueBool())
	}
	if !state.Env.IsNull() {
		attributes.SetEnv(state.Env.ValueString())
	}
	if !state.Identifier.IsNull() {
		attributes.SetIdentifier(state.Identifier.ValueString())
	}
	if !state.Service.IsNull() {
		attributes.SetService(state.Service.ValueString())
	}

	req := datadogV2.NewCreateDeploymentGateParamsWithDefaults()
	req.Data = *datadogV2.NewCreateDeploymentGateParamsDataWithDefaults()
	req.Data.Type = "deployment_gate"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *deploymentGateResource) buildDeploymentGateUpdateRequestBody(ctx context.Context, state *deploymentGateModel) (*datadogV2.UpdateDeploymentGateParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUpdateDeploymentGateParamsDataAttributesWithDefaults()

	if !state.DryRun.IsNull() {
		attributes.SetDryRun(state.DryRun.ValueBool())
	}

	req := datadogV2.NewUpdateDeploymentGateParamsWithDefaults()
	req.Data = *datadogV2.NewUpdateDeploymentGateParamsDataWithDefaults()
	req.Data.Type = "deployment_gate"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

// validateRules validates that rule types match their options and that rule names are unique
func (r *deploymentGateResource) validateRules(ctx context.Context, state *deploymentGateModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check for duplicate rule names
	nameCount := make(map[string]int)
	nameIndices := make(map[string][]int)
	for i, rule := range state.Rules {
		name := rule.Name.ValueString()
		nameCount[name]++
		nameIndices[name] = append(nameIndices[name], i)
	}

	for name, count := range nameCount {
		if count > 1 {
			indices := nameIndices[name]
			diags.AddError(
				"Duplicate rule name",
				fmt.Sprintf("Rule name '%s' is used %d times (at indices %v). "+
					"Rule names must be unique within a deployment gate since they are used to track rule identity across updates. "+
					"Please ensure each rule has a unique name.", name, count, indices),
			)
		}
	}

	for i, rule := range state.Rules {
		if isEmptyOption(rule.Options) {
			continue
		}

		ruleType := rule.Type.ValueString()

		// Check for faulty_deployment_detection specific options
		hasFddOptions := !rule.Options.ExcludedResources.IsNull()

		// Check for monitor specific options
		hasMonitorOptions := !rule.Options.Query.IsNull()

		if ruleType == "faulty_deployment_detection" && hasMonitorOptions {
			diags.AddError(
				"Invalid options for deployment rule type",
				fmt.Sprintf("Rule %d: type 'faulty_deployment_detection' cannot use monitor options (query). "+
					"Use faulty deployment detection options instead: duration, excluded_resources.", i),
			)
		} else if ruleType == "monitor" && hasFddOptions {
			diags.AddError(
				"Invalid options for deployment rule type",
				fmt.Sprintf("Rule %d: type 'monitor' cannot use faulty deployment detection options (excluded_resources). "+
					"Use monitor options instead: duration, query.", i),
			)
		}
	}

	return diags
}

func (r *deploymentGateResource) createRules(ctx context.Context, gateID string, state *deploymentGateModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for i := range state.Rules {
		rule := &state.Rules[i]

		body, ruleDiags := r.buildRuleRequestBody(ctx, rule)
		diags.Append(ruleDiags...)
		if diags.HasError() {
			return diags
		}

		resp, _, err := r.Api.CreateDeploymentRule(r.Auth, gateID, *body)
		if err != nil {
			diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error creating deployment rule '%s'", rule.Name.ValueString())))
			return diags
		}

		if err := utils.CheckForUnparsed(resp); err != nil {
			diags.AddWarning("response contains unparsedObject", err.Error())
		}

		r.updateRuleState(ctx, rule, &resp)
	}

	return diags
}

// Reads all rules from a gate and removes rules not managed from terraform
func (r *deploymentGateResource) readAndReconcileRules(ctx context.Context, gateID string, state *deploymentGateModel) diag.Diagnostics {
	var diags diag.Diagnostics

	rulesResp, _, err := r.Api.GetDeploymentGateRules(r.Auth, gateID)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing deployment rules"))
		return diags
	}

	if err := utils.CheckForUnparsed(rulesResp); err != nil {
		diags.AddWarning("response contains unparsedObject", err.Error())
	}

	data := rulesResp.GetData()
	attributes := data.GetAttributes()
	apiRules, ok := attributes.GetRulesOk()
	if !ok || apiRules == nil {
		state.Rules = []deploymentGateRuleModel{}
		return diags
	}

	// Build index by ID and by name for fast lookup
	byID := make(map[string]*datadogV2.DeploymentRuleResponseDataAttributes)
	byName := make(map[string]*datadogV2.DeploymentRuleResponseDataAttributes)
	for i := range *apiRules {
		apiRule := &(*apiRules)[i]
		byNameVal := ""
		if name, ok := apiRule.GetNameOk(); ok && name != nil {
			byNameVal = *name
		}
		byName[byNameVal] = apiRule
		if idVal, ok := apiRule.AdditionalProperties["id"]; ok {
			if idStr, ok := idVal.(string); ok {
				byID[idStr] = apiRule
			}
		}
	}

	newRules := make([]deploymentGateRuleModel, 0, len(state.Rules))
	for _, existing := range state.Rules {
		var matched *datadogV2.DeploymentRuleResponseDataAttributes

		// First try strict ID match if we have an ID
		if !existing.ID.IsNull() && !existing.ID.IsUnknown() {
			matched = byID[existing.ID.ValueString()]
		}

		// Only fallback to name if ID is null/unknown AND a name match exists
		if matched == nil && (existing.ID.IsNull() || existing.ID.IsUnknown()) {
			matched = byName[existing.Name.ValueString()]
		}
		if matched == nil {
			// Rule disappeared in the API → skip it
			continue
		}

		updated := existing
		r.updateRuleStateFromAttributes(ctx, &updated, matched)
		newRules = append(newRules, updated)
	}

	state.Rules = newRules
	return diags
}

// syncRules synchronizes rules during update: creates new, updates existing, deletes removed.
// It uses a two-phase approach:
//  1. Create/update all desired rules and record the IDs that should exist.
//  2. Delete any API rules whose IDs are not in that recorded set (including unmanaged rules).
//
// It also updates state.Rules to contain only the rules that were successfully synced.
func (r *deploymentGateResource) syncRules(ctx context.Context, gateID string, state *deploymentGateModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Phase 0: Get all existing rules from API
	rulesResp, _, err := r.Api.GetDeploymentGateRules(r.Auth, gateID)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing deployment rules"))
		return diags
	}

	if err := utils.CheckForUnparsed(rulesResp); err != nil {
		diags.AddWarning("response contains unparsedObject", err.Error())
	}

	data := rulesResp.GetData()
	attributes := data.GetAttributes()
	apiRules, _ := attributes.GetRulesOk()

	// Build a map of existing rule IDs from the API so we can tell which IDs are still valid.
	existingRuleIDs := make(map[string]bool)
	if apiRules != nil {
		for _, apiRule := range *apiRules {
			if idVal, ok := apiRule.AdditionalProperties["id"]; ok {
				if idStr, ok := idVal.(string); ok {
					existingRuleIDs[idStr] = true
				}
			}
		}
	}

	// Phase 1: Create or update desired rules.
	// Track the IDs that should exist after this sync so we can delete everything else in Phase 2.
	syncedRules := make([]deploymentGateRuleModel, 0, len(state.Rules))
	persistedRuleIDs := make(map[string]bool)

	for i := range state.Rules {
		rule := &state.Rules[i]

		needsCreate := rule.ID.IsNull() || rule.ID.IsUnknown()
		if !needsCreate {
			// If the ruleId is not in existingRuleIDs, it means it was deleted in the backend and need to be recreated
			if _, stillExists := existingRuleIDs[rule.ID.ValueString()]; !stillExists {
				needsCreate = true
			}
		}

		if needsCreate {
			// Create new rule
			body, ruleDiags := r.buildRuleRequestBody(ctx, rule)
			diags.Append(ruleDiags...)
			if diags.HasError() {
				return diags
			}

			resp, _, err := r.Api.CreateDeploymentRule(r.Auth, gateID, *body)
			if err != nil {
				diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error creating deployment rule '%s'", rule.Name.ValueString())))
				return diags
			}

			if err := utils.CheckForUnparsed(resp); err != nil {
				diags.AddWarning("response contains unparsedObject", err.Error())
			}

			r.updateRuleState(ctx, rule, &resp)
		} else {
			// Update existing rule
			ruleID := rule.ID.ValueString()
			body, ruleDiags := r.buildRuleUpdateRequestBody(ctx, rule)
			diags.Append(ruleDiags...)
			if diags.HasError() {
				return diags
			}

			resp, _, err := r.Api.UpdateDeploymentRule(r.Auth, gateID, ruleID, *body)
			if err != nil {
				diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error updating deployment rule '%s'", rule.Name.ValueString())))
				return diags
			}

			if err := utils.CheckForUnparsed(resp); err != nil {
				diags.AddWarning("response contains unparsedObject", err.Error())
			}

			r.updateRuleState(ctx, rule, &resp)
		}

		// At this point the rule has a valid ID and reflects the latest API response.
		if !rule.ID.IsNull() && !rule.ID.IsUnknown() {
			persistedRuleIDs[rule.ID.ValueString()] = true
		}

		syncedRules = append(syncedRules, *rule)
	}

	// Phase 2: Delete any existing API rules that are not in persistedRuleIDs.
	// This includes unmanaged rules and rules removed from configuration.
	if apiRules != nil {
		for _, apiRule := range *apiRules {
			var ruleID string
			if idVal, ok := apiRule.AdditionalProperties["id"]; ok {
				if idStr, ok := idVal.(string); ok {
					ruleID = idStr
				} else {
					continue
				}
			} else {
				continue
			}

			if persistedRuleIDs[ruleID] {
				// This rule was created or updated in Phase 1 and should be kept.
				continue
			}

			httpResp, err := r.Api.DeleteDeploymentRule(r.Auth, gateID, ruleID)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					// Already deleted out-of-band; ignore.
					continue
				}
				diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error deleting rule %s", ruleID)))
				return diags
			}
		}
	}

	state.Rules = syncedRules
	return diags
}

// deleteAllRules deletes all rules for a gate
func (r *deploymentGateResource) deleteAllRules(_ context.Context, gateID string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get all rules for this gate
	rulesResp, _, err := r.Api.GetDeploymentGateRules(r.Auth, gateID)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing deployment rules"))
		return diags
	}

	if err := utils.CheckForUnparsed(rulesResp); err != nil {
		diags.AddWarning("response contains unparsedObject", err.Error())
	}

	data := rulesResp.GetData()
	attributes := data.GetAttributes()
	apiRules, ok := attributes.GetRulesOk()
	if !ok || apiRules == nil {
		return diags
	}

	// Delete each rule
	for _, apiRule := range *apiRules {
		// Extract rule ID from AdditionalProperties since it's not in the spec
		var ruleID string
		if idVal, ok := apiRule.AdditionalProperties["id"]; ok {
			if idStr, ok := idVal.(string); ok {
				ruleID = idStr
			} else {
				continue
			}
		} else {
			continue
		}

		httpResp, err := r.Api.DeleteDeploymentRule(r.Auth, gateID, ruleID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				// Already deleted
				continue
			}
			diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error deleting rule %s", ruleID)))
			return diags
		}
	}

	return diags
}

// buildRuleRequestBody builds the request body for creating a rule
func (r *deploymentGateResource) buildRuleRequestBody(ctx context.Context, rule *deploymentGateRuleModel) (*datadogV2.CreateDeploymentRuleParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCreateDeploymentRuleParamsDataAttributesWithDefaults()

	if !rule.DryRun.IsNull() {
		attributes.SetDryRun(rule.DryRun.ValueBool())
	}
	if !rule.Name.IsNull() {
		attributes.SetName(rule.Name.ValueString())
	}
	if !rule.Type.IsNull() {
		attributes.SetType(rule.Type.ValueString())
	}

	options := datadogV2.DeploymentRulesOptions{}

	if rule.Type.ValueString() == "faulty_deployment_detection" {
		fdd := datadogV2.DeploymentRuleOptionsFaultyDeploymentDetection{}
		if !rule.Options.Duration.IsNull() {
			fdd.Duration = rule.Options.Duration.ValueInt64Pointer()
		}
		if !rule.Options.ExcludedResources.IsNull() {
			var excluded []string
			diags.Append(rule.Options.ExcludedResources.ElementsAs(ctx, &excluded, false)...)
			if !diags.HasError() {
				fdd.ExcludedResources = excluded
			}
		}
		options.DeploymentRuleOptionsFaultyDeploymentDetection = &fdd
		options.DeploymentRuleOptionsMonitor = nil
	} else if rule.Type.ValueString() == "monitor" {
		mon := datadogV2.DeploymentRuleOptionsMonitor{}
		if !rule.Options.Duration.IsNull() {
			mon.Duration = rule.Options.Duration.ValueInt64Pointer()
		}
		if !rule.Options.Query.IsNull() {
			mon.Query = rule.Options.Query.ValueString()
		}
		options.DeploymentRuleOptionsMonitor = &mon
		options.DeploymentRuleOptionsFaultyDeploymentDetection = nil
	}

	attributes.SetOptions(options)

	req := datadogV2.NewCreateDeploymentRuleParamsWithDefaults()
	req.Data = datadogV2.NewCreateDeploymentRuleParamsDataWithDefaults()
	req.Data.Type = "deployment_rule"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

// buildRuleUpdateRequestBody builds the request body for updating a rule
func (r *deploymentGateResource) buildRuleUpdateRequestBody(ctx context.Context, rule *deploymentGateRuleModel) (*datadogV2.UpdateDeploymentRuleParams, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUpdateDeploymentRuleParamsDataAttributesWithDefaults()

	if !rule.DryRun.IsNull() {
		attributes.SetDryRun(rule.DryRun.ValueBool())
	}
	if !rule.Name.IsNull() {
		attributes.SetName(rule.Name.ValueString())
	}

	options := datadogV2.DeploymentRulesOptions{}

	if rule.Type.ValueString() == "faulty_deployment_detection" {
		fdd := datadogV2.DeploymentRuleOptionsFaultyDeploymentDetection{}
		if !rule.Options.Duration.IsNull() {
			fdd.Duration = rule.Options.Duration.ValueInt64Pointer()
		}
		if !rule.Options.ExcludedResources.IsNull() {
			var excluded []string
			diags.Append(rule.Options.ExcludedResources.ElementsAs(ctx, &excluded, false)...)
			if !diags.HasError() {
				fdd.ExcludedResources = excluded
			}
		}
		options.DeploymentRuleOptionsFaultyDeploymentDetection = &fdd
		options.DeploymentRuleOptionsMonitor = nil
	} else if rule.Type.ValueString() == "monitor" {
		mon := datadogV2.DeploymentRuleOptionsMonitor{}
		if !rule.Options.Duration.IsNull() {
			mon.Duration = rule.Options.Duration.ValueInt64Pointer()
		}
		if !rule.Options.Query.IsNull() {
			mon.Query = rule.Options.Query.ValueString()
		}
		options.DeploymentRuleOptionsMonitor = &mon
		options.DeploymentRuleOptionsFaultyDeploymentDetection = nil
	}

	attributes.SetOptions(options)

	req := datadogV2.NewUpdateDeploymentRuleParamsWithDefaults()
	req.Data = *datadogV2.NewUpdateDeploymentRuleParamsDataWithDefaults()
	req.Data.Type = "deployment_rule"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *deploymentGateResource) updateRuleStateFromAttributes(_ context.Context, rule *deploymentGateRuleModel, attributes *datadogV2.DeploymentRuleResponseDataAttributes) {
	if idVal, ok := attributes.AdditionalProperties["id"]; ok {
		if idStr, ok := idVal.(string); ok {
			rule.ID = types.StringValue(idStr)
		}
	}

	if dryRun, ok := attributes.GetDryRunOk(); ok {
		rule.DryRun = types.BoolValue(*dryRun)
	}

	if name, ok := attributes.GetNameOk(); ok {
		rule.Name = types.StringValue(*name)
	}

	if typeVar, ok := attributes.GetTypeOk(); ok {
		rule.Type = types.StringValue(string(*typeVar))
	}

	if options, ok := attributes.GetOptionsOk(); ok {
		rule.Options = &deploymentGateRuleOptionsModel{
			ExcludedResources: types.ListNull(types.StringType),
			Duration:          types.Int64Null(),
			Query:             types.StringNull(),
		}

		// Handle options based on rule type
		if rule.Type.ValueString() == "faulty_deployment_detection" {
			if fddOptions := options.DeploymentRuleOptionsFaultyDeploymentDetection; fddOptions != nil {
				if duration, ok := fddOptions.GetDurationOk(); ok {
					rule.Options.Duration = types.Int64PointerValue(duration)
				}
				if excludedResources, ok := fddOptions.GetExcludedResourcesOk(); ok {
					if excludedResources != nil && len(*excludedResources) > 0 {
						elements := make([]attr.Value, len(*excludedResources))
						for i, resourceName := range *excludedResources {
							elements[i] = types.StringValue(resourceName)
						}
						rule.Options.ExcludedResources, _ = types.ListValue(types.StringType, elements)
					}
				}
			}
		} else if rule.Type.ValueString() == "monitor" {
			if monitorOptions := options.DeploymentRuleOptionsMonitor; monitorOptions != nil {
				if duration, ok := monitorOptions.GetDurationOk(); ok {
					rule.Options.Duration = types.Int64PointerValue(duration)
				}
				if query, ok := monitorOptions.GetQueryOk(); ok {
					rule.Options.Query = types.StringValue(*query)
				}
			}
		}

		// If the union type couldn't be disambiguated, the API client stores it in UnparsedObject
		// This happens when both variants of the union can successfully unmarshal the JSON
		// (e.g., both have optional "duration" field, so JSON with both "duration" and "query" matches both)
		// This is a bug in the client, I'll work on fixing it for the next release
		if options.DeploymentRuleOptionsFaultyDeploymentDetection == nil &&
			options.DeploymentRuleOptionsMonitor == nil &&
			options.UnparsedObject != nil {
			// Manually parse the UnparsedObject based on the rule type
			unparsedMap, ok := options.UnparsedObject.(map[string]interface{})
			if ok {
				if rule.Type.ValueString() == "faulty_deployment_detection" {
					if duration, ok := unparsedMap["duration"].(float64); ok {
						rule.Options.Duration = types.Int64Value(int64(duration))
					}
					if excludedResourcesRaw, ok := unparsedMap["excluded_resources"].([]interface{}); ok {
						elements := make([]attr.Value, len(excludedResourcesRaw))
						for i, resource := range excludedResourcesRaw {
							if resourceStr, ok := resource.(string); ok {
								elements[i] = types.StringValue(resourceStr)
							}
						}
						rule.Options.ExcludedResources, _ = types.ListValue(types.StringType, elements)
					}
				} else if rule.Type.ValueString() == "monitor" {
					if duration, ok := unparsedMap["duration"].(float64); ok {
						rule.Options.Duration = types.Int64Value(int64(duration))
					}
					if query, ok := unparsedMap["query"].(string); ok {
						rule.Options.Query = types.StringValue(query)
					}
				}
			}
		}
	}
}

// updateRuleState updates the rule state from a full API response
func (r *deploymentGateResource) updateRuleState(ctx context.Context, rule *deploymentGateRuleModel, resp *datadogV2.DeploymentRuleResponse) {
	data := resp.GetData()
	rule.ID = types.StringValue(data.GetId())
	attributes := data.GetAttributes()
	r.updateRuleStateFromAttributes(ctx, rule, &attributes)
}

func isEmptyOption(options *deploymentGateRuleOptionsModel) bool {
	if options == nil {
		return true
	}

	return options.Query.IsNull() && options.ExcludedResources.IsNull() && options.Duration.IsNull()
}
