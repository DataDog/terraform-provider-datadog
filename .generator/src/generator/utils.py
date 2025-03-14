"""Utilities methods."""

import re

GET_OPERATION = "getOperation"
CREATE_OPERATION = "createOperation"
UPDATE_OPERATION = "updateOperation"
DELETE_OPERATION = "deleteOperation"

PRIMITIVE_TYPES = ["string", "number", "boolean", "integer"]

PATTERN_DOUBLE_UNDERSCORE = re.compile(r"__+")
PATTERN_LEADING_ALPHA = re.compile(r"(.)([A-Z][a-z]+)")
PATTERN_FOLLOWING_ALPHA = re.compile(r"([a-z0-9])([A-Z])")
PATTERN_WHITESPACE = re.compile(r"\W")


def snake_case(value):
    s1 = PATTERN_LEADING_ALPHA.sub(r"\1_\2", value)
    s1 = PATTERN_FOLLOWING_ALPHA.sub(r"\1_\2", s1).lower()
    s1 = PATTERN_WHITESPACE.sub("_", s1)
    s1 = s1.rstrip("_")
    return PATTERN_DOUBLE_UNDERSCORE.sub("_", s1)


def capitalize(value):
    return value[0].upper() + value[1:]


def camel_case(value):
    return "".join(capitalize(x) for x in snake_case(value).split("_"))


def untitle_case(value):
    return value[0].lower() + value[1:]


def schema_name(schema):
    if not schema:
        return None

    if hasattr(schema, "__reference__"):
        return schema.__reference__["$ref"].split("/")[-1]


def is_primitive(schema):
    if schema and schema.get("type") in PRIMITIVE_TYPES:
        return True
    return False


def is_required(schema, attr=None):
    req_args = schema.get("required")
    if req_args is None:
        return False
    if isinstance(req_args, bool):
        return req_args
    if isinstance(req_args, list):
        return attr in req_args
    raise ValueError(f"Invalid required value: {schema} ({attr})")


def is_computed(schema):
    v = schema.get("readOnly", None) is True
    return v


def is_enum(schema):
    return "enum" in schema


def is_nullable(schema):
    return schema.get("nullable", False)


def debug_filter(value):
    print(value)
    return value


def only_keep_filters(parameters: dict):
    """
    This function removes all element from a dict that are not considered filters.
    """
    for elt in parameters.copy().keys():
        if "filter" not in elt:
            parameters.pop(elt, None)
    return parameters
