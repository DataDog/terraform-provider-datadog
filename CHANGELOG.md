## 2.14.0 (October 27, 2020)

FEATURES:

- `datadog_logs_archive_order`: Add a new resource to reorder logs archives ([#694](https://github.com/DataDog/terraform-provider-datadog/pull/694))
- `datadog_synthetics_global_variable`: Add a new resource to support global variables in synthetics tests ([#675](https://github.com/DataDog/terraform-provider-datadog/pull/675))

IMPROVEMENTS:

- `datadog_dashboard`: Add support for `apm_stats_query` request type in widgets ([#676](https://github.com/DataDog/terraform-provider-datadog/pull/676)).
- `datadog_dashboard`: Add support for dual y-axis for timeseries widgets ([#685](https://github.com/DataDog/terraform-provider-datadog/pull/685)).
- `datadog_dashboard`: Add support for `has_search_bar` and `cell_display_mode` properties on widgets ([#686](https://github.com/DataDog/terraform-provider-datadog/pull/686)).
- `datadog_dashboard`: Add support for `custom_links` property on widgets ([#696](https://github.com/DataDog/terraform-provider-datadog/pull/696)).
- `datadog_logs_archive`: Add `rehydration_tags` property ([#705](https://github.com/DataDog/terraform-provider-datadog/pull/705)).
- `datadog_logs_archive`: Add `include_tags` property ([#715](https://github.com/DataDog/terraform-provider-datadog/pull/715)).
- `datadog_logs_custom_pipeline`: Add `target_format` property to the Logs attribute remapper ([#682](https://github.com/DataDog/terraform-provider-datadog/pull/682)).
- `datadog_service_level_objective`: Add validate option ([#672](https://github.com/DataDog/terraform-provider-datadog/pull/672))
- `datadog_synthetics_test`: Add support for DNS tests ([#673](https://github.com/DataDog/terraform-provider-datadog/pull/673)).
- `datadog_synthetics_test`: Add support for global variables ([#691](https://github.com/DataDog/terraform-provider-datadog/pull/691)).
- `datadog_synthetics_test`: Add support for `dns_server` and `request_client_certificate` properties ([#711](https://github.com/DataDog/terraform-provider-datadog/pull/711)).

BUGFIXES:

- `datadog_synthetics_test`: Don't ignore options diff ([#707](https://github.com/DataDog/terraform-provider-datadog/pull/707)).
- `datadog_synthetics_test`: Make `tags` property optional ([#712](https://github.com/DataDog/terraform-provider-datadog/pull/712)).
- `datadog_ip_ranges`: Support EU site ([#713](https://github.com/DataDog/terraform-provider-datadog/pull/713)).

## 2.13.0 (September 16, 2020)

FEATURES:

- `datadog_dashboard_list`: Add a new datasource for dashboard lists ([#657](https://github.com/DataDog/terraform-provider-datadog/pull/657))
- `datadog_synthetics_locations`: Add a new datasource for locations ([#309](https://github.com/DataDog/terraform-provider-datadog/pull/309))

IMPROVEMENTS:

- `datadog_dashboard`: A new `dashboard_lists` attribute allows adding dashboard to dashboard lists in the resource itself ([#654](https://github.com/DataDog/terraform-provider-datadog/pull/654)).
- `datadog_dashboard`: Add support for `multi_compute` attribute ([#629](https://github.com/DataDog/terraform-provider-datadog/pull/629))
- `datadog_dashboard`: Add support for `metric` in `conditional_formats` ([#617](https://github.com/DataDog/terraform-provider-datadog/pull/617))
- `datadog_dashboard`: Add support for `rum_query` and `security_query` widget requests ([#416](https://github.com/DataDog/terraform-provider-datadog/pull/416))
- `datadog_monitor`: Monitors are now validated during plan ([#639](https://github.com/DataDog/terraform-provider-datadog/pull/639))
- `datadog_downtime`: Add support for recurrent rules ([#610](https://github.com/DataDog/terraform-provider-datadog/pull/610))
- `datadog_synthetics_test`: Add support for steps for browser tests ([#638](https://github.com/DataDog/terraform-provider-datadog/pull/638))
- `datadog_synthetics_test`: Add subtype TCP test support for API tests ([#632](https://github.com/DataDog/terraform-provider-datadog/pull/632))
- `datadog_synthetics_test`: Add retry and monitor options ([#636](https://github.com/DataDog/terraform-provider-datadog/pull/636))

BUGFIXES:

- `datadog_dashboard`: Prevent nil pointer dereference with template variables without prefix ([#630](https://github.com/DataDog/terraform-provider-datadog/pull/630))
- `datadog_dashboard`: Don't allow empty content in note widgets ([#607](https://github.com/DataDog/terraform-provider-datadog/pull/607))
- `datadog_downtime`: Ignore useless diff on start attribute ([#597](https://github.com/DataDog/terraform-provider-datadog/pull/597))
- `datadog_logs_custom_pipeline`: Don't allow empty pipeline filter ([#605](https://github.com/DataDog/terraform-provider-datadog/pull/605))
- `provider`: Completely skip creds validation when validate is false ([#641](https://github.com/DataDog/terraform-provider-datadog/pull/641))

NOTES:

- `datadog_synthetics_test`: The `options` attribute has been deprecated by `options_list` ([#624](https://github.com/DataDog/terraform-provider-datadog/pull/624)

## 2.12.1 (July 23, 2020)

This release doesn't contain any user-facing changes. It's done as a required part of process to finalize the transfer of the provider repository under DataDog GitHub organization: https://github.com/DataDog/terraform-provider-datadog.

## 2.12.0 (July 22, 2020)

FEATURES:

- `datadog_monitor`: Add new datasource for monitors ([#569](https://github.com/DataDog/terraform-provider-datadog/issues/569)), ([#585](https://github.com/DataDog/terraform-provider-datadog/issues/585))

IMPROVEMENTS:

- `datadog_synthetics_test`: Enable usage of `validatesJSONPath` operator ([#571](https://github.com/DataDog/terraform-provider-datadog/issues/571))
- `datadog_synthetics_test`: Allow usage of the new assertion format ([#571](https://github.com/DataDog/terraform-provider-datadog/issues/571)), ([#582](https://github.com/DataDog/terraform-provider-datadog/issues/582))
- `datadog_synthetics_test`: Add support for `basicAuth` and `query` ([#586](https://github.com/DataDog/terraform-provider-datadog/issues/586))

BUGFIXES:

- `datadog_downtime`: Replace `time.LoadLocation` by tz.LoadLocation from `4d63.com/tz` package ([#560](https://github.com/DataDog/terraform-provider-datadog/issues/560))
- `datadog_downtime`: Use `TypeSet` for monitor tags to avoid unnecessary diffs ([#540](https://github.com/DataDog/terraform-provider-datadog/issues/540))
- `provider`: Respect the debug setting in the new Go Datadog client ([#580](https://github.com/DataDog/terraform-provider-datadog/issues/580))

NOTES:

- `datadog_integration_pagerduty`: This resource is deprecated. You can use `datadog_integration_pagerduty_service_object` resources directly once the integration is activated ([#584](https://github.com/DataDog/terraform-provider-datadog/issues/584))

## 2.11.0 (June 29, 2020)

FEATURES:

- `datadog_logs_archive`: Add `datadog_logs_archive` resource ([#544](https://github.com/DataDog/terraform-provider-datadog/pull/544))
- `datadog_integration_azure`: Add `datadog_integration_azure` resource ([#556](https://github.com/DataDog/terraform-provider-datadog/pull/556))

## 2.10.0 (June 26, 2020)

FEATURES:

- `datadog_integration_aws`: Add `excluded_regions` parameter ([#549](https://github.com/DataDog/terraform-provider-datadog/pull/549))
- `datadog_dashboard`: Add `ServiceMap` widget to dashboard ([#550](https://github.com/DataDog/terraform-provider-datadog/pull/550))
- `datadog_dashboard`: Add `show_legend` and `legend_size` fields to Distribution widget ([#551](https://github.com/DataDog/terraform-provider-datadog/pull/551))
- `datadog_dashboard`: Add `network_query` and `rum_query` to timeseries widget ([#555](https://github.com/DataDog/terraform-provider-datadog/pull/555))
- `datadog_dashboard`: Add `event`, `legend_size` and `show_legend` fields to heatmap widget ([#554](https://github.com/DataDog/terraform-provider-datadog/pull/554))

IMPROVEMENTS:

- `datadog_dashboard`: Add readonly url field to dashboard ([#558](https://github.com/DataDog/terraform-provider-datadog/pull/558))

## 2.9.0 (June 22, 2020)

IMPROVEMENTS:

- `datadog_monitor`: Add monitor `force_delete` parameter ([#535](https://github.com/DataDog/terraform-provider-datadog/pull/535)) Thanks [@ykyr](https://github.com/ykyr)

BUGFIXES:

- `datadog_dashboard`: Safely access index field ([#536](https://github.com/DataDog/terraform-provider-datadog/pull/536))
- `datadog_dashboard`: Set title and title_align properly on heatmap widget ([#539](https://github.com/DataDog/terraform-provider-datadog/pull/539))
- `datadog_ip_ranges`: Fix data source for IPRanges ([#542](https://github.com/DataDog/terraform-provider-datadog/pull/542))
- `datadog_monitor`: Fix indent in datadog_monitor docs example ([#543](https://github.com/DataDog/terraform-provider-datadog/pull/543)) Thanks [@nekottyo](https://github.com/nekottyo)

NOTES:

- `datadog_synthetics_test`: `SyntheticsDeviceID` should accept all allowed values ([#538](https://github.com/DataDog/terraform-provider-datadog/issues/538))
- Thanks [@razaj92](https://github.com/razaj92) ([#547](https://github.com/DataDog/terraform-provider-datadog/pull/547)) who contributed to this release as well.

## 2.8.0 (June 10, 2020)

FEATURES:

- `provider`: Add support for `DD_API_KEY`, `DD_APP_KEY` and `DD_HOST` env variables ([#469](https://github.com/DataDog/terraform-provider-datadog/issues/469))
- `datadog_logs_custom_pipeline`: Add support for lookup processor ([#415](https://github.com/DataDog/terraform-provider-datadog/issues/415))
- `datadog_integration_aws_lambda_arn`: Add AWS Log Lambda Integration ([#436](https://github.com/DataDog/terraform-provider-datadog/issues/436))
- `datadog_integration_aws_log_collection`: Add AWS Log collection service resource ([#437](https://github.com/DataDog/terraform-provider-datadog/issues/437)) Thanks [@mhaley-miovision](https://github.com/mhaley-miovision)
- `datadog_dashboard`: Add support for tags_execution ([#524](https://github.com/DataDog/terraform-provider-datadog/issues/524))
- `datadog_dashboard`: Add `legend_size` to api request ([#421](https://github.com/DataDog/terraform-provider-datadog/issues/421))
- `provider`: Add "validate" option that can disable validation ([#474](https://github.com/DataDog/terraform-provider-datadog/issues/474)) Thanks [@bendrucker](https://github.com/bendrucker)

IMPROVEMENTS:

- `provider`: Harmonized errors across all resources ([#450](https://github.com/DataDog/terraform-provider-datadog/issues/450))
- `provider`: Add more infos in user agent header ([#455](https://github.com/DataDog/terraform-provider-datadog/issues/455))
- `provider`: Update the api error message ([#472](https://github.com/DataDog/terraform-provider-datadog/issues/472))
- `datadog_screenboard`, `datadog_timeboard`: Add deprecation messages ([#496](https://github.com/DataDog/terraform-provider-datadog/issues/496))
- `provider`: New UserAgent Header ([#455](https://github.com/DataDog/terraform-provider-datadog/issues/455)), ([#510](https://github.com/DataDog/terraform-provider-datadog/issues/510)), ([#511](https://github.com/DataDog/terraform-provider-datadog/issues/511)), and ([#512](https://github.com/DataDog/terraform-provider-datadog/issues/512))
- `datadog_integration_aws`: Add full AWS Update support ([#521](https://github.com/DataDog/terraform-provider-datadog/issues/521))

BUGFIXES:

- `datadog_logs_index`: Fail fast if index isn't imported ([#452](https://github.com/DataDog/terraform-provider-datadog/issues/452))
- `datadog_integration_aws`: Do not set empty structures in request to create aws integration ([#505](https://github.com/DataDog/terraform-provider-datadog/issues/505)) Thanks [@miguelaferreira](https://github.com/miguelaferreira)
- `datadog_dashboard`: Add default to deprecated `count` field to avoid sending 0 ([#514](https://github.com/DataDog/terraform-provider-datadog/issues/514))
- `datadog_integration_pagerduty`: Fix perpetual diff in api_token ([#518](https://github.com/DataDog/terraform-provider-datadog/issues/518)) Thanks [@bendrucker](https://github.com/bendrucker)
- `datadog_dashboard`: Add column revamp properties to dashboard log stream widget ([#517](https://github.com/DataDog/terraform-provider-datadog/issues/517))

NOTES:

- This release replaces the underlying community driven Datadog API Go client [go-datadog-api](https://github.com/zorkian/go-datadog-api) with the Datadog Official API Go client [datadog-api-client-go](https://github.com/DataDog/datadog-api-client-go) for all resources listed below:
  - `provider`: Add Datadog Go client API ([#477](https://github.com/DataDog/terraform-provider-datadog/issues/477)) and ([#456](https://github.com/DataDog/terraform-provider-datadog/issues/456))
  - `datadog_service_level_objective`: Migrate SLO resource with Datadog Go Client ([#490](https://github.com/DataDog/terraform-provider-datadog/issues/490))
  - `datadog_metric_metadata`: Migrate metric_metadata resource to use Datadog Go client ([#486](https://github.com/DataDog/terraform-provider-datadog/issues/486))
  - `datadog_integration_aws`: Migrate AWS resource to use Datadog Go client ([#481](https://github.com/DataDog/terraform-provider-datadog/issues/481))
  - `datadog_integration_gcp`: Migrate GCP resource to use Datadog Go client ([#482](https://github.com/DataDog/terraform-provider-datadog/issues/482))
  - `datadog_downtime`: Migrate Downtime resource to use Datadog Go client ([#480](https://github.com/DataDog/terraform-provider-datadog/issues/480))
  - `datadog_ip_ranges`: Migrate IP Range resource with Datadog Go client ([#491](https://github.com/DataDog/terraform-provider-datadog/issues/491))
  - `datadog_integration_pagerduty_service_object`: Migrate pagerduty_service_object resource to use Datadog Go client ([#488](https://github.com/DataDog/terraform-provider-datadog/issues/488))
  - `datadog_logs_index`, `datadog_logs_index_order`, `datadog_logs_integration_pipeline`, `datadog_logs_pipeline_order`: Migrate Logs resources to use Datadog Go client ([#483](https://github.com/DataDog/terraform-provider-datadog/issues/483))
  - `datadog_monitor`: Migrate monitor resource to use Datadog Go client ([#485](https://github.com/DataDog/terraform-provider-datadog/issues/485))
  - `datadog_dashboard_list`: Migrate Dashboard_list resource to use Datadog Go client ([#479](https://github.com/DataDog/terraform-provider-datadog/issues/479))
  - `datadog_integration_aws_log_collection`: Migrate aws_log_collection resource to use Datadog Go client ([#501](https://github.com/DataDog/terraform-provider-datadog/issues/501))
  - `datadog_logs_custom_pipeline`: Migrate Logs custom pipeline resource to utilize Datadog Go client ([#495](https://github.com/DataDog/terraform-provider-datadog/issues/495))
  - `datadog_synthetics_test_`: Migrate synthetics resource to utilize Datadog Go Client ([#499](https://github.com/DataDog/terraform-provider-datadog/issues/499))
  - `datadog_integration_aws_log_collection`, `datadog_integration_aws_lambda_arn`: Migrate AWS logs to use the Datadog Go Client ([#497](https://github.com/DataDog/terraform-provider-datadog/issues/497))
  - `datadog_dashboard`: Migrate dashboard resource to use Datadog Go client ([#489](https://github.com/DataDog/terraform-provider-datadog/issues/489))
- `datadog_screenboard` and `datadog_timeboard` resources are deprecated and should be converted to `datadog_dashboard` resources.
- Thanks [@NeverTwice](https://github.com/NeverTwice) ([#460](https://github.com/DataDog/terraform-provider-datadog/pull/460)) and [@sepulworld](https://github.com/sepulworld) ([#506](https://github.com/DataDog/terraform-provider-datadog/pull/506)) who contributed to this release as well.

## 2.7.0 (February 10, 2020)

IMPROVEMENTS:

- `datadog_dashboard`: Add `template_variable_presets` parameter ([#401](https://github.com/DataDog/terraform-provider-datadog/issues/401))
- `datadog_dashboard`: Add new Monitor Summary widget parameters: `summary_type` and `show_last_triggered` ([#396](https://github.com/DataDog/terraform-provider-datadog/issues/396))
- `datadog_dashboard`: Hide deprecated Monitor Summary widget parameters: `count` and `start` ([#403](https://github.com/DataDog/terraform-provider-datadog/issues/403))
- `datadog_monitor`: Improve monitor example with ignoring changes on silenced ([#406](https://github.com/DataDog/terraform-provider-datadog/issues/406))
- `datadog_service_level_objective`: Fix optional threshold fields handling when updating ([#400](https://github.com/DataDog/terraform-provider-datadog/issues/400))

BUGFIXES:

- `datadog_downtime`: Gracefully handle recreating downtimes that were canceled manually ([#405](https://github.com/DataDog/terraform-provider-datadog/issues/405))
- `datadog_screenboard`: Properly set screenboard attributes from client response to not produce non-empty plans ([#404](https://github.com/DataDog/terraform-provider-datadog/issues/404))

NOTES:

- This is the first release to use the new `terraform-plugin-sdk` ([#346](https://github.com/DataDog/terraform-provider-datadog/issues/346))

## 2.6.0 (January 21, 2020)

FEATURES:

- `datadog_dashboard`: Add Datadog dashboard SLO widget support ([#355](https://github.com/DataDog/terraform-provider-datadog/issues/355)) Thanks [@mbarrien](https://github.com/mbarrien)

IMPROVEMENTS:

- `datadog_logs_custom_pipeline`: Support all processors in Logs pipeline ([#357](https://github.com/DataDog/terraform-provider-datadog/pull/357)) Thanks [@tt810](https://github.com/tt810)

BUGFIXES:

- `datadog_service_level_objective`: Fix slo threshold warning value modified when storing the state ([#352](https://github.com/DataDog/terraform-provider-datadog/pull/352))
- `datadog_service_level_objective`: `monitor_search` schema removed from the SLO resource as it is not yet supported ([#358](https://github.com/DataDog/terraform-provider-datadog/issues/358)) Thanks [@unclebconnor](https://github.com/unclebconnor)
- `datadog_monitor`: Resolve non empty diff: "no_data_timeframe = 0 -> 10" on plan diff ([#384](https://github.com/DataDog/terraform-provider-datadog/issues/384)) Thanks [@abicky](https://github.com/abicky)

## 2.5.0 (October 22, 2019)

FEATURES:

- `datadog_ip_ranges`: New data source for IP ranges ([#298](https://github.com/DataDog/terraform-provider-datadog/issues/298))
- `datadog_logs_custom_pipeline`: New resource for custom logs pipelines ([#312](https://github.com/DataDog/terraform-provider-datadog/issues/312), [#332](https://github.com/DataDog/terraform-provider-datadog/issues/332))
- `datadog_logs_index`: New resource for logs indexes ([#326](https://github.com/DataDog/terraform-provider-datadog/issues/326))
- `datadog_logs_index_order`: New resource for logs index ordering ([#326](https://github.com/DataDog/terraform-provider-datadog/issues/326))
- `datadog_logs_integration_pipeline`: New resource for integration logs pipelines ([#312](https://github.com/DataDog/terraform-provider-datadog/issues/312), [#332](https://github.com/DataDog/terraform-provider-datadog/issues/332))
- `datadog_logs_pipeline_order`: New resources for logs pipeline ordering ([#312](https://github.com/DataDog/terraform-provider-datadog/issues/312))

IMPROVEMENTS:

- `datadog_dashboard`: Added documentation of `event` and `axis` ([#314](https://github.com/DataDog/terraform-provider-datadog/issues/314))
- `datadog_screenboard`: Added `count` as a valid aggregation method ([#333](https://github.com/DataDog/terraform-provider-datadog/issues/333))

BUGFIXES:

- `datadog_dashboard`: Fixed parsing of `compute.interval` and `group_by.sort.facet`, mark `group_by.facet` as optional for apm and log queries ([#322](https://github.com/DataDog/terraform-provider-datadog/issues/322), [#325](https://github.com/DataDog/terraform-provider-datadog/issues/325))
- `datadog_dashboard`: Properly respect `show_legend` ([#329](https://github.com/DataDog/terraform-provider-datadog/issues/329))
- `datadog_integration_pagerduty`: Add missing exists methods to prevent failing when resource was manually removed outside of Terraform ([#324](https://github.com/DataDog/terraform-provider-datadog/issues/324))
- `datadog_integration_pagerduty_service_object`: Add missing exists methods to prevent failing when resource was manually removed outside of Terraform ([#324](https://github.com/DataDog/terraform-provider-datadog/issues/324))

## 2.4.0 (September 11, 2019)

FEATURES:

- `datadog_dashboard_list`: New resource for dashboard lists ([#296](https://github.com/DataDog/terraform-provider-datadog/issues/296))

IMPROVEMENTS:

- `datadog_dashboard`: Allow specifying `event` and `yaxis` for timeseries definitions ([#282](https://github.com/DataDog/terraform-provider-datadog/issues/282))

## 2.3.0 (August 29, 2019)

IMPROVEMENTS:

- `datadog-dashboards`: Add resources for log, apm and process query in legacy dashboards ([#272](https://github.com/DataDog/terraform-provider-datadog/issues/272))

BUGFIXES:

- `datadog_integration_pagerduty`: Make sure PD services don't get removed by updating PD resource ([#304](https://github.com/DataDog/terraform-provider-datadog/issues/304))

## 2.2.0 (August 19, 2019)

FEATURES:

- `datadog_service_level_objective`: New resource for Service Level Objective (SLO) ([#263](https://github.com/DataDog/terraform-provider-datadog/issues/263))

IMPROVEMENTS:

- `datadog_dashbaord`: Add support for style block in dashboard widgets. ([#277](https://github.com/DataDog/terraform-provider-datadog/issues/277))
- `datadog_dashboard`: Add support for metadata block in dashboard widgets ([#278](https://github.com/DataDog/terraform-provider-datadog/issues/278))
- `datadog_synthetics_test`: Support SSL synthetics tests. ([#279](https://github.com/DataDog/terraform-provider-datadog/issues/279))

BUGFIXES:

- `datadog_dashboards`: Safely type assert optional fields from log and apm query to avoid a panic if they aren't supplied ([#283](https://github.com/DataDog/terraform-provider-datadog/issues/283))
- `datadog_synthetics_test`: Fix follow redirects field to properly apply and save in state. ([#256](https://github.com/DataDog/terraform-provider-datadog/issues/256))

## 2.1.0 (July 24, 2019)

FEATURES:

- `datadog_dashboard`: New Resource combining screenboard and timeboard, allowing a single config to manage all of your Datadog Dashboards. ([#249](https://github.com/DataDog/terraform-provider-datadog/issues/249))
- `datadog_integration_pagerduty_service_object`: New Resource that allows the configuration of individual pagerduty services for the Datadog Pagerduty Integration. ([#237](https://github.com/DataDog/terraform-provider-datadog/issues/237))

IMPROVEMENTS:

- `datadog_aws`: Add a mutex around all API operations for this resource. ([#254](https://github.com/DataDog/terraform-provider-datadog/issues/254))
- `datadog_downtime`: General improvements around allowing the resource to be ran multiple times without sending any unchanged values for the start/end times. Also fixes non empty diff when monitor_tags isn't set. ([#264](https://github.com/DataDog/terraform-provider-datadog/issues/264)] [[#267](https://github.com/DataDog/terraform-provider-datadog/issues/267))
- `datadog_monitor`: Only add a threshold window if a recovery or trigger window is set. [[#260](https://github.com/DataDog/terraform-provider-datadog/issues/260)] Thanks [@heldersepu](https://github.com/heldersepu)
- `datadog_user`: Make `is_admin` computed to continue its deprecation path and avoid spurious diffs. ([#251](https://github.com/DataDog/terraform-provider-datadog/issues/251))

NOTES:

- This release includes Terraform SDK upgrade to 0.12.5. ([#265](https://github.com/DataDog/terraform-provider-datadog/issues/265))

## 2.0.2 (June 26, 2019)

BUGFIXES:

- `datadog_monitor`: DiffSuppress the difference between `metric alert` and `query alert` no matter what is in the current state and prevent the force recreation of monitors due to this change. ([#247](https://github.com/DataDog/terraform-provider-datadog/issues/247))

## 2.0.1 (June 21, 2019)

BUGFIXES:

- `datadog_monitor`: Don't force the destruction and recreation of a monitor when the type changes between `metric alert` and `query alert`. ([#242](https://github.com/DataDog/terraform-provider-datadog/issues/242))

## 2.0.0 (June 18, 2019)

NOTES:

- `datadog_monitor`: The silence attribute is beginning its deprecation process, please use `datadog_downtime` instead ([#221](https://github.com/DataDog/terraform-provider-datadog/issues/221))

IMPROVEMENTS:

- `datadog_monitor`: Use ForceNew when changing the Monitor type ([#236](https://github.com/DataDog/terraform-provider-datadog/issues/236))
- `datadog_monitor`: Add default to `no data` timeframe of 10 minutes. ([#212](https://github.com/DataDog/terraform-provider-datadog/issues/212))
- `datadog_synthetics_test`: Support synthetics monitors in composite monitors. ([#222](https://github.com/DataDog/terraform-provider-datadog/issues/222))
- `datadog_downtime`: Add validation to tags, add timezone parameter, improve downtime id handling, add descriptions to fields. ([#204](https://github.com/DataDog/terraform-provider-datadog/issues/204))
- `datadog_screenboard`: Add support for metadata alias in graphs. ([#215](https://github.com/DataDog/terraform-provider-datadog/issues/215))
- `datadog_screenboard`: Add `custom_bg_color` to graph config. [[#189](https://github.com/DataDog/terraform-provider-datadog/issues/189)] Thanks [@milanvdm](https://github.com/milanvdm)
- Update the vendored go client to `v2.21.0`. ([#230](https://github.com/DataDog/terraform-provider-datadog/issues/230))

BUGFIXES:

- `datadog_timeboard`: Fix the `extra_col` from having a non empty plan when there are no changes. ([#231](https://github.com/DataDog/terraform-provider-datadog/issues/231))
- `datadog_timeboard`: Fix the `precision` from having a non empty plan when there are no changes. ([#228](https://github.com/DataDog/terraform-provider-datadog/issues/228))
- `datadog_monitor`: Fix the sorting of monitor tags that could lead to a non empty diff. ([#214](https://github.com/DataDog/terraform-provider-datadog/issues/214))
- `datadog_monitor`: Properly save `query_config` as to avoid to an improper non empty diff. ([#209](https://github.com/DataDog/terraform-provider-datadog/issues/209))
- `datadog_monitor`: Fix and clarify documentation on unmuting monitor scopes. ([#202](https://github.com/DataDog/terraform-provider-datadog/issues/202))
- `datadog_screenboard`: Change monitor schema to be of type String instead of Int. [[#154](https://github.com/DataDog/terraform-provider-datadog/issues/154)] Thanks [@mnaboka](https://github.com/mnaboka)

## 1.9.0 (May 09, 2019)

IMPROVEMENTS:

- `datadog_downtime`: Add `monitor_tags` getting and setting ([#167](https://github.com/DataDog/terraform-provider-datadog/issues/167))
- `datadog_monitor`: Add support for `enable_logs` in log monitors ([#151](https://github.com/DataDog/terraform-provider-datadog/issues/151))
- `datadog_monitor`: Add suport for `threshold_windows` attribute ([#131](https://github.com/DataDog/terraform-provider-datadog/issues/131))
- Support importing dashboards using the new string ID ([#184](https://github.com/DataDog/terraform-provider-datadog/issues/184))
- Various documentation fixes and improvements ([#152](https://github.com/DataDog/terraform-provider-datadog/issues/152), [#171](https://github.com/DataDog/terraform-provider-datadog/issues/171), [#176](https://github.com/DataDog/terraform-provider-datadog/issues/176), [#178](https://github.com/DataDog/terraform-provider-datadog/issues/178), [#180](https://github.com/DataDog/terraform-provider-datadog/issues/180), [#183](https://github.com/DataDog/terraform-provider-datadog/issues/183))

NOTES:

- This release includes Terraform SDK upgrade to 0.12.0-rc1. The provider is backwards compatible with Terraform v0.11.X, there should be no significant changes in behavior. Please report any issues to either [Terraform issue tracker](https://github.com/hashicorp/terraform/issues) or to [Terraform Datadog Provider issue tracker](https://github.com/DataDog/terraform-provider-datadog/issues) ([#194](https://github.com/DataDog/terraform-provider-datadog/issues/194), [#198](https://github.com/DataDog/terraform-provider-datadog/issues/198))

## 1.8.0 (April 15, 2019)

INTERNAL:

- provider: Enable request/response logging in `>=DEBUG` mode ([#153](https://github.com/DataDog/terraform-provider-datadog/issues/153))

IMPROVEMENTS:

- Add Synthetics API and Browser tests support + update go-datadog-api to latest. ([169](https://github.com/DataDog/terraform-provider-datadog/pull/169))

## 1.7.0 (March 05, 2019)

BUGFIXES:

- Bump go api client to 2.19.0 to fix TileDefStyle.fillMax type errors. ([143](https://github.com/DataDog/terraform-provider-datadog/pull/143))([144](https://github.com/DataDog/terraform-provider-datadog/pull/144))
- Fix the usage of `start_date` and `end_data` only being read on the first apply. ([145](https://github.com/DataDog/terraform-provider-datadog/pull/145))

IMPROVEMENTS:

- Upgrade to Go 1.11. ([141](https://github.com/DataDog/terraform-provider-datadog/pull/141/files))
- Add AWS Integration resource to the docs. ([146](https://github.com/DataDog/terraform-provider-datadog/pull/146))

FEATURES:

- **New Resource:** `datadog_integration_pagerduty` ([135](https://github.com/DataDog/terraform-provider-datadog/pull/135))

## 1.6.0 (November 30, 2018)

BUGFIXES:

- the graph.style.palette_flip field is a boolean but only works if it's passed as a string. ([#29](https://github.com/DataDog/terraform-provider-datadog/issues/29))
- datadog_monitor - Removal of 'silenced' resource argument has no practical effect. ([#41](https://github.com/DataDog/terraform-provider-datadog/issues/41))
- datadog_screenboard - widget swapping `x` and `y` parameters. ([#119](https://github.com/DataDog/terraform-provider-datadog/issues/119))
- datadog_screenboard - panic: interface conversion: interface {} is string, not float64. ([#117](https://github.com/DataDog/terraform-provider-datadog/issues/117))

IMPROVEMENTS:

- Feature Request: AWS Integration. ([#76](https://github.com/DataDog/terraform-provider-datadog/issues/76))
- Bump datadog api to v2.18.0 and add support for include units and zero. ([#121](https://github.com/DataDog/terraform-provider-datadog/pull/121))

## 1.5.0 (November 06, 2018)

IMPROVEMENTS:

- Add Google Cloud Platform integration ([#108](https://github.com/DataDog/terraform-provider-datadog/pull/108))
- Add new hostmap widget options: `node type`, `fill_min` and `fill_max`. ([#106](https://github.com/DataDog/terraform-provider-datadog/pull/106))
- Use dates to set downtime interval, improve docs. ([#113](https://github.com/DataDog/terraform-provider-datadog/pull/113))
- Bump Terraform provider SDK to latest. ([#110](https://github.com/DataDog/terraform-provider-datadog/pull/110))
- Better document `evaluation_delay` option. ([#112](https://github.com/DataDog/terraform-provider-datadog/pull/112))

## 1.4.0 (October 02, 2018)

IMPROVEMENTS:

- Pull changes from go-datadog-api v2.14.0 ([#99](https://github.com/DataDog/terraform-provider-datadog/pull/99))
- Add `api_url` argument to the provider ([#101](https://github.com/DataDog/terraform-provider-datadog/pull/101))

BUGFIXES:

- Allow `new_host_delay` to be unset ([#100](https://github.com/DataDog/terraform-provider-datadog/issues/100))

## 1.3.0 (September 25, 2018)

IMPROVEMENTS:

- Add full support for Datadog screenboards ([#91](https://github.com/DataDog/terraform-provider-datadog/pull/91))

BUGFIXES:

- Do not compute `new_host_delay` ([#88](https://github.com/DataDog/terraform-provider-datadog/pull/88))
- Remove buggy uptime widget ([#93](https://github.com/DataDog/terraform-provider-datadog/pull/93))

## 1.2.0 (August 27, 2018)

BUG FIXES:

- Update "monitor type" options in docs ([#81](https://github.com/DataDog/terraform-provider-datadog/pull/81))
- Fix typo in timeboard documentation ([#83](https://github.com/DataDog/terraform-provider-datadog/pull/83))

IMPROVEMENTS:

- Update `go-datadog-api` to v.2.11.0 and move vendoring from `gopkg.in/zorkian/go-datadog-api.v2` to `github.com/zorkian/go-datadog-api` ([#84](https://github.com/DataDog/terraform-provider-datadog/pull/84))
- Deprecate `is_admin` as part of the work needed to add support for `access_role` ([#85](https://github.com/DataDog/terraform-provider-datadog/pull/85))

## 1.1.0 (July 30, 2018)

IMPROVEMENTS:

- Added more docs detailing expected weird behaviours from the Datadog API. ([#79](https://github.com/DataDog/terraform-provider-datadog/pull/79))
- Added support for 'unknown' monitor threshold field. ([#45](https://github.com/DataDog/terraform-provider-datadog/pull/45))
- Deprecated the `role` argument for `User` resources since it's now a noop on the Datadog API. ([#80](https://github.com/DataDog/terraform-provider-datadog/pull/80))

## 1.0.4 (July 06, 2018)

BUG FIXES:

- Bump `go-datadog-api.v2` to v2.10.0 thus fixing tag removal on monitor updates ([#43](https://github.com/DataDog/terraform-provider-datadog/issues/43))

## 1.0.3 (January 03, 2018)

IMPROVEMENTS:

- `datadog_downtime`: adding support for setting `monitor_id` ([#18](https://github.com/DataDog/terraform-provider-datadog/issues/18))

## 1.0.2 (December 19, 2017)

IMPROVEMENTS:

- `datadog_monitor`: Add support for monitor recovery thresholds ([#37](https://github.com/DataDog/terraform-provider-datadog/issues/37))

BUG FIXES:

- Fix issue with DataDog service converting metric alerts to query alerts ([#16](https://github.com/DataDog/terraform-provider-datadog/issues/16))

## 1.0.1 (December 06, 2017)

BUG FIXES:

- Fix issue reading resources that have been updated outside of Terraform ([#34](https://github.com/DataDog/terraform-provider-datadog/issues/34))

## 1.0.0 (October 20, 2017)

BUG FIXES:

- Improved detection of "drift" when graphs are reconfigured outside of Terraform. ([#27](https://github.com/DataDog/terraform-provider-datadog/issues/27))
- Fixed API response decoding error on graphs. ([#27](https://github.com/DataDog/terraform-provider-datadog/issues/27))

## 0.1.1 (September 26, 2017)

FEATURES:

- **New Resource:** `datadog_metric_metadata` ([#17](https://github.com/DataDog/terraform-provider-datadog/issues/17))

## 0.1.0 (June 20, 2017)

NOTES:

- Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
