# CHANGELOG

## v1.0.0-beta.11 / 2020-11-06

* [Added] Add 3 new palettes to the conditional formatting options. See [#554](https://github.com/DataDog/datadog-api-client-go/pull/554).

## v1.0.0-beta.10 / 2020-11-02

* [Changed] Change teams and services objects names to be incident specific. See [#538](https://github.com/DataDog/datadog-api-client-go/pull/538).
* [Removed] Remove `require_full_window` client default value for monitors. See [#540](https://github.com/DataDog/datadog-api-client-go/pull/540).

## 1.0.0-beta.9 / 2020-10-27

* [Added] Add missing synthetics step types. See [#534](https://github.com/DataDog/datadog-api-client-go/pull/534).
* [Added] Add include_tags in logs archives. See [#530](https://github.com/DataDog/datadog-api-client-go/pull/530).
* [Added] Add dns server and client certificate support to synthetics tests. See [#523](https://github.com/DataDog/datadog-api-client-go/pull/523).
* [Added] Add rehydration_tags property to the logs archives. See [#513](https://github.com/DataDog/datadog-api-client-go/pull/513).
* [Added] Add endpoint to reorder Logs Archives. See [#505](https://github.com/DataDog/datadog-api-client-go/pull/505).
* [Added] Add has_search_bar and cell_display_mode properties to table widget definition. See [#502](https://github.com/DataDog/datadog-api-client-go/pull/502).
* [Added] Add target_format property to the Logs attribute remapper . See [#501](https://github.com/DataDog/datadog-api-client-go/pull/501).
* [Added] Add dual y-axis configuration to time-series widget in Dashboard. See [#498](https://github.com/DataDog/datadog-api-client-go/pull/498).
* [Added] Mark logs aggregate endpoint as stable. See [#496](https://github.com/DataDog/datadog-api-client-go/pull/496).
* [Added] Add endpoint to get a Synthetics global variable. See [#489](https://github.com/DataDog/datadog-api-client-go/pull/489).
* [Added] Add assertion types for DNS Synthetics tests. See [#486](https://github.com/DataDog/datadog-api-client-go/pull/486).
* [Added] Add DNS test type to Synthetics. See [#482](https://github.com/DataDog/datadog-api-client-go/pull/482).
* [Added] Add API endpoints for teams and services. See [#470](https://github.com/DataDog/datadog-api-client-go/pull/470).
* [Added] Add mobile_rum_session_count_sum property to usage responses. See [#469](https://github.com/DataDog/datadog-api-client-go/pull/469).
* [Fixed] Fix synthetics_check_id type in MonitorOptions. See [#526](https://github.com/DataDog/datadog-api-client-go/pull/526).
* [Fixed] Remove default for cell_display_mode in table widget. See [#519](https://github.com/DataDog/datadog-api-client-go/pull/519).
* [Fixed] Fix tags attribute type in event aggregation API. See [#463](https://github.com/DataDog/datadog-api-client-go/pull/463).
* [Changed] Change `columns` attribute type from string array to object array in APM stats query widget. See [#509](https://github.com/DataDog/datadog-api-client-go/pull/509).
* [Changed] Rename to ApmStats and add required properties. See [#490](https://github.com/DataDog/datadog-api-client-go/pull/490).
* [Changed] Remove unused `aggregation_key` and `related_event_id` properties from events responses. See [#480](https://github.com/DataDog/datadog-api-client-go/pull/480).
* [Changed] Define required fields for v2 requests. See [#475](https://github.com/DataDog/datadog-api-client-go/pull/475).
* [Changed] Mark required type fields in User and Roles API v2. See [#467](https://github.com/DataDog/datadog-api-client-go/pull/467).
* [Removed] Remove check_type parameter from ListTests endpoint. See [#465](https://github.com/DataDog/datadog-api-client-go/pull/465).

## v1.0.0-beta.8 / 2020-09-16

* [Added] Add `aggregation` and `metric` fields to `SecurityMonitoringRuleQuery`. See [#457](https://github.com/DataDog/datadog-api-client-go/pull/457).
* [Added] Add tracing without limits to usage API. See [#449](https://github.com/DataDog/datadog-api-client-go/pull/449).
* [Added] Add response codes for AWS API. See [#443](https://github.com/DataDog/datadog-api-client-go/pull/443).
* [Added] Add `custom_links` support for Dashboard widgets. See [#442](https://github.com/DataDog/datadog-api-client-go/pull/442).
* [Added] Add profiling to usage API. See [#436](https://github.com/DataDog/datadog-api-client-go/pull/436).
* [Added] Add synthetics CI endpoint. See [#429](https://github.com/DataDog/datadog-api-client-go/pull/429).
* [Added] Add APM resources data source to table widgets. See [#428](https://github.com/DataDog/datadog-api-client-go/pull/428).
* [Added] Add list API for security monitoring signals. See [#424](https://github.com/DataDog/datadog-api-client-go/pull/424).
* [Added] Add create, edit and delete endpoints for synthetics global variables. See [#421](https://github.com/DataDog/datadog-api-client-go/pull/421).
* [Added] Add monitor option `renotify_interval` to synthetics tests. See [#420](https://github.com/DataDog/datadog-api-client-go/pull/420).
* [Added] Add event aggregation v2 API. See [#419](https://github.com/DataDog/datadog-api-client-go/pull/419).
* [Added] Add Profiling Host to Usage endpoint. See [#417](https://github.com/DataDog/datadog-api-client-go/pull/417).
* [Added] Add `distinctFields` to `SecurityMonitoringRuleQuery`. See [#412](https://github.com/DataDog/datadog-api-client-go/pull/412).
* [Added] Add missing `security_query` on `QueryValueWidgetRequest`. See [#407](https://github.com/DataDog/datadog-api-client-go/pull/407).
* [Added] Enable security source for dashboards. See [#403](https://github.com/DataDog/datadog-api-client-go/pull/403).
* [Added] Add SLO alerts to monitor enum. See [#401](https://github.com/DataDog/datadog-api-client-go/pull/401).
* [Fixed] Add 200 response code to PATCH v2 users. See [#441](https://github.com/DataDog/datadog-api-client-go/pull/441).
* [Fixed] Fix hourly host usage descriptions. See [#438](https://github.com/DataDog/datadog-api-client-go/pull/438).
* [Fixed] Remove enum from `legend_size` widget attribute. See [#432](https://github.com/DataDog/datadog-api-client-go/pull/432).
* [Fixed] Fix content-type spelling errors. See [#423](https://github.com/DataDog/datadog-api-client-go/pull/423).
* [Fixed] Properly mark `status` and `query` field as required for creation of Security Monitoring rule. See [#422](https://github.com/DataDog/datadog-api-client-go/pull/422).
* [Fixed] Fix name of `isEnabled` parameter for Security Monitoring rule. See [#409](https://github.com/DataDog/datadog-api-client-go/pull/409).
* [Removed] Remove 204 response from PATCH v2 users. See [#446](https://github.com/DataDog/datadog-api-client-go/pull/446).

## v1.0.0-beta.7 / 2020-07-22

* [Added] Adding four usage attribution endpoints. See [#393](https://github.com/DataDog/datadog-api-client-go/pull/393).
* [Added] Fix documentation for `v1/hosts`. See [#383](https://github.com/DataDog/datadog-api-client-go/pull/383).
* [Changed] Update synthetics test to contain latest features. See [#375](https://github.com/DataDog/datadog-api-client-go/pull/375).
* [Added] Usage Billable Summary response. See [#368](https://github.com/DataDog/datadog-api-client-go/pull/368).
* [Added] Add Logs Search API v2. See [#365](https://github.com/DataDog/datadog-api-client-go/pull/365).
* [Fixed] RRULE property for Downtimes API. See [#364](https://github.com/DataDog/datadog-api-client-go/pull/364).
* [Deprecated] Dashboards List v1 has been deprecated. See [#363](https://github.com/DataDog/datadog-api-client-go/pull/363).

## v1.0.0-beta.6 / 2020-06-19

* [Fixed] Update enum of synthetics devices IDs to match API. See [#351](https://github.com/DataDog/datadog-api-client-go/pull/351).

## v1.0.0-beta.5 / 2020-06-19

* [Added] Update to the latest openapi-generator 5 snapshot. See [#338](https://github.com/DataDog/datadog-api-client-go/pull/338).
* [Added] Add synthetics location endpoint. See [#334](https://github.com/DataDog/datadog-api-client-go/pull/334).
* [Fixed] Widget legend size can also be "0". See [#336](https://github.com/DataDog/datadog-api-client-go/pull/336).
* [Fixed] Log Index as an optional parameter (default to "*") for List Queries. See [#335](https://github.com/DataDog/datadog-api-client-go/pull/335).
* [Changed] Rename payload objects to request for `users` v2 API. See [#346](https://github.com/DataDog/datadog-api-client-go/pull/346).
  * This change includes backwards incompatible changes when using the `users` v2 endpoint.
* [Changed] Split schema for roles API. See [#337](https://github.com/DataDog/datadog-api-client-go/pull/337).
  * This change includes backwards incompatible changes when using the `role` endpoint.

## v1.0.0-beta.4 / 2020-06-09

* [BREAKING] Add missing values to enums. See [#320](https://github.com/DataDog/datadog-api-client-go/pull/320).
    * This change includes backwards incompatible changes when using the `MonitorSummary` widget.
* [BREAKING] Split schemas from DashboardList v2. See [#318](https://github.com/DataDog/datadog-api-client-go/pull/318).
    * This change includes backwards incompatible changes when using corresponding endpoints methods.
* [BREAKING] Clean synthetics test CRUD endpoints. See [#317](https://github.com/DataDog/datadog-api-client-go/pull/317).
    * This change includes backwards incompatible changes when using corresponding endpoints methods.
* [Added] Add Logs Archives endpoints. See [#323](https://github.com/DataDog/datadog-api-client-go/pull/323).

## v1.0.0-beta.3 / 2020-05-21

* [BREAKING] Update to openapi-generator 5.0.0. See [#303](https://github.com/DataDog/datadog-api-client-go/pull/303).
    * This change includes backwards incompatible changes when using structs generated from `oneOf` schemas.
* [Added] Add SIEM and SNMP usage API. See [#309](https://github.com/DataDog/datadog-api-client-go/pull/309).
* [Added] Add security monitoring to clients. See [#304](https://github.com/DataDog/datadog-api-client-go/pull/304).
* [Added] Add /v1/validate endpoint. See [#290](https://github.com/DataDog/datadog-api-client-go/pull/290).
* [Added] Add generated_files file. See [#270](https://github.com/DataDog/datadog-api-client-go/pull/270).
* [Fixed] Add authentication to Go examples. See [#299](https://github.com/DataDog/datadog-api-client-go/pull/299).
* [Fixed] Add 422 error codes to users and roles v2 endpoints. See [#296](https://github.com/DataDog/datadog-api-client-go/pull/296).
* [Fixed] Update import in Go examples. See [#295](https://github.com/DataDog/datadog-api-client-go/pull/295).
* [Fixed] Check duplicate object definitions. See [#288](https://github.com/DataDog/datadog-api-client-go/pull/288).
* [Fixed] Mark unstable endpoints with beta note. See [#281](https://github.com/DataDog/datadog-api-client-go/pull/281).
* [Changed] Update ServiceLevelObjective schema names. See [#279](https://github.com/DataDog/datadog-api-client-go/pull/279).
* [Deprecated] Add deprecated fields `logset`, `count` and `start` to appropriate dashboard widgets. See [#285](https://github.com/DataDog/datadog-api-client-go/pull/285).

## v1.0.0-beta.2 / 2020-05-04

* [Added] Add RUM Monitor Type and update documentation. See [#273](https://github.com/DataDog/datadog-api-client-go/pull/273).
* [Added] Add Logs Pipeline Processor. See [#268](https://github.com/DataDog/datadog-api-client-go/pull/268).
* [Added] Add additional fields to synthetics test request. See [#262](https://github.com/DataDog/datadog-api-client-go/pull/262).
* [Added] Add Monitor Pagination. See [#253](https://github.com/DataDog/datadog-api-client-go/pull/253).
* [Fixed] Mark synthetics test request "method" and "url" as optional. See [#265](https://github.com/DataDog/datadog-api-client-go/pull/265).
* [Fixed] Update error responses for roles v2 endpoints. See [#248](https://github.com/DataDog/datadog-api-client-go/pull/248).
* [Fixed] Add missing ListSLO's 404 response. See [#245](https://github.com/DataDog/datadog-api-client-go/pull/245).
* [Removed] Remove Pagerduty endpoints from the client. See [#264](https://github.com/DataDog/datadog-api-client-go/pull/264).

## 1.0.0-beta.1 / 2020-04-22

* [Added] Initial beta release of the Datadog API Client
