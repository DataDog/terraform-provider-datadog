resource "datadog_ip_allowlist" "example" {
  enabled = false

  entry {
    cidr_block = "172.0.0.1/32"
    note       = "1st Example IP Range"
  }

  entry {
    cidr_block = "127.0.0.0/32"
    note       = "2nd Example IP Range"
  }
}