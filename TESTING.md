# Testing

This project uses Makefile targets for all test runs. Do not run raw `go test` directly.

## Required Environment Variables

Acceptance tests require sandbox credentials:

- `DD_TEST_CLIENT_API_KEY`
- `DD_TEST_CLIENT_APP_KEY`

Optional:

- `DD_TEST_SITE_URL` to override the API site for tests.

## Make Targets Overview

Use these Makefile targets to run tests:

```bash
make test     # Unit tests only
make testacc  # Acceptance tests only
make testall  # Unit + acceptance tests
```

## Unit Tests

Run unit tests with:

```bash
make test
```

## Acceptance Tests

Run acceptance tests with:

```bash
make testacc
```

### RECORD Modes

Acceptance tests use cassettes stored under `datadog/tests/cassettes/`. The `RECORD` env var controls whether tests hit the live API:

- `RECORD=false`: Replay from cassettes (default, no API calls). Use in CI and to verify cassettes.
- `RECORD=true`: Record new cassettes (hits real API). Use after fixing tests.
- `RECORD=none`: Live API only (no recording). Use for debugging.

### Running a Single Acceptance Test

```bash
RECORD=none \
  DD_TEST_CLIENT_API_KEY=$DD_TEST_CLIENT_API_KEY \
  DD_TEST_CLIENT_APP_KEY=$DD_TEST_CLIENT_APP_KEY \
  TESTARGS="-run TestAccDatadogMonitor_Basic" \
  make testacc
```

### Extra Test Arguments

Use `TESTARGS` to pass flags to the underlying test runner. Example:

```bash
TESTARGS="-run TestAccDatadogServiceLevelObjective_Basic" make testacc
```

## Recording Cassettes via CI

Maintainers can record or update cassettes directly from a PR by posting the following comment:

```
/cassettes record
```

This triggers a GitHub Actions workflow that:

1. Validates that the commenter is a maintainer (`OWNER` or `MEMBER`).
2. Detects which acceptance tests are affected by the changed files in the PR (using the same test-selection logic as the integration test workflow).
3. Mints short-lived Datadog sandbox credentials via `dd-sts-action`.
4. Runs the affected tests with `RECORD=true` to record or update cassettes.
5. Pushes the recorded cassettes back to the PR branch (same-repo PRs) or uploads them as a workflow artifact (fork PRs).

### Fork PRs

For PRs opened from a fork, the workflow cannot push directly to the fork's branch. Instead, the cassettes are uploaded as a workflow artifact. A comment on the PR will include a link to download the artifact. Extract the files into `datadog/tests/cassettes/` and commit them to the PR branch.

### Limitations

- Only maintainers can trigger the command.
- Only tests affected by the changed files in the PR are recorded. To record all tests, run `make cassettes` locally with sandbox credentials.
- The workflow cancels any in-progress recording run for the same PR if a new `/cassettes record` comment is posted.

## Notes

- Never use production credentials; tests create, update, and delete real resources.
- Use `RECORD=false` to verify cassette playback matches CI behavior.
-
