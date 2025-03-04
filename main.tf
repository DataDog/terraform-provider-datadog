terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

provider "datadog" {
  api_key = "<FILL_ME_IN>"
  app_key = "<FILL_ME_IN>"
  api_url = "https://dd.datad0g.com/"
}

