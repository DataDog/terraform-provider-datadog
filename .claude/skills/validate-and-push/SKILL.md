---
name: validate-and-push
description: >
  Validate test changes locally, re-record cassettes if needed, review the PR
  diff, then push. Orchestrates test-validator, cassette-recorder, and
  test-pr-reviewer agents to catch issues before CI.
user-invocable: true
argument-hint: "<optional: specific test names to validate, or 'auto' to detect from git diff>"
allowed-tools:
  - Read
  - Bash
  - Edit
  - Agent
  - Glob
  - Grep
  - AskUserQuestion
---

# Validate and Push

**Input:** `$ARGUMENTS`

Determine the repository root:
```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
```

---

## Phase 1 — Validate

Spawn the `test-validator` agent to run affected tests locally:

```
Agent(subagent_type="test-validator", prompt="Working directory: $REPO_ROOT. Validate all changed tests.")
```

Wait for the result. If the validator reports:

- **SAFE TO PUSH** → proceed to Phase 2
- **NEEDS CASSETTE RE-RECORD** → proceed to Phase 1b
- **HAS FAILURES** → report failures to the user and ask how to proceed

---

## Phase 1b — Re-record cassettes (if needed)

Spawn the `cassette-recorder` agent with the tests that need re-recording:

```
Agent(subagent_type="cassette-recorder", prompt="Working directory: $REPO_ROOT. Re-record cassettes for: <test names from validator>")
```

After re-recording, stage the updated cassettes:
```bash
git add datadog/tests/cassettes/
```

---

## Phase 2 — Review

Spawn the `test-pr-reviewer` agent to check for common issues:

```
Agent(subagent_type="test-pr-reviewer", prompt="Working directory: $REPO_ROOT. Review test changes.")
```

Report findings to the user. If there are FAIL-level issues, ask whether to proceed.

---

## Phase 3 — Push

If all checks pass:

1. Show the user what will be pushed:
   ```bash
   git log --oneline origin/master..HEAD
   git diff --stat origin/master..HEAD
   ```

2. Push:
   ```bash
   git push -u origin $(git branch --show-current)
   ```

3. Spawn `ci-monitor` agent in background to track CI:
   ```
   Agent(subagent_type="ci-monitor", run_in_background=true, prompt="Monitor PR for branch $(git branch --show-current)")
   ```

4. Report the PR URL and CI monitoring status to the user.
