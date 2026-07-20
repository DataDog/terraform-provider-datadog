#!/usr/bin/env python3
"""Validate every generated mini OAS: parses, openapi 3.0.0, all $refs resolve."""
import glob
import os
import yaml

try:
    from yaml import CSafeLoader as Loader
except ImportError:
    from yaml import SafeLoader as Loader

# Slices live one level up, in the mini-oas testdata dir.
MINI_OAS_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def collect_refs(node, acc):
    if isinstance(node, dict):
        for k, v in node.items():
            if k == "$ref" and isinstance(v, str):
                acc.add(v)
            else:
                collect_refs(v, acc)
    elif isinstance(node, list):
        for item in node:
            collect_refs(item, acc)


def resolve(spec, ref):
    node = spec
    for part in ref[2:].split("/"):
        node = node[part.replace("~1", "/").replace("~0", "~")]
    return node


def main():
    files = sorted(glob.glob(os.path.join(MINI_OAS_DIR, "mini-datadog_*.yaml")))
    ok = 0
    problems = []
    for f in files:
        spec = yaml.load(open(f), Loader=Loader)
        issues = []
        if spec.get("openapi") != "3.0.0":
            issues.append(f"openapi={spec.get('openapi')!r}")
        if not spec.get("paths"):
            issues.append("no paths")
        refs = set()
        collect_refs(spec, refs)
        dangling = []
        for r in sorted(refs):
            if not r.startswith("#/"):
                dangling.append(f"non-local:{r}")
                continue
            try:
                resolve(spec, r)
            except (KeyError, TypeError):
                dangling.append(r)
        if dangling:
            issues.append(f"{len(dangling)} dangling: {dangling[:3]}")
        nschemas = len(spec.get("components", {}).get("schemas", {}))
        if issues:
            problems.append((os.path.basename(f), issues))
        else:
            ok += 1
            print(f"OK  {os.path.basename(f):<62} refs={len(refs):<4} schemas={nschemas}")
    print(f"\n{ok}/{len(files)} files valid")
    for name, issues in problems:
        print(f"FAIL {name}: {issues}")


if __name__ == "__main__":
    main()
