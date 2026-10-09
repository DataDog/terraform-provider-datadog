package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// seedProviderTest writes the minimal provider_test.go the harness keeps its
// endpoint-tag map in. The real tree always has one; a generated test has to be
// registered there or it t.Fatals at startup, so generation reads it.
func seedProviderTest(t *testing.T, testsDir string) {
	t.Helper()
	if err := os.MkdirAll(testsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	const content = "package test\n\nvar testFiles2EndpointTags = map[string]string{\n" +
		"\t\"tests/provider_test\": \"terraform\",\n}\n"
	if err := os.WriteFile(filepath.Join(testsDir, "provider_test.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runCassetteGenerate drives the real command against the coherent fixture.
func runCassetteGenerate(t *testing.T, dir, reportPath string, extra ...string) error {
	t.Helper()
	seedProviderTest(t, filepath.Join(dir, "tests"))
	args := []string{"generate",
		"--spec", filepath.Join("..", "testdata", "mini-oas",
			"mini-datadog_integration_twilio_account.yaml"),
		"--include", "integration_twilio_account",
		"--output-root", filepath.Join(dir, "fwprovider"),
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--examples-output-root", filepath.Join(dir, "examples"),
		"--docs-root", filepath.Join(dir, "docs"),
		"--report", reportPath,
	}
	return runTfgen(append(args, extra...)...)
}

func readCassetteReport(t *testing.T, path string) model.RunReport {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}
	var report model.RunReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("decoding report: %v", err)
	}
	return report
}

// An artifact that does not carry cassette: true has to produce exactly what it
// produced before the feature existed — no bundle, and no cassette fields in
// the report. The opt-in is the only thing separating the two runs, so this
// uses the same fixture with that one line removed.
func TestGenerateWithoutTheOptInIsUnchanged(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join("..", "testdata", "mini-oas",
		"mini-datadog_integration_twilio_account.yaml")
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	optedOut := bytes.Replace(raw, []byte("        cassette: true\n"), nil, 1)
	if bytes.Equal(raw, optedOut) {
		t.Fatal("fixture no longer carries the cassette opt-in this test removes")
	}
	spec := filepath.Join(dir, "opted-out.yaml")
	if err := os.WriteFile(spec, optedOut, 0o644); err != nil {
		t.Fatalf("writing opted-out spec: %v", err)
	}

	reportPath := filepath.Join(dir, "report.json")
	if err := runTfgen("generate",
		"--spec", spec,
		"--include", "integration_twilio_account",
		"--output-root", filepath.Join(dir, "fwprovider"),
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--examples-output-root", filepath.Join(dir, "examples"),
		"--docs-root", filepath.Join(dir, "docs"),
		"--report", reportPath,
	); err != nil {
		t.Fatalf("generate: %v", err)
	}

	report := readCassetteReport(t, reportPath)
	if len(report.Cassettes) != 0 {
		t.Errorf("report carries %d cassette results without the opt-in", len(report.Cassettes))
	}
	if report.CassetteSummary != nil {
		t.Error("report carries a cassette summary without the opt-in")
	}
	if _, err := os.Stat(filepath.Join(dir, "tests")); !os.IsNotExist(err) {
		t.Error("a tests directory was created without the opt-in")
	}
}

// The generator emits the test and nothing else: the cassette and its freeze
// companion come from a recording run, so a generate run must not leave a
// fixture behind that would replay as if it were evidence.
func TestGenerateEmitsTheTestWithoutAFixture(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")
	if err := runCassetteGenerate(t, dir, reportPath); err != nil {
		t.Fatalf("generate: %v", err)
	}

	result := readCassetteReport(t, reportPath).Cassettes[0]
	if result.Status != model.CassetteStatusGenerated {
		t.Fatalf("status = %q; diagnostics: %v", result.Status, result.Diagnostics)
	}
	if result.WriteAction != model.CassetteWriteCreated {
		t.Errorf("write action = %q, want created", result.WriteAction)
	}

	raw, err := os.ReadFile(result.TestPath)
	if err != nil {
		t.Fatalf("reading the generated test: %v", err)
	}
	if !strings.Contains(string(raw), model.GeneratedMarker) {
		t.Error("the generated test carries no ownership marker")
	}
	// It has to be able to record, not only replay.
	if !strings.Contains(string(raw), "isRecording()") {
		t.Error("the generated test cannot record a cassette")
	}
	if _, err := os.Stat(filepath.Join(dir, "tests", "cassettes")); !os.IsNotExist(err) {
		t.Error("a fixture was generated; cassettes must come from a recording")
	}
}

// A rerun over tfgen's own test file is unchanged, not a collision.
func TestGenerateRerunLeavesTheTestUnchanged(t *testing.T) {
	dir := t.TempDir()
	if err := runCassetteGenerate(t, dir, filepath.Join(dir, "first.json")); err != nil {
		t.Fatalf("first generate: %v", err)
	}
	rerun := filepath.Join(dir, "second.json")
	if err := runCassetteGenerate(t, dir, rerun); err != nil {
		t.Fatalf("second generate: %v", err)
	}

	result := readCassetteReport(t, rerun).Cassettes[0]
	if result.WriteAction != model.CassetteWriteUnchanged {
		t.Errorf("write action = %q, want unchanged", result.WriteAction)
	}
}

// A generated test absent from testFiles2EndpointTags t.Fatals at startup, so
// emitting the bundle has to register it — the same wiring the data-source path
// has always done.
func TestGenerateRegistersTheCassetteTest(t *testing.T) {
	dir := t.TempDir()
	if err := runCassetteGenerate(t, dir, filepath.Join(dir, "report.json")); err != nil {
		t.Fatalf("generate: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "tests", "provider_test.go"))
	if err != nil {
		t.Fatalf("reading provider_test.go: %v", err)
	}
	// The harness matches a frame's file against "datadog/<key>.go", so the key
	// carries the tests/ prefix and no extension.
	const key = `"tests/resource_datadog_integration_twilio_account_openapi_example_test"`
	if !strings.Contains(string(raw), key) {
		t.Errorf("provider_test.go does not register the generated test under %s", key)
	}
}

// Check mode reports the chain's outcome without writing.
func TestGenerateEmitCassettesCheckModeWritesNothing(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")
	// --check exits 3 when anything would change, which the resource itself
	// does here; the cassette outcome is what this asserts.
	_ = runCassetteGenerate(t, dir, reportPath, "--check")

	report := readCassetteReport(t, reportPath)
	if len(report.Cassettes) != 1 {
		t.Fatalf("report carries %d cassette results, want 1", len(report.Cassettes))
	}
	// Check mode still answers the question — it would be created — while
	// leaving the tree alone, which is what lets --check exit 3 on drift.
	if action := report.Cassettes[0].WriteAction; action != model.CassetteWriteCreated {
		t.Errorf("write action = %q, want created in check mode", action)
	}
	if _, err := os.Stat(filepath.Join(dir, "tests", "cassettes")); !os.IsNotExist(err) {
		t.Error("check mode wrote a cassette")
	}
}
