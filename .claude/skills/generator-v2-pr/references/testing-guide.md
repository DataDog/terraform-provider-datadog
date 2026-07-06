# Manual testing guide (record / replay)

This produces the "How to test" section of the PR. Cassettes are recorded against the
**Datadog Frog org** and replayed offline. Fill the confirmed values into the PR body.

## Concept
Acceptance tests run in two modes: **record** (`RECORD=true`) hits the live Frog org and
writes an HTTP cassette; **replay** (`RECORD=false`) runs the same test offline against the
recorded cassette — this is the gate CI enforces. A generated test ships as a scaffold
until someone records it once against Frog.

## Get Frog test-org credentials
Pull the frog test-org keys and map them to the env vars the recorder reads. Do **not**
set `DD_TEST_SITE_URL` — the keys auto-route to frog.

```bash
eval "$(dd-auth --domain frog.datadoghq.com --force-app-key --no-cache --output)"
export DD_TEST_CLIENT_API_KEY="$DD_API_KEY"
export DD_TEST_CLIENT_APP_KEY="$DD_APP_KEY"
# sanity check — both must be non-empty
echo "api=${DD_API_KEY:0:6}…  app=${DD_APP_KEY:0:6}…"
```

## Record
```bash
# from repo root
make testacc RECORD=true TESTARGS='-run <TestAccDatadog…Datasource>'
```
This performs the real API round-trip and writes/updates the cassette. Commit the cassette
**and** its `.freeze` companion (frozen clock) alongside the test.

## Replay (what CI runs)
```bash
make testacc RECORD=false TESTARGS='-run <TestAccDatadog…Datasource>'
```
"Replays green" means this passes offline with no network. Only then may the PR claim the
source is verified. Use `make testacc` (not `make testall`, which also runs unit tests) for
a single data-source test.

## Cassette storage & naming
- Path: `datadog/tests/cassettes/`
- Files map to the test name exactly:
  - `datadog/tests/cassettes/<TestName>.yaml` — the recorded HTTP interactions
  - `datadog/tests/cassettes/<TestName>.freeze` — the frozen clock companion
- Commit both.

## Plural-specific test shape (verify the generated test uses it)
For a list-all plural data source, the generated test should:
- Add `depends_on = [<seed resource>]` on the data source so Terraform creates the seed
  **before** reading the list (the data source has no inputs, so TF sees no ordering link
  otherwise — without it the list read can fire before the create and the seed is missing
  from the recording).
- Assert with `TestCheckTypeSetElemNestedAttrs("...", "<collection>.*", {...})` (set
  membership), **not** a fixed index like `<collection>.0.name` and **not** a count like
  `<collection>.# == "1"`. The org holds other ("rogue") apps in unknown order; the count
  is unstable and a fixed index is brittle. See `risk-heuristics.md` (plural section).

## What to write in the PR
- If a cassette was recorded and replays green this run: say so and name the cassette file.
- If not: "Cassette is a scaffold — record once against the Frog org, then replay to verify.
  Steps: <record command>, then <replay command>." Flag any risk from
  `references/risk-heuristics.md` that specifically needs live verification (e.g. plural
  silent-empty trap, read-after-write lag).

## Troubleshooting
| Symptom | Cause / fix |
|---|---|
| Assertion fails: seeded entity not in list | `depends_on` missing → list read before create. Keep the `depends_on`. |
| Replay fails after a code change | The request signature changed; re-record and recommit the cassette + `.freeze`. |
| `Can't configure a value for "id"` | Config set `id` on a no-input data source. Remove it. |
| `make docs` shows no new file | Data source not registered — recheck the `Datasources` slice in `framework_provider.go`. |
