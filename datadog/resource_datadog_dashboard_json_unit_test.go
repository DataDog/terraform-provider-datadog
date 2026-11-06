package datadog

import (
	"fmt"
	"testing"
)

func TestDashboardJSONStateIgnoresComputedFields(t *testing.T) {
	stateFunc := resourceDatadogDashboardJSON().SchemaFunc()["dashboard"].StateFunc
	configured := `{"layout_type":"ordered","title":"Example dashboard","widgets":[]}`

	for _, experienceType := range []string{"default", "product_analytics"} {
		t.Run(experienceType, func(t *testing.T) {
			// Keep the response independent of the production ignore list so removing
			// experience_type from that list makes this regression test fail.
			apiResponse := fmt.Sprintf(`{
				"layout_type":"ordered","title":"Example dashboard","widgets":[],
				"id":"abc-def-ghi","author_handle":"user@example.com","author_name":"Example User",
				"created_at":"2026-10-01T00:00:00Z","modified_at":"2026-10-02T00:00:00Z",
				"url":"/dashboard/abc-def-ghi/example","experience_type":%q
			}`, experienceType)
			configuredState := stateFunc(configured)
			responseState := stateFunc(apiResponse)
			if configuredState != responseState {
				t.Fatalf("expected computed fields to produce no state diff:\nconfigured: %s\nresponse:   %s", configuredState, responseState)
			}
		})
	}
}
