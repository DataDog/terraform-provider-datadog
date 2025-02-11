from .utils import (
    GET_OPERATION,
    CREATE_OPERATION,
    UPDATE_OPERATION,
    DELETE_OPERATION,
)


def get_name(schema):
    name = None
    if hasattr(schema, "__reference__"):
        name = schema.__reference__["$ref"].split("/")[-1]

    return name


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
