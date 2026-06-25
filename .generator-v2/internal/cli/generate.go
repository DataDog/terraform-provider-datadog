package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// errCheckFailed signals that --check found files that would change.
// Execute translates this into exit code 3.
var errCheckFailed = fmt.Errorf("check: one or more files would change")

func newGenerateCmd(flags *globalFlags) *cobra.Command {
	var check bool
	var include string
	var specPath string
	var outputRoot string
	var hooksRoot string
	var trackingField string
	var maxDepth int
	var reportPath string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform artifacts from the OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			spec, err := parser.LoadSpec(specPath,
				parser.WithMaxDepth(maxDepth),
				parser.WithTrackingFieldName(trackingField))
			if err != nil {
				return err
			}

			runReport := model.RunReport{
				RunId:            uuid.NewString(),
				GeneratorVersion: cmd.Root().Version,
				SpecHash:         spec.Hash,
				StartedAt:        time.Now(),
			}

			filter := parseInclude(include)

			var registrations []emit.GeneratedRegistration
			for _, op := range spec.Operations {
				if op.Tracking == nil {
					runReport.SkippedOperations = append(runReport.SkippedOperations, model.SkippedOperation{
						OperationId: op.OperationId,
						Path:        op.Path,
						Method:      op.Method,
						Reason:      model.SkipReasonTrackingFieldAbsent,
					})
					continue
				}
				if op.Tracking.Skip {
					runReport.SkippedOperations = append(runReport.SkippedOperations, model.SkippedOperation{
						OperationId: op.OperationId,
						Path:        op.Path,
						Method:      op.Method,
						Reason:      model.SkipReasonTrackingFieldSkip,
					})
					continue
				}
				if len(filter) > 0 && !filter[op.Tracking.ArtifactName] {
					runReport.Artifacts = append(runReport.Artifacts, model.ArtifactReportEntry{
						Name:   op.Tracking.ArtifactName,
						Kind:   op.Tracking.ArtifactKind,
						Status: model.ArtifactStatusSkipped,
					})
					continue
				}

				entry, reg := generateArtifact(op, outputRoot, check)
				runReport.Artifacts = append(runReport.Artifacts, entry)
				if reg != nil {
					registrations = append(registrations, *reg)
				}
			}

			// Wire the generated data sources into the provider (register their
			// constructors, retire any they overwrite). Surface the result after
			// the report is written so a wiring I/O error still emits the report.
			wiringChanged, wiringErr := wireGeneratedDatasources(outputRoot, registrations, check)

			runReport.FinishedAt = time.Now()

			if err := runReport.Write(reportPath, cmd); err != nil {
				return err
			}

			if wiringErr != nil {
				return wiringErr
			}

			if runReport.Summary != nil && runReport.Summary.Failed > 0 {
				return fmt.Errorf("generate: %d artifact(s) failed; see report for details", runReport.Summary.Failed)
			}

			if check {
				for _, e := range runReport.Artifacts {
					if wouldChange(e.Status) {
						return errCheckFailed
					}
				}
				if wiringChanged {
					return errCheckFailed
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVar(&check, "check", false, "Read-only mode: exit 3 if any file would change")
	cmd.PersistentFlags().IntVar(&maxDepth, "max-depth", parser.DefaultMaxDepth, "Hard limit on recursive $ref expansion")
	cmd.PersistentFlags().StringVar(&specPath, "spec", ".generator/V2/openapi.yaml", "OpenAPI spec to read")
	cmd.PersistentFlags().StringVar(&include, "include", "", "Comma-separated artifact names to generate (empty = all)")
	cmd.PersistentFlags().StringVar(&outputRoot, "output-root", "datadog/fwprovider", "Root directory for generated artifacts")
	cmd.PersistentFlags().StringVar(&hooksRoot, "hooks-root", "datadog/fwprovider/hooks", "Root directory for hook subpackages")
	cmd.PersistentFlags().StringVar(&trackingField, "tracking-field", "x-datadog-tf-generator", "OpenAPI extension name for the tracking field")
	cmd.PersistentFlags().StringVar(&reportPath, "report", "-", "Where to write the run report (\"-\" = stdout)")

	return cmd
}

// generateArtifact runs the full model→emit→write pipeline for one tracked
// operation. On success it also returns the GeneratedRegistration the caller
// uses to wire the data source into the provider; it is nil for a skipped or
// failed artifact.
func generateArtifact(op *model.Operation, outputRoot string, check bool) (model.ArtifactReportEntry, *emit.GeneratedRegistration) {
	entry := model.ArtifactReportEntry{
		Name: op.Tracking.ArtifactName,
		Kind: op.Tracking.ArtifactKind,
	}

	if op.Tracking.ArtifactKind != model.ArtifactKindDataSource {
		entry.Status = model.ArtifactStatusSkipped
		entry.Diagnostics = []model.Diagnostic{{
			Severity: model.SeverityWarning,
			Message:  fmt.Sprintf("resource generation not yet supported (kind=%s)", op.Tracking.ArtifactKind),
		}}
		return entry, nil
	}

	artifact, err := model.BuildArtifact(op)
	if err != nil {
		return failEntry(entry, err), nil
	}
	artifact.SourceFile = filepath.Join(outputRoot, "data_source_datadog_"+artifact.Name+".go")
	entry.Path = artifact.SourceFile
	// Non-fatal notes (e.g. query params dropped from a plural filter set) ride
	// along on a successful entry; failEntry below overrides them on failure.
	entry.Diagnostics = append(entry.Diagnostics, artifact.Diagnostics...)

	view, err := emit.BuildDataSourceView(artifact)
	if err != nil {
		return failEntry(entry, err), nil
	}
	// Members the emit flattener dropped (e.g. relationships) ride along as info.
	for _, msg := range view.Dropped {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: model.SeverityInfo, Message: msg})
	}

	src, err := emit.RenderDataSource(view)
	if err != nil {
		return failEntry(entry, err), nil
	}

	status, err := emit.WriteFile(artifact.SourceFile, src, check)
	if err != nil {
		return failEntry(entry, err), nil
	}
	entry.Status = status
	return entry, &emit.GeneratedRegistration{
		Constructor: emit.DatasourceConstructor(artifact.Name),
		Overwrites:  op.Tracking.Overwrites,
	}
}

// wireGeneratedDatasources registers the run's generated data sources in the
// provider and retires the hand-written ones they overwrite. It rewrites the
// generatedDatasources slice with every generated constructor and, for each
// artifact whose spec set overwrites, removes the named hand-written constructor
// from the Datasources slice. It reports whether any of those files would change
// (so --check can fail) and honors check mode by not writing.
func wireGeneratedDatasources(outputRoot string, regs []emit.GeneratedRegistration, check bool) (changed bool, err error) {
	// A run that generated no data sources has nothing to register; leave the
	// provider files untouched rather than conjuring an empty generatedDatasources.
	if len(regs) == 0 {
		return false, nil
	}

	providerPath := filepath.Join(outputRoot, "framework_provider.go")
	constructors := make([]string, 0, len(regs))
	for _, reg := range regs {
		constructors = append(constructors, reg.Constructor)
		if reg.Overwrites == "" {
			continue
		}
		status, removeErr := emit.RemoveHandwrittenDatasource(providerPath, reg.Overwrites, check)
		if removeErr != nil {
			return changed, removeErr
		}
		changed = changed || wouldChange(status)
	}

	status, err := emit.SyncGeneratedDatasources(filepath.Join(outputRoot, "datasources_generated.go"), constructors, check)
	if err != nil {
		return changed, err
	}
	return changed || wouldChange(status), nil
}

// wouldChange reports whether a write status represents a file that was (or, in
// check mode, would be) modified.
func wouldChange(s model.ArtifactStatus) bool {
	return s == model.ArtifactStatusCreated || s == model.ArtifactStatusUpdated
}

func failEntry(e model.ArtifactReportEntry, err error) model.ArtifactReportEntry {
	e.Status = model.ArtifactStatusFailed
	e.Diagnostics = []model.Diagnostic{{Severity: model.SeverityError, Message: err.Error()}}
	return e
}

// parseInclude converts the --include flag value into a name-set for O(1) lookup.
// A nil map means "include all".

func parseInclude(s string) map[string]bool {
	if s == "" {
		return nil
	}
	m := make(map[string]bool)
	for _, name := range strings.Split(s, ",") {
		if n := strings.TrimSpace(name); n != "" {
			m[n] = true
		}
	}
	return m
}
