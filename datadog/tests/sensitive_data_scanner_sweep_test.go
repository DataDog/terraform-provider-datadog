package test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

var sdsSweepOnce sync.Once

// sdsGroupMinAge is the age a timestamped group must reach before the sweeper
// deletes it. The scheduled and PR integration workflows share the org, so a
// younger group can belong to a run in progress. It equals the testacc timeout.
const sdsGroupMinAge = 2 * time.Hour

// sdsGroupNameTimestamp captures the trailing unix timestamp that
// withUniqueSurrounding ("tf-<TestName>-<buildID>-<unix>") and the
// datadog-api-client-typescript BDD suite ("Test-Typescript-<Scenario>-<unix>")
// append to group names.
var sdsGroupNameTimestamp = regexp.MustCompile(`-(\d{10})$`)

// cleanupSensitiveDataScannerGroups removes stale SDS groups left behind by
// crashed or failed test runs. It uses sync.Once so that within a single test
// binary execution the cleanup runs at most once regardless of how many tests
// call it.
//
// Call this at the top of every SDS test function, before resource.Test().
func cleanupSensitiveDataScannerGroups(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	sdsSweepOnce.Do(func() {
		doSweepSensitiveDataScannerGroups(t)
	})
}

// TestSweepSensitiveDataScannerGroups is a standalone sweep test for CI / manual
// invocation via `go test -run TestSweep` or `make sweep`.
// In CI it also runs as a dedicated step before integration tests (see test_integration.yml).
func TestSweepSensitiveDataScannerGroups(t *testing.T) {
	doSweepSensitiveDataScannerGroups(t)
}

func doSweepSensitiveDataScannerGroups(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	api := datadogV2.NewSensitiveDataScannerApi(client)

	resp, _, err := api.ListScanningGroups(ctx)
	if err != nil {
		t.Logf("SDS sweep: failed to list scanning groups: %v", err)
		return
	}

	now := time.Now()
	total := 0
	var kept []string
	for _, item := range resp.GetIncluded() {
		group := item.SensitiveDataScannerGroupIncludedItem
		if group == nil {
			continue
		}
		total++
		name := group.Attributes.GetName()
		id := group.GetId()

		if !isTestGroup(name, now) {
			kept = append(kept, name)
			continue
		}

		t.Logf("SDS sweep: deleting stale group %q (id=%s)", name, id)

		body := datadogV2.NewSensitiveDataScannerGroupDeleteRequestWithDefaults()
		meta := datadogV2.NewSensitiveDataScannerMetaVersionOnlyWithDefaults()
		body.SetMeta(*meta)

		_, httpResp, err := api.DeleteScanningGroup(ctx, id, *body)
		if err != nil {
			status := 0
			if httpResp != nil {
				status = httpResp.StatusCode
			}
			t.Logf("SDS sweep: failed to delete group %q (id=%s, status=%d): %v", name, id, status, err)
		}
	}

	// The org limit is 20 groups. Kept names identify the creator of groups
	// that the sweeper does not match.
	t.Logf("SDS sweep: %d groups in org before sweep (limit 20), %d kept: %q", total, len(kept), kept)
}

// isTestGroup reports whether the sweeper deletes a group with this name at
// time now.
func isTestGroup(name string, now time.Time) bool {
	// Hardcoded group names used in SDS rule tests (current and legacy)
	switch name {
	case "my group", "another group", "my-test-group":
		return true
	}

	isTF := strings.HasPrefix(name, "tf-")
	// The datadog-api-client-typescript BDD suite runs against the same org
	// and leaks groups with this prefix.
	isTypescriptBDD := strings.HasPrefix(name, "Test-Typescript-")
	if !isTF && !isTypescriptBDD {
		return false
	}

	m := sdsGroupNameTimestamp.FindStringSubmatch(name)
	if m == nil {
		// Age unknown: keep the previous unconditional deletion for tf- names,
		// keep foreign names.
		return isTF
	}
	ts, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return isTF
	}
	return now.Sub(time.Unix(ts, 0)) >= sdsGroupMinAge
}

func TestSensitiveDataScannerSweepIsTestGroup(t *testing.T) {
	now := time.Unix(1761130000, 0)
	old := now.Add(-sdsGroupMinAge).Unix()
	recent := now.Add(-sdsGroupMinAge + time.Minute).Unix()
	future := now.Add(time.Hour).Unix()

	cases := []struct {
		name string
		want bool
	}{
		{"my group", true},
		{"another group", true},
		{"my-test-group", true},
		{fmt.Sprintf("tf-TestAccDatadogSDSGroupOrder_basic-local-%d", old), true},
		{fmt.Sprintf("tf-TestAccDatadogSDSGroupOrder_basic-12345-%d", recent), false},
		{fmt.Sprintf("tf-TestAccDatadogSDSGroupOrder_basic-local-%d", future), false},
		{"tf-no-timestamp", true},
		{fmt.Sprintf("Test-Typescript-Create_Scanning_Group_returns_OK_response-%d", old), true},
		{fmt.Sprintf("Test-Typescript-Create_Scanning_Group_returns_OK_response-%d", recent), false},
		{"Test-Typescript-Create_Scanning_Group_returns_OK_response", false},
		{fmt.Sprintf("Example-Create_Scanning_Group_returns_OK_response_%d", old), false},
		{fmt.Sprintf("production-%d", old), false},
		{"Test-Go-Create_Scanning_Group", false},
	}
	for _, c := range cases {
		if got := isTestGroup(c.name, now); got != c.want {
			t.Errorf("isTestGroup(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}
