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


def get_resources(spec: dict, config: dict) -> dict:
    """
    Generate a dictionary of resources and their CRUD operations.

    Args:
        spec (dict): The OpenAPI specification.
        config (dict): The configuration tagging resources to be generated.

    Raises:
        ValueError: If an unknown CRUD operation is encountered in the config.

    Returns:
        A dictionary containing the specifications of the resources to be generated.
    """
    resources_to_generate = {}

    # Iterate over each resource in the config
    for resource in config.get("resources", []):
        # Iterate over each CRUD operation for the resource
        for crud_operation, value in config["resources"][resource].items():
            method = value[
                "method"
            ].lower()  # Extract the HTTP method (GET, POST, etc.)
            path = value["path"]  # Extract the endpoint path
            operation = None

            # Match the CRUD operation to the corresponding constant
            match crud_operation:
                case "read":
                    operation = GET_OPERATION
                case "create":
                    operation = CREATE_OPERATION
                case "update":
                    operation = UPDATE_OPERATION
                case "delete":
                    operation = DELETE_OPERATION
                case _:
                    raise ValueError(f"Unknown operation {crud_operation}")

            # Add the operation details to the resources_to_generate dictionary
            resources_to_generate.setdefault(resource, {})[operation] = {
                "schema": spec["paths"][path][method],
                "path": path,
            }

    return resources_to_generate


def get_data_sources(spec: dict, config: dict) -> dict:
    """
    Creates a dictionary of data sources and their singular/plural endpoints.

    Args:
        spec (dict): The OpenAPI specification.
        config (dict): The configuration tagging resources to be generated.

    Returns:
        A dictionary containing the specifications of the data sources to be generated.
    """
    data_source_to_generate = {}

    for data_source in config.get("datasources", []):
        singular_path = config["datasources"][data_source]["singular"]
        data_source_to_generate.setdefault(data_source, {})["singular"] = {
            "schema": spec["paths"][singular_path]["get"],
            "path": singular_path,
        }
        plural_path = config["datasources"][data_source]["plural"]
        data_source_to_generate.setdefault(data_source, {})["plural"] = {
            "schema": spec["paths"][plural_path]["get"],
            "path": plural_path,
        }

    return data_source_to_generate


def get_terraform_primary_id(operations, path=UPDATE_OPERATION):
    update_params = parameters(operations[path]["schema"])
    primary_id = operations[path]["path"].split("/")[-1][1:-1]
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
