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
	for _, tc := range []struct{ name, schema string }{
		{"empty schema", "type: object\n              additionalProperties: {}"},
		{"boolean schema", "type: object\n              additionalProperties: true"},
		{"map inside list", "type: array\n              items:\n                type: object\n                additionalProperties: {}"},
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
				"              " + tc.schema + "\n"
			spec := strings.Replace(readThingResourceFixture(t), "    ThingUpdateAttributes:", field+"    ThingUpdateAttributes:", 1)
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
			for _, want := range []string{
				"resource.data.attributes.environments[].allocations",
				"additionalProperties declares unconstrained values",
				"concrete map value schema",
				"type: array with typed items",
			} {
				if !strings.Contains(message, want) {
					t.Errorf("diagnostic missing %q: %s", want, message)
				}
			}
			if _, err := os.Stat(filepath.Join(dir, "resource_datadog_thing.go")); !os.IsNotExist(err) {
				t.Fatalf("failed artifact must not emit resource code: %v", err)
			}
		})
	}
}
