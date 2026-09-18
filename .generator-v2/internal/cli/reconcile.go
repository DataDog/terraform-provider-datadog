package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// testAccFuncRe captures an exported acceptance-test function name, e.g.
// "TestAccDatadogTeamDataSource" — also the prefix of its cassette's filename.
var testAccFuncRe = regexp.MustCompile(`func (TestAcc[A-Za-z0-9_]+)`)

// artifactNameRe is the safe charset for an artifact name. retireArtifact
// rejects a non-matching name before building any path from it, so a --retire
// value can never traverse (../) or inject path metacharacters.
var artifactNameRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// retireArtifact deletes a generated data source's .go, _test.go, docs page,
// example and registration, reporting retired. Two guards report retire_blocked
// instead: the .go must carry the tfgen generated-code marker, and no recorded
// cassette may exist for its test (a cassette means it was adopted, so removal
// would be breaking). Check mode deletes and writes nothing.
func retireArtifact(name, outputRoot, testsOutputRoot, docsRoot, examplesOutputRoot string, check bool) model.ArtifactReportEntry {
	// Fail closed on an unsafe name before any path is built from it: an invalid
	// name deletes nothing.
	if !artifactNameRe.MatchString(name) {
		return failEntry(model.ArtifactReportEntry{Name: name, Kind: model.ArtifactKindDataSource},
			fmt.Errorf("refusing to retire %q: invalid artifact name (must match %s)", name, artifactNameRe.String()))
	}

	goPath := filepath.Join(outputRoot, "data_source_datadog_"+name+".go")
	testPath := filepath.Join(testsOutputRoot, "data_source_datadog_"+name+"_test.go")
	docPath := filepath.Join(docsRoot, name+".md")
	examplePath := filepath.Join(examplesOutputRoot, "datadog_"+name, "data-source.tf")
	genPath := filepath.Join(outputRoot, "datasources_generated.go")
	cassettesDir := filepath.Join(testsOutputRoot, "cassettes")

	entry := model.ArtifactReportEntry{Name: name, Kind: model.ArtifactKindDataSource, Path: goPath}

	src, err := os.ReadFile(goPath)
	switch {
	case err == nil:
		if !emit.IsGeneratedSource(src) {
			entry.Status = model.ArtifactStatusRetireBlocked
			entry.Diagnostics = []model.Diagnostic{{Severity: model.SeverityWarning,
				Message: fmt.Sprintf("refusing to retire %s: not tfgen-generated (missing the generated-code marker)", goPath)}}
			return entry
		}
	case errors.Is(err, os.ErrNotExist):
		// No source to guard; still clean up any stray test/doc/example/registration below.
	default:
		return failEntry(entry, err)
	}

	if adopted, cassette := hasRecordedCassette(testPath, cassettesDir); adopted {
		entry.Status = model.ArtifactStatusRetireBlocked
		entry.Diagnostics = []model.Diagnostic{{Severity: model.SeverityWarning,
			Message: fmt.Sprintf("not retiring %q: recorded cassette %q shows it was adopted; needs a manual deprecation cycle", name, cassette)}}
		return entry
	}

	if !check {
		for _, p := range []string{goPath, testPath, docPath, examplePath} {
			if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
				return failEntry(entry, err)
			}
		}
		if err := removeDirIfEmpty(filepath.Dir(examplePath)); err != nil {
			return failEntry(entry, err)
		}
	}
	if _, err := emit.RemoveGeneratedDatasource(genPath, emit.DatasourceConstructor(name), check); err != nil {
		return failEntry(entry, err)
	}
	// Drop the test's testFiles2EndpointTags entry; a missing provider_test.go
	// is tolerated.
	providerTestPath := filepath.Join(testsOutputRoot, "provider_test.go")
	if _, err := emit.RemoveEndpointTag(providerTestPath, emit.EndpointTagTestKey(name), check); err != nil {
		return failEntry(entry, err)
	}
	entry.Status = model.ArtifactStatusRetired
	return entry
}

// removeDirIfEmpty removes path only when it holds no entries, so an example
// directory with files hand-added alongside data-source.tf survives. A missing
// directory is not an error.
func removeDirIfEmpty(path string) error {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return nil
	}
	return os.Remove(path)
}

// hasRecordedCassette reports whether a cassette exists for any TestAcc
// function declared in the test file at testPath, returning its name. A
// cassette's basename is its test function, optionally with an underscore
// suffix, and only regular .yaml/.freeze files count. A missing test file or
// cassettes directory reports no adoption.
func hasRecordedCassette(testPath, cassettesDir string) (bool, string) {
	data, err := os.ReadFile(testPath)
	if err != nil {
		return false, ""
	}
	var funcs []string
	for _, m := range testAccFuncRe.FindAllStringSubmatch(string(data), -1) {
		funcs = append(funcs, m[1])
	}
	if len(funcs) == 0 {
		return false, ""
	}
	entries, err := os.ReadDir(cassettesDir)
	if err != nil {
		return false, ""
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".yaml" && ext != ".freeze" {
			continue
		}
		base := strings.TrimSuffix(e.Name(), ext)
		for _, fn := range funcs {
			if base == fn || strings.HasPrefix(base, fn+"_") {
				return true, e.Name()
			}
		}
	}
	return false, ""
}

// reconcileOrphans retires every data source still registered in
// generatedDatasources but absent from desired (the constructors this run
// registered), returning one report entry each. An orphan constructor is mapped
// back to its artifact name through the files on disk, since the
// name→constructor transform is not reversible.
func reconcileOrphans(outputRoot, testsOutputRoot, docsRoot, examplesOutputRoot string, desired map[string]bool, check bool) ([]model.ArtifactReportEntry, error) {
	genPath := filepath.Join(outputRoot, "datasources_generated.go")
	registered, err := emit.RegisteredGeneratedDatasources(genPath)
	if err != nil {
		return nil, err
	}
	ctorToName, err := generatedArtifactNames(outputRoot)
	if err != nil {
		return nil, err
	}

	var entries []model.ArtifactReportEntry
	for _, ctor := range registered {
		if desired[ctor] {
			continue
		}
		name, ok := ctorToName[ctor]
		if !ok {
			// Registered with no marker'd file to name it: drop the stale
			// registration only, since there are no files to delete.
			if _, err := emit.RemoveGeneratedDatasource(genPath, ctor, check); err != nil {
				return entries, err
			}
			entries = append(entries, model.ArtifactReportEntry{
				Name: emit.RegistrationRetirementName(ctor), Kind: model.ArtifactKindDataSource,
				Status: model.ArtifactStatusRegistrationRetired, Constructor: ctor,
				Diagnostics: []model.Diagnostic{{Severity: model.SeverityWarning,
					Message: fmt.Sprintf("registered constructor %q has no generated file; removed the stale registration only", ctor)}},
			})
			continue
		}
		entries = append(entries, retireArtifact(name, outputRoot, testsOutputRoot, docsRoot, examplesOutputRoot, check))
	}
	return entries, nil
}

// generatedArtifactNames maps constructor to artifact name by scanning
// outputRoot for data_source_datadog_*.go files carrying the tfgen
// generated-code marker. Hand-written files are excluded, so a resolved name's
// files are always safe to delete.
func generatedArtifactNames(outputRoot string) (map[string]string, error) {
	matches, err := filepath.Glob(filepath.Join(outputRoot, "data_source_datadog_*.go"))
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if !emit.IsGeneratedSource(data) {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "data_source_datadog_"), ".go")
		out[emit.DatasourceConstructor(name)] = name
	}
	return out, nil
}
