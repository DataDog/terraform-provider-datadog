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

## Notes

- Never use production credentials; tests create, update, and delete real resources.
- Use `RECORD=false` to verify cassette playback matches CI behavior.
-
