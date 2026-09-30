---
subcategory: ""
page_title: "Terraform Dynamic Provider Credentials"
description: |-
  Authenticate HCP Terraform and Terraform Enterprise runs with Workload Identity Federation.
---

# Terraform Dynamic Provider Credentials

Use Terraform-issued workload identity tokens (WITs) to authenticate the Datadog provider without storing Datadog API or application keys in your workspace. The provider exchanges a WIT for a short-lived Datadog access token through Workload Identity Federation (WIF).

## Prerequisites

- Use a Datadog provider release that includes Terraform Dynamic Provider Credentials support.
- Enable the applicable WIF integration for your Datadog organization and register the identity mapping described below. Creating the Terraform audience variable alone is insufficient.
- Run in HCP Terraform or Terraform Enterprise (TFE). Local Terraform CLI execution does not generate these workload tokens.
- For self-hosted HCP Terraform agents, tagged workload tokens require agent v1.12.0 or later; the untagged token requires v1.7.0 or later. Check [HashiCorp's requirements](https://developer.hashicorp.com/terraform/cloud-docs/dynamic-provider-credentials/manual-generation) for your deployment.

## Register a WIF persona mapping

Create a Terraform WIF identity (persona) mapping in the Datadog organization whose resources you intend to manage. Map the permitted Terraform workload identity to an active Datadog user or service account. The resulting Datadog token receives that identity's permissions, so grant the permissions required by your Terraform resources. A service account is recommended for automation.

For HCP Terraform workspace runs, the subject has this form:

```text
organization:<terraform-organization>:project:<project>:workspace:<workspace>:run_phase:<plan-or-apply>
```

For example, a mapping for both phases of a specific workspace can use:

```text
organization:example:project:Default Project:workspace:production:run_phase:*
```

Use the exact Terraform organization name. HCP mappings support exact segment values or whole-segment `*` for project, workspace, and phase; the organization must be concrete. Include every phase that needs to access Datadog, since Terraform issues a separate token for plan and apply. The Terraform organization name is different from the Datadog organization UUID used in the audience.

HCP Terraform's US and EU issuers use the dedicated Terraform integration:

- `https://app.terraform.io`
- `https://app.eu.terraform.io`

For self-hosted TFE, first register your TFE issuer with Datadog's CustomOIDC integration, configure organization resolution from the `datadog/<org-uuid>` audience, and create its persona mapping to the intended Datadog identity. The registered issuer/key source must satisfy the reachability and trust requirements of your WIF configuration. The provider routes these tokens through CustomOIDC; it does not create issuer registrations or persona mappings.

## Configure the Terraform workspace

Add an **environment variable** to the workspace, or a variable set attached to it:

| Name | Value |
| --- | --- |
| `TFC_WORKLOAD_IDENTITY_AUDIENCE_DATADOG` | `datadog/<your-datadog-org-uuid>` |

Use the audience supplied by the Datadog mapping setup. HCP Terraform/TFE injects the corresponding JWT as `TFC_WORKLOAD_IDENTITY_TOKEN_DATADOG` for each run phase. Do not populate the token variable yourself or place the token in Terraform configuration.

No additional authentication setting is required in the Datadog provider block:

```terraform
terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

provider "datadog" {
  # Set api_url for the Datadog site containing the mapping and resources.
  api_url = "https://api.datadoghq.com/"
}
```

Use the appropriate URL for your [Datadog site](https://docs.datadoghq.com/getting_started/site/). HCP's US/EU region does not select the Datadog site. The WIT audience supplies the Datadog organization UUID. `org_uuid` and `DD_ORG_UUID` are not used to select or filter WIT credentials.

The provider also supports the untagged `TFC_WORKLOAD_IDENTITY_AUDIENCE` / `TFC_WORKLOAD_IDENTITY_TOKEN` pair. Use the `DATADOG` tag to avoid sharing the untagged token with another integration. Other tags are not discovered automatically.

## Authentication precedence and failures

The provider selects authentication in this order:

1. `TFC_WORKLOAD_IDENTITY_TOKEN_DATADOG`, if addressed to an eligible Datadog organization.
2. `TFC_WORKLOAD_IDENTITY_TOKEN`, with the same audience check.
3. AWS WIF, when `cloud_provider_type = "aws"` is configured.
4. A configured bearer token.
5. Datadog API/application keys.

A candidate must be a readable JWT with an unambiguous audience of the form `datadog/<valid-org-uuid>`. The JWT string and array forms of `aud` are supported. Missing, unreadable, or unrelated candidates are skipped. The issuer is inspected only to choose the HCP or CustomOIDC exchange path; a missing or malformed issuer does not cause credential fallback. Client-side inspection is only a selection check; ETS validates signatures, issuer trust, token expiration, and persona mappings.

Once a WIT is selected, an exchange rejection is an error. The provider does not try the other token or switch to AWS/static credentials. Check the persona mapping, permitted subject/run phase, issuer registration, audience, and Datadog site when troubleshooting.

API calls and access-token refresh use the same delegated-authentication machinery as AWS WIF. Refresh exchanges the selected WIT again; if the WIT has expired or the exchange is rejected, the operation fails rather than switching identities.

The legacy `datadog_integration_pagerduty` resource uses a separate API client that does not support delegated authentication. It has the same limitation with AWS WIF and cannot be managed with WIT credentials alone.

## Verify the setup

Run a plan and apply against a test resource covered by the mapped identity's permissions. Remove alternate Datadog credentials and AWS WIF configuration from the verification workspace so success proves the WIT path. Verify the target Datadog identity through the WIF authentication/audit evidence available in your organization.

Also verify that a run excluded by the persona mapping fails authentication, and that removing the WIT permits a separately configured existing authentication method to work. An ETS rejection must still fail when alternate credentials are configured. If no authentication method is available, the provider reports a missing-credentials error.
