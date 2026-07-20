package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
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

// apiInstancesHelperPath is the provider's ApiInstances helper, the source of
// truth for SDK API accessor names. It is read relative to the working directory
// (the repo root), like the other default paths.
const apiInstancesHelperPath = "datadog/internal/utils/api_instances_helper.go"

func newGenerateCmd(flags *globalFlags) *cobra.Command {
	var check bool
	var include string
	var specPath string
	var outputRoot string
	var hooksRoot string
	var trackingField string
	var maxDepth int
	var reportPath string
	var emitTests bool
	var testsOutputRoot string
	var examplesOutputRoot string
	var docsRoot string
	var reconcile bool
	var retire string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Terraform artifacts from the OpenAPI spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Orphan detection is only valid when tfgen sees the complete annotation
			// set, so --reconcile cannot be narrowed by --include.
			if reconcile && include != "" {
				return fmt.Errorf("generate: --reconcile cannot be combined with --include (orphan detection needs the complete annotation set)")
			}

			runReport := model.RunReport{
				RunId:            uuid.NewString(),
				GeneratorVersion: cmd.Root().Version,
				StartedAt:        time.Now(),
			}

			var wiringChanged bool
			var deferredErr error // surfaced after the report is written

			if retire != "" {
				// Scoped retirement: retire the named artifacts and stop, without
				// loading the spec or regenerating anything.
				for raw := range strings.SplitSeq(retire, ",") {
					name := strings.TrimSpace(raw)
					if name == "" {
						continue
					}
					runReport.Artifacts = append(runReport.Artifacts, retireArtifact(name, outputRoot, testsOutputRoot, docsRoot, examplesOutputRoot, check))
				}
			} else {
				spec, err := parser.LoadSpec(specPath,
					parser.WithMaxDepth(maxDepth),
					parser.WithTrackingFieldName(trackingField))
				if err != nil {
					return err
				}
				runReport.SpecHash = spec.Hash

				filter := parseInclude(include)

				// Resolve the provider's SDK-accessor names, the source of truth for
				// APIAccessor. A missing file is expected outside a full checkout; a parse
				// failure is worth surfacing. Either way we fall back to derived names.
				accessors, accErr := emit.ResolveAPIAccessors(apiInstancesHelperPath)
				if accErr != nil {
					if !errors.Is(accErr, os.ErrNotExist) {
						cmd.PrintErrln("tfgen: could not resolve API accessors, using derived names:", accErr)
					}
					accessors = nil
				}

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

					entry, testEntry, exampleEntry, reg := generateArtifact(op, outputRoot, testsOutputRoot, examplesOutputRoot, emitTests, check, accessors)
					runReport.Artifacts = append(runReport.Artifacts, entry)
					if testEntry != nil {
						runReport.Artifacts = append(runReport.Artifacts, *testEntry)
					}
					if exampleEntry != nil {
						runReport.Artifacts = append(runReport.Artifacts, *exampleEntry)
					}
					if reg != nil {
						registrations = append(registrations, *reg)
					}
				}

				// Wire the generated data sources into the provider (register their
				// constructors, retire any they overwrite). Surface the result after
				// the report is written so a wiring I/O error still emits the report.
				wiringChanged, deferredErr = wireGeneratedDatasources(outputRoot, testsOutputRoot, registrations, check)

				// Reconcile: retire generated data sources whose annotation is gone.
				// Runs after wiring so the registry already holds this run's set; skip
				// it if wiring failed, since the registry state is then uncertain.
				if reconcile && deferredErr == nil {
					// A failed artifact contributes no registration, so it is absent
					// from the desired set and reconcile would retire it as a false
					// orphan. Skip reconcile entirely when any artifact failed; the
					// failure is still surfaced below. This keeps reconcile fail-closed:
					// a transient build failure must never delete a live data source.
					failed := false
					for _, e := range runReport.Artifacts {
						if e.Status == model.ArtifactStatusFailed {
							failed = true
							break
						}
					}
					if failed {
						cmd.PrintErrln("tfgen: skipping --reconcile because one or more artifacts failed to generate (retiring orphans now could delete a data source that only failed this run)")
					} else {
						desired := make(map[string]bool, len(registrations))
						for _, reg := range registrations {
							desired[reg.Constructor] = true
						}
						orphanEntries, recErr := reconcileOrphans(outputRoot, testsOutputRoot, docsRoot, examplesOutputRoot, desired, check)
						runReport.Artifacts = append(runReport.Artifacts, orphanEntries...)
						deferredErr = recErr
					}
				}
			}

			runReport.FinishedAt = time.Now()

			if err := runReport.Write(reportPath, cmd); err != nil {
				return err
			}

			if deferredErr != nil {
				return deferredErr
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
	cmd.PersistentFlags().BoolVar(&emitTests, "emit-tests", false, "Also emit a generated acceptance-test scaffold for each data source")
	cmd.PersistentFlags().StringVar(&testsOutputRoot, "tests-output-root", "datadog/tests", "Root directory for generated acceptance-test files")
	cmd.PersistentFlags().StringVar(&examplesOutputRoot, "examples-output-root", "examples/data-sources", "Root directory for generated data-source examples")
	cmd.PersistentFlags().StringVar(&docsRoot, "docs-root", "docs/data-sources", "Root directory for data-source docs pages (used when retiring)")
	cmd.PersistentFlags().BoolVar(&reconcile, "reconcile", false, "Retire generated data sources no longer annotated (requires the full spec; incompatible with --include)")
	cmd.PersistentFlags().StringVar(&retire, "retire", "", "Comma-separated artifact names to retire without generating (deletes files + registration)")

	return cmd
}

// generateArtifact runs the full model→emit→write pipeline for one tracked
// operation. On success it also returns the GeneratedRegistration the caller
// uses to wire the data source into the provider; it is nil for a skipped or
// failed artifact.
func generateArtifact(op *model.Operation, outputRoot, testsOutputRoot, examplesOutputRoot string, emitTests, check bool, accessors map[string]string) (model.ArtifactReportEntry, *model.ArtifactReportEntry, *model.ArtifactReportEntry, *emit.GeneratedRegistration) {
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
		return entry, nil, nil, nil
	}

	artifact, err := model.BuildArtifact(op)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}
	artifact.SourceFile = filepath.Join(outputRoot, "data_source_datadog_"+artifact.Name+".go")
	entry.Path = artifact.SourceFile
	// Non-fatal notes (e.g. query params dropped from a plural filter set) ride
	// along on a successful entry; failEntry below overrides them on failure.
	entry.Diagnostics = append(entry.Diagnostics, artifact.Diagnostics...)

	view, err := emit.BuildDataSourceView(artifact)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}
	// Correct APIAccessor to the name the provider's helper actually exposes,
	// which diverges from the derived name for a few acronym/aliased APIs.
	emit.ApplyAPIAccessor(&view, accessors)
	// Members the emit flattener dropped (e.g. relationships) ride along as info.
	for _, msg := range view.Dropped {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: model.SeverityInfo, Message: msg})
	}

	src, err := emit.RenderDataSource(view)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}

	status, err := emit.WriteFile(artifact.SourceFile, src, check)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}
	entry.Status = status

	exampleEntry := emitDatasourceExample(&entry, view, artifact.Name, examplesOutputRoot, check)

	var testEntry *model.ArtifactReportEntry
	if emitTests {
		testEntry = emitDatasourceTest(&entry, view, artifact.Name, testsOutputRoot, check)
	}

	reg := &emit.GeneratedRegistration{
		Constructor: emit.DatasourceConstructor(artifact.Name),
		Overwrites:  op.Tracking.Overwrites,
	}
	// A generated test must also be registered in testFiles2EndpointTags or it
	// t.Fatals at startup. Carry the key + tag only when a test was emitted, so a
	// run without --emit-tests never touches provider_test.go. Fall back to the
	// artifact name when the operation has no OpenAPI tag, so the span tag is never
	// blank.
	if testEntry != nil {
		reg.TestFileKey = emit.EndpointTagTestKey(artifact.Name)
		reg.EndpointTag = emit.NormalizeEndpointTag(op.Tag)
		if reg.EndpointTag == "" {
			reg.EndpointTag = artifact.Name
		}
	}

	return entry, testEntry, exampleEntry, reg
}

// emitDatasourceExample writes the tfplugindocs input for a data source. Like
// acceptance-test scaffolds, examples are created once and never overwritten so
// a hand-written or subsequently improved example remains authoritative.
func emitDatasourceExample(entry *model.ArtifactReportEntry, view emit.DataSourceView, name, examplesOutputRoot string, check bool) *model.ArtifactReportEntry {
	path := filepath.Join(examplesOutputRoot, "datadog_"+name, "data-source.tf")
	example := emit.RenderDataSourceExample(view)
	status, err := emit.WriteFileIfAbsent(path, example.Content, check)
	if err != nil {
		failed := failEntry(model.ArtifactReportEntry{Name: name, Kind: model.ArtifactKindDataSource, Path: path}, err)
		return &failed
	}
	// if the on-disk file still byte-matches what we would generate, nobody has touched the
	// placeholder so we should keep the diagnostics so an incomplete example is still flagged
	// rather than silently suppressed.
	untouchedPlaceholder := false
	if status == model.ArtifactStatusSkipped {
		if onDisk, readErr := os.ReadFile(path); readErr == nil {
			untouchedPlaceholder = bytes.Equal(onDisk, example.Content)
		}
	}
	if status != model.ArtifactStatusSkipped || untouchedPlaceholder {
		entry.Diagnostics = append(entry.Diagnostics, example.Diagnostics...)
	}
	return &model.ArtifactReportEntry{Name: name, Kind: model.ArtifactKindDataSource, Status: status, Path: path}
}

// emitDatasourceTest renders and writes the acceptance-test scaffold for a data
// source. It is best-effort: a render or write problem is recorded as a warning
// on the data source's own entry rather than failing the run, since the test is
// a scaffold and the data source is the real artifact. On a successful write it
// returns a report entry for the test file so --check sees it and the summary
// counts it; an existing file at the path is left untouched and reported as
// skipped, because the scaffold is completed by hand and must not be clobbered.
func emitDatasourceTest(entry *model.ArtifactReportEntry, view emit.DataSourceView, name, testsOutputRoot string, check bool) *model.ArtifactReportEntry {
	src, err := emit.RenderDataSourceTest(view)
	if err != nil {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: model.SeverityWarning, Message: fmt.Sprintf("test scaffold not generated: %v", err)})
		return nil
	}

	path := filepath.Join(testsOutputRoot, "data_source_datadog_"+name+"_test.go")
	status, err := emit.WriteFileIfAbsent(path, src, check)
	if err != nil {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: model.SeverityWarning, Message: fmt.Sprintf("test scaffold write failed: %v", err)})
		return nil
	}
	if status == model.ArtifactStatusSkipped {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: model.SeverityInfo, Message: fmt.Sprintf("test scaffold skipped: %s already exists (edit it by hand)", path)})
	}
	return &model.ArtifactReportEntry{Name: name, Kind: entry.Kind, Status: status, Path: path}
}

// wireGeneratedDatasources registers the run's generated data sources in the
// provider and retires the hand-written ones they overwrite. It rewrites the
// generatedDatasources slice with every generated constructor and, for each
// artifact whose spec set overwrites, removes the named hand-written constructor
// from the Datasources slice. Each generated test is also registered in
// provider_test.go's testFiles2EndpointTags map (under testsOutputRoot). It reports
// whether any of those files would change (so --check can fail) and honors check
// mode by not writing.
func wireGeneratedDatasources(outputRoot, testsOutputRoot string, regs []emit.GeneratedRegistration, check bool) (changed bool, err error) {
	// A run that generated no data sources has nothing to register; leave the
	// provider files untouched rather than conjuring an empty generatedDatasources.
	if len(regs) == 0 {
		return false, nil
	}

	providerPath := filepath.Join(outputRoot, "framework_provider.go")
	genPath := filepath.Join(outputRoot, "datasources_generated.go")
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
		// RemoveHandwrittenDatasource reports Unchanged only when the target was
		// not in the framework Datasources slice. That is expected on a re-run
		// where a prior run already retired it (its replacement is registered),
		// but otherwise means the target never existed — a typo, or an SDKv2
		// DataSourcesMap entry the generator cannot retire. Fail loudly so a
		// mis-targeted overwrite is caught here rather than as a mux conflict.
		if status == model.ArtifactStatusUnchanged {
			already, regErr := emit.GeneratedDatasourceRegistered(genPath, reg.Constructor)
			if regErr != nil {
				return changed, regErr
			}
			if !already {
				return changed, fmt.Errorf(
					"generate: overwrites target %q not found in the framework Datasources slice (%s); the generator can only retire hand-written framework data sources, not SDKv2 entries in provider.go's DataSourcesMap",
					reg.Overwrites, providerPath)
			}
		}
		changed = changed || wouldChange(status)
	}

	status, err := emit.SyncGeneratedDatasources(genPath, constructors, check)
	if err != nil {
		return changed, err
	}
	changed = changed || wouldChange(status)

	// Register each generated test in provider_test.go's testFiles2EndpointTags
	// map. Only regs with a test emitted this run carry a TestFileKey.
	providerTestPath := filepath.Join(testsOutputRoot, "provider_test.go")
	for _, reg := range regs {
		if reg.TestFileKey == "" {
			continue
		}
		tagStatus, tagErr := emit.InsertEndpointTag(providerTestPath, reg.TestFileKey, reg.EndpointTag, check)
		if tagErr != nil {
			return changed, tagErr
		}
		changed = changed || wouldChange(tagStatus)
	}

	return changed, nil
}

// wouldChange reports whether a write status represents a file that was (or, in
// check mode, would be) modified. A retirement deletes files and a registration
// retirement rewrites the registry, so both count; retire_blocked leaves
// everything in place, so it does not.
func wouldChange(s model.ArtifactStatus) bool {
	return s == model.ArtifactStatusCreated || s == model.ArtifactStatusUpdated ||
		s == model.ArtifactStatusRetired || s == model.ArtifactStatusRegistrationRetired
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
