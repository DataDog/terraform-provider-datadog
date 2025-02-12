# Terraform Generation

The goal of this sub-project is to generate the scaffolding to create a Terraform resource.

> [!CAUTION]
> This code is HIGHLY experimental and should stabilize over the next weeks/months. As such this code is NOT intended for production uses.

## How to use

### Requirements

- This project
- Poetry
- An OpenApi 3.0.x specification (Datadog's OpenApi spec can be found [here](https://github.com/DataDog/datadog-api-client-go/tree/master/.generator/schemas))

### Install dependencies

Install the necessary dependencies by running `poetry install`

### Marking the resources to be generated

The generator reads a configuration file in order to generate the appropriate resources.
The configuration file should look like the following:

```yaml
resources:
  { resource_name }:
    read:
      method: { read_method }
      path: { read_path }
    create:
      method: { create_method }
      path: { create_path }
    update:
      method: { update_method }
      path: { update_path }
    delete:
      method: { delete_method }
      path: { delete_path }
  ...
```

- `resource_name` is the name of the resource to be generated.
- `xxx_method` should be the HTTP method used by the relevant route
- `xxx_path` should be the HTTP route of the resource's CRUD operation

> [!NOTE]
> An example using the `team` resource would look like this:
>
> ```yaml
> resources:
>   team:
>     read:
>       method: get
>       path: /api/v2/team/{team_id}
>     create:
>       method: post
>       path: /api/v2/team
>     update:
>       method: patch
>       path: /api/v2/team/{team_id}
>     delete:
>       method: delete
>       path: /api/v2/team/{team_id}
> ```

### Running the generator

Once the configuration file is written, you can run the following command to generate the Terraform resources:

```sh
  $ poetry run python -m generator <openapi_spec_path> <configuration_path>
```

> [!NOTE]
> The generated resources will be placed in `datadog/fwprovider/`
