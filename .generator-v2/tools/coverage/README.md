# Data-source coverage checker

A throwaway tool that answers: **of the resources in an OpenAPI spec, how many
can generator-v2 emit as a singular data source, and how many as a plural one —
and why do the rest fail?**

It works by reusing the generator's own code. For each endpoint it synthesizes a
`data_source` annotation, then runs the real pipeline (`parser.NormalizeSchemas`
→ `model.BuildArtifact` → `emit.BuildDataSourceView` → `emit.RenderDataSource`)
and reads each verdict straight off the generator's errors. Because it calls the
production packages, the numbers track the generator automatically — there's
nothing to keep in sync.

> **Not production code.** This whole directory is unmaintained.
> It's an analysis aid, not part of the generator.

---

## The model (how coverage is defined)

Coverage is measured per **thing** (a resource), not per endpoint:

- A **thing** is keyed by its collection path. A by-id GET at `/X/{id}` pairs with
  a list GET at `/X` into one thing; nested resources group the same way
  (`/teams/{id}/links{/...}`). A thing is counted **once** — never twice for
  having both a by-id and a list.
- **Plural coverage** = of all things, the fraction whose **list** endpoint
  generates a plural data source.
- **Singular coverage** = of all things, the fraction that generate a singular
  data source by **any** mode, counted once:
  - **by-id** read-only (`/X/{id}`),
  - **search-only** from the list (list-the-collection-take-one — no by-id and no
    filter item required), or
  - a **singleton** `/X` that returns a single object.
- The **denominator for both percentages is all things** (shared base), reported
  separately for stable vs `x-unstable`. A thing that lacks what an axis needs is
  a **miss**, not an exclusion (e.g. no list endpoint → a plural miss).
- **"Generatable" means builds AND renders clean** (passes gofmt) — i.e. it would
  compile. The two percentages are reported separately and **never summed**
  (a single resource is usually readable both ways, so adding them double-counts).

Because singular can be reached three ways but plural needs a working list,
singular coverage normally lands **at or above** plural coverage.

---

## Prerequisites

- Go (same toolchain the generator uses).
- `curl` — only if you want the script to auto-download the upstream spec.

No build step or install: the script runs the tool with `go run`.

---

## Run it

From anywhere in the repo:

```sh
.generator-v2/tools/coverage/run.sh
```

With no arguments it fetches the upstream Datadog v2 spec to
`/tmp/dd_openapi_v2.yaml` (if not already there) and prints the summary.

### Point it at a specific spec

```sh
.generator-v2/tools/coverage/run.sh /path/to/openapi.yaml
.generator-v2/tools/coverage/run.sh .generator/V2/openapi.yaml      # the generator's own spec
```

### Pass extra flags (after `--`)

```sh
# list every uncovered thing with its per-axis reason
.generator-v2/tools/coverage/run.sh /tmp/dd_openapi_v2.yaml -- --list-fails

# list every covered thing (singular / plural / both)
.generator-v2/tools/coverage/run.sh /tmp/dd_openapi_v2.yaml -- --list-ok

# show ALL the things under each failure category, not just a 3-thing sample
# (e.g. every endpoint that fails to render / gofmt)
.generator-v2/tools/coverage/run.sh /tmp/dd_openapi_v2.yaml -- --examples -1

# dump the full per-thing report as JSON
.generator-v2/tools/coverage/run.sh /tmp/dd_openapi_v2.yaml -- --json /tmp/coverage.json

# loosen the $ref expansion limit (default 8)
.generator-v2/tools/coverage/run.sh /tmp/dd_openapi_v2.yaml -- --max-depth 16

# quieter / lighter for a huge spec (see "Memory" below)
.generator-v2/tools/coverage/run.sh /huge/spec.yaml -- --no-cycle-gate
.generator-v2/tools/coverage/run.sh /huge/spec.yaml -- --no-progress      # silence stderr

# capture profiles for a deeper look
.generator-v2/tools/coverage/run.sh /huge/spec.yaml -- --memprofile /tmp/heap.pprof
.generator-v2/tools/coverage/run.sh /huge/spec.yaml -- --cpuprofile /tmp/cpu.pprof
#   then: go tool pprof -top -inuse_space /tmp/heap.pprof
```

### Run the binary directly (skip the script)

```sh
cd .generator-v2
go run ./tools/coverage --spec /tmp/dd_openapi_v2.yaml --list-fails
```

---

## Live status

While it runs, a status stream goes to **stderr** (the report and `--json` go to
stdout, so they stay clean). Each phase shows its elapsed time and heap —
reading, building the model, the cycle scan, classifying endpoints — then a live
progress line through the evaluate phase:

```
→ building OpenAPI v3 model (the big one-time allocation)
  ✓ model built  (0.4s, heap 303.2MB)
...
classify 584/584 100% | 1250/s | heap 399MB | done
evaluate 240/406  59% | sing 168 plur 130 |  670/s | heap 470MB | api/v2/teams
done: 406 things in 0.7s | peak heap 509MB | heap now 509MB | sys 598MB | GC 9
```

On a terminal the progress line updates in place; when stderr is redirected it
prints periodic snapshots instead. `--no-progress` turns it all off.

---

## Memory (read before a massive spec)

The tool is memory-safe by construction: every synthesized operation is
normalized, evaluated, then its schema tree is **freed** before the next, so only
one thing's trees are live at a time and peak heap does not grow with the number
of things. Watch the `heap` figure in the status — it should plateau during the
evaluate phase. If it climbs steadily, that's a regression worth investigating
(capture `--memprofile`).

The dominant cost is **not** this tool — it's libopenapi holding the whole spec
in memory. As a rule of thumb the in-memory model is ~40–50× the raw spec size
(a 6.6MB spec → ~300MB heap). So budget roughly: a 50MB spec → ~2–3GB, a 100MB
spec → ~4–6GB. That's unavoidable while libopenapi is the parser.

Two levers if a huge spec strains memory or time:
- `--no-cycle-gate` skips the upfront whole-graph scan that forces every
  component schema to materialize (meaningful peak/time savings on big specs).
  You lose only the "would LoadSpec succeed?" line.
- `--memprofile` + `go tool pprof -inuse_space` confirms where the bytes are
  (expect `libyaml`, `index`, and `SchemaProxy.Schema` at the top — all
  libopenapi, none from this tool).

---

## Reading the output

```
[GLOBAL LOAD GATE: DetectComponentRefCycles, maxDepth=8]
  distinct $ref cycles : 0
  max-depth chain hit  : #/components/schemas/...
  => LoadSpec on this spec would: FAIL (analyzed per-thing instead)
```
Whether the generator's *whole-spec* load would succeed. The full upstream spec
fails this (real `$ref` cycles + over-depth chains), so the tool analyzes each
thing in isolation — mirroring how generation is run against a curated slice.

```
[THINGS]   (a thing = resource keyed by collection path; 584 GET ops total)
  distinct things                 : 406
    with a list endpoint          : 278   (plural-eligible; singular via search)
    by-id only (no list)          : 46    (no plural; singular via by-id)
    singleton only (no id, list)  : 82    (no plural; singular via read)
```
The denominator and its composition. Things without a list can never be plural,
so they're plural misses — but they're still singular-eligible via by-id/read.

```
[COVERAGE — STABLE]   (denominator = all stable things = 238)
  singular (by-id / search / read): 171/238 = 71.8%
  plural                          : 133/238 = 55.9%

[COVERAGE — UNSTABLE]   (denominator = all unstable things = 168)
  ...
```
The two headline numbers, over all things in each stability bucket. Reported
separately, never summed.

```
[SINGULAR — why things failed (best of by-id / search / read)]
   52  map / free-form additionalProperties under attributes
            /api/v2/...
            /api/v2/...
            /api/v2/...
            … +49 more (use --examples -1 to list all)
   46  generated Go fails to compile (gofmt: e.g. `type` keyword collision)
   ...
[PLURAL — why list-having things failed]
   ...
```
Each axis's misses bucketed by the generator's real error, with a sample of the
things in each. `--examples N` sets how many to list per category (default 3);
**`--examples -1` lists every one** — handy for, say, the full set of endpoints
that fail to render. `--list-fails` prints every uncovered thing with its
per-axis reason; `--list-ok` prints the covered ones.

---

## What it does NOT tell you

- **Compile-readiness of SDK bindings.** "Generatable" means the artifact builds
  and renders; it doesn't verify the SDK method/package/type the binding names
  actually exists (e.g. `/api/unstable/` paths resolve to a non-existent
  `datadogUNSTABLE` package).
- **Runtime correctness.** It checks shape and compilation, not behavior.
- **Fidelity.** `oneOf` variants and JSON:API `relationships` are *dropped, not
  fatal* — a thing can be "covered" while silently omitting those fields.
- **Which singular mode would actually be wired.** It credits a thing if *any*
  mode (by-id / search / read) generates; it doesn't decide which one the
  generator should emit, or whether an id-optional "both" data source is wanted.

---

## Files

| File | What it is |
|---|---|
| `run.sh` | Build-and-run wrapper (auto-fetches the default spec). |
| `main.go` | The tool. Reuses `internal/parser`, `internal/model`, `internal/emit`. |
| `README.md` | This file. |

Full written analysis (point-in-time): the Claudebase vault note
`tfgen-datasource-coverage-spec-wide.md`. For current numbers, re-run the tool.
