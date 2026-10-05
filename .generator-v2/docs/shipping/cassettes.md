---
ddoc:
  confluence_space: "API"
  confluence_parent: "7290816063"
  confluence_id: "7291503804"
---

# Recording and replaying the cassette

**Who this is for:** whoever is finishing a generated pull request, once the
acceptance test is written.

Acceptance tests run against recorded HTTP interactions so CI does not need live
credentials. Recording is a live run; replay is not.

## RECORD modes

| `RECORD` | Behaviour |
|---|---|
| `true` | Call the live API and write or replace the cassette |
| `false` | Replay from the committed cassette, no API calls — **the default**, and what PR CI uses |
| `none` | Call the live API and record nothing |

## 1. Get test credentials

Use the **Frog** test organisation. Never record against a production organisation.

```sh
eval "$(dd-auth --domain frog.datadoghq.com --force-app-key --no-cache --output)"
export DD_TEST_CLIENT_API_KEY="$DD_API_KEY"
export DD_TEST_CLIENT_APP_KEY="$DD_APP_KEY"
```

## 2. Record

Run only your test:

```sh
RECORD=true TESTARGS="-run TestAccDatadogIncidentTypeResource" make testacc
```

That writes two files next to each other:

```
datadog/tests/cassettes/<TestName>.yaml     the recorded HTTP interactions
datadog/tests/cassettes/<TestName>.freeze   the timestamp used while recording
```

The `.freeze` file is what makes time-derived values deterministic on replay. Commit
both.

## 3. Read the recording before committing it

This step is not optional, and it is the one people skip. The harness strips API and
application keys from request URLs and filters headers, but **response bodies are
recorded verbatim**.

Confirm that:

- the requests target the endpoint you expect
- filters and request bodies match the test configuration
- the responses actually contain the fields you assert on
- setup and cleanup requests were recorded
- no credentials, customer data, or other sensitive material appears anywhere

## 4. Replay

```sh
RECORD=false TESTARGS="-run TestAccDatadogIncidentTypeResource" make testacc
```

Replay matches each request against the cassette by HTTP method and URL — **including
the query string** — and request bodies must match too, either exactly or as
equivalent JSON. Interactions are single-use and consumed in order.

This is why deterministic test configuration matters: a different filter value
produces a different URL, which matches nothing, and the test fails with a
confusing "no interaction found" rather than an assertion failure.

## 5. Validate live, without touching the cassette

```sh
RECORD=none TESTARGS="-run TestAccDatadogIncidentTypeResource" make testacc
```

Recording with `RECORD=true` was also a live run, but this confirms the **committed**
test works independently of the act of creating the cassette. Do it last, before
taking the pull request out of draft.

## What a cassette does and does not cover

A cassette covers exactly the interactions recorded during that run. A green replay
says nothing about request or response shapes absent from it, and nothing about
whether the API has changed since. That residual risk is why live validation stays
part of merging rather than being replaced by replay.
