"""Data formatter."""
from functools import singledispatch
import warnings
import re

import dateutil.parser

from .utils import snake_case, camel_case, untitle_case, schema_name

PRIMITIVE_TYPES = ["string", "number", "boolean", "integer"]

KEYWORDS = {
    "break",
    "case",
    "chan",
    "const",
    "continue",
    "default",
    "defer",
    "else",
    "fallthrough",
    "for",
    "func",
    "go",
    "goto",
    "if",
    "import",
    "interface",
    "map",
    "package",
    "range",
    "return",
    "select",
    "struct",
    "switch",
    "type",
    "var",
}

SUFFIXES = {
    # Test
    "test",
    # $GOOS
    "aix",
    "android",
    "darwin",
    "dragonfly",
    "freebsd",
    "illumos",
    "js",
    "linux",
    "netbsd",
    "openbsd",
    "plan9",
    "solaris",
    "windows",
    # $GOARCH
    "386",
    "amd64",
    "arm",
    "arm64",
    "mips",
    "mips64",
    "mips64le",
    "mipsle",
    "ppc64",
    "ppc64le",
    "s390x",
    "wasm",
}


def block_comment(comment, prefix="#", first_line=True):
    lines = comment.split("\n")
    start = "" if first_line else lines[0] + "\n"
    return (start + "\n".join(f"{prefix} {line}".rstrip() for line in lines[(0 if first_line else 1) :])).rstrip()


def model_filename(name):
    filename = snake_case(name)
    last = filename.split("_")[-1]
    if last in SUFFIXES:
        filename += "_"
    return filename


def escape_reserved_keyword(word):
    """
    Escape reserved language keywords like openapi generator does it
    :param word: Word to escape
    :return: The escaped word if it was a reserved keyword, the word unchanged otherwise
    """
    if word in KEYWORDS:
        return f"{word}Var"
    return word


def attribute_name(attribute):
    return escape_reserved_keyword(camel_case(attribute))


def variable_name(attribute):
    return escape_reserved_keyword(untitle_case(camel_case(attribute)))


def simple_type(schema, render_nullable=False, render_new=False):
    """Return the simple type of a schema.

    :param schema: The schema to extract the type from
    :return: The simple type name
    """
    type_name = schema.get("type")
    type_format = schema.get("format")
    nullable = render_nullable and schema.get("nullable", False)

    nullable_prefix = "datadog.NewNullable" if render_new else "datadog.Nullable"

    if type_name == "integer":
        return {
            "int32": "int32" if not nullable else f"{nullable_prefix}Int32",
            "int64": "int64" if not nullable else f"{nullable_prefix}Int64",
            None: "int32" if not nullable else f"{nullable_prefix}Int32",
        }[type_format]

    if type_name == "number":
        return {
            "double": "float64" if not nullable else f"{nullable_prefix}Float64",
            None: "float" if not nullable else f"{nullable_prefix}Float",
        }[type_format]

    if type_name == "string":
        return {
            "date": "time.Time" if not nullable else f"{nullable_prefix}Time",
            "date-time": "time.Time" if not nullable else f"{nullable_prefix}Time",
            "email": "string" if not nullable else f"{nullable_prefix}String",
            "binary": "*os.File",
            None: "string" if not nullable else f"{nullable_prefix}String",
        }[type_format]
    if type_name == "boolean":
        return "bool" if not nullable else f"{nullable_prefix}Bool"

    return None


def is_reference(schema, attribute):
    """Check if an attribute is a reference."""
    is_required = attribute in schema.get("required", [])
    if is_required:
        return False

    attribute_schema = schema.get("properties", {}).get(attribute, {})

    is_nullable = attribute_schema.get("nullable", False)
    if is_nullable:
        return False

    is_anytype = attribute_schema.get("type", "object") == "object" and not (
        "properties" in attribute_schema
        or "oneOf" in attribute_schema
        or "anyOf" in attribute_schema
        or "allOf" in attribute_schema
    )
    if is_anytype:
        return False

    # no reference to arrays
    if attribute_schema.get("type", "object") == "array" or "items" in attribute_schema:
        return False

    return True


def attribute_path(attribute):
    return ".".join(attribute_name(a) for a in attribute.split("."))


def go_name(name):
    """Convert key to Go name.

    Example:

    >>> go_name("DASHBOARD_ID")
    DashboardID
    """
    return "".join(
        part.capitalize() if part not in {"API", "ID", "HTTP", "URL", "DNS"} else part for part in name.split("_")
    )


def reference_to_value(schema, value, print_nullable=True, **kwargs):
    """Return a reference to a value.

    :param schema: The schema to extract the type from
    :param value: The value to reference
    :return: The simple type name
    """
    type_name = schema.get("type")
    type_format = schema.get("format")
    nullable = schema.get("nullable", False)

    prefix = ""
    if type_name in PRIMITIVE_TYPES:
        prefix = "datadog."
    else:
        prefix = f"datadog{kwargs.get('version', '')}."

    if nullable and print_nullable:
        if value == "nil":
            formatter = "*{prefix}NewNullable{function_name}({value})"
        else:
            formatter = "*{prefix}NewNullable{function_name}({prefix}Ptr{function_name}({value}))"
    else:
        formatter = "datadog.Ptr{function_name}({value})"

    if type_name == "integer":
        function_name = {
            "int": "Int",
            "int32": "Int32",
            "int64": "Int64",
            None: "Int",
        }[type_format]
        return formatter.format(prefix=prefix, function_name=function_name, value=value)

    if type_name == "number":
        function_name = {
            "float": "Float32",
            "double": "Float64",
            None: "Float32",
        }[type_format]
        return formatter.format(prefix=prefix, function_name=function_name, value=value)

    if type_name == "string":
        function_name = {
            "date": "Time",
            "date-time": "Time",
            "email": "String",
            None: "String",
        }[type_format]
        return formatter.format(prefix=prefix, function_name=function_name, value=value)

    if type_name == "boolean":
        return formatter.format(prefix=prefix, function_name="Bool", value=value)

    if nullable:
        function_name = schema_name(schema)
        if function_name is None:
            raise NotImplementedError(f"nullable {schema} is not supported")
        return formatter.format(prefix=prefix, function_name=function_name, value=value)
    return f"&{value}"
