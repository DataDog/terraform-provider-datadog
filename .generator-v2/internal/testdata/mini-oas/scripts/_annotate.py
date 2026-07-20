#!/usr/bin/env python3
"""Annotate selected mini OAS files with the x-datadog-tf-generator data_source
extension so the generator will emit a data source for them. Originals stay
pristine; annotated copies go under gen-test/.

For each sample, annotated copies are written from the same slice:
  - gen-test/<name>.yaml    singular — resolves one record, shape per endpoints:
      * by-id GET only          -> group {read}            (read-only)
      * by-id GET + list GET     -> group {read, search}    (id-optional / both)
      * list GET only            -> group {search}          (search-only)
  - gen-test/<plural>.yaml  plural   — the collection GET as read, cardinality plural.
A slice contributes only the variants its endpoints support.
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

# Representative sample: by-id read, list read, and both-on-one-slice.
SAMPLE = [
    "cost_budget",
    "team",
    "incident_type",
    "datastore",
    "api_key",
    "user",
]


def _gets(paths):
    """All (path, 'get', operationId) GET operations in the spec."""
    out = []
    for p, item in paths.items():
        node = item.get("get") if isinstance(item, dict) else None
        if isinstance(node, dict) and "operationId" in node:
            out.append((p, "get", node["operationId"]))
    return out


def pick_read_op(paths):
    """Singular read: a by-id GET (path has a {param}), or None when list-only."""
    by_id = [g for g in _gets(paths) if "{" in g[0]]
    return by_id[0] if by_id else None


def pick_list_op(paths):
    """Plural read: a collection GET (path has no {param}), or None."""
    coll = [g for g in _gets(paths) if "{" not in g[0]]
    return coll[0] if coll else None


def pluralize(name):
    """Crude pluralization, sufficient for the sample (team->teams, user->users)."""
    return name if name.endswith("s") else name + "s"


def write_annotated(name, artifact_name, anchor, group, description, plural=False):
    """Load the pristine slice, write the extension (group + optional cardinality)
    onto the anchor operation, and save to gen-test/<artifact_name>.yaml."""
    spec = yaml.load(open(os.path.join(MINI_OAS_DIR, f"mini-datadog_{name}.yaml")), Loader=Loader)
    path, method, _ = anchor
    ext = {
        "artifact_kind": "data_source",
        "artifact_name": artifact_name,
        "tf_description": description,
        "group": group,
    }
    if plural:
        ext["cardinality"] = "plural"
    spec["paths"][path][method]["x-datadog-tf-generator"] = ext
    with open(os.path.join(OUT, f"{artifact_name}.yaml"), "w") as f:
        yaml.dump(spec, f, Dumper=Dumper, sort_keys=False, default_flow_style=False,
                  width=100, allow_unicode=True)
    kind = "plural" if plural else "singular"
    print(f"annotated {artifact_name:<16} {kind:<8} group={group}  ({method.upper()} {path})")


def main():
    os.makedirs(OUT, exist_ok=True)
    for name in SAMPLE:
        spec = yaml.load(open(os.path.join(MINI_OAS_DIR, f"mini-datadog_{name}.yaml")), Loader=Loader)
        paths = spec["paths"]
        thing = name.replace("_", " ")
        sing_desc = f"Use this data source to retrieve information about an existing Datadog {thing}."
        plural_desc = f"Use this data source to retrieve information about existing Datadog {thing}s."

        read = pick_read_op(paths)
        listed = pick_list_op(paths)

        # Singular: the resolution shape depends on which GETs the slice has — by-id
        # (read), by-id + list (both, id-optional), or list-only (search).
        if read and listed:
            write_annotated(name, name, read, {"read": read[2], "search": listed[2]}, sing_desc)
        elif read:
            write_annotated(name, name, read, {"read": read[2]}, sing_desc)
        elif listed:
            write_annotated(name, name, listed, {"search": listed[2]}, sing_desc)

        # Plural: the collection GET as a list read.
        if listed:
            write_annotated(name, pluralize(name), listed, {"read": listed[2]}, plural_desc, plural=True)

        if not read and not listed:
            raise SystemExit(f"{name}: no GET operation to annotate")


if __name__ == "__main__":
    main()
