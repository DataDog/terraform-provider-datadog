package test

import (
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// legacyLogsArchiveTestNames are fixed names used by earlier versions of the
// live logs archive tests. Archive names are unique per org, so a leaked
// archive with one of these names makes the next create fail with 400.
var legacyLogsArchiveTestNames = map[string]bool{
	"my first azure archive": true,
}

// cleanupLogsArchives removes archives with legacy fixed test names. It does
// not match the tf- prefix: a concurrent live run of the same test would lose
// its in-flight archive.
func cleanupLogsArchives(t *testing.T) {
	t.Helper()
	if isReplaying() {
		return
	}
	doSweepLogsArchives(t, func(name string) bool { return legacyLogsArchiveTestNames[name] })
}

// TestSweepLogsArchives is a standalone sweep test for CI / manual invocation.
func TestSweepLogsArchives(t *testing.T) {
	doSweepLogsArchives(t, isLogsArchiveTestName)
}

func isLogsArchiveTestName(name string) bool {
	return legacyLogsArchiveTestNames[name] || strings.HasPrefix(name, "tf-TestAccDatadogLogsArchive")
}

func doSweepLogsArchives(t *testing.T, match func(string) bool) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}
	api := datadogV2.NewLogsArchivesApi(client)

	resp, _, err := api.ListLogsArchives(ctx)
	if err != nil {
		t.Logf("Logs archive sweep: list failed: %v", err)
		return
	}
	for _, item := range resp.GetData() {
		name := item.Attributes.GetName()
		if !match(name) {
			continue
		}
		id := item.GetId()
		t.Logf("Logs archive sweep: deleting stale archive %q (id=%s)", name, id)
		if _, err := api.DeleteLogsArchive(ctx, id); err != nil {
			t.Logf("Logs archive sweep: delete %q (id=%s) failed: %v", name, id, err)
		}
	}
}
