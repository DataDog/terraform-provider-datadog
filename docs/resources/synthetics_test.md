---
page_title: "datadog_synthetics_test Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.
---

# Resource `datadog_synthetics_test`

Provides a Datadog synthetics test resource. This can be used to create and manage Datadog synthetics test.



## Schema

### Required

- **locations** (List of String, Required)
- **name** (String, Required)
- **request** (Map of String, Required)
- **status** (String, Required)
- **type** (String, Required)

### Optional

- **assertion** (Block List) (see [below for nested schema](#nestedblock--assertion))
- **assertions** (List of Map of String, Optional, Deprecated)
- **device_ids** (List of String, Optional)
- **id** (String, Optional) The ID of this resource.
- **message** (String, Optional)
- **options** (Map of String, Optional, Deprecated)
- **options_list** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list))
- **request_basicauth** (Block List, Max: 1) (see [below for nested schema](#nestedblock--request_basicauth))
- **request_client_certificate** (Block List, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate))
- **request_headers** (Map of String, Optional)
- **request_query** (Map of String, Optional)
- **step** (Block List) (see [below for nested schema](#nestedblock--step))
- **subtype** (String, Optional)
- **tags** (List of String, Optional)
- **variable** (Block List) (see [below for nested schema](#nestedblock--variable))

### Read-only

- **monitor_id** (Number, Read-only)

<a id="nestedblock--assertion"></a>
### Nested Schema for `assertion`

Required:

- **operator** (String, Required)
- **type** (String, Required)

Optional:

- **property** (String, Optional)
- **target** (String, Optional)
- **targetjsonpath** (Block List, Max: 1) (see [below for nested schema](#nestedblock--assertion--targetjsonpath))

<a id="nestedblock--assertion--targetjsonpath"></a>
### Nested Schema for `assertion.targetjsonpath`

Required:

- **jsonpath** (String, Required)
- **operator** (String, Required)
- **targetvalue** (String, Required)



<a id="nestedblock--options_list"></a>
### Nested Schema for `options_list`

Optional:

- **accept_self_signed** (Boolean, Optional)
- **allow_insecure** (Boolean, Optional)
- **follow_redirects** (Boolean, Optional)
- **min_failure_duration** (Number, Optional)
- **min_location_failed** (Number, Optional)
- **monitor_options** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list--monitor_options))
- **retry** (Block List, Max: 1) (see [below for nested schema](#nestedblock--options_list--retry))
- **tick_every** (Number, Optional)

<a id="nestedblock--options_list--monitor_options"></a>
### Nested Schema for `options_list.monitor_options`

Optional:

- **renotify_interval** (Number, Optional)


<a id="nestedblock--options_list--retry"></a>
### Nested Schema for `options_list.retry`

Optional:

- **count** (Number, Optional)
- **interval** (Number, Optional)



<a id="nestedblock--request_basicauth"></a>
### Nested Schema for `request_basicauth`

Required:

- **password** (String, Required)
- **username** (String, Required)


<a id="nestedblock--request_client_certificate"></a>
### Nested Schema for `request_client_certificate`

Required:

- **cert** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate--cert))
- **key** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--request_client_certificate--key))

<a id="nestedblock--request_client_certificate--cert"></a>
### Nested Schema for `request_client_certificate.cert`

Required:

- **content** (String, Required)

Optional:

- **filename** (String, Optional)


<a id="nestedblock--request_client_certificate--key"></a>
### Nested Schema for `request_client_certificate.key`

Required:

- **content** (String, Required)

Optional:

- **filename** (String, Optional)



<a id="nestedblock--step"></a>
### Nested Schema for `step`

Required:

- **name** (String, Required)
- **params** (String, Required)
- **type** (String, Required)

Optional:

- **allow_failure** (Boolean, Optional)
- **timeout** (Number, Optional)


<a id="nestedblock--variable"></a>
### Nested Schema for `variable`

Required:

- **name** (String, Required)
- **type** (String, Required)

Optional:

- **example** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **pattern** (String, Optional)


