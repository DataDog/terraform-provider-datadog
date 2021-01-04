---
page_title: "datadog_ip_ranges Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about Datadog's IP addresses.
---

# Data Source `datadog_ip_ranges`

Use this data source to retrieve information about Datadog's IP addresses.

## Example Usage

```terraform
data "datadog_ip_ranges" "test" {}
```

## Schema

### Optional

- **id** (String) The ID of this resource.

### Read-only

- **agents_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the Agent endpoint.
- **agents_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the Agent endpoint.
- **api_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the API endpoint.
- **api_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the API endpoint.
- **apm_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the APM endpoint.
- **apm_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the APM endpoint.
- **logs_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the Logs endpoint.
- **logs_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the Logs endpoint.
- **process_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the Process endpoint.
- **process_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the Process endpoint.
- **synthetics_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the Synthetics endpoint.
- **synthetics_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the Synthetics endpoint.
- **webhooks_ipv4** (List of String) An Array of IPv4 addresses in CIDR format specifying the A records for the Webhooks endpoint.
- **webhooks_ipv6** (List of String) An Array of IPv6 addresses in CIDR format specifying the A records for the Webhooks endpoint.


