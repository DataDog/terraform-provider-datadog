package datadog

import (
	"fmt"
	"testing"
)

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
		accountID, roleName, err := accountAndRoleFromID(tc.id)

		if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("%s: erros should be '%s', not `%s`", name, tc.err.Error(), err.Error())
		} else if err != nil && tc.err == nil {
			t.Errorf("%s: erros should be nil, not `%s`", name, err.Error())
		} else if err == nil && tc.err != nil {
			t.Errorf("%s: erros should be '%s', not nil", name, tc.err.Error())
		}

		if accountID != tc.accountID {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountID, tc.accountID)
		}
		if roleName != tc.roleName {
			t.Errorf("%s: role name '%s' didn't match `%s`", name, roleName, tc.roleName)
		}
	}
}
