import os
import pathlib

import click
from jinja2 import Environment, FileSystemLoader

from . import openapi
from . import formatter
from . import utils


@click.command()
@click.argument(
    "spec_path",
    # nargs=-1,
    type=click.Path(exists=True, file_okay=True, dir_okay=False, path_type=pathlib.Path),
)
@click.option(
    "-o",
    "--output",
    default="../datadog/",
    type=click.Path(path_type=pathlib.Path),
)
def cli(spec_path, output):
    """
    Generate a terraform code snippet from OpenAPI specification.
    """
    env = Environment(loader=FileSystemLoader(str(pathlib.Path(__file__).parent / "templates")))

    env.filters["attribute_name"] = formatter.attribute_name
    env.filters["camel_case"] = formatter.camel_case
    env.filters["sanitize_description"] = formatter.sanitize_description
    env.filters["parameter_schema"] = openapi.parameter_schema
    env.filters["parameters"] = openapi.parameters
    env.filters["response_type"] = openapi.get_type_for_response
    env.filters["return_type"] = openapi.return_type
    env.filters["simple_type"] = formatter.simple_type
    env.filters["snake_case"] = formatter.snake_case
    env.filters["untitle_case"] = formatter.untitle_case
    env.filters["upperfirst"] = utils.upperfirst
    env.filters["variable_name"] = formatter.variable_name
    env.filters["is_primitive"] = utils.is_primitive
    env.filters["is_json_api"] = openapi.is_json_api
    env.filters["tf_sort_params_by_type"] = openapi.tf_sort_params_by_type
    env.filters["tf_sort_properties_by_type"] = openapi.tf_sort_properties_by_type

    env.globals["enumerate"] = enumerate
    env.globals["get_name"] = openapi.get_name
    env.globals["get_type_for_parameter"] = openapi.get_type_for_parameter
    env.globals["get_type"] = openapi.type_to_go
    env.globals["get_terraform_schema_type"] = formatter.get_terraform_schema_type
    env.globals["get_terraform_primary_id"] = openapi.get_terraform_primary_id
    env.globals["json_api_attributes_schema"] = openapi.json_api_attributes_schema

    env.globals["GET_OPERATION"] = utils.GET_OPERATION
    env.globals["CREATE_OPERATION"] = utils.CREATE_OPERATION
    env.globals["UPDATE_OPERATION"] = utils.UPDATE_OPERATION
    env.globals["DELETE_OPERATION"] = utils.DELETE_OPERATION

    # Templates
    base_resource = env.get_template("base_resource.j2")
    base_resource_test = env.get_template("resource_test.j2")
    examples = {
        "resource.tf": env.get_template("resource_example.j2"),
        "import.sh": env.get_template("resource_import_example.j2"),
    }

    version = spec_path.parent.name
    spec = openapi.load(spec_path)
    env.globals["version"] = utils.upperfirst(version)

    operations_to_generate = openapi.operations_to_generate(spec)

    for name, operations in operations_to_generate.items():
        resource_filename = output / f"fwprovider/resource_datadog_{name}.go"
        resource_test_filename = output / "tests" / f"resource_datadog_{name}_test.go"
        with resource_filename.open("w") as fp:
            fp.write(base_resource.render(name=name, operations=operations))
        with resource_test_filename.open("w") as fp:
            fp.write(base_resource_test.render(name=name, operations=operations))

        examples_output = pathlib.Path(f"../examples/resources/datadog_{name}/")
        examples_output.mkdir(parents=True, exist_ok=True)
        for example_name, template in examples.items():
            file_name = examples_output / example_name
            with file_name.open("w") as fp:
                fp.write(template.render(name=name, operations=operations))
