---
subcategory: ""
page_title: "Terraform Dynamic Provider Credentials"
description: |-
  Authenticate HCP Terraform runs with Workload Identity Federation.
---

# Terraform Dynamic Provider Credentials

Use Terraform-issued workload identity tokens (WITs) to authenticate the Datadog provider without storing Datadog credentials in your workspace. The provider exchanges a WIT for a short-lived Datadog access token through Workload Identity Federation (WIF).

## Prerequisites

- Use a Datadog provider release that includes Terraform Dynamic Provider Credentials support (version >= 4.25.0).
- Enable the applicable WIF integration for your Datadog organization and register the identity mapping described below. Creating the Terraform audience variable alone is insufficient.
- Run in HCP Terraform. Local Terraform CLI execution does not generate these workload tokens.
- For self-hosted HCP Terraform agents, tagged workload tokens require agent v1.12.0 or later; the untagged token requires v1.7.0 or later. Check [HashiCorp's requirements](https://developer.hashicorp.com/terraform/cloud-docs/dynamic-provider-credentials/manual-generation) for your deployment.

## Register a WIF persona mapping

Create a Terraform WIF identity mapping in the Datadog organization whose resources you intend to manage. Map the permitted Terraform workload identity to an active Datadog service account or user. The resulting Datadog token receives that identity's permissions, so grant the permissions required by your Terraform resources. A service account is recommended for automation.

For HCP Terraform workspace runs, the subject has this form:

```text
organization:<terraform-organization>:project:<project>:workspace:<workspace>:run_phase:(plan|apply)
```

For example, a mapping for both phases of a specific workspace can use:

```text
organization:example:project:Default Project:workspace:production:run_phase:*
```

Use the exact Terraform organization name. HCP mappings support exact segment values or whole-segment `*` for project, workspace, and phase; the organization must be concrete. Include every phase that needs to access Datadog, since Terraform issues a separate token for plan and apply. The Terraform organization name is different from the Datadog organization UUID used in the audience.

HCP Terraform's US and EU issuers use the dedicated Terraform integration:

- `https://app.terraform.io`
- `https://app.eu.terraform.io`

## Configure the Terraform workspace

Add an **environment variable** to the workspace, or a variable set attached to it:

| Name | Value |
| --- | --- |
| `TFC_WORKLOAD_IDENTITY_AUDIENCE_DATADOG` | `datadog/<your-datadog-org-uuid>` |

Use the audience supplied by the Datadog mapping setup. HCP Terraform injects the corresponding JWT as `TFC_WORKLOAD_IDENTITY_TOKEN_DATADOG` for each run phase. Do not populate the token variable yourself or place the token in Terraform configuration.

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

Use the appropriate URL for your [Datadog site](https://docs.datadoghq.com/getting_started/site/). HCP's US/EU region does not select the Datadog site. The WIT audience supplies the Datadog organization UUID. Setting `org_uuid` is optional. If `org_uuid`, `DD_ORG_UUID`, or `DATADOG_ORG_UUID` is configured, it must match the organization UUID in the token audience. A mismatch fails before token exchange, even when `validate = false`; the provider does not switch to another token or authentication method.

The provider also supports the untagged `TFC_WORKLOAD_IDENTITY_AUDIENCE` / `TFC_WORKLOAD_IDENTITY_TOKEN` pair. Use the `DATADOG` tag to avoid sharing the untagged token with another integration. Other tags are not discovered automatically.

## Authentication precedence and failures

The provider selects authentication in this order:

1. `TFC_WORKLOAD_IDENTITY_TOKEN_DATADOG`, if addressed to an eligible Datadog organization.
2. `TFC_WORKLOAD_IDENTITY_TOKEN`, with the same audience check.
3. AWS WIF, when `cloud_provider_type = "aws"` is configured.
4. A configured bearer token.
5. Datadog API/application keys.

A candidate must be a readable JWT with an unambiguous audience of the form `datadog/<valid-org-uuid>`. The JWT string and array forms of `aud` are supported. Missing, unreadable, or unrelated candidates are skipped. A missing or malformed issuer does not cause credential fallback. Client-side inspection is only a selection check; ETS validates signatures, issuer trust, token expiration, and persona mappings.

Once a WIT is selected, an exchange rejection is an error. The provider does not try the other token or switch to AWS/static credentials. Check the persona mapping, permitted subject/run phase, issuer registration, audience, and Datadog site when troubleshooting.

API calls and access-token refresh use the same delegated-authentication machinery as AWS WIF. Refresh exchanges the selected WIT again; if the WIT has expired or the exchange is rejected, the operation fails rather than switching identities.

The legacy `datadog_integration_pagerduty` resource uses a separate API client that does not support delegated authentication. It has the same limitation with AWS WIF and cannot be managed with WIT credentials alone.

## Verify the setup

### Check the authenticated identity during plan

1. Use a dedicated test workspace with no managed resources. Keep the provider configuration above.
2. Confirm that the persona mapping permits this workspace's plan phase and that `TFC_WORKLOAD_IDENTITY_AUDIENCE_DATADOG` contains the audience returned by the mapping setup.
3. Remove alternate authentication from the provider configuration, workspace environment, inherited variable sets, and agent environment. This includes `api_key`, `app_key`, `bearer_token`, and `cloud_provider_type`, plus their `DD_*` and `DATADOG_*` environment-variable equivalents. Leave provider validation enabled.
4. Add the following configuration. Set `expected_datadog_user_id` and `expected_datadog_org_uuid` as **Terraform variables** in the workspace, using the mapped user or service account's UUID and the Datadog organization UUID.

```terraform
variable "expected_datadog_user_id" {
  type        = string
  description = "UUID of the Datadog user or service account in the persona mapping."
}

variable "expected_datadog_org_uuid" {
  type        = string
  description = "UUID of the Datadog organization containing the persona mapping."
}

data "datadog_current_user" "wif" {
  lifecycle {
    postcondition {
      condition = (
        self.id == var.expected_datadog_user_id &&
        self.org_id == var.expected_datadog_org_uuid
      )
      error_message = "The authenticated Datadog identity or organization does not match the persona mapping."
    }
  }
}

output "datadog_wif_identity" {
  value = {
    user_id         = data.datadog_current_user.wif.id
    handle          = data.datadog_current_user.wif.handle
    service_account = data.datadog_current_user.wif.service_account
    org_uuid        = data.datadog_current_user.wif.org_id
  }
}
```

Queue a new plan in HCP Terraform. In the run output, verify that:

- `datadog_current_user.wif` completes its read and the postcondition passes.
- `datadog_wif_identity` contains the expected user or service account UUID and organization UUID.
- There are no managed-resource additions, updates, or deletions. Terraform may show the new output as a change; that is expected.

This checks authentication with the plan-phase WIT and a Datadog API read. It does not establish permission to manage every resource type. Do not print the WIT or exchanged access token to inspect the result.

### Check the apply phase

Terraform issues a separate WIT for apply. A data source already read during plan may not be read again during apply, so applying the unchanged example alone does not prove the apply-phase API call works.

In the same dedicated test workspace, add this built-in Terraform resource:

```terraform
resource "terraform_data" "wif_apply_check" {
  input = "first-check"
}
```

Add `depends_on = [terraform_data.wif_apply_check]` inside the existing `data "datadog_current_user" "wif"` block, keeping its postcondition. The dependency deliberately defers the identity read until apply when the marker has a pending change. The marker stores only a value in Terraform state and creates no Datadog resource. See HashiCorp's [data source dependencies](https://developer.hashicorp.com/terraform/language/data-sources#dependencies) and [terraform_data reference](https://developer.hashicorp.com/terraform/language/resources/terraform-data).

1. Confirm the persona mapping permits the apply phase as well as plan.
2. Queue a normal plan-and-apply run. A speculative or plan-only run cannot perform this check.
3. Before applying, verify that the only resource addition is `terraform_data.wif_apply_check` and that `datadog_current_user.wif` says it will be read during apply.
4. Apply the run. Confirm that the identity read and postcondition pass and that the output contains the expected identity and organization.
5. To repeat the check, change the marker's `input` value and confirm the next plan defers the read again. For cleanup, remove the marker and the `depends_on` line, then plan and apply its removal.
