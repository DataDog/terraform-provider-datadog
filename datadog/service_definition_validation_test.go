package datadog

import (
	"testing"
)

func TestServiceDefinitionValidation(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		expectError bool
	}{
		{
			name: "valid v2 schema",
			yaml: `schema-version: v2
dd-service: test-service
team: test-team`,
			expectError: false,
		},
		{
			name: "valid v2.1 schema",
			yaml: `schema-version: v2.1
dd-service: test-service
team: test-team`,
			expectError: false,
		},
		{
			name: "valid v2.2 schema",
			yaml: `schema-version: v2.2
dd-service: test-service
team: test-team`,
			expectError: false,
		},
		{
			name: "valid v3 schema",
			yaml: `schema-version: v3
dd-service: test-service
team: test-team
tier: high
lifecycle: production`,
			expectError: false,
		},
		{
			name: "valid v3.1 schema",
			yaml: `schema-version: v3.1
dd-service: test-service
team: test-team
tier: high
lifecycle: production`,
			expectError: false,
		},
		{
			name: "invalid v1 schema",
			yaml: `schema-version: v1
dd-service: test-service
team: test-team`,
			expectError: true,
		},
		{
			name: "missing schema-version",
			yaml: `dd-service: test-service
team: test-team`,
			expectError: true,
		},
		{
			name: "missing dd-service",
			yaml: `schema-version: v3
team: test-team`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, errors := isValidServiceDefinition(tt.yaml, "service_definition")

			if tt.expectError {
				if len(errors) == 0 {
					t.Errorf("Expected validation errors but got none")
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("Expected no validation errors but got: %v", errors)
				}
			}

			// warnings are not fatal, so we just log them
			if len(warnings) > 0 {
				t.Logf("Warnings: %v", warnings)
			}
		})
	}
}
