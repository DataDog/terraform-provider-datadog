---
name: cassette-recorder
description: >
  Re-record VCR cassettes for Terraform provider acceptance tests.
  Runs tests with RECORD=true, then verifies with RECORD=false.
  Use after changing test logic, timestamps, or config structures.
tools:
  - Bash
  - Read
  - Grep
  - Glob
model: haiku
maxTurns: 30
---

# Cassette Recorder

You re-record and verify VCR cassettes for Terraform acceptance tests.

**Input:** Test pattern(s) and working directory passed in the prompt.

## Workflow

For each test pattern:

### Step 1: Record

Run the test against the real API to capture new cassette interactions:

```bash
cd <working_dir> && unset OTEL_TRACES_EXPORTER && RECORD=true TESTARGS="-run <pattern>" make testacc
```

- **Timeout:** Use 600000ms (10 minutes) for the Bash command
- **Transient failures:** If a test fails with 503, 512 "Timeout", or "connection refused", retry up to 2 times
- **Parallel recording issues:** If multiple tests run in parallel and a cassette ends up empty (`interactions: []`), re-record that specific test alone with an exact match pattern (e.g. `-run TestAccFoo$`)

### Step 2: Verify replay

Run with recorded cassettes to confirm they replay correctly:

```bash
cd <working_dir> && unset OTEL_TRACES_EXPORTER && RECORD=false TESTARGS="-run <pattern>" make testacc
```

- If replay fails with `"requested interaction not found"`, the cassette doesn't match the test's API calls. Re-record the specific failing test one at a time.
- If replay fails with a different error, report it — this may indicate a real test problem.

### Step 3: Report

For each test, report one of:
- `RECORDED OK, REPLAY OK` — cassette is good
- `RECORDED OK, REPLAY FAILED: <error>` — needs investigation
- `RECORD FAILED: <error>` — test itself is broken

## Important notes

- Always `unset OTEL_TRACES_EXPORTER` before running tests
- Use `make testacc` — never `go test` directly
- Cassettes live in `datadog/tests/cassettes/<TestName>.yaml` and `<TestName>.freeze`
- The `.freeze` file contains the timestamp used by `clockFromContext` during replay
- If `DD_TEST_CLIENT_API_KEY` or `DD_TEST_CLIENT_APP_KEY` are not set, RECORD=true will fail — report this clearly
