# Risk heuristics

**This is a quick scan, not an audit — a few minutes, not a code review.** The generator is
solid and its defects surface as build/report failures that are already gated before you get
here. So trust the generated code for correctness; your only job is to skim for the handful
of *material* runtime risks below that clearly apply to this endpoint, and flag those. If
none obviously applies, say "no material risks found" and move on. Don't diff against
goldens, don't re-verify field-by-field, and don't pad the list with boilerplate.

## Fast scan (checklist below is the reference, not a per-field to-do)
Glance at the backing operation(s) for the few things that actually bite: does the **list**
payload differ from the **detail** payload, are there enums / wide integers / required
fields, is the list paginated, and (for plural) is "get all of X in the org" even a sensible
data source. Flag only what jumps out.

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

## Prior art / context (only if a flagged risk needs backing — don't go spelunking)
- `rum_application`: PRs #2215, #3185, #3250, #1641.
- plural semantics: #759.
