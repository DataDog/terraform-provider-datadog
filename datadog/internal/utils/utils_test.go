package utils

import (
	"fmt"
	"reflect"
	"testing"
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
