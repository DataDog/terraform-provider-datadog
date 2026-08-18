#!/usr/bin/env python3
"""Slice the full Datadog v2 OpenAPI spec down to a single data source and stamp
the x-datadog-tf-generator extension on it, in one pass — an agent-driven front
end over the same slicer _build_mini.py uses.

Unlike _build_mini.py (which discovers operations by scraping the provider's Go
files against a hardcoded TARGETS list), this takes the operations and metadata
as arguments: the caller names the operationIds, the artifact name, the doc
string, the cardinality, and optionally the hand-written constructor to retire.
It writes a self-contained slice — just those operations plus their transitive
$ref closure — to a temp file and prints that path on stdout, ready to feed to
`tfgen generate --spec <path>`.

Examples:
  # singular, id-optional (by-id read + list search):
  slice_and_annotate.py --artifact-name team \
      --tf-description "Use this data source to retrieve information about an existing Datadog team." \
      --read GetTeam --search ListTeams

  # plural (the collection GET as the read):
  slice_and_annotate.py --artifact-name teams --cardinality plural \
      --tf-description "Use this data source to retrieve information about existing Datadog teams." \
      --read ListTeams

  # overwrite a hand-written data source in place:
  slice_and_annotate.py --artifact-name team --read GetTeam --search ListTeams \
      --overwrites NewDatadogTeamDataSource

The full v2 spec is not vendored; point at it with --spec or the
DATADOG_OPENAPI_V2_SPEC env var (defaults to ~/local-dev/terraform/openapi.v2.yaml).
"""
import argparse
import os
import re
import sys
import tempfile

import yaml

# Reuse the slicer + spec-path default from _build_mini (same directory).
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from _build_mini import build_slice, HTTP_METHODS, Loader, SRC  # noqa: E402

ARTIFACT_NAME_RE = re.compile(r"^[a-z][a-z0-9_]*$")


def build_op_index(spec):
    """operationId -> (path, method) across the whole spec."""
    idx = {}
    for path, item in spec["paths"].items():
        if not isinstance(item, dict):
            continue
        for method, node in item.items():
            if method in HTTP_METHODS and isinstance(node, dict) and "operationId" in node:
                idx[node["operationId"]] = (path, method)
    return idx


def parse_args(argv):
    p = argparse.ArgumentParser(
        description="Slice the full v2 OAS to one data source and stamp the "
                    "x-datadog-tf-generator extension.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    p.add_argument("--spec", default=SRC,
                   help=f"path to the full v2 OpenAPI spec (default: {SRC})")
    p.add_argument("--artifact-name", required=True,
                   help="Terraform-facing name without the datadog_ prefix (snake_case)")
    p.add_argument("--artifact-kind", default="data_source",
                   choices=["data_source", "resource"],
                   help="default: data_source")
    p.add_argument("--tf-description", default="",
                   help="doc string shown in `terraform docs`")
    p.add_argument("--cardinality", default="singular",
                   choices=["singular", "plural"],
                   help="singular resolves one item; plural returns a filtered list "
                        "(pass the collection GET as --read). default: singular")
    p.add_argument("--overwrites", default="",
                   help="hand-written constructor this artifact retires, "
                        "e.g. NewDatadogTeamDataSource (data sources only)")
    # Group operationIds. At least one of --read / --search is required.
    p.add_argument("--read", default="", help="operationId of the read (by-id) endpoint")
    p.add_argument("--search", default="",
                   help="operationId of the list endpoint used to resolve one match")
    p.add_argument("--create", default="", help="operationId of the create endpoint (resources)")
    p.add_argument("--update", default="", help="operationId of the update endpoint (resources)")
    p.add_argument("--delete", default="", help="operationId of the delete endpoint (resources)")
    p.add_argument("--out", default="",
                   help="output file (default: a temp file under $TMPDIR/tfgen-slices/)")
    return p.parse_args(argv)


def main(argv=None):
    args = parse_args(argv if argv is not None else sys.argv[1:])

    if not ARTIFACT_NAME_RE.match(args.artifact_name) or len(args.artifact_name) > 64:
        sys.exit(f"invalid --artifact-name {args.artifact_name!r}: "
                 f"must match ^[a-z][a-z0-9_]*$ and be <= 64 chars")

    # Build the group in a readable order, dropping absent operations.
    group = {k: v for k, v in (
        ("read", args.read),
        ("search", args.search),
        ("create", args.create),
        ("update", args.update),
        ("delete", args.delete),
    ) if v}
    if not group.get("read") and not group.get("search"):
        sys.exit("need at least one of --read or --search")

    if not os.path.exists(args.spec):
        sys.exit(f"spec not found at {args.spec}\n"
                 f"pass --spec or set DATADOG_OPENAPI_V2_SPEC to its path")
    print(f"loading spec {args.spec} ...", file=sys.stderr)
    with open(args.spec) as f:
        spec = yaml.load(f, Loader=Loader)

    op_index = build_op_index(spec)
    missing = sorted({op for op in group.values() if op not in op_index})
    if missing:
        sys.exit("operationId(s) not found in spec: " + ", ".join(missing))

    ext = {
        "artifact_kind": args.artifact_kind,
        "artifact_name": args.artifact_name,
    }
    if args.tf_description:
        ext["tf_description"] = args.tf_description
    if args.cardinality == "plural":
        ext["cardinality"] = "plural"
    if args.overwrites:
        ext["overwrites"] = args.overwrites
    ext["group"] = group

    # Stamp the extension on the anchor operation (read, or search when read is
    # absent) before slicing — build_slice carries the operation across verbatim,
    # so the annotation lands in the slice with no second write.
    anchor_op = group.get("read") or group.get("search")
    apath, amethod = op_index[anchor_op]
    spec["paths"][apath][amethod]["x-datadog-tf-generator"] = ext

    out_path = args.out
    if out_path:
        os.makedirs(os.path.dirname(os.path.abspath(out_path)), exist_ok=True)
    else:
        out_dir = os.path.join(tempfile.gettempdir(), "tfgen-slices")
        os.makedirs(out_dir, exist_ok=True)
        out_path = os.path.join(out_dir, f"{args.artifact_name}.yaml")

    ops = sorted(set(group.values()))
    build_slice(spec, op_index, ops, out_path)

    print(f"wrote {args.cardinality} {args.artifact_kind} {args.artifact_name!r} "
          f"(anchor {amethod.upper()} {apath}, group={group}) -> {out_path}",
          file=sys.stderr)
    # stdout carries only the path, so callers can: SLICE=$(slice_and_annotate.py ...)
    print(os.path.abspath(out_path))


if __name__ == "__main__":
    main()
