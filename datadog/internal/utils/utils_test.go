package utils

import (
	"fmt"
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

func validJSON() string {
	return fmt.Sprint(`
{
   "test":"value",
   "test_two":{
      "nested_attr":"value"
   }
}
`)
}
func invalidJSON() string {
	return fmt.Sprint(`
{
   "test":"value":"value",
   "test_two":{
      "nested_attr":"value"
   }
}
`)
}
