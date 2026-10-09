package model

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

// RunReport functions

// summarize tallies artifact entries by status into the convenience counts. The
// buckets correspond one-to-one to the ArtifactStatus values.
func summarize(entries []ArtifactReportEntry) *RunSummary {
	s := &RunSummary{}
	for _, e := range entries {
		switch e.Status {
		case ArtifactStatusCreated:
			s.Created++
		case ArtifactStatusUpdated:
			s.Updated++
		case ArtifactStatusUnchanged:
			s.Unchanged++
		case ArtifactStatusSkipped:
			s.Skipped++
		case ArtifactStatusFailed:
			s.Failed++
		case ArtifactStatusRetired:
			s.Retired++
		case ArtifactStatusRetireBlocked:
			s.RetireBlocked++
		case ArtifactStatusRegistrationRetired:
			s.RegistrationRetired++
		}
	}
	return s
}

// summarizeCassettes tallies cassette results by status. The buckets correspond
// one-to-one to the CassetteStatus values, and are kept apart from RunSummary
// so an ineligible cassette never shows up in the artifact counts CI already
// asserts on: a target whose examples cannot form a fixture is not a failed
// artifact, and per FR-012 it must not prevent other targets from being
// reported as generated.
func summarizeCassettes(results []CassetteResult) *CassetteSummary {
	s := &CassetteSummary{}
	for _, r := range results {
		switch r.Status {
		case CassetteStatusGenerated:
			s.Generated++
		case CassetteStatusPreserved:
			s.Preserved++
		case CassetteStatusIneligible:
			s.Ineligible++
		}
	}
	return s
}

// AddCassetteResult records one target's outcome. Results are kept sorted by
// (name, kind) so the report reads the same whatever order the generator
// happened to process targets in — regeneration with unchanged inputs must
// produce identical bytes, and that has to hold for the report as well as for
// the fixtures it describes.
func (r *RunReport) AddCassetteResult(result CassetteResult) {
	if r == nil {
		return
	}
	r.Cassettes = append(r.Cassettes, result)
	slices.SortStableFunc(r.Cassettes, func(a, b CassetteResult) int {
		if byName := cmp.Compare(a.Name, b.Name); byName != 0 {
			return byName
		}
		return cmp.Compare(string(a.Kind), string(b.Kind))
	})
}

// openReportWriter returns a writer for the run report and a cleanup function.
// "-" maps to the command's stdout; anything else is opened as a file.
func (r *RunReport) Write(path string, cmd *cobra.Command) error {
	writer := cmd.OutOrStdout()
	closeFunc := func() error { return nil }

	if path != "-" {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("report: opening %s: %w", path, err)
		}

		writer = f
		closeFunc = f.Close
	}
	defer closeFunc()

	r.Summary = summarize(r.Artifacts)
	// Left nil when no target was considered, so cassette_summary is omitted
	// entirely rather than emitted as a block of zeros. omitempty only drops a
	// nil pointer, so this guard is what keeps a run that did not ask for
	// cassettes byte-identical to one from before the feature existed.
	if len(r.Cassettes) > 0 {
		r.CassetteSummary = summarizeCassettes(r.Cassettes)
	}

	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(&r); err != nil {
		return fmt.Errorf("report: encoding run report: %w", err)
	}
	return nil
}
