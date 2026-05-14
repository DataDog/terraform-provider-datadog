---
name: fix-broken-integration-tests
description: >
  End-to-end workflow to diagnose, reproduce, fix, and validate a failing
  Datadog Terraform provider integration test. Takes any input pointing at
  specific tests: test function names, error messages, resource names, or any
  other description of what's failing. Runs autonomously through 8 phases —
  identify → validate in CI → reproduce locally → plan fix → execute → open
  draft PR → monitor → report.
user-invocable: true
argument-hint: "<test name(s), error description, or other context about the failing test>"
allowed-tools:
  - Read
  - Write
  - Edit
  - Glob
  - Grep
  - Bash
  - Agent
  - Skill
  - TaskCreate
  - TaskUpdate
  - TaskGet
  - AskUserQuestion
---

# Fix Broken Integration Tests

**Input:** `$ARGUMENTS`

Determine the repository root at the start of each phase:
```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
```
All relative file paths below are relative to `$REPO_ROOT`.

---

## Phase 1 — Identify Failing Tests

Parse `$ARGUMENTS` to extract concrete Go test function names.

**If input matches `[A-Z]+-\d+` (looks like a ticket ID):**
```bash
acli jira workitem view <TICKET-ID> --fields 'summary,status,description' 2>&1
```
Extract test names mentioned in the description. If the ticket is already
`Done`, warn the user: "Ticket is marked Done — tests may already be fixed.
Proceeding to validate in CI."

**If input contains Go test function names** (starts with `TestAcc` or `Test`):
Use them directly. A comma- or space-separated list is fine.

**If input is a free-form description** (resource name, error message, etc.):
Use the Grep tool to search the test directory:
```
Grep pattern "<keyword>" path "datadog/tests/" glob "*_test.go"
```
Then identify which test functions match the description. Read the test file
to confirm the relevant test names.

**Also check `flaky_tests.yaml`** to see if the tests are already tracked:
```bash
grep -A4 "<TestName>" flaky_tests.yaml
```
Note any existing reason/context — it informs the fix strategy.

Create a task to track progress:
```
TaskCreate: "Fix integration tests: <list of tests>"
```

---

## Phase 2 — Validate in Recent CI

Confirm the tests are still failing on master before doing any work.

```bash
gh run list --workflow test_integration.yml \
  --repo DataDog/terraform-provider-datadog \
  --branch master --limit 5 --json databaseId,conclusion,createdAt
```

For each completed run (check the 3 most recent):
```bash
gh run view <RUN_ID> --log-failed \
  --repo DataDog/terraform-provider-datadog 2>&1 \
  | grep -E "FAIL.*<TestName>" | head -20
```

**Decision point:**
- Tests failing in 2+ of the last 3 runs → proceed
- Tests not failing in any recent run → ask the user:
  > "These tests did not fail in the last 3 CI runs. They may have been
  > fixed already, or the failure is intermittent. How do you want to proceed?"
  > Options: "Reproduce locally anyway", "Check more CI runs", "Abort"
- Tests failing inconsistently → note intermittency, flag in PR description

Record the exact error message from CI — you'll compare it against the local
reproduction and the final CI result.

---

## Phase 3 — Reproduce Locally

Run the failing tests against the real API to confirm the current failure mode.

```
Skill: "dd-tf-provider-test-runner"
Args: "Test pattern: <TestName1>|<TestName2>  Record mode: none  Working directory: <value of REPO_ROOT>"
```

**If local reproduction matches CI error:** proceed to Phase 4.

**If local test passes:** warn the user —
> "Tests pass locally with RECORD=none. The failure may be environment-specific
> (quota, org state) or intermittent. Recommend running a few more times or
> checking org state before proceeding."

Capture the exact local error output for comparison.

---

## Phase 4 — Diagnose and Plan Fix

Read the failure patterns reference:
```
Read: .claude/skills/fix-broken-integration-tests/fix-patterns.md
```

Match the error message against the patterns to identify the fix type. Read the
relevant source files to understand the current code:

- Test file: `datadog/tests/<resource>_test.go`
- Resource file (if provider bug): `datadog/fwprovider/<resource>.go` or `datadog/<resource>.go`
- Sweep file (if quota/accumulation): `datadog/tests/<resource>_sweep_test.go` (may need creating)

Draft a concrete fix plan covering:
1. Which files change and what the change is
2. Whether cassettes need re-recording (`RECORD=true`)
3. Whether the test should be removed from `flaky_tests.yaml`

**Ask the user to confirm before making any changes:**
> "Here is my proposed fix for `<TestName(s)>`:
>   - Fix type: <e.g., sweeper, dynamic timestamps, provider read bug>
>   - Files affected: <list>
>   - Cassette re-recording needed: yes/no
>   - Summary: <one sentence>
>   Proceed?"
>
> Options: "Yes, execute the fix", "Modify the plan first", "Abort"

---

## Phase 5 — Execute the Fix

### 5a. Create a branch

```bash
git checkout -b fix/<resource>-integration-test
```

### 5b. Apply code changes

Follow the appropriate pattern from `fix-patterns.md`:

- **Timestamp fix:** replace hardcoded Unix timestamps with `clockFromContext(ctx).Now().Local().Add(...)`; update config function signatures to accept `start, end int64`
- **Sweeper fix:** create `datadog/tests/<resource>_sweep_test.go`; add `cleanupXxx(t)` call at top of each failing test function; add `TestSweepXxx` standalone function
- **Test assertion fix:** update the assertion to match new API behavior
- **Provider read bug:** fix the `Read` function; add attribute normalization or `DiffSuppressFunc`
- **Skip in live API mode:** last resort only — use when the test requires an external service that genuinely cannot be configured in the current test environment; add `if !isReplaying() { t.Skip(...) }`

### 5c. Re-record cassettes if needed

If the fix changes what the API interaction looks like:
```
Skill: "dd-tf-provider-test-runner"
Args: "Test pattern: <TestName>  Record mode: true  Working directory: <REPO_ROOT>"
```

### 5d. Validate locally via cassette replay

```
Skill: "dd-tf-provider-test-runner"
Args: "Test pattern: <TestName>  Record mode: false  Working directory: <REPO_ROOT>"
```

If cassette replay fails, investigate and re-record.

### 5e. Remove from flaky_tests.yaml

If the test is in `flaky_tests.yaml`, remove its entry.

### 5f. Quick quality checks

```bash
make fmtcheck
make test
```

---

## Phase 6 — Commit and Open Draft PR

### 6a. Stage and commit

```bash
git add <changed files>
git commit -m "[datadog_<resource>] Fix integration test — <brief root cause>

<one paragraph explaining what was failing and why, and how it is fixed>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

### 6b. Push branch

```bash
git push -u origin <branch-name>
```

### 6c. Create draft PR

Use `changelog/bugfix` only if provider code changed; use `changelog/no-changelog`
for test-only or sweeper-only changes.

```bash
gh pr create --draft \
  --title "[datadog_<resource>] Fix integration test — <brief description>" \
  --label "ci/integrations" \
  --label "<changelog/bugfix or changelog/no-changelog>" \
  --body "$(cat <<'EOF'
## Summary

- **Failing tests:** `<TestName1>`, `<TestName2>`
- **Root cause:** <one-line root cause>
- **Fix:** <one-line fix description>

## Details

<paragraph about what was failing and the error observed in CI>

<paragraph about the fix approach and what changed>

## Test plan

- [ ] Tests pass with `RECORD=none` locally
- [ ] Tests pass in CI integration run (triggered by `ci/integrations` label)
EOF
)"
```

Report the PR URL to the user.

---

## Phase 7 — Monitor Integration Tests

The `ci/integrations` label triggers `.github/workflows/test_integration.yml`.
This run typically takes **35–45 minutes**.

Poll every 5 minutes for the run to appear and complete:

```bash
# Wait for the run to be created (retry up to 10 minutes)
gh run list --repo DataDog/terraform-provider-datadog \
  --branch <branch-name> --workflow test_integration.yml \
  --limit 3 --json databaseId,status,conclusion,createdAt

# Check run status
gh run view <RUN_ID> --repo DataDog/terraform-provider-datadog
```

Continue polling until `status == "completed"`.

**Timeout:** If the run has not completed after 90 minutes, stop polling and
report the current status. Advise the user to check manually.

---

## Phase 8 — Report Results

### If all target tests pass:
- Report success with test counts
- Suggest removing the PR from draft and requesting review
- If a ticket ID was provided as input, note it can be closed

### If tests still fail:
1. Extract the failure details:
```bash
gh run view <RUN_ID> --log-failed --repo DataDog/terraform-provider-datadog 2>&1 \
  | grep -E "FAIL.*<TestName>|Error:" | head -40
```
2. Compare with the original error from Phase 3
3. **Same error:** the fix didn't work — diagnose why and propose a revised fix
4. **Different error:** the original issue is fixed but uncovered a second problem —
   treat as a new cycle starting at Phase 4
5. Report findings with a clear next-steps recommendation

---

## Reference

- Failure pattern lookup: `.claude/skills/fix-broken-integration-tests/fix-patterns.md`
- Test infrastructure: `TESTING.md`, `AGENTS.md`
- Sweep examples: `datadog/tests/sweep_test.go`, `datadog/tests/sensitive_data_scanner_sweep_test.go`
- Cassette management: `datadog/tests/provider_test.go`
