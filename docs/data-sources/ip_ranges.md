---
page_title: "datadog_ip_ranges"
---

# datadog_ip_ranges Data Source

Use this data source to retrieve information about Datadog's IP addresses.

## Example Usage

```
data "datadog_ip_ranges" "test" {}
```

## Attributes Reference

- `agents_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the agent endpoint.
- `api_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the api endpoint.
- `apm_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the apm endpoint.
- `logs_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the logs endpoint.
- `process_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the process endpoint.
- `synthetics_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the synthetics endpoint.
- `webhooks_ipv4`: An Array of IPv4 addresses in CIDR format specifying the A records for the webhooks endpoint.
- `agents_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the agent endpoint.
- `api_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the api endpoint.
- `apm_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the apm endpoint.
- `logs_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the logs endpoint.
- `process_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the process endpoint.
- `synthetics_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the synthetics endpoint.
- `webhooks_ipv6`: An Array of IPv6 addresses in CIDR format specifying the A records for the webhooks endpoint.
