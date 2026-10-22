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

const ticketCreationRuleTestPrefix = "tf-TestAccDatadogSecurityFindingsTicketCreationRule"

// ticketCreationRuleMinAge protects rules of a concurrent run (scheduled and PR
// integration runs share the org). Rules younger than this are kept.
const ticketCreationRuleMinAge = 2 * time.Hour

// ticketCreationRuleTimestamp extracts the Unix time that withUniqueSurrounding
// puts after the build ID. Tests append suffixes ("-a", "-updated"), so the
// timestamp is not always at the end of the name.
var ticketCreationRuleTimestamp = regexp.MustCompile(`-(?:local|\d+)-(\d{10})(?:-|$)`)

// cleanupTicketCreationRules removes stale ticket creation rules. The reorder
// endpoint is org-global and rejects a list that omits any rule, so one leaked
// rule makes every order test fail with "reorder must include all N rules".
// Call it at the top of each order test, before resource.Test(). No-op on replay.
func cleanupTicketCreationRules(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	doSweepTicketCreationRules(t)
}

// TestSweepTicketCreationRules is a standalone sweep test for CI / manual
// invocation via `make sweep`.
func TestSweepTicketCreationRules(t *testing.T) {
	if isReplaying() {
		t.Skip("sweep deletes real rules; not run on cassette replay")
	}
	doSweepTicketCreationRules(t)
}

func doSweepTicketCreationRules(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	// Unstable operations; the shared sweep client does not enable them.
	client.GetConfig().SetUnstableOperationEnabled("v2.ListSecurityFindingsAutomationTicketCreationRules", true)
	client.GetConfig().SetUnstableOperationEnabled("v2.DeleteSecurityFindingsAutomationTicketCreationRule", true)

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
		opts := datadogV2.NewListSecurityFindingsAutomationTicketCreationRulesOptionalParameters().
			WithPageSize(pageSize).
			WithPageNumber(pageNumber)

		resp, _, err := api.ListSecurityFindingsAutomationTicketCreationRules(ctx, *opts)
		if err != nil {
			t.Logf("Ticket creation rule sweep: failed to list rules (page %d): %v", pageNumber, err)
			return
		}

		data := resp.GetData()
		for _, rule := range data {
			attrs := rule.GetAttributes()
			name := attrs.GetName()
			if isStaleTicketCreationRuleName(name, now) {
				toDelete = append(toDelete, staleRule{id: rule.GetId(), name: name})
			}
		}

		if int64(len(data)) < pageSize {
			break
		}
	}

	for _, rule := range toDelete {
		t.Logf("Ticket creation rule sweep: deleting stale rule %q (id=%s)", rule.name, rule.id)
		if _, err := api.DeleteSecurityFindingsAutomationTicketCreationRule(ctx, rule.id); err != nil {
			t.Logf("Ticket creation rule sweep: failed to delete rule %q (id=%s): %v", rule.name, rule.id, err)
		}
	}
}

// isStaleTicketCreationRuleName reports whether a rule was created by a test
// of this resource and is older than ticketCreationRuleMinAge. A test-prefixed
// name without a parseable timestamp has no age to protect and is stale.
func isStaleTicketCreationRuleName(name string, now time.Time) bool {
	if !strings.HasPrefix(name, ticketCreationRuleTestPrefix) {
		return false
	}
	m := ticketCreationRuleTimestamp.FindStringSubmatch(name)
	if m == nil {
		return true
	}
	ts, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return true
	}
	return now.Sub(time.Unix(ts, 0)) >= ticketCreationRuleMinAge
}

func TestIsStaleTicketCreationRuleName(t *testing.T) {
	now := time.Unix(1791100000, 0)
	old := strconv.FormatInt(now.Add(-3*time.Hour).Unix(), 10)
	recent := strconv.FormatInt(now.Add(-30*time.Minute).Unix(), 10)

	for _, tc := range []struct {
		name string
		want bool
	}{
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRule-local-" + old + "-updated", true},
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRulesOrder-local-" + old + "-a", true},
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRulesOrder_OutOfBandRuleAdded-2847193-" + old + "-out-of-band", true},
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRule-local-" + old, true},
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRule-no-timestamp", true},
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRulesOrder-local-" + recent + "-a", false},
		{"tf-TestAccDatadogSecurityFindingsTicketCreationRule-2847193-" + recent, false},
		{"tf-TestAccDatadogSecurityFindingsMuteRule-local-" + old, false},
		{"Production Jira rule", false},
		{"my-tf-TestAccDatadogSecurityFindingsTicketCreationRule-local-" + old, false},
	} {
		if got := isStaleTicketCreationRuleName(tc.name, now); got != tc.want {
			t.Errorf("isStaleTicketCreationRuleName(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}
