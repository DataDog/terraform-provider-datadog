resource "datadog_tag_pipeline_ruleset" "example" {
  name    = "Complete Tag Pipeline Example"
  enabled = true

  rules {
    name    = "standardize-environment"
    enabled = true

    mapping {
      destination_key = "env"
      if_tag_exists   = "replace"
      source_keys     = ["environment", "stage", "tier"]
    }
  }

  rules {
    name    = "assign-team-tags"
    enabled = true

    query {
      query              = "service:web* OR service:frontend*"
      case_insensitivity = true
      if_tag_exists      = "append"

      addition {
        key   = "team"
        value = "frontend"
      }
    }
  }

  rules {
    name    = "enrich-service-metadata"
    enabled = true

    reference_table {
      table_name         = "service_catalog"
      case_insensitivity = true
      if_tag_exists      = "append"
      source_keys        = ["service"]

      field_pairs {
        input_column = "owner_team"
        output_key   = "owner"
      }

      field_pairs {
        input_column = "business_unit"
        output_key   = "business_unit"
      }
    }
  }
}
