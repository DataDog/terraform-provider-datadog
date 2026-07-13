package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// runTfgen builds the root command exactly as Execute does, but with explicit
// args and errors returned rather than printed, so tests can assert on them.
func runTfgen(args ...string) error {
	flags := &globalFlags{}
	root := newRootCmd("test", flags)
	root.AddCommand(newGenerateCmd(flags))
	root.AddCommand(newVerifyCmd(flags))
	root.AddCommand(newSplitCmd(flags))
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

// TestGenerateWiresEndpointTag proves that generating a data source with
// --emit-tests registers its test in provider_test.go's testFiles2EndpointTags map
// (so the generated test does not t.Fatal at startup), and that retiring it removes
// the entry again.
func TestGenerateWiresEndpointTag(t *testing.T) {
	specPath := filepath.Join("..", "testdata", "mini-oas", "scripts", "gen-test", "datastore.yaml")
	if _, err := os.Stat(specPath); err != nil {
		t.Fatalf("datastore spec fixture: %v", err)
	}

	dir := t.TempDir()
	testsDir := filepath.Join(dir, "tests")
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	providerTest := "package test\n\nvar testFiles2EndpointTags = map[string]string{\n\t\"tests/provider_test\": \"terraform\",\n}\n"
	providerTestPath := filepath.Join(testsDir, "provider_test.go")
	if err := os.WriteFile(providerTestPath, []byte(providerTest), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runTfgen("generate", "--spec", specPath, "--emit-tests",
		"--output-root", dir, "--tests-output-root", testsDir,
		"--report", filepath.Join(dir, "report.json")); err != nil {
		t.Fatalf("generate: %v", err)
	}

	key := emit.EndpointTagTestKey("datastore")
	if pt := mustRead(t, providerTestPath); !strings.Contains(pt, key) {
		t.Errorf("generated test not registered in testFiles2EndpointTags (missing %q):\n%s", key, pt)
	}

	if err := runTfgen("generate", "--retire", "datastore",
		"--output-root", dir, "--tests-output-root", testsDir,
		"--report", filepath.Join(dir, "retire-report.json")); err != nil {
		t.Fatalf("retire: %v", err)
	}
	if pt := mustRead(t, providerTestPath); strings.Contains(pt, key) {
		t.Errorf("retire did not remove the testFiles2EndpointTags entry %q:\n%s", key, pt)
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

// TestReconcileSkippedWhenAnArtifactFails proves reconcile is fail-closed: when
// any artifact fails to build this run it drops out of the desired set, so
// retiring orphans could delete a live data source. The run has one registered
// orphan ("ghost") and a spec whose only artifact fails (an unsupported anyOf
// property, which fails before any file is written). Reconcile must be skipped,
// leaving the orphan's file and registration untouched despite the failure.
func TestReconcileSkippedWhenAnArtifactFails(t *testing.T) {
	spec, err := os.ReadFile(filepath.Join("..", "testdata", "mini-oas", "scripts", "gen-test", "datastore.yaml"))
	if err != nil {
		t.Fatalf("reading datastore spec: %v", err)
	}
	// Make the datastore artifact fail: an anyOf property is unsupported, so the
	// build rejects it before writing anything to the output root.
	failing := strings.Replace(string(spec),
		"        data:\n          $ref: '#/components/schemas/DatastoreData'\n",
		"        data:\n          anyOf:\n          - type: string\n          - type: integer\n", 1)
	if !strings.Contains(failing, "anyOf:") {
		t.Fatal("failed to inject the unsupported anyOf into the datastore spec fixture")
	}

	dir := t.TempDir()
	specPath := filepath.Join(dir, "datastore.yaml")
	if err := os.WriteFile(specPath, []byte(failing), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	// Seed a registered, generated (marker'd) orphan not present in the spec.
	ghost := "package fwprovider\n\n// Code generated by tfgen; DO NOT EDIT.\n\nfunc " + emit.DatasourceConstructor("ghost") + "() {}\n"
	ghostPath := filepath.Join(out, "data_source_datadog_ghost.go")
	if err := os.WriteFile(ghostPath, []byte(ghost), 0o644); err != nil {
		t.Fatal(err)
	}
	genPath := filepath.Join(out, "datasources_generated.go")
	if _, err := emit.SyncGeneratedDatasources(genPath, []string{emit.DatasourceConstructor("ghost")}, false); err != nil {
		t.Fatalf("seeding registry: %v", err)
	}

	err = runTfgen("generate", "--reconcile", "--spec", specPath,
		"--output-root", out,
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--docs-root", filepath.Join(dir, "docs"),
		"--report", filepath.Join(dir, "report.json"))
	if err == nil {
		t.Fatal("expected the run to fail because the datastore artifact failed to build, got nil")
	}

	// The orphan must survive: reconcile was skipped, so nothing was retired.
	if _, statErr := os.Stat(ghostPath); statErr != nil {
		t.Errorf("orphan file was deleted despite a failed artifact (reconcile should be skipped): %v", statErr)
	}
	if reg := mustRead(t, genPath); !strings.Contains(reg, emit.DatasourceConstructor("ghost")) {
		t.Errorf("orphan registration was removed despite a failed artifact:\n%s", reg)
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
