//go:build integration

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"

	testinfra "github.com/terraform-providers/terraform-provider-datadog/generator/internal/testdata"
)

func TestGenerateUpdateOnlyInitialization(t *testing.T) {
	root := t.TempDir()
	source, err := os.ReadFile(filepath.Join("..", "testdata", "fixtures", "resource_incident_type", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]any
	if err := yaml.Unmarshal(source, &spec); err != nil {
		t.Fatal(err)
	}
	// Keep real SDK types and endpoints, but make description writable only
	// through Update. The Create and Read schemas share IncidentTypeAttributes.
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	attributes := schemas["IncidentTypeAttributes"].(map[string]any)["properties"].(map[string]any)
	attributes["description"].(map[string]any)["readOnly"] = true
	tag := spec["tags"].([]any)[0].(map[string]any)
	resources := tag["x-datadog-tf-generator"].(map[string]any)
	definition := resources["incident_type"].(map[string]any)
	delete(resources, "incident_type")
	resources["tfgen_initialization_fixture"] = definition
	definition["resource"].(map[string]any)["artifact_name"] = "tfgen_initialization_fixture"
	contents, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(root, "initialization.yaml")
	if err := os.WriteFile(specPath, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	fixture := writeOnlyResourceFixture{
		name: "Initialization", spec: specPath, resource: "resource_datadog_tfgen_initialization_fixture.go",
	}
	generatedPath, generated := generateWriteOnlyFixture(t, root, fixture, "first")
	probe := filepath.Join("..", "testdata", "fixtures", "resource_incident_type", "initialization_test.go.txt")
	run := func(t *testing.T, path string) testinfra.CompileResult {
		t.Helper()
		result, err := testinfra.RunTests([]testinfra.StagedFile{
			{ProviderPath: filepath.Join("datadog", "fwprovider", fixture.resource), SourcePath: path},
			{ProviderPath: filepath.Join("datadog", "fwprovider", "zz_tfgen_initialization_test.go"), SourcePath: probe},
		}, "^TestTFGenUpdateOnlyInitialization$")
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	if result := run(t, generatedPath); !result.OK() {
		t.Fatal(result.Diagnostics)
	}
	t.Run("detects a planned value mistaken for a returned value", func(t *testing.T) {
		reset := []byte("\n\tstate.Description = types.StringNull()\n")
		if bytes.Count(generated, reset) != 1 {
			t.Fatal("expected exactly one response-only initialization reset")
		}
		broken := bytes.Replace(generated, reset, []byte("\n"), 1)
		brokenPath := filepath.Join(t.TempDir(), fixture.resource)
		if err := os.WriteFile(brokenPath, broken, 0o600); err != nil {
			t.Fatal(err)
		}
		result := run(t, brokenPath)
		if result.OK() || !strings.Contains(result.Diagnostics, "initialization PATCH calls = 0, want 1") {
			t.Fatalf("expected the sparse-response regression assertion to fail:\n%s", result.Diagnostics)
		}
	})
}
