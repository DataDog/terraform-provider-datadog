"""Data formatter."""
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
