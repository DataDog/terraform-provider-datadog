package model

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RunReport functions

// summarize tallies artifact entries by status into the convenience counts. The
// five buckets correspond one-to-one to the ArtifactStatus values.
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
		}
	}
	return s
}

// openReportWriter returns a writer for the run report and a cleanup function.
// "-" maps to the command's stdout; anything else is opened as a file.
func (r *RunReport) Write(path string, cmd *cobra.Command) error {
	writer := cmd.OutOrStdout()
	closeFunc := func() error { return nil }

	if path != "-" {
		f, err := os.Create(path)
		writer = f
		closeFunc = f.Close

		if err != nil {
			return fmt.Errorf("report: opening %s: %w", path, err)
		}
	}
	defer closeFunc()

	r.Summary = summarize(r.Artifacts)

	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(&r); err != nil {
		return fmt.Errorf("report: encoding run report: %w", err)
	}
	return nil
}
