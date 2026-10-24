package fwprovider

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type rawFeatureFlagAllocation struct {
	ID               string                          `json:"id"`
	Key              string                          `json:"key"`
	Name             string                          `json:"name"`
	Type             string                          `json:"type"`
	OrderPosition    int64                           `json:"order_position"`
	ExperimentID     *string                         `json:"experiment_id"`
	TargetingRules   []rawFeatureFlagTargetingRule   `json:"targeting_rules"`
	VariantWeights   []rawFeatureFlagVariantWeight   `json:"variant_weights"`
	GuardrailMetrics []rawFeatureFlagGuardrailMetric `json:"guardrail_metrics"`
	ExposureSchedule *rawFeatureFlagExposureSchedule `json:"exposure_schedule"`
}

type rawFeatureFlagTargetingRule struct {
	Conditions []rawFeatureFlagAllocationCondition `json:"conditions"`
}

type rawFeatureFlagAllocationCondition struct {
	Attribute string   `json:"attribute"`
	Operator  string   `json:"operator"`
	Value     []string `json:"value"`
}

type rawFeatureFlagVariantWeight struct {
	Value      float64 `json:"value"`
	VariantID  string  `json:"variant_id"`
	VariantKey *string `json:"variant_key"`
	Variant    *struct {
		Key string `json:"key"`
	} `json:"variant"`
}

type rawFeatureFlagGuardrailMetric struct {
	MetricID      string `json:"metric_id"`
	TriggerAction string `json:"trigger_action"`
}

type rawFeatureFlagExposureSchedule struct {
	ID                string                        `json:"id"`
	ControlVariantID  string                        `json:"control_variant_id"`
	ControlVariantKey *string                       `json:"control_variant_key"`
	AbsoluteStartTime *string                       `json:"absolute_start_time"`
	RolloutOptions    *rawFeatureFlagRolloutOptions `json:"rollout_options"`
	RolloutSteps      []rawFeatureFlagRolloutStep   `json:"rollout_steps"`
}

type rawFeatureFlagRolloutOptions struct {
	Strategy            string `json:"strategy"`
	Autostart           bool   `json:"autostart"`
	SelectionIntervalMS int64  `json:"selection_interval_ms"`
}

type rawFeatureFlagRolloutStep struct {
	ID               string  `json:"id"`
	OrderPosition    int64   `json:"order_position"`
	ExposureRatio    float64 `json:"exposure_ratio"`
	IntervalMS       *int64  `json:"interval_ms"`
	IsPauseRecord    bool    `json:"is_pause_record"`
	GroupedStepIndex int64   `json:"grouped_step_index"`
}

func (r *featureFlagAllocationsResource) readAllocations(flagID, environmentID uuid.UUID) ([]featureFlagAllocationModel, error) {
	flag, httpResponse, err := r.Api.GetFeatureFlag(r.Auth, flagID)
	if err != nil {
		return nil, featureFlagResponseError(err, httpResponse, "error reading feature flag allocations")
	}
	if err := checkFeatureFlagResponse(&flag); err != nil {
		return nil, err
	}
	environment, err := findFeatureFlagEnvironment(&flag, environmentID)
	if err != nil {
		return nil, err
	}
	rawAllocations, ok := environment["allocations"]
	if !ok || rawAllocations == nil {
		return []featureFlagAllocationModel{}, nil
	}
	serialized, err := json.Marshal(rawAllocations)
	if err != nil {
		return nil, fmt.Errorf("marshal raw feature flag allocations: %w", err)
	}
	var raw []rawFeatureFlagAllocation
	if err := json.Unmarshal(serialized, &raw); err != nil {
		return nil, fmt.Errorf("decode feature flag allocations: %w", err)
	}
	allocations, err := rawFeatureFlagAllocationsToModel(raw, &flag)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(allocations, func(i, j int) bool {
		return allocations[i].OrderPosition.ValueInt64() < allocations[j].OrderPosition.ValueInt64()
	})
	return allocations, nil
}

func rawFeatureFlagAllocationsToModel(raw []rawFeatureFlagAllocation, flag *datadogV2.FeatureFlagResponse) ([]featureFlagAllocationModel, error) {
	variantKeys := featureFlagVariantIDToKey(flag)
	allocations := make([]featureFlagAllocationModel, 0, len(raw))
	for _, allocation := range raw {
		model := featureFlagAllocationModel{
			ID:            types.StringValue(allocation.ID),
			Key:           types.StringValue(allocation.Key),
			Name:          types.StringValue(allocation.Name),
			Type:          types.StringValue(allocation.Type),
			OrderPosition: types.Int64Value(allocation.OrderPosition),
			ExperimentID:  types.StringNull(),
		}
		if allocation.ExperimentID != nil {
			model.ExperimentID = types.StringValue(*allocation.ExperimentID)
		}
		model.TargetingRules = make([]featureFlagTargetingRuleModel, 0, len(allocation.TargetingRules))
		for _, rule := range allocation.TargetingRules {
			mappedRule := featureFlagTargetingRuleModel{Conditions: make([]featureFlagAllocationConditionModel, 0, len(rule.Conditions))}
			for _, condition := range rule.Conditions {
				mappedRule.Conditions = append(mappedRule.Conditions, featureFlagAllocationConditionModel{
					Attribute: types.StringValue(condition.Attribute),
					Operator:  types.StringValue(condition.Operator),
					Value:     append([]string{}, condition.Value...),
				})
			}
			model.TargetingRules = append(model.TargetingRules, mappedRule)
		}
		model.VariantWeights = make([]featureFlagVariantWeightModel, 0, len(allocation.VariantWeights))
		for _, weight := range allocation.VariantWeights {
			key := ""
			if weight.Variant != nil {
				key = weight.Variant.Key
			}
			if key == "" && weight.VariantKey != nil {
				key = *weight.VariantKey
			}
			if key == "" && weight.VariantID != "" {
				id, err := uuid.Parse(weight.VariantID)
				if err == nil {
					key = variantKeys[id]
				}
			}
			if key == "" {
				return nil, fmt.Errorf("allocation %q has a variant weight whose variant key cannot be resolved", allocation.Key)
			}
			model.VariantWeights = append(model.VariantWeights, featureFlagVariantWeightModel{VariantKey: types.StringValue(key), Value: types.Float64Value(weight.Value)})
		}
		for _, metric := range allocation.GuardrailMetrics {
			model.GuardrailMetrics = append(model.GuardrailMetrics, featureFlagGuardrailMetricModel{MetricID: types.StringValue(metric.MetricID), TriggerAction: types.StringValue(metric.TriggerAction)})
		}
		if allocation.ExposureSchedule != nil {
			schedule, err := rawFeatureFlagExposureScheduleToModel(*allocation.ExposureSchedule, variantKeys)
			if err != nil {
				return nil, fmt.Errorf("allocation %q: %w", allocation.Key, err)
			}
			model.ExposureSchedule = schedule
		}
		allocations = append(allocations, model)
	}
	return allocations, nil
}

func rawFeatureFlagExposureScheduleToModel(raw rawFeatureFlagExposureSchedule, variantKeys map[uuid.UUID]string) (*featureFlagExposureScheduleModel, error) {
	model := &featureFlagExposureScheduleModel{
		ID:                types.StringValue(raw.ID),
		ControlVariantKey: types.StringNull(),
		AbsoluteStartTime: types.StringNull(),
		RolloutSteps:      make([]featureFlagExposureRolloutStepModel, 0, len(raw.RolloutSteps)),
	}
	if raw.ControlVariantKey != nil {
		model.ControlVariantKey = types.StringValue(*raw.ControlVariantKey)
	} else if raw.ControlVariantID != "" {
		id, err := uuid.Parse(raw.ControlVariantID)
		if err == nil {
			if key := variantKeys[id]; key != "" {
				model.ControlVariantKey = types.StringValue(key)
			}
		}
	}
	if raw.AbsoluteStartTime != nil {
		model.AbsoluteStartTime = types.StringValue(*raw.AbsoluteStartTime)
	}
	if raw.RolloutOptions == nil {
		return nil, fmt.Errorf("exposure schedule is missing rollout_options")
	}
	model.RolloutOptions = &featureFlagRolloutOptionsModel{
		Strategy:            types.StringValue(raw.RolloutOptions.Strategy),
		Autostart:           types.BoolValue(raw.RolloutOptions.Autostart),
		SelectionIntervalMS: types.Int64Value(raw.RolloutOptions.SelectionIntervalMS),
	}
	for _, step := range raw.RolloutSteps {
		mapped := featureFlagExposureRolloutStepModel{
			ID:               types.StringValue(step.ID),
			OrderPosition:    types.Int64Value(step.OrderPosition),
			ExposureRatio:    types.Float64Value(step.ExposureRatio),
			IntervalMS:       types.Int64Null(),
			IsPauseRecord:    types.BoolValue(step.IsPauseRecord),
			GroupedStepIndex: types.Int64Value(step.GroupedStepIndex),
		}
		if step.IntervalMS != nil {
			mapped.IntervalMS = types.Int64Value(*step.IntervalMS)
		}
		model.RolloutSteps = append(model.RolloutSteps, mapped)
	}
	return model, nil
}

func (r *featureFlagAllocationsResource) writeAllocations(plan *featureFlagAllocationsResourceModel, prior []featureFlagAllocationModel) error {
	flagID, err := parseFeatureFlagUUID(plan.FeatureFlagID.ValueString(), "feature_flag_id")
	if err != nil {
		return err
	}
	environmentID, err := parseFeatureFlagUUID(plan.EnvironmentID.ValueString(), "environment_id")
	if err != nil {
		return err
	}
	flag, httpResponse, err := r.Api.GetFeatureFlag(r.Auth, flagID)
	if err != nil {
		return featureFlagResponseError(err, httpResponse, "error reading parent feature flag")
	}
	if err := checkFeatureFlagResponse(&flag); err != nil {
		return err
	}
	body, err := buildOverwriteFeatureFlagAllocationsRequest(plan.Allocations, prior, &flag)
	if err != nil {
		return err
	}
	result, httpResponse, err := r.Api.UpdateAllocationsForFeatureFlagInEnvironment(r.Auth, flagID, environmentID, body)
	if err != nil {
		return utils.TranslateClientError(err, httpResponse, "error replacing feature flag allocations")
	}
	if len(result.Data) == 0 && len(plan.Allocations) > 0 {
		return fmt.Errorf("Datadog returned no allocations after a non-empty replace request; check feature flag approval requirements and the Approvals Override permission")
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		return fmt.Errorf("invalid allocation write response: %w", err)
	}
	plan.ID = types.StringValue(featureFlagCompositeID(flagID, environmentID))
	fresh, err := r.readAllocations(flagID, environmentID)
	if err != nil {
		return fmt.Errorf("allocations were written but could not be read back: %w", err)
	}
	mergeFeatureFlagAllocationsAfterWrite(plan, fresh)
	return nil
}

func mergeFeatureFlagAllocationsAfterWrite(plan *featureFlagAllocationsResourceModel, fresh []featureFlagAllocationModel) {
	configuredByKey := make(map[string]featureFlagAllocationModel, len(plan.Allocations))
	for _, allocation := range plan.Allocations {
		configuredByKey[allocation.Key.ValueString()] = allocation
	}
	for i := range fresh {
		configured, ok := configuredByKey[fresh[i].Key.ValueString()]
		if !ok {
			continue
		}
		if configured.GuardrailMetrics == nil {
			fresh[i].GuardrailMetrics = nil
		}
		if configured.ExposureSchedule == nil {
			fresh[i].ExposureSchedule = nil
		}
	}
	plan.Allocations = fresh
}

func buildOverwriteFeatureFlagAllocationsRequest(allocations, prior []featureFlagAllocationModel, flag *datadogV2.FeatureFlagResponse) (datadogV2.OverwriteAllocationsRequest, error) {
	priorByKey := make(map[string]featureFlagAllocationModel, len(prior))
	for _, allocation := range prior {
		priorByKey[allocation.Key.ValueString()] = allocation
	}
	variantIDs := featureFlagVariantKeyToID(flag)
	data := make([]datadogV2.AllocationDataRequest, 0, len(allocations))
	keys := make(map[string]struct{}, len(allocations))
	for _, allocation := range allocations {
		key := allocation.Key.ValueString()
		if _, exists := keys[key]; exists {
			return datadogV2.OverwriteAllocationsRequest{}, fmt.Errorf("allocation key %q is declared more than once", key)
		}
		keys[key] = struct{}{}
		attributes := datadogV2.NewUpsertAllocationRequest(key, allocation.Name.ValueString(), datadogV2.AllocationType(allocation.Type.ValueString()))
		if previous, ok := priorByKey[key]; ok && !previous.ID.IsNull() && !previous.ID.IsUnknown() {
			id, err := uuid.Parse(previous.ID.ValueString())
			if err == nil {
				attributes.Id = &id
			}
		}
		if !allocation.ExperimentID.IsNull() && !allocation.ExperimentID.IsUnknown() {
			value := allocation.ExperimentID.ValueString()
			attributes.ExperimentId = *datadog.NewNullableString(&value)
		}
		attributes.TargetingRules = buildFeatureFlagTargetingRules(allocation.TargetingRules)
		weights := make([]datadogV2.VariantWeightRequest, 0, len(allocation.VariantWeights))
		for _, weight := range allocation.VariantWeights {
			variantKey := weight.VariantKey.ValueString()
			variantID, ok := variantIDs[variantKey]
			if !ok {
				return datadogV2.OverwriteAllocationsRequest{}, fmt.Errorf("allocation %q references unknown variant_key %q", key, variantKey)
			}
			weightRequest := datadogV2.NewVariantWeightRequest(weight.Value.ValueFloat64())
			weightRequest.VariantId = &variantID
			weightRequest.VariantKey = &variantKey
			weights = append(weights, *weightRequest)
		}
		attributes.VariantWeights = weights
		if allocation.GuardrailMetrics != nil {
			metrics := make([]datadogV2.GuardrailMetricRequest, 0, len(allocation.GuardrailMetrics))
			for _, metric := range allocation.GuardrailMetrics {
				metrics = append(metrics, *datadogV2.NewGuardrailMetricRequest(metric.MetricID.ValueString(), datadogV2.GuardrailTriggerAction(metric.TriggerAction.ValueString())))
			}
			attributes.GuardrailMetrics = metrics
		}
		if allocation.ExposureSchedule != nil {
			var previous *featureFlagExposureScheduleModel
			if old, ok := priorByKey[key]; ok {
				previous = old.ExposureSchedule
			}
			schedule, err := buildFeatureFlagExposureSchedule(allocation.ExposureSchedule, previous, variantIDs)
			if err != nil {
				return datadogV2.OverwriteAllocationsRequest{}, fmt.Errorf("allocation %q: %w", key, err)
			}
			attributes.ExposureSchedule = schedule
		}
		data = append(data, *datadogV2.NewAllocationDataRequest(*attributes, datadogV2.ALLOCATIONDATATYPE_ALLOCATIONS))
	}
	return *datadogV2.NewOverwriteAllocationsRequest(data), nil
}

func buildFeatureFlagTargetingRules(rules []featureFlagTargetingRuleModel) []datadogV2.TargetingRuleRequest {
	result := make([]datadogV2.TargetingRuleRequest, 0, len(rules))
	for _, rule := range rules {
		conditions := make([]datadogV2.ConditionRequest, 0, len(rule.Conditions))
		for _, condition := range rule.Conditions {
			mapped := datadogV2.NewConditionRequest()
			mapped.SetAttribute(condition.Attribute.ValueString())
			mapped.SetOperator(datadogV2.ConditionOperator(condition.Operator.ValueString()))
			mapped.SetValue(condition.Value)
			conditions = append(conditions, *mapped)
		}
		result = append(result, *datadogV2.NewTargetingRuleRequest(conditions))
	}
	return result
}

func buildFeatureFlagExposureSchedule(current, previous *featureFlagExposureScheduleModel, variantIDs map[string]uuid.UUID) (*datadogV2.ExposureScheduleRequest, error) {
	if current.RolloutOptions == nil {
		return nil, fmt.Errorf("rollout_options is required when exposure_schedule is configured")
	}
	strategy := current.RolloutOptions.Strategy.ValueString()
	if strategy == "" {
		return nil, fmt.Errorf("rollout_options.strategy is required")
	}
	options := datadogV2.NewRolloutOptionsRequest(datadogV2.RolloutStrategy(strategy))
	if !current.RolloutOptions.Autostart.IsNull() && !current.RolloutOptions.Autostart.IsUnknown() {
		value := current.RolloutOptions.Autostart.ValueBool()
		options.Autostart = *datadog.NewNullableBool(&value)
	} else if previous != nil && previous.RolloutOptions != nil && !previous.RolloutOptions.Autostart.IsNull() {
		value := previous.RolloutOptions.Autostart.ValueBool()
		options.Autostart = *datadog.NewNullableBool(&value)
	}
	if !current.RolloutOptions.SelectionIntervalMS.IsNull() && !current.RolloutOptions.SelectionIntervalMS.IsUnknown() {
		value := current.RolloutOptions.SelectionIntervalMS.ValueInt64()
		options.SelectionIntervalMs = &value
	} else if previous != nil && previous.RolloutOptions != nil && !previous.RolloutOptions.SelectionIntervalMS.IsNull() {
		value := previous.RolloutOptions.SelectionIntervalMS.ValueInt64()
		options.SelectionIntervalMs = &value
	}

	steps := make([]datadogV2.ExposureRolloutStepRequest, 0, len(current.RolloutSteps))
	for i, step := range current.RolloutSteps {
		var old *featureFlagExposureRolloutStepModel
		if previous != nil && i < len(previous.RolloutSteps) {
			old = &previous.RolloutSteps[i]
		}
		groupedStepIndex := int64(i)
		if !step.GroupedStepIndex.IsNull() && !step.GroupedStepIndex.IsUnknown() {
			groupedStepIndex = step.GroupedStepIndex.ValueInt64()
		} else if old != nil && !old.GroupedStepIndex.IsNull() {
			groupedStepIndex = old.GroupedStepIndex.ValueInt64()
		}
		isPauseRecord := false
		if !step.IsPauseRecord.IsNull() && !step.IsPauseRecord.IsUnknown() {
			isPauseRecord = step.IsPauseRecord.ValueBool()
		} else if old != nil {
			isPauseRecord = old.IsPauseRecord.ValueBool()
		}
		mapped := datadogV2.NewExposureRolloutStepRequest(step.ExposureRatio.ValueFloat64(), groupedStepIndex, isPauseRecord)
		if old != nil && !old.ID.IsNull() && !old.ID.IsUnknown() {
			id, err := uuid.Parse(old.ID.ValueString())
			if err == nil {
				mapped.Id = &id
			}
		}
		if !step.IntervalMS.IsNull() && !step.IntervalMS.IsUnknown() {
			value := step.IntervalMS.ValueInt64()
			mapped.IntervalMs = *datadog.NewNullableInt64(&value)
		} else if old != nil && !old.IntervalMS.IsNull() {
			value := old.IntervalMS.ValueInt64()
			mapped.IntervalMs = *datadog.NewNullableInt64(&value)
		}
		steps = append(steps, *mapped)
	}
	schedule := datadogV2.NewExposureScheduleRequest(*options, steps)
	if previous != nil && !previous.ID.IsNull() && !previous.ID.IsUnknown() {
		id, err := uuid.Parse(previous.ID.ValueString())
		if err == nil {
			schedule.Id = &id
		}
	}

	controlVariantKey := ""
	if !current.ControlVariantKey.IsNull() && !current.ControlVariantKey.IsUnknown() {
		controlVariantKey = current.ControlVariantKey.ValueString()
	} else if previous != nil && !previous.ControlVariantKey.IsNull() {
		controlVariantKey = previous.ControlVariantKey.ValueString()
	}
	if controlVariantKey != "" {
		if _, ok := variantIDs[controlVariantKey]; !ok {
			return nil, fmt.Errorf("unknown control_variant_key %q", controlVariantKey)
		}
		schedule.ControlVariantKey = *datadog.NewNullableString(&controlVariantKey)
	}

	absoluteStartTime := ""
	if !current.AbsoluteStartTime.IsNull() && !current.AbsoluteStartTime.IsUnknown() {
		absoluteStartTime = current.AbsoluteStartTime.ValueString()
	} else if previous != nil && !previous.AbsoluteStartTime.IsNull() {
		absoluteStartTime = previous.AbsoluteStartTime.ValueString()
	}
	if absoluteStartTime != "" {
		value, err := time.Parse(time.RFC3339Nano, absoluteStartTime)
		if err != nil {
			return nil, fmt.Errorf("absolute_start_time must be an RFC3339 timestamp: %w", err)
		}
		schedule.AbsoluteStartTime = *datadog.NewNullableTime(&value)
	}
	return schedule, nil
}
