package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readThingResourceFixture returns the annotated full-CRUD resource spec at
// testdata/mini-oas/scripts/gen-test/thing_resource.yaml.
func readThingResourceFixture(t *testing.T) string {
	t.Helper()
	spec, err := os.ReadFile(filepath.Join("..", "testdata", "mini-oas", "scripts", "gen-test", "thing_resource.yaml"))
	if err != nil {
		t.Fatalf("reading thing resource spec: %v", err)
	}
	return string(spec)
}

// TestGenerateCreatesAndRegistersResource drives a full-CRUD resource through
// parsing, model construction, resource-view building and rendering, and
// proves the CLI wires it into the provider: the resource file is written
// under --output-root (not a path hard-coded relative to the working
// directory), and resources_generated.go registers its constructor. A second,
// unmodified run is idempotent.
func TestGenerateCreatesAndRegistersResource(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "thing.yaml")
	if err := os.WriteFile(specPath, []byte(readThingResourceFixture(t)), 0o644); err != nil {
		t.Fatal(err)
	}

	run := func() error {
		return runTfgen("generate", "--spec", specPath,
			"--output-root", dir,
			"--examples-output-root", filepath.Join(dir, "examples"),
			"--tests-output-root", filepath.Join(dir, "tests"),
			"--report", filepath.Join(dir, "report.json"))
	}

	if err := run(); err != nil {
		t.Fatalf("generate: %v", err)
	}

	resourcePath := filepath.Join(dir, "resource_datadog_thing.go")
	generated := mustRead(t, resourcePath)
	for _, want := range []string{
		"func NewDatadogThingResource() resource.Resource",
		"func (r *datadogThingResource) Create(",
		"func (r *datadogThingResource) Read(",
		"func (r *datadogThingResource) Update(",
		"func (r *datadogThingResource) Delete(",
		// The JSON:API envelope is built level by level, each from its own
		// New<Type>WithDefaults(), so the data component's constructor sets the
		// "type" discriminator the API requires (T138).
		"bodyAttributes := datadogV2.NewThingAttributesWithDefaults()",
		"bodyAttributes.SetName(state.Name.ValueString())",
		"bodyData := datadogV2.NewThingCreateDataWithDefaults()",
		"bodyData.SetAttributes(*bodyAttributes)",
		"body := datadogV2.NewThingCreateRequestWithDefaults()",
		"body.SetData(*bodyData)",
	} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated resource missing %q:\n%s", want, generated)
		}
	}
	// Reaching through the wrapper is what left the discriminator empty; it
	// must not come back.
	if strings.Contains(generated, "body.Data.Attributes.") {
		t.Errorf("generated resource reaches through the request wrapper instead of building the envelope:\n%s", generated)
	}

	registered := mustRead(t, filepath.Join(dir, "resources_generated.go"))
	if !strings.Contains(registered, "NewDatadogThingResource") {
		t.Errorf("generated constructor not registered in resources_generated.go:\n%s", registered)
	}

	// A second run over the identical spec must not rewrite either file.
	beforeResource, err := os.Stat(resourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := run(); err != nil {
		t.Fatalf("second generate: %v", err)
	}
	afterResource, err := os.Stat(resourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if !afterResource.ModTime().Equal(beforeResource.ModTime()) {
		t.Errorf("second identical run rewrote %s (want unchanged)", resourcePath)
	}
}

// TestGenerateWiresResourceOverwrite proves that opting a spec into
// "overwrites" removes the named hand-written constructor from
// framework_provider.go's Resources slice without disturbing its neighbors,
// while the generated constructor lands in both the new resource file and
// resources_generated.go.
func TestGenerateWiresResourceOverwrite(t *testing.T) {
	spec := readThingResourceFixture(t)
	withOverwrites := strings.Replace(spec,
		"        artifact_name: thing\n",
		"        artifact_name: thing\n        overwrites: NewDatadogThingResource\n", 1)
	if !strings.Contains(withOverwrites, "overwrites: NewDatadogThingResource") {
		t.Fatal("failed to inject overwrites into the thing resource spec fixture")
	}

	dir := t.TempDir()
	specPath := filepath.Join(dir, "thing.yaml")
	if err := os.WriteFile(specPath, []byte(withOverwrites), 0o644); err != nil {
		t.Fatal(err)
	}

	provider := "package fwprovider\n\n" +
		"var Resources = []func() resource.Resource{\n" +
		"\tNewAPIKeyResource,\n\tNewDatadogThingResource,\n\tNewHostsResource,\n}\n"
	providerPath := filepath.Join(dir, "framework_provider.go")
	if err := os.WriteFile(providerPath, []byte(provider), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runTfgen("generate", "--spec", specPath,
		"--output-root", dir,
		"--examples-output-root", filepath.Join(dir, "examples"),
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--report", filepath.Join(dir, "report.json")); err != nil {
		t.Fatalf("generate: %v", err)
	}

	registered := mustRead(t, filepath.Join(dir, "resources_generated.go"))
	if !strings.Contains(registered, "NewDatadogThingResource") {
		t.Errorf("generated constructor not registered in resources_generated.go:\n%s", registered)
	}

	prov := mustRead(t, providerPath)
	if strings.Contains(prov, "NewDatadogThingResource") {
		t.Errorf("overwritten hand-written constructor not removed from Resources:\n%s", prov)
	}
	if !strings.Contains(prov, "NewAPIKeyResource") || !strings.Contains(prov, "NewHostsResource") {
		t.Errorf("removal disturbed neighboring Resources entries:\n%s", prov)
	}
}
