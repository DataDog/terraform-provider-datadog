resource "datadog_synthetics_private_location" "private_location" {
    name = "First private location"
    description = "Description of the private location"
    tags = ["foo:bar", "env:test"]
}
