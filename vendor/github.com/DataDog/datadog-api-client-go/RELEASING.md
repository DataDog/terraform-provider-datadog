# Releasing

This document summarizes the process of doing a new release of this project.
Release can only be performed by Datadog maintainers of this repository.

## Schedule
This project does not have a strict release schedule. However, we would make a release at least every 2 months.
  - No release will be done if no changes got merged to the `master` branch during the above mentioned window.
  - Releases may be done more frequently than the above mentioned window.

### Prerequisites
- Install [datadog_checks_dev](https://datadog-checks-base.readthedocs.io/en/latest/datadog_checks_dev.cli.html#installation) using Python 3
- Have [Golang 1.13+](https://golang.org/doc/install) (since this process requires `go mod` command)
- Have the latest [godoc](https://github.com/golang/tools/tree/master/godoc) version.
    - *NOTE* With go 1.13, this isn't the binary included with Go and must be installed separately, otherwise it won't include module support.
- Ensure all CIs are passing on the master branch that we're about to release. 

## Release
Note that once the release process is started, nobody should be merging/pushing anything.

### Bumping Major Versions
When moving this package from one major semver version to the next, there's a couple extra steps needed:
1) Bump the `module` line in `go.mod`. E.g. from `module github.com/DataDog/datadog-api-client-go` to `module github.com/DataDog/datadog-api-client-go/v2`
2) Update all imports in `.go` files to utilize this new import path. 

### Commands

- See changes ready for release by running `ddev release show changes --tag-prefix "v" .` at the root of this project. Add any missing labels to PRs if needed.
- Run `ddev release changelog . <NEW_VERSION>` to update the `CHANGELOG.md` file at the root of this repository
- Run `go mod tidy` to clean up the dependencies defined in `go.mod` and `go.sum`
- Update the version in `version.go` you want to release, following semver. This file is used to know the version when sending telemetry.
- Commit the changes to the repository in a release branch and get it approved/merged after you:
    - Make sure that all CIs are passing, as this is the commit we will be releasing!
    - Check the built godoc looks OK by running `godoc -http=:<PORT_NUM>` and opening in your browser.
- Merge the above PR and create a release on the [releases page](https://github.com/DataDog/datadog-api-client-go/releases).
    - Specify the version you want to release, following semver.
    - Place the changelog contents into the description of the release.
    - Create/Publish the release, which will automatically create a tag on the `HEAD` commit.
- Bump the version again in `version.go` to start the new release cycle.

Check that the release is available by running:
`go get github.com/Datadog/datadog-api-client-go@<VERSION>`
where `VERSION` is the version that was just tagged (e.g. `v0.1.0`)
