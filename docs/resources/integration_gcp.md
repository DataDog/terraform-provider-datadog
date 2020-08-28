---
page_title: "datadog_integration_gcp"
---

# datadog_integration_gcp Resource

Provides a Datadog - Google Cloud Platform integration resource. This can be used to create and manage Datadog - Google Cloud Platform integration.

## Example Usage

```hcl
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

## Argument Reference

The following arguments are supported:

* `project_id`: (Required) Your Google Cloud project ID found in your JSON service account key.
* `private_key_id`: (Required) Your private key ID found in your JSON service account key.
* `private_key`: (Required) Your private key name found in your JSON service account key.
* `client_email`: (Required) Your email found in your JSON service account key.
* `client_id`: (Required) Your ID found in your JSON service account key.
* `host_filters`: (Optional) Limit the GCE instances that are pulled into Datadog by using tags. Only hosts that match one of the defined tags are imported into Datadog.

### See also
* [Google Cloud > Creating and Managing Service Account Keys](https://cloud.google.com/iam/docs/creating-managing-service-account-keys)
* [Datadog API Reference > Integrations > Google Cloud Platform](https://docs.datadoghq.com/api/v1/gcp-integration/)

## Attributes Reference

The following attributes are exported:

* `project_id` - Google Cloud project ID
* `client_email` - Google Cloud project service account email
* `host_filters` - Host filters

## Import

Google Cloud Platform integrations can be imported using their project ID, e.g.

```
$ terraform import datadog_integration_gcp.awesome_gcp_project_integration project_id
```
