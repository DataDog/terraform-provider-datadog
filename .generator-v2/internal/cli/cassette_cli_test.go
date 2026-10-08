package cli

import (
	"bytes"
	"encoding/json"
	"errors"
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

// mutatedSpec copies the coherent fixture with one textual substitution, so a
// case can make the description self-contradicting without a second fixture to
// keep in step with this one.
func mutatedSpec(t *testing.T, dir, old, replacement string) string {
	t.Helper()
	source := filepath.Join("..", "testdata", "mini-oas",
		"mini-datadog_integration_twilio_account.yaml")
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if !bytes.Contains(raw, []byte(old)) {
		t.Fatalf("fixture no longer contains %q; the mutation would be a no-op", old)
	}
	mutated := bytes.Replace(raw, []byte(old), []byte(replacement), 1)
	path := filepath.Join(dir, "mutated.yaml")
	if err := os.WriteFile(path, mutated, 0o644); err != nil {
		t.Fatalf("writing mutated spec: %v", err)
	}
	return path
}

// contradictedNameSpec gives the create request's name a format its own
// example does not satisfy. A format needs no SDK type, so the artifact still
// binds and generates; only the cassette target becomes ineligible, which is
// the mixed condition these cases need. (An invented enum would also
// contradict the example, but enums bind to SDK types and would fail the
// artifact instead.)
func contradictedNameSpec(t *testing.T, dir string) string {
	t.Helper()
	return mutatedSpec(t, dir,
		"          example: twilio-prod\n          type: string",
		"          example: twilio-prod\n          type: string\n          format: uuid")
}

// runGenerateWithSpec drives the real command against an arbitrary spec.
func runGenerateWithSpec(t *testing.T, specPath, dir, reportPath string, extra ...string) error {
	t.Helper()
	seedProviderTest(t, filepath.Join(dir, "tests"))
	args := []string{"generate",
		"--spec", specPath,
		"--include", "integration_twilio_account",
		"--output-root", filepath.Join(dir, "fwprovider"),
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--examples-output-root", filepath.Join(dir, "examples"),
		"--docs-root", filepath.Join(dir, "docs"),
		"--report", reportPath,
	}
	return runTfgen(append(args, extra...)...)
}

// Opting in with cassette: true is a request for a test. A target that could
// not produce one is a failure of the run, not a silent success — otherwise
// the annotation looks satisfied when it is not.
func TestGenerateFailsWhenARequestedTargetIsIneligible(t *testing.T) {
	dir := t.TempDir()
	report := filepath.Join(dir, "report.json")

	err := runGenerateWithSpec(t, contradictedNameSpec(t, dir), dir, report)
	if err == nil {
		t.Fatal("an ineligible target exited 0")
	}
	if errors.Is(err, errCheckFailed) {
		t.Errorf("ineligibility reported as check drift: %v", err)
	}
	if !strings.Contains(err.Error(), "integration_twilio_account") {
		t.Errorf("error does not name the ineligible target: %v", err)
	}

	// Every independent result still has to survive the failure.
	got := readCassetteReport(t, report)
	if len(got.Cassettes) != 1 || got.Cassettes[0].Status != model.CassetteStatusIneligible {
		t.Fatalf("report does not record the ineligibility: %+v", got.Cassettes)
	}
	if got.CassetteSummary == nil || got.CassetteSummary.Ineligible != 1 {
		t.Errorf("cassette summary = %+v, want ineligible 1", got.CassetteSummary)
	}
	// The source artifact is independent of the cassette and must still be
	// generated and reported.
	if len(got.Artifacts) == 0 {
		t.Fatal("the source artifact result was discarded")
	}
	for _, a := range got.Artifacts {
		if a.Status == model.ArtifactStatusFailed {
			t.Errorf("cassette ineligibility failed the source artifact %q", a.Name)
		}
	}
	if _, statErr := os.Stat(filepath.Join(dir, "fwprovider",
		"resource_datadog_integration_twilio_account.go")); statErr != nil {
		t.Errorf("the eligible source artifact was not written: %v", statErr)
	}
}

// Exit 1 takes precedence over exit 3: a --check run finding both an
// ineligible target and output drift is a failure, not drift.
func TestGenerateIneligibilityOutranksCheckDrift(t *testing.T) {
	dir := t.TempDir()
	report := filepath.Join(dir, "report.json")

	// Nothing exists yet, so every file would change — drift is guaranteed,
	// and the contradiction makes the target ineligible at the same time.
	err := runGenerateWithSpec(t, contradictedNameSpec(t, dir), dir, report, "--check")
	if err == nil {
		t.Fatal("check mode with an ineligible target exited 0")
	}
	if errors.Is(err, errCheckFailed) {
		t.Fatalf("exit 3 took precedence over exit 1: %v", err)
	}
	if !strings.Contains(err.Error(), "ineligible") {
		t.Errorf("error does not report ineligibility: %v", err)
	}
	// Both facts belong in the report regardless of which one set the code.
	got := readCassetteReport(t, report)
	if got.CassetteSummary == nil || got.CassetteSummary.Ineligible != 1 {
		t.Errorf("report lost the ineligibility: %+v", got.CassetteSummary)
	}
}

// Drift alone is still exit 3, which is what makes the precedence meaningful.
func TestGenerateCheckDriftAloneIsStillCheckFailed(t *testing.T) {
	dir := t.TempDir()
	err := runCassetteGenerate(t, dir, filepath.Join(dir, "report.json"), "--check")
	if !errors.Is(err, errCheckFailed) {
		t.Fatalf("check error = %v, want errCheckFailed", err)
	}
}
