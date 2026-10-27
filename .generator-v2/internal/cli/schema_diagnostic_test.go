package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// Exercise parsing, resource-schema merging and report serialization together:
// collection conversion used to discard both the field path and the reason.
func TestGenerateReportsUnconstrainedMapPath(t *testing.T) {
	unconstrained := []string{
		"additionalProperties declares unconstrained values",
		"concrete map value schema",
		"type: array with typed items",
	}
	for _, tc := range []struct {
		name, schema, components string
		want, notWant            []string
	}{
		{name: "empty schema", schema: "type: object\n              additionalProperties: {}", want: unconstrained},
		{name: "boolean schema", schema: "type: object\n              additionalProperties: true", want: unconstrained},
		{name: "map inside list", schema: "type: array\n              items:\n                type: object\n                additionalProperties: {}", want: unconstrained},
		{name: "free-form object value", schema: "type: object\n              additionalProperties:\n                type: object", want: unconstrained},
		{
			name:       "ref to empty component",
			schema:     "type: object\n              additionalProperties:\n                $ref: '#/components/schemas/AnyValue'",
			components: "    AnyValue:\n      description: Any value.\n",
			want:       append([]string{`its $ref "AnyValue" has no type or properties`}, unconstrained...),
			notWant:    []string{"under additionalProperties"},
		},
		{
			// An alias only points elsewhere; the type is missing at the end of
			// the chain, so that is the component to name.
			name:       "ref chain to empty component",
			schema:     "type: object\n              additionalProperties:\n                $ref: '#/components/schemas/AnyValueAlias'",
			components: "    AnyValueAlias:\n      $ref: '#/components/schemas/AnyValue'\n    AnyValue:\n      description: Any value.\n",
			want:       append([]string{`its $ref "AnyValue" has no type or properties`}, unconstrained...),
			notWant:    []string{`"AnyValueAlias"`},
		},
		{
			// The siblings carry no type, but the referenced component does
			// constrain its values: it is not unconstrained.
			name:       "ref with siblings to itemless array",
			schema:     "type: object\n              additionalProperties:\n                $ref: '#/components/schemas/ItemlessList'\n                description: A list.",
			components: "    ItemlessList:\n      type: array\n",
			notWant:    unconstrained,
		},
		{name: "object value counting properties", schema: "type: object\n              additionalProperties:\n                type: object\n                minProperties: 1", want: unconstrained},
		{name: "object value requiring properties", schema: "type: object\n              additionalProperties:\n                type: object\n                required: [id]", want: unconstrained},
		{name: "empty object value", schema: "type: object\n              additionalProperties:\n                type: object\n                additionalProperties: false", want: unconstrained},
		{
			// Constrained but unsupported: the unconstrained advice would point
			// at the wrong fix.
			name:    "anyOf value",
			schema:  "type: object\n              additionalProperties:\n                anyOf:\n                  - type: string\n                  - type: integer",
			notWant: unconstrained,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			specPath, reportPath := filepath.Join(dir, "spec.yaml"), filepath.Join(dir, "report.json")
			field := "        environments:\n" +
				"          type: array\n" +
				"          items:\n" +
				"            $ref: '#/components/schemas/Environment'\n" +
				"    Environment:\n" +
				"      type: object\n" +
				"      properties:\n" +
				"        allocations:\n" +
				"              " + tc.schema + "\n" +
				tc.components
			const anchor = "    ThingUpdateAttributes:"
			fixture := readThingResourceFixture(t)
			if !strings.Contains(fixture, anchor) {
				t.Fatalf("fixture lost the %q anchor the field is inserted before", anchor)
			}
			spec := strings.Replace(fixture, anchor, field+anchor, 1)
			if err := os.WriteFile(specPath, []byte(spec), 0600); err != nil {
				t.Fatal(err)
			}
			if err := runTfgen("generate", "--spec", specPath, "--include", "thing", "--output-root", dir, "--report", reportPath); err == nil {
				t.Fatal("expected unsupported map to fail generation")
			}
			var report model.RunReport
			if err := json.Unmarshal([]byte(mustRead(t, reportPath)), &report); err != nil {
				t.Fatal(err)
			}
			if len(report.Artifacts) != 1 || report.Artifacts[0].Status != model.ArtifactStatusFailed {
				t.Fatalf("expected exactly one failed artifact: %+v", report.Artifacts)
			}
			var messages []string
			for _, diagnostic := range report.Artifacts[0].Diagnostics {
				messages = append(messages, diagnostic.Message)
			}
			message := strings.Join(messages, "\n")
			for _, want := range append([]string{"resource.data.attributes.environments[].allocations"}, tc.want...) {
				if !strings.Contains(message, want) {
					t.Errorf("diagnostic missing %q: %s", want, message)
				}
			}
			for _, notWant := range tc.notWant {
				if strings.Contains(message, notWant) {
					t.Errorf("diagnostic unexpectedly contains %q: %s", notWant, message)
				}
			}
			if _, err := os.Stat(filepath.Join(dir, "resource_datadog_thing.go")); !os.IsNotExist(err) {
				t.Fatalf("failed artifact must not emit resource code: %v", err)
			}
		})
	}
}
