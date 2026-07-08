# Phase 1 — Collecting inputs

The only output of this phase is a **complete, validated parameter set** for
`slice_and_annotate.py`. Generate nothing here. Confirm every value with the user before
handing off to Phase 2 — a wrong operationId or cardinality wastes the whole generation run.

## The parameter set to produce

| Parameter | Script flag | Required | How to determine |
|---|---|---|---|
| Spec path | `--spec` | yes* | Default: curl the upstream v2 spec to a temp file. An explicit path or `$DATADOG_OPENAPI_V2_SPEC` overrides (see below). |
| Read group | `--read` / `--search` | yes (≥1) | operationIds — via discovery or direct (see below). |
| Cardinality | `--cardinality` | no | `singular` (default) or `plural`. |
| Artifact name | `--artifact-name` | yes | default from the resource; validate. |
| TF description | `--tf-description` | no | default sentence unless the user wants custom. |
| Overwrite target | `--overwrites` | no | auto-detected constructor, confirmed. |

\* Always pass `--spec` explicitly (the curled temp file, or the override); confirm it exists first.

## 1. Resolve the spec path

Routes can't be discovered without a local copy of the spec. Resolution order:
1. An explicit path the user gives.
2. `$DATADOG_OPENAPI_V2_SPEC`, if set and present.
3. **Default** — curl the upstream Datadog v2 spec (datadog-api-client-go `master`) to a
   temp file:

   ```bash
   SPEC="$(mktemp)"
   curl -fsSL https://raw.githubusercontent.com/DataDog/datadog-api-client-go/refs/heads/master/.generator/schemas/v2/openapi.yaml -o "$SPEC"
   ```

Reuse the same `$SPEC` for the whole run — discover routes from it, and pass it to the
script as `--spec "$SPEC"`. Confirm the curl succeeded (non-empty file). If you have no
network and no local override, you can still take operationIds directly from the user, but
you can't validate them until Phase 2 — tell the user that.

## 2. Determine the read group (the routes)

Two paths, per the user's starting point:

**Discovery (preferred when the user names a resource/service and a spec is available).**
List candidate GET operations and propose the scenario:

```bash
# every operationId + method + path, filtered to the thing of interest
python3 -c "import yaml; s=yaml.safe_load(open('$SPEC'));
[print(n['operationId'], m.upper(), p) for p,i in s['paths'].items()
 if isinstance(i,dict) for m,n in i.items()
 if isinstance(n,dict) and 'operationId' in n]" | grep -i <resource>
```

- **by-id GET** — path has a `{param}` (e.g. `GET /api/v2/team/{team_id}` → `GetTeam`).
- **collection GET** — no `{param}` (e.g. `GET /api/v2/team` → `ListTeams`).

Present the candidates and let the user pick. **Direct** path: the user just gives the
operationIds; validate they exist in the spec.

## 3. Decide cardinality + scenario

The scenario follows from which GETs the user wants and the cardinality they choose:

| The user wants… | GETs used | Flags | Scenario |
|---|---|---|---|
| One record, by id only | by-id | `--read GetX` | singular, read-only |
| One record, by id **or** filter a list | by-id + list | `--read GetX --search ListX` | singular, "both" (id-optional) |
| One record, resolved from a list | list | `--search ListX` | singular, search-only |
| The whole filtered collection | list | `--read ListX --cardinality plural` | plural |

> ⚠️ For **plural**, the list op goes in `--read`, not `--search`. `search` only helps a
> *singular* source resolve a single match. If only a collection GET exists, ask whether
> they want singular search-only or plural — the same op maps differently.

## 4. Artifact name

- Default: the resource name in snake_case, without the `datadog_` prefix (plural gets the
  pluralized form, e.g. `teams`).
- Validate: `^[a-z][a-z0-9_]*$`, ≤64 chars. The script rejects anything else.
- Confirm — this becomes `datadog_<name>` and the generated file paths.

## 5. TF description

- Default (singular): `Use this data source to retrieve information about an existing Datadog <thing>.`
- Default (plural): `Use this data source to retrieve information about existing Datadog <thing>s.`
- `<thing>` is the artifact name with underscores turned to spaces.
- Only prompt for a custom string if the user wants one; otherwise pass the default.

## 6. Overwrite — auto-detect, then confirm

Check whether a hand-written data source already exists for this name:

```bash
ls datadog/fwprovider/data_source_datadog_<name>.go 2>/dev/null
grep -rn "func NewDatadog.*DataSource" datadog/fwprovider/data_source_datadog_<name>.go 2>/dev/null
```

- **If it exists:** find the constructor (`NewDatadog…DataSource`) and confirm it's listed
  in `framework_provider.go`'s `Datasources` slice. Ask the user whether to retire it —
  if yes, pass `--overwrites <ConstructorName>`. The generated file then overwrites the
  hand-written one in place and the generator removes it from `Datasources`.
- **If it doesn't exist:** additive generation, no `--overwrites`.

> The generator only retires hand-written **framework** data sources. If the existing one
> is an SDKv2 entry in `provider.go`'s `DataSourcesMap`, `--overwrites` will fail in Phase
> 2 — flag that to the user instead of passing it.

## 7. Confirm and hand off

Echo the full parameter set back to the user as the exact `slice_and_annotate.py`
invocation Phase 2 will run, and get an explicit go-ahead. Example:

```bash
python3 .generator-v2/internal/testdata/mini-oas/scripts/slice_and_annotate.py \
  --spec "$SPEC" \
  --artifact-name team \
  --tf-description "Use this data source to retrieve information about an existing Datadog team." \
  --read GetTeam --search ListTeams
```

Carry into Phase 2: the invocation above, plus the known **scenario/cardinality** (Phase 3
needs it and should not have to re-derive it).
