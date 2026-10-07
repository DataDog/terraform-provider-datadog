package datadog

import (
	"fmt"
	"strings"
	"testing"
)

func TestDashboardJSONStateIgnoresComputedFields(t *testing.T) {
	stateFunc := resourceDatadogDashboardJSON().SchemaFunc()["dashboard"].StateFunc
	configured := `{"layout_type":"ordered","title":"Example dashboard","widgets":[]}`

	computed := make([]string, 0, len(computedFields))
	for _, field := range computedFields {
		computed = append(computed, fmt.Sprintf(`%q:%q`, field, "server value"))
	}
	apiResponse := fmt.Sprintf(
		`{"layout_type":"ordered","title":"Example dashboard","widgets":[],%s}`,
		strings.Join(computed, ","),
	)

	configuredState := stateFunc(configured)
	responseState := stateFunc(apiResponse)
	if configuredState != responseState {
		t.Fatalf("expected computed fields to produce no state diff:\nconfigured: %s\nresponse:   %s", configuredState, responseState)
	}
}
