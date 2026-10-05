---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
---

# Extending tfgen

**Who this is for:** you are adding support for something the generator currently
rejects.

## Before writing code

Check [reference/scope.md](../reference/scope.md) for whether the thing is
*deliberately* unsupported. `anyOf` and a `oneOf` as a map value are known gaps;
dynamic-value fallbacks are a deliberate non-goal, and adding one would change the
generator's contract from "fails honestly" to "emits something approximate".

## Adding a schema shape

Work outward from the parser; each stage will tell you what it is missing.

1. **Parse** — stop classifying it as unsupported in `internal/parser/schema.go`, and
   represent it in the recursive `Schema`. Preserve component identity and
   `UnsupportedReason` for the sub-cases you still reject.
2. **Model** — map it to a Framework type in `framework_types.go` and represent it in
   the `AttributeTree`. If a resource can carry it, handle it in the create/update/read
   merge in `merge.go` — in particular decide how it widens for validation and how it
   ORs sensitivity.
3. **Bind** — if it changes the SDK call surface (new argument type, new setter
   shape), handle it in `sdkbind`/`sdkbinding`.
4. **Emit** — render it, and map state both ways. Omitted values must become null.
5. **Goldens** — add a fixture, regenerate, read the diff.

Add a fixture at each stage rather than at the end. `internal/testdata/mini-oas/`
holds production-shaped slices; `mini-oas/scripts/gen-test/` holds annotated copies;
`internal/testdata/fixtures/` holds full fixtures with committed `out/` directories.

## Adding an artifact kind or lifecycle role

The model has one `Artifact` type and lifecycle roles rather than per-kind structs.
That is what made resources mostly a matter of populating Create/Update/Delete and
adding a template. Follow the same route: add the role, populate it in
`lifecycle.go`, bind it, template it. Resist adding a parallel artifact type.

## Changing a contract

The annotation contract in `internal/contracts/` is consumed outside this module — by
spec authors and by `openapi-transformer`. So:

1. Edit `tracking-field.schema.json`.
2. Run `make tfgen-docs` and commit the regenerated reference page in the same
   change. `make tfgen-docs-check` fails otherwise.
3. Check whether `openapi-transformer` also needs the field. **A field accepted
   upstream but absent here aborts the whole run**, which is exactly how `ignore`
   became a documented feature that broke generation.

Adding a new contract — a run-report schema being the obvious one — means an
`//go:embed` directive plus an exported var in `contracts.go`, and a `page` entry in
`cmd/tfgen-docs/main.go`. The schema renderer is generic; nothing else is needed to
get a generated reference page out of it.

## Testing your change

```sh
make tfgen-test              # unit + goldens
make tfgen-test-integration  # SDK corroboration + generated-code compile gate
make tfgen-docs-check        # reference pages still match the contracts
```

Then generate into a scratch directory against a real slice and read the output:

```sh
make tfgen-build
./bin/tfgen generate \
  --spec .generator-v2/internal/testdata/mini-oas/scripts/gen-test/datastore.yaml \
  --output-root /tmp/tfgen-scratch/datadog/fwprovider \
  --examples-output-root /tmp/tfgen-scratch/examples/data-sources \
  --report -
```

Generator tests prove behaviour against known inputs. They do not prove the
generated artifact works against the API — that still needs a recorded cassette.
