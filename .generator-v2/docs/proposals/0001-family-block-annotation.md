# Proposal 0001: One family registry per document, not one annotation per operation

- **Status:** accepted for the container question (root), implemented behind the legacy
  shape — see [Implementation status](#implementation-status)
- **Scope:** `x-datadog-tf-generator` — the annotation surface only. No change to what
  the generator emits.

## Summary

Replace the current *one annotation per operation* shape with a single **family registry**
at the document root: a map keyed by family name, each entry declaring up to three
artifacts under the keys `resource`, `datasource_singular`, and `datasource_list`.

```yaml
openapi: 3.0.0
info: {title: Datadog API, version: "1"}

x-datadog-tf-generator:
  integration_databricks_account:
    resource:
      tf_description: Provides a Datadog Databricks integration account resource.
      group:
        create: CreateDatabricksIntegrationAccount
        read: GetDatabricksIntegrationAccount
        update: UpdateDatabricksIntegrationAccount
        delete: DeleteDatabricksIntegrationAccount
    datasource_singular:
      tf_description: Use this data source to retrieve information about an existing Databricks integration account.
      group:
        read: GetDatabricksIntegrationAccount
    datasource_list:
      tf_description: Use this data source to list Databricks integration accounts.
      group:
        read: ListDatabricksIntegrationAccounts

paths:
  /api/v2/integration/databricks:
    post:
      operationId: CreateDatabricksIntegrationAccount
      x-datadog-generators: {public: true, generators: {terraform: true}}
    get:
      operationId: ListDatabricksIntegrationAccounts
```

The map key is the family name; the block key replaces `artifact_kind` + `cardinality`.
The list data source derives `integration_databricks_accounts` by pluralisation,
overridable per block. Every other field keeps its current meaning, scoped to the block
that owns it.

## Motivation

### `cardinality` is a field whose meaning depends on another field

The reference documents it as *"For a data source … Ignored for resources."* A flat schema
cannot express that, so the constraint lives in prose and in `buildArtifact`, which already
dispatches on the `(kind, cardinality, group-shape)` triple. The block key *is* that
discriminator, named once.

### A family is declared three times

A resource plus its two data sources means three annotations on three operations, each
repeating `artifact_name` and its own copy of `group`. Nothing ties them together, so
nothing catches drift — visibly, in the Databricks annotation as written today, where the
data source is described as *"Provides a Datadog Databricks integration account
resource."* One entry makes that one description to get right.

### The annotation's position already carries no meaning

Group references resolve through a **global** `operationId` index
([`indexOperationsByID`](../../internal/parser/group.go)), never relative to the annotated
operation. So the operation an annotation sits on is purely somewhere to put it: the
resource block lives on `post` by convention alone, and the singular data source could
equally be declared from the `get`.

An earlier draft of this proposal moved the block to the path item (sibling of
`get`/`post`). That was still arbitrary, because a family spans two path items —
`…/databricks` hosts create and list, `…/databricks/{id}` hosts read, update and delete —
so it needed an enforced rule about which of the two may host the block, plus a diagnostic
for getting it wrong. Lifting one more level deletes that rule and its diagnostic outright:
a family is a property of the document, not of any one path.

There is precedent for the key being location-polymorphic: `sensitive` already applies
*"when present on a Schema Object"*, so one flat schema already validates occurrences in
two unrelated positions. Splitting by location is overdue regardless.

## Design

### D1 — The registry lives at the document root

Verified against libopenapi v0.37.2: extensions are extracted at the document root, on the
`Paths` object, and on each path item, and a root- or `paths`-level `x-datadog-tf-generator`
is **not** mistaken for a path — only real path keys appear in `Paths.PathItems`.

Root is preferred over `paths:` level. Both are legal OpenAPI and both are reachable, but a
`paths:`-level extension reads as *about the collection of paths*, which will surprise
anyone who thinks of `paths` as a plain map; a document-wide registry belongs at the
document level, next to `info` and `servers`.

What this buys, concretely: the anchor rule disappears, its diagnostic disappears, and the
whole annotation surface of a spec is one reviewable hunk.

What it costs — and this is the real trade — is **locality**. The annotation no longer sits
next to the operations it names, so an endpoint pull request that renames an `operationId`
will not show the annotation in its diff. Two things make that tolerable:

- A group naming an unknown `operationId` is already reported by
  `unresolvedGroupDiagnostics`, so a rename breaks loudly at generation time rather than
  silently producing a wrong artifact. The safety net exists today.
- `x-datadog-generators` (`public: true`, `terraform: true`) stays per-operation and
  unchanged, so the opt-in signal remains visible on the endpoint itself.

### D2 — The family name is the map key

```yaml
x-datadog-tf-generator:
  monitor_config_policy:                        # family name, validated as the key
    datasource_singular: {...}                  # → data.datadog_monitor_config_policy
    datasource_list:
      artifact_name: monitor_config_policies_v2 # optional per-block override
      group: {read: ListMonitorConfigPolicies}
```

`artifact_name` stops being a required field: the key carries it, so it cannot be repeated
inconsistently across blocks, and the YAML structure itself makes one name own one family.
Keys are validated with `propertyNames` against the existing
`^[a-z][a-z0-9_]*$` / 1–64 constraint. `artifact_name` survives as an optional
**override** inside any block, for the case where one artifact genuinely diverges.

**Pluralisation** for `datasource_list`, applied to the last underscore-separated word:

| Ending | Rule | Example |
|---|---|---|
| consonant + `y` | `y` → `ies` | `monitor_config_policy` → `monitor_config_policies` |
| `s`, `x`, `z`, `ch`, `sh` | `+es` | `logs_index` → `logs_indexes`, `_status` → `_statuses` |
| anything else | `+s` | `integration_databricks_account` → `…_accounts`, `api_key` → `api_keys` |

Checked against names the provider already ships: `logs_indexes` (not *indices*),
`service_level_objectives`, `restriction_policies`, `team_memberships`,
`sensitive_data_scanner_standard_patterns` all derive correctly.

**Derivation is validated, not trusted** — but not for the reason this proposal first
claimed. An earlier draft wanted a guard against a family name with no distinct plural
(`_metadata`, `_data`) silently putting two data sources at the same Terraform address.
Writing the test showed that guard is unreachable: every rule above either appends or
replaces a suffix, so a derived name can never equal its stem. A mass noun inflects
regularly to `metadatas` — poor reading, not a collision — and that is what the
`artifact_name` override is for.

What is validated is what can actually fire: the derived name is checked against the same
pattern and 64-character limit the contract puts on a written one, since inflection is the
one way to reach a name no author reviewed. A derived name that genuinely collides — with
another family, or with a hand-written artifact — is caught spec-wide by
`CheckDuplicateArtifactNames`, which sees every artifact rather than one family at a time.

**Consequence for the duplicate check.** `CheckDuplicateArtifactNames` must keep keying on
`(kind, artifact_name)` and must **not** gain `cardinality` — a singular and a plural data
source both land in `data.`, so a shared name is a genuine collision the existing check
already rejects. (This corrects an earlier suggestion of mine: with derivation in place the
names differ by construction, and widening the key would only let a real collision
through.) The check stays necessary even with names as map keys, because overrides can
collide and because duplicate YAML keys are not reliably rejected by every parser.

### D3 — `group` keeps today's meaning in every block

`search` stays what it is: *a list endpoint used to resolve a single match*, singular data
sources only. The list data source names its list endpoint in **`read`**, matching
`buildPluralArtifact`. Nothing in the group vocabulary changes.

Because each block declares exactly one artifact kind, the per-kind constraints become
expressible in JSON Schema instead of discovered in Go:

| Block | Required | Forbidden |
|---|---|---|
| `resource` | `create`, `read`, `delete` (per `requireResolvedRoles`) | `search` |
| `datasource_singular` | at least one of `read`, `search` | `create`, `update`, `delete` |
| `datasource_list` | `read` | `search`, `create`, `update`, `delete` |

A resource missing `delete` then fails at parse time with a schema error naming the family
and block, rather than partway through `BuildArtifact`.

### D4 — Per-artifact fields live inside the blocks

`tf_description`, `group`, `id_strategy`, `overwrites`, `skip` and the optional
`artifact_name` override all describe one artifact, so all of them sit inside the block.
`overwrites` names one specific constructor, so it has no coherent family-level meaning.
Any per-artifact field added to the tracking extension later belongs in the block for the
same reason — `cassette`, in flight on another branch, is one such field.

`skip: true` in a block drops that artifact; on the family entry it drops all three.

`sensitive` is unaffected — it lives on Schema Objects, not here.

### D5 — The registry desugars into today's internal model

This is what keeps the change small: **the registry is parse-time sugar.** Each block
becomes one `model.TrackingFieldMetadata` with `ArtifactKind` and `Cardinality` synthesised
from the block key, carried by one `model.Operation` anchored on that block's **primary
operation**:

| Block | Primary operation |
|---|---|
| `resource` | `group.create` |
| `datasource_singular` | `group.read`, else `group.search` |
| `datasource_list` | `group.read` |

Downstream, `BuildArtifact`, `CheckDuplicateArtifactNames`, the `cli/generate.go` loop over
`spec.Operations`, and every later consumer all keep working unchanged. Note this table is no
longer a constraint on authors — it was the anchor rule in the path-item draft, and is now
an internal detail of the fan-out.

Anchoring on the primary operation is what makes the desugaring behaviour-preserving:
`unstableOperationKeys` adds the annotated operation plus every resolved group role, and
for each kind the primary operation is already in the group, so the `x-unstable` gate set is
identical to today's.

Two plumbing details:

1. **Synthetic operations must not pollute the `operationId` index.** Fan-out produces a
   second `*model.Operation` for an already-enumerated operation; `indexOperationsByID`
   documents that *"the last one wins"*, so a group could resolve to the clone. Add
   `Operation.Synthetic bool` and skip those when indexing.
2. **Clones need raw handles.** `LoadSpec`'s `rawOps` map is keyed by `*model.Operation`
   pointer, so each clone must inherit the primary's entry (`rawOps[clone] =
   rawOps[primary]`) or schema normalisation will not find its request and response bodies.

Ordering inside `LoadSpec`: enumerate real operations → build the id index → read the root
registry and fan out → sort → duplicate check → `ResolveOperationGroups` → normalise.

Diagnostics improve by losing the path: instead of *"invalid tracking-field extension on
POST /api/v2/… (operationId X)"*, errors name the address the author actually wrote —
`family "integration_databricks_account", block datasource_list: …`.

### D6 — Two schemas instead of one

Split `tracking-field.schema.json` by location:

- `family-registry.schema.json` — the root registry: `propertyNames` for family names, the
  three optional block objects, `additionalProperties: false` at every level.
- `schema-annotation.schema.json` — the `sensitive` flag on Schema Objects.

`cmd/tfgen-docs` renders [`annotation.md`](../reference/annotation.md) from the contract, so
the reference regenerates with `make tfgen-docs` and needs no hand editing — the docs
generator just renders two sections instead of one.

## Migration

Annotations live in the API spec repositories; `rg` finds **zero** occurrences of
`x-datadog-tf-generator` outside `.generator-v2`. A hard shape flip breaks every annotated
spec at once.

1. Land `family-registry.schema.json` **alongside** the existing schema. The parser reads
   the root registry when present and operation-level annotations otherwise —
   discriminating on location, so there is no `oneOf` ambiguity and both may coexist during
   the transition.
2. Migrate the spec repositories family by family. Both shapes generate byte-identical
   output, so each migration reviews as a no-op diff.
3. Emit a deprecation diagnostic on operation-level annotations once the spec repos are
   clean, then delete that branch.

In-repo cost: **45 fixture files, 112 occurrences** under `internal/testdata` and
`examples`. Mechanical, but it is the bulk of the work and worth scripting rather than
hand-editing.

## Upstream changes required

The registry cannot be adopted by `datadog-api-spec` until `openapi-transformer` learns it.
Surveyed 2026-10-12 against `openapi-transformer` @ `ca370cf` and
`openapi-transformer-tf-dispatch` @ `c4b8128`.

How the pipeline reaches tfgen: 278 per-domain fragments under `spec/v2/**` (partial
documents — `components`/`paths`/`tags`, no `openapi`/`info`) are merged by
`parser/folder_loader.py`, encoded to `spec/v2/full_spec.terraform.yaml` with only
terraform-allowed extensions, and tfgen runs on that bundle
(`config/client_generator_config.yaml`). So a root registry does land exactly where
`DecodeRegistry` looks — in the bundle's document root.

**This is also the answer to the merge-point worry.** Because each domain owns its own
fragment, every team writes its registry at the top of its own file: no shared block, no
human merge contention. What has to change is the machine merge.

### 1. The decoder rejects the shape outright — verified

`OpenAPIObject.known_extensions` (`openapi/base.py`) maps extension key → class with no
regard for location, and `parser/decoder.py:_decode_extension` instantiates that class
wherever the key appears. A root `x-datadog-tf-generator` is therefore decoded as the
operation-level `XDatadogTfGenerator`, whose fields are `artifact_kind`/`artifact_name`/…,
and the family name is an unregistered field. Loading the document raises:

```
ValueError: Unexpected key 'integration_databricks_account' at
'#/x-datadog-tf-generator/integration_databricks_account' — no field registered
```

Needs location-aware dispatch (key + parent type) or a wildcard-keyed registry class
alongside the existing one, so both shapes decode during the migration.

### 2. Fragment merge silently keeps only the first registry — verified

`folder_loader.py` merges `components`, `paths`, `security` and `tags` structurally, and
copies every *other* root key only when absent:

```python
if key not in _SKIP_MERGE and key not in self._raw_data:
    self._raw_data[key] = val
```

Two fragments each declaring a root registry key therefore collapse to whichever loads
first, with no error. Demonstrated with a two-fragment probe: `paths` merged to
`['/alpha', '/zeta']` while the registry merged to `['family_from_alpha']` — the second
domain's families vanished.

That is the most dangerous failure mode in this list: it loses artifacts quietly. The
registry key needs a structural union keyed by family name, with a duplicate-family error.
The same first-wins applies to a `paths`-level key, so moving the registry down one level
would not avoid it.

### 3. tf-dispatch would stop dispatching Terraform — from code, not tested

`spec.py:terraform_operation_ids` builds the set of operations Terraform cares about by
iterating operations and reading `operation.x_datadog_tf_generator`; `diff/spec_diff.py`
uses it to decide whether a spec change affects the terraform generator. Under a registry
**no operation carries the extension**, so the set is empty, no change ever looks
Terraform-affecting, and tfgen is never run. This fails silently too — the pipeline simply
stops producing provider PRs. It must read the root registry and union each block's group
operationIds.

### 4. The Python validation rule needs a sibling

`validation/rules/validate_x_datadog_tf_generator.py` duplicates, in Python, what this
change now expresses in the JSON Schema. It is bound to the operation shape twice over:
`RULE_DEFINITION["object"] = "XDatadogTfGenerator"`, and every message is addressed via
`cast("Operation", extension.parent()).operation_id`, which has no meaning when the parent
is the spec root. A registry rule should address errors by family and block, and can drop
the checks the schema now enforces structurally (per-kind group constraints).

### 5. No change needed in `generators.py`

Worth stating because it is the one place that looks like it would need touching:
`GENERATOR_ALLOWED_EXTENSIONS["terraform"]` already lists `x-datadog-tf-generator`, and
`parser/encoder.py:encode_object_extensions` filters by extension *name* only, on any
object. A root-level registry survives into `full_spec.terraform.yaml` with the allowlist
untouched.

### 6. `datadog-api-spec` migration is small

Only six files mention the extension: the Databricks, Elastic Cloud, Snowflake and Twilio
interface specs, `actions_datastore/datastores.yaml`, and a describing comment in
`config/client_generator_config.yaml` that says tfgen "reads operations annotated with
x-datadog-tf-generator" — which the comment will need to stop saying. Far smaller than the
45 fixture files in this repository.

### Sequencing

1. openapi-transformer: decode (1), merge (2), validate (4) — the registry round-trips but
   nothing authors it yet.
2. tf-dispatch: (3), before any spec adopts the registry, or Terraform generation goes
   quiet.
3. datadog-api-spec: migrate the six files, one domain at a time; output is byte-identical
   per `TestRegistryShapeGeneratesIdenticalSource`.
4. This repository: migrate fixtures, then drop the operation-level branch.

## Settled questions

1. **Container: the document root.** Decided. The remaining risk is operational rather than
   technical: if `datadog-api-spec` is authored as per-endpoint or per-tag fragments and
   bundled, one root registry becomes a **merge point** — every team adding an artifact edits
   the same block. If that bites, the fallbacks are to author the registry once per bundled
   output file (cleanest, if the bundle is generated anyway), or to move it to a sibling
   `tfgen.yaml` manifest, which is honest about being a manifest but breaks the
   OpenAPI-only-input constraint.
2. **`artifact_name` overrides are permitted in every block**, not just `datasource_list`:
   free in the schema, symmetric, and it covers a resource that must diverge from its data
   source.
3. **Resource-only fields are enforced structurally**, by living in the `resource` block
   rather than being policed in Go. This baseline has none, but it is how `cassette` should
   be admitted when that work lands: as a key in the resource block only.

## Implementation status

Landed on a branch off `master`, with the legacy operation-level shape untouched and still
working — migration step 1. Every pre-existing test passes, `go vet` is clean, and
`make tfgen-docs-check` is current.

The registry carries no `cassette` key: that field does not exist in this baseline, and
adding it here would couple the annotation change to an unmerged stack. It slots into the
`resource` block when that work lands.

| Step | Where | State |
|---|---|---|
| Family-registry contract | `internal/contracts/family-registry.schema.json` | done |
| Decode + desugar + pluralise | `internal/parser/registry.go` | done |
| Fan out, phase ordering | `internal/parser/parser.go` | done |
| `Synthetic` flag, excluded from the id index | `internal/model/types.go`, `internal/parser/group.go` | done |
| Clones inherit raw handles | `internal/parser/parser.go` | done |
| Registry address in errors | `internal/parser/registry.go` | done |
| Unit tests + `Pluralize` table | `internal/parser/registry_test.go` | done |
| Equivalence proof | `internal/registry_equivalence_test.go` | done |
| Registry section in the annotation reference | `cmd/tfgen-docs` | **not done** |
| Fixture migration (45 files, 112 occurrences) | `internal/testdata`, `examples` | not started (gated on the spec repos) |

`cli/generate.go` needed no change, which is the desugaring claim holding up: an untracked
operation already reports as `tracking_field_absent`, so the real operation behind a
synthetic artifact behaves exactly like any unannotated one.

The docs gap is deliberate: `cmd/tfgen-docs` renders a flat single-object schema, and
teaching it the nested registry is its own change. The committed reference still documents
the legacy shape, which is still the shape in use, so nothing it says is wrong yet.

### The equivalence proof

`TestRegistryShapeGeneratesIdenticalSource` drives the real incident-type fixture through
the full offline pipeline twice — once as committed, once with its annotation lifted into a
root registry — and compares the emitted Go source byte for byte. It passes.

That fixture is a better witness than it looks: its annotation sits on the **GET**, not the
POST, while the registry anchors a resource on `group.create`. So the two runs enter
`BuildArtifact` from different operations and still render identically, which is the
strongest available evidence that the anchor is a diagnostics detail and nothing more. The
test asserts that divergence rather than assuming it, so it cannot pass vacuously.
