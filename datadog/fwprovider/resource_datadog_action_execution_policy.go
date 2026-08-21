package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &actionExecutionPolicyResource{}
	_ resource.ResourceWithImportState    = &actionExecutionPolicyResource{}
	_ resource.ResourceWithValidateConfig = &actionExecutionPolicyResource{}
)

// executionPolicyScopeIntegrations maps each scope variant to the
// `action_pattern.integration` it is only valid for.
var executionPolicyScopeIntegrations = map[string]string{
	"kubernetes":           string(datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_KUBERNETES),
	"scripts":              string(datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_SCRIPT),
	"remote_action_rshell": string(datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_REMOTE_ACTION),
}

type actionExecutionPolicyResource struct {
	Api  *datadogV2.ExecutionPolicyApi
	Auth context.Context
}

type actionExecutionPolicyModel struct {
	ID            types.String                       `tfsdk:"id"`
	Name          types.String                       `tfsdk:"name"`
	Effect        types.String                       `tfsdk:"effect"`
	Version       types.Int32                        `tfsdk:"version"`
	CreatedAt     types.String                       `tfsdk:"created_at"`
	CreatedBy     types.String                       `tfsdk:"created_by"`
	UpdatedAt     types.String                       `tfsdk:"updated_at"`
	UpdatedBy     types.String                       `tfsdk:"updated_by"`
	ActionPattern *executionPolicyActionPatternModel `tfsdk:"action_pattern"`
	Scope         *executionPolicyScopeModel         `tfsdk:"scope"`
	Target        []*executionPolicyTargetModel      `tfsdk:"target"`
}

type executionPolicyActionPatternModel struct {
	Integration types.String `tfsdk:"integration"`
	ActionFqns  types.List   `tfsdk:"action_fqns"`
}

type executionPolicyScopeModel struct {
	Kubernetes         *executionPolicyKubernetesScopeModel         `tfsdk:"kubernetes"`
	Scripts            *executionPolicyScriptScopeModel             `tfsdk:"scripts"`
	RemoteActionRshell *executionPolicyRemoteActionRshellScopeModel `tfsdk:"remote_action_rshell"`
}

type executionPolicyKubernetesScopeModel struct {
	Rule []*executionPolicyKubernetesScopeRuleModel `tfsdk:"rule"`
}

type executionPolicyKubernetesScopeRuleModel struct {
	TargetNamespaces types.List `tfsdk:"target_namespaces"`
}

type executionPolicyScriptScopeModel struct {
	Rule []*executionPolicyScriptScopeRuleModel `tfsdk:"rule"`
}

type executionPolicyScriptScopeRuleModel struct {
	TargetScriptNames types.List `tfsdk:"target_script_names"`
}

type executionPolicyRemoteActionRshellScopeModel struct {
	Rule []*executionPolicyRemoteActionRshellScopeRuleModel `tfsdk:"rule"`
}

type executionPolicyRemoteActionRshellScopeRuleModel struct {
	TargetPaths types.List   `tfsdk:"target_paths"`
	Access      types.String `tfsdk:"access"`
}

type executionPolicyTargetModel struct {
	Name      types.String `tfsdk:"name"`
	AgentTags types.List   `tfsdk:"agent_tags"`
}

func NewActionExecutionPolicyResource() resource.Resource {
	return &actionExecutionPolicyResource{}
}

func (r *actionExecutionPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetExecutionPolicyApiV2()
	r.Auth = providerData.Auth
}

func (r *actionExecutionPolicyResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "action_execution_policy"
}

// ValidateConfig covers the cross-field rules the schema cannot express
func (r *actionExecutionPolicyResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config actionExecutionPolicyModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	if config.ActionPattern == nil {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("action_pattern"),
			"Missing action_pattern block",
			"You must specify an action_pattern block.",
		)
		return
	}

	if config.Scope == nil {
		return
	}

	var set []string
	if config.Scope.Kubernetes != nil {
		set = append(set, "kubernetes")
	}
	if config.Scope.Scripts != nil {
		set = append(set, "scripts")
	}
	if config.Scope.RemoteActionRshell != nil {
		set = append(set, "remote_action_rshell")
	}

	if len(set) > 1 {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("scope"),
			"Conflicting scope blocks",
			fmt.Sprintf(
				"At most one of `kubernetes`, `scripts` or `remote_action_rshell` may be set, but %s were.",
				strings.Join(set, ", "),
			),
		)
		return
	}

	// An empty scope block is valid: it means the policy has no scope restriction.
	if len(set) == 0 {
		return
	}

	var ruleCount int
	switch set[0] {
	case "kubernetes":
		ruleCount = len(config.Scope.Kubernetes.Rule)
	case "scripts":
		ruleCount = len(config.Scope.Scripts.Rule)
	case "remote_action_rshell":
		ruleCount = len(config.Scope.RemoteActionRshell.Rule)
	}
	if ruleCount == 0 {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("scope").AtName(set[0]).AtName("rule"),
			"Missing scope rule block",
			fmt.Sprintf("A configured `%s` scope must contain at least one `rule` block.", set[0]),
		)
		return
	}

	integration := config.ActionPattern.Integration
	if integration.IsNull() || integration.IsUnknown() {
		return
	}

	if want := executionPolicyScopeIntegrations[set[0]]; integration.ValueString() != want {
		response.Diagnostics.AddAttributeError(
			frameworkPath.Root("scope").AtName(set[0]),
			"Scope does not match integration",
			fmt.Sprintf(
				"A `%s` scope requires `action_pattern.integration` to be `%s`, but it is `%s`.",
				set[0], want, integration.ValueString(),
			),
		)
	}
}

func (r *actionExecutionPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Execution Policy resource. Execution policies control which Action Platform actions may run against your infrastructure, and where. Each policy pairs an effect (`allow` or `deny`) with a pattern of actions, optionally narrowed to specific Kubernetes namespaces, scripts or remote shell paths, and optionally scoped to agents matching a set of Fleet Automation tags.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the execution policy.",
			},
			"effect": schema.StringAttribute{
				Required:    true,
				Description: "Whether the policy allows or denies the matched actions.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(datadogV2.EXECUTIONPOLICYEFFECT_ALLOW),
						string(datadogV2.EXECUTIONPOLICYEFFECT_DENY),
					),
				},
			},
			"version": schema.Int32Attribute{
				Computed:    true,
				Description: "The version of the execution policy. Incremented by Datadog on every update.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the execution policy was created, as an RFC3339 timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the user who created the execution policy.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the execution policy was last updated, as an RFC3339 timestamp.",
			},
			"updated_by": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the user who last updated the execution policy.",
			},
		},
		Blocks: map[string]schema.Block{
			"action_pattern": schema.SingleNestedBlock{
				Description: "The set of actions this policy applies to. Required.",
				Attributes: map[string]schema.Attribute{
					"integration": schema.StringAttribute{
						Required:    true,
						Description: "The integration the actions belong to.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								string(datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_KUBERNETES),
								string(datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_SCRIPT),
								string(datadogV2.EXECUTIONPOLICYINTEGRATION_INTEGRATION_REMOTE_ACTION),
							),
						},
					},
					"action_fqns": schema.ListAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "The fully qualified action names this policy matches. Use `*` to match all actions of the integration, or a fully qualified name prefixed with the integration's action namespace (for example `com.datadoghq.script.*` for the Script integration).",
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
				},
			},
			"scope": schema.SingleNestedBlock{
				Description: "Restricts where the policy applies, beyond `action_pattern`. At most one of `kubernetes`, `scripts` or `remote_action_rshell` may be set, and it must match `action_pattern.integration`. Omitting this block means the policy has no scope restriction.",
				Blocks: map[string]schema.Block{
					"kubernetes": schema.SingleNestedBlock{
						Description: "Restricts the policy to specific Kubernetes namespaces. Requires `action_pattern.integration` to be `INTEGRATION_KUBERNETES`.",
						Blocks: map[string]schema.Block{
							"rule": schema.ListNestedBlock{
								Description: "A rule restricting the Kubernetes scope to specific namespaces.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"target_namespaces": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "The Kubernetes namespaces this rule applies to.",
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
					"scripts": schema.SingleNestedBlock{
						Description: "Restricts the policy to specific scripts. Requires `action_pattern.integration` to be `INTEGRATION_SCRIPT`.",
						Blocks: map[string]schema.Block{
							"rule": schema.ListNestedBlock{
								Description: "A rule restricting the script scope to specific script names.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"target_script_names": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "The script names this rule applies to.",
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
					"remote_action_rshell": schema.SingleNestedBlock{
						Description: "Restricts the policy to specific remote shell paths. Requires `action_pattern.integration` to be `INTEGRATION_REMOTE_ACTION`.",
						Blocks: map[string]schema.Block{
							"rule": schema.ListNestedBlock{
								Description: "A rule restricting remote shell access to specific paths.",
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"target_paths": schema.ListAttribute{
											Required:    true,
											ElementType: types.StringType,
											Description: "The filesystem paths this rule applies to.",
											Validators: []validator.List{
												listvalidator.SizeAtLeast(1),
											},
										},
										"access": schema.StringAttribute{
											Required:    true,
											Description: "The level of remote shell access granted for the target paths.",
											Validators: []validator.String{
												stringvalidator.OneOf(
													string(datadogV2.EXECUTIONPOLICYREMOTEACTIONRSHELLACCESS_READ_ONLY),
													string(datadogV2.EXECUTIONPOLICYREMOTEACTIONRSHELLACCESS_READ_WRITE),
												),
											},
										},
									},
								},
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
								},
							},
						},
					},
				},
			},
			"target": schema.ListNestedBlock{
				Description: "A target this policy is scoped to, expressed as a set of Agent tags. Each target is matched independently; omitting all `target` blocks applies the policy fleet-wide.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "A human-readable name for the target.",
						},
						"agent_tags": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
							Description: "The Agent tags identifying the target, for example `env:prod`.",
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

func (r *actionExecutionPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *actionExecutionPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan actionExecutionPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	attributes := buildExecutionPolicyWriteAttributes(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewExecutionPolicyCreateRequest(
		*datadogV2.NewExecutionPolicyCreateRequestData(*attributes, datadogV2.EXECUTIONPOLICYTYPE_EXECUTION_POLICY),
	)

	res, _, err := r.Api.CreateExecutionPolicy(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating execution policy"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating execution policy"))
		return
	}

	// Only the computed attributes are taken from the response; the writable ones stay as
	// planned so that Terraform never sees an inconsistent-result-after-apply error.
	setComputedExecutionPolicyState(&plan, &res.Data)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *actionExecutionPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state actionExecutionPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, httpResp, err := r.Api.GetExecutionPolicy(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading execution policy"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading execution policy"))
		return
	}

	updateExecutionPolicyStateFromResponse(ctx, &state, &res.Data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *actionExecutionPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state actionExecutionPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan actionExecutionPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	attributes := buildExecutionPolicyWriteAttributes(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	body := datadogV2.NewExecutionPolicyUpdateRequest(
		*datadogV2.NewExecutionPolicyUpdateRequestData(*attributes, id, datadogV2.EXECUTIONPOLICYTYPE_EXECUTION_POLICY),
	)

	res, _, err := r.Api.UpdateExecutionPolicy(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating execution policy"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating execution policy"))
		return
	}

	setComputedExecutionPolicyState(&plan, &res.Data)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *actionExecutionPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state actionExecutionPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.Api.DeleteExecutionPolicy(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting execution policy"))
	}
}

// setComputedExecutionPolicyState copies the server-owned attributes of a response onto the
// model, leaving every writable attribute untouched.
func setComputedExecutionPolicyState(state *actionExecutionPolicyModel, data *datadogV2.ExecutionPolicyResponseData) {
	state.ID = types.StringValue(data.Id)
	state.Version = types.Int32Value(data.Attributes.Version)
	state.CreatedAt = types.StringValue(data.Attributes.CreatedAt.Format(time.RFC3339))
	state.CreatedBy = types.StringValue(data.Attributes.CreatedBy)
	state.UpdatedAt = types.StringValue(data.Attributes.UpdatedAt.Format(time.RFC3339))
	state.UpdatedBy = types.StringValue(data.Attributes.UpdatedBy)
}

// updateExecutionPolicyStateFromResponse rebuilds the whole model from an API response. Used
// by Read, so that drift and imports both reflect what the API actually holds.
func updateExecutionPolicyStateFromResponse(ctx context.Context, state *actionExecutionPolicyModel, data *datadogV2.ExecutionPolicyResponseData, diags *diag.Diagnostics) {
	setComputedExecutionPolicyState(state, data)

	attributes := data.Attributes
	state.Name = types.StringValue(attributes.Name)
	state.Effect = types.StringValue(string(attributes.Effect))

	state.ActionPattern = &executionPolicyActionPatternModel{
		Integration: types.StringValue(string(attributes.ActionPattern.Integration)),
		ActionFqns:  executionPolicyStringListValue(ctx, attributes.ActionPattern.ActionFqns, diags),
	}

	apiScope := flattenExecutionPolicyScope(ctx, attributes.Scope, diags)
	stateHasExplicitEmptyScope := state.Scope != nil &&
		state.Scope.Kubernetes == nil &&
		state.Scope.Scripts == nil &&
		state.Scope.RemoteActionRshell == nil
	if apiScope != nil || !stateHasExplicitEmptyScope {
		state.Scope = apiScope
	}

	if len(attributes.Targets) == 0 {
		state.Target = nil
	} else {
		targets := make([]*executionPolicyTargetModel, 0, len(attributes.Targets))
		for _, target := range attributes.Targets {
			model := &executionPolicyTargetModel{
				AgentTags: executionPolicyStringListValue(ctx, target.AgentTags, diags),
			}
			if name, ok := target.GetNameOk(); ok && name != nil {
				model.Name = types.StringValue(*name)
			} else {
				model.Name = types.StringNull()
			}
			targets = append(targets, model)
		}
		state.Target = targets
	}
}

// flattenExecutionPolicyScope maps an API scope onto the model. A nil scope, or one where the
// API echoed an empty object, both mean "no scope restriction" and map to a nil block so that
// Terraform does not see a phantom `scope {}` in state.
func flattenExecutionPolicyScope(ctx context.Context, scope *datadogV2.ExecutionPolicyScope, diags *diag.Diagnostics) *executionPolicyScopeModel {
	if scope == nil {
		return nil
	}

	model := &executionPolicyScopeModel{}

	if scope.Kubernetes != nil {
		rules := make([]*executionPolicyKubernetesScopeRuleModel, 0, len(scope.Kubernetes.Rules))
		for _, rule := range scope.Kubernetes.Rules {
			rules = append(rules, &executionPolicyKubernetesScopeRuleModel{
				TargetNamespaces: executionPolicyStringListValue(ctx, rule.TargetNamespaces, diags),
			})
		}
		model.Kubernetes = &executionPolicyKubernetesScopeModel{Rule: rules}
	}

	if scope.Scripts != nil {
		rules := make([]*executionPolicyScriptScopeRuleModel, 0, len(scope.Scripts.Rules))
		for _, rule := range scope.Scripts.Rules {
			rules = append(rules, &executionPolicyScriptScopeRuleModel{
				TargetScriptNames: executionPolicyStringListValue(ctx, rule.TargetScriptNames, diags),
			})
		}
		model.Scripts = &executionPolicyScriptScopeModel{Rule: rules}
	}

	if scope.RemoteActionRshell != nil {
		rules := make([]*executionPolicyRemoteActionRshellScopeRuleModel, 0, len(scope.RemoteActionRshell.Rules))
		for _, rule := range scope.RemoteActionRshell.Rules {
			rules = append(rules, &executionPolicyRemoteActionRshellScopeRuleModel{
				TargetPaths: executionPolicyStringListValue(ctx, rule.TargetPaths, diags),
				Access:      types.StringValue(string(rule.Access)),
			})
		}
		model.RemoteActionRshell = &executionPolicyRemoteActionRshellScopeModel{Rule: rules}
	}

	if model.Kubernetes == nil && model.Scripts == nil && model.RemoteActionRshell == nil {
		return nil
	}
	return model
}

func buildExecutionPolicyWriteAttributes(ctx context.Context, plan *actionExecutionPolicyModel, diags *diag.Diagnostics) *datadogV2.ExecutionPolicyWriteAttributes {
	if plan.ActionPattern == nil {
		diags.AddAttributeError(
			frameworkPath.Root("action_pattern"),
			"Missing action_pattern block",
			"You must specify an action_pattern block.",
		)
		return nil
	}

	actionPattern := datadogV2.NewExecutionPolicyActionPattern(
		executionPolicyStringSlice(ctx, plan.ActionPattern.ActionFqns, diags),
		datadogV2.ExecutionPolicyIntegration(plan.ActionPattern.Integration.ValueString()),
	)

	attributes := datadogV2.NewExecutionPolicyWriteAttributes(
		*actionPattern,
		datadogV2.ExecutionPolicyEffect(plan.Effect.ValueString()),
		plan.Name.ValueString(),
	)

	if scope := expandExecutionPolicyScope(ctx, plan.Scope, diags); scope != nil {
		attributes.SetScope(*scope)
	}

	if len(plan.Target) > 0 {
		targets := make([]datadogV2.ExecutionPolicyTarget, 0, len(plan.Target))
		for _, target := range plan.Target {
			apiTarget := datadogV2.NewExecutionPolicyTarget(
				executionPolicyStringSlice(ctx, target.AgentTags, diags),
			)
			if !target.Name.IsNull() && !target.Name.IsUnknown() {
				apiTarget.SetName(target.Name.ValueString())
			}
			targets = append(targets, *apiTarget)
		}
		attributes.SetTargets(targets)
	}

	return attributes
}

func expandExecutionPolicyScope(ctx context.Context, scope *executionPolicyScopeModel, diags *diag.Diagnostics) *datadogV2.ExecutionPolicyScope {
	if scope == nil {
		return nil
	}

	apiScope := datadogV2.NewExecutionPolicyScope()
	set := false

	if scope.Kubernetes != nil {
		rules := make([]datadogV2.ExecutionPolicyKubernetesScopeRule, 0, len(scope.Kubernetes.Rule))
		for _, rule := range scope.Kubernetes.Rule {
			rules = append(rules, *datadogV2.NewExecutionPolicyKubernetesScopeRule(
				executionPolicyStringSlice(ctx, rule.TargetNamespaces, diags),
			))
		}
		apiScope.SetKubernetes(*datadogV2.NewExecutionPolicyKubernetesScope(rules))
		set = true
	}

	if scope.Scripts != nil {
		rules := make([]datadogV2.ExecutionPolicyScriptScopeRule, 0, len(scope.Scripts.Rule))
		for _, rule := range scope.Scripts.Rule {
			rules = append(rules, *datadogV2.NewExecutionPolicyScriptScopeRule(
				executionPolicyStringSlice(ctx, rule.TargetScriptNames, diags),
			))
		}
		apiScope.SetScripts(*datadogV2.NewExecutionPolicyScriptScope(rules))
		set = true
	}

	if scope.RemoteActionRshell != nil {
		rules := make([]datadogV2.ExecutionPolicyRemoteActionRshellScopeRule, 0, len(scope.RemoteActionRshell.Rule))
		for _, rule := range scope.RemoteActionRshell.Rule {
			rules = append(rules, *datadogV2.NewExecutionPolicyRemoteActionRshellScopeRule(
				datadogV2.ExecutionPolicyRemoteActionRshellAccess(rule.Access.ValueString()),
				executionPolicyStringSlice(ctx, rule.TargetPaths, diags),
			))
		}
		apiScope.SetRemoteActionRshell(*datadogV2.NewExecutionPolicyRemoteActionRshellScope(rules))
		set = true
	}

	if !set {
		return nil
	}
	return apiScope
}

func executionPolicyStringSlice(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(list.ElementsAs(ctx, &out, false)...)
	return out
}

func executionPolicyStringListValue(ctx context.Context, values []string, diags *diag.Diagnostics) types.List {
	list, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return list
}
