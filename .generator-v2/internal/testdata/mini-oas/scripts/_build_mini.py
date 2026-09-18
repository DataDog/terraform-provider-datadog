#!/usr/bin/env python3
"""Build minimal OpenAPI v2 specs for singular Datadog TF-provider datasources.

For each target datasource it:
  1. locates the Go datasource file,
  2. detects whether it uses the V1 API (skipped: no V1 spec available),
  3. extracts the operationId(s) it calls (Go client method name == operationId),
  4. slices the full v2 spec down to those operations + the transitive $ref closure.

Run with no args for a recon table (writes nothing). Pass --build to write files.

The full v2 spec is not vendored in this repo; point at it with the
DATADOG_OPENAPI_V2_SPEC env var (defaults to ~/local-dev/terraform/openapi.v2.yaml).
Slices are written next to this scripts/ directory (the mini-oas testdata dir).
"""
import os
import re
import sys
import yaml

try:
    from yaml import CSafeLoader as Loader, CSafeDumper as Dumper
except ImportError:
    from yaml import SafeLoader as Loader, SafeDumper as Dumper

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
OUT_DIR = os.path.dirname(SCRIPT_DIR)  # the mini-oas testdata dir holds the slices
# scripts/ -> mini-oas -> testdata -> internal -> .generator-v2 -> <repo root>
REPO = os.path.abspath(os.path.join(SCRIPT_DIR, *([os.pardir] * 5)))
SRC = os.environ.get(
    "DATADOG_OPENAPI_V2_SPEC",
    os.path.expanduser("~/local-dev/terraform/openapi.v2.yaml"),
)
FW_DIR = os.path.join(REPO, "datadog", "fwprovider")
SDK_DIR = os.path.join(REPO, "datadog")

# Targets: singular datasources #4-#44 (datadog_role #3 already done; #1,#2 are V1).
TARGETS = [
    "datadog_sensitive_data_scanner_standard_pattern",
    "datadog_service_level_objective",
    "datadog_synthetics_test",
    "datadog_user",
    "datadog_action_connection",
    "datadog_api_key",
    "datadog_app_builder_app",
    "datadog_apm_retention_filters_order",
    "datadog_aws_cur_config",
    "datadog_azure_uc_config",
    "datadog_cost_budget",
    "datadog_custom_allocation_rule",
    "datadog_dashboard_list",
    "datadog_datastore",
    "datadog_datastore_item",
    "datadog_gcp_uc_config",
    "datadog_incident_notification_rule",
    "datadog_incident_notification_template",
    "datadog_incident_type",
    "datadog_integration_aws_external_id",
    "datadog_integration_aws_iam_permissions",
    "datadog_integration_aws_iam_permissions_resource_collection",
    "datadog_integration_aws_iam_permissions_standard",
    "datadog_integration_aws_namespace_rules",
    "datadog_logs_pipelines_order",
    "datadog_metric_metadata",
    "datadog_metric_tags",
    "datadog_organization_settings",
    "datadog_powerpack",
    "datadog_reference_table",
    "datadog_rum_application",
    "datadog_security_monitoring_critical_asset",
    "datadog_security_monitoring_suppression",
    "datadog_sensitive_data_scanner_group_order",
    "datadog_service_account",
    "datadog_software_catalog",
    "datadog_synthetics_global_variable",
    "datadog_tag_pipeline_ruleset",
    "datadog_team",
    "datadog_team_notification_rule",
    "datadog_workflow_automation",
]

HTTP_METHODS = {"get", "post", "put", "delete", "patch", "head", "options"}
# API calls go through a receiver ending in `api`/`Api`: `d.Api.Op(`, `d.api.Op(`,
# `api.Op(` (helper param), or the SDK getter chain `GetXApiV2().Op(`.
CALL_RECV_RE = re.compile(r"[Aa]pi\.([A-Z][A-Za-z0-9]+)\s*\(")
CALL_SDK_RE = re.compile(r"ApiV[12]\(\)\s*\.\s*([A-Z][A-Za-z0-9]+)\s*\(")
# Broad fallback (any method call) — only when scoped + helper following find nothing.
CALL_ANY_RE = re.compile(r"\.([A-Z][A-Za-z0-9]+)\s*\(")
# A helper called with both the auth context and api client among its args.
HELPER_CALL_RE = re.compile(r"\b([A-Za-z_]\w*)\(([^)]*)")
# V1 usage: a V1 API receiver field or a V1 client getter.
V1_RE = re.compile(r"\*datadogV1\.\w*Api\b|ApiV1\(\)")
TYPENAME_RE = re.compile(r'\.TypeName\s*=\s*"([a-z0-9_]+)"')
FUNC_DEF_RE = re.compile(r"^func\s+(?:\([^)]*\)\s*)?([A-Za-z_]\w*)\s*\(", re.M)


def _go_files(*dirs):
    for d in dirs:
        for fn in os.listdir(d):
            if fn.endswith(".go") and (fn.startswith("data_source_") or fn.startswith("resource_")):
                yield os.path.join(d, fn)


def build_fw_index():
    """Map datadog_<suffix> -> framework datasource file path (any receiver var)."""
    idx = {}
    for path in _go_files(FW_DIR):
        if not os.path.basename(path).startswith("data_source_"):
            continue
        with open(path) as f:
            text = f.read()
        for m in TYPENAME_RE.finditer(text):
            idx["datadog_" + m.group(1)] = path
    return idx


def build_func_index():
    """Map top-level func name -> list of body texts across fwprovider + SDK dirs."""
    idx = {}
    for path in _go_files(FW_DIR, SDK_DIR):
        with open(path) as f:
            text = f.read()
        matches = list(FUNC_DEF_RE.finditer(text))
        for i, m in enumerate(matches):
            end = matches[i + 1].start() if i + 1 < len(matches) else len(text)
            idx.setdefault(m.group(1), []).append(text[m.start():end])
    return idx


def locate(name, fw_index):
    if name in fw_index:
        return fw_index[name]
    suffix = name[len("datadog_"):]
    candidates = [
        os.path.join(SDK_DIR, f"data_source_datadog_{suffix}.go"),
        os.path.join(SDK_DIR, f"data_source_datadog_{suffix}_.go"),  # e.g. synthetics_test_
    ]
    for c in candidates:
        if os.path.exists(c):
            return c
    return None


def _scoped_calls(text):
    """Method names invoked on an api receiver, incl. ...WithPagination variants."""
    names = set(CALL_RECV_RE.findall(text)) | set(CALL_SDK_RE.findall(text))
    out = set()
    for n in names:
        out.add(n)
        if n.endswith("WithPagination"):
            out.add(n[: -len("WithPagination")])
    return out


def extract_ops(text, op_ids, func_index):
    """Return (ops, was_fallback). Follows read-helpers passed the api client."""
    found = set(_scoped_calls(text))
    # Follow helpers that receive both auth + api (read logic often lives in resource files).
    for m in HELPER_CALL_RE.finditer(text):
        helper, arglist = m.group(1), m.group(2)
        if re.search(r"\.[Aa]uth\b", arglist) and re.search(r"\.[Aa]pi\b", arglist):
            for body in func_index.get(helper, []):
                found |= _scoped_calls(body)
    scoped = found & op_ids
    if scoped:
        return sorted(scoped), False
    broad = set(CALL_ANY_RE.findall(text)) & op_ids
    return sorted(broad), True


def find_refs(node, acc):
    if isinstance(node, dict):
        for k, v in node.items():
            if k == "$ref" and isinstance(v, str):
                acc.add(v)
            else:
                find_refs(v, acc)
    elif isinstance(node, list):
        for item in node:
            find_refs(item, acc)


def resolve(ref, spec):
    node = spec
    for part in ref[2:].split("/"):
        part = part.replace("~1", "/").replace("~0", "~")
        node = node[part]
    return node


def trim_oauth_scopes(scheme, keep):
    if scheme.get("type") != "oauth2":
        return scheme
    for flow in scheme.get("flows", {}).values():
        if "scopes" in flow:
            flow["scopes"] = {s: d for s, d in flow["scopes"].items() if s in keep}
    return scheme


def build_slice(spec, op_index, ops, out_path):
    """Slice the spec down to the given operationIds + transitive closure."""
    out = {
        "openapi": spec.get("openapi"),
        "info": spec.get("info"),
        "servers": spec.get("servers"),
        "security": spec.get("security"),
        "tags": None,
        "paths": {},
        "components": {},
    }

    tag_names, sec_blocks = set(), [spec.get("security")]
    for op in sorted(ops):
        path, method = op_index[op]
        path_item = spec["paths"][path]
        dst = out["paths"].setdefault(path, {})
        dst[method] = path_item[method]
        if "parameters" in path_item:  # path-level shared params
            dst["parameters"] = path_item["parameters"]
        tag_names.update(path_item[method].get("tags", []))
        sec_blocks.append(path_item[method].get("security"))

    out["tags"] = [t for t in spec.get("tags", []) if t.get("name") in tag_names]

    # Transitive $ref closure.
    collected = {}
    queue = []
    seed = set()
    find_refs(out["paths"], seed)
    queue.extend(seed)
    while queue:
        ref = queue.pop()
        if ref in collected:
            continue
        node = resolve(ref, spec)
        collected[ref] = node
        nested = set()
        find_refs(node, nested)
        queue.extend(r for r in nested if r not in collected)

    for ref, node in sorted(collected.items()):
        parts = ref[2:].split("/")
        assert parts[0] == "components" and len(parts) == 3, f"unexpected ref: {ref}"
        out["components"].setdefault(parts[1], {})[parts[2]] = node

    # Security schemes (referenced by name, not $ref) + trimmed oauth scopes.
    sec_names, keep_scopes = set(), {}
    for block in sec_blocks:
        for requirement in (block or []):
            for scheme, scopes in requirement.items():
                sec_names.add(scheme)
                keep_scopes.setdefault(scheme, set()).update(scopes)
    src_schemes = spec.get("components", {}).get("securitySchemes", {})
    schemes = {}
    for n in sorted(sec_names):
        if n in src_schemes:
            schemes[n] = trim_oauth_scopes(src_schemes[n], keep_scopes.get(n, set()))
    if schemes:
        out["components"]["securitySchemes"] = schemes

    with open(out_path, "w") as f:
        yaml.dump(out, f, Dumper=Dumper, sort_keys=False, default_flow_style=False,
                  width=100, allow_unicode=True)
    return collected


def main():
    do_build = "--build" in sys.argv
    if not os.path.exists(SRC):
        sys.exit(f"full v2 spec not found at {SRC}\n"
                 f"set DATADOG_OPENAPI_V2_SPEC to its path")
    print("loading spec ...", file=sys.stderr)
    with open(SRC) as f:
        spec = yaml.load(f, Loader=Loader)

    op_index = {}
    for path, item in spec["paths"].items():
        if not isinstance(item, dict):
            continue
        for method, node in item.items():
            if method in HTTP_METHODS and isinstance(node, dict) and "operationId" in node:
                op_index[node["operationId"]] = (path, method)
    op_ids = set(op_index)

    fw_index = build_fw_index()
    func_index = build_func_index()

    rows = []
    for name in TARGETS:
        path = locate(name, fw_index)
        if path is None:
            rows.append((name, "NO_FILE", "", "", ""))
            continue
        with open(path) as f:
            text = f.read()
        loc = os.path.relpath(path, REPO)
        if V1_RE.search(text):
            rows.append((name, "SKIP_V1", loc, "", ""))
            continue
        called, fallback = extract_ops(text, op_ids, func_index)
        if not called:
            rows.append((name, "NO_OPS", loc, "", ""))
            continue
        status = "REVIEW" if fallback else "BUILD"
        out_file = os.path.join(OUT_DIR, f"mini-{name}.yaml")
        ncomp = ""
        if do_build:
            collected = build_slice(spec, op_index, called, out_file)
            ncomp = str(len(collected))
            status = "WROTE*" if fallback else "WROTE"
        rows.append((name, status, loc, ",".join(called), ncomp))

    # Report
    w = max(len(r[0]) for r in rows)
    print(f"\n{'DATASOURCE':<{w}}  {'STATUS':<8}  OPERATIONS")
    print("-" * (w + 60))
    for name, status, loc, ops, ncomp in rows:
        extra = f"  [{ncomp} comps]" if ncomp else ""
        print(f"{name:<{w}}  {status:<8}  {ops}{extra}")
    n_build = sum(1 for r in rows if r[1].startswith(("BUILD", "WROTE", "REVIEW")))
    n_v1 = sum(1 for r in rows if r[1] == "SKIP_V1")
    n_bad = sum(1 for r in rows if r[1] in ("NO_FILE", "NO_OPS"))
    print(f"\n{n_build} to build (V2) | {n_v1} skipped (V1) | {n_bad} needs attention")
    print("(REVIEW/WROTE* = ops found via broad fallback, eyeball these)")


if __name__ == "__main__":
    main()
