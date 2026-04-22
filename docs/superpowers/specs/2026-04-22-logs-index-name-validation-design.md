# Design: Logs Index Name Validation

**Date:** 2026-04-22  
**Status:** Approved

## Problem

The Datadog API rejects log index names that contain characters outside `[a-z0-9-]` with a 400 at apply time:

> Scope name must be a lowercase letter followed only by digits, lowercase letters or - character.

This means invalid names (e.g. `mbff-user_trips` with an underscore) only fail during `tofu apply`, not `tofu plan`, giving no early feedback.

## Goal

Surface the validation error at plan time so users catch it before any API call is made.

## Constraint

Exact constraint encoded: name must match `^[a-z][a-z0-9-]*$` — starts with a lowercase letter, followed by zero or more lowercase letters, digits, or hyphens.

## Design

### Resource change

In `datadog/resource_datadog_logs_index.go`, add `ValidateFunc` to the `name` field in `indexSchema`:

```go
"name": {
    Description: "The name of the index. Index names cannot be modified after creation. If this value is changed, a new index will be created.",
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    ValidateFunc: validation.StringMatch(
        regexp.MustCompile(`^[a-z][a-z0-9-]*$`),
        "must start with a lowercase letter and contain only lowercase letters, digits, or hyphens",
    ),
},
```

Both `regexp` and `validation` are already imported. This matches the existing `ValidateFunc` pattern used for `reset_time` and `reset_utc_offset` in the same file.

### Test change

In `datadog/tests/resource_datadog_logs_index_test.go`, add a test step with an invalid name and `ExpectError`:

```go
{
    Config:      testAccCheckDatadogLogsIndexInvalidNameConfig("mbff-user_trips"),
    ExpectError: regexp.MustCompile("must start with a lowercase letter and contain only lowercase letters, digits, or hyphens"),
},
```

Add a helper `testAccCheckDatadogLogsIndexInvalidNameConfig(name string) string` that produces a minimal index config. This test does not hit the API — validation fires before the plan is applied.

## Files Changed

| File | Change |
|------|--------|
| `datadog/resource_datadog_logs_index.go` | Add `ValidateFunc` to `name` in `indexSchema` |
| `datadog/tests/resource_datadog_logs_index_test.go` | Add one test step + config helper for invalid name |

## Out of Scope

- Length limit (not stated by the API constraint)
- Validating other string fields on this resource
- Migrating existing `ValidateFunc` usages to `ValidateDiagFunc`
