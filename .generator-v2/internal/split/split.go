// Package split turns one aggregate tfgen push — all N generated data sources on
// a single branch — into per-artifact bundles, one directory per data source, so
// each can land as its own PR. It is pure: it diffs two on-disk trees (a master
// checkout and the pushed-branch checkout), reuses tfgen's registry serializer to
// reconstruct each artifact's datasources_generated.go as "the base set plus this
// one artifact", and never touches git, the network, or the OpenAPI spec.
//
// Attribution is exact and fail-loud: a changed file that maps to no artifact and
// is not a known shared registry file is a hard error, never silently dropped, so
// a future shared helper file can't vanish. Removals (retirement) are out of scope
// for this create/update MVP and are reported as errors.
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
	Name   string `json:"name"`
	Status string `json:"status"` // "created" (net-new) or "updated" (overwrote an existing file)
	// ServiceTag is the [service] PR-title prefix; derived in Phase 2 (the incoming
	// diff carries no OpenAPI tag), so it is left empty here.
	ServiceTag string   `json:"service_tag,omitempty"`
	Files      []string `json:"files"`
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
	dataSourceFileRe     = regexp.MustCompile(`^data_source_datadog_([a-z][a-z0-9_]*)\.go$`)
	dataSourceTestFileRe = regexp.MustCompile(`^data_source_datadog_([a-z][a-z0-9_]*)_test\.go$`)
	// constructorDeclRe recovers the constructor(s) a data-source file declares.
	constructorDeclRe = regexp.MustCompile(`func (New[A-Za-z0-9_]+DataSource)\(`)
	// constructorRe recovers constructor identifiers listed inside a slice literal.
	constructorRe = regexp.MustCompile(`New[A-Za-z0-9_]+DataSource`)
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
)

// classify maps a changed file's basename to its role in the split.
func classify(base string) (fileKind, string) {
	switch base {
	case registryFile:
		return kindRegistry, ""
	case providerFile:
		return kindProvider, ""
	case providerTestFile:
		return kindProviderTest, ""
	}
	if m := dataSourceTestFileRe.FindStringSubmatch(base); m != nil {
		return kindDataSourceTest, m[1]
	}
	if m := dataSourceFileRe.FindStringSubmatch(base); m != nil {
		return kindDataSource, m[1]
	}
	return kindUnknown, ""
}

// changedFile is a file differing between base and generated.
type changedFile struct {
	rel   string // repo-relative path
	base  string // basename
	added bool   // absent in base (net-new) vs. modified
}

// artifactPlan gathers the self-contained files discovered for one data source.
type artifactPlan struct {
	source *changedFile
	test   *changedFile
}

// Split diffs the two trees, routes each artifact into OutDir/<name>/, and returns
// a report. Attribution problems are recorded in report.Errors and also returned
// as an error (so the caller exits non-zero); the report is still complete.
func Split(opts Options) (*Result, error) {
	rep := &Result{BaseDir: opts.BaseDir, GeneratedDir: opts.GeneratedDir, OutDir: opts.OutDir}

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
		kind, name := classify(cf.base)
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
		default:
			fail("changed file maps to no artifact and is not a known shared file: %s", cf.rel)
		}
	}
	for _, d := range deleted {
		fail("file removed (retirement is out of scope for the create/update MVP): %s", d)
	}
	// The shared endpoint-tag map only appears with --emit-tests, which the pipeline
	// omits; reconstructing it needs the OpenAPI tag the branch does not carry.
	if providerTest != nil {
		fail("%s changed, but test routing (--emit-tests) is deferred; the incoming push must omit --emit-tests", providerTest.rel)
	}

	if len(plans) == 0 {
		// Nothing to split; a shared-file change with no data source is a problem.
		if registry != nil {
			fail("%s changed but no data source files changed", registry.rel)
		}
		return finish(rep, problems)
	}
	if registry == nil {
		fail("data source files changed but %s did not", registryFile)
		return finish(rep, problems)
	}

	baseSet, err := emit.RegisteredGeneratedDatasources(filepath.Join(opts.BaseDir, registry.rel))
	if err != nil {
		return rep, fmt.Errorf("split: reading base registry: %w", err)
	}

	overwrites, provProblems := attributeOverwrites(opts, plans, provider)
	problems = append(problems, provProblems...)

	// Cross-check: base ∪ every artifact's constructor must equal the incoming
	// registry, or an artifact is unaccounted for (a registration with no file, or
	// vice versa).
	incoming, err := emit.RegisteredGeneratedDatasources(filepath.Join(opts.GeneratedDir, registry.rel))
	if err != nil {
		return rep, fmt.Errorf("split: reading generated registry: %w", err)
	}
	want := append(slices.Clone(baseSet), constructorsFor(plans)...)
	if diff := onlyIn(incoming, want); len(diff) > 0 {
		fail("%s registers constructor(s) not accounted for by any changed data source file: %s", registry.rel, strings.Join(diff, ", "))
	}
	if diff := onlyIn(want, incoming); len(diff) > 0 {
		fail("data source(s) changed whose constructor is missing from the incoming %s: %s", registry.rel, strings.Join(diff, ", "))
	}

	for _, name := range sortedNames(plans) {
		art, artProblems := materialize(opts, name, plans[name], registry.rel, baseSet, overwrites[name], provider)
		problems = append(problems, artProblems...)
		if art != nil {
			rep.Artifacts = append(rep.Artifacts, *art)
		}
	}

	return finish(rep, problems)
}

// materialize writes one artifact's bundle under OutDir/<name>/ and returns its
// report entry. The self-contained files (source, optional test) are copied
// verbatim; datasources_generated.go is reconstructed as baseSet ∪ {this
// constructor}; framework_provider.go is base with this artifact's overwritten
// constructor(s) removed.
func materialize(opts Options, name string, p *artifactPlan, registryRel string, baseSet, overwritten []string, provider *changedFile) (*Artifact, []string) {
	var problems []string
	art := &Artifact{Name: name}
	if p.source.added {
		art.Status = "created"
	} else {
		art.Status = "updated"
	}
	outDir := filepath.Join(opts.OutDir, name)

	files := []string{p.source.rel}
	if p.test != nil {
		files = append(files, p.test.rel)
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

	// An overwrite retires hand-written constructor(s) from framework_provider.go;
	// replay that removal against the base copy (there is no from-set rebuild for it).
	if len(overwritten) > 0 {
		provOut := filepath.Join(outDir, provider.rel)
		if !opts.Check {
			if err := copyFile(filepath.Join(opts.BaseDir, provider.rel), provOut); err != nil {
				problems = append(problems, fmt.Sprintf("copying %s for %q: %v", provider.rel, name, err))
			}
		}
		for _, c := range overwritten {
			if _, err := emit.RemoveHandwrittenDatasource(provOut, c, opts.Check); err != nil {
				problems = append(problems, fmt.Sprintf("retiring %s from %s for %q: %v", c, provider.rel, name, err))
			}
		}
		art.Files = append(art.Files, provider.rel)
	}

	sort.Strings(art.Files)
	art.Files = slices.Compact(art.Files)
	return art, problems
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

// diffTrees returns the files that differ between base and generated (added or
// modified), plus files present in base but gone from generated (deleted). The
// .git directory is skipped; clean checkouts are assumed.
func diffTrees(baseDir, genDir string) (changed []changedFile, deleted []string, err error) {
	err = filepath.WalkDir(genDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(genDir, path)
		if relErr != nil {
			return relErr
		}
		same, existed, cmpErr := sameContent(filepath.Join(baseDir, rel), path)
		if cmpErr != nil {
			return cmpErr
		}
		if !same {
			changed = append(changed, changedFile{rel: rel, base: d.Name(), added: !existed})
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	err = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(baseDir, path)
		if relErr != nil {
			return relErr
		}
		if _, statErr := os.Stat(filepath.Join(genDir, rel)); os.IsNotExist(statErr) {
			deleted = append(deleted, rel)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	sort.Strings(deleted)
	return changed, deleted, nil
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

func constructorsFor(m map[string]*artifactPlan) []string {
	out := make([]string, 0, len(m))
	for _, n := range sortedNames(m) {
		out = append(out, emit.DatasourceConstructor(n))
	}
	return out
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
