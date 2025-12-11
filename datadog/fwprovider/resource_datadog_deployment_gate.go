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
							Description: "The rule name.",
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
									Optional:    true,
									Description: "The duration for the rule.",
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

	// Read all rules and reconcile with state
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

	// Validate rules before creating
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

	// Create all rules
	gateID := state.ID.ValueString()
	response.Diagnostics.Append(r.createRules(ctx, gateID, &state)...)
	if response.Diagnostics.HasError() {
		// Rollback: delete the gate if rule creation failed
		_, _ = r.Api.DeleteDeploymentGate(r.Auth, gateID)
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *deploymentGateResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate rules before updating
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

	// Sync rules (create, update, delete as needed)
	response.Diagnostics.Append(r.syncRules(ctx, id, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *deploymentGateResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state deploymentGateModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	// Delete all rules first
	response.Diagnostics.Append(r.deleteAllRules(ctx, id)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Then delete the gate
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

// validateRules validates that rule types match their options
func (r *deploymentGateResource) validateRules(ctx context.Context, state *deploymentGateModel) diag.Diagnostics {
	var diags diag.Diagnostics

	for i, rule := range state.Rules {
		if rule.Options == nil {
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

// createRules creates all rules for the gate
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

		// Update rule with returned ID
		r.updateRuleState(ctx, rule, &resp)
	}

	return diags
}

// readAndReconcileRules reads all rules from API and reconciles with desired state
// This deletes any unmanaged rules
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

	managedRuleIDs := make(map[string]struct{})
	for _, rule := range state.Rules {
		if !rule.ID.IsNull() && !rule.ID.IsUnknown() {
			managedRuleIDs[rule.ID.ValueString()] = struct{}{}
		}
	}

	// Get rules from API response
	data := rulesResp.GetData()
	attributes := data.GetAttributes()
	apiRules, ok := attributes.GetRulesOk()
	if !ok || apiRules == nil {
		return diags
	}

	// Build a map of API rules by ID for easy lookup
	apiRulesByID := make(map[string]*datadogV2.DeploymentRuleResponseDataAttributes)
	for i := range *apiRules {
		apiRule := &(*apiRules)[i]
		// Extract rule ID from AdditionalProperties since it's not in the spec
		if idVal, ok := apiRule.AdditionalProperties["id"]; ok {
			if idStr, ok := idVal.(string); ok {
				apiRulesByID[idStr] = apiRule
			}
		}
	}

	// Delete any unmanaged rules
	for ruleID := range apiRulesByID {
		if _, present := managedRuleIDs[ruleID]; !present {
			// This rule is not managed by Terraform, delete it
			httpResp, err := r.Api.DeleteDeploymentRule(r.Auth, gateID, ruleID)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					// Already deleted, continue
					continue
				}
				diags.AddWarning(
					"Failed to delete unmanaged rule",
					fmt.Sprintf("Could not delete unmanaged rule %s: %v", ruleID, err),
				)
			}
		}
	}

	// Update state with current rule details from API response
	// Note: We use the data we already have from GetDeploymentGateRules
	for i := range state.Rules {
		rule := &state.Rules[i]
		if rule.ID.IsNull() || rule.ID.IsUnknown() {
			continue
		}

		ruleID := rule.ID.ValueString()

		// Check if the rule still exists in the API
		if apiRule, exists := apiRulesByID[ruleID]; exists {
			// Update state with the API rule data
			r.updateRuleStateFromAttributes(ctx, rule, apiRule)
		} else {
			// Rule was deleted outside Terraform - this will cause drift
			// Terraform will detect this and prompt for recreation on next apply
			diags.AddWarning(
				"Managed rule not found",
				fmt.Sprintf("Rule %s (name: %s) was deleted outside of Terraform. "+
					"Run terraform apply to recreate it.", ruleID, rule.Name.ValueString()),
			)
		}
	}

	return diags
}

// syncRules synchronizes rules during update: creates new, updates existing, deletes removed
func (r *deploymentGateResource) syncRules(ctx context.Context, gateID string, state *deploymentGateModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get all existing rules from API
	rulesResp, _, err := r.Api.GetDeploymentGateRules(r.Auth, gateID)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing deployment rules"))
		return diags
	}

	if err := utils.CheckForUnparsed(rulesResp); err != nil {
		diags.AddWarning("response contains unparsedObject", err.Error())
	}

	// Build maps for comparison
	data := rulesResp.GetData()
	attributes := data.GetAttributes()
	apiRules, _ := attributes.GetRulesOk()

	existingRuleIDs := make(map[string]bool)
	if apiRules != nil {
		for _, apiRule := range *apiRules {
			// Extract rule ID from AdditionalProperties since it's not in the spec
			if idVal, ok := apiRule.AdditionalProperties["id"]; ok {
				if idStr, ok := idVal.(string); ok {
					existingRuleIDs[idStr] = true
				}
			}
		}
	}

	desiredRuleIDs := make(map[string]bool)
	for _, rule := range state.Rules {
		if !rule.ID.IsNull() && !rule.ID.IsUnknown() {
			desiredRuleIDs[rule.ID.ValueString()] = true
		}
	}

	// Delete rules not in desired state (including unmanaged rules)
	if apiRules != nil {
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

			if !desiredRuleIDs[ruleID] {
				httpResp, err := r.Api.DeleteDeploymentRule(r.Auth, gateID, ruleID)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						continue
					}
					diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error deleting rule %s", ruleID)))
					return diags
				}
			}
		}
	}

	// Create or update desired rules
	for i := range state.Rules {
		rule := &state.Rules[i]

		if rule.ID.IsNull() || rule.ID.IsUnknown() || !existingRuleIDs[rule.ID.ValueString()] {
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
	}

	return diags
}

// deleteAllRules deletes all rules for a gate
func (r *deploymentGateResource) deleteAllRules(ctx context.Context, gateID string) diag.Diagnostics {
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

	// Handle options based on rule type
	if rule.Options != nil {
		options := datadogV2.DeploymentRulesOptions{}

		if rule.Type.ValueString() == "faulty_deployment_detection" {
			fddOptions := rule.Options
			options.DeploymentRuleOptionsFaultyDeploymentDetection = &datadogV2.DeploymentRuleOptionsFaultyDeploymentDetection{}

			if !fddOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsFaultyDeploymentDetection.Duration = fddOptions.Duration.ValueInt64Pointer()
			}
			if !fddOptions.ExcludedResources.IsNull() {
				var excludedResources []string
				diags.Append(fddOptions.ExcludedResources.ElementsAs(ctx, &excludedResources, false)...)
				if !diags.HasError() {
					options.DeploymentRuleOptionsFaultyDeploymentDetection.ExcludedResources = excludedResources
				}
			}
		} else if rule.Type.ValueString() == "monitor" {
			monitorOptions := rule.Options
			options.DeploymentRuleOptionsMonitor = &datadogV2.DeploymentRuleOptionsMonitor{}

			if !monitorOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsMonitor.Duration = monitorOptions.Duration.ValueInt64Pointer()
			}
			if !monitorOptions.Query.IsNull() {
				options.DeploymentRuleOptionsMonitor.Query = monitorOptions.Query.ValueString()
			}
		}

		attributes.SetOptions(options)
	}

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

	// Handle options based on rule type
	if rule.Options != nil {
		options := datadogV2.DeploymentRulesOptions{}

		if rule.Type.ValueString() == "faulty_deployment_detection" {
			fddOptions := rule.Options
			options.DeploymentRuleOptionsFaultyDeploymentDetection = &datadogV2.DeploymentRuleOptionsFaultyDeploymentDetection{}

			if !fddOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsFaultyDeploymentDetection.Duration = fddOptions.Duration.ValueInt64Pointer()
			}
			if !fddOptions.ExcludedResources.IsNull() {
				var excludedResources []string
				diags.Append(fddOptions.ExcludedResources.ElementsAs(ctx, &excludedResources, false)...)
				if !diags.HasError() {
					options.DeploymentRuleOptionsFaultyDeploymentDetection.ExcludedResources = excludedResources
				}
			}
		} else if rule.Type.ValueString() == "monitor" {
			monitorOptions := rule.Options
			options.DeploymentRuleOptionsMonitor = &datadogV2.DeploymentRuleOptionsMonitor{}

			if !monitorOptions.Duration.IsNull() {
				options.DeploymentRuleOptionsMonitor.Duration = monitorOptions.Duration.ValueInt64Pointer()
			}
			if !monitorOptions.Query.IsNull() {
				options.DeploymentRuleOptionsMonitor.Query = monitorOptions.Query.ValueString()
			}
		}

		attributes.SetOptions(options)
	}

	req := datadogV2.NewUpdateDeploymentRuleParamsWithDefaults()
	req.Data = *datadogV2.NewUpdateDeploymentRuleParamsDataWithDefaults()
	req.Data.Type = "deployment_rule"
	req.Data.SetAttributes(*attributes)

	return req, diags
}

// updateRuleStateFromAttributes updates the rule state from rule attributes
func (r *deploymentGateResource) updateRuleStateFromAttributes(ctx context.Context, rule *deploymentGateRuleModel, attributes *datadogV2.DeploymentRuleResponseDataAttributes) {
	// Extract rule ID from AdditionalProperties since it's not in the spec
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

	// Handle options from the response
	if options, ok := attributes.GetOptionsOk(); ok {
		if rule.Options == nil {
			rule.Options = &deploymentGateRuleOptionsModel{
				ExcludedResources: types.ListNull(types.StringType),
				Duration:          types.Int64Null(),
				Query:             types.StringNull(),
			}
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
						for i, resource := range *excludedResources {
							elements[i] = types.StringValue(resource)
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
	}
}

// updateRuleState updates the rule state from a full API response
func (r *deploymentGateResource) updateRuleState(ctx context.Context, rule *deploymentGateRuleModel, resp *datadogV2.DeploymentRuleResponse) {
	data := resp.GetData()
	rule.ID = types.StringValue(data.GetId())
	attributes := data.GetAttributes()
	r.updateRuleStateFromAttributes(ctx, rule, &attributes)
}
