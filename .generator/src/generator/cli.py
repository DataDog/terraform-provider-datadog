import os
import pathlib
import click
import subprocess

from jinja2 import Template

from . import setup
from . import openapi


@click.command()
@click.argument(
    "spec_path",
    type=click.Path(
        exists=True, file_okay=True, dir_okay=False, path_type=pathlib.Path
    ),
)
@click.argument(
    "config_path",
    type=click.Path(
        exists=True, file_okay=True, dir_okay=False, path_type=pathlib.Path
    ),
)
@click.option("--go-fmt/--no-go-fmt", default=True)
def cli(spec_path, config_path, go_fmt):
    """
    Generate a terraform code snippet from OpenAPI specification.
    """
    env = setup.load_environment(version=spec_path.parent.name)

    templates = setup.load_templates(env=env)

    spec = setup.load(spec_path)
    config = setup.load(config_path)

    data_sources_to_generate = openapi.get_data_sources(spec, config)
    for name, data_source in data_sources_to_generate.items():
        generate_data_source(
            name=name, data_source=data_source, templates=templates, go_fmt=go_fmt
        )

    # resources_to_generate = openapi.get_resources(spec, config)

    # for name, resource in resources_to_generate.items():
    #     generate_resource(
    #         name=name,
    #         resource=resource,
    #         templates=templates,
    #         go_fmt=go_fmt,
    #     )


def generate_data_source(
    name: str, data_source: dict, templates: dict[str, Template], go_fmt: bool
) -> None:
    output = pathlib.Path("../datadog/")
    filename = output / f"fwprovider/data_source_datadog_{name}.go"
    with filename.open("w") as fp:
        fp.write(templates["datasource"].render(name=name, operations=data_source))
    if go_fmt:
        subprocess.call(["go", "fmt", filename])


def generate_resource(
    name: str, resource: dict, templates: dict[str, Template], go_fmt: bool
) -> None:
    """
    Generates files related to a resource.

    :param name: The name of the resource.
    :param operation: The intermediate representation of the resource.
    :param output: The root where the files will be generated.
    :param templates: The templates of the generated files.
    """
    # TF resource file
    output = pathlib.Path("../datadog/")
    filename = output / f"fwprovider/resource_datadog_{name}.go"
    with filename.open("w") as fp:
        fp.write(templates["base"].render(name=name, operations=resource))
    if go_fmt:
        subprocess.call(["go", "fmt", filename])

    # TF test file
    filename = output / "tests" / f"resource_datadog_{name}_test.go"
    with filename.open("w") as fp:
        fp.write(templates["test"].render(name=name, operations=resource))
    if go_fmt:
        subprocess.call(["go", "fmt", filename])

    dirname = output.parent / f"examples/resources/datadog_{name}"
    if not dirname.exists():
        os.makedirs(dirname)

    # TF resource example
    filename = dirname / "resource.tf"
    with filename.open("w") as fp:
        fp.write(templates["example"].render(name=name, operations=resource))

    # TF import example
    filename = dirname / "import.sh"
    with filename.open("w") as fp:
        fp.write(templates["import"].render(name=name, operations=resource))
