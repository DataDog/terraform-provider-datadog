import json
import pathlib
import warnings
import yaml

from jsonref import JsonRef

from . import formatter
from .utils import (
    GET_OPERATION,
    CREATE_OPERATION,
    UPDATE_OPERATION,
    DELETE_OPERATION,
    is_primitive,
)


def load(filename):
    path = pathlib.Path(filename)
    with path.open() as fp:
        return JsonRef.replace_refs(yaml.safe_load(fp))


def get_name(schema):
    name = None
    if hasattr(schema, "__reference__"):
        name = schema.__reference__["$ref"].split("/")[-1]

    return name


def type_to_go(schema, alternative_name=None, render_nullable=False, render_new=False):
    """Return Go type name for the type."""
    if render_nullable and schema.get("nullable", False):
        prefix = "Nullable"
    else:
        prefix = ""

    # special case for additionalProperties: true
    if schema is True:
        return "interface{}"

    if "enum" not in schema:
        name = formatter.simple_type(
            schema, render_nullable=render_nullable, render_new=render_new
        )
        if name is not None:
            return name

    name = get_name(schema)
    if name:
        if "enum" in schema:
            return prefix + name
        if (
            not (schema.get("additionalProperties") and not schema.get("properties"))
            and schema.get("type", "object") == "object"
        ):
            return prefix + name

    type_ = schema.get("type")
    if type_ is None:
        if "items" in schema:
            type_ = "array"
        elif "properties" in schema:
            type_ = "object"
        else:
            type_ = "object"
            warnings.warn(
                f"Unknown type for schema: {schema} ({name or alternative_name})"
            )

    if type_ == "array":
        if name and schema.get("x-generate-alias-as-model", False):
            return prefix + name
        if name or alternative_name:
            alternative_name = (name or alternative_name) + "Item"
        name = type_to_go(schema["items"], alternative_name=alternative_name)
        # handle nullable arrays
        if formatter.simple_type(schema["items"]) and schema["items"].get("nullable"):
            name = "*" + name
        return "[]{}".format(name)
    elif type_ == "object":
        if "additionalProperties" in schema:
            return "map[string]{}".format(type_to_go(schema["additionalProperties"]))
        return (
            prefix + alternative_name
            if alternative_name
            and (
                "properties" in schema
                or "oneOf" in schema
                or "anyOf" in schema
                or "allOf" in schema
            )
            else "interface{}"
        )

    raise ValueError(f"Unknown type {type_}")


def get_type_for_parameter(parameter):
    """Return Go type name for the parameter."""
    if "content" in parameter:
        assert "in" not in parameter
        for content in parameter["content"].values():
            return type_to_go(content["schema"])
    return type_to_go(parameter.get("schema"))


def get_type_for_response(response):
    """Return Go type name for the response."""
    if "content" in response:
        for content in response["content"].values():
            if "schema" in content:
                return type_to_go(content["schema"])


def operations_to_generate(spec):
    """
    {
        "resourceName": {
            "getOperation": {
                "path": "endpoint/path",
                "schema": {...}
            },
            ...
        }
    }
    """
    operations = {}
    for path in spec["paths"]:
        for method in spec["paths"][path]:
            operation = spec["paths"][path][method]
            if "x-terraform-resource" in operation:
                if method == "get":
                    operations.setdefault(operation["x-terraform-resource"], {})[
                        GET_OPERATION
                    ] = {"schema": operation, "path": path}
                elif method == "post":
                    operations.setdefault(operation["x-terraform-resource"], {})[
                        CREATE_OPERATION
                    ] = {"schema": operation, "path": path}
                elif method == "patch" or method == "put":
                    operations.setdefault(operation["x-terraform-resource"], {})[
                        UPDATE_OPERATION
                    ] = {"schema": operation, "path": path}
                elif method == "delete":
                    operations.setdefault(operation["x-terraform-resource"], {})[
                        DELETE_OPERATION
                    ] = {"schema": operation, "path": path}

    return operations


def get_terraform_primary_id(operations):
    update_params = parameters(operations[UPDATE_OPERATION]["schema"])
    primary_id = operations[UPDATE_OPERATION]["path"].split("/")[-1][1:-1]
    primary_id_param = update_params.pop(primary_id)

    return {"schema": parameter_schema(primary_id_param), "name": primary_id}


def parameters(operation):
    parametersDict = {}
    for content in operation.get("parameters", []):
        if "schema" in content:
            parametersDict[content["name"]] = content

    if "requestBody" in operation:
        if "multipart/form-data" in operation["requestBody"]["content"]:
            parent = operation["requestBody"]["content"]["multipart/form-data"][
                "schema"
            ]
            for name, schema in parent["properties"].items():
                parametersDict[name] = {
                    "in": "form",
                    "schema": schema,
                    "name": name,
                    "description": schema.get("description"),
                    "required": name in parent.get("required", []),
                }
        else:
            name = operation.get("x-codegen-request-body-name", "body")
            parametersDict[name] = operation["requestBody"]

    return parametersDict


def parameter_schema(parameter):
    if "schema" in parameter:
        return parameter["schema"]
    if "content" in parameter:
        for content in parameter.get("content", {}).values():
            if "schema" in content:
                return content["schema"]
    raise ValueError(f"Unknown schema for parameter {parameter}")


def tf_sort_params_by_type(parameters):
    """
    Sort parameters by primitive and non primitive types since
    we use Blocks in terraform instead of NestedAttributes for
    non primitives
    """
    primitive = {}
    primitive_array = {}
    non_primitive_array = {}
    non_primitive_obj = {}
    for name, p in parameters.items():
        schema = parameter_schema(p)
        if is_json_api(schema):
            json_attr_schema = json_api_attributes_schema(schema)
            for attr, s in json_attr_schema["properties"].items():
                required = attr in json_attr_schema.get("required", [])
                s["_tf_required"] = required
                if is_primitive(s):
                    primitive[attr] = s
                elif s.get("type") == "array":
                    if is_primitive(s.get("items")):
                        primitive_array[attr] = s
                    else:
                        non_primitive_array[attr] = s
                else:
                    non_primitive_obj[attr] = s
        else:
            for attr, s in schema["properties"].items():
                required = attr in schema.get("required", [])
                s["_tf_required"] = required
                if is_primitive(s):
                    primitive[attr] = s
                elif s.get("type") == "array":
                    if is_primitive(s.get("items")):
                        primitive_array[attr] = s
                    else:
                        non_primitive_array[attr] = s
                else:
                    non_primitive_obj[attr] = s

    return primitive, primitive_array, non_primitive_array, non_primitive_obj


def tf_sort_properties_by_type(schema):
    """
    Sort schema properties by primitive and non primitive types since
    we use Blocks in terraform instead of NestedAttributes for
    non primitives
    """
    primitive = {}
    primitive_array = {}
    non_primitive_array = {}
    non_primitive_obj = {}
    for name, cSchema in schema.get("properties", {}).items():
        if is_json_api(cSchema):
            json_attr_schema = json_api_attributes_schema(cSchema)
            for attr, s in json_attr_schema["properties"].items():
                required = attr in json_attr_schema.get("required", [])
                s["_tf_required"] = required
                if is_primitive(s):
                    primitive[attr] = s
                elif s.get("type") == "array":
                    if is_primitive(s.get("items")):
                        primitive_array[attr] = s
                    else:
                        non_primitive_array[attr] = s
                else:
                    non_primitive_obj[attr] = s
        else:
            if is_primitive(cSchema):
                primitive[name] = cSchema
            elif cSchema.get("type") == "array":
                if is_primitive(cSchema.get("items")):
                    primitive_array[name] = cSchema
                else:
                    non_primitive_array[name] = cSchema
            else:
                non_primitive_obj[name] = cSchema

    if "oneOf" in schema:
        for oneOf in schema.get("oneOf"):
            schemaName = formatter.snake_case(get_name(oneOf))
            non_primitive_obj[schemaName] = oneOf

    return primitive, primitive_array, non_primitive_array, non_primitive_obj


def return_type(operation):
    for response in operation.get("responses", {}).values():
        for content in response.get("content", {}).values():
            if "schema" in content:
                return type_to_go(content["schema"]), content["schema"]
        return


def is_json_api(schema):
    properties = schema.get("properties", {})
    if "data" in properties:
        data_properties = properties["data"].get("properties", {})
        if "type" in data_properties and "attributes" in data_properties:
            return True
    return False


def json_api_attributes_schema(schema):
    return (
        schema.get("properties", {})
        .get("data", {})
        .get("properties", {})
        .get("attributes", {})
    )
