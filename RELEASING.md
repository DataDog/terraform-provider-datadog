# Releasing

This document summarizes the process of doing a new release of this project.
Release can only be performed by Datadog maintainers of this repository.

## Schedule

This project does not have a strict release schedule. However, we would make a release at least every 2 months.
- No release will be done if no changes got merged to the `master` branch during the above mentioned window.
- Releases may be done more frequently than the above mentioned window.

### Prerequisites

- Ensure all CIs are passing on the master branch that we're about to release.

## Release Process

The release process is controlled and run by GitHub Actions.

### Prerequisite

1. Make sure you have `write_repo` access.
1. Share your plan for the release with other maintainers to avoid conflicts during the release process.

### Update Changelog

1. Open [prepare release](https://github.com/DataDog/terraform-provider-datadog/actions/workflows/prepare_release.yml) workflow and click on `Run workflow` dropdown.
2. Enter new version identifier in the `New version number` input box (e.g. `3.11.0`).
3. Trigger the action by clicking on `Run workflow` button.

### Review

1. Review the generated pull-request for `release/<New version tag>` branch.
2. If everything is fine, merge the pull-request.
3. Check that the [release](https://github.com/DataDog/terraform-provider-datadog/actions/workflows/release.yml) action created new draft [release](https://github.com/DataDog/terraform-provider-datadog/releases) on GitHub.
4. Review and publish the draft release by clicking `Publish Release``.

Hashicorp machinery will take over and sync the release to the Terraform Registry. Check that the release is available on the [Terraform Registry](https://registry.terraform.io/providers/DataDog/datadog):
where `VERSION` is the version that was just tagged (e.g. `3.11.0`)