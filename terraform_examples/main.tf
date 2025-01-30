# terraform_examples/main.tf
terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

variable "datadog_api_key" {
  type        = string
  description = "TFC Organization"
}

variable "datadog_app_key" {
  type        = string
  description = "TFC Workspace"
}

# Configure the Datadog provider
provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}

resource "datadog_monitor" "foo" {
  name = "test klara [terraform]"
  type = "cost alert"
  message = "this is a test"

  query = "formula(\"exclude_null(query1)\").last(\"7d\").anomaly(direction=\"above\", threshold=10) >= 6"

  variables {
  	cloud_cost_query {
		data_source = "cloud_cost"
		name        = "query1"
		query 	    = "sum:aws.cost.net.amortized.shared.resources.allocated{aws_product IN (amplify ,athena, backup, bedrock ) } by {aws_product}.rollup(sum, 86400)"
		aggregator  = "sum"
	}
  }

  monitor_thresholds {
      critical = 6
  }
}
