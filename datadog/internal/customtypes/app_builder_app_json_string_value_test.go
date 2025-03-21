// reference: https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/v0.2.0/jsontypes/normalized_value_test.go

package customtypes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// these tests are similar to the NormalizedValue tests
func TestAppBuilderAppStringValueSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentJson   AppBuilderAppStringValue
		givenJson     basetypes.StringValuable
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"not equal - mismatched field values": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "dlrow", "queries": [3, 2, 1], "components": {"test-bool": false}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - mismatched field names": {
			currentJson:   NewAppBuilderAppStringValue(`{"Name": "world", "Queries": [1, 2, 3], "Components": {"Test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - object additional field": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "description": null}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - array additional field": {
			currentJson:   NewAppBuilderAppStringValue(`[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true, "description": null}}]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"not equal - array item order difference": {
			currentJson:   NewAppBuilderAppStringValue(`[{"queries":[1, 2, 3]}, {"name": "world"}, {"components": {"test-bool": true}}]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"semantically equal - object byte-for-byte match": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - array byte-for-byte match": {
			currentJson:   NewAppBuilderAppStringValue(`[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"semantically equal - object field order difference": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"queries": [1, 2, 3], "components": {"test-bool": true}, "name": "world"}`),
			expectedMatch: true,
		},
		"semantically equal - object whitespace difference": {
			currentJson: NewAppBuilderAppStringValue(`{
				"name": "world",
				"queries": [1, 2, 3],
				"components": {
					"test-bool": true
				}
			}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name":"world","queries":[1,2,3],"components":{"test-bool":true}}`),
			expectedMatch: true,
		},
		"semantically equal - array whitespace difference": {
			currentJson: NewAppBuilderAppStringValue(`[
				{
				  "name": "world"
				},
				{
				  "queries": [
					1,
					2,
					3
				  ]
				},
				{
				  "components": {
					"test-bool": true
				  }
				}
			  ]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"error - invalid json": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`&#$^"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected error occurred while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Error: invalid character '&' looking for beginning of value",
				),
			},
		},
		"error - not given normalized json value": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			givenJson:     basetypes.NewStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: customtypes.AppBuilderAppStringValue\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
		// JSON Semantic equality uses (decoder).UseNumber to avoid Go parsing JSON numbers into float64. This ensures that Go
		// won't normalize the JSON number representation or impose limits on numeric range.
		"not equal - different JSON number representations": {
			currentJson:   NewAppBuilderAppStringValue(`{"rootInstanceName": 12423434}`),
			givenJson:     NewAppBuilderAppStringValue(`{"rootInstanceName": 1.2423434e+07}`),
			expectedMatch: false,
		},
		"semantically equal - larger than max float64 values": {
			currentJson:   NewAppBuilderAppStringValue(`{"rootInstanceName": 1.79769313486231570814527423731704356798070e+309}`),
			givenJson:     NewAppBuilderAppStringValue(`{"rootInstanceName": 1.79769313486231570814527423731704356798070e+309}`),
			expectedMatch: true,
		},
		// JSON Semantic equality uses Go's encoding/json library, which replaces some characters to escape codes
		"semantically equal - HTML escape characters are equal": {
			currentJson:   NewAppBuilderAppStringValue(`{"description": "http://example.com?foo=bar&name=world", "left-caret": "<", "right-caret": ">"}`),
			givenJson:     NewAppBuilderAppStringValue(`{"description": "http://example.com?foo=bar\u0026name=world", "left-caret": "\u003c", "right-caret": "\u003e"}`),
			expectedMatch: true,
		},
		// additional tests specific to AppBuilderAppStringValue below
		"semantically equal - existence of id field is ignored": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - differences in id field is ignored": {
			currentJson:   NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`),
			givenJson:     NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555554"}`),
			expectedMatch: true,
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := testCase.currentJson.StringSemanticEquals(context.Background(), testCase.givenJson)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}

func TestAppBuilderAppStringValueUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		json          AppBuilderAppStringValue
		target        any
		output        any
		expectedDiags diag.Diagnostics
	}{
		"AppBuilderAppStringValue value is null ": {
			json: NewAppBuilderAppStringValueNull(),
			target: struct {
				Name string `json:"name"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppStringValue Unmarshal Error",
					"json string value is null",
				),
			},
		},
		"AppBuilderAppStringValue value is unknown ": {
			json: NewAppBuilderAppStringValueUnknown(),
			target: struct {
				Name string `json:"name"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppStringValue Unmarshal Error",
					"json string value is unknown",
				),
			},
		},
		"invalid target - not a pointer ": {
			json: NewAppBuilderAppStringValue(`{"name": "world"}`),
			target: struct {
				Name string `json:"name"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppStringValue Unmarshal Error",
					"json: Unmarshal(non-pointer struct { Name string \"json:\\\"name\\\"\" })",
				),
			},
		},
		"valid target ": {
			json: NewAppBuilderAppStringValue(`{"name": "world", "queries": [1, 2, 3], "test-bool": true}`),
			target: &struct {
				Name    string `json:"name"`
				Numbers []int  `json:"queries"`
				Test    bool   `json:"test-bool"`
			}{},
			output: &struct {
				Name    string `json:"name"`
				Numbers []int  `json:"queries"`
				Test    bool   `json:"test-bool"`
			}{
				Name:    "world",
				Numbers: []int{1, 2, 3},
				Test:    true,
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.json.Unmarshal(testCase.target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
