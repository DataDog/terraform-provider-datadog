---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
  confluence_id: "7291144348"
---

# 0001 — Resources shipped before test automation

- **Date:** 2026-10-02 (recording a decision taken over 2026-09)
- **Status:** accepted, in progress ([APIR-2494](https://datadoghq.atlassian.net/browse/APIR-2494))

## Context

The generator's original plan sequenced resources *after* automated acceptance-test
generation. The reasoning was that data-source coverage was already high, so the
bottleneck was no longer emitting code but proving a generated artifact safe to
merge — and that solving testing first would make resources straightforward and
keep humans out of the loop.

That is not what happened.

## Decision

Resource generation was implemented and shipped first, with the human-in-the-loop
testing workflow unchanged. Generation landed in #4227 (2026-09-11) and four
generated resources were adopted on 2026-09-28: the Databricks, Elastic Cloud,
Snowflake, and Twilio integration-account resources.

Automated acceptance-test generation and agentic review remain unbuilt.

## Consequences

- **Resources are the path with precedent**, and data sources are the path with
  none — the inverse of the original framing. `generatedDatasources` is still an
  empty slice. The documentation leads with resources accordingly.
- **Completing the acceptance test is still the dominant cost** of adopting a
  generated artifact, and it is now paid on the more complex of the two kinds, since
  a resource test must exercise create, update, and destroy.
- The original reasoning still stands on its merits: testing remains the bottleneck.
  It was simply not a prerequisite for emitting resource code.

## Why this is recorded

The intent was documented as though it were the plan of record across several
reference pages, which then described resources as unimplemented for seven weeks
after they shipped. Intent now lives in decision records so that reference pages can
describe only what runs. See house rule 5 in [../README.md](../README.md).
