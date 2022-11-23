import hashlib
import json
import pathlib
import random
import uuid
import warnings
import yaml

from jsonref import JsonRef
from urllib.parse import urlparse
from yaml import CSafeLoader

from . import formatter, utils


def load(filename):
    path = pathlib.Path(filename)
    with path.open() as fp:
        return JsonRef.replace_refs(yaml.load(fp, Loader=CSafeLoader))


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
        name = formatter.simple_type(schema, render_nullable=render_nullable, render_new=render_new)
        if name is not None:
            return name

    name = get_name(schema)
    if name:
        if "enum" in schema:
            return prefix + name
        if not (schema.get("additionalProperties") and not schema.get("properties")) and schema.get("type", "object") == "object":
            return prefix + name

    type_ = schema.get("type")
    if type_ is None:
        if "items" in schema:
            type_ = "array"
        elif "properties" in schema:
            type_ = "object"
        else:
            type_ = "object"
            warnings.warn(f"Unknown type for schema: {schema} ({name or alternative_name})")

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
            and ("properties" in schema or "oneOf" in schema or "anyOf" in schema or "allOf" in schema)
            else "interface{}"
        )

    raise ValueError(f"Unknown type {type_}")


def get_type_for_attribute(schema, attribute, current_name=None):
    """Return Go type name for the attribute."""
    child_schema = schema.get("properties", {}).get(attribute)
    alternative_name = current_name + formatter.camel_case(attribute) if current_name else None
    return type_to_go(child_schema, alternative_name=alternative_name)


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
    operations = {}

    for path in spec["paths"]:
        for method in spec["paths"][path]:
            operation = spec["paths"][path][method]
            if "x-terraform-resource" in operation:
                if method == "get":
                    operations.setdefault(operation["x-terraform-resource"], {})[utils.GET_OPERATION] = operation
                elif method == "post":
                    operations.setdefault(operation["x-terraform-resource"], {})[utils.CREATE_OPERATION] = operation
                elif method == "patch":
                    operations.setdefault(operation["x-terraform-resource"], {})[utils.UPDATE_OPERATION] = operation
                elif method == "delete":
                    operations.setdefault(operation["x-terraform-resource"], {})[utils.DELETE_OPERATION] = operation

    return operations


def parameters(operationList):
    parametersDict = {}

    for operation in operationList:
        for content in operation.get("parameters", []):
            if "schema" in content and content.get("required"):
                parametersDict[content["name"]] = content

        if "requestBody" in operation:
            if "multipart/form-data" in operation["requestBody"]["content"]:
                parent = operation["requestBody"]["content"]["multipart/form-data"]["schema"]
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

        for content in operation.get("parameters", []):
            if "schema" in content and not content.get("required"):
                parametersDict[content["name"]] = content

    return parametersDict


def parameter_schema(parameter):
    if "schema" in parameter:
        return parameter["schema"]
    if "content" in parameter:
        for content in parameter.get("content", {}).values():
            if "schema" in content:
                return content["schema"]
    raise ValueError(f"Unknown schema for parameter {parameter}")


def return_type(operation):
    for response in operation.get("responses", {}).values():
        for content in response.get("content", {}).values():
            if "schema" in content:
                return type_to_go(content["schema"])
        return


def response_code_and_accept_type(operation, status_code=None):
    for response in operation["responses"]:
        if status_code is None:
            return int(response), next(iter(operation["responses"][response].get("content", {None: None})))
        if response == str(status_code):
            return status_code, next(iter(operation["responses"][response].get("content", {None: None})))
    return status_code, None


def request_content_type(operation, status_code=None):
    return next(iter(operation.get("requestBody", {}).get("content", {None: None})))


def response(operation, status_code=None):
    for response in operation["responses"]:
        if status_code is None or response == str(status_code):
            return list(operation["responses"][response]["content"].values())[0]["schema"]
    return None


def generate_value(schema, use_random=False, prefix=None):
    spec = schema.spec
    if not use_random:
        if "example" in spec:
            return spec["example"]
        if "default" in spec:
            return spec["default"]

    if spec["type"] == "string":
        if use_random:
            return str(
                uuid.UUID(
                    bytes=hashlib.sha256(
                        str(prefix or schema.keys).encode("utf-8"),
                    ).digest()[:16]
                )
            )
        return "string"
    elif spec["type"] == "integer":
        return random.randint(0, 32000) if use_random else len(str(prefix or schema.keys))
    elif spec["type"] == "number":
        return random.random() if use_random else 1.0 / len(str(prefix or schema.keys))
    elif spec["type"] == "boolean":
        return True
    elif spec["type"] == "array":
        return [generate_value(schema[0], use_random=use_random)]
    elif spec["type"] == "object":
        return {key: generate_value(schema[key], use_random=use_random) for key in spec["properties"]}
    else:
        raise TypeError(f"Unknown type: {spec['type']}")


def is_primitive(schema):
    # We resolve enums to ClassName.ENUM so don't treat enum's as primitive
    if schema.get("type") in utils.PRIMITIVE_TYPES:
        return True
    return False


def get_terraform_type(schema):
    return {
        "string": "TypeString",
        "boolean": "TypeBool",
        "integer": "TypeInt",
        "number": "TypeInt",
        "array": "TypeList",
        "object": "TypeList",
        None: "String",
    }[schema.get("type")]



class Schema:
    def __init__(self, spec, value=None, keys=None):
        self.spec = spec
        self.value = value if value is not None else generate_value
        self.keys = keys or tuple()

    def __getattr__(self, key):
        return self[key]

    def __getitem__(self, key):
        type_ = self.spec.get("type", "object")
        if type_ == "object":
            try:
                return self.__class__(
                    self.spec["properties"][key],
                    value=self.value,
                    keys=self.keys + (key,),
                )
            except KeyError:
                if "oneOf" in self.spec:
                    for schema in self.spec["oneOf"]:
                        if schema.get("type", "object") == "object":
                            try:
                                return self.__class__(
                                    schema["properties"][key],
                                    value=self.value,
                                    keys=self.keys + (key,),
                                )
                            except KeyError:
                                pass
            raise KeyError(f"{key} not found in {self.spec.get('properties', {}).keys()}: {self.spec}")
        if type_ == "array":
            return self.__class__(self.spec["items"], value=self.value, keys=self.keys + (key,))

        raise KeyError(f"{key} not found in {self.spec}")

    def __repr__(self):
        value = self.value(self)
        if isinstance(value, (dict, list)):
            return json.dumps(value, indent=2)
        return str(value)


class Operation:
    def __init__(self, name, spec, method, path):
        self.name = name
        self.spec = spec
        self.method = method
        self.path = path

    def server_url_and_method(self, spec, server_index=0, server_variables=None):
        def format_server(server, path):
            url = server["url"] + path
            # replace potential path variables
            for variable, value in server_variables.items():
                url = url.replace("{" + variable + "}", value)
            # replace server variables if they were not replace before
            for variable in server["variables"]:
                if variable in server_variables:
                    continue
                url = url.replace(
                    "{" + variable + "}",
                    server["variables"][variable]["default"],
                )
            return url

        server_variables = server_variables or {}
        if "servers" in self.spec:
            server = self.spec["servers"][server_index]
        else:
            server = spec["servers"][server_index]
        return format_server(server, self.path), self.method

    def response_code_and_accept_type(self):
        for response in self.spec["responses"]:
            return int(response), next(iter(self.spec["responses"][response].get("content", {None: None})))
        return None, None

    def request_content_type(self):
        return next(iter(self.spec.get("requestBody", {}).get("content", {None: None})))

    def response(self):
        for response in self.spec["responses"]:
            return Schema(next(iter((self.spec["responses"][response]["content"].values())))["schema"])

    def request(self):
        return Schema(next(iter(self.spec["requestBody"]["content"].values()))["schema"])
