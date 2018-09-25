## 1.3.0 (Unreleased)

IMPROVEMENTS:

* Add full support for Datadog screenboards ([#91](https://github.com/terraform-providers/terraform-provider-datadog/pull/91))

BUGFIXES:

* Do not compute `new_host_delay` ([#88](https://github.com/terraform-providers/terraform-provider-datadog/pull/88))
* Remove buggy uptime widget ([#93](https://github.com/terraform-providers/terraform-provider-datadog/pull/93))

## 1.2.0 (August 27, 2018)

BUG FIXES:

* Update "monitor type" options in docs ([#81](https://github.com/terraform-providers/terraform-provider-datadog/pull/81))
* Fix typo in timeboard documentation ([#83](https://github.com/terraform-providers/terraform-provider-datadog/pull/83))

IMPROVEMENTS:

* Update `go-datadog-api` to v.2.11.0 and move vendoring from `gopkg.in/zorkian/go-datadog-api.v2` to `github.com/zorkian/go-datadog-api` ([#84](https://github.com/terraform-providers/terraform-provider-datadog/pull/84))
* Deprecate `is_admin` as part of the work needed to add support for `access_role` ([#85](https://github.com/terraform-providers/terraform-provider-datadog/pull/85))

## 1.1.0 (July 30, 2018)

IMPROVEMENTS:

* Added more docs detailing expected weird behaviours from the Datadog API. ([#79](https://github.com/terraform-providers/terraform-provider-datadog/pull/79))
* Added support for 'unknown' monitor threshold field. ([#45](https://github.com/terraform-providers/terraform-provider-datadog/pull/45))
* Deprecated the `role` argument for `User` resources since it's now a noop on the Datadog API. ([#80](https://github.com/terraform-providers/terraform-provider-datadog/pull/80))

## 1.0.4 (July 06, 2018)

BUG FIXES:

* Bump `go-datadog-api.v2` to v2.10.0 thus fixing tag removal on monitor updates ([#43](https://github.com/terraform-providers/terraform-provider-datadog/issues/43))

## 1.0.3 (January 03, 2018)

IMPROVEMENTS:

* `datadog_downtime` - adding support for setting `monitor_id` ([#18](https://github.com/terraform-providers/terraform-provider-datadog/issues/18))

## 1.0.2 (December 19, 2017)

IMPROVEMENTS:

* `datadog_monitor` - Add support for monitor recovery thresholds ([#37](https://github.com/terraform-providers/terraform-provider-datadog/issues/37))

BUG FIXES:

* Fix issue with DataDog service converting metric alerts to query alerts ([#16](https://github.com/terraform-providers/terraform-provider-datadog/issues/16))

## 1.0.1 (December 06, 2017)

BUG FIXES:

* Fix issue reading resources that have been updated outside of Terraform ([#34](https://github.com/terraform-providers/terraform-provider-datadog/issues/34))

## 1.0.0 (October 20, 2017)

BUG FIXES:

* Improved detection of "drift" when graphs are reconfigured outside of Terraform. ([#27](https://github.com/terraform-providers/terraform-provider-datadog/issues/27))
* Fixed API response decoding error on graphs. ([#27](https://github.com/terraform-providers/terraform-provider-datadog/issues/27))

## 0.1.1 (September 26, 2017)

FEATURES:

* **New Resource:** `datadog_metric_metadata` ([#17](https://github.com/terraform-providers/terraform-provider-datadog/issues/17))


## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
