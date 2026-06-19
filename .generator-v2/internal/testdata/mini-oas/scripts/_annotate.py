#!/usr/bin/env python3
"""Annotate selected mini OAS files with the x-datadog-tf-generator data_source
extension so the generator will emit a data source for them.

Picks the read operation (by-id GET preferred, else a GET, else the only op),
writes an annotated copy to gen-test/<name>.yaml. Originals stay pristine.
"""
import os
import yaml

try:
    from yaml import CSafeLoader as Loader, CSafeDumper as Dumper
except ImportError:
    from yaml import SafeLoader as Loader, SafeDumper as Dumper

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
MINI_OAS_DIR = os.path.dirname(SCRIPT_DIR)  # source slices live one level up
OUT = os.path.join(SCRIPT_DIR, "gen-test")

# Representative sample: small/medium, by-id read, list-only read, multi-op.
SAMPLE = [
    "cost_budget",
    "team",
    "incident_type",
    "datastore",
    "api_key",
    "user",
]


def pick_read_op(paths):
    """Return (path, method, operationId) for the data-source read operation."""
    gets = []
    for p, item in paths.items():
        for m, node in item.items():
            if m == "get" and isinstance(node, dict) and "operationId" in node:
                gets.append((p, m, node["operationId"]))
    if not gets:
        raise SystemExit("no GET operation to annotate")
    # Prefer a by-id GET (path has a {param}); that is the singular read.
    by_id = [g for g in gets if "{" in g[0]]
    return (by_id or gets)[0]


def main():
    os.makedirs(OUT, exist_ok=True)
    for name in SAMPLE:
        src = os.path.join(MINI_OAS_DIR, f"mini-datadog_{name}.yaml")
        spec = yaml.load(open(src), Loader=Loader)
        path, method, op_id = pick_read_op(spec["paths"])
        spec["paths"][path][method]["x-datadog-tf-generator"] = {
            "artifact_kind": "data_source",
            "artifact_name": name,
            "tf_description": f"Use this data source to retrieve information about an existing Datadog {name}.",
            "group": {"read": op_id},
        }
        dst = os.path.join(OUT, f"{name}.yaml")
        with open(dst, "w") as f:
            yaml.dump(spec, f, Dumper=Dumper, sort_keys=False, default_flow_style=False,
                      width=100, allow_unicode=True)
        print(f"annotated {name:<14} read={op_id:<28} {method.upper()} {path}")


if __name__ == "__main__":
    main()
