"""Declare Terraform artifacts on the tag owning their anchor operation."""


def declare_artifact(spec, operation, name, block, fields):
    if not operation.get("tags"):
        operation["tags"] = ["Terraform"]
    tag_name = operation["tags"][0]
    tags = spec.setdefault("tags", [])
    tag = next((tag for tag in tags if tag.get("name") == tag_name), None)
    if tag is None:
        tag = {"name": tag_name}
        tags.append(tag)
    registry = tag.setdefault("x-datadog-tf-generator", {})
    registry.setdefault(name, {})[block] = {"artifact_name": name, **fields}
