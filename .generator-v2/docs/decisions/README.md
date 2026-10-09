---
ddoc:
  confluence_space: "API"
  confluence_parent: "6733104627"
  confluence_id: "7291274458"
---

# Decisions

**Who this is for:** you want to know why the generator is the way it is, or why a
plan that is written down somewhere did not happen.

Reference pages describe what runs today. Intent, superseded direction, and
"we chose X over Y" belong here instead, dated, so a reference page is never the
place someone discovers that a feature was only ever planned.

| Record | Decision |
|---|---|
| [0001](0001-resources-before-test-automation.md) | Resources shipped before test automation, inverting the original sequencing. |
| [0002](0002-split-moves-upstream.md) | The split/fan-out workflow moved out of the provider repository. |

Add a record when a choice would otherwise only be legible from a commit message, or
when something documented as the plan stops being it. Unmerged work gets a record, not
a reference page; see house rule 5 in [../README.md](../README.md).
