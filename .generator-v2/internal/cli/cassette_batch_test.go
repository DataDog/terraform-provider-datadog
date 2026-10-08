package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// One incomplete description must not cost every other target its test. These
// cases drive a run over two opted-in resources where exactly one is
// ineligible, which is the only way to tell "the run stopped at the first
// problem" from "the run processed everything and reported it".
//
// The fixture is internal/testdata/cassette/mixed_eligibility.yaml; its header
// records how it was built and why the contradiction sits where it does.
// The ineligible target is the one the run processes first — paths sort
// elastic-cloud before twilio — so a run that aborted at the first problem
// would skip the eligible one and these cases would catch it.
const (
	mixedEligibleTarget   = "integration_twilio_account"
	mixedIneligibleTarget = "integration_elastic_cloud"
	mixedEligibleTestFile = "resource_datadog_integration_twilio_account_openapi_example_test.go"
	mixedIneligibleTest   = "resource_datadog_integration_elastic_cloud_openapi_example_test.go"
)

// runMixedGenerate drives the real command over both targets. No --include, so
// the run covers the whole fixture the way a real one would.
func runMixedGenerate(t *testing.T, dir, reportPath string, extra ...string) error {
	t.Helper()
	seedProviderTest(t, filepath.Join(dir, "tests"))
	args := []string{"generate",
		"--spec", filepath.Join("..", "testdata", "cassette", "mixed_eligibility.yaml"),
		"--output-root", filepath.Join(dir, "fwprovider"),
		"--tests-output-root", filepath.Join(dir, "tests"),
		"--examples-output-root", filepath.Join(dir, "examples"),
		"--docs-root", filepath.Join(dir, "docs"),
		"--report", reportPath,
	}
	return runTfgen(append(args, extra...)...)
}

// cassetteByName finds one target's result, so a case asserts about the target
// it means rather than about report ordering.
func cassetteByName(t *testing.T, report model.RunReport, name string) model.CassetteResult {
	t.Helper()
	for _, c := range report.Cassettes {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("report has no cassette result for %q; got %d results", name, len(report.Cassettes))
	return model.CassetteResult{}
}

func TestGenerateMixedRunProcessesEveryTarget(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")

	err := runMixedGenerate(t, dir, reportPath)

	// The ineligible target fails the run, per the CLI contract's exit codes.
	if err == nil {
		t.Fatal("a run with an ineligible target exited 0")
	}
	if errors.Is(err, errCheckFailed) {
		t.Errorf("ineligibility reported as check drift: %v", err)
	}
	if !strings.Contains(err.Error(), mixedIneligibleTarget) {
		t.Errorf("error does not name the ineligible target: %v", err)
	}
	// And it names only that one, so a reader is not sent looking at the other.
	if strings.Contains(err.Error(), mixedEligibleTarget) {
		t.Errorf("error blames the eligible target too: %v", err)
	}

	report := readCassetteReport(t, reportPath)

	// Every requested target is summarized, whichever way it went.
	if len(report.Cassettes) != 2 {
		t.Fatalf("got %d cassette results, want one per target", len(report.Cassettes))
	}
	if got := cassetteByName(t, report, mixedEligibleTarget); got.Status != model.CassetteStatusGenerated {
		t.Errorf("%s status = %q, want generated", mixedEligibleTarget, got.Status)
	}
	ineligible := cassetteByName(t, report, mixedIneligibleTarget)
	if ineligible.Status != model.CassetteStatusIneligible {
		t.Errorf("%s status = %q, want ineligible", mixedIneligibleTarget, ineligible.Status)
	}
	if len(ineligible.Diagnostics) == 0 {
		t.Errorf("%s carries no diagnostic explaining why", mixedIneligibleTarget)
	}
	if report.CassetteSummary == nil ||
		report.CassetteSummary.Generated != 1 || report.CassetteSummary.Ineligible != 1 {
		t.Errorf("cassette summary = %+v, want generated 1 ineligible 1", report.CassetteSummary)
	}
}

func TestGenerateMixedRunWritesTheEligibleTest(t *testing.T) {
	dir := t.TempDir()
	if err := runMixedGenerate(t, dir, filepath.Join(dir, "report.json")); err == nil {
		t.Fatal("expected the ineligible target to fail the run")
	}

	testsDir := filepath.Join(dir, "tests")
	if _, err := os.Stat(filepath.Join(testsDir, mixedEligibleTestFile)); err != nil {
		t.Errorf("the eligible target's test was not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(testsDir, mixedIneligibleTest)); !os.IsNotExist(err) {
		t.Errorf("the ineligible target wrote a test: %v", err)
	}

	// Registration follows the file: the eligible test is in the endpoint-tag
	// map, which it must be or it t.Fatals at startup, and the other is not.
	providerTest, err := os.ReadFile(filepath.Join(testsDir, "provider_test.go"))
	if err != nil {
		t.Fatalf("reading provider_test.go: %v", err)
	}
	if !strings.Contains(string(providerTest), strings.TrimSuffix(mixedEligibleTestFile, ".go")) {
		t.Error("the eligible test was not registered in testFiles2EndpointTags")
	}
	if strings.Contains(string(providerTest), strings.TrimSuffix(mixedIneligibleTest, ".go")) {
		t.Error("the ineligible target was registered despite writing no test")
	}
}

// A cassette is not a precondition for its resource. Both source artifacts
// have to generate, including the one whose cassette was rejected, or one
// incomplete description would cost the provider a resource.
func TestGenerateMixedRunPreservesSourceArtifacts(t *testing.T) {
	dir := t.TempDir()
	if err := runMixedGenerate(t, dir, filepath.Join(dir, "report.json")); err == nil {
		t.Fatal("expected the ineligible target to fail the run")
	}
	report := readCassetteReport(t, filepath.Join(dir, "report.json"))

	if report.Summary == nil || report.Summary.Failed != 0 {
		t.Errorf("summary = %+v, want no failed artifacts", report.Summary)
	}
	seen := map[string]model.ArtifactStatus{}
	for _, a := range report.Artifacts {
		seen[a.Name] = a.Status
	}
	for _, name := range []string{mixedEligibleTarget, mixedIneligibleTarget} {
		status, ok := seen[name]
		if !ok {
			t.Errorf("no artifact result for %q", name)
			continue
		}
		if status == model.ArtifactStatusFailed {
			t.Errorf("artifact %q failed because of its cassette", name)
		}
	}
	for _, name := range []string{
		"resource_datadog_integration_twilio_account.go",
		"resource_datadog_integration_elastic_cloud.go",
	} {
		if _, err := os.Stat(filepath.Join(dir, "fwprovider", name)); err != nil {
			t.Errorf("source artifact %s was not written: %v", name, err)
		}
	}
}

// Check mode reports the same mixed outcome and still writes nothing, so a
// reviewer can see both facts before anything lands.
func TestGenerateMixedRunCheckModeReportsBothAndWritesNothing(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")

	err := runMixedGenerate(t, dir, reportPath, "--check")
	if err == nil {
		t.Fatal("check mode over a mixed run exited 0")
	}
	// Exit 1 outranks check drift even though this run has plenty of drift.
	if errors.Is(err, errCheckFailed) {
		t.Errorf("check drift outranked ineligibility: %v", err)
	}

	report := readCassetteReport(t, reportPath)
	if report.CassetteSummary == nil ||
		report.CassetteSummary.Generated != 1 || report.CassetteSummary.Ineligible != 1 {
		t.Errorf("cassette summary = %+v, want generated 1 ineligible 1", report.CassetteSummary)
	}
	if _, err := os.Stat(filepath.Join(dir, "fwprovider")); !os.IsNotExist(err) {
		t.Errorf("check mode created the output root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "tests", mixedEligibleTestFile)); !os.IsNotExist(err) {
		t.Error("check mode wrote the eligible target's test")
	}
}
