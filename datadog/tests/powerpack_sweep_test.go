package test

import (
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

// cleanupPowerpacks removes stale powerpacks left behind by failed test runs.
// Call this at the top of powerpack test functions, before resource.Test().
func cleanupPowerpacks(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	doSweepPowerpacks(t)
}

// TestSweepPowerpacks is a standalone sweep test for CI / manual invocation.
func TestSweepPowerpacks(t *testing.T) {
	doSweepPowerpacks(t)
}

func doSweepPowerpacks(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	api := datadogV2.NewPowerpackApi(client)

	// Use small page sizes and retry to handle 504 timeouts on large orgs
	var offset int64
	pageSize := int64(5)
	for {
		opts := datadogV2.NewListPowerpacksOptionalParameters().
			WithPageLimit(pageSize).
			WithPageOffset(offset)

		var data []datadogV2.PowerpackData
		var listErr error
		for attempt := 0; attempt < 3; attempt++ {
			resp, _, err := api.ListPowerpacks(ctx, *opts)
			if err != nil {
				listErr = err
				t.Logf("Powerpack sweep: list attempt %d failed (offset=%d): %v", attempt+1, offset, err)
				time.Sleep(5 * time.Second)
				continue
			}
			data = resp.GetData()
			listErr = nil
			break
		}
		if listErr != nil {
			t.Logf("Powerpack sweep: giving up listing after 3 attempts (offset=%d)", offset)
			return
		}

		if len(data) == 0 {
			break
		}

		for _, item := range data {
			name := item.Attributes.GetName()
			id := item.GetId()

			if !strings.HasPrefix(name, "tf-") {
				offset++
				continue
			}

			t.Logf("Powerpack sweep: deleting stale powerpack %q (id=%s)", name, id)

			_, err := api.DeletePowerpack(ctx, id)
			if err != nil {
				t.Logf("Powerpack sweep: failed to delete powerpack %q (id=%s): %v", name, id, err)
				offset++
			}
		}

		if int64(len(data)) < pageSize {
			break
		}
	}
}
