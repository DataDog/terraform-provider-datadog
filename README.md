# Terraform Provider

-   Website: https://www.terraform.io
-   [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
-   Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

-   [Terraform](https://www.terraform.io/downloads.html) 0.12.x

## Building The Provider

Clone repository to: `$GOPATH/src/github.com/DataDog/terraform-provider-datadog`

```sh
$ mkdir -p $GOPATH/src/github.com/DataDog; cd $GOPATH/src/github.com/DataDog
$ git clone git@github.com:DataDog/terraform-provider-datadog
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/DataDog/terraform-provider-datadog
$ make build
```

**Note**: For contributions created from forks, the repository should still be cloned under the `$GOPATH/src/github.com/DataDog/terraform-provider-datadog` directory to allow the provided `make` commands to properly run, build, and test this project.

## Using the provider

If you're building the provider, follow the instructions to [install it as a plugin.](./DEVELOPMENT.md) After placing it into your plugins directory, run `terraform init` to initialize it.

Further [usage documentation is available on the Terraform website](https://www.terraform.io/docs/providers/datadog/index.html).
