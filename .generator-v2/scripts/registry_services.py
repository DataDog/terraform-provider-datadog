"""Resolve artifact service prefixes from tag registries for generation scripts."""
import argparse

import yaml


def pluralize(name):
    """Match the generator's deliberately small pluralization rules."""
    prefix, separator, word = name.rpartition("_")
    head = prefix + separator
    if word.endswith("y") and len(word) > 1 and word[-2] not in "aeiou":
        return head + word[:-1] + "ies"
    return head + word + ("es" if word.endswith(("s", "x", "z", "ch", "sh")) else "s")


def artifact_services(spec):
    for tag in spec.get("tags") or []:
        for family, declaration in (tag.get("x-datadog-tf-generator") or {}).items():
            if family.lower().startswith("x-") or declaration.get("skip"):
                continue
            for kind in ("resource", "datasource_singular", "datasource_list"):
                block = declaration.get(kind)
                if block is None or block.get("skip"):
                    continue
                name = block.get("artifact_name") or (
                    pluralize(family) if kind == "datasource_list" else family)
                yield name, tag["name"].lower().replace(" ", "-")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("spec")
    parser.add_argument("--artifact")
    args = parser.parse_args()
    with open(args.spec) as source:
        spec = yaml.safe_load(source)
    for name, service in artifact_services(spec):
        if args.artifact is None:
            print(f"{name}\t{service}")
        elif name == args.artifact:
            print(service)
            return


if __name__ == "__main__":
    main()
