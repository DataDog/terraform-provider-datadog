//go:build integration

package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/config"
)

// repoRoot returns the absolute path to the provider repo root.
func repoRoot(t *testing.T) string {
	t.Helper()
	// From generator/internal/codegen/ the repo root is ../../../
	abs, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return abs
}

// goBuild runs "go build ./datadog/fwprovider/" from the repo root and returns any error.
func goBuild(t *testing.T) error {
	t.Helper()
	root := repoRoot(t)
	cmd := exec.Command("go", "build", "./datadog/fwprovider/")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &buildError{output: string(output), err: err}
	}
	return nil
}

type buildError struct {
	output string
	err    error
}

func (e *buildError) Error() string {
	return e.output + "\n" + e.err.Error()
}

// realSpecPath returns the path to the real Datadog V2 OpenAPI spec, or skips the test.
func realSpecPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join(repoRoot(t), ".generator", "V2", "openapi.yaml")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Skip("real spec not found at .generator/V2/openapi.yaml; skipping integration test")
	}
	return p
}

// generateToFwprovider generates a data source and copies it to the provider's fwprovider dir.
// Returns the path to the generated file. Files are cleaned up automatically.
func generateToFwprovider(t *testing.T, cfg *config.Config, configDir, name string) string {
	t.Helper()
	outputDir := t.TempDir()

	err := Generate(cfg, configDir, outputDir, false, 5)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	genPath := filepath.Join(outputDir, "data_source_datadog_"+name+"_generated.go")
	content, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	root := repoRoot(t)
	destPath := filepath.Join(root, "datadog", "fwprovider", "data_source_datadog_"+name+"_generated.go")
	if err := os.WriteFile(destPath, content, 0644); err != nil {
		t.Fatalf("copying generated file to provider: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(destPath)
	})

	// Also copy hooks file if not exists
	hooksPath := filepath.Join(outputDir, "data_source_datadog_"+name+"_hooks.go")
	hooksDestPath := filepath.Join(root, "datadog", "fwprovider", "data_source_datadog_"+name+"_hooks.go")
	if _, err := os.Stat(hooksDestPath); os.IsNotExist(err) {
		hooksContent, err := os.ReadFile(hooksPath)
		if err == nil {
			if err := os.WriteFile(hooksDestPath, hooksContent, 0644); err == nil {
				t.Cleanup(func() {
					os.Remove(hooksDestPath)
				})
			}
		}
	}

	return destPath
}

// T090: go build integration test — verifies generated code compiles within the provider.
// Uses the real Datadog V2 spec to generate a team data source, then runs go build
// to catch struct field mismatches, missing imports, and invalid SDK accessor names.
func TestGenerateAndBuild(t *testing.T) {
	specPath := realSpecPath(t)

	cfg := &config.Config{
		Specs: map[string]config.SpecConfig{
			"v2": {Path: specPath},
		},
		DataSources: map[string]config.DataSourceConfig{
			"integ_build_test": {
				Spec: "v2",
				Read: config.OperationRef{
					Path:   "/api/v2/team/{team_id}",
					Method: "get",
				},
			},
		},
	}

	destPath := generateToFwprovider(t, cfg, ".", "integ_build_test")

	// Verify the file was placed correctly
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatal("generated file not found in fwprovider directory")
	}

	// Run go build — this catches struct field mismatches, missing imports,
	// and invalid SDK accessor names that go/parser cannot detect.
	if err := goBuild(t); err != nil {
		t.Fatalf("go build failed:\n%v", err)
	}
}

// T091: go build with ListNestedBlock — verifies NestedObject wrapper compiles.
// Uses the real spec with cost_budget endpoint which has an array-of-objects field
// that produces a ListNestedBlock.
func TestGenerateAndBuild_ListNestedBlock(t *testing.T) {
	specPath := realSpecPath(t)

	cfg := &config.Config{
		Specs: map[string]config.SpecConfig{
			"v2": {Path: specPath},
		},
		DataSources: map[string]config.DataSourceConfig{
			"integ_nested_test": {
				Spec: "v2",
				Read: config.OperationRef{
					Path:   "/api/v2/cost/budget/{budget_id}",
					Method: "get",
				},
			},
		},
	}

	destPath := generateToFwprovider(t, cfg, ".", "integ_nested_test")

	// Verify ListNestedBlock uses NestedObject wrapper
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	s := string(content)

	if !strings.Contains(s, "schema.ListNestedBlock{") {
		t.Error("generated file should contain schema.ListNestedBlock{")
	}
	if !strings.Contains(s, "NestedObject: schema.NestedBlockObject{") {
		t.Error("ListNestedBlock should use NestedObject wrapper")
	}

	// Run go build — validates the NestedObject wrapper produces valid Go
	if err := goBuild(t); err != nil {
		t.Fatalf("go build failed (ListNestedBlock test):\n%v", err)
	}
}

// T092: go build with SDK casing — verifies SDK accessor names match actual SDK methods.
// Uses cost_budget which has fields with trailing acronyms (org_id).
func TestGenerateAndBuild_SDKCasing(t *testing.T) {
	specPath := realSpecPath(t)

	cfg := &config.Config{
		Specs: map[string]config.SpecConfig{
			"v2": {Path: specPath},
		},
		DataSources: map[string]config.DataSourceConfig{
			"integ_sdk_test": {
				Spec: "v2",
				Read: config.OperationRef{
					Path:   "/api/v2/cost/budget/{budget_id}",
					Method: "get",
				},
			},
		},
	}

	destPath := generateToFwprovider(t, cfg, ".", "integ_sdk_test")

	// Verify SDK casing patterns
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	s := string(content)

	// Should use SDK-style casing (lowercased trailing acronyms)
	if strings.Contains(s, "GetOrgID()") {
		t.Error("should use GetOrgId() not GetOrgID() for SDK accessor")
	}
	if !strings.Contains(s, "GetOrgId()") {
		t.Error("should contain GetOrgId() SDK accessor")
	}

	// Run go build — this is the ultimate validation that SDK accessor names
	// match the actual SDK method names
	if err := goBuild(t); err != nil {
		t.Fatalf("go build failed (SDK casing test):\n%v", err)
	}
}

// T109: go build integration test for composition — verifies code generated from
// schemas with allOf/oneOf/additionalProperties compiles within the provider.
// Uses the real Datadog V2 spec — the team endpoint's response schema uses allOf
// for inheritance (TeamResponse -> TeamData -> TeamAttributes).
func TestGenerateAndBuild_Composition(t *testing.T) {
	specPath := realSpecPath(t)

	// The team endpoint uses allOf in the real spec (TeamAttributes allOf composition).
	// This also exercises additionalProperties that may exist in some response schemas.
	cfg := &config.Config{
		Specs: map[string]config.SpecConfig{
			"v2": {Path: specPath},
		},
		DataSources: map[string]config.DataSourceConfig{
			"integ_composition_test": {
				Spec: "v2",
				Read: config.OperationRef{
					Path:   "/api/v2/team/{team_id}",
					Method: "get",
				},
			},
		},
	}

	destPath := generateToFwprovider(t, cfg, ".", "integ_composition_test")

	// Verify the file was generated
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	s := string(content)

	// Should contain standard generated patterns
	if !strings.Contains(s, "package fwprovider") {
		t.Error("generated file should contain package fwprovider")
	}
	if !strings.Contains(s, "datadogIntegCompositionTestDataSource") {
		t.Error("generated file should contain the data source struct")
	}

	// Run go build — validates that composition-generated code compiles
	// against the real SDK and Terraform framework
	if err := goBuild(t); err != nil {
		t.Fatalf("go build failed (composition test):\n%v", err)
	}
}
