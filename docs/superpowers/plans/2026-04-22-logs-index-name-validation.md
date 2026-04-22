# Logs Index Name Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add schema-level validation to `datadog_logs_index.name` so invalid names (e.g. containing underscores) fail at `tofu plan` time instead of returning a 400 from the Datadog API at apply time.

**Architecture:** Add a single `ValidateFunc` to the `name` field in `indexSchema` using the existing `validation.StringMatch` helper. Add a unit test (no API calls needed) that asserts an invalid name produces the expected validation error.

**Tech Stack:** Go, `github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema`, `github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation`, `github.com/hashicorp/terraform-plugin-testing/helper/resource`

---

## File Map

| File | Change |
|------|--------|
| `datadog/resource_datadog_logs_index.go` | Add `ValidateFunc` to `name` field in `indexSchema` |
| `datadog/tests/resource_datadog_logs_index_test.go` | Add `TestUnitDatadogLogsIndex_InvalidName` test function + config helper |

---

### Task 1: Write the failing test

**Files:**
- Modify: `datadog/tests/resource_datadog_logs_index_test.go`

- [ ] **Step 1: Add the test function and config helper**

Append to the end of `datadog/tests/resource_datadog_logs_index_test.go`:

```go
func TestUnitDatadogLogsIndex_InvalidName(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogLogsIndexInvalidNameConfig("mbff-user_trips"),
				ExpectError: regexp.MustCompile("must start with a lowercase letter and contain only lowercase letters, digits, or hyphens"),
			},
		},
	})
}

func testAccCheckDatadogLogsIndexInvalidNameConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "invalid_index" {
  name = "%s"
  filter {
    query = "service:test"
  }
}
`, name)
}
```

- [ ] **Step 2: Add `regexp` to imports**

The `regexp` package is not yet imported in the test file. Update the import block at the top of `datadog/tests/resource_datadog_logs_index_test.go`:

```go
import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)
```

- [ ] **Step 3: Run the test to verify it fails**

```bash
cd /Users/eligio.marino/Code/terraform-provider-datadog
TF_ACC=1 go test ./datadog/tests/ -run TestUnitDatadogLogsIndex_InvalidName -v
```

Expected: test **fails** because no `ExpectError` match is found — the provider currently accepts any string for `name` and no validation error is produced.

---

### Task 2: Add the ValidateFunc

**Files:**
- Modify: `datadog/resource_datadog_logs_index.go`

- [ ] **Step 1: Add ValidateFunc to the `name` field in `indexSchema`**

In `datadog/resource_datadog_logs_index.go`, find the `name` entry in `indexSchema` (around line 20) and add `ValidateFunc`:

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

Both `regexp` and `validation` are already imported — no import changes needed.

- [ ] **Step 2: Run the test to verify it passes**

```bash
cd /Users/eligio.marino/Code/terraform-provider-datadog
TF_ACC=1 go test ./datadog/tests/ -run TestUnitDatadogLogsIndex_InvalidName -v
```

Expected output (abbreviated):
```
--- PASS: TestUnitDatadogLogsIndex_InvalidName (...)
PASS
```

- [ ] **Step 3: Confirm the existing Basic test still compiles**

```bash
cd /Users/eligio.marino/Code/terraform-provider-datadog
go build ./datadog/...
```

Expected: exits 0 with no output.

---

### Task 3: Commit

- [ ] **Step 1: Stage and commit**

```bash
cd /Users/eligio.marino/Code/terraform-provider-datadog
git add datadog/resource_datadog_logs_index.go datadog/tests/resource_datadog_logs_index_test.go
git commit -m "feat(logs-index): validate name at plan time

Adds ValidateFunc to the name field so names containing characters
outside [a-z0-9-] (e.g. underscores) fail during tofu plan rather
than returning a 400 from the Datadog API at apply time.

Regex: ^[a-z][a-z0-9-]*$  (matches Datadog API constraint)

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```
