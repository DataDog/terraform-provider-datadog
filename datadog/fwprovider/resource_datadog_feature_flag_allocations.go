package fwprovider

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_                                resource.ResourceWithConfigure   = &featureFlagAllocationsResource{}
	_                                resource.ResourceWithImportState = &featureFlagAllocationsResource{}
	featureFlagAllocationsWriteMutex sync.Mutex
)

func NewFeatureFlagAllocationsResource() resource.Resource {
	return &featureFlagAllocationsResource{}
}

type featureFlagAllocationsResource struct {
	Api  *datadogV2.FeatureFlagsApi
	Auth context.Context
}

type featureFlagAllocationsResourceModel struct {
	ID            types.String                 `tfsdk:"id"`
	FeatureFlagID types.String                 `tfsdk:"feature_flag_id"`
	EnvironmentID types.String                 `tfsdk:"environment_id"`
	Allocations   []featureFlagAllocationModel `tfsdk:"allocations"`
}

type featureFlagAllocationModel struct {
	ID               types.String                      `tfsdk:"id"`
	Key              types.String                      `tfsdk:"key"`
	Name             types.String                      `tfsdk:"name"`
	Type             types.String                      `tfsdk:"type"`
	OrderPosition    types.Int64                       `tfsdk:"order_position"`
	ExperimentID     types.String                      `tfsdk:"experiment_id"`
	TargetingRules   []featureFlagTargetingRuleModel   `tfsdk:"targeting_rules"`
	VariantWeights   []featureFlagVariantWeightModel   `tfsdk:"variant_weights"`
	GuardrailMetrics []featureFlagGuardrailMetricModel `tfsdk:"guardrail_metrics"`
	ExposureSchedule *featureFlagExposureScheduleModel `tfsdk:"exposure_schedule"`
}

type featureFlagTargetingRuleModel struct {
	Conditions []featureFlagAllocationConditionModel `tfsdk:"conditions"`
}

type featureFlagAllocationConditionModel struct {
	Attribute types.String `tfsdk:"attribute"`
	Operator  types.String `tfsdk:"operator"`
	Value     []string     `tfsdk:"value"`
}

type featureFlagVariantWeightModel struct {
	VariantKey types.String  `tfsdk:"variant_key"`
	Value      types.Float64 `tfsdk:"value"`
}

type featureFlagGuardrailMetricModel struct {
	MetricID      types.String `tfsdk:"metric_id"`
	TriggerAction types.String `tfsdk:"trigger_action"`
}

type featureFlagExposureScheduleModel struct {
	ID                types.String                          `tfsdk:"id"`
	ControlVariantKey types.String                          `tfsdk:"control_variant_key"`
	AbsoluteStartTime types.String                          `tfsdk:"absolute_start_time"`
	RolloutOptions    *featureFlagRolloutOptionsModel       `tfsdk:"rollout_options"`
	RolloutSteps      []featureFlagExposureRolloutStepModel `tfsdk:"rollout_steps"`
}

type featureFlagRolloutOptionsModel struct {
	Strategy            types.String `tfsdk:"strategy"`
	Autostart           types.Bool   `tfsdk:"autostart"`
	SelectionIntervalMS types.Int64  `tfsdk:"selection_interval_ms"`
}

type featureFlagExposureRolloutStepModel struct {
	ID               types.String  `tfsdk:"id"`
	OrderPosition    types.Int64   `tfsdk:"order_position"`
	ExposureRatio    types.Float64 `tfsdk:"exposure_ratio"`
	IntervalMS       types.Int64   `tfsdk:"interval_ms"`
	IsPauseRecord    types.Bool    `tfsdk:"is_pause_record"`
	GroupedStepIndex types.Int64   `tfsdk:"grouped_step_index"`
}

func (r *featureFlagAllocationsResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "feature_flag_allocations"
}

func (r *featureFlagAllocationsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Manages the complete ordered allocation list for one feature flag in one environment. Every write uses the API's replace-all operation. Progressive rollout start, pause, resume, and stop actions are intentionally not managed.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"feature_flag_id": schema.StringAttribute{
				Description: "UUID of the feature flag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "UUID of the feature flag environment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allocations": featureFlagAllocationsAttribute(),
		},
	}
}

func (r *featureFlagAllocationsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetFeatureFlagsApiV2()
	r.Auth = providerData.Auth
}

func (r *featureFlagAllocationsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan featureFlagAllocationsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	featureFlagAllocationsWriteMutex.Lock()
	err := r.writeAllocations(&plan, nil)
	featureFlagAllocationsWriteMutex.Unlock()
	if err != nil {
		response.Diagnostics.AddError("Feature flag allocations creation failed", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagAllocationsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state featureFlagAllocationsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	flagID, environmentID, err := parseFeatureFlagCompositeID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag allocations ID", err.Error())
		return
	}
	allocations, err := r.readAllocations(flagID, environmentID)
	if err != nil {
		if errors.Is(err, errFeatureFlagEnvironmentNotFound) || isFeatureFlagHTTPNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Feature flag allocations read failed", err.Error())
		return
	}
	state.FeatureFlagID = types.StringValue(flagID.String())
	state.EnvironmentID = types.StringValue(environmentID.String())
	state.Allocations = allocations
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *featureFlagAllocationsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan featureFlagAllocationsResourceModel
	var prior featureFlagAllocationsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &prior)...)
	if response.Diagnostics.HasError() {
		return
	}
	featureFlagAllocationsWriteMutex.Lock()
	err := r.writeAllocations(&plan, prior.Allocations)
	featureFlagAllocationsWriteMutex.Unlock()
	if err != nil {
		response.Diagnostics.AddError("Feature flag allocations update failed", err.Error())
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *featureFlagAllocationsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state featureFlagAllocationsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	flagID, environmentID, err := parseFeatureFlagCompositeID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Invalid feature flag allocations ID", err.Error())
		return
	}
	body := *datadogV2.NewOverwriteAllocationsRequest([]datadogV2.AllocationDataRequest{})
	featureFlagAllocationsWriteMutex.Lock()
	_, httpResponse, err := r.Api.UpdateAllocationsForFeatureFlagInEnvironment(r.Auth, flagID, environmentID, body)
	featureFlagAllocationsWriteMutex.Unlock()
	if err != nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
		response.Diagnostics.AddError("Feature flag allocations delete failed", utils.TranslateClientError(err, httpResponse, "error clearing feature flag allocations").Error())
	}
}

func (r *featureFlagAllocationsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importFeatureFlagCompositeState(ctx, request, response)
}
