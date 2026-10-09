package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestDefaultCassetteWriteFailureFailsRun(t *testing.T) {
	for _, check := range []bool{false, true} {
		name := "generate"
		if check {
			name = "check"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			spec := mutatedSpec(t, dir, "        cassette: true\n", "")
			testsDir := filepath.Join(dir, "tests")
			seedProviderTest(t, testsDir)
			testPath := filepath.Join(testsDir, "resource_datadog_integration_twilio_account_openapi_example_test.go")
			handwritten := []byte("package test\n// handwritten\n")
			if err := os.WriteFile(testPath, handwritten, 0o644); err != nil {
				t.Fatal(err)
			}
			reportPath := filepath.Join(dir, "report.json")
			var extra []string
			if check {
				extra = []string{"--check"}
			}
			err := runGenerateWithSpec(t, spec, dir, reportPath, extra...)
			if err == nil || errors.Is(err, errCheckFailed) {
				t.Fatalf("write failure must fail the run, not succeed or report drift: %v", err)
			}
			report := readCassetteReport(t, reportPath)
			if len(report.Cassettes) != 1 {
				t.Fatalf("expected one cassette result: %+v", report.Cassettes)
			}
			result := report.Cassettes[0]
			if result.Status != model.CassetteStatusIneligible || result.WriteAction != model.CassetteWriteNone {
				t.Fatalf("unexpected write failure outcome: %+v", result)
			}
			foundWrite := false
			for _, diagnostic := range result.Diagnostics {
				foundWrite = foundWrite || diagnostic.Category == model.DiagnosticCategoryWrite
			}
			if !foundWrite {
				t.Fatalf("missing write diagnostic: %+v", result.Diagnostics)
			}
			contents, err := os.ReadFile(testPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(contents, handwritten) {
				t.Fatal("handwritten file was changed")
			}
			registration, err := os.ReadFile(filepath.Join(testsDir, "provider_test.go"))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(registration, []byte("integration_twilio_account_openapi_example")) {
				t.Fatal("failed test was registered")
			}
		})
	}
}
