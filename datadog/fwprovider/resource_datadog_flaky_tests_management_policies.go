package fwprovider

import (
	"context"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &flakyTestsManagementPoliciesResource{}
	_ resource.ResourceWithImportState = &flakyTestsManagementPoliciesResource{}
)

type flakyTestsManagementPoliciesResource struct {
	Api  *datadogV2.TestOptimizationApi
	Auth context.Context
}

type flakyTestsManagementPoliciesModel struct {
	ID           types.String                       `tfsdk:"id"`
	RepositoryId types.String                       `tfsdk:"repository_id"`
	AttemptToFix *flakyTestsPolicyAttemptToFixModel `tfsdk:"attempt_to_fix"`
	Disabled     *flakyTestsPolicyDisabledModel     `tfsdk:"disabled"`
	Quarantined  *flakyTestsPolicyQuarantinedModel  `tfsdk:"quarantined"`
}

type flakyTestsPolicyAttemptToFixModel struct {
	Retries types.Int64 `tfsdk:"retries"`
}

type flakyTestsPolicyDisabledModel struct {
	Enabled         types.Bool                                    `tfsdk:"enabled"`
	AutoDisableRule *flakyTestsPolicyAutoDisableRuleModel         `tfsdk:"auto_disable_rule"`
	BranchRule      *flakyTestsPolicyBranchRuleModel              `tfsdk:"branch_rule"`
	FailureRateRule *flakyTestsPolicyDisabledFailureRateRuleModel `tfsdk:"failure_rate_rule"`
}

type flakyTestsPolicyQuarantinedModel struct {
	Enabled            types.Bool                                       `tfsdk:"enabled"`
	AutoQuarantineRule *flakyTestsPolicyAutoQuarantineRuleModel         `tfsdk:"auto_quarantine_rule"`
	BranchRule         *flakyTestsPolicyBranchRuleModel                 `tfsdk:"branch_rule"`
	FailureRateRule    *flakyTestsPolicyQuarantinedFailureRateRuleModel `tfsdk:"failure_rate_rule"`
}

type flakyTestsPolicyAutoDisableRuleModel struct {
	Enabled       types.Bool   `tfsdk:"enabled"`
	Status        types.String `tfsdk:"status"`
	WindowSeconds types.Int64  `tfsdk:"window_seconds"`
}

type flakyTestsPolicyAutoQuarantineRuleModel struct {
	Enabled       types.Bool  `tfsdk:"enabled"`
	WindowSeconds types.Int64 `tfsdk:"window_seconds"`
}

type flakyTestsPolicyBranchRuleModel struct {
	Branches             types.Set  `tfsdk:"branches"`
	Enabled              types.Bool `tfsdk:"enabled"`
	ExcludedBranches     types.Set  `tfsdk:"excluded_branches"`
	ExcludedTestServices types.Set  `tfsdk:"excluded_test_services"`
}

type flakyTestsPolicyDisabledFailureRateRuleModel struct {
	Branches  types.Set     `tfsdk:"branches"`
	Enabled   types.Bool    `tfsdk:"enabled"`
	MinRuns   types.Int64   `tfsdk:"min_runs"`
	Status    types.String  `tfsdk:"status"`
	Threshold types.Float64 `tfsdk:"threshold"`
}

type flakyTestsPolicyQuarantinedFailureRateRuleModel struct {
	Branches  types.Set     `tfsdk:"branches"`
	Enabled   types.Bool    `tfsdk:"enabled"`
	MinRuns   types.Int64   `tfsdk:"min_runs"`
	Threshold types.Float64 `tfsdk:"threshold"`
}

func NewFlakyTestsManagementPoliciesResource() resource.Resource {
	return &flakyTestsManagementPoliciesResource{}
}

func (r *flakyTestsManagementPoliciesResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "flaky_tests_management_policies"
}

func (r *flakyTestsManagementPoliciesResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTestOptimizationApiV2()
	r.Auth = providerData.Auth
}

func flakyTestsPolicyStatusValidators() []validator.String {
	return []validator.String{
		stringvalidator.OneOf("active", "quarantined"),
	}
}

func flakyTestsStringSetAttribute(description string) schema.SetAttribute {
	return schema.SetAttribute{
		Optional:    true,
		Computed:    true,
		Description: description,
		ElementType: types.StringType,
	}
}

func (r *flakyTestsManagementPoliciesResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	branchRuleAttributes := func() map[string]schema.Attribute {
		return map[string]schema.Attribute{
			"branches":               flakyTestsStringSetAttribute("List of branches to which the policy applies."),
			"enabled":                flakyTestsBoolAttribute("Whether this branch rule is enabled."),
			"excluded_branches":      flakyTestsStringSetAttribute("List of branches excluded from the policy."),
			"excluded_test_services": flakyTestsStringSetAttribute("List of test services excluded from the policy."),
		}
	}

	response.Schema = schema.Schema{
		Description: "Provides a Datadog Flaky Tests Management policies resource. This can be used to configure the flaky test lifecycle policies (quarantine, disable, attempt to fix) for a repository. Note: the API does not support deleting policies; removing the resource from your configuration stops managing it but leaves the policies in place.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "The repository identifier (e.g. `github.com/org/repo`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attempt_to_fix": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Configuration for the attempt-to-fix Flaky Tests Management policy.",
				Attributes: map[string]schema.Attribute{
					"retries": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Number of retries when attempting to fix a flaky test. Must be greater than 0.",
					},
				},
			},
			"disabled": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Configuration for the disabled Flaky Tests Management policy.",
				Attributes: map[string]schema.Attribute{
					"enabled": flakyTestsBoolAttribute("Whether the disabled policy is enabled."),
					"auto_disable_rule": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Automatic disable triggering rule based on a time window and test status.",
						Attributes: map[string]schema.Attribute{
							"enabled":        flakyTestsBoolAttribute("Whether this auto-disable rule is enabled."),
							"status":         flakyTestsStatusAttribute("Test status that the disable policy applies to."),
							"window_seconds": flakyTestsInt64Attribute("Time window in seconds over which flakiness is evaluated. Must be greater than 0."),
						},
					},
					"branch_rule": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Branch filtering rule for the disabled policy.",
						Attributes:  branchRuleAttributes(),
					},
					"failure_rate_rule": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Failure-rate-based rule for the disabled policy.",
						Attributes: map[string]schema.Attribute{
							"branches":  flakyTestsStringSetAttribute("List of branches to which this rule applies."),
							"enabled":   flakyTestsBoolAttribute("Whether this failure rate rule is enabled."),
							"min_runs":  flakyTestsInt64Attribute("Minimum number of runs required before the rule is evaluated. Must be greater than or equal to 0."),
							"status":    flakyTestsStatusAttribute("Test status that the disable policy applies to."),
							"threshold": flakyTestsFloat64Attribute("Failure rate threshold (0.0-1.0) above which the rule triggers."),
						},
					},
				},
			},
			"quarantined": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Configuration for the quarantined Flaky Tests Management policy.",
				Attributes: map[string]schema.Attribute{
					"enabled": flakyTestsBoolAttribute("Whether the quarantined policy is enabled."),
					"auto_quarantine_rule": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Automatic quarantine triggering rule based on a time window.",
						Attributes: map[string]schema.Attribute{
							"enabled":        flakyTestsBoolAttribute("Whether this auto-quarantine rule is enabled."),
							"window_seconds": flakyTestsInt64Attribute("Time window in seconds over which flakiness is evaluated. Must be greater than 0."),
						},
					},
					"branch_rule": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Branch filtering rule for the quarantined policy.",
						Attributes:  branchRuleAttributes(),
					},
					"failure_rate_rule": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Failure-rate-based rule for the quarantined policy.",
						Attributes: map[string]schema.Attribute{
							"branches":  flakyTestsStringSetAttribute("List of branches to which this rule applies."),
							"enabled":   flakyTestsBoolAttribute("Whether this failure rate rule is enabled."),
							"min_runs":  flakyTestsInt64Attribute("Minimum number of runs required before the rule is evaluated. Must be greater than or equal to 0."),
							"threshold": flakyTestsFloat64Attribute("Failure rate threshold (0.0-1.0) above which the rule triggers."),
						},
					},
				},
			},
		},
	}
}

func flakyTestsBoolAttribute(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}

func flakyTestsInt64Attribute(description string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}

func flakyTestsFloat64Attribute(description string) schema.Float64Attribute {
	return schema.Float64Attribute{
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}

func flakyTestsStatusAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: description,
		Validators:  flakyTestsPolicyStatusValidators(),
	}
}

func (r *flakyTestsManagementPoliciesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("repository_id"), request.ID)...)
}

func (r *flakyTestsManagementPoliciesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan flakyTestsManagementPoliciesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildUpdateRequest(&plan)
	res, _, err := r.Api.UpdateFlakyTestsManagementPolicies(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating flaky tests management policies"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(&plan, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *flakyTestsManagementPoliciesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state flakyTestsManagementPoliciesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	repositoryId := state.RepositoryId.ValueString()
	if repositoryId == "" {
		repositoryId = state.ID.ValueString()
	}

	getRequest := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesGetRequest(
		*datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesGetRequestData(
			*datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesGetRequestAttributes(repositoryId),
			datadogV2.TESTOPTIMIZATIONGETFLAKYTESTSMANAGEMENTPOLICIESREQUESTDATATYPE_TEST_OPTIMIZATION_GET_FLAKY_TESTS_MANAGEMENT_POLICIES_REQUEST,
		),
	)
	res, httpResponse, err := r.Api.GetFlakyTestsManagementPolicies(r.Auth, *getRequest)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching flaky tests management policies"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(&state, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *flakyTestsManagementPoliciesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan flakyTestsManagementPoliciesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildUpdateRequest(&plan)
	res, _, err := r.Api.UpdateFlakyTestsManagementPolicies(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating flaky tests management policies"))
		return
	}
	if err := utils.CheckForUnparsed(res); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "response contains unparsed object"))
		return
	}

	r.updateStateFromResponse(&plan, &res)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *flakyTestsManagementPoliciesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// The API does not expose a delete endpoint for flaky tests management
	// policies; deleting the resource only removes it from the Terraform state.
}

func stringSetToSlice(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var out []string
	for _, elem := range set.Elements() {
		if s, ok := elem.(types.String); ok {
			out = append(out, s.ValueString())
		}
	}
	return out
}

func stringSetFromSlice(in []string) types.Set {
	if in == nil {
		return types.SetNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(in))
	for _, s := range in {
		elems = append(elems, types.StringValue(s))
	}
	return types.SetValueMust(types.StringType, elems)
}

func (r *flakyTestsManagementPoliciesResource) buildUpdateRequest(model *flakyTestsManagementPoliciesModel) *datadogV2.TestOptimizationFlakyTestsManagementPoliciesUpdateRequest {
	attributes := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesUpdateRequestAttributes(model.RepositoryId.ValueString())

	if m := model.AttemptToFix; m != nil && !m.Retries.IsUnknown() && !m.Retries.IsNull() {
		attemptToFix := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesAttemptToFix()
		attemptToFix.SetRetries(m.Retries.ValueInt64())
		attributes.SetAttemptToFix(*attemptToFix)
	}

	if m := model.Disabled; m != nil {
		disabled := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesDisabled()
		if !m.Enabled.IsUnknown() && !m.Enabled.IsNull() {
			disabled.SetEnabled(m.Enabled.ValueBool())
		}
		if rule := m.AutoDisableRule; rule != nil {
			autoDisable := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesAutoDisableRule()
			if !rule.Enabled.IsUnknown() && !rule.Enabled.IsNull() {
				autoDisable.SetEnabled(rule.Enabled.ValueBool())
			}
			if !rule.Status.IsUnknown() && !rule.Status.IsNull() {
				autoDisable.SetStatus(datadogV2.TestOptimizationFlakyTestsManagementPoliciesDisabledStatus(rule.Status.ValueString()))
			}
			if !rule.WindowSeconds.IsUnknown() && !rule.WindowSeconds.IsNull() {
				autoDisable.SetWindowSeconds(rule.WindowSeconds.ValueInt64())
			}
			disabled.SetAutoDisableRule(*autoDisable)
		}
		if rule := m.BranchRule; rule != nil {
			disabled.SetBranchRule(*buildBranchRule(rule))
		}
		if rule := m.FailureRateRule; rule != nil {
			failureRate := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesDisabledFailureRateRule()
			if branches := stringSetToSlice(rule.Branches); branches != nil {
				failureRate.SetBranches(branches)
			}
			if !rule.Enabled.IsUnknown() && !rule.Enabled.IsNull() {
				failureRate.SetEnabled(rule.Enabled.ValueBool())
			}
			if !rule.MinRuns.IsUnknown() && !rule.MinRuns.IsNull() {
				failureRate.SetMinRuns(rule.MinRuns.ValueInt64())
			}
			if !rule.Status.IsUnknown() && !rule.Status.IsNull() {
				failureRate.SetStatus(datadogV2.TestOptimizationFlakyTestsManagementPoliciesDisabledStatus(rule.Status.ValueString()))
			}
			if !rule.Threshold.IsUnknown() && !rule.Threshold.IsNull() {
				failureRate.SetThreshold(rule.Threshold.ValueFloat64())
			}
			disabled.SetFailureRateRule(*failureRate)
		}
		attributes.SetDisabled(*disabled)
	}

	if m := model.Quarantined; m != nil {
		quarantined := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesQuarantined()
		if !m.Enabled.IsUnknown() && !m.Enabled.IsNull() {
			quarantined.SetEnabled(m.Enabled.ValueBool())
		}
		if rule := m.AutoQuarantineRule; rule != nil {
			autoQuarantine := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesAutoQuarantineRule()
			if !rule.Enabled.IsUnknown() && !rule.Enabled.IsNull() {
				autoQuarantine.SetEnabled(rule.Enabled.ValueBool())
			}
			if !rule.WindowSeconds.IsUnknown() && !rule.WindowSeconds.IsNull() {
				autoQuarantine.SetWindowSeconds(rule.WindowSeconds.ValueInt64())
			}
			quarantined.SetAutoQuarantineRule(*autoQuarantine)
		}
		if rule := m.BranchRule; rule != nil {
			quarantined.SetBranchRule(*buildBranchRule(rule))
		}
		if rule := m.FailureRateRule; rule != nil {
			failureRate := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesQuarantinedFailureRateRule()
			if branches := stringSetToSlice(rule.Branches); branches != nil {
				failureRate.SetBranches(branches)
			}
			if !rule.Enabled.IsUnknown() && !rule.Enabled.IsNull() {
				failureRate.SetEnabled(rule.Enabled.ValueBool())
			}
			if !rule.MinRuns.IsUnknown() && !rule.MinRuns.IsNull() {
				failureRate.SetMinRuns(rule.MinRuns.ValueInt64())
			}
			if !rule.Threshold.IsUnknown() && !rule.Threshold.IsNull() {
				failureRate.SetThreshold(rule.Threshold.ValueFloat64())
			}
			quarantined.SetFailureRateRule(*failureRate)
		}
		attributes.SetQuarantined(*quarantined)
	}

	data := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesUpdateRequestData(
		*attributes,
		datadogV2.TESTOPTIMIZATIONUPDATEFLAKYTESTSMANAGEMENTPOLICIESREQUESTDATATYPE_TEST_OPTIMIZATION_UPDATE_FLAKY_TESTS_MANAGEMENT_POLICIES_REQUEST,
	)
	return datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesUpdateRequest(*data)
}

func buildBranchRule(m *flakyTestsPolicyBranchRuleModel) *datadogV2.TestOptimizationFlakyTestsManagementPoliciesBranchRule {
	rule := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesBranchRule()
	if branches := stringSetToSlice(m.Branches); branches != nil {
		rule.SetBranches(branches)
	}
	if !m.Enabled.IsUnknown() && !m.Enabled.IsNull() {
		rule.SetEnabled(m.Enabled.ValueBool())
	}
	if excluded := stringSetToSlice(m.ExcludedBranches); excluded != nil {
		rule.SetExcludedBranches(excluded)
	}
	if excluded := stringSetToSlice(m.ExcludedTestServices); excluded != nil {
		rule.SetExcludedTestServices(excluded)
	}
	return rule
}

func (r *flakyTestsManagementPoliciesResource) updateStateFromResponse(model *flakyTestsManagementPoliciesModel, res *datadogV2.TestOptimizationFlakyTestsManagementPoliciesResponse) {

	data := res.GetData()
	model.ID = types.StringValue(data.GetId())
	attributes := data.GetAttributes()
	model.RepositoryId = types.StringValue(attributes.GetRepositoryId())

	if attemptToFix, ok := attributes.GetAttemptToFixOk(); ok {
		model.AttemptToFix = &flakyTestsPolicyAttemptToFixModel{
			Retries: types.Int64Value(attemptToFix.GetRetries()),
		}
	}

	if disabled, ok := attributes.GetDisabledOk(); ok {
		m := &flakyTestsPolicyDisabledModel{
			Enabled: types.BoolValue(disabled.GetEnabled()),
		}
		if rule, ok := disabled.GetAutoDisableRuleOk(); ok {
			m.AutoDisableRule = &flakyTestsPolicyAutoDisableRuleModel{
				Enabled:       types.BoolValue(rule.GetEnabled()),
				Status:        types.StringValue(string(rule.GetStatus())),
				WindowSeconds: types.Int64Value(rule.GetWindowSeconds()),
			}
		}
		if rule, ok := disabled.GetBranchRuleOk(); ok {
			m.BranchRule = branchRuleModelFromAPI(rule)
		}
		if rule, ok := disabled.GetFailureRateRuleOk(); ok {
			m.FailureRateRule = &flakyTestsPolicyDisabledFailureRateRuleModel{
				Branches:  stringSetFromSlice(rule.GetBranches()),
				Enabled:   types.BoolValue(rule.GetEnabled()),
				MinRuns:   types.Int64Value(rule.GetMinRuns()),
				Status:    types.StringValue(string(rule.GetStatus())),
				Threshold: types.Float64Value(rule.GetThreshold()),
			}
		}
		model.Disabled = m
	}

	if quarantined, ok := attributes.GetQuarantinedOk(); ok {
		m := &flakyTestsPolicyQuarantinedModel{
			Enabled: types.BoolValue(quarantined.GetEnabled()),
		}
		if rule, ok := quarantined.GetAutoQuarantineRuleOk(); ok {
			m.AutoQuarantineRule = &flakyTestsPolicyAutoQuarantineRuleModel{
				Enabled:       types.BoolValue(rule.GetEnabled()),
				WindowSeconds: types.Int64Value(rule.GetWindowSeconds()),
			}
		}
		if rule, ok := quarantined.GetBranchRuleOk(); ok {
			m.BranchRule = branchRuleModelFromAPI(rule)
		}
		if rule, ok := quarantined.GetFailureRateRuleOk(); ok {
			m.FailureRateRule = &flakyTestsPolicyQuarantinedFailureRateRuleModel{
				Branches:  stringSetFromSlice(rule.GetBranches()),
				Enabled:   types.BoolValue(rule.GetEnabled()),
				MinRuns:   types.Int64Value(rule.GetMinRuns()),
				Threshold: types.Float64Value(rule.GetThreshold()),
			}
		}
		model.Quarantined = m
	}
}

func branchRuleModelFromAPI(rule *datadogV2.TestOptimizationFlakyTestsManagementPoliciesBranchRule) *flakyTestsPolicyBranchRuleModel {
	return &flakyTestsPolicyBranchRuleModel{
		Branches:             stringSetFromSlice(rule.GetBranches()),
		Enabled:              types.BoolValue(rule.GetEnabled()),
		ExcludedBranches:     stringSetFromSlice(rule.GetExcludedBranches()),
		ExcludedTestServices: stringSetFromSlice(rule.GetExcludedTestServices()),
	}
}
