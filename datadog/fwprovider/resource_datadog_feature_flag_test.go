package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFeatureFlagResourceSchemas(t *testing.T) {
	t.Parallel()
	resources := map[string]func() resource.Resource{
		"feature_flag":                    NewFeatureFlagResource,
		"feature_flag_environment":        NewFeatureFlagEnvironmentResource,
		"feature_flag_environment_config": NewFeatureFlagEnvironmentConfigResource,
		"feature_flag_allocations":        NewFeatureFlagAllocationsResource,
	}
	for name, constructor := range resources {
		t.Run(name, func(t *testing.T) {
			r := constructor()
			var response resource.SchemaResponse
			r.Schema(context.Background(), resource.SchemaRequest{}, &response)
			response.Diagnostics.Append(response.Schema.Validate()...)
			response.Diagnostics.Append(response.Schema.ValidateImplementation(context.Background())...)
			if response.Diagnostics.HasError() {
				t.Fatalf("invalid schema: %v", response.Diagnostics.Errors())
			}
		})
	}
}

func TestFeatureFlagCompositeIDRoundTrip(t *testing.T) {
	t.Parallel()
	flagID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	environmentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	gotFlagID, gotEnvironmentID, err := parseFeatureFlagCompositeID(featureFlagCompositeID(flagID, environmentID))
	if err != nil {
		t.Fatalf("parse composite ID: %v", err)
	}
	if gotFlagID != flagID || gotEnvironmentID != environmentID {
		t.Fatalf("got %s:%s, want %s:%s", gotFlagID, gotEnvironmentID, flagID, environmentID)
	}
	for _, invalid := range []string{"", "not-a-pair", "not-a-uuid:also-not-a-uuid", "a:b:c"} {
		if _, _, err := parseFeatureFlagCompositeID(invalid); err == nil {
			t.Errorf("parseFeatureFlagCompositeID(%q) unexpectedly succeeded", invalid)
		}
	}
}

func TestFeatureFlagEnvironmentWithAllocationsUsesRawJSON(t *testing.T) {
	t.Parallel()
	const payload = `{
		"data": {
			"id": "11111111-1111-1111-1111-111111111111",
			"type": "feature-flags",
			"attributes": {
				"description": "Checkout rollout",
				"key": "checkout",
				"name": "Checkout",
				"value_type": "BOOLEAN",
				"variants": [
					{"id": "33333333-3333-3333-3333-333333333333", "key": "on", "name": "On", "value": "true"},
					{"id": "55555555-5555-5555-5555-555555555555", "key": "off", "name": "Off", "value": "false"}
				],
				"feature_flag_environments": [{
					"environment_id": "22222222-2222-2222-2222-222222222222",
					"status": "ENABLED",
					"default_variant_id": "33333333-3333-3333-3333-333333333333",
					"allocations": [{
						"id": "44444444-4444-4444-4444-444444444444",
						"key": "checkout-staff",
						"name": "Checkout staff",
						"type": "FEATURE_GATE",
						"order_position": 0,
						"targeting_rules": [{"conditions": [{"attribute": "user.tier", "operator": "EQUALS", "value": ["staff"]}]}],
						"variant_weights": [{"variant_id": "33333333-3333-3333-3333-333333333333", "value": 100}],
						"guardrail_metrics": []
					}]
				}]
			}
		}
	}`
	var response datadogV2.FeatureFlagResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if err := checkFeatureFlagResponse(&response); err != nil {
		t.Fatalf("expected environment mismatch to be isolated: %v", err)
	}
	environmentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	raw, err := findFeatureFlagEnvironment(&response, environmentID)
	if err != nil {
		t.Fatalf("find environment: %v", err)
	}
	if raw["status"] != "ENABLED" {
		t.Fatalf("status = %v, want ENABLED", raw["status"])
	}
	if got := defaultFeatureFlagVariantKey(&response); got.ValueString() != "on" {
		t.Fatalf("default variant = %q, want on", got.ValueString())
	}
	serialized, err := json.Marshal(raw["allocations"])
	if err != nil {
		t.Fatalf("marshal allocations: %v", err)
	}
	var rawAllocations []rawFeatureFlagAllocation
	if err := json.Unmarshal(serialized, &rawAllocations); err != nil {
		t.Fatalf("decode allocations: %v", err)
	}
	allocations, err := rawFeatureFlagAllocationsToModel(rawAllocations, &response)
	if err != nil {
		t.Fatalf("map allocations: %v", err)
	}
	if len(allocations) != 1 || allocations[0].VariantWeights[0].VariantKey.ValueString() != "on" {
		t.Fatalf("unexpected allocations: %+v", allocations)
	}
	if allocations[0].TargetingRules[0].Conditions[0].Value[0] != "staff" {
		t.Fatalf("unexpected targeting rules: %+v", allocations[0].TargetingRules)
	}
}

func TestBuildOverwriteFeatureFlagAllocationsPreservesIdentity(t *testing.T) {
	t.Parallel()
	variantID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	allocationID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	flag := &datadogV2.FeatureFlagResponse{Data: datadogV2.FeatureFlag{Attributes: datadogV2.FeatureFlagAttributes{
		Variants: []datadogV2.Variant{{Id: variantID, Key: "on", Name: "On", Value: "true"}},
	}}}
	allocation := featureFlagAllocationModel{
		ID:             types.StringUnknown(),
		Key:            types.StringValue("checkout-staff"),
		Name:           types.StringValue("Checkout staff"),
		Type:           types.StringValue("FEATURE_GATE"),
		ExperimentID:   types.StringNull(),
		TargetingRules: []featureFlagTargetingRuleModel{},
		VariantWeights: []featureFlagVariantWeightModel{{
			VariantKey: types.StringValue("on"),
			Value:      types.Float64Value(100),
		}},
	}
	prior := allocation
	prior.ID = types.StringValue(allocationID.String())
	body, err := buildOverwriteFeatureFlagAllocationsRequest([]featureFlagAllocationModel{allocation}, []featureFlagAllocationModel{prior}, flag)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("got %d allocations, want 1", len(body.Data))
	}
	attributes := body.Data[0].Attributes
	if attributes.Id == nil || *attributes.Id != allocationID {
		t.Fatalf("allocation ID = %v, want %s", attributes.Id, allocationID)
	}
	if len(attributes.VariantWeights) != 1 || attributes.VariantWeights[0].VariantId == nil || *attributes.VariantWeights[0].VariantId != variantID {
		t.Fatalf("variant weights = %+v", attributes.VariantWeights)
	}
	if _, err := buildOverwriteFeatureFlagAllocationsRequest([]featureFlagAllocationModel{allocation, allocation}, nil, flag); err == nil {
		t.Fatal("duplicate allocation keys unexpectedly accepted")
	}
	allocation.VariantWeights[0].VariantKey = types.StringValue("missing")
	if _, err := buildOverwriteFeatureFlagAllocationsRequest([]featureFlagAllocationModel{allocation}, nil, flag); err == nil {
		t.Fatal("unknown variant key unexpectedly accepted")
	}
}

func TestBuildFeatureFlagExposureScheduleDefaults(t *testing.T) {
	t.Parallel()
	schedule := &featureFlagExposureScheduleModel{
		ID:                types.StringUnknown(),
		ControlVariantKey: types.StringUnknown(),
		AbsoluteStartTime: types.StringUnknown(),
		RolloutOptions: &featureFlagRolloutOptionsModel{
			Strategy:            types.StringValue("NO_ROLLOUT"),
			Autostart:           types.BoolUnknown(),
			SelectionIntervalMS: types.Int64Unknown(),
		},
		RolloutSteps: []featureFlagExposureRolloutStepModel{{
			ID:               types.StringUnknown(),
			ExposureRatio:    types.Float64Value(1),
			IntervalMS:       types.Int64Unknown(),
			IsPauseRecord:    types.BoolValue(false),
			GroupedStepIndex: types.Int64Unknown(),
		}},
	}
	body, err := buildFeatureFlagExposureSchedule(schedule, nil, map[string]uuid.UUID{})
	if err != nil {
		t.Fatalf("build schedule: %v", err)
	}
	if body.RolloutOptions.Autostart.IsSet() {
		t.Fatal("autostart was sent even though it was unknown")
	}
	if body.AbsoluteStartTime.IsSet() {
		t.Fatal("absolute_start_time was sent even though it was unknown")
	}
	if len(body.RolloutSteps) != 1 || body.RolloutSteps[0].GroupedStepIndex != 0 {
		t.Fatalf("rollout steps = %+v", body.RolloutSteps)
	}
}

func TestValidateArchivedFeatureFlagAdoption(t *testing.T) {
	t.Parallel()
	archived := datadogV2.FeatureFlagListItem{Attributes: datadogV2.FeatureFlagListItemAttributes{
		ValueType: datadogV2.VALUETYPE_BOOLEAN,
	}}
	plan := featureFlagResourceModel{
		ValueType:         types.StringValue("BOOLEAN"),
		RequireApproval:   types.BoolUnknown(),
		DefaultVariantKey: types.StringUnknown(),
	}
	if err := validateArchivedFeatureFlagAdoption(archived, plan); err != nil {
		t.Fatalf("adoption with computed fields rejected: %v", err)
	}
	plan.ValueType = types.StringValue("STRING")
	if err := validateArchivedFeatureFlagAdoption(archived, plan); err == nil {
		t.Fatal("value_type mismatch unexpectedly accepted")
	}
}
