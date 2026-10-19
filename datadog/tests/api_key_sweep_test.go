package test

import (
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

var apiKeySweepOnce sync.Once

// apiKeyMinAge is how old a key must be before the sweeper considers it stale.
// Keys younger than this may belong to a run in progress.
const apiKeyMinAge = time.Hour

// providerTestKeyName matches the names withUniqueSurrounding generates:
// "tf-<TestName>-<buildID>-<unix ts>", where buildID is a CI build number or
// the literal "local". Keys that do not match are left alone.
var providerTestKeyName = regexp.MustCompile(`^tf-.+-(local|\d+)-\d{10}$`)

// sweepingIsSafe reports whether the current RECORD mode may delete real API
// keys. Replay must not reach the network, and recording runs skip the api key
// tests, so only a live run sweeps.
func sweepingIsSafe() bool {
	return !isRecording() && !isReplaying()
}

// cleanupApiKeys removes stale API keys left behind by crashed or failed test
// runs. The org caps API keys at 200, so leaked keys eventually make every api
// key test fail with "You have exceeded the maximum number of permitted API
// keys".
//
// Call this at the top of every api key test function, before t.Parallel().
func cleanupApiKeys(t *testing.T) {
	t.Helper()

	if !sweepingIsSafe() {
		return
	}

	apiKeySweepOnce.Do(func() {
		doSweepApiKeys(t)
	})
}

// TestSweepApiKeys is a standalone sweep test for CI / manual invocation via
// `go test -run TestSweep` or `make sweep`.
func TestSweepApiKeys(t *testing.T) {
	if !sweepingIsSafe() {
		t.Skip("sweeping deletes real API keys; not supported under RECORD=true or RECORD=false")
	}
	doSweepApiKeys(t)
}

func doSweepApiKeys(t *testing.T) {
	t.Helper()

	if !sweepingIsSafe() {
		t.Log("API key sweep: refusing to delete keys in a cassette mode")
		return
	}

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	api := datadogV2.NewKeyManagementApi(client)

	// Collect before deleting: deletions shift the pagination window.
	type staleKey struct {
		id   string
		name string
	}
	var toDelete []staleKey
	var totalSeen int
	const pageSize int64 = 100
	cutoff := time.Now().Add(-apiKeyMinAge)

	for pageNumber := int64(0); ; pageNumber++ {
		opts := datadogV2.NewListAPIKeysOptionalParameters().
			WithPageSize(pageSize).
			WithPageNumber(pageNumber)

		resp, _, err := api.ListAPIKeys(ctx, *opts)
		if err != nil {
			t.Logf("API key sweep: failed to list keys (page %d): %v", pageNumber, err)
			return
		}

		data := resp.GetData()
		totalSeen += len(data)

		for _, key := range data {
			attrs := key.Attributes
			if attrs == nil {
				continue
			}
			name := attrs.GetName()
			if !isTestApiKeyName(name) {
				continue
			}
			createdAt, err := time.Parse(time.RFC3339Nano, attrs.GetCreatedAt())
			if err != nil {
				t.Logf("API key sweep: skipping %q, unparseable created_at %q", name, attrs.GetCreatedAt())
				continue
			}
			if createdAt.After(cutoff) {
				continue
			}
			toDelete = append(toDelete, staleKey{id: key.GetId(), name: name})
		}

		if int64(len(data)) < pageSize {
			break
		}
	}

	t.Logf("API key sweep: found %d keys, %d stale", totalSeen, len(toDelete))

	var deleted int
	for _, key := range toDelete {
		httpResp, err := api.DeleteAPIKey(ctx, key.id)
		if err != nil {
			status := 0
			if httpResp != nil {
				status = httpResp.StatusCode
			}
			t.Logf("API key sweep: failed to delete %q (id=%s, status=%d): %v", key.name, key.id, status, err)
			continue
		}
		deleted++
	}

	t.Logf("API key sweep: deleted %d stale keys", deleted)
}

func isTestApiKeyName(name string) bool {
	return providerTestKeyName.MatchString(name)
}

func TestIsTestApiKeyName(t *testing.T) {
	// Every "keep" case below is a name seen on a real key in the test org.
	sweep := []string{
		"tf-TestAccDatadogApiKeyDatasource_matchId-local-1785463711",
		"tf-TestDatadogApiKey_import-2847193-1785463711",
		"tf-TestAccDatadogApiKey_Update-local-1785463711",
	}
	for _, name := range sweep {
		if !isTestApiKeyName(name) {
			t.Errorf("expected %q to be swept, it was kept", name)
		}
	}

	keep := []string{
		"tf-prod-deploy",
		"tf-terraform-cloud-runner",
		"terraform",
		"Test-Key",
		"Datadog Agent",
		"CI Runner 1785463711",
		"my-tf-key-local-1785463711",
		"Test-Go-Create-an-API-key-returns-OK-response-1785463711",
		"Test-Python-List-API-keys-1785463711",
		"test-go-something-1785463711",
	}
	for _, name := range keep {
		if isTestApiKeyName(name) {
			t.Errorf("expected %q to be kept, it would have been swept", name)
		}
	}
}

func TestSweepingIsSafe(t *testing.T) {
	for _, tc := range []struct {
		record string
		want   bool
	}{
		{"false", false}, // replay
		{"true", false},  // recording
		{"none", true},   // live run
		{"", true},       // unset behaves as a live run
	} {
		t.Setenv("RECORD", tc.record)
		if got := sweepingIsSafe(); got != tc.want {
			t.Errorf("RECORD=%q: sweepingIsSafe() = %v, want %v", tc.record, got, tc.want)
		}
	}
}
