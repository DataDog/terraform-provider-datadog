---
subcategory: ""
page_title: "Datadog Provider v4.0.0 Upgrade Guide"
description: |-
    Migration guide for upgrading from Datadog Provider v3.x to v4.0.0
---

# Datadog Provider v4.0.0 Upgrade Guide

Version 4.0.0 of the Datadog Provider for Terraform is a major release that includes breaking changes. This guide covers the changes and steps you need to follow to upgrade from v3.x to v4.0.0.

## Changes in v4.0.0

### Terraform Version Requirement

**Terraform 1.1.5 or later is now required.**

The provider has migrated from Terraform Plugin Protocol Version 5 to Protocol Version 6, which requires Terraform 1.1.5+. Users on earlier versions should either:

- Upgrade Terraform to 1.1.5 or later (recommended)
- Pin to v3.x of the provider: `version = "~> 3.0"`

**Staying on v3.x:**

```terraform
terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
      version = "~> 3.0"
    }
  }
}
```

**Upgrading to v4.0.0:**

```terraform
terraform {
  required_providers {
    datadog = {
      source  = "DataDog/datadog"
      version = "~> 4.0"
    }
  }
  required_version = ">= 1.1.5"
}
```

**Migration steps:**
Follow the official Terraform documentation depending on your version to upgrade to at least 1.1.5.

<!--
================================================================================
PLACEHOLDER: Resource-Specific Breaking Changes
================================================================================

Other teams should add their breaking changes below this comment. For each
resource with breaking changes, add a section following this format:

### <Name of changes> datadog_<resource_name> <resource_type>

Brief description of the breaking change.

**Before (v3.x):**

```terraform
# Old configuration
```

**After (v4.0.0):**

```terraform
# New configuration
```

**Migration steps:**
1. Step one
2. Step two

================================================================================
-->

## Getting Help

If you encounter issues upgrading to v4.0.0:

- Check the [Datadog Terraform Provider documentation](https://registry.terraform.io/providers/DataDog/datadog/latest/docs)
- Open an issue on [GitHub](https://github.com/DataDog/terraform-provider-datadog/issues)
- Contact [Datadog Support](https://docs.datadoghq.com/help/)
