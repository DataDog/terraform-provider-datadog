package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// runTfgen builds the root command exactly as Execute does, but with explicit
// args and errors returned rather than printed, so tests can assert on them.
func runTfgen(args ...string) error {
	flags := &globalFlags{}
	root := newRootCmd("test", flags)
	root.AddCommand(newGenerateCmd(flags))
	root.AddCommand(newVerifyCmd(flags))
	root.SetArgs(args)
	root.SilenceErrors = true
	return root.Execute()
}

// TestGenerateWiresMaxDepth proves the --max-depth flag value reaches LoadSpec:
// the same deep-but-acyclic spec loads at a high bound and fails at a low one.
// If the flag were ignored, both runs would behave identically.
func TestGenerateWiresMaxDepth(t *testing.T) {
	deep := filepath.Join("..", "testdata", "parser", "deep_chain.yaml")

	if err := runTfgen("generate", "--spec", deep, "--max-depth", "20"); err != nil {
		t.Errorf("--max-depth 20 should load the 8-deep chain, got: %v", err)
	}
	if err := runTfgen("generate", "--spec", deep, "--max-depth", "4"); err == nil {
		t.Error("--max-depth 4 should fail on the 8-deep chain, got nil")
	}
}

// TestGenerateSurfacesCycleError proves cycle detection is reachable through the
// command and surfaces the typed error.
func TestGenerateSurfacesCycleError(t *testing.T) {
	self := filepath.Join("..", "testdata", "parser", "cycle_self.yaml")

	err := runTfgen("generate", "--spec", self)
	var cycleErr *parser.RefCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error %v (%T) is not a *parser.RefCycleError", err, err)
	}
}

// TestGenerateWiresOverwrite drives the full generate path with a spec that sets
// overwrites and proves the three wiring effects: the generated data source is
// written, its constructor is registered in datasources_generated.go, and the
// overwritten hand-written constructor is removed from the Datasources slice
// without disturbing its neighbors.
func TestGenerateWiresOverwrite(t *testing.T) {
	spec, err := os.ReadFile(filepath.Join("..", "testdata", "mini-oas", "scripts", "gen-test", "datastore.yaml"))
	if err != nil {
		t.Fatalf("reading datastore spec: %v", err)
	}
	// Opt the datastore data source into overwriting the hand-written one.
	withOverwrites := strings.Replace(string(spec),
		"        artifact_name: datastore\n",
		"        artifact_name: datastore\n        overwrites: NewDatadogDatastoreDataSource\n", 1)
	if !strings.Contains(withOverwrites, "overwrites: NewDatadogDatastoreDataSource") {
		t.Fatal("failed to inject overwrites into the datastore spec fixture")
	}

	dir := t.TempDir()
	specPath := filepath.Join(dir, "datastore.yaml")
	if err := os.WriteFile(specPath, []byte(withOverwrites), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := "package fwprovider\n\n" +
		"var Datasources = []func() datasource.DataSource{\n" +
		"\tNewAPIKeyDataSource,\n\tNewDatadogDatastoreDataSource,\n\tNewHostsDataSource,\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "framework_provider.go"), []byte(provider), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runTfgen("generate", "--spec", specPath, "--output-root", dir, "--report", filepath.Join(dir, "report.json")); err != nil {
		t.Fatalf("generate: %v", err)
	}

	generated := mustRead(t, filepath.Join(dir, "data_source_datadog_datastore.go"))
	if !strings.Contains(generated, "func NewDatadogDatastoreDataSource()") {
		t.Errorf("generated data source missing NewDatadogDatastoreDataSource constructor:\n%s", generated)
	}

	registered := mustRead(t, filepath.Join(dir, "datasources_generated.go"))
	if !strings.Contains(registered, "NewDatadogDatastoreDataSource") {
		t.Errorf("generated constructor not registered in datasources_generated.go:\n%s", registered)
	}

	prov := mustRead(t, filepath.Join(dir, "framework_provider.go"))
	if strings.Contains(prov, "NewDatadogDatastoreDataSource") {
		t.Errorf("overwritten hand-written constructor not removed from Datasources:\n%s", prov)
	}
	if !strings.Contains(prov, "NewAPIKeyDataSource") || !strings.Contains(prov, "NewHostsDataSource") {
		t.Errorf("removal disturbed neighboring Datasources entries:\n%s", prov)
	}
}

// TestGenerateFailsOnMissingOverwriteTarget proves the safety guard: an
// overwrites target absent from the framework Datasources slice — a typo, or an
// SDKv2 DataSourcesMap entry the generator cannot retire — fails the run rather
// than silently leaving both registrations in place to collide at mux time. The
// temp dir has no datasources_generated.go, so the target is genuinely absent
// rather than already retired by a prior run.
func TestGenerateFailsOnMissingOverwriteTarget(t *testing.T) {
	spec, err := os.ReadFile(filepath.Join("..", "testdata", "mini-oas", "scripts", "gen-test", "datastore.yaml"))
	if err != nil {
		t.Fatalf("reading datastore spec: %v", err)
	}
	withOverwrites := strings.Replace(string(spec),
		"        artifact_name: datastore\n",
		"        artifact_name: datastore\n        overwrites: NewNonexistentDataSource\n", 1)
	if !strings.Contains(withOverwrites, "overwrites: NewNonexistentDataSource") {
		t.Fatal("failed to inject overwrites into the datastore spec fixture")
	}

	dir := t.TempDir()
	specPath := filepath.Join(dir, "datastore.yaml")
	if err := os.WriteFile(specPath, []byte(withOverwrites), 0o644); err != nil {
		t.Fatal(err)
	}
	provider := "package fwprovider\n\n" +
		"var Datasources = []func() datasource.DataSource{\n" +
		"\tNewAPIKeyDataSource,\n\tNewHostsDataSource,\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "framework_provider.go"), []byte(provider), 0o644); err != nil {
		t.Fatal(err)
	}

	err = runTfgen("generate", "--spec", specPath, "--output-root", dir, "--report", filepath.Join(dir, "report.json"))
	if err == nil {
		t.Fatal("expected an error when overwrites names a constructor absent from the Datasources slice, got nil")
	}
	if !strings.Contains(err.Error(), "NewNonexistentDataSource") {
		t.Errorf("error should name the missing overwrites target, got: %v", err)
	}
}

// TestGenerateRejectsReconcileWithInclude proves the guard: orphan detection is
// only sound over the complete annotation set, so --reconcile refuses --include.
// The check runs before the spec loads, so the (nonexistent) spec is never read.
func TestGenerateRejectsReconcileWithInclude(t *testing.T) {
	err := runTfgen("generate", "--reconcile", "--include", "team", "--spec", "does-not-matter.yaml")
	if err == nil {
		t.Fatal("expected --reconcile combined with --include to be rejected, got nil")
	}
	if !strings.Contains(err.Error(), "reconcile") {
		t.Errorf("error should mention reconcile, got: %v", err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(content)
}
