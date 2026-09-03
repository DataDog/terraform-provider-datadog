#!/usr/bin/env python3
"""Build mini-OAS slices for the two integration-account interfaces (Twilio,
Elastic Cloud) that live under datadog-api-spec's spec/v2/integration_accounts/.

Unlike the other mini-oas slices, this feature hasn't been merged into the
bundled full v2 spec yet (DATADOG_OPENAPI_V2_SPEC), so there is no single file
to slice from. Instead this script assembles an in-memory "spec" by merging the
handful of raw spec-repo fragments that make up the feature -- header.yaml and
shared.yaml (global info/servers/security/shared error responses) plus the
integration_accounts fragments themselves -- and then reuses _build_mini's
generic build_slice() against that merged spec.

Point at the datadog-api-spec checkout with DATADOG_API_SPEC_REPO (defaults to
~/go/src/github.com/DataDog/datadog-api-spec).
"""
import os
import sys

import yaml

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _build_mini import build_slice, HTTP_METHODS, Loader, OUT_DIR  # noqa: E402

SPEC_REPO = os.environ.get(
    "DATADOG_API_SPEC_REPO",
    os.path.expanduser("~/go/src/github.com/DataDog/datadog-api-spec"),
)
V2_DIR = os.path.join(SPEC_REPO, "spec", "v2")
IA_DIR = os.path.join(V2_DIR, "integration_accounts")

FRAGMENTS = [
    os.path.join(V2_DIR, "header.yaml"),
    os.path.join(V2_DIR, "shared.yaml"),
    os.path.join(IA_DIR, "integration_accounts.yaml"),
    os.path.join(IA_DIR, "integration_account_auth_methods.yaml"),
    os.path.join(IA_DIR, "interfaces", "twilio", "twilio.yaml"),
    os.path.join(IA_DIR, "interfaces", "twilio", "twilio_config.yaml"),
    os.path.join(IA_DIR, "interfaces", "elastic_cloud", "elastic_cloud.yaml"),
    os.path.join(IA_DIR, "interfaces", "elastic_cloud", "elastic_cloud_config.yaml"),
]

SLICES = [
    (
        "mini-datadog_integration_twilio_account.yaml",
        [
            "ListTwilioIntegrationAccounts",
            "CreateTwilioIntegrationAccount",
            "GetTwilioIntegrationAccount",
            "UpdateTwilioIntegrationAccount",
            "DeleteTwilioIntegrationAccount",
        ],
    ),
    (
        "mini-datadog_integration_elastic_cloud_account.yaml",
        [
            "ListElasticCloudIntegrationAccounts",
            "CreateElasticCloudIntegrationAccount",
            "GetElasticCloudIntegrationAccount",
            "UpdateElasticCloudIntegrationAccount",
            "DeleteElasticCloudIntegrationAccount",
        ],
    ),
]


def deep_merge(dst, src):
    for k, v in src.items():
        if k == "tags" and isinstance(v, list):
            # Each fragment can contribute its own domain tag; concatenate
            # rather than letting the first (empty, from shared.yaml) win.
            existing = {t.get("name") for t in dst.setdefault("tags", [])}
            dst["tags"].extend(t for t in v if t.get("name") not in existing)
        elif isinstance(v, dict) and isinstance(dst.get(k), dict):
            deep_merge(dst[k], v)
        elif k not in dst:
            dst[k] = v
        # else: first fragment wins (only header/shared.yaml define top-level
        # info/servers/security/openapi, so there's no real collision).


def load_merged_spec():
    spec = {}
    for path in FRAGMENTS:
        if not os.path.exists(path):
            sys.exit(f"spec fragment not found: {path}\n"
                      f"set DATADOG_API_SPEC_REPO to the datadog-api-spec checkout")
        with open(path) as f:
            frag = yaml.load(f, Loader=Loader) or {}
        deep_merge(spec, frag)
    # UnprocessableEntityResponse is authored per-domain across the spec repo
    # (not in shared.yaml) but every copy is identical; the integration_accounts
    # operations reference it without defining their own, so add the canonical
    # one (matching e.g. notification_rules.yaml) here.
    spec["components"].setdefault("responses", {}).setdefault(
        "UnprocessableEntityResponse",
        {
            "description": "The server cannot process the request because it contains invalid data.",
            "content": {
                "application/json": {
                    "schema": {"$ref": "#/components/schemas/JSONAPIErrorResponse"}
                }
            },
        },
    )
    return spec


def main():
    spec = load_merged_spec()

    op_index = {}
    for path, item in spec["paths"].items():
        if not isinstance(item, dict):
            continue
        for method, node in item.items():
            if method in HTTP_METHODS and isinstance(node, dict) and "operationId" in node:
                op_index[node["operationId"]] = (path, method)

    for filename, ops in SLICES:
        missing = [op for op in ops if op not in op_index]
        if missing:
            sys.exit(f"operationId(s) not found: {missing}")
        out_path = os.path.join(OUT_DIR, filename)
        collected = build_slice(spec, op_index, ops, out_path)
        print(f"wrote {out_path} ({len(collected)} components, {len(ops)} operations)")


if __name__ == "__main__":
    main()
