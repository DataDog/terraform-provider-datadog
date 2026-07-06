# Risk heuristics

Read this every run. These are the logical, runtime-level risks the reviewer needs
flagged — the generator is solid, so most real defects already surface as build errors or
report failures and are gated before you get here. What's left is code that is *valid* but
whose scenario default may be a poor fit for the endpoint. Flag what genuinely applies to
*this* endpoint, and say why. Tie each flagged risk to something concrete in the generated
code or the spec.

## How to reason (before the checklist)
1. Read the spec for the backing operation(s). Note: what the **list** payload contains vs.
   the **detail** payload, which fields are required, enums, numeric widths, and any
   pagination markers.
2. Determine what the endpoint *does* — is "get all of X in the org" even a sensible
   Terraform data source? Is there a natural filter, or is the id computed?
3. Compare against the emit golden for the scenario to see the intended shape, then judge
   whether that shape fits this endpoint.

## Plural (`cardinality: plural`, list read)

- **List-all / no filter.** Plural sources typically return the *whole org* with no filter
  input (the `id` is computed, not an input). Flag that the reviewer / user must narrow
  client-side. In tests, **never** assert `<collection>.# == "1"` against a live org — the
  count is unstable. The correct pattern is `TestCheckTypeSetElemNestedAttrs` (set
  membership), not a fixed index. Check the generated test for a brittle count/index assertion.
- **Item type must match the SDK response element.** The mapped item type must equal the
  SDK response `.Data` element type (e.g. `RUMApplicationList`'s element). A mismatch is a
  logical bug even if it compiles.
- **Silent-empty trap.** The SDK response `UnmarshalJSON` swallows strict-parse failures
  (unknown enum value, missing required field, `int32` overflow such as a large `org_id`)
  into an `UnparsedObject` and returns an **empty** list — no error. So an "empty" result
  can be a parse failure, not a truly empty org. If the endpoint has enums, required
  fields, or wide integers, flag this and require cassette verification.
- **Pagination.** If the operation is paginated (`.Read.Paginated` / `x-pagination`), the
  generated read should use the SDK's `...WithPagination` method. If pagination markers
  exist but the generated code doesn't paginate, flag it — results will be truncated.
- **Read-after-write lag.** create-then-list within a single `apply` can transiently
  return 0. This is the canonical reason a plural can build + report `created` cleanly yet
  return nothing live. Always call this out for plurals and require cassette replay before
  claiming it works.

## Singular (by-id / both / search-only)

- **`both` type-degradation.** In the `both` (id-optional) scenario, if the by-id detail
  type ≠ the list item type, the generator auto-degrades to by-id-only. Note this in the
  PR — the search path the annotation implied may not actually be wired.
- **`both` needs id; search must resolve exactly one.** The search path must resolve to
  exactly one match and error on 0 or >1. Check that the generated search enforces this.
- **404 / empty handling.** Detail reads should map absence via a `GetXOk`-guarded mapper
  so a missing value becomes `null`, not `""`. An empty-string default where `null` is
  correct is a logical risk.
- **Sensitive fields on detail but not list.** Fields like `client_token` may exist on the
  detail payload but not on list items. For a `both`/search source, selecting via list and
  then expecting a detail-only field will yield null. Flag any field the schema exposes
  that the chosen read path can't populate.

## Prior art / context
Relevant history worth checking when reasoning about a specific source:
- `rum_application`: PRs #2215, #3185, #3250, #1641.
- plural semantics: #759.

If you have access to the coding agent with repo context, it's often faster to ask it
whether a specific heuristic applies to this endpoint than to infer from the spec alone.
