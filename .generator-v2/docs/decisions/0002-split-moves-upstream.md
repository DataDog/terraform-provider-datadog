---
ddoc:
  confluence_space: "API"
  confluence_parent: "7291274458"
  confluence_id: "7291732553"
---

# 0002: The split/fan-out workflow moved upstream

- **Date:** 2026-10-02 (recording #4262, merged 2026-09-23)
- **Status:** accepted; the upstream description is an open gap

## Context

Generating a batch of artifacts produces one aggregate branch. Something has to
attribute each changed file to an artifact and open one draft pull request per
artifact. That was `.github/workflows/tfgen-split.yml` in this repository: 746
lines of workflow that ran on pushes to `datadog-api-spec/generated/**`.

Hosting it here meant the provider repository carried CI whose purpose was to
produce provider pull requests, and that workflow appeared in the repository as
though it were part of the provider's own build.

## Decision

The workflow was deleted in #4262, with the rationale: *"This job should be part of
an upstream pipeline, not part of the final PRs."*

The `tfgen split` command and `internal/split/` stay here. The generator still
computes the attribution; the provider repository no longer runs the fan-out.

## Consequences

- [pipeline.md](../pipeline.md) describes triggers and ownership only, and links out
  rather than re-documenting machinery this repository does not own.
- **Open gap:** at the time of writing, the authoritative description of where the
  fan-out now runs has not been written in the owning repository. Anyone debugging
  fan-out starts from the API-spec pipeline configuration. This should be closed by
  documentation upstream, not by re-adding a description here.
- Two artefacts of the old workflow are still checked in and should be removed: the
  `self.generator.tfgen-split.sts.yaml` STS policy, which still grants
  `contents: write` and `pull_requests: write` to a workflow ref that no longer
  exists, and a stale comment in `.github/workflows/test.yml`.

## Why this is recorded

The previous documentation described the removed workflow in detail (branch naming,
the pull-request cap, the fail-slow loop) for weeks after it was deleted, because
nothing connected the deletion to the page describing it. A dated decision record is
the connection.
