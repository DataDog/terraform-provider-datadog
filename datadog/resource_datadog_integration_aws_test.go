package datadog

import (
	"testing"
)

func TestAccountAndRoleFromID(t *testing.T) {
	cases := map[string]struct {
		id        string
		accountID string
		roleName  string
	}{
		"basic":       {"1234_qwe", "1234", "qwe"},
		"underscores": {"1234_qwe_rty_asd", "1234", "qwe_rty_asd"},
	}
	for name, tc := range cases {
		accountID, roleName := accountAndRoleFromID(tc.id)
		if accountID != tc.accountID {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountID, tc.accountID)
		}
		if roleName != tc.roleName {
			t.Errorf("%s: role name '%s' didn't match `%s`", name, roleName, tc.roleName)
		}
	}
}
