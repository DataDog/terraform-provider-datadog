package test

import (
	"os"
	"testing"
)

// sweepRequestedEnvName is set by `make sweep`. TestSweep* tests delete real
// resources through a live client that ignores RECORD, so a -run pattern that
// matches a sweep test name must not start a sweep by itself.
const sweepRequestedEnvName = "DD_TEST_SWEEP"

func sweepRequested() bool {
	return os.Getenv(sweepRequestedEnvName) == "true" && !isReplaying()
}

// skipUnlessSweepRequested must be the first call in every standalone TestSweep* test.
func skipUnlessSweepRequested(t *testing.T) {
	t.Helper()
	if !sweepRequested() {
		t.Skipf("sweep deletes real resources; runs only through `make sweep` (%s=true) and never under RECORD=false", sweepRequestedEnvName)
	}
}

func TestIsSweepRequested(t *testing.T) {
	for _, tc := range []struct {
		sweep, record string
		want          bool
	}{
		{"true", "", true},
		{"true", "none", true},
		{"true", "false", false},
		{"", "none", false},
		{"", "false", false},
		{"1", "", false},
	} {
		t.Setenv(sweepRequestedEnvName, tc.sweep)
		t.Setenv("RECORD", tc.record)
		if got := sweepRequested(); got != tc.want {
			t.Errorf("%s=%q RECORD=%q: sweepRequested() = %v, want %v", sweepRequestedEnvName, tc.sweep, tc.record, got, tc.want)
		}
	}
}
