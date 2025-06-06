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


def clean_response_for_datasource(schema: dict):
    schema_save = schema.copy()
    try:
        schema["properties"] = remove_all_but(
            schema=schema["properties"], field_to_keep="data"
        )
        schema["properties"]["data"]["properties"] = remove_all_but(
            schema=schema["properties"]["data"]["properties"],
            field_to_keep="attributes",
        )
        schema = move_fields_to_top(
            schema=schema,
            path_to_fields=["properties", "data", "properties", "attributes"],
        )
    except KeyError:
        print("Error while cleaning response for datasource, restoring original schema")
        return schema_save
    return schema


def remove_all_but(schema: dict, field_to_keep: str):
    """
    This function removes elements from a schema that are unwanted.
    This function is meant to be used when generating data sources as we do not want to generate models for all fields (eg: "relationships" or "included").
    """
    for elt in schema.copy().keys():
        if field_to_keep not in elt:
            schema.pop(elt, None)
    return schema


def move_fields_to_top(schema: dict, path_to_fields: list[str]):
    """
    This function moves a field to the top of the schema.
    This function is meant to be used when generating data sources as we want to have the fields in [properties][data][properties][attributes] at the top level of the schema.
    """
    tmp = schema
    for field in path_to_fields:
        tmp = tmp[field]

    for fields in tmp:
        schema[fields] = tmp[fields]

    return schema
