#!/usr/bin/env python3
"""Build a minimal OpenAPI spec for the datadog_role (singular) datasource.

The datasource calls GetRolesApiV2().ListRoles(auth, ...WithFilter), i.e. the
GET /api/v2/roles operation (operationId: ListRoles), and reads from the
RolesResponse body. This script slices the full v2 spec down to just that
operation plus the transitive closure of every component it references.

The full v2 spec is not vendored in this repo; point at it with the
DATADOG_OPENAPI_V2_SPEC env var (defaults to ~/local-dev/terraform/openapi.v2.yaml).
The slice is written next to this scripts/ directory (the mini-oas testdata dir).
"""
import os
import sys
import yaml

try:
    from yaml import CSafeLoader as Loader, CSafeDumper as Dumper
except ImportError:
    from yaml import SafeLoader as Loader, SafeDumper as Dumper

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MINI_OAS_DIR = os.path.dirname(SCRIPT_DIR)
SRC = os.environ.get(
    "DATADOG_OPENAPI_V2_SPEC",
    os.path.expanduser("~/local-dev/terraform/openapi.v2.yaml"),
)
DST = os.path.join(MINI_OAS_DIR, "mini-datadog_role.yaml")

PATH = "/api/v2/roles"
OPERATION = "get"  # ListRoles


def find_refs(node, acc):
    """Collect every $ref string reachable from node."""
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
    """Resolve a local JSON pointer like '#/components/schemas/Foo'."""
    assert ref.startswith("#/"), f"non-local ref: {ref}"
    node = spec
    for part in ref[2:].split("/"):
        part = part.replace("~1", "/").replace("~0", "~")
        node = node[part]
    return node


def security_scheme_names(*security_blocks):
    names = set()
    for block in security_blocks:
        for requirement in (block or []):
            names.update(requirement.keys())
    return names


def referenced_scopes(scheme_name, *security_blocks):
    """Scopes requested for scheme_name across the kept security requirements."""
    scopes = set()
    for block in security_blocks:
        for requirement in (block or []):
            scopes.update(requirement.get(scheme_name, []))
    return scopes


def trim_oauth_scopes(scheme, keep_scopes):
    """For an oauth2 scheme, keep only the referenced scopes (minimal + valid)."""
    if scheme.get("type") != "oauth2":
        return scheme
    for flow in scheme.get("flows", {}).values():
        if "scopes" in flow:
            flow["scopes"] = {s: d for s, d in flow["scopes"].items() if s in keep_scopes}
    return scheme


def main():
    if not os.path.exists(SRC):
        sys.exit(f"full v2 spec not found at {SRC}\n"
                 f"set DATADOG_OPENAPI_V2_SPEC to its path")
    with open(SRC) as f:
        spec = yaml.load(f, Loader=Loader)

    op = spec["paths"][PATH][OPERATION]

    out = {
        "openapi": spec.get("openapi"),
        "info": spec.get("info"),
        "servers": spec.get("servers"),
        "security": spec.get("security"),
        "tags": [t for t in spec.get("tags", []) if t.get("name") in set(op.get("tags", []))],
        "paths": {PATH: {OPERATION: op}},
        "components": {},
    }

    # BFS over the transitive $ref closure starting from the kept operation.
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

    # Place every collected component back under components/<section>/<name>.
    for ref, node in sorted(collected.items()):
        parts = ref[2:].split("/")
        assert parts[0] == "components" and len(parts) == 3, f"unexpected ref shape: {ref}"
        out["components"].setdefault(parts[1], {})[parts[2]] = node

    # Security schemes are referenced by name (not $ref) in security requirements.
    sec_names = security_scheme_names(out.get("security"), op.get("security"))
    src_schemes = spec.get("components", {}).get("securitySchemes", {})
    schemes = {}
    for n in sorted(sec_names):
        if n not in src_schemes:
            continue
        scheme = src_schemes[n]
        keep = referenced_scopes(n, out.get("security"), op.get("security"))
        schemes[n] = trim_oauth_scopes(scheme, keep)
    if schemes:
        out["components"]["securitySchemes"] = schemes

    with open(DST, "w") as f:
        yaml.dump(out, f, Dumper=Dumper, sort_keys=False, default_flow_style=False, width=100, allow_unicode=True)

    # Report
    print(f"wrote {DST}")
    for section, items in out["components"].items():
        print(f"  components/{section}: {len(items)} -> {', '.join(sorted(items))}")
    missing = [n for n in sec_names if n not in src_schemes]
    if missing:
        print(f"  WARNING missing security schemes: {missing}", file=sys.stderr)


if __name__ == "__main__":
    main()
