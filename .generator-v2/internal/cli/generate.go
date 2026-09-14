package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/providermod"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/sdkbind"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/sdkbinding"
)

// errCheckFailed signals that --check found files that would change; Execute
// maps it to exit code 3.
var errCheckFailed = fmt.Errorf("check: one or more files would change")

// apiInstancesHelperRelPath points at the provider's ApiInstances helper, which
// names the SDK API accessors. It is relative to --output-root, not to the
// working directory.
const apiInstancesHelperRelPath = "../internal/utils/api_instances_helper.go"

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
		Use:               "generate",
		Short:             "Generate Terraform artifacts from the OpenAPI spec",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Orphan detection needs the complete annotation set, so --reconcile
			// cannot be narrowed by --include.
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

				// Resolve the provider's helper accessors first: they take precedence
				// over derived names, keeping cached clients and aliased spellings
				// (RUM, APM, Observability Pipelines) as the initialization path.
				accessors, accErr := emit.ResolveAPIAccessors(filepath.Join(outputRoot, apiInstancesHelperRelPath))
				if accErr != nil {
					cmd.PrintErrln("tfgen: could not resolve provider API accessors; SDK constructor names will be derived from OpenAPI tags:", accErr)
					accessors = nil
				}

				// The pinned SDK is best-effort corroboration only, so a cold cache or
				// missing package warns and generation continues from the
				// authoritative OpenAPI derivation.
				var sdkBindings *sdkbinding.Inventory
				sdkPackageDir, sdkDirErr := providermod.SDKPackageDir(outputRoot)
				if sdkDirErr != nil {
					cmd.PrintErrln("tfgen: pinned SDK corroboration unavailable; generating from OpenAPI bindings:", sdkDirErr)
				} else {
					sdkBindings, err = sdkbinding.Load(sdkPackageDir)
					if err != nil {
						cmd.PrintErrln("tfgen: pinned SDK corroboration unavailable; generating from OpenAPI bindings:", err)
						sdkBindings = nil
					}
				}

				var registrations []emit.GeneratedRegistration
				var resourceRegistrations []emit.GeneratedRegistration
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

					if op.Tracking.ArtifactKind == model.ArtifactKindResource {
						entry, reg := generateResourceArtifact(op, outputRoot, check, accessors, sdkBindings)
						runReport.Artifacts = append(runReport.Artifacts, entry)
						if reg != nil {
							resourceRegistrations = append(resourceRegistrations, *reg)
						}
						continue
					}

					entry, testEntry, exampleEntry, reg := generateArtifact(op, outputRoot, testsOutputRoot, examplesOutputRoot, emitTests, check, accessors, sdkBindings)
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

				// Register the generated constructors and retire any they overwrite.
				// Errors are deferred so a wiring failure still emits the report.
				wiringChanged, deferredErr = wireGeneratedDatasources(outputRoot, testsOutputRoot, registrations, check)
				if deferredErr == nil {
					var resourceWiringChanged bool
					resourceWiringChanged, deferredErr = wireGeneratedResources(outputRoot, resourceRegistrations, check)
					wiringChanged = wiringChanged || resourceWiringChanged
				}
				// One registry for both kinds: the SDK gates a beta endpoint per
				// operation, so an x-unstable GET behind a data source is enabled
				// exactly like a resource's create.
				if deferredErr == nil {
					var unstableChanged bool
					unstableChanged, deferredErr = wireUnstableOperations(
						outputRoot, check, slices.Concat(registrations, resourceRegistrations))
					wiringChanged = wiringChanged || unstableChanged
				}

				// Retire generated data sources whose annotation is gone. Runs after
				// wiring so the registry holds this run's set, and is skipped if
				// wiring failed, since the registry state is then uncertain.
				if reconcile && deferredErr == nil {
					// A failed artifact contributes no registration, so reconcile
					// would see it as an orphan and delete a live data source. Skip
					// reconcile entirely when anything failed; the failure is still
					// surfaced below.
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

// generateArtifact builds, renders and writes one tracked data-source
// operation, returning its report entry, the optional test and example entries,
// and its registration. The registration is nil for a failed artifact.
func generateArtifact(op *model.Operation, outputRoot, testsOutputRoot, examplesOutputRoot string, emitTests, check bool, accessors map[string]string, sdkBindings *sdkbinding.Inventory) (model.ArtifactReportEntry, *model.ArtifactReportEntry, *model.ArtifactReportEntry, *emit.GeneratedRegistration) {
	entry := model.ArtifactReportEntry{
		Name: op.Tracking.ArtifactName,
		Kind: op.Tracking.ArtifactKind,
	}

	// Bind the SDK oneOf wrapper, members and constructors before projection,
	// which copies those bindings onto each envelope rather than deriving them.
	// A search operation distinct from the read op is bound alongside it.
	bindOps := []*model.Operation{op}
	if searchOp := op.ResolvedGroup.Op(model.GroupRoleSearch); searchOp != nil && searchOp != op {
		bindOps = append(bindOps, searchOp)
	}
	bindingDiagnostics, err := bindOperations(bindOps, sdkBindings)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}

	artifact, err := model.BuildArtifact(op)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}
	artifact.SourceFile = filepath.Join(outputRoot, "data_source_datadog_"+artifact.Name+".go")
	entry.Path = artifact.SourceFile
	// Non-fatal notes (e.g. query params dropped from a plural filter set) ride
	// along on a successful entry; failEntry below overrides them on failure.
	entry.Diagnostics = append(entry.Diagnostics, bindingDiagnostics...)
	entry.Diagnostics = append(entry.Diagnostics, artifact.Diagnostics...)

	view, err := emit.BuildDataSourceView(artifact)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}
	// Prefer the provider's existing helper accessor (including acronym/alias
	// spellings), then derive the same constructor the SDK generator will emit.
	if err := emit.ApplyAPIAccessor(&view, accessors); err != nil {
		return failEntry(entry, err), nil, nil, nil
	}
	// Members the emit flattener dropped (e.g. relationships) ride along at the
	// severity the drop warrants.
	for _, d := range view.Dropped {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: d.Severity, Message: d.Message})
	}

	src, err := emit.RenderDataSource(view)
	if err != nil {
		return failEntry(entry, err), nil, nil, nil
	}

	status, err := emit.WriteArtifactSource(artifact.SourceFile, src, check, op.Tracking.Overwrites != "")
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
		Constructor:        emit.DatasourceConstructor(artifact.Name),
		Overwrites:         op.Tracking.Overwrites,
		UnstableOperations: artifact.UnstableOperations,
	}
	entry.Diagnostics = append(entry.Diagnostics, unstableOperationDiagnostics(artifact)...)
	// A generated test must appear in testFiles2EndpointTags or it t.Fatals at
	// startup. Set only when a test was emitted, so a run without --emit-tests
	// never touches provider_test.go; the artifact name stands in for a missing
	// OpenAPI tag so the tag is never blank.
	if testEntry != nil {
		reg.TestFileKey = emit.EndpointTagTestKey(artifact.Name)
		reg.EndpointTag = emit.NormalizeEndpointTag(op.Tag)
		if reg.EndpointTag == "" {
			reg.EndpointTag = artifact.Name
		}
	}

	return entry, testEntry, exampleEntry, reg
}

// generateResourceArtifact builds, renders and writes one tracked resource
// operation, returning its report entry and its registration. The registration
// is nil for a failed artifact.
func generateResourceArtifact(op *model.Operation, outputRoot string, check bool, accessors map[string]string, sdkBindings *sdkbinding.Inventory) (model.ArtifactReportEntry, *emit.GeneratedRegistration) {
	entry := model.ArtifactReportEntry{
		Name: op.Tracking.ArtifactName,
		Kind: op.Tracking.ArtifactKind,
	}

	// Bind each role of the CRUD quad independently: a oneOf can appear in any
	// of the Create/Read/Update bodies the merge later unions. A groupless op
	// falls back to binding itself.
	roleOps := op.ResolvedGroup.Operations(
		model.GroupRoleCreate, model.GroupRoleRead, model.GroupRoleUpdate, model.GroupRoleDelete)
	if len(roleOps) == 0 {
		roleOps = []*model.Operation{op}
	}
	bindingDiagnostics, err := bindOperations(roleOps, sdkBindings)
	if err != nil {
		return failEntry(entry, err), nil
	}
	entry.Diagnostics = append(entry.Diagnostics, bindingDiagnostics...)

	artifact, err := model.BuildArtifact(op)
	if err != nil {
		return failEntry(entry, err), nil
	}
	artifact.SourceFile = filepath.Join(outputRoot, "resource_datadog_"+artifact.Name+".go")
	entry.Path = artifact.SourceFile
	entry.Diagnostics = append(entry.Diagnostics, artifact.Diagnostics...)

	view, err := emit.BuildResourceView(artifact)
	if err != nil {
		return failEntry(entry, err), nil
	}
	if err := emit.ApplyResourceAPIAccessor(&view, accessors); err != nil {
		return failEntry(entry, err), nil
	}
	for _, d := range view.Dropped {
		entry.Diagnostics = append(entry.Diagnostics, model.Diagnostic{Severity: d.Severity, Message: d.Message})
	}

	src, err := emit.RenderResource(view)
	if err != nil {
		return failEntry(entry, err), nil
	}

	status, err := emit.WriteArtifactSource(artifact.SourceFile, src, check, op.Tracking.Overwrites != "")
	if err != nil {
		return failEntry(entry, err), nil
	}
	entry.Status = status

	reg := &emit.GeneratedRegistration{
		Constructor:        emit.ResourceConstructor(artifact.Name),
		Overwrites:         op.Tracking.Overwrites,
		UnstableOperations: artifact.UnstableOperations,
	}
	entry.Diagnostics = append(entry.Diagnostics, unstableOperationDiagnostics(artifact)...)
	return entry, reg
}

// bindOperations resolves the SDK bindings for each operation in ops and
// returns their non-fatal diagnostics. The first unresolvable union aborts and
// returns its error, failing only the artifact being generated.
func bindOperations(ops []*model.Operation, inv *sdkbinding.Inventory) ([]model.Diagnostic, error) {
	var diags []model.Diagnostic
	for _, op := range ops {
		if err := sdkbind.BindOperation(op); err != nil {
			return nil, err
		}
		opDiags, err := sdkbinding.Bind(op, inv)
		if err != nil {
			return nil, err
		}
		diags = append(diags, opDiags...)
	}
	return diags, nil
}

// unstableOperationDiagnostics returns one info diagnostic naming the unstable
// operations the artifact needs enabled, or nil when they are all stable.
func unstableOperationDiagnostics(artifact *model.Artifact) []model.Diagnostic {
	if len(artifact.UnstableOperations) == 0 {
		return nil
	}
	return []model.Diagnostic{{
		Severity: model.SeverityInfo,
		Message: fmt.Sprintf(
			"enabled %d unstable operation(s) the artifact calls: %s",
			len(artifact.UnstableOperations), strings.Join(artifact.UnstableOperations, ", ")),
	}}
}

// emitDatasourceExample writes a data source's example .tf. Written only when
// absent, so a hand-edited example is never overwritten.
func emitDatasourceExample(entry *model.ArtifactReportEntry, view emit.DataSourceView, name, examplesOutputRoot string, check bool) *model.ArtifactReportEntry {
	path := filepath.Join(examplesOutputRoot, "datadog_"+name, "data-source.tf")
	example := emit.RenderDataSourceExample(view)
	status, err := emit.WriteFileIfAbsent(path, example.Content, check)
	if err != nil {
		failed := failEntry(model.ArtifactReportEntry{Name: name, Kind: model.ArtifactKindDataSource, Path: path}, err)
		return &failed
	}
	// A skipped file that still byte-matches what we would generate is an
	// untouched placeholder, so keep its diagnostics instead of suppressing
	// them as belonging to someone's hand-written example.
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

// emitDatasourceTest renders and writes a data source's acceptance-test
// scaffold, returning a report entry for it. Best-effort: a render or write
// problem becomes a warning on the data source's own entry and returns nil. An
// existing file is left untouched and reported as skipped, since the scaffold
// is completed by hand.
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

// generatedKind parameterizes wireGenerated over data sources and resources,
// which share the registration shape — rewrite the tfgen-owned slice, retire
// each overwritten hand-written constructor — and differ only by these hooks.
type generatedKind struct {
	// sliceName is the hand-written framework slice the kind retires from, as
	// framework_provider.go names it.
	sliceName string
	// genFileName is the tfgen-owned registry file's base name.
	genFileName string
	// retireHint completes the "can only retire hand-written framework ..."
	// error, which names the kind and any registry the generator cannot reach.
	retireHint string

	removeHandwritten func(path, constructor string, check bool) (model.ArtifactStatus, error)
	registered        func(path string) ([]string, error)
	sync              func(path string, constructors []string, check bool) (model.ArtifactStatus, error)
}

var (
	datasourceKind = generatedKind{
		sliceName:         "Datasources",
		genFileName:       "datasources_generated.go",
		retireHint:        "data sources, not SDKv2 entries in provider.go's DataSourcesMap",
		removeHandwritten: emit.RemoveHandwrittenDatasource,
		registered:        emit.RegisteredGeneratedDatasources,
		sync:              emit.SyncGeneratedDatasources,
	}
	resourceKind = generatedKind{
		sliceName:         "Resources",
		genFileName:       "resources_generated.go",
		retireHint:        "resources",
		removeHandwritten: emit.RemoveHandwrittenResource,
		registered:        emit.RegisteredGeneratedResources,
		sync:              emit.SyncGeneratedResources,
	}
)

// wireGenerated rewrites the tfgen-owned registry slice with every generated
// constructor of one kind and removes each hand-written constructor an artifact
// overwrites from the framework slice. It reports whether any file was (or, in
// check mode, would be) changed; check mode writes nothing.
func wireGenerated(outputRoot string, k generatedKind, regs []emit.GeneratedRegistration, check bool) (changed bool, err error) {
	// A run that generated nothing of this kind has nothing to register; leave
	// the provider files untouched rather than conjuring an empty slice.
	if len(regs) == 0 {
		return false, nil
	}

	providerPath := filepath.Join(outputRoot, "framework_provider.go")
	genPath := filepath.Join(outputRoot, k.genFileName)

	// The registered set only changes when sync runs below, so read it once
	// rather than re-scanning the registry per overwriting artifact.
	registered, err := k.registered(genPath)
	if err != nil {
		return false, err
	}

	constructors := make([]string, 0, len(regs))
	for _, reg := range regs {
		constructors = append(constructors, reg.Constructor)
		if reg.Overwrites == "" {
			continue
		}
		status, removeErr := k.removeHandwritten(providerPath, reg.Overwrites, check)
		if removeErr != nil {
			return changed, removeErr
		}
		// Unchanged means the target was not in the framework slice: expected on
		// a re-run that already retired it (its replacement is registered),
		// otherwise the target never existed. Fail rather than let a
		// mis-targeted overwrite surface later as a mux conflict.
		if status == model.ArtifactStatusUnchanged && !slices.Contains(registered, reg.Constructor) {
			return changed, fmt.Errorf(
				"generate: overwrites target %q not found in the framework %s slice (%s); the generator can only retire hand-written framework %s",
				reg.Overwrites, k.sliceName, providerPath, k.retireHint)
		}
		changed = changed || wouldChange(status)
	}

	status, err := k.sync(genPath, constructors, check)
	if err != nil {
		return changed, err
	}
	return changed || wouldChange(status), nil
}

// wireGeneratedDatasources runs the shared registration for data sources, then
// records each generated test in provider_test.go's testFiles2EndpointTags map
// under testsOutputRoot.
func wireGeneratedDatasources(outputRoot, testsOutputRoot string, regs []emit.GeneratedRegistration, check bool) (changed bool, err error) {
	changed, err = wireGenerated(outputRoot, datasourceKind, regs, check)
	if err != nil {
		return changed, err
	}

	// Only regs whose test was emitted this run carry a TestFileKey.
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

// wireGeneratedResources registers each generated resource constructor in
// resources_generated.go and drops any hand-written resource it overwrites from
// the framework Resources slice.
func wireGeneratedResources(outputRoot string, regs []emit.GeneratedRegistration, check bool) (changed bool, err error) {
	return wireGenerated(outputRoot, resourceKind, regs, check)
}

// wireUnstableOperations merges every artifact's x-unstable operation keys into
// unstable_operations_generated.go. A run with no such keys leaves the file
// alone rather than creating an empty one.
func wireUnstableOperations(outputRoot string, check bool, regs []emit.GeneratedRegistration) (changed bool, err error) {
	var keys []string
	for _, reg := range regs {
		keys = append(keys, reg.UnstableOperations...)
	}
	if len(keys) == 0 {
		return false, nil
	}

	path := filepath.Join(outputRoot, "unstable_operations_generated.go")
	status, err := emit.SyncUnstableOperations(path, keys, check)
	if err != nil {
		return false, err
	}
	return status != model.ArtifactStatusUnchanged, nil
}

// wouldChange reports whether a status means a file was (or, in check mode,
// would be) modified. Retired and registration-retired count, since they
// delete files and rewrite the registry; retire_blocked changes nothing.
func wouldChange(s model.ArtifactStatus) bool {
	return s == model.ArtifactStatusCreated || s == model.ArtifactStatusUpdated ||
		s == model.ArtifactStatusRetired || s == model.ArtifactStatusRegistrationRetired
}

func failEntry(e model.ArtifactReportEntry, err error) model.ArtifactReportEntry {
	e.Status = model.ArtifactStatusFailed
	e.Diagnostics = []model.Diagnostic{{Severity: model.SeverityError, Message: err.Error()}}
	return e
}

// parseInclude converts the --include value into a name set. A nil map means
// "include all".

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
