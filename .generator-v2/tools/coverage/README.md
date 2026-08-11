# tfgen coverage sweep

This harness measures which Datadog v2 read surfaces tfgen can turn into data
sources that both generate and compile. It is deliberately non-destructive:
generated provider files are discarded after every emit probe and build batch,
and the command fails if the provider generation paths were dirty before the
run or differ afterward.

Every sweep rebuilds `bin/tfgen` from the checked-out branch before probing.
An existing binary is never reused, so reports cannot accidentally measure a
different branch.

## Run it

From the provider repository root:

```bash
./.generator-v2/tools/coverage/run.sh
```

The default run uses the api-spec checkout at
`/Users/jason.tenczar/projects/datadog-api-spec`, reads
`spec/v2/full_spec.terraform.yaml`, probes every addressable candidate with
tfgen, and compile-checks emit-clean candidates in attributed batches using
`make build`.

Useful development modes:

```bash
# Recompute the set model and exports without invoking tfgen.
./.generator-v2/tools/coverage/run.sh --enumerate-only

# Use the curated slim bundle for a quick harness check.
./.generator-v2/tools/coverage/run.sh --fast

# Exercise a deterministic prefix of the candidate inventory.
./.generator-v2/tools/coverage/run.sh --fast --limit 10
```

Use `--help` for path overrides and filtering. Results are written under
`.generator-v2/tools/coverage/results/`; `latest.csv`, `latest.md`, and
`latest.json` are refreshed on every completed run.

## Set model

Resource identity comes from the canonical OAS collection path: a trailing path
parameter is removed from a by-id route, while embedded parameters remain. This
pairs `/widgets` with `/widgets/{widget_id}` without merging unrelated resources
that happen to end in the same noun.

- `S`: a resource has a single-object read, either a by-id GET or a singleton
  GET whose JSON:API `data` member is an object.
- `P`: a resource has a true collection GET whose JSON:API `data` member is an
  array.
- `S∩P`: both endpoint kinds exist.
- `S\P`: only a singular endpoint exists.
- `P\S`: only a true collection endpoint exists.

Singleton GETs are therefore singular candidates, never plural candidates.
Responses without a JSON:API `data` envelope and the dashboards surface are
named permanent exclusions and do not enter the denominators.

Coverage views use their natural denominators:

- singular: passing singular candidates divided by SDK-bound `S`;
- plural: passing plural candidates divided by SDK-bound `P`;
- both: resources in `S∩P` for which both generated artifacts compile, split
  into full and degraded response-shape compatibility;
- in-scope spec surface: the same views before removing `sdk-not-released`, so
  the release-timing gap remains visible without depressing the headline
  generator percentage;
- total OAS inventory: the in-scope spec surface plus explicit `out-of-scope`
  rows, reported as a count rather than blended into either coverage percentage.

The CSV contains singular and plural probe rows plus derived `both` rows. The
Markdown report contains the set views, coverage matrix, gap histogram,
exclusions, and an explicit unclassified count.

## Safety

The harness only generates names prefixed with `tfgen_coverage_`. Cleanup is
limited to those generated Go files and example directories plus restoration
of `datadog/fwprovider/datasources_generated.go`. It never edits docs, tests, or
existing data sources. Build verification always uses `make build`; it never
invokes raw Go build or test commands.
