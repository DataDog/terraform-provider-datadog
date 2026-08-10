// Package split turns one aggregate tfgen push — all N generated data sources on
// a single branch — into per-artifact bundles, one directory per data source, so
// each can land as its own PR. It is pure: it diffs the provider, docs and
// examples paths in two on-disk trees (a master checkout and the pushed-branch
// checkout), reuses tfgen's registry serializer to reconstruct each artifact's
// datasources_generated.go as "the base set plus this one artifact", and never
// touches git, the network, or the OpenAPI spec.
//
// Attribution is exact and fail-loud: a changed file that maps to no artifact and
// is not a known shared registry file is a hard error, never silently dropped, so
// a future shared helper file can't vanish. Retirements require the upstream
// generation report and are emitted as explicit file-removal plans.
package split

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// Options configures a split run.
type Options struct {
	// BaseDir is a checkout of the base branch (master) — the tree the push is diffed against.
	BaseDir string
	// GeneratedDir is a checkout of the pushed branch carrying all N generated artifacts.
	GeneratedDir string
	// OutDir receives one <name>/ subdirectory per artifact, each holding the exact
	// files to commit at their repo-relative paths.
	OutDir string
	// GenerationReport is the upstream tfgen generate report. When supplied, it
	// authorizes and cross-checks retirement and retire_blocked routing.
	GenerationReport string
	// Check computes the plan and report without writing any output bundle.
	Check bool
}

// Result is the JSON output of a split run.
type Result struct {
	BaseDir      string     `json:"base_dir"`
	GeneratedDir string     `json:"generated_dir"`
	OutDir       string     `json:"out_dir"`
	Artifacts    []Artifact `json:"artifacts"`
	// Errors holds attribution problems that failed the run; empty on success.
	Errors []string `json:"errors,omitempty"`
}

// Artifact is one routed data source: the files materialized under OutDir/<name>/
// (repo-relative), and how it changed relative to base.
type Artifact struct {
	Name         string               `json:"name"`
	Status       model.ArtifactStatus `json:"status"`
	Files        []string             `json:"files"`
	RemovedFiles []string             `json:"removed_files,omitempty"`
	Diagnostics  []model.Diagnostic   `json:"diagnostics,omitempty"`
}

// WriteJSON encodes the report as indented JSON.
func (r *Result) WriteJSON(w io.Writer) error {
	sort.Slice(r.Artifacts, func(i, j int) bool { return r.Artifacts[i].Name < r.Artifacts[j].Name })
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(r)
}

// The known shared files (rewritten per artifact) and the generated data-source
// file name pattern. datasources_generated.go and framework_provider.go are shared:
// reconstructed for each branch, never attributed to a single artifact by presence.
const (
	registryFile     = "datasources_generated.go"
	providerFile     = "framework_provider.go"
	providerTestFile = "provider_test.go"
)

var (
	dataSourceFileRe       = regexp.MustCompile(`^data_source_datadog_([a-z][a-z0-9_]*)\.go$`)
	dataSourceTestFileRe   = regexp.MustCompile(`^data_source_datadog_([a-z][a-z0-9_]*)_test\.go$`)
	dataSourceDocFileRe    = regexp.MustCompile(`^([a-z][a-z0-9_]*)\.md$`)
	dataSourceExampleDirRe = regexp.MustCompile(`^datadog_([a-z][a-z0-9_]*)$`)
	artifactNameRe         = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// constructorDeclRe recovers the constructor(s) a data-source file declares.
	constructorDeclRe = regexp.MustCompile(`func (New[A-Za-z0-9_]+DataSource)\(`)
	// constructorRe recovers constructor identifiers listed inside a slice literal.
	constructorRe = regexp.MustCompile(`New[A-Za-z0-9_]+DataSource`)
	// Directory scopes are walked in full. Test routing is added separately by
	// scopedFiles so unrelated acceptance-test and cassette drift is ignored.
	diffDirectoryScopes = []string{
		filepath.FromSlash("datadog/fwprovider"),
		filepath.FromSlash("docs/data-sources"),
		filepath.FromSlash("examples/data-sources"),
	}
)

// datasourcesSliceHeader opens the hand-written Datasources slice in
// framework_provider.go — the block an overwrite retires a constructor from.
// Matches emit's constant of the same name.
const datasourcesSliceHeader = "var Datasources = []func() datasource.DataSource{"

type fileKind int

const (
	kindUnknown fileKind = iota
	kindDataSource
	kindDataSourceTest
	kindRegistry
	kindProvider
	kindProviderTest
	kindDataSourceDoc
	kindDataSourceExample
)

// classify maps an exact repo-relative path to its role in the split.
func classify(rel string) (fileKind, string) {
	clean := filepath.Clean(rel)
	base := filepath.Base(clean)
	switch clean {
	case filepath.Join("datadog", "fwprovider", registryFile):
		return kindRegistry, ""
	case filepath.Join("datadog", "fwprovider", providerFile):
		return kindProvider, ""
	case filepath.Join("datadog", "tests", providerTestFile):
		return kindProviderTest, ""
	}
	if filepath.Dir(clean) == filepath.Join("datadog", "fwprovider") {
		if m := dataSourceFileRe.FindStringSubmatch(base); m != nil {
			return kindDataSource, m[1]
		}
	}
	if filepath.Dir(clean) == filepath.Join("datadog", "tests") {
		if m := dataSourceTestFileRe.FindStringSubmatch(base); m != nil {
			return kindDataSourceTest, m[1]
		}
	}
	if filepath.Dir(clean) == filepath.Join("docs", "data-sources") {
		if m := dataSourceDocFileRe.FindStringSubmatch(base); m != nil {
			return kindDataSourceDoc, m[1]
		}
	}
	if base == "data-source.tf" && filepath.Dir(filepath.Dir(clean)) == filepath.Join("examples", "data-sources") {
		if m := dataSourceExampleDirRe.FindStringSubmatch(filepath.Base(filepath.Dir(clean))); m != nil {
			return kindDataSourceExample, m[1]
		}
	}
	return kindUnknown, ""
}

// changedFile is a file differing between base and generated.
type changedFile struct {
	rel   string // repo-relative path
	added bool   // absent in base (net-new) vs. modified
}

// artifactPlan gathers the self-contained files discovered for one data source.
type artifactPlan struct {
	source  *changedFile
	test    *changedFile
	doc     *changedFile
	example *changedFile
	removed []string
	report  *generationArtifact
}

type generationArtifact struct {
	status      model.ArtifactStatus
	diagnostics []model.Diagnostic
	// constructor is set only for registration_retired artifacts; it is the
	// registration to drop, since no file names the artifact.
	constructor string
}

// Split diffs the two trees, routes each artifact into OutDir/<name>/, and returns
// a report. Attribution problems are recorded in report.Errors and also returned
// as an error (so the caller exits non-zero); the report is still complete.
func Split(opts Options) (*Result, error) {
	rep := &Result{BaseDir: opts.BaseDir, GeneratedDir: opts.GeneratedDir, OutDir: opts.OutDir}

	reportArtifacts := map[string]generationArtifact{}
	if opts.GenerationReport != "" {
		var err error
		reportArtifacts, err = readGenerationArtifacts(opts.GenerationReport)
		if err != nil {
			return rep, fmt.Errorf("split: reading generation report: %w", err)
		}
	}

	changed, deleted, err := diffTrees(opts.BaseDir, opts.GeneratedDir)
	if err != nil {
		return rep, err
	}

	var problems []string
	fail := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }

	plans := map[string]*artifactPlan{}
	var registry, provider, providerTest *changedFile
	for i := range changed {
		cf := &changed[i]
		kind, name := classify(cf.rel)
		switch kind {
		case kindRegistry:
			registry = cf
		case kindProvider:
			provider = cf
		case kindProviderTest:
			providerTest = cf
		case kindDataSource:
			planFor(plans, name).source = cf
		case kindDataSourceTest:
			planFor(plans, name).test = cf
		case kindDataSourceDoc:
			planFor(plans, name).doc = cf
		case kindDataSourceExample:
			planFor(plans, name).example = cf
		default:
			fail("changed file maps to no artifact and is not a known shared file: %s", cf.rel)
		}
	}
	for _, d := range deleted {
		kind, name := classify(d)
		switch kind {
		case kindDataSource, kindDataSourceTest, kindDataSourceDoc, kindDataSourceExample:
			planFor(plans, name).removed = append(planFor(plans, name).removed, d)
		default:
			fail("removed file maps to no artifact or is a shared file that cannot be deleted: %s", d)
		}
	}
	for name, reported := range reportArtifacts {
		p := planFor(plans, name)
		reportedCopy := reported
		p.report = &reportedCopy
	}

	if len(plans) == 0 {
		// Nothing to split; a shared-file change with no data source is a problem.
		if registry != nil {
			fail("%s changed but no data source files changed", registry.rel)
		}
		return finish(rep, problems)
	}

	for _, name := range sortedNames(plans) {
		p := plans[name]
		status, statusProblems := validatePlan(name, p, opts.GenerationReport != "")
		problems = append(problems, statusProblems...)
		if p.report == nil && status != "" {
			p.report = &generationArtifact{status: status}
		}
	}

	registryRel := filepath.Join("datadog", "fwprovider", registryFile)
	baseSet, err := emit.RegisteredGeneratedDatasources(filepath.Join(opts.BaseDir, registryRel))
	if err != nil {
		return rep, fmt.Errorf("split: reading base registry: %w", err)
	}

	overwrites, provProblems := attributeOverwrites(opts, plans, provider)
	problems = append(problems, provProblems...)

	incoming, err := emit.RegisteredGeneratedDatasources(filepath.Join(opts.GeneratedDir, registryRel))
	if err != nil {
		return rep, fmt.Errorf("split: reading generated registry: %w", err)
	}
	want := slices.Clone(baseSet)
	prProducing := 0
	for name, p := range plans {
		if p.report == nil {
			continue
		}
		switch p.report.status {
		case model.ArtifactStatusCreated, model.ArtifactStatusUpdated:
			want = append(want, emit.DatasourceConstructor(name))
			prProducing++
		case model.ArtifactStatusRetired:
			want = without(want, emit.DatasourceConstructor(name))
			prProducing++
		case model.ArtifactStatusRegistrationRetired:
			want = without(want, p.report.constructor)
			prProducing++
		}
	}
	registrySetChanged := len(onlyIn(want, baseSet)) > 0 || len(onlyIn(baseSet, want)) > 0
	if prProducing > 0 && registrySetChanged && registry == nil {
		fail("PR-producing data source changes require %s to change", registryRel)
	}
	if diff := onlyIn(incoming, want); len(diff) > 0 {
		fail("%s registers constructor(s) not accounted for by the generation report: %s", registryRel, strings.Join(diff, ", "))
	}
	if diff := onlyIn(want, incoming); len(diff) > 0 {
		fail("generation report expects constructor(s) missing from incoming %s: %s", registryRel, strings.Join(diff, ", "))
	}

	providerTestRel := filepath.Join("datadog", "tests", providerTestFile)
	baseTags, err := emit.RegisteredEndpointTags(filepath.Join(opts.BaseDir, providerTestRel))
	if err != nil {
		return rep, fmt.Errorf("split: reading base endpoint tags: %w", err)
	}
	incomingTags, err := emit.RegisteredEndpointTags(filepath.Join(opts.GeneratedDir, providerTestRel))
	if err != nil {
		return rep, fmt.Errorf("split: reading generated endpoint tags: %w", err)
	}
	wantTags := cloneMap(baseTags)
	for name, p := range plans {
		if p.report == nil {
			continue
		}
		key := emit.EndpointTagTestKey(name)
		switch p.report.status {
		case model.ArtifactStatusCreated, model.ArtifactStatusUpdated:
			if p.test != nil {
				tag, ok := incomingTags[key]
				if !ok {
					fail("%s changed for %q, but %s has no %q entry", p.test.rel, name, providerTestRel, key)
				} else {
					wantTags[key] = tag
				}
			}
		case model.ArtifactStatusRetired:
			delete(wantTags, key)
		}
	}
	if !mapsEqual(wantTags, incomingTags) {
		fail("%s changes are not fully attributable to the reported artifacts", providerTestRel)
	}
	if !mapsEqual(baseTags, incomingTags) && providerTest == nil {
		fail("%s content changed but was not discovered in the scoped diff", providerTestRel)
	}

	for _, name := range sortedNames(plans) {
		p := plans[name]
		if p.report == nil {
			continue
		}
		var art *Artifact
		var artProblems []string
		switch p.report.status {
		case model.ArtifactStatusCreated, model.ArtifactStatusUpdated:
			art, artProblems = materializeActive(opts, name, p, registryRel, baseSet, overwrites[name], provider, providerTestRel, incomingTags)
		case model.ArtifactStatusRetired:
			art, artProblems = materializeRetired(opts, name, p, registryRel, providerTestRel, baseTags)
		case model.ArtifactStatusRegistrationRetired:
			art, artProblems = materializeRegistrationRetired(opts, name, p, registryRel)
		case model.ArtifactStatusRetireBlocked:
			art = &Artifact{
				Name:        name,
				Status:      model.ArtifactStatusRetireBlocked,
				Files:       []string{},
				Diagnostics: slices.Clone(p.report.diagnostics),
			}
		}
		problems = append(problems, artProblems...)
		if art != nil {
			rep.Artifacts = append(rep.Artifacts, *art)
		}
	}

	return finish(rep, problems)
}

func materializeActive(opts Options, name string, p *artifactPlan, registryRel string, baseSet, overwritten []string, provider *changedFile, providerTestRel string, incomingTags map[string]string) (*Artifact, []string) {
	var problems []string
	art := &Artifact{
		Name:        name,
		Status:      p.report.status,
		Files:       []string{},
		Diagnostics: slices.Clone(p.report.diagnostics),
	}
	outDir := filepath.Join(opts.OutDir, name)

	var files []string
	if p.source != nil {
		files = append(files, p.source.rel)
	}
	if p.test != nil {
		files = append(files, p.test.rel)
	}
	if p.example != nil {
		files = append(files, p.example.rel)
	}
	for _, rel := range files {
		if !opts.Check {
			if err := copyFile(filepath.Join(opts.GeneratedDir, rel), filepath.Join(outDir, rel)); err != nil {
				problems = append(problems, fmt.Sprintf("copying %s for %q: %v", rel, name, err))
			}
		}
	}
	art.Files = append(art.Files, files...)

	// Reconstruct the registry from the base set plus this one constructor. Writing
	// to a fresh path unions with nothing, so the result is exactly base + this.
	constructors := append(slices.Clone(baseSet), emit.DatasourceConstructor(name))
	if _, err := emit.SyncGeneratedDatasources(filepath.Join(outDir, registryRel), constructors, opts.Check); err != nil {
		problems = append(problems, fmt.Sprintf("reconstructing %s for %q: %v", registryRel, name, err))
	}
	art.Files = append(art.Files, registryRel)

	if len(overwritten) > 0 {
		provOut := filepath.Join(outDir, provider.rel)
		if !opts.Check {
			if err := copyFile(filepath.Join(opts.BaseDir, provider.rel), provOut); err != nil {
				problems = append(problems, fmt.Sprintf("copying %s for %q: %v", provider.rel, name, err))
			}
			for _, c := range overwritten {
				if _, err := emit.RemoveHandwrittenDatasource(provOut, c, false); err != nil {
					problems = append(problems, fmt.Sprintf("retiring %s from %s for %q: %v", c, provider.rel, name, err))
				}
			}
		}
		art.Files = append(art.Files, provider.rel)
	}

	if p.test != nil {
		key := emit.EndpointTagTestKey(name)
		tag, ok := incomingTags[key]
		if !ok {
			problems = append(problems, fmt.Sprintf("reconstructing %s for %q: endpoint tag %q is absent", providerTestRel, name, key))
		} else {
			testOut := filepath.Join(outDir, providerTestRel)
			if !opts.Check {
				if err := copyFile(filepath.Join(opts.BaseDir, providerTestRel), testOut); err != nil {
					problems = append(problems, fmt.Sprintf("copying %s for %q: %v", providerTestRel, name, err))
				} else if _, err := emit.InsertEndpointTag(testOut, key, tag, false); err != nil {
					problems = append(problems, fmt.Sprintf("reconstructing %s for %q: %v", providerTestRel, name, err))
				}
			}
			art.Files = append(art.Files, providerTestRel)
		}
	}

	sort.Strings(art.Files)
	art.Files = slices.Compact(art.Files)
	return art, problems
}

func materializeRetired(opts Options, name string, p *artifactPlan, registryRel, providerTestRel string, baseTags map[string]string) (*Artifact, []string) {
	var problems []string
	art := &Artifact{
		Name:         name,
		Status:       model.ArtifactStatusRetired,
		Files:        []string{registryRel},
		RemovedFiles: slices.Clone(p.removed),
		Diagnostics:  slices.Clone(p.report.diagnostics),
	}
	outDir := filepath.Join(opts.OutDir, name)
	registryOut := filepath.Join(outDir, registryRel)
	if !opts.Check {
		if err := copyFile(filepath.Join(opts.BaseDir, registryRel), registryOut); err != nil {
			problems = append(problems, fmt.Sprintf("copying %s for %q: %v", registryRel, name, err))
		} else if _, err := emit.RemoveGeneratedDatasource(registryOut, emit.DatasourceConstructor(name), false); err != nil {
			problems = append(problems, fmt.Sprintf("reconstructing %s for retirement %q: %v", registryRel, name, err))
		}
	}

	key := emit.EndpointTagTestKey(name)
	if _, ok := baseTags[key]; ok {
		testOut := filepath.Join(outDir, providerTestRel)
		if !opts.Check {
			if err := copyFile(filepath.Join(opts.BaseDir, providerTestRel), testOut); err != nil {
				problems = append(problems, fmt.Sprintf("copying %s for %q: %v", providerTestRel, name, err))
			} else if _, err := emit.RemoveEndpointTag(testOut, key, false); err != nil {
				problems = append(problems, fmt.Sprintf("reconstructing %s for retirement %q: %v", providerTestRel, name, err))
			}
		}
		art.Files = append(art.Files, providerTestRel)
	}

	sort.Strings(art.Files)
	art.Files = slices.Compact(art.Files)
	sort.Strings(art.RemovedFiles)
	art.RemovedFiles = slices.Compact(art.RemovedFiles)
	return art, problems
}

// materializeRegistrationRetired reconstructs the sole change of a
// registration-only retirement: dropping the orphan's constructor from the
// generated registry. Its files were already gone, so nothing else is touched
// and the constructor — not the artifact name — drives the edit.
func materializeRegistrationRetired(opts Options, name string, p *artifactPlan, registryRel string) (*Artifact, []string) {
	var problems []string
	art := &Artifact{
		Name:        name,
		Status:      model.ArtifactStatusRegistrationRetired,
		Files:       []string{registryRel},
		Diagnostics: slices.Clone(p.report.diagnostics),
	}
	if !opts.Check {
		registryOut := filepath.Join(opts.OutDir, name, registryRel)
		if err := copyFile(filepath.Join(opts.BaseDir, registryRel), registryOut); err != nil {
			problems = append(problems, fmt.Sprintf("copying %s for %q: %v", registryRel, name, err))
		} else if _, err := emit.RemoveGeneratedDatasource(registryOut, p.report.constructor, false); err != nil {
			problems = append(problems, fmt.Sprintf("reconstructing %s for registration retirement %q: %v", registryRel, name, err))
		}
	}
	return art, problems
}

func validatePlan(name string, p *artifactPlan, reportRequired bool) (model.ArtifactStatus, []string) {
	var problems []string
	fail := func(format string, a ...any) { problems = append(problems, fmt.Sprintf(format, a...)) }

	if !artifactNameRe.MatchString(name) {
		fail("artifact name %q from diff/report is unsafe", name)
		return "", problems
	}
	if reportRequired && p.report == nil {
		fail("artifact %q changed but is absent from the generation report", name)
		return "", problems
	}

	status := model.ArtifactStatus("")
	if p.report != nil {
		status = p.report.status
	} else if p.source != nil {
		if p.source.added {
			status = model.ArtifactStatusCreated
		} else {
			status = model.ArtifactStatusUpdated
		}
	}

	switch status {
	case model.ArtifactStatusCreated:
		if p.source == nil || !p.source.added {
			fail("generation report marks %q created, but its source file is not newly added", name)
		}
		if len(p.removed) > 0 {
			fail("created artifact %q also removes files: %s", name, strings.Join(p.removed, ", "))
		}
	case model.ArtifactStatusUpdated:
		if p.source == nil && p.test == nil && p.example == nil {
			fail("generation report marks %q updated, but no source, test or example file changed", name)
		}
		if p.source != nil && p.source.added {
			fail("generation report marks %q updated, but its source file is newly added", name)
		}
		if len(p.removed) > 0 {
			fail("updated artifact %q also removes files: %s", name, strings.Join(p.removed, ", "))
		}
	case model.ArtifactStatusRetired:
		if p.source != nil || p.test != nil || p.doc != nil || p.example != nil {
			fail("retired artifact %q contains added or modified self-contained files", name)
		}
	case model.ArtifactStatusRetireBlocked:
		if p.source != nil || p.test != nil || p.doc != nil || p.example != nil || len(p.removed) > 0 {
			fail("retire_blocked artifact %q unexpectedly changes files", name)
		}
	case model.ArtifactStatusRegistrationRetired:
		if p.source != nil || p.test != nil || p.doc != nil || len(p.removed) > 0 {
			fail("registration_retired artifact %q unexpectedly changes files", name)
		}
	case "":
		if len(p.removed) > 0 {
			fail("removed file for %q requires a generation report declaring retired", name)
		} else {
			fail("artifact %q has no routable source change or report status", name)
		}
	default:
		fail("artifact %q has unsupported generation status %q", name, status)
	}
	if p.doc != nil {
		fail("documentation file changed upstream for %q (%s); docs must be generated per artifact in CI", name, p.doc.rel)
	}
	return status, problems
}

func readGenerationArtifacts(path string) (map[string]generationArtifact, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var report model.RunReport
	if err := json.NewDecoder(f).Decode(&report); err != nil {
		return nil, err
	}

	type accumulated struct {
		mainSeen       bool
		mainStatus     model.ArtifactStatus
		testChanged    bool
		exampleChanged bool
		diagnostics    []model.Diagnostic
		constructor    string
	}
	acc := map[string]*accumulated{}
	for _, entry := range report.Artifacts {
		if entry.Kind != model.ArtifactKindDataSource {
			continue
		}
		if !artifactNameRe.MatchString(entry.Name) {
			if entry.Status == model.ArtifactStatusRetired ||
				entry.Status == model.ArtifactStatusRetireBlocked ||
				entry.Status == model.ArtifactStatusRegistrationRetired {
				return nil, fmt.Errorf("unsafe retired artifact name %q", entry.Name)
			}
			continue
		}
		a := acc[entry.Name]
		if a == nil {
			a = &accumulated{}
			acc[entry.Name] = a
		}
		base := filepath.Base(entry.Path)
		if base == "data_source_datadog_"+entry.Name+"_test.go" {
			if entry.Status == model.ArtifactStatusCreated || entry.Status == model.ArtifactStatusUpdated {
				a.testChanged = true
			}
			continue
		}
		if base == "data-source.tf" && filepath.Base(filepath.Dir(entry.Path)) == "datadog_"+entry.Name {
			if entry.Status == model.ArtifactStatusCreated || entry.Status == model.ArtifactStatusUpdated {
				a.exampleChanged = true
			}
			continue
		}
		if base != "" && base != "data_source_datadog_"+entry.Name+".go" &&
			entry.Status != model.ArtifactStatusRetired &&
			entry.Status != model.ArtifactStatusRetireBlocked &&
			entry.Status != model.ArtifactStatusRegistrationRetired {
			continue
		}
		if a.mainSeen {
			return nil, fmt.Errorf("duplicate generation report entry for artifact %q", entry.Name)
		}
		a.mainSeen = true
		a.mainStatus = entry.Status
		a.diagnostics = slices.Clone(entry.Diagnostics)
		a.constructor = entry.Constructor
	}

	out := map[string]generationArtifact{}
	for name, a := range acc {
		if !a.mainSeen {
			if a.testChanged || a.exampleChanged {
				return nil, fmt.Errorf("generation report has a changed auxiliary file for %q without a data-source entry", name)
			}
			continue
		}
		status := a.mainStatus
		if status == model.ArtifactStatusUnchanged && (a.testChanged || a.exampleChanged) {
			status = model.ArtifactStatusUpdated
		}
		switch status {
		case model.ArtifactStatusCreated, model.ArtifactStatusUpdated, model.ArtifactStatusRetired, model.ArtifactStatusRetireBlocked, model.ArtifactStatusRegistrationRetired:
			out[name] = generationArtifact{status: status, diagnostics: a.diagnostics, constructor: a.constructor}
		case model.ArtifactStatusUnchanged, model.ArtifactStatusSkipped:
			continue
		case model.ArtifactStatusFailed:
			return nil, fmt.Errorf("generation report contains failed artifact %q", name)
		default:
			return nil, fmt.Errorf("generation report contains unknown status %q for %q", status, name)
		}
	}
	return out, nil
}

// attributeOverwrites maps each hand-written constructor removed from
// framework_provider.go to the artifact whose base source file declared it (an
// overwrite generates in place at the hand-written data source's path). A removed
// constructor no artifact claims, or an addition to the hand-written slice, is a
// problem.
func attributeOverwrites(opts Options, plans map[string]*artifactPlan, provider *changedFile) (map[string][]string, []string) {
	overwrites := map[string][]string{}
	if provider == nil {
		return overwrites, nil
	}
	var problems []string

	baseCtors, err := handwrittenConstructors(filepath.Join(opts.BaseDir, provider.rel))
	if err != nil {
		return overwrites, []string{fmt.Sprintf("%s: %v", provider.rel, err)}
	}
	genCtors, err := handwrittenConstructors(filepath.Join(opts.GeneratedDir, provider.rel))
	if err != nil {
		return overwrites, []string{fmt.Sprintf("%s: %v", provider.rel, err)}
	}
	if added := onlyIn(genCtors, baseCtors); len(added) > 0 {
		problems = append(problems, fmt.Sprintf("%s adds hand-written constructor(s) — generation only ever retires them: %s", provider.rel, strings.Join(added, ", ")))
	}
	removed := onlyIn(baseCtors, genCtors)

	// Map each removed constructor to the artifact whose base file declared it.
	claimed := map[string]bool{}
	for _, name := range sortedNames(plans) {
		p := plans[name]
		if p.source == nil || p.report == nil ||
			(p.report.status != model.ArtifactStatusCreated && p.report.status != model.ArtifactStatusUpdated) {
			continue
		}
		if p.source.added {
			continue // net-new file: nothing pre-existed to overwrite
		}
		declared, derr := declaredConstructors(filepath.Join(opts.BaseDir, p.source.rel))
		if derr != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", p.source.rel, derr))
			continue
		}
		for _, c := range declared {
			if slices.Contains(removed, c) {
				overwrites[name] = append(overwrites[name], c)
				claimed[c] = true
			}
		}
	}
	for _, c := range removed {
		if !claimed[c] {
			problems = append(problems, fmt.Sprintf("%s retires %s, but no changed data source file declares it", provider.rel, c))
		}
	}
	return overwrites, problems
}

// diffTrees returns scoped files that differ between base and generated (added
// or modified), plus scoped files present in base but gone from generated
// (deleted). scopedFiles limits the comparison to generator-owned directories
// and selected test files so unrelated master drift cannot invalidate an
// otherwise self-contained generator push.
func diffTrees(baseDir, genDir string) (changed []changedFile, deleted []string, err error) {
	generatedFiles, err := scopedFiles(genDir)
	if err != nil {
		return nil, nil, err
	}
	baseFiles, err := scopedFiles(baseDir)
	if err != nil {
		return nil, nil, err
	}

	for _, rel := range generatedFiles {
		path := filepath.Join(genDir, rel)
		same, existed, cmpErr := sameContent(filepath.Join(baseDir, rel), path)
		if cmpErr != nil {
			return nil, nil, cmpErr
		}
		if !same {
			changed = append(changed, changedFile{rel: rel, added: !existed})
		}
	}

	for _, rel := range baseFiles {
		_, statErr := os.Stat(filepath.Join(genDir, rel))
		if os.IsNotExist(statErr) {
			deleted = append(deleted, rel)
		} else if statErr != nil {
			return nil, nil, statErr
		}
	}
	sort.Strings(deleted)
	return changed, deleted, nil
}

// scopedFiles returns every file under provider/docs/examples plus only
// generated data-source tests and provider_test.go. This preserves fail-loud
// attribution without making unrelated test-suite or cassette drift part of
// the split.
func scopedFiles(root string) ([]string, error) {
	var out []string
	for _, scope := range diffDirectoryScopes {
		scopePath := filepath.Join(root, scope)
		if _, err := os.Stat(scopePath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if err := filepath.WalkDir(scopePath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			out = append(out, rel)
			return nil
		}); err != nil {
			return nil, err
		}
	}

	providerTestRel := filepath.Join("datadog", "tests", providerTestFile)
	if info, err := os.Stat(filepath.Join(root, providerTestRel)); err == nil && !info.IsDir() {
		out = append(out, providerTestRel)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	testMatches, err := filepath.Glob(filepath.Join(root, "datadog", "tests", "data_source_datadog_*_test.go"))
	if err != nil {
		return nil, err
	}
	for _, path := range testMatches {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	sort.Strings(out)
	return slices.Compact(out), nil
}

// sameContent reports whether basePath and genPath hold identical bytes, and
// whether basePath existed at all (a missing base file means the generated file
// is net-new).
func sameContent(basePath, genPath string) (same, existed bool, err error) {
	bi, berr := os.Stat(basePath)
	if berr != nil {
		if os.IsNotExist(berr) {
			return false, false, nil
		}
		return false, false, berr
	}
	gi, gerr := os.Stat(genPath)
	if gerr != nil {
		return false, true, gerr
	}
	if bi.Size() != gi.Size() {
		return false, true, nil
	}
	bb, err := os.ReadFile(basePath)
	if err != nil {
		return false, true, err
	}
	gb, err := os.ReadFile(genPath)
	if err != nil {
		return false, true, err
	}
	return bytes.Equal(bb, gb), true, nil
}

// handwrittenConstructors lists the New<...>DataSource constructors inside the
// hand-written Datasources slice in framework_provider.go, scoped to that block so
// a like-named Resources entry is never counted.
func handwrittenConstructors(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == datasourcesSliceHeader {
			start = i
			break
		}
	}
	if start == -1 {
		return nil, fmt.Errorf("Datasources slice not found")
	}
	var out []string
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "}" {
			return out, nil
		}
		out = append(out, constructorRe.FindAllString(lines[i], -1)...)
	}
	return nil, fmt.Errorf("Datasources slice is not terminated")
}

// declaredConstructors returns the New<...>DataSource constructor(s) a data-source
// file declares.
func declaredConstructors(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, m := range constructorDeclRe.FindAllStringSubmatch(string(data), -1) {
		out = append(out, m[1])
	}
	return out, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func planFor(m map[string]*artifactPlan, name string) *artifactPlan {
	if p := m[name]; p != nil {
		return p
	}
	p := &artifactPlan{}
	m[name] = p
	return p
}

func sortedNames(m map[string]*artifactPlan) []string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func without(items []string, target string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item != target {
			out = append(out, item)
		}
	}
	return out
}

func cloneMap[K comparable, V any](in map[K]V) map[K]V {
	out := make(map[K]V, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mapsEqual[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// onlyIn returns the sorted, de-duplicated members of a not present in b.
func onlyIn(a, b []string) []string {
	set := map[string]struct{}{}
	for _, x := range b {
		set[x] = struct{}{}
	}
	seen := map[string]struct{}{}
	var out []string
	for _, x := range a {
		if _, in := set[x]; in {
			continue
		}
		if _, dup := seen[x]; dup {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

func finish(rep *Result, problems []string) (*Result, error) {
	if len(problems) == 0 {
		return rep, nil
	}
	sort.Strings(problems)
	rep.Errors = problems
	return rep, fmt.Errorf("split: %d attribution problem(s); see report", len(problems))
}
