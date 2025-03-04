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
func TestAppBuilderAppJSONStringValueSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentJson   AppBuilderAppJSONStringValue
		givenJson     basetypes.StringValuable
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"not equal - mismatched field values": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "dlrow", "nums": [3, 2, 1], "nested": {"test-bool": false}}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - mismatched field names": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"Hello": "world", "Nums": [1, 2, 3], "Nested": {"Test-bool": true}}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - object additional field": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "new-field": null}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - array additional field": {
			currentJson:   NewAppBuilderAppJSONStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true, "new-field": null}}]`),
			givenJson:     NewAppBuilderAppJSONStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"not equal - array item order difference": {
			currentJson:   NewAppBuilderAppJSONStringValue(`[{"nums":[1, 2, 3]}, {"hello": "world"}, {"nested": {"test-bool": true}}]`),
			givenJson:     NewAppBuilderAppJSONStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"semantically equal - object byte-for-byte match": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - array byte-for-byte match": {
			currentJson:   NewAppBuilderAppJSONStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			givenJson:     NewAppBuilderAppJSONStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"semantically equal - object field order difference": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"nums": [1, 2, 3], "nested": {"test-bool": true}, "hello": "world"}`),
			expectedMatch: true,
		},
		"semantically equal - object whitespace difference": {
			currentJson: NewAppBuilderAppJSONStringValue(`{
				"hello": "world",
				"nums": [1, 2, 3],
				"nested": {
					"test-bool": true
				}
			}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello":"world","nums":[1,2,3],"nested":{"test-bool":true}}`),
			expectedMatch: true,
		},
		"semantically equal - array whitespace difference": {
			currentJson: NewAppBuilderAppJSONStringValue(`[
				{
				  "hello": "world"
				},
				{
				  "nums": [
					1,
					2,
					3
				  ]
				},
				{
				  "nested": {
					"test-bool": true
				  }
				}
			  ]`),
			givenJson:     NewAppBuilderAppJSONStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"error - invalid json": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`&#$^"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
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
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     basetypes.NewStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: AppBuilderAppJSONStringValue\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
		// JSON Semantic equality uses (decoder).UseNumber to avoid Go parsing JSON numbers into float64. This ensures that Go
		// won't normalize the JSON number representation or impose limits on numeric range.
		"not equal - different JSON number representations": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"large": 12423434}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"large": 1.2423434e+07}`),
			expectedMatch: false,
		},
		"semantically equal - larger than max float64 values": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"large": 1.79769313486231570814527423731704356798070e+309}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"large": 1.79769313486231570814527423731704356798070e+309}`),
			expectedMatch: true,
		},
		// JSON Semantic equality uses Go's encoding/json library, which replaces some characters to escape codes
		"semantically equal - HTML escape characters are equal": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"url_ampersand": "http://example.com?foo=bar&hello=world", "left-caret": "<", "right-caret": ">"}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"url_ampersand": "http://example.com?foo=bar\u0026hello=world", "left-caret": "\u003c", "right-caret": "\u003e"}`),
			expectedMatch: true,
		},
		// additional tests specific to AppBuilderAppJSONStringValue below
		"semantically equal - existence of id field is ignored": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - differences in id field is ignored": {
			currentJson:   NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`),
			givenJson:     NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555554"}`),
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

func TestAppBuilderAppJSONStringValueUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		json          AppBuilderAppJSONStringValue
		target        any
		output        any
		expectedDiags diag.Diagnostics
	}{
		"AppBuilderAppJSONStringValue value is null ": {
			json: NewAppBuilderAppJSONStringValueNull(),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppJSONStringValue Unmarshal Error",
					"json string value is null",
				),
			},
		},
		"AppBuilderAppJSONStringValue value is unknown ": {
			json: NewAppBuilderAppJSONStringValueUnknown(),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppJSONStringValue Unmarshal Error",
					"json string value is unknown",
				),
			},
		},
		"invalid target - not a pointer ": {
			json: NewAppBuilderAppJSONStringValue(`{"hello": "world"}`),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppJSONStringValue Unmarshal Error",
					"json: Unmarshal(non-pointer struct { Hello string \"json:\\\"hello\\\"\" })",
				),
			},
		},
		"valid target ": {
			json: NewAppBuilderAppJSONStringValue(`{"hello": "world", "nums": [1, 2, 3], "test-bool": true}`),
			target: &struct {
				Hello   string `json:"hello"`
				Numbers []int  `json:"nums"`
				Test    bool   `json:"test-bool"`
			}{},
			output: &struct {
				Hello   string `json:"hello"`
				Numbers []int  `json:"nums"`
				Test    bool   `json:"test-bool"`
			}{
				Hello:   "world",
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
