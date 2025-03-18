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
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "dlrow", "nums": [3, 2, 1], "nested": {"test-bool": false}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - mismatched field names": {
			currentJson:   NewAppBuilderAppStringValue(`{"Hello": "world", "Nums": [1, 2, 3], "Nested": {"Test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - object additional field": {
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "new-field": null}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
		},
		"not equal - array additional field": {
			currentJson:   NewAppBuilderAppStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true, "new-field": null}}]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"not equal - array item order difference": {
			currentJson:   NewAppBuilderAppStringValue(`[{"nums":[1, 2, 3]}, {"hello": "world"}, {"nested": {"test-bool": true}}]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: false,
		},
		"semantically equal - object byte-for-byte match": {
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - array byte-for-byte match": {
			currentJson:   NewAppBuilderAppStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			givenJson:     NewAppBuilderAppStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"semantically equal - object field order difference": {
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`{"nums": [1, 2, 3], "nested": {"test-bool": true}, "hello": "world"}`),
			expectedMatch: true,
		},
		"semantically equal - object whitespace difference": {
			currentJson: NewAppBuilderAppStringValue(`{
				"hello": "world",
				"nums": [1, 2, 3],
				"nested": {
					"test-bool": true
				}
			}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello":"world","nums":[1,2,3],"nested":{"test-bool":true}}`),
			expectedMatch: true,
		},
		"semantically equal - array whitespace difference": {
			currentJson: NewAppBuilderAppStringValue(`[
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
			givenJson:     NewAppBuilderAppStringValue(`[{"hello": "world"}, {"nums":[1, 2, 3]}, {"nested": {"test-bool": true}}]`),
			expectedMatch: true,
		},
		"error - invalid json": {
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     NewAppBuilderAppStringValue(`&#$^"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
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
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			givenJson:     basetypes.NewStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: false,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Semantic Equality Check Error",
					"An unexpected value type was received while performing semantic equality checks. "+
						"Please report this to the provider developers.\n\n"+
						"Expected Value Type: AppBuilderAppStringValue\n"+
						"Got Value Type: basetypes.StringValue",
				),
			},
		},
		// JSON Semantic equality uses (decoder).UseNumber to avoid Go parsing JSON numbers into float64. This ensures that Go
		// won't normalize the JSON number representation or impose limits on numeric range.
		"not equal - different JSON number representations": {
			currentJson:   NewAppBuilderAppStringValue(`{"large": 12423434}`),
			givenJson:     NewAppBuilderAppStringValue(`{"large": 1.2423434e+07}`),
			expectedMatch: false,
		},
		"semantically equal - larger than max float64 values": {
			currentJson:   NewAppBuilderAppStringValue(`{"large": 1.79769313486231570814527423731704356798070e+309}`),
			givenJson:     NewAppBuilderAppStringValue(`{"large": 1.79769313486231570814527423731704356798070e+309}`),
			expectedMatch: true,
		},
		// JSON Semantic equality uses Go's encoding/json library, which replaces some characters to escape codes
		"semantically equal - HTML escape characters are equal": {
			currentJson:   NewAppBuilderAppStringValue(`{"url_ampersand": "http://example.com?foo=bar&hello=world", "left-caret": "<", "right-caret": ">"}`),
			givenJson:     NewAppBuilderAppStringValue(`{"url_ampersand": "http://example.com?foo=bar\u0026hello=world", "left-caret": "\u003c", "right-caret": "\u003e"}`),
			expectedMatch: true,
		},
		// additional tests specific to AppBuilderAppStringValue below
		"semantically equal - existence of id field is ignored": {
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}}`),
			expectedMatch: true,
		},
		"semantically equal - differences in id field is ignored": {
			currentJson:   NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`),
			givenJson:     NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "nested": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555554"}`),
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
				Hello string `json:"hello"`
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
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppStringValue Unmarshal Error",
					"json string value is unknown",
				),
			},
		},
		"invalid target - not a pointer ": {
			json: NewAppBuilderAppStringValue(`{"hello": "world"}`),
			target: struct {
				Hello string `json:"hello"`
			}{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"AppBuilderAppStringValue Unmarshal Error",
					"json: Unmarshal(non-pointer struct { Hello string \"json:\\\"hello\\\"\" })",
				),
			},
		},
		"valid target ": {
			json: NewAppBuilderAppStringValue(`{"hello": "world", "nums": [1, 2, 3], "test-bool": true}`),
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
