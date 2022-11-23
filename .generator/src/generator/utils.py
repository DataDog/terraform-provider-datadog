"""Utilities methods."""
import re

GET_OPERATION = "getOperation"
CREATE_OPERATION = "createOperation"
UPDATE_OPERATION = "updateOperation"
DELETE_OPERATION = "deleteOperation"


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


def upperfirst(value):
    return value[0].upper() + value[1:]


def camel_case(value):
    return "".join(upperfirst(x) for x in snake_case(value).split("_"))


def untitle_case(value):
    return value[0].lower() + value[1:]


def schema_name(schema):
    if not schema:
        return None

    if hasattr(schema, "__reference__"):
        return schema.__reference__["$ref"].split("/")[-1]


def given_variables(context):
    """Return a list of variables using in given steps."""
    return {key for values in context.get("_given", {}).values() for key in values}
