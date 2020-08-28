# Terraform Provider

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.10.x
- [Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

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

If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-a-plugin) After placing it into your plugins directory, run `terraform init` to initialize it.

Further [usage documentation is available on the Terraform website](https://www.terraform.io/docs/providers/datadog/index.html).

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is _required_). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-datadog
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

Note that the above command runs acceptance tests by replaying pre-recorded API responses (cassettes) stored in `datadog/cassettes/`. When tests are modified, the cassettes need to be re-recorded.

In order to make tests cassette friendly, it's necessary to ensure that resources always get manipulated in a predictable order. When creating a testing Terraform config that defines multiple resources at the same time, you need to set inter-resource dependencies (using `depends_on`) in such a way that there is only one way for Terraform to manipulate them. For example, given resources A, B and C in the same config string, you can achieve this by making A depend on B and B depend on C. See [PR #442](https://github.com/DataDog/terraform-provider-datadog/pull/442) for an example of this.

_Note:_ Recording cassettes creates/updates/destroys real resources. Never run this on a production Datadog organization.

In order to re-record all cassettes you need to have `DD_API_KEY` and `DD_APP_KEY` for your testing organization in your environment. With that, run `make cassettes`. Do note that this would regenerate all cassettes and thus take a very long time; if you only need to re-record cassettes for one or two tests, you can run `make cassettes TESTARGS ="-run XXX"` - this will effectively execute `go test -run=XXX`, which would run all testcases that contain `XXX` in their name.

In order to run the full suite of Acceptance tests, run `make testacc`.

_Note:_ Acceptance tests create/update/destroy real resources. Never run this on a production Datadog organization.

```sh
$ make testacc
```

In order to update the underlying API Clients that are used by this provider to interact with the Datadog API, run:

```sh
API_CLIENT_VERSION=vx.y.z ZORKIAN_VERSION=vx.y.z make update-go-client
```

where:

- `API_CLIENT_VERSION` is the version or commit ref of the https://github.com/DataDog/datadog-api-client-go client.
- `ZORKIAN_VERSION` is the version or commit ref of the https://github.com/zorkian/go-datadog-api client.
