---
ddoc:
  confluence_space: "API"
  confluence_parent: "7290816063"
  confluence_id: "7291438200"
---

# Completing the acceptance test

**Who this is for:** whoever is finishing a generated pull request.

tfgen emits a scaffold into `datadog/tests/`. It is a starting point with TODOs, not
a test. tfgen will not overwrite the file once it exists, so your edits are safe
across regeneration.

Use the test function name **as it appears in the generated file**. If another
instruction quotes a differently-capitalised name, the file wins.

## What the scaffold gives you

- a framework provider test case
- a placeholder Terraform configuration
- an assertion that the artifact's ID is set
- for a plural data source, an additional collection assertion
- TODOs where test data and real assertions belong

## Arrange the object the test needs

A data source reads something that already exists; a resource creates it. In both
cases prefer creating the object **in the same Terraform configuration** rather than
relying on permanent state in the test organisation — a test that depends on an
object someone created by hand in the UI will eventually fail for reasons unrelated
to your change.

- Create the prerequisite with its Terraform resource where one exists.
- Pass its ID, or a unique search value, into the artifact under test.
- Use `depends_on` where Terraform cannot infer ordering.
- Keep names deterministic with the existing unique-name helpers — replay matches on
  the request URL, so a random name produces a request the cassette has never seen.

## Assert something that would catch a regression

Replace the placeholder checks. The bar is: *if the generated state mapping broke,
would this test fail?*

- Assert real values, not just that `id` is set.
- Populate optional fields so their mappings are exercised at all.
- Include nested objects and arrays where the response has them.
- For a plural data source, record more than one item.
- For a resource, exercise the full lifecycle: create, then **update in place**, then
  destroy. An update step is what catches a broken Update body or a field that
  should have been ForceNew.
- Cross a pagination boundary when the endpoint and the test org make that practical.

Not every combination needs its own test. The cassette should cover the request
shapes and state mappings that actually matter for this artifact.

## Then record it

[cassettes.md](cassettes.md).

## Testing tfgen itself

If you changed the generator rather than consumed it:

```sh
make tfgen-test
```

That covers the parser, model, SDK-binding, emission, CLI, and split tests, plus the
golden snapshots in `internal/snapshots/` — which compare rendered output byte for
byte, so an intentional emitter change shows up as a reviewable snapshot diff.
Refresh them deliberately with `make tfgen-update-goldens`.

For focused debugging, the small specs in `internal/testdata/mini-oas/` can be fed
straight to `tfgen generate --spec`. The base `mini-datadog_<name>.yaml` files carry
no annotation; annotated copies live under `mini-oas/scripts/gen-test/`.

Generator tests prove tfgen behaves as implemented against known inputs. They are
not a substitute for an acceptance test against the API.
