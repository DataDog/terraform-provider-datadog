---
page_title: "datadog_integration_gcp Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - Google Cloud Platform integration resource. This can be used to create and manage Datadog - Google Cloud Platform integration.
---

# Resource `datadog_integration_gcp`

Provides a Datadog - Google Cloud Platform integration resource. This can be used to create and manage Datadog - Google Cloud Platform integration.

## Example Usage

```terraform
# Create a new Datadog - Google Cloud Platform integration
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "awesome-project-id"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  host_filters   = "foo:bar,buzz:lightyear"
}
```

## Schema

### Required

- **client_email** (String, Required)
- **client_id** (String, Required)
- **private_key** (String, Required)
- **private_key_id** (String, Required)
- **project_id** (String, Required)

### Optional

- **host_filters** (String, Optional)
- **id** (String, Optional) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Google Cloud Platform integrations can be imported using their project ID, e.g.

terraform import datadog_integration_gcp.awesome_gcp_project_integration project_id
```
