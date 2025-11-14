# ============================================================================
# Example 1: Basic Usage - Manage the order of tag pipeline rulesets
# ============================================================================
# This example shows the default behavior where UI-defined rulesets that are
# not in Terraform will be preserved at the end of the order.

resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "Standardize Environment Tags"
  enabled = true

  rules {
    name    = "map-env"
    enabled = true

    mapping {
      destination_key = "env"
      if_not_exists   = true
      source_keys     = ["environment", "stage"]
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "Assign Team Tags"
  enabled = true

  rules {
    name    = "assign-team"
    enabled = true

    query {
      query         = "service:web* OR service:api*"
      if_not_exists = false

      addition {
        key   = "team"
        value = "backend"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "Enrich Service Metadata"
  enabled = true

  rules {
    name    = "lookup-service"
    enabled = true

    reference_table {
      table_name         = "service_catalog"
      case_insensitivity = true
      if_not_exists      = true
      source_keys        = ["service"]

      field_pairs {
        input_column = "owner_team"
        output_key   = "owner"
      }
    }
  }
}

# Manage the order of tag pipeline rulesets
# Rulesets are executed in the order specified in ruleset_ids
# UI-defined rulesets not in this list will be preserved at the end
resource "datadog_tag_pipeline_rulesets" "order" {
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}

# ============================================================================
# Example 2: Override UI-defined rulesets (override_ui_defined_resources = true)
# ============================================================================
# When set to true, any rulesets created via the UI that are not defined in Terraform
# will be automatically deleted during terraform apply.

resource "datadog_tag_pipeline_ruleset" "managed_first" {
  name    = "Standardize Environment Tags"
  enabled = true

  rules {
    name    = "map-env"
    enabled = true

    mapping {
      destination_key = "env"
      if_not_exists   = true
      source_keys     = ["environment", "stage"]
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "managed_second" {
  name    = "Assign Team Tags"
  enabled = true

  rules {
    name    = "assign-team"
    enabled = true

    query {
      query         = "service:web*"
      if_not_exists = false

      addition {
        key   = "team"
        value = "frontend"
      }
    }
  }
}

# Manage order with override_ui_defined_resources = true
# This will delete any rulesets created via the UI that are not in this list
resource "datadog_tag_pipeline_rulesets" "order_override" {
  override_ui_defined_resources = true

  ruleset_ids = [
    datadog_tag_pipeline_ruleset.managed_first.id,
    datadog_tag_pipeline_ruleset.managed_second.id
  ]
}

# ============================================================================
# Example 3: Preserve UI-defined rulesets (override_ui_defined_resources = false)
# ============================================================================
# When set to false (default), UI-defined rulesets that are not in Terraform
# will be preserved at the end of the order. However, if unmanaged rulesets
# are in the middle of the order, Terraform will error and require you to either:
# 1. Import the unmanaged rulesets
# 2. Set override_ui_defined_resources = true
# 3. Manually reorder or delete them in the Datadog UI

resource "datadog_tag_pipeline_ruleset" "preserve_first" {
  name    = "Standardize Environment Tags"
  enabled = true

  rules {
    name    = "map-env"
    enabled = true

    mapping {
      destination_key = "env"
      if_not_exists   = true
      source_keys     = ["environment", "stage"]
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "preserve_second" {
  name    = "Assign Team Tags"
  enabled = true

  rules {
    name    = "assign-team"
    enabled = true

    query {
      query         = "service:web*"
      if_not_exists = false

      addition {
        key   = "team"
        value = "frontend"
      }
    }
  }
}

# Manage order with override_ui_defined_resources = false (default)
# UI-defined rulesets will be preserved at the end of the order
# Terraform will warn if unmanaged rulesets exist at the end
# Terraform will error if unmanaged rulesets are in the middle
resource "datadog_tag_pipeline_rulesets" "order_preserve" {
  override_ui_defined_resources = false

  ruleset_ids = [
    datadog_tag_pipeline_ruleset.preserve_first.id,
    datadog_tag_pipeline_ruleset.preserve_second.id
  ]
}