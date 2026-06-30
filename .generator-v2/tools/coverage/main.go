// Command coverage reports data-source generation coverage for an OpenAPI spec,
// measured per "thing" (resource), not per endpoint.
//
// THROWAWAY ANALYSIS TOOL — not part of the generator and not maintained.
// Run it via ./run.sh. It reuses the REAL generator internals, so
// its verdicts track the generator automatically.
//
// Model (decided deliberately — see README):
//   - A "thing" is a resource keyed by its collection path: a by-id GET at
//     /X/{id} pairs with a list GET at /X into one thing.
//   - PLURAL coverage: of all things, the fraction whose list endpoint generates
//     a plural data source.
//   - SINGULAR coverage ("both" mode): of all things, the fraction that have BOTH
//     a by-id endpoint AND a searchable list AND generate the id-or-search
//     singular data source. A thing counted once — never twice for id + search.
//   - Denominator for both percentages is ALL things (shared base); a thing that
//     lacks what an axis needs is a miss, not an exclusion.
//   - "Generatable" means builds AND renders clean (gofmt) — must compile.
//   - Stable and x-unstable things are reported separately.
//
// It skips LoadSpec's global ref-cycle gate (which fails fast on the full
// upstream spec), reports that gate's result separately, and evaluates each thing
// in isolation. Schema trees are normalized then freed per op/thing, so peak heap
// stays bounded to the libopenapi model plus one thing's trees at a time.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/emit"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

func main() {
	specPath := flag.String("spec", ".generator/V2/openapi.yaml", "OpenAPI spec to analyze")
	maxDepth := flag.Int("max-depth", parser.DefaultMaxDepth, "$ref expansion limit (per operation)")
	jsonPath := flag.String("json", "", "Optional path to dump the full per-thing report as JSON")
	listFails := flag.Bool("list-fails", false, "List every uncovered thing with its per-axis reason")
	listOK := flag.Bool("list-ok", false, "List every covered thing")
	examples := flag.Int("examples", 3, "Example things to list under each failure category (-1 = all)")
	noProgress := flag.Bool("no-progress", false, "Suppress the live status output on stderr")
	noCycleGate := flag.Bool("no-cycle-gate", false, "Skip the global ref-cycle scan (slow on huge specs)")
	cpuProfile := flag.String("cpuprofile", "", "Write a CPU profile to this path")
	memProfile := flag.String("memprofile", "", "Write a heap profile to this path (after the run)")
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	st := newStatus(!*noProgress)
	rep, err := analyzeCoverage(*specPath, *maxDepth, !*noCycleGate, st)
	if err != nil {
		fatal(err)
	}
	st.summary(len(rep.Things))

	rep.print(os.Stdout, *listFails, *listOK, *examples)
	if *jsonPath != "" {
		if err := rep.writeJSON(*jsonPath); err != nil {
			fatal(err)
		}
		fmt.Printf("\n[detail written to %s]\n", *jsonPath)
	}

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			fatal(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "coverage:", err)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// thingOutcome is one resource's verdict on both axes.
type thingOutcome struct {
	Key      string `json:"key"` // collection path
	Unstable bool   `json:"unstable,omitempty"`

	HasList      bool `json:"has_list"`
	HasByID      bool `json:"has_by_id"`
	HasSingleton bool `json:"has_singleton,omitempty"`

	// PluralOK: list endpoint generates a plural data source.
	PluralOK     bool   `json:"plural_ok"`
	PluralReason string `json:"plural_reason,omitempty"`
	PluralCat    string `json:"plural_category,omitempty"`
	// SingularOK: id+search "both"-mode singular data source generates. Requires
	// both a by-id endpoint and a searchable list.
	SingularOK     bool   `json:"singular_ok"`
	SingularReason string `json:"singular_reason,omitempty"`
	SingularCat    string `json:"singular_category,omitempty"`
}

// coverageReport is the full analysis result.
type coverageReport struct {
	Spec        string         `json:"spec"`
	MaxDepth    int            `json:"max_depth"`
	CycleGate   bool           `json:"cycle_gate_run"`
	Cycles      []string       `json:"ref_cycles"`
	MaxDepthHit string         `json:"max_depth_error,omitempty"`
	LoadGateOK  bool           `json:"load_gate_ok"`
	TotalGETs   int            `json:"total_gets"`
	Things      []thingOutcome `json:"things"`
}

// endpointInfo is a classified GET endpoint (pass 1).
type endpointInfo struct {
	path     string
	opID     string
	raw      *v3.Operation
	unstable bool
	role     string // "list" | "byid" | "singleton" | "" (non-candidate)
	thingKey string
}

// thing groups the endpoints of one resource (pass 2).
type thing struct {
	key       string
	list      *endpointInfo
	byid      *endpointInfo
	singleton *endpointInfo
	unstable  bool
}

// ---------------------------------------------------------------------------
// Analysis
// ---------------------------------------------------------------------------

func analyzeCoverage(specPath string, maxDepth int, runCycleGate bool, st *status) (*coverageReport, error) {
	st.step("reading spec %q", specPath)
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("reading spec %q: %w", specPath, err)
	}
	st.ok("read %s", humanBytes(uint64(len(data))))

	// SkipCircularReferenceCheck lets the model build on specs whose component
	// graph contains $ref cycles (the upstream Datadog spec does).
	st.step("building OpenAPI v3 model (the big one-time allocation)")
	doc, err := libopenapi.NewDocumentWithConfiguration(data, &datamodel.DocumentConfiguration{
		SkipCircularReferenceCheck: true,
	})
	if err != nil {
		return nil, fmt.Errorf("parsing spec %q: %w", specPath, err)
	}
	v3doc, errs := doc.BuildV3Model()
	if v3doc == nil {
		return nil, fmt.Errorf("building OpenAPI v3 model for %q: %w", specPath, errs)
	}
	data = nil
	components := v3doc.Model.Components
	st.ok("model built")

	// Collect every GET (path + raw + unstable flag).
	type rawGet struct {
		path     string
		raw      *v3.Operation
		unstable bool
	}
	var gets []rawGet
	if paths := v3doc.Model.Paths; paths != nil && paths.PathItems != nil {
		for opPath, item := range paths.PathItems.FromOldest() {
			if item == nil {
				continue
			}
			for method, op := range item.GetOperations().FromOldest() {
				if op == nil || strings.ToUpper(method) != "GET" || op.OperationId == "" {
					continue
				}
				gets = append(gets, rawGet{
					path:     opPath,
					raw:      op,
					unstable: op.Extensions != nil && op.Extensions.GetOrZero("x-unstable") != nil,
				})
			}
		}
	}

	rep := &coverageReport{Spec: specPath, MaxDepth: maxDepth, CycleGate: runCycleGate, TotalGETs: len(gets)}

	if runCycleGate {
		st.step("scanning component graph for $ref cycles (global load gate)")
		cycles, gateErr := parser.DetectComponentRefCycles(components, maxDepth)
		for _, c := range cycles {
			rep.Cycles = append(rep.Cycles, c.Ref)
		}
		if gateErr != nil {
			var mde *parser.MaxDepthError
			if errors.As(gateErr, &mde) {
				rep.MaxDepthHit = mde.Ref
			} else {
				rep.MaxDepthHit = gateErr.Error()
			}
		}
		rep.LoadGateOK = len(rep.Cycles) == 0 && rep.MaxDepthHit == ""
		st.ok("cycle scan done: %d cycle(s), gate %s", len(rep.Cycles), gateLabel(rep.LoadGateOK))
	} else {
		st.note("cycle gate skipped (--no-cycle-gate)")
	}

	// Pass 1: classify each GET by response shape (normalize, read data kind, free).
	st.step("classifying %d GET endpoints", len(gets))
	infos := make([]*endpointInfo, 0, len(gets))
	for i, g := range gets {
		info := &endpointInfo{path: g.path, opID: g.raw.OperationId, raw: g.raw, unstable: g.unstable}
		info.role, info.thingKey = classifyEndpoint(components, g.path, g.raw.OperationId, g.raw, maxDepth)
		infos = append(infos, info)
		st.tick("classify", i+1, len(gets), 0, 0, g.path, false)
	}
	st.tick("classify", len(gets), len(gets), 0, 0, "done", true)

	// Group endpoints into things by collection path.
	things := map[string]*thing{}
	for _, info := range infos {
		if info.role == "" {
			continue // non-candidate
		}
		t := things[info.thingKey]
		if t == nil {
			t = &thing{key: info.thingKey}
			things[info.thingKey] = t
		}
		switch info.role {
		case "list":
			t.list = info
		case "byid":
			t.byid = info
		case "singleton":
			t.singleton = info
		}
	}
	// A thing's stability follows its primary endpoint (list, else by-id, else singleton).
	keys := make([]string, 0, len(things))
	for k, t := range things {
		rep := t.list
		if rep == nil {
			rep = t.byid
		}
		if rep == nil {
			rep = t.singleton
		}
		t.unstable = rep != nil && rep.unstable
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Pass 2: evaluate each thing through the real pipeline (only things with a
	// list can be covered on either axis; others are recorded as structural misses).
	st.step("evaluating %d things through the real pipeline", len(keys))
	rep.Things = make([]thingOutcome, 0, len(keys))
	covS, covP := 0, 0
	for i, k := range keys {
		out := evaluateThing(components, things[k], maxDepth)
		if out.SingularOK {
			covS++
		}
		if out.PluralOK {
			covP++
		}
		rep.Things = append(rep.Things, out)
		st.tick("evaluate", i+1, len(keys), covS, covP, k, false)
	}
	st.tick("evaluate", len(keys), len(keys), covS, covP, "done", true)
	return rep, nil
}

// classifyEndpoint normalizes one GET's response (real parser), reads the data
// shape, then lets the tree be freed. role is "list" (data array), "byid" (data
// object at /X/{id}), "singleton" (data object at /X), or "" (non-candidate).
func classifyEndpoint(components *v3.Components, path, opID string, raw *v3.Operation, maxDepth int) (role, thingKey string) {
	mop := &model.Operation{
		Path: path, Method: "GET", OperationId: opID, Tag: firstTag(raw.Tags),
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindDataSource, ArtifactName: opID,
			Cardinality: model.CardinalitySingular, IdStrategy: model.IdStrategyDataID,
			Group: &model.OperationGroup{Read: opID},
		},
	}
	spec := &model.Spec{Components: components, Operations: []*model.Operation{mop}}
	if err := parser.NormalizeSchemas(spec, map[*model.Operation]*v3.Operation{mop: raw}, maxDepth, parser.DefaultTrackingFieldName); err != nil {
		return "", ""
	}
	defer func() { mop.ResponseSchema = nil }()

	rs := mop.ResponseSchema
	if rs == nil || rs.Kind != model.SchemaKindObject {
		return "", ""
	}
	resultsPath := "data"
	if mop.Pagination != nil && mop.Pagination.ResultsPath != "" {
		resultsPath = mop.Pagination.ResultsPath
	}
	data := rs.Properties[resultsPath]
	if data == nil {
		return "", ""
	}
	switch data.Kind {
	case model.SchemaKindArray:
		return "list", path
	case model.SchemaKindObject:
		if isItemPath(path) {
			return "byid", collectionKey(path)
		}
		return "singleton", path
	default:
		return "", ""
	}
}

// evaluateThing builds and renders a thing's data sources via the real pipeline.
//
//   - PLURAL: requires a list endpoint; builds the plural data source from it.
//   - SINGULAR: a thing is covered if ANY mode generates, counted once — by-id
//     read-only (/X/{id}), a singleton read-only (/X returning one object), or
//     SEARCH-ONLY from the list (list-the-collection-take-one; no by-id and no
//     filter item required). Modes are tried in that order and OR'd.
func evaluateThing(components *v3.Components, t *thing, maxDepth int) thingOutcome {
	out := thingOutcome{Key: t.key, Unstable: t.unstable, HasList: t.list != nil, HasByID: t.byid != nil, HasSingleton: t.singleton != nil}

	// Plural: list endpoint only.
	if t.list != nil {
		out.PluralOK, out.PluralReason, out.PluralCat = generate(
			components, t.list.raw, t.list.path, t.list.opID,
			model.CardinalityPlural, &model.OperationGroup{Read: t.list.opID}, maxDepth)
	} else {
		out.PluralReason, out.PluralCat = "thing has no list endpoint", "no list endpoint"
	}

	// Singular: try each available mode, stop at the first that generates. The
	// first failing mode's reason is kept when none succeed.
	trySingular := func(raw *v3.Operation, path, opID string, group *model.OperationGroup) bool {
		ok, reason, cat := generate(components, raw, path, opID, model.CardinalitySingular, group, maxDepth)
		if ok {
			out.SingularOK, out.SingularReason, out.SingularCat = true, "", ""
			return true
		}
		if out.SingularReason == "" {
			out.SingularReason, out.SingularCat = reason, cat
		}
		return false
	}

	switch {
	case t.byid != nil:
		_ = trySingular(t.byid.raw, t.byid.path, t.byid.opID, &model.OperationGroup{Read: t.byid.opID}) ||
			(t.list != nil && trySingular(t.list.raw, t.list.path, t.list.opID, &model.OperationGroup{Search: t.list.opID}))
	case t.singleton != nil:
		_ = trySingular(t.singleton.raw, t.singleton.path, t.singleton.opID, &model.OperationGroup{Read: t.singleton.opID}) ||
			(t.list != nil && trySingular(t.list.raw, t.list.path, t.list.opID, &model.OperationGroup{Search: t.list.opID}))
	case t.list != nil:
		_ = trySingular(t.list.raw, t.list.path, t.list.opID, &model.OperationGroup{Search: t.list.opID})
	default:
		out.SingularReason, out.SingularCat = "no candidate endpoint", "no candidate endpoint"
	}
	return out
}

// generate normalizes one synthesized operation (1-op spec, freed after) and runs
// its artifact through BuildArtifact -> BuildDataSourceView -> RenderDataSource.
// The Tracking group selects the mode: {Read} -> by-id/plural, {Search} ->
// search-only singular.
func generate(components *v3.Components, raw *v3.Operation, path, opID string, card model.Cardinality, group *model.OperationGroup, maxDepth int) (ok bool, reason, category string) {
	mop := &model.Operation{
		Path: path, Method: "GET", OperationId: opID, Tag: firstTag(raw.Tags),
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindDataSource, ArtifactName: opID,
			Cardinality: card, IdStrategy: model.IdStrategyDataID, Group: group,
		},
	}
	spec := &model.Spec{Components: components, Operations: []*model.Operation{mop}}
	if err := parser.NormalizeSchemas(spec, map[*model.Operation]*v3.Operation{mop: raw}, maxDepth, parser.DefaultTrackingFieldName); err != nil {
		return false, "normalize: " + firstLine(err.Error()), "normalize error"
	}
	defer func() {
		mop.ResponseSchema = nil
		mop.RequestSchema = nil
		mop.QueryParams = nil
	}()
	return tryGenerate(mop)
}

// tryGenerate runs one operation's artifact through BuildArtifact ->
// BuildDataSourceView -> RenderDataSource and returns (ok, reason, category).
func tryGenerate(op *model.Operation) (ok bool, reason, category string) {
	defer func() {
		if r := recover(); r != nil {
			ok, reason, category = false, fmt.Sprintf("panic: %v", r), "panic during evaluation"
		}
	}()
	artifact, err := model.BuildArtifact(op)
	if err == nil {
		var view emit.DataSourceView
		view, err = emit.BuildDataSourceView(artifact)
		if err == nil {
			_, err = emit.RenderDataSource(view)
		}
	}
	if err != nil {
		r, c := describeErr(err)
		return false, r, c
	}
	return true, "", ""
}

// ---------------------------------------------------------------------------
// Path helpers (thing grouping)
// ---------------------------------------------------------------------------

func lastSegment(path string) string {
	p := strings.TrimRight(path, "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// isItemPath reports whether the path's final segment is a single path parameter
// (e.g. /api/v2/teams/{team_id}).
func isItemPath(path string) bool {
	s := lastSegment(path)
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}

// collectionKey strips a trailing /{param} segment, mapping a by-id path to its
// collection (e.g. /api/v2/teams/{team_id} -> /api/v2/teams).
func collectionKey(path string) string {
	p := strings.TrimRight(path, "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[:i]
	}
	return p
}

// ---------------------------------------------------------------------------
// Error categorization (reused from the generator's own errors)
// ---------------------------------------------------------------------------

func describeErr(err error) (reason, category string) {
	var emitErr *emit.UnsupportedEmitError
	if errors.As(err, &emitErr) && len(emitErr.Nodes) > 0 {
		n := emitErr.Nodes[0]
		return firstLine(n.Path + ": " + n.Reason), bucket(n.Reason)
	}
	var kindErr *model.UnsupportedKindError
	if errors.As(err, &kindErr) {
		switch kindErr.Kind {
		case model.SchemaKindRefCycle:
			return firstLine(err.Error()), "recursive schema ($ref cycle / depth>8)"
		default:
			return firstLine(err.Error()), "unrepresentable node (anyOf / typeless / free-form / empty)"
		}
	}
	msg := firstLine(err.Error())
	return msg, bucket(msg)
}

func bucket(r string) string {
	switch {
	case strings.Contains(r, "nesting under attributes"),
		strings.Contains(r, "nesting under item attributes"):
		return "nested object/map directly under attributes (emit hoisting gap)"
	case strings.Contains(r, "map value kind"),
		strings.Contains(r, "map not yet supported"),
		strings.Contains(r, "map-of-object"):
		return "map / free-form additionalProperties under attributes"
	case strings.Contains(r, "nested-attribute form"):
		return "nested-attribute form not yet supported"
	case strings.Contains(r, "single-member JSON:API envelope"):
		return "not a single {data:{...}} envelope (extra siblings / bare shape)"
	case strings.Contains(r, "missing an attributes object"):
		return "data envelope has no attributes object"
	case strings.Contains(r, "attributes must be an object"):
		return "envelope attributes is not an object"
	case strings.Contains(r, "missing results array block"):
		return "list results not an array-of-object"
	case strings.Contains(r, "missing list item type"),
		strings.Contains(r, "result-array element"):
		return "list element has no SDK $ref item type"
	case strings.Contains(r, "gofmt of generated"):
		return "generated Go fails to compile (gofmt: e.g. `type` keyword collision)"
	case strings.Contains(r, "ref_cycle"), strings.Contains(r, "depth"):
		return "recursive schema ($ref cycle / depth>8)"
	case strings.Contains(r, "not representable"):
		return "unrepresentable node (anyOf / typeless / free-form / empty)"
	default:
		return "other: " + r
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// ---------------------------------------------------------------------------
// Reporting
// ---------------------------------------------------------------------------

func (rep *coverageReport) print(w io.Writer, listFails, listOK bool, maxExamples int) {
	p := func(format string, a ...any) { fmt.Fprintf(w, format, a...) }

	// Aggregate. Failure maps hold the thing keys per category, so we can list
	// them (counts are derived from the slice length).
	var total, withList, byIDOnly, singletonOnly int
	stable := map[bool]*struct{ n, singOK, plurOK int }{
		false: {}, true: {},
	}
	singFailCat := map[string][]string{}
	plurFailCat := map[string][]string{}
	for _, t := range rep.Things {
		total++
		g := stable[t.Unstable]
		g.n++
		if t.HasList {
			withList++
		} else if t.HasByID {
			byIDOnly++
		} else if t.HasSingleton {
			singletonOnly++
		}
		// Every thing is singular-eligible (some mode can be attempted), so any
		// thing that doesn't generate a singular is a real miss.
		if t.SingularOK {
			g.singOK++
		} else {
			singFailCat[t.SingularCat] = append(singFailCat[t.SingularCat], t.Key)
		}
		// Plural only applies to things with a list endpoint.
		if t.PluralOK {
			g.plurOK++
		} else if t.HasList {
			plurFailCat[t.PluralCat] = append(plurFailCat[t.PluralCat], t.Key)
		}
	}

	p("==============================================================================\n")
	p("DATA SOURCE COVERAGE  (real generator internals; unit = thing/resource)\n")
	p("  spec: %s\n", rep.Spec)
	p("==============================================================================\n\n")

	p("[GLOBAL LOAD GATE: DetectComponentRefCycles, maxDepth=%d]\n", rep.MaxDepth)
	if !rep.CycleGate {
		p("  (skipped via --no-cycle-gate)\n\n")
	} else {
		p("  distinct $ref cycles : %d\n", len(rep.Cycles))
		for i, c := range rep.Cycles {
			if i >= 10 {
				p("      ... +%d more\n", len(rep.Cycles)-10)
				break
			}
			p("      - %s\n", c)
		}
		if rep.MaxDepthHit != "" {
			p("  max-depth chain hit  : %s\n", rep.MaxDepthHit)
		}
		if rep.LoadGateOK {
			p("  => LoadSpec on this spec would: succeed\n\n")
		} else {
			p("  => LoadSpec on this spec would: FAIL (analyzed per-thing instead)\n\n")
		}
	}

	p("[THINGS]   (a thing = resource keyed by collection path; %d GET ops total)\n", rep.TotalGETs)
	p("  distinct things                 : %d\n", total)
	p("    with a list endpoint          : %d   (plural-eligible; singular via search)\n", withList)
	p("    by-id only (no list)          : %d   (no plural; singular via by-id)\n", byIDOnly)
	p("    singleton only (no id, list)  : %d   (no plural; singular via read)\n", singletonOnly)
	p("\n")

	printAxis := func(label string, g *struct{ n, singOK, plurOK int }) {
		p("[COVERAGE — %s]   (denominator = all %s things = %d)\n", label, strings.ToLower(label), g.n)
		if g.n == 0 {
			p("  (none)\n\n")
			return
		}
		p("  singular (by-id / search / read): %d/%d = %.1f%%\n",
			g.singOK, g.n, 100*float64(g.singOK)/float64(g.n))
		p("  plural                          : %d/%d = %.1f%%\n\n",
			g.plurOK, g.n, 100*float64(g.plurOK)/float64(g.n))
	}
	printAxis("STABLE", stable[false])
	printAxis("UNSTABLE", stable[true])

	// printCats lists each failure category (count-sorted), then up to maxExamples
	// of the things in it (sorted); maxExamples < 0 lists every one.
	printCats := func(title string, m map[string][]string) {
		p("[%s]\n", title)
		type kv struct {
			k     string
			items []string
		}
		rows := make([]kv, 0, len(m))
		for k, items := range m {
			rows = append(rows, kv{k, items})
		}
		sort.Slice(rows, func(i, j int) bool {
			if len(rows[i].items) != len(rows[j].items) {
				return len(rows[i].items) > len(rows[j].items)
			}
			return rows[i].k < rows[j].k
		})
		for _, r := range rows {
			p("  %4d  %s\n", len(r.items), r.k)
			items := append([]string(nil), r.items...)
			sort.Strings(items)
			shown := len(items)
			if maxExamples >= 0 && maxExamples < shown {
				shown = maxExamples
			}
			for _, key := range items[:shown] {
				p("            %s\n", key)
			}
			if shown < len(items) {
				p("            … +%d more (use --examples -1 to list all)\n", len(items)-shown)
			}
		}
		p("\n")
	}
	printCats("SINGULAR — why things failed (best of by-id / search / read)", singFailCat)
	printCats("PLURAL — why list-having things failed", plurFailCat)

	if listFails {
		p("[UNCOVERED THINGS]\n")
		for _, t := range rep.Things {
			if t.SingularOK && t.PluralOK {
				continue
			}
			tag := ""
			if t.Unstable {
				tag = " (unstable)"
			}
			p("  %s%s\n", t.Key, tag)
			if !t.SingularOK {
				p("      singular: %s\n", t.SingularReason)
			}
			if !t.PluralOK {
				p("      plural:   %s\n", t.PluralReason)
			}
		}
		p("\n")
	}
	if listOK {
		p("[COVERED THINGS]\n")
		for _, t := range rep.Things {
			if !t.SingularOK && !t.PluralOK {
				continue
			}
			marks := ""
			switch {
			case t.SingularOK && t.PluralOK:
				marks = "singular+plural"
			case t.SingularOK:
				marks = "singular"
			default:
				marks = "plural"
			}
			tag := ""
			if t.Unstable {
				tag = " (unstable)"
			}
			p("  %-16s %s%s\n", marks, t.Key, tag)
		}
		p("\n")
	}
}

func (rep *coverageReport) writeJSON(path string) error {
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func firstTag(tags []string) string {
	if len(tags) > 0 {
		return tags[0]
	}
	return ""
}

func gateLabel(ok bool) string {
	if ok {
		return "would PASS"
	}
	return "would FAIL"
}

// ---------------------------------------------------------------------------
// status: live progress + memory reporting on stderr.
// ---------------------------------------------------------------------------

type status struct {
	w         io.Writer
	enabled   bool
	tty       bool
	start     time.Time
	stepStart time.Time
	lastTick  time.Time
	peakHeap  uint64
}

func newStatus(enabled bool) *status {
	s := &status{w: os.Stderr, enabled: enabled, start: time.Now(), stepStart: time.Now()}
	if fi, err := os.Stderr.Stat(); err == nil {
		s.tty = fi.Mode()&os.ModeCharDevice != 0
	}
	return s
}

func (s *status) heap() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.HeapAlloc > s.peakHeap {
		s.peakHeap = m.HeapAlloc
	}
	return m.HeapAlloc
}

func (s *status) step(format string, a ...any) {
	s.stepStart = time.Now()
	if !s.enabled {
		return
	}
	fmt.Fprintf(s.w, "→ "+format+"\n", a...)
}

func (s *status) ok(format string, a ...any) {
	if !s.enabled {
		return
	}
	fmt.Fprintf(s.w, "  ✓ %s  (%.1fs, heap %s)\n",
		fmt.Sprintf(format, a...), time.Since(s.stepStart).Seconds(), humanBytes(s.heap()))
}

func (s *status) note(format string, a ...any) {
	if !s.enabled {
		return
	}
	fmt.Fprintf(s.w, "  · "+format+"\n", a...)
}

// tick draws a phase progress line, throttled (100ms TTY / 2s otherwise).
func (s *status) tick(phase string, done, total, a, b int, label string, final bool) {
	if !s.enabled {
		return
	}
	now := time.Now()
	interval := 100 * time.Millisecond
	if !s.tty {
		interval = 2 * time.Second
	}
	if !final && now.Sub(s.lastTick) < interval {
		return
	}
	s.lastTick = now

	elapsed := now.Sub(s.start).Seconds()
	rate := 0.0
	if elapsed > 0 {
		rate = float64(done) / elapsed
	}
	var counts string
	if phase == "evaluate" {
		counts = fmt.Sprintf("sing %d plur %d | ", a, b)
	}
	line := fmt.Sprintf("%s %d/%d %3d%% | %s%4.0f/s | heap %s | %s",
		phase, done, total, pct(done, total), counts, rate, humanBytes(s.heap()), truncate(label, 40))
	if s.tty {
		fmt.Fprintf(s.w, "\r\033[2K%s", line)
		if final {
			fmt.Fprint(s.w, "\n")
		}
	} else {
		fmt.Fprintln(s.w, line)
	}
}

func (s *status) summary(things int) {
	if !s.enabled {
		return
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.HeapAlloc > s.peakHeap {
		s.peakHeap = m.HeapAlloc
	}
	fmt.Fprintf(s.w, "done: %d things in %.1fs | peak heap %s | heap now %s | sys %s | GC %d\n",
		things, time.Since(s.start).Seconds(),
		humanBytes(s.peakHeap), humanBytes(m.HeapAlloc), humanBytes(m.Sys), m.NumGC)
}

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func pct(a, b int) int {
	if b == 0 {
		return 0
	}
	return a * 100 / b
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}
