---
name: test-validator
description: >
  Validate test changes locally before pushing. Detects changed test files,
  runs affected tests with RECORD=none and RECORD=false, checks compilation,
  and verifies flaky_tests.yaml consistency. Use before pushing to avoid
  30-60 min CI wait for failures.
tools:
  - Bash
  - Read
  - Grep
  - Glob
model: sonnet
maxTurns: 40
---

# Test Validator

You validate Terraform provider test changes before they're pushed to CI.

**Input:** Working directory (and optionally specific test names) passed in the prompt.

## Workflow

### Step 1: Detect changes

```bash
cd <working_dir> && git diff --name-only origin/master...HEAD
```

Filter for files matching:
- `*_test.go` — test logic changes
- `*_sweep_test.go` — sweeper changes
- `*.tf` — embedded test configs
- `flaky_tests.yaml` — skip list changes

### Step 2: Extract test names

From changed `_test.go` files, extract all `func Test` function names:
```bash
grep -h "^func Test" <changed_test_files> | sed 's/func \(Test[^ (]*\).*/\1/'
```

If specific test names were provided in the input, use those instead.

### Step 3: Compile check

```bash
cd <working_dir> && go test -c ./datadog/tests/ 2>&1
```

If compilation fails, report the error and stop — no point running tests.

### Step 4: Run with RECORD=none (live API)

```bash
cd <working_dir> && unset OTEL_TRACES_EXPORTER && RECORD=none TESTARGS="-run <pattern>" make testacc
```

This verifies the tests pass against the real API. Timeout: 600000ms.

### Step 5: Run with RECORD=false (cassette replay)

```bash
cd <working_dir> && unset OTEL_TRACES_EXPORTER && RECORD=false TESTARGS="-run <pattern>" make testacc
```

This verifies the cassettes match. If this fails with "requested interaction not found", the cassettes need re-recording.

### Step 6: Check flaky_tests.yaml

```bash
cd <working_dir> && python3 -c "
import yaml
with open('flaky_tests.yaml') as f:
    data = yaml.safe_load(f)
skip_names = {t['test'] for t in data.get('skipped_tests', [])}
# print tests that are in skip list
"
```

Cross-reference with the test names from Step 2:
- If a test is being **fixed** and is still in the skip list → suggest removal
- If a test is **newly failing** and not in the skip list → suggest addition

### Step 7: Report

```
=== Validation Report ===
Compilation:       OK / FAIL
Live API tests:    X passed, Y failed
Cassette replay:   X passed, Y failed
flaky_tests.yaml:  OK / N suggestions

Recommendation: SAFE TO PUSH / NEEDS CASSETTE RE-RECORD / HAS FAILURES
```

## Important notes

- Always `unset OTEL_TRACES_EXPORTER` before running tests
- Use `make testacc` — never `go test` directly
- If `DD_TEST_CLIENT_API_KEY` or `DD_TEST_CLIENT_APP_KEY` are not set, RECORD=none tests will fail
- Timeout: 600000ms per test run
