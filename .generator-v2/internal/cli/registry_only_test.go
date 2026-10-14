package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestGenerateIgnoresLegacyOperationAnnotations(t *testing.T) {
	dir := t.TempDir()
	spec := filepath.Join(dir, "legacy.yaml")
	report := filepath.Join(dir, "report.json")
	body := `openapi: 3.0.0
info: {title: Legacy annotations, version: "1"}
paths:
  /things:
    get:
      operationId: ListThings
      x-datadog-tf-generator:
        artifact_kind: data_source
        artifact_name: things
        cardinality: plural
        group: {read: ListThings}
      responses:
        '200': {description: OK}
    post:
      operationId: CreateThing
      x-datadog-tf-generator: malformed
      responses:
        '201': {description: Created}
`
	if err := os.WriteFile(spec, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runTfgen("generate", "--spec", spec, "--output-root", dir, "--report", report); err != nil {
		t.Fatal(err)
	}
	var result model.RunReport
	if err := json.Unmarshal([]byte(mustRead(t, report)), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Artifacts) != 0 {
		t.Fatalf("legacy annotations generated artifacts: %+v", result.Artifacts)
	}
}
