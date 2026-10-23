package model

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

// writeReport runs the real Write path to a temp file and returns its bytes,
// so these specs assert on what tfgen emits rather than on a struct.
func writeReport(r *RunReport) string {
	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	path := filepath.Join(GinkgoT().TempDir(), "report.json")
	Expect(r.Write(path, cmd)).To(Succeed())
	raw, err := os.ReadFile(path)
	Expect(err).To(Succeed())
	return strings.TrimSpace(string(raw))
}

func artifactOnlyReport() *RunReport {
	return &RunReport{
		RunId:            "run-1",
		GeneratorVersion: "dev",
		SpecHash:         "abc",
		StartedAt:        time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC),
		FinishedAt:       time.Date(2026, 9, 30, 12, 0, 1, 0, time.UTC),
		Artifacts: []ArtifactReportEntry{
			{Name: "team", Kind: ArtifactKindDataSource, Status: ArtifactStatusCreated, Path: "p"},
		},
	}
}

var _ = Describe("summarizeCassettes", func() {
	It("tallies each status into its own bucket", func() {
		got := summarizeCassettes([]CassetteResult{
			{Status: CassetteStatusGenerated},
			{Status: CassetteStatusGenerated},
			{Status: CassetteStatusPreserved},
			{Status: CassetteStatusIneligible},
			{Status: CassetteStatusIneligible},
			{Status: CassetteStatusIneligible},
		})
		Expect(*got).To(Equal(CassetteSummary{Generated: 2, Preserved: 1, Ineligible: 3}))
	})

	It("returns zeroed counts for no results", func() {
		Expect(*summarizeCassettes(nil)).To(Equal(CassetteSummary{}))
	})

	It("ignores a status it does not recognize rather than miscounting it", func() {
		got := summarizeCassettes([]CassetteResult{
			{Status: CassetteStatusGenerated}, {Status: "half_written"},
		})
		Expect(*got).To(Equal(CassetteSummary{Generated: 1}))
	})
})

var _ = Describe("RunReport.AddCassetteResult", func() {
	// Regeneration with unchanged inputs must produce identical bytes, and
	// that has to hold for the report as well as the fixtures it describes.
	It("keeps results sorted by name so report order is independent of processing order", func() {
		r := &RunReport{}
		for _, name := range []string{"widget", "api_key", "team"} {
			r.AddCassetteResult(CassetteResult{Name: name, Kind: ArtifactKindResource})
		}
		var names []string
		for _, c := range r.Cassettes {
			names = append(names, c.Name)
		}
		Expect(names).To(Equal([]string{"api_key", "team", "widget"}))
	})

	It("orders a resource and a data source sharing a name deterministically", func() {
		r := &RunReport{}
		r.AddCassetteResult(CassetteResult{Name: "team", Kind: ArtifactKindResource})
		r.AddCassetteResult(CassetteResult{Name: "team", Kind: ArtifactKindDataSource})
		Expect(r.Cassettes[0].Kind).To(Equal(ArtifactKindDataSource))
		Expect(r.Cassettes[1].Kind).To(Equal(ArtifactKindResource))
	})

	It("tolerates a nil report", func() {
		var r *RunReport
		Expect(func() { r.AddCassetteResult(CassetteResult{}) }).NotTo(Panic())
	})
})

var _ = Describe("RunReport.Write with cassettes", func() {
	// The whole constraint on this phase: a run that did not ask for cassettes
	// must emit exactly what it emitted before the feature existed.
	It("omits both cassette fields entirely when no target was considered", func() {
		content := writeReport(artifactOnlyReport())
		Expect(content).NotTo(ContainSubstring("cassette"))
		Expect(content).To(ContainSubstring(`"summary"`))
	})

	// omitempty only drops a nil pointer, so an eagerly-built zero summary
	// would leak a cassette_summary block into every existing report.
	It("leaves CassetteSummary nil rather than building a zeroed one", func() {
		r := artifactOnlyReport()
		_ = writeReport(r)
		Expect(r.CassetteSummary).To(BeNil())
	})

	It("emits the summary once a target is recorded", func() {
		r := artifactOnlyReport()
		r.AddCassetteResult(CassetteResult{
			Name: "team", Kind: ArtifactKindDataSource, TestName: "TestAccTeamOpenAPIExample",
			Status: CassetteStatusGenerated, WriteAction: CassetteWriteCreated,
		})
		r.AddCassetteResult(CassetteResult{
			Name: "widget", Kind: ArtifactKindResource, TestName: "TestAccWidgetOpenAPIExample",
			Status: CassetteStatusIneligible, WriteAction: CassetteWriteNone,
		})
		content := writeReport(r)

		var decoded RunReport
		Expect(json.Unmarshal([]byte(content), &decoded)).To(Succeed())
		Expect(decoded.CassetteSummary).NotTo(BeNil())
		Expect(*decoded.CassetteSummary).To(Equal(CassetteSummary{Generated: 1, Ineligible: 1}))
		Expect(decoded.Cassettes).To(HaveLen(2))

		// An ineligible cassette must not disturb the artifact tallies.
		Expect(decoded.Summary.Created).To(Equal(1))
		Expect(decoded.Summary.Failed).To(BeZero())
	})

	It("produces identical bytes for the same results added in a different order", func() {
		first, second := artifactOnlyReport(), artifactOnlyReport()
		a := CassetteResult{Name: "api_key", Kind: ArtifactKindDataSource,
			Status: CassetteStatusPreserved, WriteAction: CassetteWriteUnchanged}
		b := CassetteResult{Name: "widget", Kind: ArtifactKindResource,
			Status: CassetteStatusGenerated, WriteAction: CassetteWriteCreated}
		first.AddCassetteResult(a)
		first.AddCassetteResult(b)
		second.AddCassetteResult(b)
		second.AddCassetteResult(a)
		Expect(writeReport(first)).To(Equal(writeReport(second)))
	})

	It("validates against the embedded run-report contract", func() {
		r := artifactOnlyReport()
		r.AddCassetteResult(CassetteResult{
			Name: "widget", Kind: ArtifactKindResource, TestName: "TestAccWidgetOpenAPIExample",
			Status: CassetteStatusGenerated, WriteAction: CassetteWriteCreated,
			SelectedExample: ScenarioNameDefault,
		})
		content := writeReport(r)
		var decoded any
		Expect(json.Unmarshal([]byte(content), &decoded)).To(Succeed())
		Expect(validateAgainstContract(runReportSchema(), decoded)).To(Succeed())
	})
})
