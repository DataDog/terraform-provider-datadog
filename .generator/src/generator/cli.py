import pathlib
import click

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
@click.option(
    "-o",
    "--output",
    default="../datadog/",
    type=click.Path(path_type=pathlib.Path),
)
def cli(spec_path, config_path, output):
    """
    Generate a terraform code snippet from OpenAPI specification.
    """
    env = setup.load_environment(version=spec_path.parent.name)

    templates = setup.load_templates(env=env)

    spec = setup.load(spec_path)
    config = setup.load(config_path)

    operations = openapi.operations_to_generate(spec)

    for name, operations in operations.items():
        generate_resource(
            name=name,
            operations=operations,
            output=output,
            templates=templates,
        )


def generate_resource(
    name: str, operations: dict, output: pathlib.Path, templates: dict[str, Template]
) -> None:
    """
    Generates files related to a resource.

    :param name: The name of the resource.
    :param operation: The intermediate representation of the resource.
    :param output: The root where the files will be generated.
    :param templates: The templates of the generated files.
    """
    # TF resource file
    filename = output / f"fwprovider/resource_datadog_{name}.go"
    with filename.open("w") as fp:
        fp.write(templates["base"].render(name=name, operations=operations))

    # TF test file
    filename = output / "tests" / f"resource_datadog_{name}_test.go"
    with filename.open("w") as fp:
        fp.write(templates["test"].render(name=name, operations=operations))

    # TF resource example
    filename = output.parent / f"examples/resources/datadog_{name}/resource.tf"
    with filename.open("w") as fp:
        fp.write(templates["example"].render(name=name, operations=operations))

    # TF import example
    filename = output.parent / f"examples/resources/datadog_{name}/import.sh"
    with filename.open("w") as fp:
        fp.write(templates["import"].render(name=name, operations=operations))
