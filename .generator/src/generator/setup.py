from jinja2 import Environment, FileSystemLoader, Template
import pathlib
import yaml

from jsonref import JsonRef

from . import openapi
from . import formatter
from . import utils
from . import type


def load_environment(version: str) -> Environment:
    env = Environment(
        loader=FileSystemLoader(str(pathlib.Path(__file__).parent / "templates"))
    )

    env.filters["attribute_name"] = formatter.attribute_name
    env.filters["camel_case"] = formatter.camel_case
    env.filters["sanitize_description"] = formatter.sanitize_description
    env.filters["simple_type"] = formatter.simple_type
    env.filters["snake_case"] = formatter.snake_case
    env.filters["untitle_case"] = formatter.untitle_case
    env.filters["variable_name"] = formatter.variable_name
    env.filters["parameter_schema"] = openapi.parameter_schema
    env.filters["parameters"] = openapi.parameters
    env.filters["is_json_api"] = openapi.is_json_api
    env.filters["capitalize"] = utils.capitalize
    env.filters["is_primitive"] = utils.is_primitive
    env.filters["response_type"] = type.get_type_for_response
    env.filters["return_type"] = type.return_type
    env.filters["tf_sort_params_by_type"] = type.tf_sort_params_by_type
    env.filters["tf_sort_properties_by_type"] = type.tf_sort_properties_by_type

    env.globals["enumerate"] = enumerate
    env.globals["get_name"] = openapi.get_name
    env.globals["get_terraform_primary_id"] = openapi.get_terraform_primary_id
    env.globals["json_api_attributes_schema"] = openapi.json_api_attributes_schema
    env.globals["get_terraform_schema_type"] = formatter.get_terraform_schema_type
    env.globals["get_type_for_parameter"] = type.get_type_for_parameter
    env.globals["get_type"] = type.type_to_go

    env.globals["GET_OPERATION"] = utils.GET_OPERATION
    env.globals["CREATE_OPERATION"] = utils.CREATE_OPERATION
    env.globals["UPDATE_OPERATION"] = utils.UPDATE_OPERATION
    env.globals["DELETE_OPERATION"] = utils.DELETE_OPERATION

    env.globals["version"] = utils.capitalize(version)

    return env


def load_templates(env: Environment) -> dict[str, Template]:
    templates = {
        "base": env.get_template("base_resource.j2"),
        "test": env.get_template("resource_test.j2"),
        "example": env.get_template("resource_example.j2"),
        "import": env.get_template("resource_import_example.j2"),
    }
    return templates


def load(filename):
    path = pathlib.Path(filename)
    with path.open() as fp:
        return JsonRef.replace_refs(yaml.safe_load(fp))
