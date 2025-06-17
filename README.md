# Datadog Terraform Provider

- Website: https://registry.terraform.io/providers/DataDog/datadog/latest
- Documentation: https://registry.terraform.io/providers/DataDog/datadog/latest/docs
- Terraform website: https://www.terraform.io
- [Support](https://help.datadoghq.com/hc/en-us/requests/new?_gl=1*rmfzc4*_gcl_au*OTc0MzI0MjMyLjE3NDM2Nzc1MjQ.*_ga*MjI2ODYyNDMxLjE3NDY0NDMyNjU.*_ga_KN80RDFSQK*czE3NDcxMjUxMjgkbzIkZzAkdDE3NDcxMjUxMjgkajAkbDAkaDE3NDY0ODczNDk.*_fplc*dXIzUEdsS2htcE1kY0ZGZGtIYSUyQlFFVjJJRmFWaTFVYzlZUUtoSmoxMW5NNFlXbWppdzZORUhVcHJQdDFXZ2k5bHFrNEJneWV1bW1YRVNBdno5dVZFaERoZDclMkZRbUY4R0FVNm1hSUJ6UzZoUWJuOEJJY0lNZUo4WWpIdEh6dyUzRCUzRA..)

## Requirements

-   [Terraform](https://www.terraform.io/downloads.html) >= 0.12.x

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
