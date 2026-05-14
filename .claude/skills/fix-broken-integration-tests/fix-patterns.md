# Integration Test Failure Patterns and Fixes

Match the error message from the failing test against a pattern below to identify the right fix strategy.

---

## Pattern 1: Quota / Resource Accumulation

**Errors:**
- `402 Payment Required: Quota reached for api tests`
- `403 Forbidden: Reached the maximum amount of RUM applications`
- `422 Unprocessable Entity: quota`
- `400 Bad Request: Max groups quota reached`

**Root cause:** Stale resources from previous test runs accumulated in the org and hit a per-org limit.

**Fix: Add a sweeper**

Create `datadog/tests/<resource>_sweep_test.go`:

```go
package test

import (
    "sync"
    "testing"
)

var <resource>SweepOnce sync.Once

// cleanup<Resource> removes stale <resource> resources before tests run.
// Uses sync.Once so it runs at most once per test binary execution.
// Safe to call from every failing test — only the first call does real work.
func cleanup<Resource>(t *testing.T) {
    t.Helper()
    if isReplaying() {
        return // never hit the real API during cassette replay
    }
    <resource>SweepOnce.Do(func() {
        doSweep<Resource>(t)
    })
}

// TestSweep<Resource> is a standalone sweep test for CI / manual invocation via `make sweep`.
func TestSweep<Resource>(t *testing.T) {
    doSweep<Resource>(t)
}

func doSweep<Resource>(t *testing.T) {
    t.Helper()
    ctx, client := newSweepAPIClient(t)
    if client == nil {
        return
    }
    // List all resources, delete ones matching test naming patterns or that are stale.
    // Log what was deleted.
}
```

Then add `cleanup<Resource>(t)` as the **first line** of each failing test body,
before `t.Parallel()` and before `resource.Test(...)`.

**Reference:** `datadog/tests/sensitive_data_scanner_sweep_test.go`

---

## Pattern 2: Hardcoded Timestamps

**Errors:**
- `400 Bad Request: end timestamp 2025-01-01 must not be older than 15 months`
- `400 Bad Request: Scheduled downtime start cannot be in the past`
- `400 Bad Request: start time must be in the future`

**Root cause:** Test configs use hardcoded Unix timestamps that are now in the past.

**Fix: Replace with dynamic timestamps using `clockFromContext`**

```go
func TestAccDatadogXxx_Basic(t *testing.T) {
    ctx, accProviders := testAccProviders(context.Background(), t)
    // Generate timestamps relative to test time, not hardcoded
    start := clockFromContext(ctx).Now().Local().Add(time.Hour * 1)
    end := start.Add(time.Hour * 2)

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {Config: testAccXxxConfig(start.Unix(), end.Unix())},
        },
    })
}

func testAccXxxConfig(start, end int64) string {
    return fmt.Sprintf(`
resource "datadog_xxx" "foo" {
  start = %d
  end   = %d
}`, start, end)
}
```

After changing test logic, **re-record cassettes** with `RECORD=true`.

**Reference:** `datadog/tests/resource_datadog_downtime_test.go` (commit `56c0c2e6`)

---

## Pattern 3: Non-Empty Plan After Apply (Provider Read Bug)

**Errors:**
- `After applying this test step, the refresh plan was not empty`
- Diff shows resource attributes changing on every refresh

**Root cause:** The resource's `Read` function returns attribute values that differ from
what was applied — Terraform detects a perpetual diff.

**Fix: Normalize the drifting attribute in the Read function**

1. Identify which attribute drifts (read the plan diff carefully)
2. Find the resource: `datadog/fwprovider/<resource>.go` (plugin framework) or `datadog/<resource>.go` (SDKv2)
3. In the `Read` method, normalize the value to match what the provider writes on create:

```go
// Example: API returns RFC3339Nano but provider stores short RFC3339
if v, ok := resp.GetCreatedAtOk(); ok {
    parsed, _ := time.Parse(time.RFC3339Nano, *v)
    state.CreatedAt = types.StringValue(parsed.UTC().Format(time.RFC3339))
}
```

For SDKv2 resources, add a `DiffSuppressFunc`:
```go
"created_at": {
    Type: schema.TypeString,
    DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
        tOld, err1 := time.Parse(time.RFC3339Nano, old)
        tNew, err2 := time.Parse(time.RFC3339Nano, new)
        if err1 != nil || err2 != nil {
            return false
        }
        return tOld.Equal(tNew)
    },
},
```

Validate with `RECORD=none` after fixing.

---

## Pattern 4: Eventual Consistency (Resource Not Found After Create)

**Errors:**
- `failed to satisfy the condition after N times`
- `Filter keyword returned no results` right after creating the resource
- Datasource returns empty results immediately after a Create step

**Root cause:** The API is eventually consistent — writes don't immediately appear in
list/read endpoints.

**Fix A: Two-step test pattern**

Split a single-step test that creates-then-queries into two steps:
```go
Steps: []resource.TestStep{
    {
        // Step 1: create the resource only
        Config: testAccCreateOnlyConfig(name),
    },
    {
        // Step 2: add the datasource query in the same config
        Config: testAccWithDatasourceConfig(name),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr(...),
        ),
    },
},
```

**Fix B: Retry in the provider datasource**

Add retry-with-backoff in the datasource's `Read` function:
```go
err = retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
    items, _, err := api.ListXxx(ctx)
    if err != nil {
        return retry.NonRetryableError(err)
    }
    if len(items) == 0 {
        return retry.RetryableError(fmt.Errorf("no items found yet"))
    }
    // populate state
    return nil
})
```

---

## Pattern 5: Missing External Service / Cloud Integration (Last Resort)

**Errors:**
- `500 Internal Server Error: Azure integration not configured`
- `404 Not Found: AWS account not integrated`
- `BUCKET_ACCESS: Cloud storage account not found`
- `Invalid MS Teams webhook URL`

**Root cause:** The test requires a cloud provider integration or external service
that isn't available in the current test environment.

**Preferred approach first:** Ask whether the integration can be enabled in the test environment.
If it can, do that instead of skipping.

**Last-resort fix:** Make the test cassette-replay only. This does NOT fix the
underlying test — it only prevents it from surfacing as a live-API failure
while still exercising the provider logic via recorded cassettes.

```go
func TestAccAwsCurConfigBasic(t *testing.T) {
    if !isReplaying() {
        t.Skip("Requires AWS Cost & Usage Report integration — not available in this test environment")
    }
    // cassette replay only below
    ...
}
```

---

## Pattern 6: Rate Limiting

**Errors:**
- `429 Too Many Requests` with `X-Ratelimit-Name` header
- Typically hits endpoints with low per-minute limits

**Root cause:** Multiple tests hitting the same rate-limited endpoint in parallel.

**Fix A: Remove `t.Parallel()` from the affected test family**

Tests for the same resource share API endpoints. Serializing them reduces
concurrent request volume.

**Fix B: Add 429 retry in the provider**

```go
err = retry.RetryContext(ctx, timeout, func() *retry.RetryError {
    _, httpResp, err := client.CreateXxx(ctx, body)
    if err != nil {
        if httpResp != nil && httpResp.StatusCode == 429 {
            return retry.RetryableError(err)
        }
        return retry.NonRetryableError(err)
    }
    return nil
})
```

---

## Pattern 7: API Behavior Change

**Errors:**
- `Attribute 'field' expected "old_value", got "new_value"`
- `Step N: expected an error but got none` (API no longer rejects this)
- `Additional properties are not allowed ('deprecated_field' was unexpected)`

**Root cause:** The Datadog API changed — field renamed, default changed, or validation removed.

**Fix: Update the test to match current API behavior**

For renamed field values:
```go
// Before: API returned "Sessions with replays"
// After:  API returns "Sessions with forced replays"
resource.TestCheckResourceAttr("datadog_rum_application.foo", "name", "Sessions with forced replays"),
```

For removed API fields (e.g., OpsWorks EOL):
```hcl
# Replace removed field with a valid alternative in test HCL configs
services = ["lambda"]  # was "opsworks"
```

For tests that expected an error which the API no longer returns:
Remove or update the `ExpectError` step to reflect the new valid behavior.

Re-record cassettes after updating configs if the API response shape changed.

---

## Pattern 8: State Pollution Between Tests

**Errors:**
- `409 Conflict: A configuration already exists for this tag key`
- `409 Conflict: connection between these orgs already exists`
- Test assumes no pre-existing resource but finds one

**Root cause:** A previous test created a singleton or near-singleton resource and
didn't clean it up.

**Fix A: Use `uniqueEntityName` for resource names**
```go
name := uniqueEntityName(ctx, t)
```

**Fix B: Add `CheckDestroy` to ensure cleanup**
```go
resource.Test(t, resource.TestCase{
    CheckDestroy: testAccCheckDatadogXxxDestroy(accProvider),
    ...
})
```

**Fix C: Add pre-test cleanup for singletons**
```go
PreCheck: func() {
    testAccPreCheck(t)
    cleanupExistingXxx(t)
},
```

---

## Pattern 9: Import State Verify Mismatch

**Errors:**
- `ImportStateVerify failed: created_at: "2026-02-06T12:31:20Z" vs "2026-02-06T12:31:20.671242+00:00"`
- Field value differs between the original state and the re-imported state

**Root cause:** The provider writes an attribute in one format on Create but the API
returns it in a different format on subsequent reads.

**Fix: Normalize in the Read function (preferred)**
```go
if v, ok := resp.GetCreatedAtOk(); ok {
    parsed, err := time.Parse(time.RFC3339Nano, *v)
    if err == nil {
        state.CreatedAt = types.StringValue(parsed.UTC().Format(time.RFC3339))
    }
}
```

**Fix: `ImportStateVerifyIgnore` (last resort)**
```go
resource.TestStep{
    ResourceName:            "datadog_xxx.foo",
    ImportState:             true,
    ImportStateVerify:       true,
    ImportStateVerifyIgnore: []string{"created_at"},
},
```

Prefer fixing the normalization — `ImportStateVerifyIgnore` hides the inconsistency
rather than resolving it.

---

## Reference Files

| Pattern | Example file |
|---------|-------------|
| Sweeper | `datadog/tests/sensitive_data_scanner_sweep_test.go` |
| Dynamic timestamps | `datadog/tests/resource_datadog_downtime_test.go` |
| PreCheck cleanup | `datadog/tests/resource_datadog_logs_custom_destination_test.go` |
| `isReplaying()` helper | `datadog/tests/provider_test.go` |
| Retry in provider | `datadog/fwprovider/resource_datadog_synthetics_test.go` |
