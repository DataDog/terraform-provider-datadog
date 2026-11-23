package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// Exercise the full parser -> SDK binding -> lifecycle -> report pipeline. Each
// invalid explicit selection must fail closed while a resource sharing the
// original operations still generates with its original positional inference.
func TestIDParameterFailureIsArtifactLocal(t *testing.T) {
	parameter := func(name, schema string) string {
		return fmt.Sprintf("      - name: %s\n        in: path\n        required: true\n        schema: %s\n", name, schema)
	}
	for _, tc := range []struct {
		name, role, method, route, parameters, reason string
	}{
		{"missing read", "read", "get", "/broken/{other_id}", parameter("other_id", "{type: string}"), "no required SDK path argument"},
		{"missing update", "update", "patch", "/broken/{other_id}", parameter("other_id", "{type: string}"), "no required SDK path argument"},
		{"missing delete", "delete", "delete", "/broken/{other_id}/archive", parameter("other_id", "{type: string}"), "no required SDK path argument"},
		{"exact name required", "read", "get", "/broken/{ThingId}", parameter("ThingId", "{type: string}"), "no required SDK path argument"},
		{"unsupported conversion", "read", "get", "/broken/{thing_id}", parameter("thing_id", "{type: array, items: {type: string}}"), "supported conversion"},
		{"unknown scalar format", "read", "get", "/broken/{thing_id}", parameter("thing_id", "{type: integer, format: unsupported}"), "cannot derive SDK binding"},
		{"anonymous enum", "read", "get", "/broken/{thing_id}", parameter("thing_id", "{type: string, enum: [first, second]}"), "cannot derive SDK binding"},
		{"conflicting id", "read", "get", "/broken/{thing_id}/{id}", parameter("thing_id", "{type: string}") + parameter("id", "{type: string}"), "also claims Terraform's id"},
		{"create needs response id", "create", "post", "/broken/{thing_id}", parameter("thing_id", "{type: string}"), "before response data.id exists"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const operation = "BrokenLifecycleOperation"
			spec := readThingResourceFixture(t)
			roleOps := map[string]string{"create": "CreateThing", "read": "GetThing", "update": "UpdateThing", "delete": "DeleteThing"}
			roleOps[tc.role] = operation
			broken := fmt.Sprintf("    broken:\n      resource:\n        id_parameter: thing_id\n        group: {create: %s, read: %s, update: %s, delete: %s}\n",
				roleOps["create"], roleOps["read"], roleOps["update"], roleOps["delete"])
			// The fixture ends in the tag registry, so this adds a sibling family.
			spec += broken
			extra := fmt.Sprintf("  /api/v2%s:\n    %s:\n      operationId: %s\n      tags: [Things]\n      parameters:\n%s      responses:\n        '200':\n          description: Broken lifecycle response.\n          content:\n            application/json:\n              schema:\n                $ref: '#/components/schemas/ThingResponse'\n",
				tc.route, tc.method, operation, tc.parameters)
			spec = strings.Replace(spec, "components:\n", extra+"components:\n", 1)
			dir := t.TempDir()
			specPath, reportPath := filepath.Join(dir, "spec.yaml"), filepath.Join(dir, "report.json")
			if err := os.WriteFile(specPath, []byte(spec), 0600); err != nil {
				t.Fatal(err)
			}
			if err := runTfgen("generate", "--spec", specPath, "--output-root", dir,
				"--tests-output-root", filepath.Join(dir, "tests"),
				"--examples-output-root", filepath.Join(dir, "examples"), "--report", reportPath); err == nil {
				t.Fatal("invalid explicit selection unexpectedly generated")
			}
			var report model.RunReport
			if err := json.Unmarshal([]byte(mustRead(t, reportPath)), &report); err != nil {
				t.Fatal(err)
			}
			var healthy, failed bool
			for _, artifact := range report.Artifacts {
				if artifact.Name == "thing" && artifact.Status == model.ArtifactStatusCreated {
					healthy = true
				}
				if artifact.Name == "broken" && artifact.Status == model.ArtifactStatusFailed {
					failed = true
					diagnostics, err := json.Marshal(artifact.Diagnostics)
					if err != nil {
						t.Fatal(err)
					}
					for _, want := range []string{"broken", "group." + tc.role, operation, "thing_id", tc.reason} {
						if !strings.Contains(string(diagnostics), want) {
							t.Errorf("diagnostic missing %q: %s", want, diagnostics)
						}
					}
				}
			}
			if !healthy || !failed || len(report.Artifacts) != 2 {
				t.Fatalf("want one healthy and one failed artifact: %+v", report.Artifacts)
			}
			if _, err := os.Stat(filepath.Join(dir, "resource_datadog_broken.go")); !os.IsNotExist(err) {
				t.Fatalf("failed artifact emitted code: %v", err)
			}
			if source := mustRead(t, filepath.Join(dir, "resource_datadog_thing.go")); !strings.Contains(source, "state.ID.ValueString()") {
				t.Fatal("healthy sibling lost its inferred ID binding")
			}
		})
	}
}
