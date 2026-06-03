// Package report serializes a model.RunReport into the structured JSON consumed
// by CI (to gate regeneration drift), by maintainers (to audit per-artifact
// status), and by downstream tooling.
package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// WriteJSON encodes report as indented JSON to writer.
//
// The summary block is derived from the artifact statuses rather than trusted
// from the caller, so its counts always agree with the artifacts array. CI
// will eventually gates on these counts, so they must not drift.
// The caller's report is not mutated.
//
// Output is deterministic: field order is fixed and a RunReport carries no maps,
// so identical reports serialize to identical bytes.
func WriteJSON(writer io.Writer, report *model.RunReport) error {
	if report == nil {
		return fmt.Errorf("report: nil RunReport")
	}

	out := *report // shallow copy: set Summary for output without touching the caller's report
	out.Summary = summarize(report.Artifacts)

	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(&out); err != nil {
		return fmt.Errorf("report: encoding run report: %w", err)
	}
	return nil
}

// summarize tallies artifact entries by status into the convenience counts. The
// five buckets correspond one-to-one to the ArtifactStatus values.
func summarize(entries []model.ArtifactReportEntry) *model.RunSummary {
	s := &model.RunSummary{}
	for _, e := range entries {
		switch e.Status {
		case model.ArtifactStatusCreated:
			s.Created++
		case model.ArtifactStatusUpdated:
			s.Updated++
		case model.ArtifactStatusUnchanged:
			s.Unchanged++
		case model.ArtifactStatusSkipped:
			s.Skipped++
		case model.ArtifactStatusFailed:
			s.Failed++
		}
	}
	return s
}
