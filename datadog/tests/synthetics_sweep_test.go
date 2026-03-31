package test

import (
	"strings"
	"sync"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

var (
	syntheticsTestsSweepOnce     sync.Once
	syntheticsGlobalVarSweepOnce sync.Once
)

// cleanupSyntheticsTests removes stale Synthetics tests left behind by crashed
// or failed test runs. It uses sync.Once so that within a single test binary
// execution the cleanup runs at most once regardless of how many tests call it.
//
// Call this at the top of every Synthetics test function, before t.Parallel().
func cleanupSyntheticsTests(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	syntheticsTestsSweepOnce.Do(func() {
		doSweepSyntheticsTests(t)
	})
}

// TestSweepSyntheticsTests is a standalone sweep test for CI / manual
// invocation via `go test -run TestSweep` or `make sweep`.
func TestSweepSyntheticsTests(t *testing.T) {
	doSweepSyntheticsTests(t)
}

func doSweepSyntheticsTests(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	api := datadogV1.NewSyntheticsApi(client)

	// Collect all test IDs to delete by paginating through all tests.
	var toDelete []string
	var pageNumber int64
	const pageSize int64 = 100
	var totalSeen int

	for {
		opts := datadogV1.NewListTestsOptionalParameters().
			WithPageSize(pageSize).
			WithPageNumber(pageNumber)

		resp, _, err := api.ListTests(ctx, *opts)
		if err != nil {
			t.Logf("Synthetics sweep: failed to list tests (page %d): %v", pageNumber, err)
			return
		}

		tests := resp.GetTests()
		totalSeen += len(tests)
		t.Logf("Synthetics sweep: page %d returned %d tests", pageNumber, len(tests))

		for _, test := range tests {
			name := test.GetName()
			id := test.GetPublicId()
			if isSyntheticsTestResource(name) {
				t.Logf("Synthetics sweep: will delete test %q (id=%s)", name, id)
				toDelete = append(toDelete, id)
			} else {
				t.Logf("Synthetics sweep: skipping test %q (id=%s)", name, id)
			}
		}

		if int64(len(tests)) < pageSize {
			break
		}
		pageNumber++
	}

	t.Logf("Synthetics sweep: found %d total tests, %d to delete", totalSeen, len(toDelete))

	if len(toDelete) == 0 {
		t.Log("Synthetics sweep: no stale tests found")
		return
	}

	t.Logf("Synthetics sweep: deleting %d stale tests", len(toDelete))

	forceDelete := true
	payload := datadogV1.SyntheticsDeleteTestsPayload{
		PublicIds:               toDelete,
		ForceDeleteDependencies: &forceDelete,
	}

	_, httpResp, err := api.DeleteTests(ctx, payload)
	if err != nil {
		status := 0
		if httpResp != nil {
			status = httpResp.StatusCode
		}
		t.Logf("Synthetics sweep: failed to delete tests (status=%d): %v", status, err)
	}
}

// cleanupSyntheticsGlobalVariables removes stale Synthetics global variables
// left behind by crashed or failed test runs. Uses sync.Once like the tests sweeper.
//
// Call this at the top of every Synthetics global variable test function.
func cleanupSyntheticsGlobalVariables(t *testing.T) {
	t.Helper()

	if isReplaying() {
		return
	}

	syntheticsGlobalVarSweepOnce.Do(func() {
		doSweepSyntheticsGlobalVariables(t)
	})
}

// TestSweepSyntheticsGlobalVariables is a standalone sweep test for CI / manual invocation.
func TestSweepSyntheticsGlobalVariables(t *testing.T) {
	doSweepSyntheticsGlobalVariables(t)
}

func doSweepSyntheticsGlobalVariables(t *testing.T) {
	t.Helper()

	ctx, client := newSweepAPIClient(t)
	if client == nil {
		return
	}

	api := datadogV1.NewSyntheticsApi(client)

	resp, _, err := api.ListGlobalVariables(ctx)
	if err != nil {
		t.Logf("Synthetics global variable sweep: failed to list variables: %v", err)
		return
	}

	var deleted int
	for _, v := range resp.GetVariables() {
		name := v.GetName()
		id := v.GetId()

		if !isSyntheticsTestResource(name) {
			continue
		}

		t.Logf("Synthetics global variable sweep: deleting %q (id=%s)", name, id)

		httpResp, err := api.DeleteGlobalVariable(ctx, id)
		if err != nil {
			status := 0
			if httpResp != nil {
				status = httpResp.StatusCode
			}
			t.Logf("Synthetics global variable sweep: failed to delete %q (id=%s, status=%d): %v", name, id, status, err)
			continue
		}
		deleted++
	}

	if deleted == 0 {
		t.Log("Synthetics global variable sweep: no stale variables found")
	} else {
		t.Logf("Synthetics global variable sweep: deleted %d stale variables", deleted)
	}
}

// isSyntheticsTestResource returns true if the name looks like it was created by tests.
// Test names from uniqueEntityName() start with "tf-", global variable names
// from getUniqueVariableName() start with "TF_", and Datadog example tests
// start with "Example-".
func isSyntheticsTestResource(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, "tf-") || strings.HasPrefix(lower, "tf_") || strings.HasPrefix(lower, "example-")
}
