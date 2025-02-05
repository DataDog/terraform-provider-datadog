# Terraform Generation

The goal of this sub-project is to generate the scaffolding to create a terraform resource.

> [!CAUTION]
> This code is HIGHLY experimental and should stabilize over the next weeks/months. As such this code is NOT intended for production uses.

## How to use

### Requirements

- This project
- Poetry
- An OpenApi 3.0.x specification (Datadog's OpenApi spec can be found [here](https://github.com/DataDog/datadog-api-client-go/tree/master/.generator/schemas))

### Install dependencies

Install the necessary dependencies by running `poetry install`

### Mark resources to be generated

For the generator to create a resource one must tag the resource's CRUD actions with `x-terraform-resource: <resource_name>` in the openApi specification.

```yaml
paths:
  /users:
    post:
      #  [...]
      x-terraform-resource: user
  /users/{user_id}:
    get:
      #  [...]
      x-terraform-resource: user
    patch:
      #  [...]
      x-terraform-resource: user
    delete:
      #  [...]
      x-terraform-resource: user
```

The routes tagged in this example will generate a `user` resource.

### Run the generator

When all resources to be generated are tagged run `poetry run python -m generator <openapi_spec_path>`.
the generated resources will be placed in `/datadog/fwprovider/`.
