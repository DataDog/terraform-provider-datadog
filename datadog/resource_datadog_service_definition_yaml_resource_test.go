package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestServiceDefinitionYAMLResourceV3Schema(t *testing.T) {
	resource := resourceDatadogServiceDefinitionYAML()

	// Test that the resource is properly configured
	if resource.CreateContext == nil {
		t.Fatal("CreateContext should not be nil")
	}
	if resource.ReadContext == nil {
		t.Fatal("ReadContext should not be nil")
	}
	if resource.UpdateContext == nil {
		t.Fatal("UpdateContext should not be nil")
	}
	if resource.DeleteContext == nil {
		t.Fatal("DeleteContext should not be nil")
	}

	// Test schema
	schemaMap := resource.SchemaFunc()
	serviceDefSchema, ok := schemaMap["service_definition"]
	if !ok {
		t.Fatal("service_definition should be in schema")
	}

	if serviceDefSchema.Type != schema.TypeString {
		t.Error("service_definition should be of type string")
	}

	if !serviceDefSchema.Required {
		t.Error("service_definition should be required")
	}

	// Test v3 schema validation through the ValidateFunc
	validV3YAML := `schema-version: v3
dd-service: test-service-v3
team: test-team
tier: high
lifecycle: production
contacts:
  - name: Support Email
    type: email
    contact: team@example.com
links:
  - name: Runbook
    type: runbook
    url: https://runbook/test-service
tags:
  - env:prod
  - team:platform
integrations:
  pagerduty:
    service-url: https://my-org.pagerduty.com/service-directory/Ptest-service`

	warnings, errors := serviceDefSchema.ValidateFunc(validV3YAML, "service_definition")

	if len(errors) > 0 {
		t.Errorf("Valid v3 YAML should not produce errors: %v", errors)
	}

	if len(warnings) > 0 {
		t.Logf("Warnings for v3 YAML: %v", warnings)
	}
}

func TestServiceDefinitionYAMLStateFunc(t *testing.T) {
	resource := resourceDatadogServiceDefinitionYAML()
	schemaMap := resource.SchemaFunc()
	serviceDefSchema := schemaMap["service_definition"]

	// Test that StateFunc normalizes v3 YAML properly
	inputV3YAML := `schema-version: v3
dd-service: test-service-v3
team: test-team
tier: high
lifecycle: production
tags:
  - env:prod
  - team:platform
contacts:
  - name: Support Email
    type: email
    contact: team@example.com`

	normalizedYAML := serviceDefSchema.StateFunc(inputV3YAML)

	// The StateFunc should return a normalized version
	if normalizedYAML == "" {
		t.Error("StateFunc should not return empty string for valid v3 YAML")
	}

	// Verify the normalized YAML is still valid
	warnings, errors := serviceDefSchema.ValidateFunc(normalizedYAML, "service_definition")
	if len(errors) > 0 {
		t.Errorf("Normalized v3 YAML should be valid: %v", errors)
	}
	if len(warnings) > 0 {
		t.Logf("Warnings for normalized v3 YAML: %v", warnings)
	}
}
