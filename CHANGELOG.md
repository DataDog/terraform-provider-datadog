## 2.2.1 (Unreleased)
## 2.2.0 (August 19, 2019)

FEATURES:
* `datadog_service_level_objective`: New resource for Service Level Objective (SLO) ([#263](https://github.com/terraform-providers/terraform-provider-datadog/issues/263))

IMPROVEMENTS:
* `datadog_dashbaord`: Add support for style block in dashboard widgets. ([#277](https://github.com/terraform-providers/terraform-provider-datadog/issues/277))
* `datadog_dashboard`: Add support for metadata block in dashboard widgets ([#278](https://github.com/terraform-providers/terraform-provider-datadog/issues/278))
* `datadog_synthetics_test`: Support SSL synthetics tests. ([#279](https://github.com/terraform-providers/terraform-provider-datadog/issues/279))

BUGFIXES:
* `datadog_dashboards`: Safely type assert optional fields from log and apm query to avoid a panic if they aren't supplied ([#283](https://github.com/terraform-providers/terraform-provider-datadog/issues/283))
* `datadog_synthetics_test`: Fix follow redirects field to properly apply and save in state. ([#256](https://github.com/terraform-providers/terraform-provider-datadog/issues/256))

## 2.1.0 (July 24, 2019)

FEATURES:
* `datadog_dashboard`: New Resource combining screenboard and timeboard, allowing a single config to manage all of your Datadog Dashboards. ([#249](https://github.com/terraform-providers/terraform-provider-datadog/issues/249))
* `datadog_integration_pagerduty_service_object`: New Resource that allows the configuration of individual pagerduty services for the Datadog Pagerduty Integration. ([#237](https://github.com/terraform-providers/terraform-provider-datadog/issues/237))

IMPROVEMENTS:
* `datadog_aws`: Add a mutex around all API operations for this resource. ([#254](https://github.com/terraform-providers/terraform-provider-datadog/issues/254))
* `datadog_downtime`: General improvements around allowing the resource to be ran multiple times without sending any unchanged values for the start/end times. Also fixes non empty diff when monitor_tags isn't set. ([#264](https://github.com/terraform-providers/terraform-provider-datadog/issues/264)] [[#267](https://github.com/terraform-providers/terraform-provider-datadog/issues/267))
* `datadog_monitor`: Only add a threshold window if a recovery or trigger window is set. [[#260](https://github.com/terraform-providers/terraform-provider-datadog/issues/260)] Thanks [@heldersepu](https://github.com/heldersepu)
* `datadog_user`: Make `is_admin` computed to continue its deprecation path and avoid spurious diffs. ([#251](https://github.com/terraform-providers/terraform-provider-datadog/issues/251))

NOTES:
* This release includes Terraform SDK upgrade to 0.12.5. ([#265](https://github.com/terraform-providers/terraform-provider-datadog/issues/265))

## 2.0.2 (June 26, 2019)

BUGFIXES:
* `datadog_monitor`: DiffSuppress the difference between `metric alert` and `query alert` no matter what is in the current state and prevent the force recreation of monitors due to this change. ([#247](https://github.com/terraform-providers/terraform-provider-datadog/issues/247))

## 2.0.1 (June 21, 2019)

BUGFIXES:
* `datadog_monitor`: Don't force the destruction and recreation of a monitor when the type changes between `metric alert` and `query alert`. ([#242](https://github.com/terraform-providers/terraform-provider-datadog/issues/242))

## 2.0.0 (June 18, 2019)

NOTES:
* `datadog_monitor`: The silence attribute is beginning its deprecation process, please use `datadog_downtime` instead ([#221](https://github.com/terraform-providers/terraform-provider-datadog/issues/221))

IMPROVEMENTS:
* `datadog_monitor`: Use ForceNew when changing the Monitor type ([#236](https://github.com/terraform-providers/terraform-provider-datadog/issues/236))
* `datadog_monitor`: Add default to `no data` timeframe of 10 minutes. ([#212](https://github.com/terraform-providers/terraform-provider-datadog/issues/212))
* `datadog_synthetics_test`: Support synthetics monitors in composite monitors. ([#222](https://github.com/terraform-providers/terraform-provider-datadog/issues/222))
* `datadog_downtime`: Add validation to tags, add timezone parameter, improve downtime id handling, add descriptions to fields. ([#204](https://github.com/terraform-providers/terraform-provider-datadog/issues/204))
* `datadog_screenboard`: Add support for metadata alias in graphs. ([#215](https://github.com/terraform-providers/terraform-provider-datadog/issues/215))
* `datadog_screenboard`: Add `custom_bg_color` to graph config. [[#189](https://github.com/terraform-providers/terraform-provider-datadog/issues/189)] Thanks [@milanvdm](https://github.com/milanvdm)
* Update the vendored go client to `v2.21.0`. ([#230](https://github.com/terraform-providers/terraform-provider-datadog/issues/230))

BUGFIXES:
* `datadog_timeboard`: Fix the `extra_col` from having a non empty plan when there are no changes. ([#231](https://github.com/terraform-providers/terraform-provider-datadog/issues/231))
* `datadog_timeboard`: Fix the `precision` from having a non empty plan when there are no changes. ([#228](https://github.com/terraform-providers/terraform-provider-datadog/issues/228))
* `datadog_monitor`: Fix the sorting of monitor tags that could lead to a non empty diff. ([#214](https://github.com/terraform-providers/terraform-provider-datadog/issues/214))
* `datadog_monitor`: Properly save `query_config` as to avoid to an improper non empty diff. ([#209](https://github.com/terraform-providers/terraform-provider-datadog/issues/209))
* `datadog_monitor`: Fix and clarify documentation on unmuting monitor scopes. ([#202](https://github.com/terraform-providers/terraform-provider-datadog/issues/202))
* `datadog_screenboard`: Change monitor schema to be of type String instead of Int. [[#154](https://github.com/terraform-providers/terraform-provider-datadog/issues/154)] Thanks [@mnaboka](https://github.com/mnaboka)

## 1.9.0 (May 09, 2019)

IMPROVEMENTS:

* `datadog_downtime`:  Add `monitor_tags` getting and setting ([#167](https://github.com/terraform-providers/terraform-provider-datadog/issues/167))
* `datadog_monitor`: Add support for `enable_logs` in log monitors ([#151](https://github.com/terraform-providers/terraform-provider-datadog/issues/151))
* `datadog_monitor`: Add suport for `threshold_windows` attribute ([#131](https://github.com/terraform-providers/terraform-provider-datadog/issues/131))
* Support importing dashboards using the new string ID ([#184](https://github.com/terraform-providers/terraform-provider-datadog/issues/184))
* Various documentation fixes and improvements ([#152](https://github.com/terraform-providers/terraform-provider-datadog/issues/152), [#171](https://github.com/terraform-providers/terraform-provider-datadog/issues/171), [#176](https://github.com/terraform-providers/terraform-provider-datadog/issues/176), [#178](https://github.com/terraform-providers/terraform-provider-datadog/issues/178), [#180](https://github.com/terraform-providers/terraform-provider-datadog/issues/180), [#183](https://github.com/terraform-providers/terraform-provider-datadog/issues/183))

NOTES:

* This release includes Terraform SDK upgrade to 0.12.0-rc1. The provider is backwards compatible with Terraform v0.11.X, there should be no significant changes in behavior. Please report any issues to either [Terraform issue tracker](https://github.com/hashicorp/terraform/issues) or to [Terraform Datadog Provider issue tracker](https://github.com/terraform-providers/terraform-provider-datadog/issues) ([#194](https://github.com/terraform-providers/terraform-provider-datadog/issues/194), [#198](https://github.com/terraform-providers/terraform-provider-datadog/issues/198))

## 1.8.0 (April 15, 2019)

INTERNAL:

* provider: Enable request/response logging in `>=DEBUG` mode ([#153](https://github.com/terraform-providers/terraform-provider-datadog/issues/153))

IMPROVEMENTS:

* Add Synthetics API and Browser tests support + update go-datadog-api to latest. ([169](https://github.com/terraform-providers/terraform-provider-datadog/pull/169))


## 1.7.0 (March 05, 2019)

BUGFIXES:

* Bump go api client to 2.19.0 to fix TileDefStyle.fillMax type errors.  ([143](https://github.com/terraform-providers/terraform-provider-datadog/pull/143))([144](https://github.com/terraform-providers/terraform-provider-datadog/pull/144))
* Fix the usage of `start_date` and `end_data` only being read on the first apply. ([145](https://github.com/terraform-providers/terraform-provider-datadog/pull/145))

IMPROVEMENTS:

* Upgrade to Go 1.11. ([141](https://github.com/terraform-providers/terraform-provider-datadog/pull/141/files))
* Add AWS Integration resource to the docs. ([146](https://github.com/terraform-providers/terraform-provider-datadog/pull/146))


FEATURES:

* **New Resource:** `datadog_integration_pagerduty` ([135](https://github.com/terraform-providers/terraform-provider-datadog/pull/135))



## 1.6.0 (November 30, 2018)

BUGFIXES:

* the graph.style.palette_flip field is a boolean but only works if it's passed as a string. ([#29](https://github.com/terraform-providers/terraform-provider-datadog/issues/29))
* datadog_monitor - Removal of 'silenced' resource argument has no practical effect. ([#41](https://github.com/terraform-providers/terraform-provider-datadog/issues/41))
* datadog_screenboard - widget swapping `x` and `y` parameters. ([#119](https://github.com/terraform-providers/terraform-provider-datadog/issues/119))
* datadog_screenboard - panic: interface conversion: interface {} is string, not float64. ([#117](https://github.com/terraform-providers/terraform-provider-datadog/issues/117))

IMPROVEMENTS:

* Feature Request: AWS Integration. ([#76](https://github.com/terraform-providers/terraform-provider-datadog/issues/76))
* Bump datadog api to v2.18.0 and add support for include units and zero. ([#121](https://github.com/terraform-providers/terraform-provider-datadog/pull/121))

## 1.5.0 (November 06, 2018)

IMPROVEMENTS:

* Add Google Cloud Platform integration ([#108](https://github.com/terraform-providers/terraform-provider-datadog/pull/108))
* Add new hostmap widget options: `node type`, `fill_min` and `fill_max`. ([#106](https://github.com/terraform-providers/terraform-provider-datadog/pull/106))
* Use dates to set downtime interval, improve docs. ([#113](https://github.com/terraform-providers/terraform-provider-datadog/pull/113))
* Bump Terraform provider SDK to latest. ([#110](https://github.com/terraform-providers/terraform-provider-datadog/pull/110))
* Better document `evaluation_delay` option. ([#112](https://github.com/terraform-providers/terraform-provider-datadog/pull/112))

## 1.4.0 (October 02, 2018)

IMPROVEMENTS:

* Pull changes from go-datadog-api v2.14.0 ([#99](https://github.com/terraform-providers/terraform-provider-datadog/pull/99))
* Add `api_url` argument to the provider ([#101](https://github.com/terraform-providers/terraform-provider-datadog/pull/101))

BUGFIXES:

* Allow `new_host_delay` to be unset ([#100](https://github.com/terraform-providers/terraform-provider-datadog/issues/100))


## 1.3.0 (September 25, 2018)

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
