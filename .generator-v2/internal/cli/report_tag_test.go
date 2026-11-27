package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func TestReportUsesDecodedNamesAndRegistryOwner(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("..", "testdata", "mini-oas", "scripts", "gen-test", "teams.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// Operations still belong to Teams. The registry owner is deliberately
	// different, and a metadata-only tag precedes it.
	source := string(fixture)
	index := strings.LastIndex(source, "\ntags:\n")
	if index < 0 {
		t.Fatal("fixture has no tag registry")
	}
	for _, tc := range []struct {
		name, family, override, want string
		include                      string
		invalid                      bool
	}{
		{name: "derived plural", family: "policy", want: "policies"},
		{name: "explicit override", family: "policy", override: "custom", want: "custom"},
		{name: "maximum length override", family: "policy", override: strings.Repeat("a", 64), want: strings.Repeat("a", 64)},
		{name: "filtered artifact", family: "policy", want: "policies", include: "other"},
		{name: "too long derived plural", family: strings.Repeat("a", 64), invalid: true},
		{name: "invalid family", family: "bad-name", invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := source[:index] + `
tags:
- name: Metadata Only
  x-datadog-tf-generator:
    X-owner: unrelated
    x-fake: {datasource_list: {group: {read: MissingOperation}}}
- name: Registry Owner
  x-datadog-tf-generator:
    x-owner: api-platform
    X-labels: [metadata]
    x-enabled: true
    x-null: null
    ignored: {skip: true}
    ` + tc.family + `:
      datasource_list:
        group: {read: ListTeams}
`
			if tc.override != "" {
				body += "        artifact_name: " + tc.override + "\n"
			}
			dir := t.TempDir()
			specPath, reportPath := filepath.Join(dir, "spec.yaml"), filepath.Join(dir, "report.json")
			if err := os.WriteFile(specPath, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"generate", "--spec", specPath, "--report", reportPath,
				"--output-root", filepath.Join(dir, "out"), "--examples-output-root", filepath.Join(dir, "examples")}
			if tc.include != "" {
				args = append(args, "--include", tc.include)
			}
			err := runTfgen(args...)
			if tc.invalid {
				if err == nil {
					t.Fatal("invalid name should fail before report generation")
				}
				if _, statErr := os.Stat(reportPath); !os.IsNotExist(statErr) {
					t.Fatalf("unexpected report: %v", statErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var report model.RunReport
			if err := json.Unmarshal([]byte(mustRead(t, reportPath)), &report); err != nil {
				t.Fatal(err)
			}
			if len(report.Artifacts) == 0 {
				t.Fatal("no artifact reported")
			}
			found := false
			for _, entry := range report.Artifacts {
				if entry.Tag != "Registry Owner" {
					t.Errorf("wrong registry owner: %+v", entry)
				}
				if entry.Name == tc.want {
					found = true
					wantStatus := model.ArtifactStatusCreated
					if tc.include != "" {
						wantStatus = model.ArtifactStatusSkipped
					}
					if entry.Status != wantStatus {
						t.Errorf("status %s, want %s", entry.Status, wantStatus)
					}
				}
			}
			if !found {
				t.Fatalf("decoded name %q absent: %+v", tc.want, report.Artifacts)
			}
		})
	}
}
