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

- **id** (String, Optional) The ID of this resource.

### Read-only

- **agents_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the agent endpoint.
- **agents_ipv6** (List of String, Read-only) An Array of IPv6 addresses in CIDR format specifying the A records for the agent endpoint.
- **api_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the api endpoint.
- **api_ipv6** (List of String, Read-only) An Array of IPv6 addresses in CIDR format specifying the A records for the api endpoint.
- **apm_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the apm endpoint.
- **apm_ipv6** (List of String, Read-only) An Array of IPv6 addresses in CIDR format specifying the A records for the apm endpoint.
- **logs_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the logs endpoint.
- **logs_ipv6** (List of String, Read-only) An Array of IPv6 addresses in CIDR format specifying the A records for the logs endpoint.
- **process_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the process endpoint.
- **process_ipv6** (List of String, Read-only) An Array of IPv6 addresses in CIDR format specifying the A records for the process endpoint.
- **synthetics_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the synthetics endpoint
- **synthetics_ipv6** (List of String, Read-only) An Array of IPv6 addresses in CIDR format specifying the A records for the synthetics endpoint.
- **webhooks_ipv4** (List of String, Read-only) An Array of IPv4 addresses in CIDR format specifying the A records for the webhooks endpoint.
- **webhooks_ipv6** (List of String, Read-only) n Array of IPv6 addresses in CIDR format specifying the A records for the webhooks endpoint.


