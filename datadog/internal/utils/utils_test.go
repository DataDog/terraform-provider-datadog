package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestAccountAndLambdaArnFromID(t *testing.T) {
	cases := map[string]struct {
		id        string
		accountID string
		lambdaArn string
		err       error
	}{
		"basic":               {"123456789 /aws/lambda/my-fun-funct", "123456789", "/aws/lambda/my-fun-funct", nil},
		"no delimiter":        {"123456789", "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: 123456789")},
		"multiple delimiters": {"123456789 extra bits", "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: 123456789 extra bits")},
	}
	for name, tc := range cases {
		accountID, lambdaArn, err := AccountAndLambdaArnFromID(tc.id)

		if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("%s: errors should be '%s', not `%s`", name, tc.err.Error(), err.Error())
		} else if err != nil && tc.err == nil {
			t.Errorf("%s: errors should be nil, not `%s`", name, err.Error())
		} else if err == nil && tc.err != nil {
			t.Errorf("%s: errors should be '%s', not nil", name, tc.err.Error())
		}

		if accountID != tc.accountID {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountID, tc.accountID)
		}
		if lambdaArn != tc.lambdaArn {
			t.Errorf("%s: lambda arn '%s' didn't match `%s`", name, lambdaArn, tc.lambdaArn)
		}
	}
}

func TestAccountAndRoleFromID(t *testing.T) {
	cases := map[string]struct {
		id        string
		accountID string
		roleName  string
		err       error
	}{
		"basic":        {"1234:qwe", "1234", "qwe", nil},
		"underscores":  {"1234:qwe_rty_asd", "1234", "qwe_rty_asd", nil},
		"no delimeter": {"1234", "", "", fmt.Errorf("error extracting account ID and Role name from an Amazon Web Services integration id: 1234")},
	}
	for name, tc := range cases {
		accountID, roleName, err := AccountAndRoleFromID(tc.id)

		if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("%s: errors should be '%s', not `%s`", name, tc.err.Error(), err.Error())
		} else if err != nil && tc.err == nil {
			t.Errorf("%s: errors should be nil, not `%s`", name, err.Error())
		} else if err == nil && tc.err != nil {
			t.Errorf("%s: errors should be '%s', not nil", name, tc.err.Error())
		}

		if accountID != tc.accountID {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountID, tc.accountID)
		}
		if roleName != tc.roleName {
			t.Errorf("%s: role name '%s' didn't match `%s`", name, roleName, tc.roleName)
		}
	}
}

func TestAccountNameAndChannelNameFromID(t *testing.T) {
	cases := map[string]struct {
		id          string
		accountName string
		channelName string
		err         error
	}{
		"basic":        {"test-account:#channel", "test-account", "#channel", nil},
		"no delimeter": {"test-account", "", "", fmt.Errorf("error extracting account name and channel name: test-account")},
	}
	for name, tc := range cases {
		accountID, roleName, err := AccountNameAndChannelNameFromID(tc.id)

		if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("%s: errors should be '%s', not `%s`", name, tc.err.Error(), err.Error())
		} else if err != nil && tc.err == nil {
			t.Errorf("%s: errors should be nil, not `%s`", name, err.Error())
		} else if err == nil && tc.err != nil {
			t.Errorf("%s: errors should be '%s', not nil", name, tc.err.Error())
		}

		if accountID != tc.accountName {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountID, tc.accountName)
		}
		if roleName != tc.channelName {
			t.Errorf("%s: role name '%s' didn't match `%s`", name, roleName, tc.channelName)
		}
	}
}

func TestConvertResponseByteToMap(t *testing.T) {
	cases := map[string]struct {
		js     string
		errMsg string
	}{
		"validJSON":   {validJSON(), ""},
		"invalidJSON": {invalidJSON(), "invalid character ':' after object key:value pair"},
	}
	for name, tc := range cases {
		_, err := ConvertResponseByteToMap([]byte(tc.js))
		if err != nil && tc.errMsg != "" && err.Error() != tc.errMsg {
			t.Fatalf("%s: error should be %s, not %s", name, tc.errMsg, err.Error())
		}
		if err != nil && tc.errMsg == "" {
			t.Fatalf("%s: error should be nil, not %s", name, err.Error())
		}
	}
}

func TestDeleteKeyInMap(t *testing.T) {
	cases := map[string]struct {
		keyPath   []string
		mapObject map[string]interface{}
		expected  map[string]interface{}
	}{
		"basic":  {[]string{"test"}, map[string]interface{}{"test": true, "field-two": false}, map[string]interface{}{"field-two": false}},
		"nested": {[]string{"test", "nested"}, map[string]interface{}{"test": map[string]interface{}{"nested": "field"}, "field-two": false}, map[string]interface{}{"test": map[string]interface{}{}, "field-two": false}},
	}
	for name, tc := range cases {
		DeleteKeyInMap(tc.mapObject, tc.keyPath)
		if !reflect.DeepEqual(tc.mapObject, tc.expected) {
			t.Fatalf("%s: expected %s, but got %s", name, tc.expected, tc.mapObject)
		}
	}
}

func TestGetStringSlice(t *testing.T) {
	cases := []struct {
		testCase string
		resource map[string]interface{}
		expected []string
	}{
		{"emptyMap", map[string]interface{}{}, []string{}},
		{"emptyArrayValue", map[string]interface{}{"key": []interface{}{}}, []string{}},
		{"nonEmptyarrayValue", map[string]interface{}{"key": []interface{}{"value1", "value2"}}, []string{"value1", "value2"}},
		{"otherKeyInMap", map[string]interface{}{"otherKey": []interface{}{"value1"}}, []string{}},
	}

	for _, tc := range cases {
		actual := GetStringSlice(&mockResourceData{tc.resource}, "key")
		if !reflect.DeepEqual(actual, tc.expected) {
			t.Errorf("%s: expected %s, but got %s", tc.testCase, tc.expected, actual)
		}
	}
}

func validJSON() string {
	return `
{
   "test":"value",
   "test_two":{
      "nested_attr":"value"
   }
}
`
}
func invalidJSON() string {
	return `
{
   "test":"value":"value",
   "test_two":{
      "nested_attr":"value"
   }
}
`
}

type mockResourceData struct {
	values map[string]interface{}
}

func (r *mockResourceData) Get(key string) interface{} {
	return r.values[key]
}

func (r *mockResourceData) GetOk(key string) (interface{}, bool) {
	v, ok := r.values[key]
	return v, ok
}

func (r *mockResourceData) GetRawConfigAt(cty.Path) (cty.Value, sdkdiag.Diagnostics) {
	return cty.NilVal, nil
}

var testMetricNames = map[string]string{
	// bad metric names, need remapping
	"test*&(*._-_Metrictastic*(*)(  wtf_who_doesthis??": "test.Metrictastic_wtf_who_doesthis",
	"?does.this.work?":                        "does.this.work",
	"5-2 arsenal over spurs":                  "arsenal_over_spurs",
	"dd.crawler.amazon web services.run_time": "dd.crawler.amazon_web_services.run_time",

	//  multiple metric names that normalize to the same thing
	"multiple-norm-1": "multiple_norm_1",
	"multiple_norm-1": "multiple_norm_1",

	// for whatever reason, invalid characters x
	"a$.b":            "a.b",
	"a_.b":            "a.b",
	"__init__.metric": "init.metric",
	"a___..b":         "a..b",
	"a_.":             "a.",
}

func TestNormMetricNameParse(t *testing.T) {
	for src, target := range testMetricNames {
		normed := NormMetricNameParse(src)
		if normed != target {
			t.Errorf("Expected tag '%s' normalized to '%s', got '%s' instead.", src, target, normed)
			return
		}
		// double check that we're idempotent
		again := NormMetricNameParse(normed)
		if again != normed {
			t.Errorf("Expected tag '%s' to be idempotent', got '%s' instead.", normed, again)
			return
		}
	}
}

func TestRemoveEmptyValuesInMap(t *testing.T) {
	cases := map[string]struct {
		input    map[string]any
		expected map[string]any
	}{
		"empty_map": {
			input:    map[string]any{},
			expected: map[string]any{},
		},
		"nil_values": {
			input: map[string]any{
				"keep":    "value",
				"remove":  nil,
				"number":  42,
				"remove2": nil,
			},
			expected: map[string]any{
				"keep":   "value",
				"number": 42,
			},
		},
		"empty_arrays": {
			input: map[string]any{
				"keep":      []any{"value"},
				"remove":    []any{},
				"keep_nums": []any{1, 2, 3},
			},
			expected: map[string]any{
				"keep":      []any{"value"},
				"keep_nums": []any{1, 2, 3},
			},
		},
		"nested_maps": {
			input: map[string]any{
				"keep": map[string]any{
					"nested_keep": "value",
					"nested_nil":  nil,
				},
				"remove": map[string]any{},
				"deep": map[string]any{
					"deeper": map[string]any{
						"empty_array": []any{},
						"keep_this":   "value",
					},
				},
			},
			expected: map[string]any{
				"keep": map[string]any{
					"nested_keep": "value",
				},
				"deep": map[string]any{
					"deeper": map[string]any{
						"keep_this": "value",
					},
				},
			},
		},
		"array_with_maps": {
			input: map[string]any{
				"items": []any{
					map[string]any{
						"keep":   "value",
						"remove": nil,
					},
					map[string]any{
						"empty_map": map[string]any{},
						"keep":      "value2",
					},
				},
			},
			expected: map[string]any{
				"items": []any{
					map[string]any{
						"keep": "value",
					},
					map[string]any{
						"keep": "value2",
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Create a copy of the input to avoid modifying the test case
			input := make(map[string]any)
			for k, v := range tc.input {
				input[k] = v
			}

			// Run the function
			RemoveEmptyValuesInMap(input)

			// Compare the result
			if !reflect.DeepEqual(input, tc.expected) {
				t.Errorf("%s: expected %#v, but got %#v", name, tc.expected, input)
			}
		})
	}
}

// reference: https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/v0.2.0/jsontypes/normalized_value_test.go
func TestAppBuilderAppJSONStringSemanticEquals(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		currentJson   string
		givenJson     string
		expectedMatch bool
		expectedDiags diag.Diagnostics
	}{
		"not equal - mismatched field values": {
			currentJson:   `{"name": "dlrow", "queries": [3, 2, 1], "components": {"test-bool": false}}`,
			givenJson:     `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			expectedMatch: false,
		},
		"not equal - mismatched field names": {
			currentJson:   `{"Name": "world", "Queries": [1, 2, 3], "Components": {"Test-bool": true}}`,
			givenJson:     `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			expectedMatch: false,
		},
		"not equal - object additional field": {
			currentJson:   `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "description": null}`,
			givenJson:     `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			expectedMatch: false,
		},
		"not equal - array additional field": {
			currentJson:   `[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true, "description": null}}]`,
			givenJson:     `[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`,
			expectedMatch: false,
		},
		"not equal - array item order difference": {
			currentJson:   `[{"queries":[1, 2, 3]}, {"name": "world"}, {"components": {"test-bool": true}}]`,
			givenJson:     `[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`,
			expectedMatch: false,
		},
		"semantically equal - object byte-for-byte match": {
			currentJson:   `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			givenJson:     `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			expectedMatch: true,
		},
		"semantically equal - array byte-for-byte match": {
			currentJson:   `[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`,
			givenJson:     `[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`,
			expectedMatch: true,
		},
		"semantically equal - object field order difference": {
			currentJson:   `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			givenJson:     `{"queries": [1, 2, 3], "components": {"test-bool": true}, "name": "world"}`,
			expectedMatch: true,
		},
		"semantically equal - object whitespace difference": {
			currentJson: `{
				"name": "world",
				"queries": [1, 2, 3],
				"components": {
					"test-bool": true
				}
			}`,
			givenJson:     `{"name":"world","queries":[1,2,3],"components":{"test-bool":true}}`,
			expectedMatch: true,
		},
		"semantically equal - array whitespace difference": {
			currentJson: `[
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
			  ]`,
			givenJson:     `[{"name": "world"}, {"queries":[1, 2, 3]}, {"components": {"test-bool": true}}]`,
			expectedMatch: true,
		},
		"error - invalid json": {
			currentJson:   `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			givenJson:     `&#$^"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
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
		// JSON Semantic equality uses (decoder).UseNumber to avoid Go parsing JSON numbers into float64. This ensures that Go
		// won't normalize the JSON number representation or impose limits on numeric range.
		"not equal - different JSON number representations": {
			currentJson:   `{"rootInstanceName": 12423434}`,
			givenJson:     `{"rootInstanceName": 1.2423434e+07}`,
			expectedMatch: false,
		},
		"semantically equal - larger than max float64 values": {
			currentJson:   `{"rootInstanceName": 1.79769313486231570814527423731704356798070e+309}`,
			givenJson:     `{"rootInstanceName": 1.79769313486231570814527423731704356798070e+309}`,
			expectedMatch: true,
		},
		// JSON Semantic equality uses Go's encoding/json library, which replaces some characters to escape codes
		"semantically equal - HTML escape characters are equal": {
			currentJson:   `{"description": "http://example.com?foo=bar&name=world", "left-caret": "<", "right-caret": ">"}`,
			givenJson:     `{"description": "http://example.com?foo=bar\u0026name=world", "left-caret": "\u003c", "right-caret": "\u003e"}`,
			expectedMatch: true,
		},
		// additional tests specific to AppBuilderAppStringValue below
		"semantically equal - existence of id field is ignored": {
			currentJson:   `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`,
			givenJson:     `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}}`,
			expectedMatch: true,
		},
		"semantically equal - differences in id field is ignored": {
			currentJson:   `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555555"}`,
			givenJson:     `{"name": "world", "queries": [1, 2, 3], "components": {"test-bool": true}, "id": "11111111-2222-3333-4444-555555555554"}`,
			expectedMatch: true,
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			match, diags := AppJSONStringSemanticEquals(testCase.currentJson, testCase.givenJson)

			if testCase.expectedMatch != match {
				t.Errorf("Expected StringSemanticEquals to return: %t, but got: %t", testCase.expectedMatch, match)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("Unexpected diagnostics (-got, +expected): %s", diff)
			}
		})
	}
}
