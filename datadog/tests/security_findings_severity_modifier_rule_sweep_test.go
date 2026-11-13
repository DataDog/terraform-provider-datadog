package test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
)

const severityModifierRuleTestPrefix = "tf-TestAccDatadogSecurityFindingsSeverityModifierRule"

// severityModifierRuleMinAge is the age a rule must reach before the sweeper
// deletes it, so that a test that is still running keeps its rules.
const severityModifierRuleMinAge = 2 * time.Hour

// severityModifierRuleTimestamp extracts the Unix time that withUniqueSurrounding
// puts after the build ID. Tests append suffixes ("-a", "-updated"), so the
// timestamp is not always at the end of the name.
var severityModifierRuleTimestamp = regexp.MustCompile(`-(?:local|\d+)-(\d{10})(?:-|$)`)

// cleanupSeverityModifierRules removes stale severity modifier rules. The reorder
// endpoint is org-global and rejects a list that omits any rule, so one leaked
// rule makes every order test fail with "reorder must include all N rules".
// Call it at the top of each order test, before resource.Test(). No-op on replay.
func cleanupSeverityModifierRules(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	doSweepSeverityModifierRules(t)
}

// TestSweepSeverityModifierRules is a standalone sweep test for CI / manual
// invocation via `make sweep`.
func TestSweepSeverityModifierRules(t *testing.T) {
	if isReplaying() {
		t.Skip("sweep deletes real rules; not run on cassette replay")
	}
	doSweepSeverityModifierRules(t)
}

func doSweepSeverityModifierRules(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	// Unstable operations; the shared sweep client does not enable them.
	client.GetConfig().SetUnstableOperationEnabled("v2.ListSecurityFindingsAutomationSeverityModifierRules", true)
	client.GetConfig().SetUnstableOperationEnabled("v2.DeleteSecurityFindingsAutomationSeverityModifierRule", true)

	api := datadogV2.NewSecurityMonitoringApi(client)

	type staleRule struct {
		id   uuid.UUID
		name string
	}
	// Collect before deleting: deletions shift the pagination window.
	var toDelete []staleRule
	const pageSize int64 = 100
	// Upper bound in case the endpoint ignores page parameters.
	const maxPages = 50
	now := time.Now()

	for pageNumber := int64(0); pageNumber < maxPages; pageNumber++ {
		opts := datadogV2.NewListSecurityFindingsAutomationSeverityModifierRulesOptionalParameters().
			WithPageSize(pageSize).
			WithPageNumber(pageNumber)

		resp, _, err := api.ListSecurityFindingsAutomationSeverityModifierRules(ctx, *opts)
		if err != nil {
			t.Logf("Severity modifier rule sweep: failed to list rules (page %d): %v", pageNumber, err)
			return
		}

		data := resp.GetData()
		for _, rule := range data {
			attrs := rule.GetAttributes()
			name := attrs.GetName()
			if isStaleSeverityModifierRuleName(name, now) {
				toDelete = append(toDelete, staleRule{id: rule.GetId(), name: name})
			}
		}

		if int64(len(data)) < pageSize {
			break
		}
	}

	for _, rule := range toDelete {
		t.Logf("Severity modifier rule sweep: deleting stale rule %q (id=%s)", rule.name, rule.id)
		if _, err := api.DeleteSecurityFindingsAutomationSeverityModifierRule(ctx, rule.id); err != nil {
			t.Logf("Severity modifier rule sweep: failed to delete rule %q (id=%s): %v", rule.name, rule.id, err)
		}
	}
}

// isStaleSeverityModifierRuleName reports whether a rule was created by a test
// of this resource and is older than severityModifierRuleMinAge. A test-prefixed
// name without a parseable timestamp has no age to protect and is stale.
func isStaleSeverityModifierRuleName(name string, now time.Time) bool {
	if !strings.HasPrefix(name, severityModifierRuleTestPrefix) {
		return false
	}
	m := severityModifierRuleTimestamp.FindStringSubmatch(name)
	if m == nil {
		return true
	}
	ts, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return true
	}
	return now.Sub(time.Unix(ts, 0)) >= severityModifierRuleMinAge
}

func TestIsStaleSeverityModifierRuleName(t *testing.T) {
	now := time.Unix(1791100000, 0)
	old := strconv.FormatInt(now.Add(-3*time.Hour).Unix(), 10)
	recent := strconv.FormatInt(now.Add(-30*time.Minute).Unix(), 10)

	for _, tc := range []struct {
		name string
		want bool
	}{
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRule-local-" + old + "-updated", true},
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRulesOrder-local-" + old + "-a", true},
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRulesOrder_OutOfBandRuleAdded-2847193-" + old + "-out-of-band", true},
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRule-local-" + old, true},
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRule-no-timestamp", true},
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRulesOrder-local-" + recent + "-a", false},
		{"tf-TestAccDatadogSecurityFindingsSeverityModifierRule-2847193-" + recent, false},
		{"tf-TestAccDatadogSecurityFindingsMuteRule-local-" + old, false},
		{"Production severity rule", false},
		{"my-tf-TestAccDatadogSecurityFindingsSeverityModifierRule-local-" + old, false},
	} {
		if got := isStaleSeverityModifierRuleName(tc.name, now); got != tc.want {
			t.Errorf("isStaleSeverityModifierRuleName(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}
