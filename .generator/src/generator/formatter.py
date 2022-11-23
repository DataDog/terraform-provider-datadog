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


def format_value(value, quotes='"', schema=None):
    if schema and "enum" in schema:
        index = schema["enum"].index(value)
        enum_varnames = schema["x-enum-varnames"][index]
        name = schema_name(schema)
        return f"{name.upper()}_{enum_varnames}"

    if isinstance(value, str):
        return f"{quotes}{value}{quotes}"
    elif isinstance(value, bool):
        return "true" if value else "false"
    elif value is None:
        return "nil"
    return value


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


def format_parameters(data, spec, replace_values=None, has_body=False, **kwargs):
    parameters_spec = {p["name"]: p for p in spec.get("parameters", [])}
    if "requestBody" in spec and "multipart/form-data" in spec["requestBody"]["content"]:
        parent = spec["requestBody"]["content"]["multipart/form-data"]["schema"]
        for name, schema in parent["properties"].items():
            parameters_spec[name] = {
                "in": "form",
                "schema": schema,
                "name": name,
                "description": schema.get("description"),
                "required": name in parent.get("required", []),
            }

    parameters = ""
    has_optional = False
    for p in parameters_spec.values():
        required = p.get("required", False)
        if required:
            k = p["name"]
            v = data.pop(k)  # otherwise there is a missing required parameters
            value = format_data_with_schema(
                v["value"],
                p["schema"],
                name_prefix=f"datadog{kwargs.get('version', '')}.",
                replace_values=replace_values,
                required=True,
                **kwargs,
            )
            parameters += f"{value}, "
        else:
            has_optional = True

    body_is_required = spec.get("requestBody", {"required": None}).get("required", False)

    if has_body and body_is_required:
        parameters += "body, "
    if has_optional or body_is_required is False:
        parameters += f"*datadog{kwargs.get('version', '')}.New{spec['operationId'][0].upper()}{spec['operationId'][1:]}OptionalParameters()"
        if has_body and not body_is_required:
            parameters += ".WithBody(body)"

        for k, v in data.items():
            value = format_data_with_schema(
                v["value"],
                parameters_spec[k]["schema"],
                name_prefix=f"datadog{kwargs.get('version', '')}.",
                replace_values=replace_values,
                required=True,
                **kwargs,
            )
            parameters += f".With{camel_case(k)}({value})"

        parameters += ", "

    return parameters


@singledispatch
def format_data_with_schema(
    data,
    schema,
    name_prefix="",
    replace_values=None,
    default_name=None,
    required=False,
    in_list=False,
    **kwargs,
):
    if not schema:
        return ""

    nullable = schema.get("nullable", False)
    variables = kwargs.get("variables", set())

    name = schema_name(schema)

    if "enum" in schema and data not in schema["enum"]:
        raise ValueError(f"{data} is not valid enum value {schema['enum']}")

    if replace_values and data in replace_values:
        parameters = replace_values[data]

        # Make sure that variables used in given statements are camelCase for Go linter
        if parameters in variables:
            parameters = go_name(parameters)

        simple_type_value = simple_type(schema)
        if isinstance(data, int) and simple_type_value in {
            "float",
            "float32",
            "float64",
        }:
            parameters = f"{simple_type_value}({parameters})"
    else:
        if nullable and data is None:
            parameters = "nil"
        else:

            def format_string(x):
                if "`" in x:
                    x = re.sub(r"(`+)", r'` + "\1" + `', x)
                if x and ('"' in x or "\n" in x):
                    x = f"`{x}`"
                    x = re.sub(r" \+ ``$", "", x)
                    return x
                return f'"{x}"' if x else '""'

            def format_datetime(x):
                d = dateutil.parser.isoparse(x)
                return f"time.Date({d.year}, {d.month}, {d.day}, {d.hour}, {d.minute}, {d.second}, {d.microsecond}, time.UTC)"

            schema = schema.copy()

            def format_interface(x):
                if isinstance(x, int):
                    return str(x)
                if isinstance(x, float):
                    return str(x)
                if isinstance(x, str):
                    return format_string(x)
                raise TypeError(f"{x} is not supported type {schema}")

            def format_bool(x):
                if not isinstance(x, bool):
                    raise TypeError(f"{x} is not supported type {schema}")
                return "true" if x else "false"

            def open_file(x):
                return f"func() *os.File {{ fp, _ := os.Open({format_string(x)}); return fp }}()"

            formatter = {
                "int32": str,
                "int64": str,
                "double": str,
                "date-time": format_datetime,
                "number": str,
                "integer": str,
                "boolean": format_bool,
                "string": format_string,
                "email": format_string,
                "binary": open_file,
                None: format_interface,
            }[schema.get("format", schema.get("type"))]

            # TODO format date and datetime
            parameters = formatter(data)

    if "enum" in schema and name:
        # find schema index and get name from x-enum-varnames
        index = schema["enum"].index(data)
        enum_varnames = schema["x-enum-varnames"][index]
        parameters = f"{name_prefix}{name.upper()}_{enum_varnames}"
        if not required:
            parameters = f"{parameters}.Ptr()"
        return parameters
        # TODO handle nullable enums if necessary
        # return f"{name_prefix}{name}({parameters}){'.Ptr()' if not required else ''}"

    if in_list and nullable:
        schema = schema.copy()
        schema["nullable"] = False

    if (not required or schema.get("nullable")) and schema.get("type") is not None:
        return reference_to_value(schema, parameters, print_nullable=not in_list, **kwargs)

    if "oneOf" in schema:
        matched = 0
        one_of_schema = None
        for sub_schema in schema["oneOf"]:
            try:
                if sub_schema.get("nullable") and data is None:
                    # only one schema can be nullable
                    formatted = "nil"
                else:
                    sub_schema["nullable"] = False
                    formatted = format_data_with_schema(
                        data,
                        sub_schema,
                        name_prefix=name_prefix,
                        replace_values=replace_values,
                        **kwargs,
                    )
                if matched == 0:
                    one_of_schema = sub_schema
                    # NOTE we do not support mixed schemas with oneOf
                    # parameters += formatted
                    parameters = formatted
                matched += 1
            except (KeyError, ValueError, TypeError) as e:
                print(f"{e}")

        if matched == 0:
            raise ValueError(f"[{matched}] {data} is not valid for schema {name}")
        elif matched > 1:
            warnings.warn(f"[{matched}] {data} is not valid for schema {name}")

        one_of_schema_name = schema_name(one_of_schema)
        if not one_of_schema_name:
            one_of_schema_name = simple_type(one_of_schema).title()
        reference = "" if required or nullable else "&"
        if name:
            return f"{reference}{name_prefix}{name}{{\n{one_of_schema_name}: {parameters}}}"
        else:
            return f"{{{one_of_schema_name}: {reference}{parameters}}}"

    return parameters


@format_data_with_schema.register(list)
def format_data_with_schema_list(
    data,
    schema,
    name_prefix="",
    replace_values=None,
    default_name=None,
    required=False,
    in_list=False,
    **kwargs,
):
    if not schema:
        return ""

    if "oneOf" in schema:
        parameters = ""
        matched = 0
        one_of_schema = None
        for sub_schema in schema["oneOf"]:
            try:
                if sub_schema.get("nullable") and data is None:
                    formatted = "nil"
                else:
                    sub_schema["nullable"] = False
                    formatted = format_data_with_schema(
                        data,
                        sub_schema,
                        name_prefix=name_prefix,
                        replace_values=replace_values,
                        **kwargs,
                    )
                if matched == 0:
                    one_of_schema = sub_schema
                    parameters = formatted
                matched += 1
            except (KeyError, ValueError) as e:
                print(f"{e}")

        if matched == 0:
            raise ValueError(f"[{matched}] {data} is not valid for schema {schema}")
        elif matched > 1:
            warnings.warn(f"[{matched}] {data} is not valid for schema {schema}")

        one_of_schema_name = simple_type(one_of_schema) or f"{schema_name(one_of_schema)}"
        reference = "" if one_of_schema.get("required", False) else "&"
        return f"{{{one_of_schema_name}: {reference}{parameters}}}"

    parameters = ""
    # collect nested array types until you find a non-array type
    schema_parts = [(required, "[]")]
    list_schema = schema["items"]
    depth = 1
    while list_schema.get("type") == "array":
        schema_parts.append((not list_schema.get("nullable", False), "[]"))
        list_schema = list_schema["items"]
        depth += 1

    nested_prefix = list_schema.get("nullable", False) and "*" or ""
    nested_schema_name = schema_name(list_schema)
    if "oneOf" in list_schema:
        if schema_name(list_schema):
            nested_schema_name = f"{name_prefix}{schema_name(list_schema)}"
        elif schema_name(schema['items']):
            nested_schema_name = f"{name_prefix}{schema_name(schema['items'])}Item"
        else:
            nested_schema_name = "interface{}"
    else:
        nested_schema_name = schema_name(list_schema)
        nested_schema_name = f"{name_prefix}{nested_schema_name}" if nested_schema_name else "interface{}"

    nested_type = simple_type(list_schema)
    schema_parts.append(
        (
            not list_schema.get("nullable", False),
            nested_prefix + (nested_type if nested_type and not list_schema.get("enum") else nested_schema_name),
        )
    )
    nested_simple_type_name = "".join(p[1] for p in schema_parts)

    parameters = ""
    for d in data:
        value = format_data_with_schema(
            d,
            schema["items"],
            name_prefix=name_prefix,
            replace_values=replace_values,
            required=not schema["items"].get("nullable", False),
            in_list=True,
            **kwargs,
        )
        parameters += f"{value},\n"

    if in_list:
        for _ in range(depth):
            parameters = f"{{\n{parameters}}}"
        return parameters

    return f"{nested_simple_type_name}{{\n{parameters}}}"


@format_data_with_schema.register(dict)
def format_data_with_schema_dict(
    data,
    schema,
    name_prefix="",
    replace_values=None,
    default_name=None,
    required=False,
    in_list=False,
    **kwargs,
):
    if not schema:
        return ""

    reference = "" if required else "&"
    nullable = schema.get("nullable", False)

    name = schema_name(schema) or default_name

    parameters = ""
    if "properties" in schema:
        required_properties = set(schema.get("required", []))

        for k, v in data.items():
            if k not in schema["properties"]:
                continue
            value = format_data_with_schema(
                v,
                schema["properties"][k],
                name_prefix=name_prefix,
                replace_values=replace_values,
                default_name=name + camel_case(k) if name else None,
                required=k in required_properties,
                **kwargs,
            )
            parameters += f"{camel_case(k)}: {value},\n"

    if schema.get("additionalProperties"):
        saved_parameters = ""
        if schema.get("properties"):
            saved_parameters = parameters
            parameters = ""
        nested_schema = schema["additionalProperties"]
        nested_schema_name = simple_type(nested_schema)
        if not nested_schema_name:
            nested_schema_name = schema_name(nested_schema)
            if nested_schema_name:
                nested_schema_name = name_prefix + nested_schema_name

        has_properties = schema.get("properties")

        for k, v in data.items():
            if has_properties and k in schema["properties"]:
                continue
            value = format_data_with_schema(
                v,
                schema["additionalProperties"],
                name_prefix=name_prefix,
                replace_values=replace_values,
                required=True,
                **kwargs,
            )
            parameters += f'"{k}": {value},\n'

            # IMPROVE: find a better way to get nested schema name
            if not nested_schema_name:
                nested_schema_name = value.split("{")[0]

        if has_properties:
            if parameters:
                parameters = f"{saved_parameters}AdditionalProperties: map[string]{nested_schema_name}{{\n{parameters}}},\n"
            else:
                parameters = saved_parameters
        else:
            return f"map[string]{nested_schema_name}{{\n{parameters}}}"

    if "oneOf" in schema:
        matched = 0
        one_of_schema = None
        for sub_schema in schema["oneOf"]:
            try:
                if sub_schema.get("nullable") and data is None:
                    # only one schema can be nullable
                    formatted = "nil"
                else:
                    sub_schema["nullable"] = False
                    formatted = format_data_with_schema(
                        data,
                        sub_schema,
                        name_prefix=name_prefix,
                        replace_values=replace_values,
                        **kwargs,
                    )
                if matched == 0:
                    one_of_schema = sub_schema
                    # NOTE we do not support mixed schemas with oneOf
                    # parameters += formatted
                    parameters = formatted
                matched += 1
            except (KeyError, ValueError) as e:
                print(f"{e}")

        if matched == 0:
            raise ValueError(f"[{matched}] {data} is not valid for schema {name}")
        elif matched > 1:
            warnings.warn(f"[{matched}] {data} is not valid for schema {name}")

        one_of_schema_name = simple_type(one_of_schema) or f"{schema_name(one_of_schema)}"
        return f"{reference}{name_prefix}{name}{{\n{one_of_schema_name}: {parameters}}}"

    if schema.get("type") == "object" and "properties" not in schema:
        if schema.get("additionalProperties") == {}:
            name_prefix = ""
            name = "map[string]interface{}"
            reference = ""
            for k, v in data.items():
                parameters += f'"{k}": "{v}",\n'
        else:
            return "new(interface{})"

    if not name:
        warnings.warn(f"Unnamed schema {schema} for {data}")
        name_prefix = ""
        name = "map[string]interface"

    if parameters == "":
        # TODO select oneOf based on data
        warnings.warn(f"No schema matched for {data}")

    if nullable:
        return f"*{name_prefix}NewNullable{name}(&{name_prefix}{name}{{\n{parameters}}})"

    if in_list:
        return f"{{\n{parameters}}}"
    return f"{reference}{name_prefix}{name}{{\n{parameters}}}"
