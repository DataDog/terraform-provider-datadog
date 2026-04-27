#!/usr/bin/env bash
#
# Self-contained walkthrough of the SAT ephemeral POC against staging.
#
# Each stage of the demo (re)writes the same `main.tf`, prints it, then runs
# `terraform apply` on it. Reading top-to-bottom shows exactly what changes
# between applies — the only thing the demo varies is `rotation_trigger`.
#
# Required tooling: terraform >=1.11, jq, go.
# Required env:     DDPAT (a Datadog Personal Access Token, format ddpat_...)
# Usage:            DDPAT=ddpat_xxx ./run_demo.sh

set -euo pipefail

# ─── required env ───────────────────────────────────────────────────────────

if [[ -z "${DDPAT:-}" ]]; then
  printf "\033[1;31mERROR:\033[0m DDPAT env var is required. Set it to a Datadog Personal Access Token (ddpat_...).\n" >&2
  printf "Example:\n  DDPAT=ddpat_xxx %s\n" "${0##*/}" >&2
  exit 1
fi

if [[ ! "${DDPAT}" == ddpat_* ]]; then
  printf "\033[1;31mERROR:\033[0m DDPAT does not look like a PAT (should start with 'ddpat_').\n" >&2
  exit 1
fi

# ─── paths and constants ────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WORK_DIR="${SCRIPT_DIR}/.work"
TF_FILE="${WORK_DIR}/main.tf"
PROVIDER_BIN_DIR="${WORK_DIR}/provider"
DEV_TFRC="${WORK_DIR}/dev.tfrc"

# Staging targets app.datad0g.com; the API is on api.datad0g.com.
DD_HOST_URL="${DD_HOST_URL:-https://api.datad0g.com}"
DD_HOST_DOMAIN="${DD_HOST_URL#https://}"

# The Datadog Terraform provider only sends DD-API-KEY/DD-APPLICATION-KEY
# headers (no native Bearer support), so we put the PAT in DD_APP_KEY.
# Datadog's auth backend recognizes the ddpat_ prefix and validates accordingly.
export DD_APP_KEY="${DDPAT}"
export DD_HOST="${DD_HOST_URL}"

# ─── pretty logging ─────────────────────────────────────────────────────────

log()      { printf "\n\033[1;36m▶ %s\033[0m\n" "$*"; }
log_step() { printf "\n\033[1;33m═══════════════════════════════════════════════════════════════════\n  %s\n═══════════════════════════════════════════════════════════════════\033[0m\n" "$*"; }

# ─── helpers ────────────────────────────────────────────────────────────────

# Writes the demo HCL with the supplied rotation_trigger value and prints the
# file to stdout, so each apply is preceded by the exact HCL Terraform will see.
write_and_show_tf() {
  local rotation_trigger="$1"
  # Quoted heredoc so bash does NOT interpret $... or backticks. Use a __ROTATION__
  # placeholder for the only value that varies per stage, and substitute it after.
  cat > "${TF_FILE}" <<'EOF'
terraform {
  required_providers {
    datadog = { source = "DataDog/datadog" }
  }
}
provider "datadog" {
  # PATs aren't accepted by the standard /api/v1/validate credential check, so disable it.
  validate = "false"
}

# Service account that owns the SAT.
resource "datadog_service_account" "demo" {
  name  = "tf-sat-poc"
  email = "tf-sat-poc-${formatdate("YYYYMMDDhhmm", timestamp())}@example.com"
  lifecycle {
    ignore_changes = [email]
  }
}

# Anchor: random ID generated once on first apply, stable forever.
# anchor.id does NOT change on rotation.
resource "terraform_data" "anchor" {}

# Locals so the ephemeral and the diagnostic outputs share the same source values.
# (Ephemeral attributes can't flow into non-ephemeral outputs, so we express the
# constructed name in terms of locals here.)
locals {
  rotation_trigger = "__ROTATION__"
  anchor_short     = substr(terraform_data.anchor.id, 0, 8)
  constructed_name = "tf-poc-${local.anchor_short}-${local.rotation_trigger}"
}

# Ephemeral SAT — creates/finds the SAT during apply; secret value never lands in tfstate.
# Server-side name is constructed by the provider as: tf-poc-<anchor.id[:8]>-<rotation_trigger>.
ephemeral "datadog_service_account_token" "ci" {
  service_account_id = datadog_service_account.demo.id
  prefix             = "tf-poc"
  anchor             = local.anchor_short
  rotation_trigger   = local.rotation_trigger
  scopes             = ["dashboards_read"]
}

# Lifecycle — tracks the active SAT UUID and any previous ones for cutover/cleanup.
# key_id_version is the non-ephemeral version trigger. When it changes (e.g. rotation_trigger
# bumps), the framework triggers Update which then reads the fresh write-only key_id.
resource "datadog_service_account_token_lifecycle" "ci" {
  service_account_id = datadog_service_account.demo.id
  key_id             = ephemeral.datadog_service_account_token.ci.id
  key_id_version     = local.constructed_name
  retain_count       = 1
}

# Outputs make state easy to inspect between applies.
output "service_account_id"   { value = datadog_service_account.demo.id }
output "anchor_id"            { value = terraform_data.anchor.id }
output "constructed_sat_name" { value = local.constructed_name }
output "active_sat_uuid"      { value = datadog_service_account_token_lifecycle.ci.active_key_id }
output "previous_sat_uuids"   { value = datadog_service_account_token_lifecycle.ci.previous_keys }
EOF
  # Substitute the rotation_trigger placeholder.
  sed -i.bak "s/__ROTATION__/${rotation_trigger}/g" "${TF_FILE}" && rm "${TF_FILE}.bak"

  log "About to apply (rotation_trigger=\"${rotation_trigger}\"):"
  printf "\033[2m"
  cat "${TF_FILE}"
  printf "\033[0m\n"
}

# Runs `terraform <args>` from WORK_DIR with the dev_overrides config picked up.
run_tf() {
  TF_CLI_CONFIG_FILE="${DEV_TFRC}" \
    bash -c "cd '${WORK_DIR}' && terraform $*"
}

# Hits the SAT API directly with the PAT in a Bearer header.
curl_api() {
  local method="$1"
  local path="$2"
  curl -sS -X "${method}" "${DD_HOST_URL}${path}" \
    -H "Accept: application/json" \
    -H "Authorization: Bearer ${DDPAT}"
}

# Track every SAT UUID we mint across the demo so we can assert all of them
# are revoked at the end.
MINTED_SATS=()

# Captures the currently-active SAT UUID from terraform outputs and appends it
# to MINTED_SATS (deduplicated). Safe to call after no-op applies.
record_active_sat() {
  local uuid
  uuid=$(run_tf output -raw active_sat_uuid 2>/dev/null | tail -n 1 || echo "")
  [[ -z "${uuid}" ]] && return
  for existing in "${MINTED_SATS[@]+"${MINTED_SATS[@]}"}"; do
    [[ "${existing}" == "${uuid}" ]] && return
  done
  MINTED_SATS+=("${uuid}")
  log "Tracking SAT ${uuid} for end-of-demo revocation check."
}

# ─── stage 0: bootstrap ─────────────────────────────────────────────────────

log_step "Stage 0 — bootstrap (build provider, configure dev_overrides, reset state)"

mkdir -p "${WORK_DIR}" "${PROVIDER_BIN_DIR}"

log "Building local provider binary"
( cd "${REPO_ROOT}" && go build -o "${PROVIDER_BIN_DIR}/terraform-provider-datadog" . )

log "Writing dev_overrides config to ${DEV_TFRC}"
cat > "${DEV_TFRC}" <<EOF
provider_installation {
  dev_overrides { "DataDog/datadog" = "${PROVIDER_BIN_DIR}" }
  direct {}
}
EOF

log "Resetting prior Terraform state"
rm -f "${WORK_DIR}/terraform.tfstate" "${WORK_DIR}/terraform.tfstate.backup" "${WORK_DIR}/tfplan.out"
rm -rf "${WORK_DIR}/.terraform" "${WORK_DIR}/.terraform.lock.hcl"

log "Auth target: ${DD_HOST_URL}  (PAT supplied via DD_APP_KEY → DD-APPLICATION-KEY header)"

# ─── stage 1: first apply ───────────────────────────────────────────────────

log_step "Stage 1 — first apply (rotation_trigger=\"1\")"
log "Expect: service account created, anchor generated, SAT-A minted, lifecycle stores SAT-A."

write_and_show_tf "1"
run_tf apply -auto-approve
record_active_sat

log "State after stage 1:"
run_tf output

# ─── stage 2: steady-state apply ────────────────────────────────────────────

log_step "Stage 2 — steady-state apply (no HCL changes)"
log "Expect: zero changes; same anchor; same active SAT."

write_and_show_tf "1"
run_tf plan -out=tfplan.out
run_tf apply -auto-approve tfplan.out
record_active_sat

# ─── stage 3: rotation (1 → 2) ──────────────────────────────────────────────

log_step "Stage 3 — rotation (rotation_trigger=\"1\" → \"2\")"
log "Expect: anchor unchanged, new SAT-B minted, lifecycle moves SAT-A → previous_keys."
log "        With retain_count=1, SAT-A stays alive for graceful cutover until next rotation."

write_and_show_tf "2"
run_tf apply -auto-approve
record_active_sat

log "State after stage 3:"
run_tf output
log "Lifecycle state details:"
run_tf state show datadog_service_account_token_lifecycle.ci

# ─── stage 4: rotation (2 → 3) ──────────────────────────────────────────────

log_step "Stage 4 — rotation (rotation_trigger=\"2\" → \"3\")"
log "Expect: SAT-C minted, SAT-B moves to previous_keys, retain_count=1 prunes and revokes SAT-A."

write_and_show_tf "3"
run_tf apply -auto-approve
record_active_sat

log "State after stage 4:"
run_tf output

# ─── stage 5: API-side verification ─────────────────────────────────────────

log_step "Stage 5 — API-side verification of remaining SATs"

SA_ID=$(run_tf output -raw service_account_id 2>/dev/null | tail -n 1 || echo "")
if [[ -n "${SA_ID}" ]]; then
  log "Listing SATs on service account ${SA_ID}:"
  curl_api GET "/api/v2/service_accounts/${SA_ID}/access_tokens" \
    | jq '.data[] | {id: .id, name: .attributes.name, revoker_uuid: .attributes.revoker_uuid}'
fi

# ─── stage 6: destroy ───────────────────────────────────────────────────────

log_step "Stage 6 — destroy"
log "Expect: lifecycle revokes active + previous SATs; service account torn down."

run_tf destroy -auto-approve

# ─── final verification ─────────────────────────────────────────────────────
#
# Assert that every SAT we minted during the demo is unreachable. The check has
# two passing paths:
#   1. The parent service account returns 404 — by definition, none of its SATs
#      can be reached anymore. This is the typical outcome of a clean destroy.
#   2. The service account still exists for some reason — fall back to listing
#      and asserting each tracked UUID has revoker_uuid populated.
log_step "Final verification — every minted SAT must be revoked or unreachable"

if [[ ${#MINTED_SATS[@]} -eq 0 ]]; then
  printf "\033[1;31mERROR:\033[0m no SATs were tracked during the demo; verification cannot run.\n" >&2
  exit 1
fi

log "Tracked ${#MINTED_SATS[@]} SAT UUID(s) across the demo:"
for uuid in "${MINTED_SATS[@]}"; do
  printf "    - %s\n" "${uuid}"
done

if [[ -z "${SA_ID:-}" ]]; then
  printf "\033[1;31mERROR:\033[0m SA_ID was never captured; cannot verify revocation.\n" >&2
  exit 1
fi

sa_status=$(curl -sS -o /dev/null -w "%{http_code}" \
  -X GET "${DD_HOST_URL}/api/v2/service_accounts/${SA_ID}" \
  -H "Authorization: Bearer ${DDPAT}")

if [[ "${sa_status}" == "404" ]]; then
  log "PASS: parent service account returns 404 — every tracked SAT is unreachable."
else
  log "Service account still resolves (HTTP ${sa_status}); checking each SAT individually."
  list_json=$(curl_api GET "/api/v2/service_accounts/${SA_ID}/access_tokens")
  fail=0
  for uuid in "${MINTED_SATS[@]}"; do
    match=$(echo "${list_json}" | jq -r --arg id "${uuid}" '.data[] | select(.id==$id) | .attributes.revoker_uuid // "PRESENT_BUT_NULL"')
    if [[ -z "${match}" ]]; then
      printf "    OK: SAT %s absent from listing\n" "${uuid}"
    elif [[ "${match}" == "PRESENT_BUT_NULL" ]]; then
      printf "    \033[1;31mFAIL\033[0m: SAT %s still active (revoker_uuid is null)\n" "${uuid}"
      fail=1
    else
      printf "    OK: SAT %s revoked (revoker_uuid=%s)\n" "${uuid}" "${match}"
    fi
  done
  if [[ ${fail} -ne 0 ]]; then
    printf "\033[1;31mERROR:\033[0m one or more SATs are not revoked.\n" >&2
    exit 1
  fi
  log "PASS: all ${#MINTED_SATS[@]} SAT(s) verified revoked."
fi

log_step "DONE"
