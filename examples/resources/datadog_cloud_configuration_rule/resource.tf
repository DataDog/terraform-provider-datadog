resource "datadog_cloud_configuration_rule" "myrule" {
  name                   = "My cloud configuration rule"
  message                = "Rule has triggered"
  enabled                = true
  policy                 = <<-EOT
        package datadog

        import data.datadog.output as dd_output

        import future.keywords.contains
        import future.keywords.if
        import future.keywords.in

        eval(resource) = "skip" if {
            # Logic that evaluates to true if the resource should be skipped
            true
        } else = "pass" {
            # Logic that evaluates to true if the resource is compliant
            true
        } else = "fail" {
            # Logic that evaluates to true if the resource is not compliant
            true
        }

        # This part remains unchanged for all rules
        results contains result if {
            some resource in input.resources[input.main_resource_type]
            result := dd_output.format(resource, eval(resource))
        }
    EOT
  resource_type          = "aws_s3_bucket"
  related_resource_types = []
  severity               = "high"
  group_by               = ["@resource"]
  notifications          = ["@channel"]
  tags                   = ["some:tag"]
}
