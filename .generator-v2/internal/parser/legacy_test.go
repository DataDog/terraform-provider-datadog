package parser

import (
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v4"
)

func TestLegacySchemaWarningsTraverseSchemasOnly(t *testing.T) {
	source := `components:
  schemas:
    Thing:
      type: object
      properties:
        example:
          type: string
          x-custom: {sensitive: true}
        items:
          type: array
          items:
            allOf:
            - type: string
              x-custom: {sensitive: false}
      example:
        schema:
          x-custom: {sensitive: true}
      x-metadata:
        schema:
          x-custom: {sensitive: true}
paths:
  /things:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: string
                x-custom: {sensitive: true}
tags:
- name: Things
  x-custom:
    x-metadata:
      schema:
        x-custom: {sensitive: true}
`
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(source), &root); err != nil {
		t.Fatal(err)
	}
	warnings := legacySchemaWarnings(&root, "x-custom")
	if len(warnings) != 3 {
		t.Fatalf("want 3 schema warnings, got %v", warnings)
	}
	for _, warning := range warnings {
		if !strings.Contains(warning, "x-custom.sensitive at line") {
			t.Errorf("missing location: %s", warning)
		}
	}
	if warnings := legacySchemaWarnings(&root, DefaultTrackingFieldName); len(warnings) != 0 {
		t.Fatalf("wrong extension was reported: %v", warnings)
	}
}
