package test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

var sdsSweepOnce sync.Once

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

// TestSweepSensitiveDataScannerGroups is a standalone test for CI / manual
// invocation via `go test -run TestSweep` or `make sweep`.
func TestSweepSensitiveDataScannerGroups(t *testing.T) {
	doSweepSensitiveDataScannerGroups(t)
}

func doSweepSensitiveDataScannerGroups(t *testing.T) {
	t.Helper()

	apiKey := os.Getenv(testAPIKeyEnvName)
	appKey := os.Getenv(testAPPKeyEnvName)
	apiURL := os.Getenv(testAPIUrlEnvName)

	if apiKey == "" || appKey == "" {
		t.Log("SDS sweep: DD_TEST_CLIENT_API_KEY or DD_TEST_CLIENT_APP_KEY not set, skipping cleanup")
		return
	}

	ctx, err := buildContext(context.Background(), apiKey, appKey, apiURL)
	if err != nil {
		t.Logf("SDS sweep: failed to build API context: %v", err)
		return
	}

	cfg := common.NewConfiguration()
	client := common.NewAPIClient(cfg)
	api := datadogV2.NewSensitiveDataScannerApi(client)

	resp, _, err := api.ListScanningGroups(ctx)
	if err != nil {
		t.Logf("SDS sweep: failed to list scanning groups: %v", err)
		return
	}

	for _, item := range resp.GetIncluded() {
		group := item.SensitiveDataScannerGroupIncludedItem
		if group == nil {
			continue
		}
		name := group.Attributes.GetName()
		id := group.GetId()

		if !isTestGroup(name) {
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
}

// isTestGroup returns true if the group name looks like it was created by tests.
func isTestGroup(name string) bool {
	// Groups created via uniqueEntityName() always start with "tf-"
	if strings.HasPrefix(name, "tf-") {
		return true
	}
	// Hardcoded group names used in SDS rule tests
	switch name {
	case "my group", "another group":
		return true
	}
	return false
}
