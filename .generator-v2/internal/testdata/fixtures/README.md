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
├── terraform/           # (resources) config driving the cassette lifecycle
│   └── main.tf
├── cassettes/           # (optional) recorded HTTP traffic for the lifecycle
└── README.md            # (optional) what this fixture exercises
```

- **`openapi.yaml`** — the OpenAPI spec the generator reads. The resources and data sources to generate are marked with the `x-datadog-tf-generator` extension.
- **`out/`** — the golden output: the `.go` files the generator is expected to emit. These are committed and diffed against on every run. When generator behavior changes intentionally, regenerate this directory and review the diff by hand.
- **`terraform/main.tf`** *(resource fixtures)* — the Terraform configuration the E2E suite drives through create → refresh → update → destroy against the fixture's cassettes.
- **`cassettes/`** *(optional)* — recorded and sanitized HTTP traffic for that lifecycle, replayed via `go-vcr`.
- **`README.md`** *(optional)* — a short note describing the capability the fixture targets (e.g. a particular schema shape, a hook, or an error path).

## Reproducing a golden

Generating into a bare temporary directory does not always reproduce a committed `out/` — the generator reads parts of the provider checkout through paths relative to `--output-root`. Each fixture's own README states what its golden depends on.

## Naming

Name each fixture after the artifact it generates so the catalogue reads as a list of capabilities:

- `data_source_<name>` — a generated data source (e.g. `data_source_pet`).
- `resource_<name>` — a generated resource.
- Append a distinguishing suffix when a fixture targets a variant or edge case (e.g. `data_source_team_with_hooks`, `broken_hook_signature`).

## Contents

| Fixture | What it pins |
|---|---|
| [`resource_incident_type`](./resource_incident_type/) | A full-CRUD resource from a real v2 slice: the three-body schema merge, the JSON:API request envelope and its `type` discriminator, per-role request components, a request-settable nested object and an enum leaf inside it. |

More fixtures are added as the corresponding generator capabilities land.
