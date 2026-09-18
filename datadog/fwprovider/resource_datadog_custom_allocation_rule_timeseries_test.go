package fwprovider

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// TestBasedOnTimeseriesValidator mirrors the server-side cases in dd-source
// (cost-onboarding-api arbitrarycostrule.validateTimeseriesBasedRule). The two
// implementations enforce the same rules and should be kept in sync.
//
// The two "real payload" cases are taken verbatim from production rules and must
// always pass: they are the reason based_on_timeseries is modeled as opaque JSON
// rather than a typed schema.
func TestBasedOnTimeseriesValidator(t *testing.T) {
	t.Parallel()

	const metricsPayload = `{"response_format":"timeseries","queries":[{"name":"query1","data_source":"metrics","query":"avg:system.cpu.user{*} by {host}"}],"formulas":[{"formula":"query1"}]}`
	const aggregatePayload = `{"response_format":"timeseries","queries":[{"data_source":"aggregate_augmented_query","name":"query1","base_query":{"data_source":"metrics","name":"query1","query":"sum:aip_serving_cost_reallocation{*} by {dd.team}"},"augment_query":{"data_source":"reference_table","name":"filter_query","table_name":"aip_serving_team_mapping","columns":[{"name":"team"},{"name":"dd_team"}]},"compute":[{"aggregation":"sum","name":"compute_result"}],"group_by":[{"facet":"team","source":"filter_query"}],"join_condition":{"join_type":"inner","is_negated":false,"base_attribute":"dd.team","augment_attribute":"dd_team"}}],"formulas":[{"formula":"query1"}]}`

	tests := []struct {
		name        string
		configValue basetypes.StringValue
		wantSummary string
		wantDetail  string
	}{
		// --- must pass ---
		{
			name:        "real production metrics payload",
			configValue: basetypes.NewStringValue(metricsPayload),
		},
		{
			name:        "real production aggregate_augmented_query payload",
			configValue: basetypes.NewStringValue(aggregatePayload),
		},
		{
			name:        "arithmetic formula is not resolved against query names",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"},{"name":"query2","data_source":"metrics"}],"formulas":[{"formula":"query1 / query2"}]}`),
		},
		{
			name:        "formulas may be omitted entirely",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"}]}`),
		},
		{
			name:        "response_format may be omitted",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"}],"formulas":[{"formula":"query1"}]}`),
		},
		{
			name:        "unknown data_source is accepted",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"some_future_source"}]}`),
		},
		{
			name:        "null config is skipped",
			configValue: basetypes.NewStringNull(),
		},
		{
			name:        "unknown config is skipped",
			configValue: basetypes.NewStringUnknown(),
		},

		// --- must fail ---
		{
			name:        "not valid JSON",
			configValue: basetypes.NewStringValue(`not json`),
			wantSummary: "Invalid based_on_timeseries",
		},
		{
			name:        "JSON array instead of object",
			configValue: basetypes.NewStringValue(`[]`),
			wantSummary: "Invalid based_on_timeseries",
		},
		{
			name:        "wrong response_format",
			configValue: basetypes.NewStringValue(`{"response_format":"scalar","queries":[{"name":"query1","data_source":"metrics"}]}`),
			wantSummary: "Invalid response_format",
			wantDetail:  `got "scalar"`,
		},
		{
			name:        "queries missing",
			configValue: basetypes.NewStringValue(`{"response_format":"timeseries"}`),
			wantSummary: "Missing queries",
		},
		{
			name:        "queries empty",
			configValue: basetypes.NewStringValue(`{"queries":[]}`),
			wantSummary: "Invalid queries",
		},
		{
			name:        "queries not an array",
			configValue: basetypes.NewStringValue(`{"queries":{"name":"query1"}}`),
			wantSummary: "Invalid queries",
		},
		{
			name:        "query is not an object",
			configValue: basetypes.NewStringValue(`{"queries":["query1"]}`),
			wantSummary: "Invalid query",
		},
		{
			name:        "query missing name",
			configValue: basetypes.NewStringValue(`{"queries":[{"data_source":"metrics"}]}`),
			wantSummary: "Missing query name",
		},
		{
			name:        "query missing data_source",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1"}]}`),
			wantSummary: "Missing data_source",
		},
		{
			name:        "duplicate query names",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"},{"name":"query1","data_source":"metrics"}]}`),
			wantSummary: "Duplicate query name",
			wantDetail:  `"query1"`,
		},
		{
			name:        "formulas not an array",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"}],"formulas":{"formula":"query1"}}`),
			wantSummary: "Invalid formulas",
		},
		{
			name:        "formula empty",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"}],"formulas":[{"formula":""}]}`),
			wantSummary: "Missing formula",
		},
		{
			name:        "bare identifier references undefined query",
			configValue: basetypes.NewStringValue(`{"queries":[{"name":"query1","data_source":"metrics"}],"formulas":[{"formula":"nonexistent"}]}`),
			wantSummary: "Unknown query reference",
			wantDetail:  `references query "nonexistent"`,
		},
	}

	v := basedOnTimeseriesValidator{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var resp validator.StringResponse
			v.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("based_on_timeseries"),
				ConfigValue: tc.configValue,
			}, &resp)

			if tc.wantSummary == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("expected no error, got: %v", resp.Diagnostics.Errors())
				}
				return
			}

			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected error with summary %q, got none", tc.wantSummary)
			}
			err := resp.Diagnostics.Errors()[0]
			if got := err.Summary(); got != tc.wantSummary {
				t.Fatalf("summary = %q, want %q (detail: %s)", got, tc.wantSummary, err.Detail())
			}
			if tc.wantDetail != "" && !strings.Contains(err.Detail(), tc.wantDetail) {
				t.Fatalf("detail = %q, want it to contain %q", err.Detail(), tc.wantDetail)
			}
		})
	}
}

func TestBasedOnTimeseriesRoundTrip(t *testing.T) {
	t.Parallel()

	original := map[string]interface{}{
		"response_format": "timeseries",
		"formulas":        []interface{}{map[string]interface{}{"formula": "query1"}},
		"queries": []interface{}{
			map[string]interface{}{
				"data_source": "aggregate_augmented_query",
				"name":        "query1",
				"base_query": map[string]interface{}{
					"data_source": "metrics",
					"name":        "query1",
					"query":       "sum:aip_serving_cost_reallocation{*} by {dd.team}",
				},
				"augment_query": map[string]interface{}{
					"data_source": "reference_table",
					"name":        "filter_query",
					"table_name":  "aip_serving_team_mapping",
					"columns": []interface{}{
						map[string]interface{}{"name": "team"},
						map[string]interface{}{"name": "dd_team"},
					},
				},
				"compute": []interface{}{
					map[string]interface{}{"aggregation": "sum", "name": "compute_result"},
				},
				"group_by": []interface{}{
					map[string]interface{}{"facet": "team", "source": "filter_query"},
				},
				"join_condition": map[string]interface{}{
					"join_type":         "inner",
					"is_negated":        false,
					"base_attribute":    "dd.team",
					"augment_attribute": "dd_team",
				},
			},
		},
	}

	// Expand: the JSON a user writes must reach the API client as a non-empty map.
	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var expanded map[string]interface{}
	if err := json.Unmarshal(raw, &expanded); err != nil {
		t.Fatal(err)
	}
	strategy := datadogV2.NewArbitraryCostUpsertRequestDataAttributesStrategyWithDefaults()
	strategy.SetBasedOnTimeseries(expanded)
	sent, ok := strategy.GetBasedOnTimeseriesOk()
	if !ok || len(*sent) == 0 {
		t.Fatal("expected non-empty based_on_timeseries on strategy")
	}
	if !reflect.DeepEqual(*sent, original) {
		t.Fatalf("expand altered the payload\nbefore: %#v\nafter:  %#v", original, *sent)
	}

	// Flatten: an API response carrying the same map must land in state unchanged.
	state := &datadogCustomAllocationRuleModel{Strategy: &strategyModel{}}
	diags := diag.Diagnostics{}
	resp := datadogV2.NewArbitraryRuleResponseWithDefaults()
	data := datadogV2.NewArbitraryRuleResponseDataWithDefaults()
	attrs := datadogV2.NewArbitraryRuleResponseDataAttributesWithDefaults()
	strategyResp := datadogV2.NewArbitraryRuleResponseDataAttributesStrategyWithDefaults()
	strategyResp.SetBasedOnTimeseries(original)
	attrs.SetStrategy(*strategyResp)
	data.SetAttributes(*attrs)
	resp.SetData(*data)

	r := &datadogCustomAllocationRuleResource{}
	r.updateState(context.Background(), state, resp, &diags)
	if diags.HasError() {
		t.Fatalf("updateState errors: %v", diags.Errors())
	}
	if state.Strategy.BasedOnTimeseries == nil || state.Strategy.BasedOnTimeseries.Json.IsNull() {
		t.Fatal("expected based_on_timeseries in state")
	}

	// Nothing may be dropped on the way back out. `is_negated: false` is the
	// case to watch: a bool zero value is the most likely thing to disappear.
	var flattened map[string]interface{}
	if err := json.Unmarshal([]byte(state.Strategy.BasedOnTimeseries.Json.ValueString()), &flattened); err != nil {
		t.Fatalf("state value is not valid JSON: %v", err)
	}
	if !reflect.DeepEqual(flattened, original) {
		t.Fatalf("flatten altered the payload\nbefore: %#v\nafter:  %#v", original, flattened)
	}

	joinCondition := flattened["queries"].([]interface{})[0].(map[string]interface{})["join_condition"].(map[string]interface{})
	if _, present := joinCondition["is_negated"]; !present {
		t.Fatal("is_negated was dropped from join_condition")
	}
}
