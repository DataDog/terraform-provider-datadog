package cli

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/providermod"
)

// updateFixtures rewrites each fixture's out/ directory from the current
// generator output. Run `make tfgen-update-goldens` after an intentional
// generator change and review the diff by hand.
var updateFixtures = flag.Bool("update", false, "rewrite fixture out/ directories from generator output")

// TestFixtureGoldensMatchGeneratorOutput regenerates every fixture in the
// catalogue and requires its committed out/ directory to match: the same file
// set, byte for byte. Compiling a golden (internal/testdata) proves it is valid
// Go, not that tfgen still produces it.
func TestFixtureGoldensMatchGeneratorOutput(t *testing.T) {
	providerRoot, err := providermod.Root("")
	if err != nil {
		t.Fatal(err)
	}
	specs, err := filepath.Glob(filepath.Join("..", "testdata", "fixtures", "*", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) == 0 {
		t.Fatal("no fixtures found under internal/testdata/fixtures")
	}

	for _, spec := range specs {
		fixture := filepath.Dir(spec)
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			generated := generateFixture(t, providerRoot, spec)
			golden := filepath.Join(fixture, "out")
			if *updateFixtures {
				rewriteGolden(t, golden, generated)
				return
			}
			compareGolden(t, golden, generated)
		})
	}
}

// generateFixture runs tfgen on spec into an empty output root and returns it.
// The fixture README's two reproduction conditions hold here: the output root
// starts empty, and its ../internal/utils resolves to the provider checkout so
// Configure binds the ApiInstances accessor rather than the fallback.
func generateFixture(t *testing.T, providerRoot, spec string) string {
	t.Helper()
	dir := t.TempDir()
	outputRoot := filepath.Join(dir, "datadog", "fwprovider")
	if err := os.MkdirAll(outputRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(providerRoot, "datadog", "internal"), filepath.Join(dir, "datadog", "internal")); err != nil {
		t.Fatal(err)
	}
	if err := runTfgen("generate", "--spec", spec,
		"--output-root", outputRoot,
		"--examples-output-root", filepath.Join(dir, "examples"),
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--report", filepath.Join(dir, "report.json")); err != nil {
		t.Fatalf("generate: %v", err)
	}
	return outputRoot
}

func compareGolden(t *testing.T, golden, generated string) {
	t.Helper()
	want, got := goFiles(t, golden), goFiles(t, generated)
	if !slices.Equal(want, got) {
		t.Fatalf("%s holds %v but tfgen generates %v; run `make tfgen-update-goldens`", golden, want, got)
	}
	for _, name := range want {
		wantSrc, gotSrc := mustReadBytes(t, filepath.Join(golden, name)), mustReadBytes(t, filepath.Join(generated, name))
		if !bytes.Equal(wantSrc, gotSrc) {
			t.Errorf("%s differs from tfgen output; run `make tfgen-update-goldens` and review the diff", filepath.Join(golden, name))
		}
	}
}

func rewriteGolden(t *testing.T, golden, generated string) {
	t.Helper()
	for _, name := range goFiles(t, golden) {
		if err := os.Remove(filepath.Join(golden, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range goFiles(t, generated) {
		if err := os.WriteFile(filepath.Join(golden, name), mustReadBytes(t, filepath.Join(generated, name)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// goFiles lists the .go files directly in dir, sorted.
func goFiles(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = filepath.Base(m)
	}
	slices.Sort(names)
	return names
}

func mustReadBytes(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
