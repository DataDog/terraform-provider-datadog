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

### Removed import support for `datadog_application_key`

Import functionality has been removed for the `datadog_application_key` resource. This was previously deprecated with a warning.

Application keys contain sensitive credentials that cannot be retrieved after creation. When you import an existing application key, the `key` attribute cannot be populated from the API, which leads to state inconsistencies and potential security issues.

**Note:** If your organization has [One-Time Read mode](https://docs.datadoghq.com/account_management/api-app-keys/#one-time-read-mode) enabled for Application Keys, then no action is needed to migrate for this resource because import is already unavailable.

**Before (v3.x):**

```shell
# Import command
terraform import datadog_application_key.foo 11111111-2222-3333-4444-555555555555
```

```terraform
# Import block (Terraform 1.5+)
import {
  to = datadog_application_key.foo
  id = "11111111-2222-3333-4444-555555555555"
}
```

**After (v4.0.0):**

Import is no longer supported. Attempting to import will result in:

```
Error: Resource Import Not Implemented

This resource does not support import.
```

**Migration steps:**

1. If you have `import` blocks for `datadog_application_key` resources, ensure they have been applied before upgrading, then remove the import blocks from your configuration.
2. Previously imported application keys continue to work after upgrading. No action is required for keys already in your Terraform state.
3. For new application keys, use the `datadog_application_key` resource to create them directly and securely store the key values using a secret management system.

### Removed `locked` on `datadog_monitor`

Removed `locked` and changed the default behavior of `restricted_roles` on `datadog_monitor`. These changes are intended 
to encourage users to migrate and manage monitor permissions through the `datadog_restriction_policy` resource. 

**Note:** Migrating off `restricted_roles` is not required. This field is still supported by the monitor provider. However, we
strongly recommend migrating to `datadog_restriction_policy` as the preferred way to manage monitor permissions going forward.

**Before (v3.x):**

```terraform
# Old configuration
# Monitor with `locked`
resource "datadog_monitor" "foo" {
  name               = "Name for monitor foo"
  type               = "metric alert"
  message            = "Monitor triggered. Notify: @hipchat-channel"
  escalation_message = "Escalation message @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
  monitor_thresholds {
    critical = 4
  }
  locked = true
}

# Monitor with `restricted_roles`
resource "datadog_monitor" "foo" {
  name               = "Name for monitor foo"
  type               = "metric alert"
  message            = "Monitor triggered. Notify: @hipchat-channel"
  escalation_message = "Escalation message @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
  monitor_thresholds {
    critical = 4
  }
  restricted_roles = ["aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"]
}
```

**After (v4.0.0):**

```terraform
# New configuration
resource "datadog_monitor" "foo" {
  name               = "Name for monitor foo"
  type               = "metric alert"
  message            = "Monitor triggered. Notify: @hipchat-channel"
  escalation_message = "Escalation message @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
  monitor_thresholds {
    critical = 4
  }
}

resource "datadog_restriction_policy" "bar" {
  resource_id = "monitor:${datadog_monitor.foo.id}"
  bindings {
    principals = ["role:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"]
    relation   = "editor"
  }
}
```

**Migration steps:**
1. Remove the `locked` or `restricted_roles` field from monitor resources.
2. Create a `datadog_restriction_policy` the associated monitor ID and the roles you want to restrict.

## Getting Help

If you encounter issues upgrading to v4.0.0:

- Check the [Datadog Terraform Provider documentation](https://registry.terraform.io/providers/DataDog/datadog/latest/docs)
- Open an issue on [GitHub](https://github.com/DataDog/terraform-provider-datadog/issues)
- Contact [Datadog Support](https://docs.datadoghq.com/help/)
