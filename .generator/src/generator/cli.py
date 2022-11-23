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
    Generate a Go code snippet from OpenAPI specification.
    """
    env = Environment(loader=FileSystemLoader(str(pathlib.Path(__file__).parent / "templates")))

    env.filters["attribute_name"] = formatter.attribute_name
    env.filters["camel_case"] = formatter.camel_case
    env.filters["format_value"] = formatter.format_value
    env.filters["is_reference"] = formatter.is_reference
    env.filters["parameter_schema"] = openapi.parameter_schema
    env.filters["parameters"] = openapi.parameters
    env.filters["response_type"] = openapi.get_type_for_response
    env.filters["responses_by_types"] = openapi.responses_by_types
    env.filters["return_type"] = openapi.return_type
    env.filters["simple_type"] = formatter.simple_type
    env.filters["snake_case"] = formatter.snake_case
    env.filters["untitle_case"] = formatter.untitle_case
    env.filters["upperfirst"] = utils.upperfirst
    env.filters["variable_name"] = formatter.variable_name

    env.globals["enumerate"] = enumerate
    env.globals["get_name"] = openapi.get_name
    env.globals["get_type_for_attribute"] = openapi.get_type_for_attribute
    env.globals["get_type_for_parameter"] = openapi.get_type_for_parameter
    env.globals["get_type"] = openapi.type_to_go
    env.globals["get_default"] = openapi.get_default
    env.globals["get_container"] = openapi.get_container
    env.globals["get_container_type"] = openapi.get_container_type
    env.globals["get_type_at_path"] = openapi.get_type_at_path
    
    env.globals["GET_OPERATION"] = utils.GET_OPERATION
    env.globals["CREATE_OPERATION"] = utils.CREATE_OPERATION
    env.globals["UPDATE_OPERATION"] = utils.UPDATE_OPERATION
    env.globals["DELETE_OPERATION"] = utils.DELETE_OPERATION

    
    spec = openapi.load(spec_path)
    env.globals["version"] = spec_path.parent.name
    operations_to_generate = openapi.operations_to_generate(spec)
    
    base_resource = env.get_template("base_resource.j2")

    for name, operations in operations_to_generate.items():
        resource_filename = output / f"resource_datadog_{name}.go"
        with resource_filename.open("w") as fp:
            fp.write(base_resource.render(name=name, spec=spec, operations=operations))
