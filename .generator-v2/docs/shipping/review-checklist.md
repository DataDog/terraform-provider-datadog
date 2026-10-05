---
ddoc:
  confluence_space: "API"
  confluence_parent: "7290816063"
  confluence_id: "7291634859"
---

# Reviewing generated code

**Who this is for:** whoever approves a generated pull request — including the
author, since generated code gets less scrutiny than hand-written code precisely
because it looks uniform.

Generation is deterministic, so you are not reviewing for typos. You are reviewing
whether the *right* thing was generated. Work the list against the diff and the run
report together.

## The annotation

- [ ] `artifact_name` reads well in HCL. Practitioners type `datadog_<name>`, and it
      is effectively permanent once released.
- [ ] `tf_description` describes what a practitioner gets, not how the endpoint is
      implemented. It is the first line of the published documentation.
- [ ] The artifact kind matches what practitioners need to do, not what was easiest
      to annotate.
- [ ] For a resource with no `update`: replace-on-change is intended and the
      description says so.

## The schema

- [ ] Attribute names, types, and optionality look right against the API reference.
- [ ] Required vs Optional+Computed is correct — a field the API populates should not
      be plain Optional, or practitioners get perpetual drift.
- [ ] Descriptions came through. Empty descriptions mean the OpenAPI properties lack
      them; fix the spec rather than the generated file.
- [ ] Enums are validated, and the member list matches the spec.
- [ ] Secrets: anything credential-shaped is write-only or sensitive, and you know
      which of those keeps the value out of state. See
      [reference/scope.md](../reference/scope.md).

## The request and state mapping

- [ ] The SDK method called is the one the annotation named.
- [ ] Path and query parameters are bound to the right schema fields.
- [ ] For a resource, the update body sources `data.id` from state.
- [ ] Omitted API fields map to Terraform **null**, not to empty strings or zeros.
- [ ] A `oneOf` mapper rejects ambiguous or unmatched unions rather than producing
      partial state.

## The run report

- [ ] Every `info` diagnostic is something you are content to lose. Dropped audit
      fields are normal; a dropped field practitioners need is not.
- [ ] No `warning` left unexamined — an unresolved warning is the most common way a
      real problem reaches merge.
- [ ] Nothing in the diff contradicts the report.

## The test and cassette

- [ ] The test would fail if the state mapping broke. Asserting only that `id` is set
      does not meet that bar.
- [ ] For a resource: create, update, and destroy are all exercised.
- [ ] The cassette was read, not just produced — see [cassettes.md](cassettes.md).
- [ ] `RECORD=none` passed against the live API.

## Registration and blast radius

- [ ] The artifact is registered in exactly one place, and the generated registry
      file was not hand-edited.
- [ ] If `overwrites` was used: the hand-written constructor is gone, nothing else
      referenced its model types, and the package still builds.
- [ ] The diff contains only this artifact's files. Anything else means the split or
      the run was wider than intended.

## When to reject rather than fix

Push the fix upstream to the annotation or the spec — not into the generated file —
when the problem is a missing description, a wrong artifact kind, or an absent enum.
Editing generated code is undone by the next run, and the marker makes CI reject
hand edits anyway.
