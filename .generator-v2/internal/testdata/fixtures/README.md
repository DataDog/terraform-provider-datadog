# Fixture Catalogue

Curated test inputs for the `tfgen` end-to-end suite. Each fixture is a self-contained directory holding everything needed to drive the generator and assert on its output: an OpenAPI input spec and the expected golden code the generator should produce from it.

The E2E suite runs `tfgen` against a fixture's `openapi.yaml` and compares the result against the committed `out/` directory. Each fixture therefore exercises one specific generator capability end to end.

## Layout

Each fixture is a directory under this one:

```
<fixture_name>/
├── openapi.yaml         # Input spec with x-datadog-tf-generator annotations
├── out/                 # Expected golden output (committed)
│   └── *.go
└── README.md            # (optional) what this fixture exercises
```

- **`openapi.yaml`** — the OpenAPI spec the generator reads. The resources and data sources to generate are marked with the `x-datadog-tf-generator` extension.
- **`out/`** — the golden output: the `.go` files the generator is expected to emit. These are committed and diffed against on every run. When generator behavior changes intentionally, regenerate this directory and review the diff by hand.
- **`README.md`** *(optional)* — a short note describing the capability the fixture targets (e.g. a particular schema shape, a hook, or an error path).

## Naming

Name each fixture after the artifact it generates so the catalogue reads as a list of capabilities:

- `data_source_<name>` — a generated data source (e.g. `data_source_pet`).
- `resource_<name>` — a generated resource.
- Append a distinguishing suffix when a fixture targets a variant or edge case (e.g. `data_source_team_with_hooks`, `broken_hook_signature`).

> [!NOTE]
> This catalogue is currently empty. Fixtures are added in later phases as the corresponding generator capabilities land.
