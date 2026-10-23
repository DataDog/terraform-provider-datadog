package test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// cleanupTagIndexingRules removes stale tag indexing rules left behind by failed
// or interrupted test runs. Call this at the top of tag indexing rule test
// functions, before resource.Test(). It is a no-op on replay so it never runs
// against a cassette or affects recorded interactions.
func cleanupTagIndexingRules(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	doSweepTagIndexingRules(t)
}

// TestSweepTagIndexingRules is a standalone sweep test for CI / manual invocation
// via `go test -run TestSweep` or `make sweep`.
func TestSweepTagIndexingRules(t *testing.T) {
	doSweepTagIndexingRules(t)
}

// Rules leaked into the shared test org by the datadog-api-client BDD suites. They are
// deleted only after foreignTagIndexingRuleMinAge, so that a test that is still running
// keeps its rules. tf-* rules have no age guard: datadog_tag_indexing_rule_order requires
// rule_ids to be the complete set of active rules, so any rule left by an interrupted
// provider run, however recent, fails that test.
var foreignTagIndexingRuleTimestampName = regexp.MustCompile(`^TestTypescript\w*?(\d{10})$`)

const (
	foreignTagIndexingRuleFixedName = "my-indexing-rule"
	foreignTagIndexingRuleMinAge    = 2 * time.Hour
)

func shouldSweepTagIndexingRule(attrs datadogV2.TagIndexingRuleAttributes, now time.Time) bool {
	name := attrs.GetName()
	if strings.HasPrefix(name, "tf-") {
		return true
	}

	var createdAt time.Time
	if v, ok := attrs.GetCreatedAtOk(); ok {
		createdAt = *v
	}
	if m := foreignTagIndexingRuleTimestampName.FindStringSubmatch(name); m != nil {
		if createdAt.IsZero() {
			sec, err := strconv.ParseInt(m[1], 10, 64)
			if err != nil {
				return false
			}
			createdAt = time.Unix(sec, 0)
		}
	} else if name != foreignTagIndexingRuleFixedName {
		return false
	}

	return !createdAt.IsZero() && now.Sub(createdAt) >= foreignTagIndexingRuleMinAge
}

func doSweepTagIndexingRules(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	// Tag indexing rule endpoints are unstable operations; the shared sweep client doesn't enable
	// them, so opt in here (mirrors the provider's framework_provider.go configuration).
	client.GetConfig().SetUnstableOperationEnabled("v2.ListTagIndexingRules", true)
	client.GetConfig().SetUnstableOperationEnabled("v2.DeleteTagIndexingRule", true)

	api := datadogV2.NewMetricsApi(client)

	resp, _, err := api.ListTagIndexingRules(ctx)
	if err != nil {
		t.Logf("Tag indexing rule sweep: failed to list rules: %v", err)
		return
	}

	now := time.Now()
	for _, rule := range resp.GetData() {
		attrs := rule.GetAttributes()
		name := attrs.GetName()
		id := rule.GetId()

		if !shouldSweepTagIndexingRule(attrs, now) {
			continue
		}

		t.Logf("Tag indexing rule sweep: deleting stale rule %q (id=%s)", name, id)

		if _, err := api.DeleteTagIndexingRule(ctx, id); err != nil {
			t.Logf("Tag indexing rule sweep: failed to delete rule %q (id=%s): %v", name, id, err)
		}
	}
}

func TestShouldSweepTagIndexingRule(t *testing.T) {
	now := time.Unix(1790000000, 0)
	old := now.Add(-3 * time.Hour)
	recent := now.Add(-time.Minute)
	cases := []struct {
		name      string
		createdAt *time.Time
		want      bool
	}{
		{"tf-test-rule", &recent, true},
		{"my-indexing-rule", &old, true},
		{"my-indexing-rule", &recent, false},
		{"my-indexing-rule", nil, false},
		{"TestTypescriptUpdateatagindexingrulereturnsOKresponse1787930239", &old, true},
		{"TestTypescriptUpdateatagindexingrulereturnsOKresponse1787930239", &recent, false},
		{"TestTypescriptUpdateatagindexingrulereturnsOKresponse1787930239", nil, true},
		{"TestTypescriptCreateRule" + strconv.FormatInt(recent.Unix(), 10), nil, false},
		{"TestTypescriptNoTimestamp", &old, false},
		{"production-rule", &old, false},
	}
	for _, c := range cases {
		attrs := datadogV2.TagIndexingRuleAttributes{}
		attrs.SetName(c.name)
		if c.createdAt != nil {
			attrs.SetCreatedAt(*c.createdAt)
		}
		if got := shouldSweepTagIndexingRule(attrs, now); got != c.want {
			t.Errorf("shouldSweepTagIndexingRule(%q, createdAt=%v) = %v, want %v", c.name, c.createdAt, got, c.want)
		}
	}
}
