"""Data formatter."""

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
    "meta",
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


def sanitize_description(description):
    escaped_description = description.replace('"', '\\"')
    return " ".join(escaped_description.splitlines())


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


def get_terraform_schema_type(schema):
    return {
        "string": "String",
        "boolean": "Bool",
        "integer": "Int64",
        "number": "Int64",
        "array": "List",
        "object": "Block",
        None: "String",
    }[schema.get("type")]


def go_to_terraform_type_formatter(name: str, schema: dict) -> str:
    """
    This function is intended to be used in the Jinja2 templates.
    It was made to support the format enrichment of the OpenAPI schema.
    The format enrichment allows for a more appropriate Go type to be used in the provider (eg: string + date-time enrichment -> time.Time).
    However when updating the state we wish to use the primitive type that Terraform support instead.
    Args:
        name (str): The name of the variable to format.
        schema (dict): OpenApi spec as a dictionary. May contain a "format" key.
    Returns:
        str: The string representation of the variable in Go.
    """
    match schema.get("format"):
        case "date-time":
            return f"{variable_name(name)}.String()"
        case "date":
            return f"{variable_name(name)}.String()"
        case "binary":
            return f"string({variable_name(name)})"

        # primitive types should fall through
        case _:
            return f"*{variable_name(name)}"
