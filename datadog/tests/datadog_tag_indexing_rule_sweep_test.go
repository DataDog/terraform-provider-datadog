package test

import (
	"strings"
	"testing"

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

	for _, rule := range resp.GetData() {
		attrs := rule.GetAttributes()
		name := attrs.GetName()
		id := rule.GetId()

		if !strings.HasPrefix(name, "tf-") {
			continue
		}

		t.Logf("Tag indexing rule sweep: deleting stale rule %q (id=%s)", name, id)

		if _, err := api.DeleteTagIndexingRule(ctx, id); err != nil {
			t.Logf("Tag indexing rule sweep: failed to delete rule %q (id=%s): %v", name, id, err)
		}
	}
}
