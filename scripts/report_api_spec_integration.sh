#!/usr/bin/env bash
# Match the SDK integration status contract for a workflow_run event.
# Run from the trusted base checkout with an api-spec-scoped GH_TOKEN.
set -euo pipefail

if [[ ! "${HEAD_BRANCH}" =~ ^datadog-api-spec/generated/([1-9][0-9]*)$ ]]; then
  echo "Expected a generated branch containing the source spec PR number" >&2
  exit 1
fi
spec_pr="${BASH_REMATCH[1]}"

case "${INTEGRATION_CONCLUSION}" in
  success) state=success ;;
  pending|cancelled) state=pending ;;
  failure) state=failure ;;
  *) echo "Unexpected integration conclusion: ${INTEGRATION_CONCLUSION}" >&2; exit 1 ;;
esac

# Like the SDK action, report against the source PR's current head.
spec_sha=$(gh api "repos/ddoghq/datadog-api-spec/pulls/${spec_pr}" --jq .head.sha)
gh api --method POST "repos/ddoghq/datadog-api-spec/statuses/${spec_sha}" \
  -f "state=${state}" \
  -f "context=${GITHUB_REPOSITORY#*/}/integration" \
  -f "description=${GITHUB_WORKFLOW}" \
  -f "target_url=${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
