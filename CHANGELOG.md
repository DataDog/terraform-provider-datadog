## 3.84.0 (January 8, 2026)

### BUGFIXES
* [datadog_software_catalog] Fix software_catalog resource to handle empty API responses for deleted entities by @KruthiVuppala in https://github.com/DataDog/terraform-provider-datadog/pull/3371
* [datadog_csm_threat_agent] fix multiple actions support by @safchain in https://github.com/DataDog/terraform-provider-datadog/pull/3375
* [datadog_security_monitoring_rule] Replace security monitoring rules with an updated detection method by @paul-hf in https://github.com/DataDog/terraform-provider-datadog/pull/3392
### IMPROVEMENTS
* [datadog_csm_threats_agent_rule] add silent attribute to rules by @safchain in https://github.com/DataDog/terraform-provider-datadog/pull/3380
### FEATURES
* [datadog_observability_pipeline] Add processor groups instead of standalone processors by @vladimir-dd in https://github.com/DataDog/terraform-provider-datadog/pull/3346
* [datadog_deployment_gate] Add deployment gates terraform provider by @Aaron-9900 in https://github.com/DataDog/terraform-provider-datadog/pull/3286
* [datadog_dashboard] Add semantic_mode support to metric queries by @brtu in https://github.com/DataDog/terraform-provider-datadog/pull/3356
* [datadog_observability_pipeline] Add `display_name` for processors and processor groups by @vladimir-dd in https://github.com/DataDog/terraform-provider-datadog/pull/3377
* [datadog_security_monitoring_rule] Add anomaly detection options to security monitoring rule resource by @paul-hf in https://github.com/DataDog/terraform-provider-datadog/pull/3376
* [datadog_security_monitoring_rule] Add instantaneous baseline option to NewValue by @hachem-dd in https://github.com/DataDog/terraform-provider-datadog/pull/3400
### NOTES
* Remove note about log index deletion not being supported by @kelly-kong-ddog in https://github.com/DataDog/terraform-provider-datadog/pull/3373

## New Contributors
* @KruthiVuppala made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3371
* @Aaron-9900 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3286
* @kelly-kong-ddog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3373
* @hachem-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3400

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.83.0...v3.84.0

## 3.83.0 (December 18, 2025)

### BUGFIXES
* [datadog_security_monitoring] Fix how tags and calculated fields array are handled by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/3347
* [datadog_security_monitoring] Fix missing group_signals_by in rules by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/3365
### IMPROVEMENTS
* [datadog_sensitive_data_scanner_rule] drop ForceNew from included_keyword_configuration by @OBe95 in https://github.com/DataDog/terraform-provider-datadog/pull/3355
* [datadog_cost_budget] Add Validation for terraform plan by @zeinageb in https://github.com/DataDog/terraform-provider-datadog/pull/3349
### FEATURES
* [datadog_integration_aws_iam_permissions_resource_collection] Add new data source for AWS Integration resource collection IAM permissions by @raymondeah in https://github.com/DataDog/terraform-provider-datadog/pull/3265
* [datadog_monitor_notification_rule] Add support for scope and conditional_recipients by @shaneyuandd in https://github.com/DataDog/terraform-provider-datadog/pull/3363
* [datadog_resource_datadog_monitor] Add Support for Saving Assets to Monitors by @kevinpombo-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/3334
* [datadog_logs_restriction_query] Add LRQ Terraform Resource by @dwijetunga in https://github.com/DataDog/terraform-provider-datadog/pull/3352

## New Contributors
* @OBe95 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3355
* @dwijetunga made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3352

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.82.0...v3.83.0

## 3.82.0 (December 10, 2025)

### BUGFIXES
* [datadog_cost_budget] Fix update creating duplicate instead of updating existing by @zeinageb in https://github.com/DataDog/terraform-provider-datadog/pull/3341
### FEATURES
* [datadog_custom_allocation_rules] add the functionality to allow users decide custom allocation rule reorder behaviors by @congx4 in https://github.com/DataDog/terraform-provider-datadog/pull/3324
* [datadog_reference_table] Add terraform support for reference tables by @g-brizolier in https://github.com/DataDog/terraform-provider-datadog/pull/3317
* [datadog_team_hierarchy_links] Add TF provider for team hierarchy links by @dshivara in https://github.com/DataDog/terraform-provider-datadog/pull/3348

## New Contributors
* @zeinageb made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3341
* @dshivara made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3358
* @g-brizolier made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3317

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.81.0...v3.82.0

## 3.81.0 (November 20, 2025)

### BUGFIXES
* [datadog_agentless_scanning_gcp_scan_options] Better support of resource conflict in agentless GCP scan options TF provider by @mohamed-challal in https://github.com/DataDog/terraform-provider-datadog/pull/3328
* [datadog_security_monitoring_suppression] Fix issue when dealing with empty array of tags in Rules by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/3331
* [datadog_observability_pipeline] make google auth optional by @vladimir-dd in https://github.com/DataDog/terraform-provider-datadog/pull/3332
### FEATURES
* [datadog_tag_pipeline_rulesets] add the functionality to allow users decide tag pipeline reorder behaviors by @congx4 in https://github.com/DataDog/terraform-provider-datadog/pull/3320
* [datadog_user] Add new exclude_service_accounts attribute to the datadog_user ds by @tquidetdd in https://github.com/DataDog/terraform-provider-datadog/pull/3323
* [datadog_metrics] add 'datadog_metrics' data source by @mdb in https://github.com/DataDog/terraform-provider-datadog/pull/3214

## New Contributors
* @mfu00 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3330
* @mgaonach made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3327

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.80.0...v3.81.0

## 3.80.0 (November 12, 2025)

### BUGFIXES
* [datadog_software_catalog] Allow backstage imports in software_catalog resource by @christinexlee in https://github.com/DataDog/terraform-provider-datadog/pull/3319
* [datadog_monitor] Make cloud cost query `aggregator` field required by @fpighi in https://github.com/DataDog/terraform-provider-datadog/pull/3318
### FEATURES
* [datadog_agentless_scanning_gcp_scan_options] Add Terraform provider for GCP scan options by @mohamed-challal in https://github.com/DataDog/terraform-provider-datadog/pull/3321
* [datadog_integration_gcp_sts] Add granular namespace filters to GCP V2 Terraform Provider by @ash-ddog in https://github.com/DataDog/terraform-provider-datadog/pull/3299 - NOTE: due to limitations with the Terraform protocol v5 currently in use by our provider, users should add `filters = []` to each `metric_namespace_configs` key. Omitting it would result in an error. For more information on `filters`, see [Integrations - Google Cloud Platform](https://docs.datadoghq.com/integrations/google-cloud-platform/?tab=organdfolderlevelprojectdiscovery#configuration)

## New Contributors
* @mohamed-challal made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3321

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.79.0...v3.80.0

## 3.79.0 (November 3, 2025)

### BUGFIXES
* [datadog_integration_gcp_sts] Fix handling prometheus default in mncs by @katherinekim-51 in https://github.com/DataDog/terraform-provider-datadog/pull/3297
* [datadog_tag_pipeline_ruleset] Ensure resource handles tag pipeline rulesets deleted via the UI that are managed in tfstate by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3313
* [datadog_synthetics_private_location] Fix restriction_policy_resource_id format by @Matt-Cam in https://github.com/DataDog/terraform-provider-datadog/pull/3310
* [datadog_app_builder_app] fix bug with publish/unpublish by @tyffical in https://github.com/DataDog/terraform-provider-datadog/pull/3312
### IMPROVEMENTS
* [scripts] Update Cloud Cost import existing resources script  by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3309
* [datadog_logs_pipelines] Add support for schema processor by @seohyunh in https://github.com/DataDog/terraform-provider-datadog/pull/3289

## New Contributors
* @katherinekim-51 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3297
* @sean-r-williams made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3280
* @seohyunh made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3289

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.78.0...v3.79.0

## 3.78.0 (October 29, 2025)

### BUGFIXES
* [datadog_dashboards] Fix type error in unitScale by @brlee19 in https://github.com/DataDog/terraform-provider-datadog/pull/3238
* [datadog_csm_threats] Fix tags field in csm_threat_policy resource by @QuentinGuillard in https://github.com/DataDog/terraform-provider-datadog/pull/3298
* [datadog_monitor] Fix monitor validation on default tags by @shaneyuandd in https://github.com/DataDog/terraform-provider-datadog/pull/3296
* [datadog_application_key] Update docs for application key by @wangwillson1 in https://github.com/DataDog/terraform-provider-datadog/pull/3305
### IMPROVEMENTS
* [datdog_csm_threats_agent_rule] add missing set action fields by @safchain in https://github.com/DataDog/terraform-provider-datadog/pull/3295

## New Contributors
* @brlee19 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3238
* @safchain made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3295

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.77.0...v3.78.0

## 3.77.0 (October 21, 2025)

### BUGFIXES
* [datadog_synthetics_test] Allow deletion of Synthetics test steps by @ALenfant in https://github.com/DataDog/terraform-provider-datadog/pull/3284
* [datadog_custom_allocation_rule] update terraform go api-client by @congx4 in https://github.com/DataDog/terraform-provider-datadog/pull/3283
* [datadog_aws_cur_config][datadog_custom_allocation_rule] Minor fixes to Cloud Cost resources to prevent apply needed after imports by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3287
### FEATURES
* [datadog_synthetics_test] Support the `blockedRequestPatterns` option for browser tests by @SKYWy in https://github.com/DataDog/terraform-provider-datadog/pull/3257
* [datadog_monitor] Update Monitors terraform to support draft monitors by @kevinpombo-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/3275
* [scripts] Create helper script to automatically find and import any existing Datadog Cloud Cost resources into terraform by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3278

## New Contributors
* @ALenfant made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3284
* @SKYWy made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3257
* @kevinpombo-datadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3275

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.76.0...v3.77.0

## 3.76.0 (October 15, 2025)

### BUGFIXES
* [datadog_azure_uc_config] Add support for Cloud Cost Azure Cloud Usage Cost Config resource and data source by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3263
* [datadog_csm_threats] Fix csm threats agent rule expression whitespace inconsistency by @QuentinGuillard in https://github.com/DataDog/terraform-provider-datadog/pull/3269
* [datadog_tag_pipeline_ruleset] Fix an edge case when imports tag pipeline rulesets without any rules by @congx4 in https://github.com/DataDog/terraform-provider-datadog/pull/3277
### IMPROVEMENTS
* [datadog_synthetics_private_location] Add restriction_policy_resource_id to synthetics private locations to use with restriction policy by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/3271
* [datadog_azure_uc_config][datadog_aws_cur_config][datadog_gcp_uc_config] Update Cloud cost provider resources to support exposing computed fields by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3274
### FEATURES
* [datadog_security_monitoring_rule] Add sequence detection method to security monitoring rule resource by @paul-hf in https://github.com/DataDog/terraform-provider-datadog/pull/3239
* [datadog_gcp_uc_config] Add support for Cloud Cost Google Cloud Usage Cost Config resource and data source by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3258
* [datadog_integration_aws_iam_permissions_standard] Add new data source for AWS Integration standard IAM permissions by @raymondeah in https://github.com/DataDog/terraform-provider-datadog/pull/3261
* [datadog_tag_pipeline_ruleset] terraform support for tag-pipeline by @congx4 in https://github.com/DataDog/terraform-provider-datadog/pull/3236
* [datadog_custom_allocation_rule] Add support for CCM custom allocation rules by @elihschiff in https://github.com/DataDog/terraform-provider-datadog/pull/3262

## New Contributors
* @congx4 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3236
* @elihschiff made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3262

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.75.0...v3.76.0

## 3.75.0 (October 1, 2025)

### BUGFIXES
* [datadog_synthetics_test] Fix typo in browser test steps press key modifiers validation by @zach-b in https://github.com/DataDog/terraform-provider-datadog/pull/3234
* [datadog_security_monitoring_rule] Defer JSON filtering to Read to fix “inconsistent result after apply” by @Datadog-MartinG in https://github.com/DataDog/terraform-provider-datadog/pull/3212
* [datadog_synthetics_global_variable] Fix dynamic block parsing by @Matt-Cam in https://github.com/DataDog/terraform-provider-datadog/pull/3251
### IMPROVEMENTS
* [datadog_rum_application] Expose RUM application API key ID by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/3250
### FEATURES
* [datadog_org_connection] Terraform Resource for Org Connection  by @abe-lin-dd in https://github.com/DataDog/terraform-provider-datadog/pull/3229
* [datadog_observability_pipeline] add Google PubSub destination by @gwenaskell in https://github.com/DataDog/terraform-provider-datadog/pull/3235
* [datadog_integration_aws_external_id] Add data source to retrieve AWS external ID of an AWS integration by @LorisFriedel in https://github.com/DataDog/terraform-provider-datadog/pull/3247
* [datadog_aws_cur_config] Add support for Cloud Cost AWS CUR Config resource and datasource by @alextmeyer in https://github.com/DataDog/terraform-provider-datadog/pull/3252

## New Contributors
* @abe-lin-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3229
* @LiuVII made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3248
* @alextmeyer made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3252

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.74.0...v3.75.0

## 3.74.0 (September 17, 2025)

### BUGFIXES
* [datadog_security_monitoring_rule] Handle empty vs null for security monitoring reference table deletion by @paul-hf in https://github.com/DataDog/terraform-provider-datadog/pull/3192
* [datadog_metric_tags] fix: data source incorrect metric name by @bharling in https://github.com/DataDog/terraform-provider-datadog/pull/3220
* [datadog_team_membership] Remove resource from state if not found by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/3228
### IMPROVEMENTS
* [datadog_security_monitoring_suppression] Add Suppressions Validation during TF Plan by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/3203
* [datadog_synthetics_test] Keep Synthetics multistep test steps id in state by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/3204
### FEATURES
* [datadog_incident_notification_template] Add datadog_incident_notification_template resource and datasource by @codhah92 in https://github.com/DataDog/terraform-provider-datadog/pull/3218
* [datadog_incident_notification_rule] Add incident notification rule tf resource & data source by @codhah92 in https://github.com/DataDog/terraform-provider-datadog/pull/3224
* [datadog_monitor] Migrate datadog_monitor to plugin framework by @shaneyuandd in https://github.com/DataDog/terraform-provider-datadog/pull/3161
* [datadog_integration_gcp] [datadog_integration_gcp_sts] Add monitored_resource_configs to v1 and v2 terraform modules by @ash-ddog in https://github.com/DataDog/terraform-provider-datadog/pull/3219
* [datadog_security_monitoring_rule] Support Cloud SIEM scheduled rules by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/3181

## New Contributors
* @Datadog-MartinG made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3208
* @ash-ddog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3219
* @Matt-Cam made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3217

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.73.0...v3.74.0

## 3.73.0 (September 3, 2025)

### BUGFIXES
* [datadog_service_account] Add pagination support for datadog_service_account data by @shuheiktgw in https://github.com/DataDog/terraform-provider-datadog/pull/3172
### IMPROVEMENTS
* [datadog_synthetics_test] Make port in synthetic tests support both int and string by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/3194
### FEATURES
* [datadog_integration_aws_account] Add support for log_source_config by @ktmq in https://github.com/DataDog/terraform-provider-datadog/pull/3173
* [datadog_metric_active_tags_and_aggregations] add `datadog_metric_active_tags_and_aggregations` data source by @mdb in https://github.com/DataDog/terraform-provider-datadog/pull/3103
* [datadog_integration_aws_account] Disable AWS/Usage metric collection by default by @joaquinrios in https://github.com/DataDog/terraform-provider-datadog/pull/3156
* [datadog_rum_application] Add Product Scales support by @jasonscottdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/3185
* [datadog_sensitive_data_scanner_group] Add support for SDS `should_save_match` by @fuchsnj in https://github.com/DataDog/terraform-provider-datadog/pull/3189

## New Contributors
* @joaquinrios made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3156
* @tquidetdd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3182
* @jasonscottdatadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3185
* @fuchsnj made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3189
* @diogocp made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3200

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.72.0...v3.73.0

## 3.72.0 (August 20, 2025)

### BUGFIXES
* [workflow_automation] Clean up state if workflow not found by @GabeWM in https://github.com/DataDog/terraform-provider-datadog/pull/3171
### FEATURES
* [datadog_observability_pipeline] Add v 2.6.0 components, plus Amazon S3 and Crowdstrike destinations by @tessneau in https://github.com/DataDog/terraform-provider-datadog/pull/3166
* [datadog_integration_aws_iam_permissions] Add new data source for AWS Integration required IAM permissions by @raymondeah in https://github.com/DataDog/terraform-provider-datadog/pull/3178
* [dataset] Dataset terraform resource by @kisharne in https://github.com/DataDog/terraform-provider-datadog/pull/3138

## New Contributors
* @kisharne made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3138

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.71.0...v3.72.0

## v3.71.0 (August 13, 2025)

### BUGFIXES
* [datadog_metric_tags] chore: correct `datadogMetricTagsDataSource` interface check by @mdb in https://github.com/DataDog/terraform-provider-datadog/pull/3098
### IMPROVEMENTS
* [datadog_log_custom_pipeline] Add support for decoder processor by @pavel-kim-dd in https://github.com/DataDog/terraform-provider-datadog/pull/3125
### FEATURES
* [datadog_security_notification_rule] Add Host and IaC security notification rule types by @ltaloc in https://github.com/DataDog/terraform-provider-datadog/pull/3158
* [datadog_metric_metadata] feat: add `datadog_metric_metadata` data source by @mdb in https://github.com/DataDog/terraform-provider-datadog/pull/3100
* [datadog_agentless_scanning_aws_scan_options] Add Terraform provider for AWS scan options by @k3nz0 in https://github.com/DataDog/terraform-provider-datadog/pull/3164

## New Contributors
* @ikraemer-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3152
* @ltaloc made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3158
* @pavel-kim-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3125
* @k3nz0 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3164

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.70.0...v3.71.0

## 3.70.0 (August 6, 2025)

### BUGFIXES
* [datadog_team_membership] Correct handling of 404 when reading team memberships by @heavi5ide in https://github.com/DataDog/terraform-provider-datadog/pull/3140
### FEATURES
* [authentication] Add ability for terraform-provider to use AWS cloud-based authentication by @wynbennett in https://github.com/DataDog/terraform-provider-datadog/pull/3051
* [datadog_incident_type] Add resouce and data source by @codhah92 in https://github.com/DataDog/terraform-provider-datadog/pull/3147

## New Contributors
* @larrys made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3127
* @heavi5ide made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3140
* @wynbennett made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3051
* @codhah92 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3147

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.69.0...v3.70.0

## 3.69.0 (July 23, 2025)

### BUGFIXES
* [synthetics_test] Make gRPC `message` required by @mtalec in https://github.com/DataDog/terraform-provider-datadog/pull/3126
### IMPROVEMENTS
* [datadog_downtime_schedule] Validate recurrence start by @gmeligio in https://github.com/DataDog/terraform-provider-datadog/pull/3096
* [datadog_synthetics_test] Add QoL feedback when converting multistep API test by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/3049

## New Contributors
* @gmeligio made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3096

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.68.0...v3.69.0

## 3.68.0 (July 18, 2025)

### BUGFIXES
* [synthetics_test] Fix file upload issue in the terraform provider by @mtalec in https://github.com/DataDog/terraform-provider-datadog/pull/3118
### IMPROVEMENTS
* [datadog_synthetics_test] Rewrite all `log.Printf` into `diags` by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/3066
* [datadog_logs_pipelines] Updates Datadog Go API client to v2.43.0 by @LorisFriedel in https://github.com/DataDog/terraform-provider-datadog/pull/3119
* [datadog_logs_pipelines] Add support for Logs Array Processor in Logs Pipelines by @LorisFriedel in https://github.com/DataDog/terraform-provider-datadog/pull/3120
### FEATURES
* [app_key_registration] Add App Key Registration Resource  by @GabeWM in https://github.com/DataDog/terraform-provider-datadog/pull/3054
* [datadog_synthetics_test] Add `disableAiaIntermediateFetching` option by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/3121

## New Contributors
* @iGranDav made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3110

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.67.0...v3.68.0

## 3.67.0 (July 9, 2025)

### BUGFIXES
* [datadog_synthetics_test] Fix basic auth by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/3062
* [datadog_synthetics_test] Make request client certificate content optional by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/3063
* [datadog_synthetics_test] Support `dns_server_port` as string or number by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/3070
* [datadog_synthetics_test] Fix provider crash due to absent 'relative_position' in config by @Alexis-Jacob-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/3067
* [datadog_synthetics_test] update api client to ensure the `message` field is set for mobile tests by @zach-b in https://github.com/DataDog/terraform-provider-datadog/pull/3082
* [datadog_on_call] update validation and documentation by @bndbsh in https://github.com/DataDog/terraform-provider-datadog/pull/3077
* [datadog_on_call_escalation_policy] correctly set retries on update by @bndbsh in https://github.com/DataDog/terraform-provider-datadog/pull/3078
* [datadog_on_call_escalation_policy] fix validation for assignment by @bndbsh in https://github.com/DataDog/terraform-provider-datadog/pull/3085
### FEATURES
* [datadog_logs_custom_destination] Support Microsoft Sentinel Custom Destinations by @mgrunenberger in https://github.com/DataDog/terraform-provider-datadog/pull/3061
* [datadog_sensitive_data_scanner_group] Handle sampling rate in scanning group by @isabella-garza-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/3076
* [datadog_synthetics_test] Add extract body email step  by @Alexis-Jacob-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/3091
* [datadog_synthetics_test] Add append_to_content option for text input by @Alexis-Jacob-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/3106

## New Contributors
* @mgrunenberger made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3061
* @zach-b made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3082
* @SimoneT-DD made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3079
* @isabella-garza-datadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3076
* @l-teles made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3023

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.66.0...v3.67.0

## 3.66.0 (June 17, 2025)

### BUGFIXES
* [datadog_synthetics_test] Exclude TLS version types from being converted to number by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/3042
### IMPROVEMENTS
* [datadog_monitor] add opt-in behavior on restricted_roles by @shaneyuandd in https://github.com/DataDog/terraform-provider-datadog/pull/3017
### FEATURES
* [datadog_compliance_resource_evaluation_filter] Adding resource evaluation filters as a terraform resource by @Matzoc in https://github.com/DataDog/terraform-provider-datadog/pull/3004
* [datadog_cost_budget] add support for cost budgets by @klara-kr in https://github.com/DataDog/terraform-provider-datadog/pull/3001
* [datadog_csm_threats] supporting cws multi-policy in terraform by @homoeconomics in https://github.com/DataDog/terraform-provider-datadog/pull/2681
* [datadog_synthetics_test] Add `notification_preset_name` monitor option for browser & api tests by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/3044
* [synthetics_test] Add private location api key field by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/3053
* [datadog_synthetics_test] Add support for all api subtypes for multistep tests by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/3041
* [datadog_synthetics_test] Add `form` for `multipart/form-data` API tests by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/3048
### NOTES
* [Chore] Update the dependencies by @Supam in https://github.com/DataDog/terraform-provider-datadog/pull/3037

## New Contributors
* @Matzoc made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3004
* @homoeconomics made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2681
* @sarrazaghbib-dev made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3047

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.65.0...v3.66.0

## 3.65.0 (May 30, 2025)

### FEATURES
* [datadog_compliance_custom_framework] Terraform Provider for Custom Frameworks by @nkonjeti in https://github.com/DataDog/terraform-provider-datadog/pull/2975

## New Contributors
* @nkonjeti made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2975

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.64.0...v3.65.0

## 3.64.0 (May 28, 2025)

### BUGFIXES
* [datadog_synthetics_test] Change target and targetvalue to be number or string by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/3027
### IMPROVEMENTS
* [datadog_dashboard] List Stream Widget Request - make column fields optional by @brtu in https://github.com/DataDog/terraform-provider-datadog/pull/3026
### FEATURES
* [datadog_on_call] Add On-Call resources by @bndbsh in https://github.com/DataDog/terraform-provider-datadog/pull/3029


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.63.0...v3.64.0

## 3.63.0 (May 20, 2025)

### BUGFIXES
* [datadog_monitors] Fix `datadog_monitors` datasource not failing if no monitors found by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2808
### IMPROVEMENTS
* [datadog_logs_index] Add delete logs indexes support in terraform by @szakon in https://github.com/DataDog/terraform-provider-datadog/pull/2885
### FEATURES
* [datadog_synthetics_test] Add `click_with_javascript` for browser tests by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/3013
* [datadog_service_level_objective] Add support for default tag by @Supam in https://github.com/DataDog/terraform-provider-datadog/pull/2972

## New Contributors
* @szakon made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2885

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.62.0...v3.63.0

## 3.62.0 (May 14, 2025)

### FEATURES
* [datadog_observability_pipeline] Added the remaining Observability Pipelines components by @vladimir-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2988
* [datadog_dashboards] [datadog_powerpacks] add custom_links support to treemap widgets by @dashashifrina in https://github.com/DataDog/terraform-provider-datadog/pull/2923
* [datadog_integration_gcp_sts] Add IsPerProjectQuotaEnabled field by @tedkahwaji in https://github.com/DataDog/terraform-provider-datadog/pull/2999
* [datadog_security_monitoring_rule_json] Add security_monitoring_rule_json resource by @sabraham2 in https://github.com/DataDog/terraform-provider-datadog/pull/2985

## New Contributors
* @sabraham2 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2985
* @bndbsh made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/3007

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.61.0...v3.62.0

## 3.61.0 (April 30, 2025)

### BUGFIXES
* [datadog_spans_metric] Fix validation with unintended side effects by @Supam in https://github.com/DataDog/terraform-provider-datadog/pull/2983
### FEATURES
* [synthetics_test] add escalation message parameter as a monitor option in the terraform provider as a monitor option  by @mtalec in https://github.com/DataDog/terraform-provider-datadog/pull/2979
* [datadog_metric_tags] add `datadog_metric_tags` data source by @mdb in https://github.com/DataDog/terraform-provider-datadog/pull/2966
* [datadog_monitor_notification_rule] Add support for Monitor Notification Rule by @shaneyuandd in https://github.com/DataDog/terraform-provider-datadog/pull/2980
* [datadog_security_monitoring_rule] Add default tag support by @Supam in https://github.com/DataDog/terraform-provider-datadog/pull/2968

## New Contributors
* @mtalec made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2979
* @aleks-dog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2978
* @mdb made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2966
* @nguyening made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2986
* @shaneyuandd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2980

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.60.1...v3.61.0

## 3.60.1 (April 17, 2025)

### IMPROVEMENTS
* [slo_query_type] bump datadog-api-client-go v2.37.1 by @pablito-perez in https://github.com/DataDog/terraform-provider-datadog/pull/2967


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.60.0...v3.60.1

## 3.60.0 (April 16, 2025)

### BUGFIXES
* [datadog_synthetics_private_location] Fix error messages by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2955
### IMPROVEMENTS
* [datadog_integration_slack_channel] Add mute_button argument by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2950
### FEATURES
* [datadog_observability_pipeline] Add Observability Pipeline resource by @vladimir-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2945

## New Contributors
* @vladimir-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2945
* @daniel-m-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2958
* @haoran-shao-datadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2931

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.59.1...v3.60.0

## 3.59.1 (April 4, 2025)

### BUGFIXES
* [datadog_synthetics_tests] Fix encoding bug on plain_proto_file / compressedProtoFile by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2947


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.59.0...v3.59.1

## 3.59.0 (April 2, 2025)

### BUGFIXES
* [datadog_metric_metadata] Fix drift on imported resources by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2922
* [datadog_spans_metric] Suppress unknown include_percentiles state by @fpighi in https://github.com/DataDog/terraform-provider-datadog/pull/2906
* [datadog_azure_integration] Fix import command by @agulen in https://github.com/DataDog/terraform-provider-datadog/pull/2924
* [datadog_rum_metric] Set `tag_name` parameter as computed by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2934
* [datadog_rum_metric] Fix issue with rum metric resource. by @hemeryar in https://github.com/DataDog/terraform-provider-datadog/pull/2936
### FEATURES
* [datadog_logs_archive] Added storage class to s3 archive terraform provider by @jcausey8 in https://github.com/DataDog/terraform-provider-datadog/pull/2875
* [datadog_app_builder_app] Add App Builder Terraform Support by @tyffical in https://github.com/DataDog/terraform-provider-datadog/pull/2723
* [data_source_datadog_software_catalog] Adding Software Catalog data source by @cmcmacs in https://github.com/DataDog/terraform-provider-datadog/pull/2822
* [datadog_rum_retention_filter] Add support for the resource. by @kwangjunkim in https://github.com/DataDog/terraform-provider-datadog/pull/2899

## New Contributors
* @agulen made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2924
* @jcausey8 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2875
* @tyffical made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2723
* @cmcmacs made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2822
* @kwangjunkim made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2899

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.58.0...v3.59.0

## 3.58.0 (March 18, 2025)

### BUGFIXES
* [datadog_synthetics_private_location] Fix unremovable metadata bug by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2892
* [datadog_webhook] Handle unknown encode_as on update by @fpighi in https://github.com/DataDog/terraform-provider-datadog/pull/2907
### IMPROVEMENTS
* [datadog_synthetics_private_location] Migrate synthetics resource to framework by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2881
* [datadog_synthetics_global_variable] Migrate synthetics resource and data_source to framework by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2886
### FEATURES
* [datadog_integration_aws_external_id] Add new resource to generate an AWS Integration external ID by @raymondeah in https://github.com/DataDog/terraform-provider-datadog/pull/2880
* [datadog_appsec] introduce the Application Security resources by @nizox in https://github.com/DataDog/terraform-provider-datadog/pull/2874
* [datadog_workflow_automation] Add Workflows to Terraform by @GabeWM in https://github.com/DataDog/terraform-provider-datadog/pull/2833
* [datadog_security_monitoring_rule] Add group_signals_by and case actions by @Moutix in https://github.com/DataDog/terraform-provider-datadog/pull/2895
* [datadog_synthetics_test] Add `extractedValuesFromScript` to multistep API tests by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2894
### NOTES
* [Dependencies] Bump golang.org/x/net from 0.23.0 to 0.36.0 by @dependabot in https://github.com/DataDog/terraform-provider-datadog/pull/2902

## New Contributors
* @raymondeah made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2880
* @janine-c made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2882
* @RaphaelAllier made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2887
* @GabeWM made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2833
* @Moutix made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2895
* @fpighi made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2914

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.57.0...v3.58.0

## 3.57.0 (February 28, 2025)

### BUGFIXES
* [datadog_synthetics_test] Only set renotify_occurrences if renotify_interval is set  by @Alexis-Jacob-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/2871
### IMPROVEMENTS
* [datadog_dashboard] Handles `cell_display_mode_options` in Query Table widget by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2873
* [datadog_dashboard] Add `clustering_pattern_field_path` and `group_by` for ListStream Widget by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2869
### FEATURES
* [datadog_action_connection] Support Action Connection API by @embeaken in https://github.com/DataDog/terraform-provider-datadog/pull/2718
* [datadog_integration_ms_teams_microsoft_workflows_webhook_handle] Added Resource by @kblackdd in https://github.com/DataDog/terraform-provider-datadog/pull/2849

## New Contributors
* @embeaken made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2718

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.56.0...v3.57.0

## 3.56.0 (February 25, 2025)

### BUGFIXES
* [datadog_spans_metric] Fix count aggregation by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2860
* [datadog_api_key] Fix bug introduced in 3.55.0 for organisation that doesn't have Remove config enabled by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2857
### IMPROVEMENTS
* [synthetics_global_variable] Add is_fido by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2836
* [datadog_user] Add computed attributes to `datadog_user` datasource by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2787
* [datadog_logs_custom_pipeline] Add support for default tags in log pipelines by @shuheiktgw in https://github.com/DataDog/terraform-provider-datadog/pull/2845
* [datadog_dashboard] Add dashboard unit support by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2858
* [datadog_logs_custom_pipeline] Manage `span-id-remapper` processor by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2861
* [datadog_dashboard] Add `scaling` argument for style of `toplist_definition` widget by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2862
### NOTES
* [datadog_openapi_api] Deprecate openapi resource by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/2863

## New Contributors
* @shuheiktgw made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2845

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.55.0...v3.56.0

## 3.55.0 (February 18, 2025)

### BUGFIXES
* [datadog_user] Fix the `name` field as Optional + Computed by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2784
* [datadog_synthetics_test] Fix browser request_client_certificate by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2840
### IMPROVEMENTS
* [datadog_users] Add attributes to `datadog_users` datasource by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2786
### FEATURES
* [datadog_synthetics_test] Add secure field for extractFromJavascript browser step variables by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2818
* [datadog_teams] Add new datasource  by @Supam in https://github.com/DataDog/terraform-provider-datadog/pull/2789
* [datadog_security_notification_rule] Add notification rule terraform provider by @FredMoster in https://github.com/DataDog/terraform-provider-datadog/pull/2806
* [datadog_api_key] Support `remote_config_read_enabled` argument by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2831
* [datadog_synthetics_global_variable] Add is_totp by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2832
* [datadog_synthetics_test] Add renotify occurrences option by @Alexis-Jacob-datadog in https://github.com/DataDog/terraform-provider-datadog/pull/2820

## New Contributors
* @Supam made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2809
* @FredMoster made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2806
* @Alexis-Jacob-datadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2820

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.54.0...v3.55.0

## 3.54.0 (February 6, 2025)

### BUGFIXES
* [datadog_spans_metric] Mark `include_percentiles` property as computed by @VenelinMartinov in https://github.com/DataDog/terraform-provider-datadog/pull/2707
* [datadog_service_account] Fix : `roles` attribute is now always updated by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2761
* [datadog_service_definition_yaml] Fix panic when contact attributes exists but are empty (nil) by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2802
* [datadog_synthetics_global_variable] Remove drift on `option` argument by @ecdatadog in https://github.com/DataDog/terraform-provider-datadog/pull/2768
### IMPROVEMENTS
* [datadog_logs_archives] Handle encryption field for S3 destinations for Logs Archives by @michelledeng30 in https://github.com/DataDog/terraform-provider-datadog/pull/2740
### FEATURES
* [datadog_application_key] Support scoped app key management for service accounts by @wangwillson1 in https://github.com/DataDog/terraform-provider-datadog/pull/2770
* [datadog_monitor] Add cost alert monitors by @klara-kr in https://github.com/DataDog/terraform-provider-datadog/pull/2769
* [datadog_logs_custom_pipeline] Add tags and description to logs pipelines by @joboccara in https://github.com/DataDog/terraform-provider-datadog/pull/2773
### NOTES
* Bump datadog-api-client to v2.35.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2810

## New Contributors
* @klara-kr made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2769
* @joboccara made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2773
* @michelledeng30 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2740
* @VenelinMartinov made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2707
* @ecdatadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2761

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.53.0...v3.54.0

## 3.53.0 (January 23, 2025)

### BUGFIXES
* [datadog_azure_integration] change default value of `resource_provider_configs` and rerecord tests by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2783
* [datadog_synthetics_test] Fix steps reordering without `local_key` by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2795
### FEATURES
* [datadog_security_monitoring_suppression] Add Start to Suppressions by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2780
* [datadog_azure_integration] Support metric filtering by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2781


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.52.1...v3.53.0

## 3.52.1 (January 16, 2025)

### BUGFIXES
* [data_datadog_application_key] remove scopes from schema by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2774


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.52.0...v3.52.1

## 3.52.0 (January 14, 2025)

### BUGFIXES
* [datadog_application_key] Use plan modifiers for app keys by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2763
* [datadog_api_key] Use plan modifier for API keys. by @tyjet in https://github.com/DataDog/terraform-provider-datadog/pull/2765
### FEATURES
* [datadog_application_key] Support scoped app key management by @wangwillson1 in https://github.com/DataDog/terraform-provider-datadog/pull/2760


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.51.0...v3.52.0

## 3.51.0 (January 8, 2025)

### BUGFIXES
* [provider] Prevent the experimental default_tags field from deduplicating tags by @Amaury-Behague in https://github.com/DataDog/terraform-provider-datadog/pull/2728
* [datadog_synthetics_test] Prevent updating `files` with the backend response to keep the plan stable by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2731
* [datadog_webhook] fix import  by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2742
* [datadog_integration_aws_account] Fix panic when resources_config is omitted by @ktmq in https://github.com/DataDog/terraform-provider-datadog/pull/2749
* [datadog_synthetics_test] update the step params when reordering steps by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2741
### IMPROVEMENTS
* [datadog_synthetics_test] Add step public id in the state by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2744
* [resource_datadog_role] ACCESS-2867 - Support `default_permissions_opt_out` attribute in Terraform by @wangwillson1 in https://github.com/DataDog/terraform-provider-datadog/pull/2710
### FEATURES
* [datadog_logs_pipelines_order] Add datasource for Logs Pipelines Order by @LorisFriedel in https://github.com/DataDog/terraform-provider-datadog/pull/2732

## New Contributors
* @LorisFriedel made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2732
* @gpalmz made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2738

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.50.0...v3.51.0

## 3.50.0 (December 17, 2024)

### BUGFIXES
* [datadog_monitor] Allow default_tags values to contain colons. by @Amaury-Behague in https://github.com/DataDog/terraform-provider-datadog/pull/2703
* [datadog_synthetics_test] Fix multistep client certificate by @AntoineDona in https://github.com/DataDog/terraform-provider-datadog/pull/2683
* [resource_datadog_logs_index] Fix having zero retention logs indexes with flex by @Prepager in https://github.com/DataDog/terraform-provider-datadog/pull/2716
* [datadog_integration_cloudflare_account] Don't set email left empty and ignore resource order by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2724
* [datadog_monitor] Change enable_samples to optional by @lexicalunit in https://github.com/DataDog/terraform-provider-datadog/pull/2680
### FEATURES
* [datadog_synthetics_test] Add mobile support by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/2701
* [datadog_sensitive_data_standard_pattern] add description and deprecate pattern by @vinckama in https://github.com/DataDog/terraform-provider-datadog/pull/2704
* [datadog_synthetics_test] Add `exitIfSucceed` to multistep API tests by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2708
* [datadog_synthetics_test] Add support for mobile steps  by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/2717
* [datadog_integration_aws_account] Add `datadog_integration_aws_account` resource  by @ktmq in https://github.com/DataDog/terraform-provider-datadog/pull/2682
* [datadog_integration_aws_available_namespaces] Create AWS Integrations V2 datasources `datadog_integration_aws_available_logs_services/namespaces` by @ktmq in https://github.com/DataDog/terraform-provider-datadog/pull/2709
### NOTES
* [datadog_dashboard] Make the `is_read_only` deprecation warning wording stronger by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2720

## New Contributors
* @bthuilot made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2713
* @ktmq made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2682
* @lexicalunit made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2680

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.49.0...v3.50.0

## 3.49.0 (November 26, 2024)

### BUGFIXES
* [resource_monitor_json] Avoid restricted role diffs with restriction_policy by @phillip-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2638
* [datadog_webhook] Fix panic on missing encode_as field by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2689
* [datadog_authn_mapping] Read team assignment when importing AuthN mappings by @chris-pinola-rf in https://github.com/DataDog/terraform-provider-datadog/pull/2686
### IMPROVEMENTS
* [provider] Bump Datadog Go SDK by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2697
### FEATURES
* [datadog_synthetics_test] Add `always_execute` and `exit_if_succeed` to browser steps by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2647
* [datadog_integration_ms_teams_tenant_based_handle] Added Resource by @kblackdd in https://github.com/DataDog/terraform-provider-datadog/pull/2657
* [datadog_rum_metric] Add support for the resource. by @hemeryar in https://github.com/DataDog/terraform-provider-datadog/pull/2643
### NOTES
* [datadog_sensitive_data_scanner_rule] make recommended keywords an opt-out option by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2674

## New Contributors
* @kblackdd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2657
* @hemeryar made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2643
* @AntoineDona made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2688
* @chris-pinola-rf made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2686
* @jegutierrez made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2692

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.48.1...v3.49.0

## 3.48.1 (November 14, 2024)

### BUGFIXES
* [datadog_integration_aws_tag_filter] include `step_functions` as valid enum for v1 AWS tag filter  by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2671


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.48.0...v3.48.1

## 3.48.0 (November 12, 2024)

### BUGFIXES
* [datadog_cloudflare] Remove default and update doc by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2658
* [synthetics_test] Handle `request_file` when state does not exist by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2659
### FEATURES
* [resource_datadog_domain_allowlist] Domain allowlist provider support by @diab42 in https://github.com/DataDog/terraform-provider-datadog/pull/2637
* [datadog_dashboard][datadog_powerpack] Update go client to support `good_minutes` and `bad_minutes` for Time Slice SLOs by @amaskara-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2662

## New Contributors
* @diab42 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2637

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.47.0...v3.48.0

## 3.47.0 (November 8, 2024)

### BUGFIXES
* [datadog_software_catalog] v3 terraform needs to hydrate the current state from raw content by @christinexlee in https://github.com/DataDog/terraform-provider-datadog/pull/2591
* [datadog_user_role] lookup roles per user, rather than users per role by @pablito-perez in https://github.com/DataDog/terraform-provider-datadog/pull/2609
* [cloud_configuration_rule] Make order of cloud configuration rule tags deterministic by @volnt in https://github.com/DataDog/terraform-provider-datadog/pull/2628
* [datadog_synthetics_test] Handle secure `config_variables` on import by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2653
### FEATURES
* [datadog_dashboard] Add support for x and y axis controls on distribution widgets by @RobThomas02 in https://github.com/DataDog/terraform-provider-datadog/pull/2611
* [datadog_security_monitoring_rule] Add reference_tables field to security monitoring rule resource by @paul-hf in https://github.com/DataDog/terraform-provider-datadog/pull/2610
* [datadog_integration_gcp_sts] Add CloudRevisionRunFilters field to GCPStsServiceAccount Tf module by @tedkahwaji in https://github.com/DataDog/terraform-provider-datadog/pull/2612
* [datadog_integration_gcp_sts] Add Support for MetricNamesapceConfig & IsResourceCollectionChangeEnabled for GCP Service Accounts by @tedkahwaji in https://github.com/DataDog/terraform-provider-datadog/pull/2650

## New Contributors
* @christinexlee made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2591
* @pablito-perez made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2607
* @RobThomas02 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2611
* @tedkahwaji made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2612

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.46.0...v3.47.0

## 3.46.0 (October 3, 2024)


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.45.0...v3.46.0

## 3.45.0 (October 3, 2024)

### BUGFIXES
* [datadog_synthetics_test] Do not send `value` if empty by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2580
* [datadog_synthetics_test] Fix http2 pseudo-headers not accepted by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2581
* [datadog_dashboard] Add support for text_formats in query table widget requests by @hyungl in https://github.com/DataDog/terraform-provider-datadog/pull/2587
### IMPROVEMENTS
* [datadog_logs_index] Add Flex Index Configuration support for Logs Indexes by @KillerPaperMice in https://github.com/DataDog/terraform-provider-datadog/pull/2569
### FEATURES
* [resource_datadog_dashboard] add support for toplist widget styles in dashboards by @brtu in https://github.com/DataDog/terraform-provider-datadog/pull/2566
* [datadog_synthetics_test] Add javascript assertion to Synthetics Assertion by @RomainGuarinoni in https://github.com/DataDog/terraform-provider-datadog/pull/2576

## New Contributors
* @KillerPaperMice made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2569
* @hyungl made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2587

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.44.1...v3.45.0

## 3.44.1 (September 11, 2024)

### BUGFIXES
* [datadog_logs_custom_destination] Fix creation failing without forward_tags_restriction_list set by @Prepager in https://github.com/DataDog/terraform-provider-datadog/pull/2565
* [datadog_synthetics_test] Handle both integer and string values for port by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2577
### IMPROVEMENTS
* [datadog_provider] Bump go version in provider by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2573


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.44.0...v3.44.1

## 3.44.0 (September 5, 2024)

### BUGFIXES
* [datadog_api_key] Stop overwriting key values in state when the API omits the key by @tyjet in https://github.com/DataDog/terraform-provider-datadog/pull/2544
* [datadog_restriction_policy] Allow policy deletions outside of terraform to prompt resource recreation by @dchayes-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2545
### FEATURES
* [datadog_software_catalog] Entity resource by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/2537
* [datadog_logs_custom_destination] Introduce logs custom destination resource by @Prepager in https://github.com/DataDog/terraform-provider-datadog/pull/2542
* [datadog_synthetics_test] Support `grpc_message` extracted variables by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2552
* [datadog_logs_custom_destination] Introduce logs custom destination resource by @Prepager in https://github.com/DataDog/terraform-provider-datadog/pull/2555
### NOTES
* [datadog_synthetics_test] allow variable in port by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2522
* [datadog_api_key] Add deprecation warning to datadog_api_key data source by @tyjet in https://github.com/DataDog/terraform-provider-datadog/pull/2541
* [datadog_api_key] Add deprecation warning for importing datadog_api_key resources by @tyjet in https://github.com/DataDog/terraform-provider-datadog/pull/2543
* [SOCE-1320] update description for slo validate option by @zhengshizhao in https://github.com/DataDog/terraform-provider-datadog/pull/2532

## New Contributors
* @tyjet made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2541
* @dchayes-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2545

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.43.1...v3.44.0

## 3.43.1 (August 20, 2024)

### BUGFIXES
* [datadog_synthetics_test] Do not send `field` if empty by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2533


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.43.0...v3.43.1

## 3.43.0 (August 14, 2024)

### BUGFIXES
* [datadog_synthetics_test] prevent unnecessary marshalling unmarshalling of proto files / json descriptors by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2503
### IMPROVEMENTS
* [RUM] Update terraform definition with `rum_stream` for `data_source` by @ThibautGeriz in https://github.com/DataDog/terraform-provider-datadog/pull/2497
* [datadog_ip_allowlist] add ip allowlist example by @afrieds in https://github.com/DataDog/terraform-provider-datadog/pull/2509
### FEATURES
* [datadog_synthetics_test] Support upload file in multistep API test by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2516
* [datadog_dashboard] Update docs for new `incident_analytics` datasource by @brtu in https://github.com/DataDog/terraform-provider-datadog/pull/2523
* [provider] Add a default_tags attribute in the Datadog provider config, currently only used by the `datadog_monitor` resource by @Amaury-Behague in https://github.com/DataDog/terraform-provider-datadog/pull/2486
### NOTES
* [provider] bump go, fix type change by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2528

## New Contributors
* @ThibautGeriz made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2497
* @afrieds made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2509
* @brtu made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2523
* @Amaury-Behague made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2486
* @JHPart made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2512

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.42.0...v3.43.0

## 3.42.0 (July 22, 2024)

### BUGFIXES
* [datadog_synthetics_test] remove body from non-POST requests by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2492
### IMPROVEMENTS
* [provider] Bump datadog-api-client to latest master by @AmitaiBl in https://github.com/DataDog/terraform-provider-datadog/pull/2494
* [datadog_provider] Reduce idle connection timeout by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2496

## New Contributors
* @AmitaiBl made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2494
* @amaskara-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2498

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.41.0...v3.42.0

## 3.41.0 (July 16, 2024)

### BUGFIXES
* [datadog_synthetics_test] Add default value to `http_version` by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2406
* [datadog_synthetics_test] Fix `renotify_interval` drift when saving from UI by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2457
* [datadog_synthetics_test] Remove `request_definition.timeout` default value by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2459
* [datadog_synthetics_test] prevent overriding the local state element from the backend response by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2460
* [datadog_service_account] Role-assignment improvements for service accounts by @retsguj in https://github.com/DataDog/terraform-provider-datadog/pull/2434
* [datadog_application_key] Stop overwriting key values in state when the API omits the key by @jackakeller in https://github.com/DataDog/terraform-provider-datadog/pull/2451
### IMPROVEMENTS
* [datadog_hosts] Add tags_by_source for datadog_hosts data source by @Nothing4You in https://github.com/DataDog/terraform-provider-datadog/pull/2375
### FEATURES
* [datadog_synthetics_test] Support `unary` gRPC steps in multistep API tests by @Drarig29 in https://github.com/DataDog/terraform-provider-datadog/pull/2445
* [datadog_dashboard] Add `cross_org_uuid support` by @cheungringo in https://github.com/DataDog/terraform-provider-datadog/pull/2450
* [datadog_monitor] Handle zero-based priority values by @carlmartensen in https://github.com/DataDog/terraform-provider-datadog/pull/2455
* [datadog_powerpack] add support for datadog_powerpack data source by @bharling in https://github.com/DataDog/terraform-provider-datadog/pull/2458
### NOTES
* [datadog_application_key] Add deprecation warning for importing datadog_application_key resources by @jackakeller in https://github.com/DataDog/terraform-provider-datadog/pull/2468
* [datadog_application_key] Add deprecation warning to datadog_application_key data source by @jackakeller in https://github.com/DataDog/terraform-provider-datadog/pull/2465

## New Contributors
* @cheungringo made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2450
* @jackakeller made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2468
* @Nothing4You made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2375
* @bharling made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2458

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.40.0...v3.41.0

## 3.40.0 (July 3, 2024)

### BUGFIXES
* [datadog_security_monitoring_suppression] Fix error handling when updating suppression payload by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/2383
* [datadog_security_monitoring_suppression] Fix error handling when creating suppression payload by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2386
* [datadog_synthetics_test] Ensure tags are non-empty and castable by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2393
* [datadog_integration_gcp_sts] Add a mutex around the mutating calls in the GCP STS module by @smuhit in https://github.com/DataDog/terraform-provider-datadog/pull/2419
* [provider] Bump SDK and fix style nits by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2400
* [datadog_azure_integration] Fix error when updating client_id and tenant_name by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2420
* [datadog_security_monitoring_rule] Upgrade go client. by @jack-edmonds-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2430
* [datadog_security_monitoring_suppression] Fix API calls returning 409 by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/2442
* [datadog_logs_metric][datadog_metric_metadata] Normalize invalid metric name by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2433
* [datadog_logs_index] Set index name to replace on modify by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2449
### IMPROVEMENTS
* [datadog_sensitive_data_scanner_rule] Add priority level in rules by @trosenblatt in https://github.com/DataDog/terraform-provider-datadog/pull/2376
* [resource_datadog_security_monitoring_rule] Add a validation step in security_monitoring_rule_resource provider. by @cpatel24 in https://github.com/DataDog/terraform-provider-datadog/pull/2357
* [datadog_synthetics_documentation] Update synthetics GRPC example by @RomainGuarinoni in https://github.com/DataDog/terraform-provider-datadog/pull/2389
* [datadog_synthetics_test] Support the bodyHash assertion. by @jack-edmonds-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2431
* [datadog_synthetics_test] Add original file name to API files by @RomainGuarinoni in https://github.com/DataDog/terraform-provider-datadog/pull/2427
* [datadog_synthetics_test] Add validation on synthetics test header by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2469
* [datadog_synthetics_test] Add support for `force_delete_dependencies` field by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2474
### FEATURES
* [datadog_synthetics_test] Add support for variablesFromScript for API tests by @ruizb in https://github.com/DataDog/terraform-provider-datadog/pull/2369
* [datadog_security_monitoring_default_rule] Add support for custom tags in security monitoring default rules by @paul-hf in https://github.com/DataDog/terraform-provider-datadog/pull/2399
* [datadog_synthetics_test] support file upload for api test by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2422
* [datadog_synthetics_test] support json schema assertion in api and multi test by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2437
* [datadog_synthetics_test] support elementsOperator in JSONPath assertions by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2444
* [datadog_organization_settings] Add Security Contacts support by @alexandre-pocquet in https://github.com/DataDog/terraform-provider-datadog/pull/2396
* [datadog_synthetics_test] Add handling of a wait step for multistep api tests by @teodor2312 in https://github.com/DataDog/terraform-provider-datadog/pull/2472
### NOTES
* [CWS-2164] - Put the cloud workload security resources as deprecated by @Malo10LeGoff in https://github.com/DataDog/terraform-provider-datadog/pull/2407
* Add deprecation message on cloud_workload_security_agent_rules resource and data source by @Malo10LeGoff in https://github.com/DataDog/terraform-provider-datadog/pull/2412
* [datadog_authn_mapping] Only allow one of role and team to be set by @JonasBak in https://github.com/DataDog/terraform-provider-datadog/pull/2391

## New Contributors
* @trosenblatt made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2376
* @cpatel24 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2357
* @ruizb made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2369
* @paul-hf made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2399
* @fabrideci made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2425
* @xafardero made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2463
* @alexandre-pocquet made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2396
* @teodor2312 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2472

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.39.0...v3.40.0

## 3.39.0 (April 23, 2024)

### BUGFIXES
* [synthetics] support http_version in multi-step API tests by @etnbrd in https://github.com/DataDog/terraform-provider-datadog/pull/2323
* [datadog_logs_index] Don't panic on empty exclusion filter by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2328
* [datadog_csm_threats_agent_rules] Add a by resource lock to prevent 409s by @Malo10LeGoff in https://github.com/DataDog/terraform-provider-datadog/pull/2332
* [datadog_security_monitoring_rule] Explicitly remove filters when none remain in config by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2334
* [datadog_user] Re-order role updates by @srosenthal-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2346
* [datadog_monitor] Handle empty `scheduling_options` struct by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2359
* [datadog_monitor_json] Ignore the POST null for the restriction_policy field by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2373
### IMPROVEMENTS
* [datadog_service_level_objective] Add query_interval_seconds to time-slice SLO condition parameters by @ddjamesfrullo in https://github.com/DataDog/terraform-provider-datadog/pull/2349
* [framework_provider] handle empty import id by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2361
* [datadog_integration_aws] Add extended resource collection field by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2365
### FEATURES
* [datadog_csm_threats_agent_rules] Add ressource and datasource for CSM Threats agent rule by @Malo10LeGoff in https://github.com/DataDog/terraform-provider-datadog/pull/2316
* [datadog_user_role] Creating a new `datadog_user_role` resource by @wangwillson1 in https://github.com/DataDog/terraform-provider-datadog/pull/2311
* [datadog_authn_mapping] Add team relationship support to AuthNMappings resource by @eric-wood in https://github.com/DataDog/terraform-provider-datadog/pull/2363
* [datadog_security_monitoring_suppression] Add data_exclusion_query by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/2360
### NOTES
* [security_monitoring_rule] Deprecate filter field for some rule types by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/2347

## New Contributors
* @anton-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2322
* @etnbrd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2324
* @barweiss made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2325
* @tim-chaplin-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2339
* @srosenthal-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2346
* @JonasBak made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2317
* @wangwillson1 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2311
* @eric-wood made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2363

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.38.0...v3.39.0

## 3.38.0 (March 13, 2024)

### BUGFIXES
* [datadog_sensitive_data_scanner_group] Mark resource as missing when not found by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2301
* [datadog_monitor] Check eval window cast before using by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2307
* [datadog_logs_archive] Fix project id for logsArchive by @Malo10LeGoff in https://github.com/DataDog/terraform-provider-datadog/pull/2312
* [datadog_users] Don't panic if no users is found by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2302
### IMPROVEMENTS
* [datadog_logs_archive] Make project id optional in GCS destination by @artslidd in https://github.com/DataDog/terraform-provider-datadog/pull/2299
* [datadog_integration_gcp] Migrate to FW Provider, Add ResourceCollectionEnabled and IsSecurityCommandCenterEnabled fields by @smuhit in https://github.com/DataDog/terraform-provider-datadog/pull/2230
### FEATURES
* [datadog_integration_gcp_sts] Add Account tags, ResourceCollectionEnabled and IsSecurityCommandCenterEnabled support to gcp tf module by @dtru-ddog in https://github.com/DataDog/terraform-provider-datadog/pull/2134
### NOTES
* [datadog_provider] Bump datadog-api-client-go version  by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2318

## New Contributors
* @smuhit made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2230
* @Malo10LeGoff made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2312

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.37.0...v3.38.0

## 3.37.0 (February 28, 2024)

### IMPROVEMENTS
* [datadog_service_account] Migrate resource to framework by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2269
* [datadog_integration_pagerduty_service_object] Add importer by @ruben-chainalysis in https://github.com/DataDog/terraform-provider-datadog/pull/2286
* [provider] Update go version by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2295
### NOTES
* [datadog_security_monitoring_suppression] Add an example for suppression by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/2288

## New Contributors
* @Drarig29 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2282
* @ruben-chainalysis made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2286
* @smaeda-ks made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2287

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.36.1...v3.37.0

## 3.36.1 (February 13, 2024)

### BUGFIXES
* [datadog_monitor] Sync search query validation with backend API by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2275
* [datadog_logs_index] Fix permadiff on omitted reset time by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2280
### FEATURES
* [datadog_synthetics_test] Add plain_proto_file and deprecate compressed_json_descriptor by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/2273


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.36.0...v3.36.1

## 3.36.0 (February 7, 2024)

### BUGFIXES
* [datadog_synthetics_test] Fix no_saving_response_body option to the request config for multi api test by @RomainGuarinoni in https://github.com/DataDog/terraform-provider-datadog/pull/2249
* [datadog_synthetics_test] Check if ci options block is valid by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2252
* [datadog_apm_retention_filter] User mutex to prevent concurrent operation by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2244
* [datadog_monitor] Omit recurrence start field if its an empty string by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2258
### IMPROVEMENTS
* [ip_allow_list] Migrate ip allow list to framework by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2239
* [datadog_service_level_objective] Add support for time-slice SLOs by @ddjamesfrullo in https://github.com/DataDog/terraform-provider-datadog/pull/2245
* [datadog_api_key][datadog_application_key] Add exact_match option to data sources by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2247
* [datadog_sensitive_data_scanner] Handle new `included_keyword_configuration` attribute in scanning rule by @artslidd in https://github.com/DataDog/terraform-provider-datadog/pull/2238
* [datadog_dashboards] Bump limit for allowed graphs in split graph widget by @adil-alimbetov in https://github.com/DataDog/terraform-provider-datadog/pull/2265
* [datadog_logs_index] Add daily limit reset options to log indexes by @Prepager in https://github.com/DataDog/terraform-provider-datadog/pull/2263
### FEATURES
* [security_monitoring_rule] Support Third Party rules by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/2231
* [datadog_security_monitoring_suppression] Add support for suppressions by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/2246
* [datadog_openapi_api] Add support for API management resource by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/2262

## New Contributors
* @RomainGuarinoni made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2249
* @buraizu made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2254
* @michellealbright12 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2259
* @dsinn made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2253
* @artslidd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2238

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.35.0...v3.36.0

## 3.35.0 (January 10, 2024)

### IMPROVEMENTS
* [datadog_rum_application] migrate to framework by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2215
* [datadog_dashboard][datadog_powerpack] Support cloud cost query in F&F queries by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2211
* [application_key] Migrate application key resource to framework by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2224
* [datadog_metric_metadata] correct consistency of type by @fl0Lec in https://github.com/DataDog/terraform-provider-datadog/pull/2222
* [datadog_gcp_sts] Avoid initializing attributes using `WithDefault()` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2236
### FEATURES
* [datadog_integration_aws_event_bridge] Add support for AWS Event Bridge integration by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2204
* [datadog_users] Added support for the Users data source by @aldrickdev in https://github.com/DataDog/terraform-provider-datadog/pull/2220
### NOTES
* Bump datadog-api-client client to 2.21.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2234

## New Contributors
* @NBParis made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2223
* @aldrickdev made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2220
* @fl0Lec made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2222

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.34.0...v3.35.0

## 3.34.0 (December 14, 2023)

### BUGFIXES
* [datadog_dashboard] Don't crash on empty group_by in widgets by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/2189
* [datadog_service_definition_yaml] Validate tag type is string by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2196
### IMPROVEMENTS
* [datadog_integration_azure] Migrate resource to framework by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2202
* [datadog_dashboards] [datadog_powerpacks] clarify widget layout schema component descriptions by @dashashifrina in https://github.com/DataDog/terraform-provider-datadog/pull/2182
### FEATURES
* [datadog_integration_azure] Add resource_collection_enabled by @ava-silver in https://github.com/DataDog/terraform-provider-datadog/pull/2203
* [dashboards] Support for powerpack widgets by @dashashifrina in https://github.com/DataDog/terraform-provider-datadog/pull/2205
* [datadog_powerpack] Add Powerpack resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2187
* [datadog_security_monitoring_rule] Support for Application Security rule type by @nizox in https://github.com/DataDog/terraform-provider-datadog/pull/2195
### NOTES
* [datadog_integration_aws_*] Validate AWS account ID and improve error handling by @rjhornsby in https://github.com/DataDog/terraform-provider-datadog/pull/2201

## New Contributors
* @nizox made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2195
* @rjhornsby made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2201

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.33.0...v3.34.0

## 3.33.0 (November 21, 2023)

### BUGFIXES
* [datadog_service_definition_yaml] Keep `ci-pipeline-fingerprints` sorted in state by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2178
### IMPROVEMENTS
* [provider] Remove terraform logger transport by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2167
* [data_source_datadog_service_level_objectives] Add ability to query slo with multiple tags by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2172
* [datadog_monitor] Add custom_schedule fields by @unclebconnor in https://github.com/DataDog/terraform-provider-datadog/pull/2177
### FEATURES
* [datadog_integration_azure] Add container_app_filters by @ava-silver in https://github.com/DataDog/terraform-provider-datadog/pull/2138
* [apm_retention_filters] Add resources and datasource by @SalahEddineBC in https://github.com/DataDog/terraform-provider-datadog/pull/2137
* [datadog_metric_tag_configuration] Adds exclude_tags_mode; updates include_percentiles by @jmfurlott in https://github.com/DataDog/terraform-provider-datadog/pull/1949
### NOTES
* [provider] bump Go SDK to 2.19.0 by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2185

## New Contributors
* @romainmeunier made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2175
* @SalahEddineBC made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2137
* @jmfurlott made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1949

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.32.0...v3.33.0

## 3.32.0 (October 31, 2023)

### BUGFIXES
* [datadog_service_definition_yaml] Fix typo in v2.2 example by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/2152
* [datadog_integration_aws] Allow account_id to be specified in an access_key configured integration by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2154
### IMPROVEMENTS
* [data_source_datadog_team_memberships] add exact match support by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2118


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.31.0...v3.32.0

## 3.31.0 (October 18, 2023)

### BUGFIXES
* [datadog_sensitive_data_scanner_group] Avoid panic when nil returned by @mikutas in https://github.com/DataDog/terraform-provider-datadog/pull/2097
* [datadog_synthetics_test] Prevent setting empty values for rum settings by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/2124
* [datadog_cloud_configuration_rule] Set userActivationStatus to true when at least one group_by field is present by @christophetd in https://github.com/DataDog/terraform-provider-datadog/pull/2131
### IMPROVEMENTS
* [datadog_synthetics_test] Add support for timings scope for response assertions by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/2101
* [datadog_dashboard] Support Formula and Function requests in heatmap widget by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2103
* [datadog_provider] Document default values of schemas by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2139
* [datadog_ip_ranges] Adds global prefixes to the data source by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1783
* [datadog_dashboards] Implement support for split graph widget by @adil-alimbetov in https://github.com/DataDog/terraform-provider-datadog/pull/2129
### FEATURES
* [datadog_service_definition_yaml] Support service definition schema v2.2 by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/2126
* [datadog_synthetics_test] Add support for proto files for grpc tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/2116
### NOTES
* [provider] Bump Go SDK by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2147

## New Contributors
* @mikutas made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2097
* @dtru-ddog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2104
* @traegerdog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2120

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.30.0...v3.31.0

## 3.30.0 (September 12, 2023)

### BUGFIXES
* [datadog_downtimes_schedule] Fix downtimes schedule with monitor id identifier by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2092
### IMPROVEMENTS
* [datadog_synthetics_tests] Retrieve synthetics test in datasource by making 2 api calls by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2068
* [datadog_dashboard] Add support for sunburst widget request style by @matt-miller-ddog in https://github.com/DataDog/terraform-provider-datadog/pull/2089
### FEATURES
* [datadog_synthetics_test] Add support for metadata for Synthetics grpc tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/2094
### NOTES
* [datadog_downtime_schedule] GA the resource by @carlmartensen in https://github.com/DataDog/terraform-provider-datadog/pull/2090
* [datadog_provider] Bump datadog-api-client to `2.17.0` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2099


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.29.0...v3.30.0

## 3.29.0 (August 23, 2023)

### BUGFIXES
* [datadog_sensitive_data_scanner_rule] Handle empty string for `excluded_namespaces` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2041
* [datadog_dashboard] Fix handling of nil notify_by list by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/2046
* [datadog_spans_metric] Normalize tag value by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2056
* [datadog_integration_aws_tag_filter] Prevent concurrency issues with mutex by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2063
### IMPROVEMENTS
* [restriction_policy] Update restriction_policy resource documentation for private beta dashboard support by @sanjaythakkar-dd in https://github.com/DataDog/terraform-provider-datadog/pull/2028
* [datadog_dashboard_list] Migrate to tf framework by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/2038
* [data_sources] Relax unparsed object check for datasources by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2043
* Use validators to mark objects as required by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2053
* [datadog_synthetics_test] Add persistCookies option by @luissmg in https://github.com/DataDog/terraform-provider-datadog/pull/2058
### FEATURES
* [datadog_cloud_configuration_rule] Added filter field by @symphony-elias in https://github.com/DataDog/terraform-provider-datadog/pull/2015
* [datadog_downtime_schedule] Add support for Downtime V2 by @carlmartensen in https://github.com/DataDog/terraform-provider-datadog/pull/2023
* [datadog_integration_azure] Add app_service_plan_filters, cspm, and custom metrics by @ava-silver in https://github.com/DataDog/terraform-provider-datadog/pull/2059
* [datadog_team_permission] Add team permission setting resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2070
### NOTES
* [datadog_provider] Bump go version to `1.20` and other terraform package versions by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2030
* [datadog_provider] Switch sdkv2 schemas to `SchemaFunc` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2034
* [datadog_provider] Bump datadog-api-client to 2.16.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2074

## New Contributors
* @ava-silver made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2059
* @luissmg made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/2058

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.28.0...v3.29.0

## 3.28.0 (July 20, 2023)

### BUGFIXES
* [datadog_dashboard_json] Check if widgets are valid while casting by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2012
### IMPROVEMENTS
* [data_source_datadog_synthetics_test] Allow datasource to take a test URL as well as id by @Cgg in https://github.com/DataDog/terraform-provider-datadog/pull/1993
* [datadog_dashboard] Add a new field additional_query_filters to Formula and Function SLO query by @zhengshizhao in https://github.com/DataDog/terraform-provider-datadog/pull/1997
* [restriction_policy] Update restriction_policy resource documentation by @sanjaythakkar-dd in https://github.com/DataDog/terraform-provider-datadog/pull/1995
* [datadog_dashboard] Support overlay display type in dashboard widget definition requests by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/2006
* [datadog_integration_confluent_resource] Add support for `enable_custom_metrics` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/2011
### FEATURES
* [datadog_service_definition] support backstage schema validation by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/2004
* [datadog_service_definition] support backstage schema validation v2 by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/2008

## New Contributors
* @Cgg made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1993
* @sanjaythakkar-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1995

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.27.0...v3.28.0

## 3.27.0 (July 3, 2023)

### BUGFIXES
* [datadog_synthetics_test] Fix for optional fields in synthetics oauth by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1956
* [datadog_synthetics_test] Handle both example and secure missing from variables by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1959
* [datadog_monitor] Fix unsetting of `restricted_roles` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1970
* [datadog_spans_metric] Convert `group_by` to set by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1972
* [datadog_synthetics_test] Fix targetvalue for isUndefined json path assertions by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1966
* [datadog_synthetics_test] Allow synthetics web basicauth without password set by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1989
### IMPROVEMENTS
* [datadog_user] Add support for `exact_match` in data source by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1968
* [datadog_dashboard] Add note about tf managed dashboard lists when using `dashboard_lists` resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1975
### FEATURES
* [datadog_service_account_application_key] Add service_account_application_key resource by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1955
* [datadog_service_account] Add service_account data source by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1955
* [restriction_policy] Add restriction policy resource by @agnesjang98 in https://github.com/DataDog/terraform-provider-datadog/pull/1957
### NOTES
* [datadog_integration_gcp] Mark resource as deprecated by @joshhuie in https://github.com/DataDog/terraform-provider-datadog/pull/1965
* [provider] Add datasource wrapper by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1962
* [provider] Remove `ResourceDataKey` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1961
* Upgrade testing framework to `terraform-plugin-testing` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1967
* [provider] bump datadog-api-client Go SDK to v2.14.0 by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1988

## New Contributors
* @joshhuie made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1965
* @agnesjang98 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1957

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.26.0...v3.27.0

## 3.26.0 (June 5, 2023)

### BUGFIXES
* [provider] Check for empty api_url string by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1907
* [datadog_logs_custom_pipelines] Add empty string validation to grok parser samples list by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1917
* [datadog_logs_archive] Don't crash on empty logs archives destination by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1915
* [provider] Fix segfault on invalid API URL by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1921
* [datadog_synthetics_test] Fix segfault on uninitialized basic auth fields by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1948
### IMPROVEMENTS
* [datadog_dashboard] Add slo datasource to dash widget formulas and functions query by @zhengshizhao in https://github.com/DataDog/terraform-provider-datadog/pull/1924
* [datadog_synthetics_test] Add secure field to browser variables and update docs by @alai97 in https://github.com/DataDog/terraform-provider-datadog/pull/1918
* [multiple_resources] Cleanup refs in resources by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1925
* [datadog_provider] Add resource wrapper and support adding enum values to docs by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1922
* [datadog_synthetics_test] Add secure options to extracted values in synthetics multistep tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1938
* [datadog_dashboard] Add sort block to Dashboard list_stream query block by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1941
### FEATURES
* [datadog_integration_aws_namespace_rules] Add data source for available aws namespace rules by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1905
* [datadog_spans_metric] Add spans metric support by @vitormaco in https://github.com/DataDog/terraform-provider-datadog/pull/1858
* [datadog_synthetics_concurrency_cap] Add concurrency cap resource for synthetics tests by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1913
* [datadog_integration_gcp_sts] generate GCP STS account resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1936
* [datadog_team] Add support for datadog team resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1908
* [datadog_team_link] Add support for datadog team link resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1908
* [datadog_team_membership] Add support for datadog team membership resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1908
* [datadog_team] Add datadog team data source by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1946
* [datadog_team_memberships] Add datadog team memberships data source by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1946
### NOTES
* [provider] Bump `datadog-api-client` version to `v2.13.0` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1940

## New Contributors
* @vitormaco made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1858
* @alai97 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1918
* @boraborah made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1923
* @IanMoroney made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1928

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.25.0...v3.26.0

## 3.25.0 (May 5, 2023)

### BUGFIXES
* [provider] Update API key validation call to use official SDK by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1902
### IMPROVEMENTS
* [datadog_synthetics_test] Add support for new assertions by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1901
* [datadog_dashboard] Add additional_query_filters to SLO widget by @zhengshizhao in https://github.com/DataDog/terraform-provider-datadog/pull/1899
### FEATURES
* [datadog_sensitive_data_scanner_group_order] Add resource and datasource for SDS group order by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1820


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.24.1...v3.25.0

## 3.24.1 (May 3, 2023)

### BUGFIXES
* [datadog_logs_metric] Convert `group_by` to set by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1896
* [datadog_dashboard_json] skip prepResource if attrMap is nil by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1898


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.24.0...v3.24.1

## 3.24.0 (April 26, 2023)

### BUGFIXES
* [datadog_dashboard] Safeguard against empty formula definitions by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1838
* [datadog_synthetics_test] Check error when building token API authentication by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1837
* [datadog_dashboard] Handle empty section in template variables by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1842
* [datadog_synthetics_test] Handle advanced scheduling in browser tests by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1856
* [datadog_dashboard_json] Handle continuous diff if `is_read_only` is not set by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1854
* [datadog_monitor] Fix dropped `hour_starts` field when value is zero by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1878
* [datadog_service_level_objective] Fix edge case removing fields that are both optional and computed by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1879
* [datadog_synthetics_test] Fix missing default value for browser variables by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1884
* [datadog_monitor] Fix panic on scheduling option with empty evaluation window by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1885
* [datadog_synthetics_test] Remove required from jsonpath and xpath targetvalue fields by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1880
* [datadog_api_key] Set ID in state and change error message back to original by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1892
### IMPROVEMENTS
* [datadog_service_level_objective] Expose more SLO data in the SLO datasource by @ali-sattari in https://github.com/DataDog/terraform-provider-datadog/pull/1829
* [datadog_monitor] remove Beta notice from notify_by Monitor resource by @Dalje-et in https://github.com/DataDog/terraform-provider-datadog/pull/1840
* [provider] Begin upgrading the provider to use `terraform-framework-provider` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1776
* [datadog_dashboard] Add sort block in slo_list_definition by @sbecker59 in https://github.com/DataDog/terraform-provider-datadog/pull/1835
* [datadog_synthetics_test] Refactor options for Synthetics tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1861
* [datadog_api_key] Upgrade API key data source to framework by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1780
* [datadog_service_definition_yaml] Add v2.1 service definition example by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/1883
* [datadog_service_definition_yaml] Add support for v2.1 schema of service catalog by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/1833
* [datadog_dashboard] Implement support for team tags in Dashboards by @DrkSephy in https://github.com/DataDog/terraform-provider-datadog/pull/1881
* [datadog_synthetics_test] Add no screenshot option to browser steps by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1887
### FEATURES
* [datadog_hosts] Add a host list datasource by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1847
* [datadog_integration_fastly_account] Add Fastly integration resources #1851 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1852
* [datadog_integration_fastly_service] Add Fastly integration resources #1851 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1852
* [datadog_integration_confluent_account] Add Confluent integration resources by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1851
* [datadog_integration_confluent_resource] Add Confluent integration resources by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1851
* [datadog_integration_cloudflare_account] Add Cloudflare account resource by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1853
### NOTES
* [provider] Bump go version to `1.19` and `terraform-plugin-framework` to `1.2.0` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1845
* [provider] Bump datadog api client to 2.12.0 by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1873

## New Contributors
* @apiazza3 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1827
* @ali-sattari made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1829
* @sbecker59 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1835

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.23.0...v3.24.0

## 3.23.0 (March 21, 2023)

### BUGFIXES
* [datadog_downtime] Ignore start/end comparison on recurring downtimes by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1811
* [datadog_monitor] Make search required for f&f monitors by @JoannaYe-Datadog in https://github.com/DataDog/terraform-provider-datadog/pull/1814
* [datadog_synthetics_test] Allow empty values for service for Synthetics grpc tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1813
* [datadog_sensitive_data_scanner] Add mutex to SDS resource by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1818
* [datadog_metric_tag_configuration] Fix tag validation for metric tags configs by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1819
* [datadog_security_monitioring_rule] Set `tags` in state by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1821
* [datadog_sensitive_data_scanner_rule] Handle diffs when importing by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1823
### IMPROVEMENTS
* [datadog_logs_custom_pipeline] Add reference table logs processor by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1800
### FEATURES
* [datadog_ip_allowlist] Support IP allowlist management in terraform by @retsguj in https://github.com/DataDog/terraform-provider-datadog/pull/1788
### NOTES
* [provider] Remove custom transport by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/1812


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.22.0...v3.23.0

## 3.22.0 (March 7, 2023)

### BUGFIXES
* [datadog_service_definition_yaml] Add ordering of contacts in state func by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1785
* [datadog_security_monitoring_rule] Don't cast nil values in `group_by_fields` or `distinct_fields` by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1796
* [datadog_logs_index_order] Make name optional and computed by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1799
### IMPROVEMENTS
* [datadog_permissions] Add option to show restricted permissions by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1782
* [datadog_dashboard] Add run-workflow widget definition by @DavidRobertAnsart in https://github.com/DataDog/terraform-provider-datadog/pull/1779
* [datadog_ip_ranges] Adds orchestrator to the resource by @tbavelier in https://github.com/DataDog/terraform-provider-datadog/pull/1781
### FEATURES
* [datadog_synthetics_test] add advanced scheduling by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1777
* [datadog_sensitive_data_scanner_group] Generate Sensitive Data Scanner group resources by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1737
* [datadog_sensitive_data_scanner_rule] Create Sensitive Data Scanner rule resource and data source by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1738
### NOTES
* [provider] Bump `terraform-plugin-sdk/v2`to latest by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1765
* [Provider] Bump datadog api client to 2.10.0 by @HantingZhang2 in https://github.com/DataDog/terraform-provider-datadog/pull/1803

## New Contributors
* @tbavelier made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1781
* @dependabot made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1789

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.21.0...v3.22.0

## 3.21.0 (February 9, 2023)

### BUGFIXES
* [datadog_service_level_objective] Set thresholds fields as optional computed by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1725
* [datadog_synthetics_test] Prevent setting secure property on config variables of type global by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1755
### IMPROVEMENTS
* [datadog_service_level_objective] Update terraform-provider-datadog to allow primary timeframe, target, and warning to be specified by @ddjamesfrullo in https://github.com/DataDog/terraform-provider-datadog/pull/1704
* [datadog_synthetics_test] Add support for http version in test options by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1722
* [datadog_security_monitoring_default_rule] Warn when the rule has a deprecation date by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/1728
* [datadog_monitor] Add support for `notification_preset_name` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1749
* [datadog_integration_gcp] Support enabling the CSPM integration by @christophetd in https://github.com/DataDog/terraform-provider-datadog/pull/1748
* [datadog_dashboard] Add event_size fields to list stream query by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1761
### FEATURES
* [datadog_synthetics] Add TOTP Parameters for Global Variables by @thestefanristovski in https://github.com/DataDog/terraform-provider-datadog/pull/1708
* [datadog_monitor_config_policy] Add monitor config policies resource and data source by @carlmartensen in https://github.com/DataDog/terraform-provider-datadog/pull/1750

## New Contributors
* @ddjamesfrullo made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1704
* @thestefanristovski made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1708
* @carlmartensen made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1752
* @christophetd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1748

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.20.0...v3.21.0

## 3.20.0 (January 11, 2023)

### BUGFIXES
* [security_monitoring_default_rule] Fix acceptance tests for default rules by @muffix in https://github.com/DataDog/terraform-provider-datadog/pull/1707
* [datadog_dashboard] Fix palette_index not set in the API when set in formula style by @valerian-roche in https://github.com/DataDog/terraform-provider-datadog/pull/1714
### IMPROVEMENTS
* [service_definition_yaml] Add tag normalization util by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1686
* [datadog_monitor] Mark `notify_by` field as private beta by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1694
* [datadog_synthetics_test] Add secure field to synthetics config variable by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1696
* [datadog_role] Add ability to skip pre-flight `permission` validation by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1703
### FEATURES
* [datadog_cloud_configuration_rule] Implemented resource to manage cloud_configuration rules by @symphony-elias in https://github.com/DataDog/terraform-provider-datadog/pull/1677
* [datadog_service_account] Add service account resource by @mnguyendatadog in https://github.com/DataDog/terraform-provider-datadog/pull/1685
* [datadog_integration_aws_logs_services] Add an AWS log ready services data source by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1690
### NOTES
* [provider] Bump `datadog-api-client` to 2.7.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1717

## New Contributors
* @mnguyendatadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1685
* @valerian-roche made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1714

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.19.1...v3.20.0

## 3.19.1 (December 22, 2022)

### BUGFIXES
* [datadog_logs_metric] Fix `getUpdateCompute ` method for non distribution aggregation type by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1683

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.19.0...v3.19.1

## 3.19.0 (December 21, 2022)

### BUGFIXES
* [datadog_service_definition_yaml] Fix panic on missing name in service definition links by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1660
* [datadog_logs_custom_pipeline] Handle nested empty filter query by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1655
* [datadog_integration_aws] Properly handle missing resource when importing by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1657
* [datadog_logs_archive] mark `path` as optional by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1658
* [datadog_integration_aws] Change excluded_regions to TypeSet by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1656
### IMPROVEMENTS
* [datadog_security_monitoring_rule] Updating schema validation for field type by @symphony-elias in https://github.com/DataDog/terraform-provider-datadog/pull/1640
* [datadog_logs_metric] Add include_percentiles attribute to distribution compute by @JeanCoquelet in https://github.com/DataDog/terraform-provider-datadog/pull/1645
* [datadog_synthetics_test] Add Digest auth by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1669
* [datadog_dashboard] Add support for style field in dashboard widget formulas by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1671
* [datadog_monitor] Add enable_samples monitor options by @liashenko in https://github.com/DataDog/terraform-provider-datadog/pull/1670
* [datadog_synthetics_test] Add support for oauth authentication by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1668
### FEATURES
* [datadog_logs_archive_order] Add a logs archive order data source by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1661
* [datadog_rum_application] Add a RUM application data source by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1641
### NOTES
* Bump `datadog-api-client` version to v2.6.1 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1678

## New Contributors
* @symphony-elias made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1640
* @JeanCoquelet made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1645
* @piotrekkr made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1663

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.18.0...v3.19.0

## 3.18.0 (November 10, 2022)

### BUGFIXES
* [datadog_monitor] Handle explicit null for `new_host_delay` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1615
* [datadog_dashboard] Suppress URL attribute diff by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1614
* [datadog_dashboard] Fix panic when `slo_list` widget is in `group` widget by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1624
* [datadog_service_definition_yaml] Move `404 statusCode` check into the error check block by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1634
### IMPROVEMENTS
* [datadog_dashboard] Add support for `values` and `defaults` in template variables by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1622
* [datadog_monitor] Support monitor `scheduling_options` by @bmay2 in https://github.com/DataDog/terraform-provider-datadog/pull/1630
* [datadog_synthetics] Add support for xpath assertions by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1632
* [datadog_synthetics] Add body_type field to SyntheticsTest request_definition by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1629
### NOTES
* [datadog_provider] Bump `go` and `terraform-plugin-sdk` versions by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1597

## New Contributors
* @bmay2 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1630

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.17.0...v3.18.0

## 3.17.0 (October 24, 2022)

### BUGFIXES
* [datadog_logs_metric] add nil check to Logs Metrics getGroupBys by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1608
### IMPROVEMENTS
* [datadog_dashboard] add support for storage parameter in widget queries by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1569
* [datadog_dashboard] Add support for the SLO List widget by @mmeyer724 in https://github.com/DataDog/terraform-provider-datadog/pull/1595
* [datadog_security_monitoring_rule] Add Terraform Support for Signal Correlation Rules by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/1593
* [datadog_monitor] add notify_by option by @chrismdd in https://github.com/DataDog/terraform-provider-datadog/pull/1599
* [datadog_synthetics_test] Add missing disable_cors option by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1605
### NOTES
* [datadog_security_monitoring_rules] Deprecate metric field of Security Monitoring Rules by @clementgbcn in https://github.com/DataDog/terraform-provider-datadog/pull/1604
* Bump `datadog-api-client` to v2.4.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1610

## New Contributors
* @mmeyer724 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1595
* @clementgbcn made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1593

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.16.0...v3.17.0

## 3.16.0 (September 28, 2022)

### BUGFIXES
* [datadog_dashboard] Handle empty widgets by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1568
* [datadog_dashboards] Handle empty group definition by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1576
* [datadog_security_monitoring_default_rule] Add missing schema attribute `type` by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1579
### IMPROVEMENTS
* [datadog_synthetics_test] Add missing options for synthetics tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1553
* [datadog_dashboard] add Topology Map definition by @anbnyc in https://github.com/DataDog/terraform-provider-datadog/pull/1557
* [datadog_synthetics_global_variable] Add support for local variable extract from test by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1567
### FEATURES
* [datadog_service_definition] Add terraform support for service catalog definition (yaml/json) by @hyperloglogy in https://github.com/DataDog/terraform-provider-datadog/pull/1556
* [datadog_logs_pipelines] Add a pipelines datasource by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1555

## New Contributors
* @hyperloglogy made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1556
* @jketcham made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1428

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.15.1...v3.16.0

## 3.15.1 (September 8, 2022)

### BUGFIXES
* [datadog_ip_ranges] Fix `IPRanges` server configuration by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1560


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.15.0...v3.15.1

## 3.15.0 (September 7, 2022)

### BUGFIXES
* [datadog_security_monitoring_rule] Check for nil `metrics` values. by @juliendoutre in https://github.com/DataDog/terraform-provider-datadog/pull/1506
* [datadog_synthetics_private_location] Improve restricted_roles handling by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1519
* [datadog_synthetics_test] Fix target for packet loss assertions by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1527
* [provider] Handle redirects by retaining the original request/redirect body by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1534
* [datadog_synthetics_test] suppress whitespace diff when comparing files by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1518
* [datadog_dashboard] Handle empty requests definition in hostmap widget by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1546
### IMPROVEMENTS
* [datadog_synthetics_test] Add DiffSupressFunc for rum settings by @bhaoui in https://github.com/DataDog/terraform-provider-datadog/pull/1532
* [datadog_monitor] Suppress diff when using `ok` and `unknown` thresholds in non `service check` monitors. by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1529
* [datadog_dashboard] add support for priority parameters in ManageStatus widget by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1525
* [datadog_monitor] add group_retention_duration and on_missing_data monitor options by @chrismdd in https://github.com/DataDog/terraform-provider-datadog/pull/1535
* [datadog_monitor] Add formula & functions by @phillip-dd in https://github.com/DataDog/terraform-provider-datadog/pull/1357
* [datadog_role] Upgrade provider from old single-permission APIs to newer UpdateRole API by @retsguj in https://github.com/DataDog/terraform-provider-datadog/pull/1542
### FEATURES
* [datadog_rum_application] Add RUM Application resource support by @nkzou in https://github.com/DataDog/terraform-provider-datadog/pull/1537
### NOTES
* [provider] Bump datadog-api-client to V2 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1515
* Bump datadog-api-client to v2.2.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1547

## New Contributors
* @bhaoui made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1532
* @nkzou made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1525
* @buranmert made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1538

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.14.0...v3.15.0

## 3.14.0 (July 25, 2022)

### BUGFIXES
* [datadog_synthetics_test] Fix ci execution rule options for browser tests by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1491
* [datadog_synthetics_test] defaults api_step timeout to 60 to avoid it defaulting to 0 by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1497
### FEATURES
* [datadog_security_monitoring_rule] NewValue detection type supports threshold learning duration and metrics by @juliendoutre in https://github.com/DataDog/terraform-provider-datadog/pull/1479
* [datadog_security_monitoring_rule] Dynamic Criticality Support by @pietrodll in https://github.com/DataDog/terraform-provider-datadog/pull/1483
* [datadog_synthetics_test] Add support for grpc subtype by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1500
### NOTES
* [provider] Update go client by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1501

## New Contributors
* @juliendoutre made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1479
* @pietrodll made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1483

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.13.1...v3.14.0

## 3.13.1 (July 11, 2022)

### BUGFIXES
* [provider] Update client by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1488


**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.13.0...v3.13.1

## 3.13.0 (July 5, 2022)

### IMPROVEMENTS
* [datadog_monitor] Add ci-tests alert to terraform monitor docs. by @liashenko in https://github.com/DataDog/terraform-provider-datadog/pull/1451
* [datadog_monitor_json] Avoid unnecessary restricted_roles diff by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1457
* [datadog_dashboard] Add list_stream_definition resource type by @luisvalini in https://github.com/DataDog/terraform-provider-datadog/pull/1470
* [datadog_synthetics_test] Add rum settings by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1464
* [datadog_synthetics_test] Add support for ci execution rule option by @romainberger in https://github.com/DataDog/terraform-provider-datadog/pull/1474
### FEATURES
* [datadog_integration_opsgenie_service_object] Add support for Opsgenie service resource by @abravo3641 in https://github.com/DataDog/terraform-provider-datadog/pull/1466

## New Contributors
* @liashenko made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1451
* @NouemanKHAL made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1463
* @luisvalini made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1470
* @abravo3641 made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1466

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.12.0...v3.13.0

## 3.12.0 (May 24, 2022)

### BUGFIXES
* [datadog_synthetics_test] Allow users to set user locator with no element by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1437
### IMPROVEMENTS
* [datadog_synthetics_test] [datadog_synthetics_private_location] Add support for restricted roles on private locations and synthetics tests by @dajofo in https://github.com/DataDog/terraform-provider-datadog/pull/1423
* [datadog_logs_archive] Handle new rehydration_max_scan_size_in_gb field for Logs Archives by @corentinmarc in https://github.com/DataDog/terraform-provider-datadog/pull/1440
* [datadog_downtime]: Update docs for rrule/type by @mikebryant in https://github.com/DataDog/terraform-provider-datadog/pull/1434
* [datadog_monitor] Add documentation on limits for Monitors timeout_h option by @Dalje-et in https://github.com/DataDog/terraform-provider-datadog/pull/1432
* [datadog_synthetics_test] Fix `config_variable` example by @ethan-lowman-dd in https://github.com/DataDog/terraform-provider-datadog/pull/1397
### FEATURES
* [mute_first_recovery_notification] add mute first recovery notification to monitor options by @JoannaYe-Datadog in https://github.com/DataDog/terraform-provider-datadog/pull/1417
### NOTES
* Exponential backoff period for `5xx` errors and enabled retries by default by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1444
* Bump datadog-api-client to v1.14.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1448

## New Contributors
* @Dalje-et made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1432
* @dajofo made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1423
* @JoannaYe-Datadog made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1417
* @corentinmarc made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1440
* @ethan-lowman-dd made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1397
* @mikebryant made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1434

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.11.0...v3.12.0

## 3.11.0 (April 20, 2022)

### BUGFIXES
* [datadog_synthetics_test] Remove unparsed check in the data source by @sdeprez in https://github.com/DataDog/terraform-provider-datadog/pull/1403
* [datadog_security_monitoring_rule] Add default for aggregation by @muffix in https://github.com/DataDog/terraform-provider-datadog/pull/1407
* [datadog_synthetics_test] Use a correct regex for variables by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1398
* [datadog_monitor] Fix spurious "changes made outside of terraform" by @markadev in https://github.com/DataDog/terraform-provider-datadog/pull/1384
### IMPROVEMENTS
* [resource_datadog_monitor] Add separate validation for existing monitors by @gkharkwal in https://github.com/DataDog/terraform-provider-datadog/pull/1406
* [datadog_resource_dashboard] Implement support for timeseries background in query value widgets by @DrkSephy in https://github.com/DataDog/terraform-provider-datadog/pull/1415
### FEATURES
* [datadog_security_monitoring_rule] Update provider for detection method impossible travel by @muffix in https://github.com/DataDog/terraform-provider-datadog/pull/1402
### NOTES
* [datadog_monitor] Mark locked as deprecated by @phillip-dd in https://github.com/DataDog/terraform-provider-datadog/pull/1400
* Add debug mode for developers by @AlaricCalmette in https://github.com/DataDog/terraform-provider-datadog/pull/1399
* Bump datadog-api-client-go to v1.13.0 by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1419

## New Contributors
* @muffix made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1407
* @keisku made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1404
* @markadev made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1384

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.10.0...v3.11.0

## 3.10.0 (March 28, 2022)

### BUGFIXES
* [datadog_cloud_workload_security_agent_rule] Fix `enabled` attribute reading by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1379
* [datadog_dashboard] Fix `sunburst_definition` inside `group_definition` by @volnt in https://github.com/DataDog/terraform-provider-datadog/pull/1377
### IMPROVEMENTS
* [datadog_user] Match existing users based on email by @skarimo in https://github.com/DataDog/terraform-provider-datadog/pull/1383
### FEATURES
* [datadog_synthetics_test] Add synthetics test data source by @sdeprez in https://github.com/DataDog/terraform-provider-datadog/pull/1391
### NOTES
* Update Datadog client to v1.11.0 by @therve in https://github.com/DataDog/terraform-provider-datadog/pull/1393

## New Contributors
* @volnt made their first contribution in https://github.com/DataDog/terraform-provider-datadog/pull/1377

**Full Changelog**: https://github.com/DataDog/terraform-provider-datadog/compare/v3.9.0...v3.10.0

## 3.9.0 (March 9, 2022)

IMPROVEMENTS:

-   `datadog_dashboard`: Implement support for Treemap widget ([#1331](https://github.com/DataDog/terraform-provider-datadog/pull/1331))
-   `datadog_dashboard`: Add support for `apm_stats_query` for distributions widget ([#1326](https://github.com/DataDog/terraform-provider-datadog/pull/1326))
-   `datadog_synthetics_test`: Add support new authentication types and request proxy for Synthetics tests ([#1336](https://github.com/DataDog/terraform-provider-datadog/pull/1336))
-   `datadog_synthetics_test`: Add element user locator field for browser steps ([#1346](https://github.com/DataDog/terraform-provider-datadog/pull/1346))
-   `datadog_integration_aws`: Add support for `metrics`, `cspm_resource` and `resource` collections ([#1343](https://github.com/DataDog/terraform-provider-datadog/pull/1343)) Thanks [@nikohaa](https://github.com/nikohaa)
-   `datadog_synthetics_test`: Add certificate check option for ssl tests ([#1368](https://github.com/DataDog/terraform-provider-datadog/pull/1368))
-   `datadog_synthetics_test`: Add support for is_critical option on browser steps ([#1359](https://github.com/DataDog/terraform-provider-datadog/pull/1359))

FEATURES:

-   `datadog_cloud_workload_security_agent_rules`: Add terraform support for Cloud Workload Security Agent Rules ([#1338](https://github.com/DataDog/terraform-provider-datadog/pull/1338))
-   `data_source_datadog_logs_indexes`: Add logs indexes datasource ([#1349](https://github.com/DataDog/terraform-provider-datadog/pull/1349))
-   `datadog_authn_mapping`: Add new resource SAML AuthN Mappings ([#1349](https://github.com/DataDog/terraform-provider-datadog/pull/1349))

BUGFIXES:

-   `datadog_dashboard_json`: Handle perpetual diff when both `is_read_only` and `restricted_roles` is set ([#1280](https://github.com/DataDog/terraform-provider-datadog/pull/1280))
-   `datadog_security_monitoring_rule`: Set evaluation_window to optional and fix tests ([#1347](https://github.com/DataDog/terraform-provider-datadog/pull/1347))
-   `datadog_integration_gcp`: Use mutex in GCP resource to limit concurrent changes ([#1360](https://github.com/DataDog/terraform-provider-datadog/pull/1360))
-   `datadog_integration_aws_lambda_arn`: Use mutex in aws lambda arn resource to limit concurrent changes ([#1370](https://github.com/DataDog/terraform-provider-datadog/pull/1370))
-   `datadog_aws_log_collection`: Use mutex to limit concurrent changes ([#1370](https://github.com/DataDog/terraform-provider-datadog/pull/1370))

NOTES:

-   Update Datadog client to [v1.10.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.10.0)

## 3.8.1 (January 19, 2022)

BUGFIXES:

-   `datadog_synthetics_test`: Handle empty `retry` option in API step ([#1332](https://github.com/DataDog/terraform-provider-datadog/pull/1332))

## 3.8.0 (January 18, 2022)

IMPROVEMENTS:

-   `datadog_dashboard`: Implement support for sunburst widget ([#1324](https://github.com/DataDog/terraform-provider-datadog/pull/1324))
-   `datadog_monitor`: Add support `ci-pipelines alert` monitor type ([#1315](https://github.com/DataDog/terraform-provider-datadog/pull/1315))
-   `datadog_security_monitoring_rules`: Raise the case limit in security rules ([#1313](https://github.com/DataDog/terraform-provider-datadog/pull/1313))
-   `datadog_service_level_objective`: Fix SLO-correction examples ([#1307](https://github.com/DataDog/terraform-provider-datadog/pull/1307))
-   `datadog_slo_correction`: Update documentation to list supported correction rules ([#1308](https://github.com/DataDog/terraform-provider-datadog/pull/1308))
-   `datadog_synthetics_test`: Add GET call after create to ensure resource is created successfully ([#1312](https://github.com/DataDog/terraform-provider-datadog/pull/1312))
-   `datadog_synthetics_test`: Add retry options to Synthetics multi step ([#1317](https://github.com/DataDog/terraform-provider-datadog/pull/1317))
-   `datadog_synthetics_test`: Add support for websocket synthetics tests ([#1287](https://github.com/DataDog/terraform-provider-datadog/pull/1287))
-   `datadog_synthetics_test`: Allow variables in `moreThan` operator with JSONPath ([#1322](https://github.com/DataDog/terraform-provider-datadog/pull/1322))

NOTES:

-   `datadog_application_key`: Deprecate `agent_rule` field ([#1318](https://github.com/DataDog/terraform-provider-datadog/pull/1318))
-   Update Datadog client to [v1.8.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.8.0)

## 3.7.0 (December 15, 2021)

IMPROVEMENTS:

-   `datadog_dashboard`: Implement support for formulas and functions in Scatterplot Widgets ([#1275](https://github.com/DataDog/terraform-provider-datadog/pull/1275))
-   `datadog_webhook`: Use mutex in webhook to prevent concurrent modifications ([#1279](https://github.com/DataDog/terraform-provider-datadog/pull/1279))
-   `datadog_webhook_custom_variable`: Use mutex in webhook to prevent concurrent modifications ([#1279](https://github.com/DataDog/terraform-provider-datadog/pull/1279))
-   `datadog_monitor`: Fix invalid monitor `timeout_h` example ([#1281](https://github.com/DataDog/terraform-provider-datadog/pull/1281))
-   `datadog_service_level_objective`: Mark `target_display` and `warning_display display` values as computed ([#1286](https://github.com/DataDog/terraform-provider-datadog/pull/1286))
-   `datadog_synthetics_test`: Add support for UDP tests ([#1277](https://github.com/DataDog/terraform-provider-datadog/pull/1277))
-   `datadog_dashboard`: Implement support for Change widgets using formulas and functions ([#1191](https://github.com/DataDog/terraform-provider-datadog/pull/1191))
-   `datadog_monitor`: Update `new_group_delay` and `new_host_delay` docs ([#1293](https://github.com/DataDog/terraform-provider-datadog/pull/1293))

BUGFIXES:

-   `datadog_dashboard_json`: Handle perpetual diff when both `is_read_only` and `restricted_roles` is set ([#1280](https://github.com/DataDog/terraform-provider-datadog/pull/1280))
-   `datadog_monitor_json`: Fix panic on resource name change ([#1278](https://github.com/DataDog/terraform-provider-datadog/pull/1278))
-   `datadog_monitor_json`: Fix perpetual diff on some monitor JSON fields ([#1291](https://github.com/DataDog/terraform-provider-datadog/pull/1291))
-   `datadog_dashboard_json`: Handle `notify_list` diffs for dashboard resource ([#1295](https://github.com/DataDog/terraform-provider-datadog/pull/1295))
-   `datadog_monitor`: Don't set `new_host_delay` if it's not returned by api ([#1281](https://github.com/DataDog/terraform-provider-datadog/pull/1281))
-   `datadog_dashboard`: Handle perpetual diff in `notify_list` attribute ([#1295](https://github.com/DataDog/terraform-provider-datadog/pull/1295))

NOTES:

-   Update Datadog client to [v1.7.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.7.0)
-   Update `terraform-plugin-sdk` to [v2.10.0](https://github.com/hashicorp/terraform-plugin-sdk/releases/tag/v2.10.0)

## 3.6.0 (November 10, 2021)

IMPROVEMENTS:

-   `datadog_metric_tag_configuration`: add aggregations option to `metric_tag_configuration` TF resource ([#1179](https://github.com/DataDog/terraform-provider-datadog/pull/1179))
-   `datadog_dashboard`: Safeguard against empty widget requests ([#1253](https://github.com/DataDog/terraform-provider-datadog/pull/1253))
-   `datadog_downtime`: Correct `downtime.monitor_tags` description ([#1252](https://github.com/DataDog/terraform-provider-datadog/pull/1252)) Thanks [@antonioalmeida](https://github.com/antonioalmeida)
-   `datadog_dashboard`: Update property descriptions for Dashboard RBAC release ([#1251](https://github.com/DataDog/terraform-provider-datadog/pull/1251))
-   `datadog_monitor`: Fix typo in the `datadog_monitor` page ([#1257](https://github.com/DataDog/terraform-provider-datadog/pull/1257)) Thanks [@jtamagnan-delphix](https://github.com/jtamagnan-delphix)
-   `datadog_slo_correction`: Add docs for recurring slo correction ([#1256](https://github.com/DataDog/terraform-provider-datadog/pull/1256))
-   `datadog_integration_azure`: Add support for Azure resource automute option ([#1262](https://github.com/DataDog/terraform-provider-datadog/pull/1262)).

FEATURES:

-   `datadog_logs_indexes_order`: Add datasource datadog_logs_indexes_order ([#1244](https://github.com/DataDog/terraform-provider-datadog/pull/1244))
-   `datadog_integration_azure`: Fix azure resource state when duplicate tenants are present ([#1255](https://github.com/DataDog/terraform-provider-datadog/pull/1255)).

BUGFIXES:

-   `datadog_dashboard`: Safeguard against empty widget requests ([#1253](https://github.com/DataDog/terraform-provider-datadog/pull/1253))
-   `datadog_synthetics_test`: Properly handle empty `basicAuth` values ([#1263](https://github.com/DataDog/terraform-provider-datadog/pull/1263))
-   `datadog_synthetics_test`: Handle empty `request_definition` values ([#1268](https://github.com/DataDog/terraform-provider-datadog/pull/1268))

NOTES:

-   Update Datadog client to [v1.6.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.6.0)

## 3.5.0 (October 25, 2021)

IMPROVEMENTS:

-   `datadog_dashboard`: Implement support for APM Dependency Stats query in Query Table ([#1199](https://github.com/DataDog/terraform-provider-datadog/pull/1199))
-   `datadog_synthetics_test`: Add missing follow redirects option for multistep requests ([#1194](https://github.com/DataDog/terraform-provider-datadog/pull/1194))
-   `datadog_dashboard`: Implement support for APM Resource Stats query in Query Table ([#1200](https://github.com/DataDog/terraform-provider-datadog/pull/1200))
-   `datadog_logs_custom_pipeline`: Document how to find pipeline ID for import ([#1220](https://github.com/DataDog/terraform-provider-datadog/pull/1220))
-   `datadog_security_monitoring_rules`: Add CWS support to terraform provider ([#1222](https://github.com/DataDog/terraform-provider-datadog/pull/1222))
-   `datadog_dashboard`: Set dashboard resource's widget attribute to Optional ([#1224](https://github.com/DataDog/terraform-provider-datadog/pull/1224))
-   `datadog_synthetics_test`: Add support for `servername` in Synthetics test request ([#1232](https://github.com/DataDog/terraform-provider-datadog/pull/1232))
-   `datadog_monitor`: Add support for new renotify options ([#1235](https://github.com/DataDog/terraform-provider-datadog/pull/1235))
-   `datadog_logs_index`: Use mutex to avoid creating/modifying logs indexes in parallel ([#1245](https://github.com/DataDog/terraform-provider-datadog/pull/1245))

FEATURES:

-   `datadog_webhook`: Add Webhook resource ([#1205](https://github.com/DataDog/terraform-provider-datadog/pull/1205))
-   `datadog_webhook_custom_variable`: Add Webhook Custom Variable resource ([#1206](https://github.com/DataDog/terraform-provider-datadog/pull/1206))
-   `datadog_roles`: Add datadog roles data source ([#1240](https://github.com/DataDog/terraform-provider-datadog/pull/1240))
-   `datadog_monitor_json`: Add datadog monitor json resource ([#1240](https://github.com/DataDog/terraform-provider-datadog/pull/1240))

BUGFIXES:

-   `datadog_synthetics_test`: Document `device_ids` as required for browser type only ([#1216](https://github.com/DataDog/terraform-provider-datadog/pull/1216)) Thanks [@alexjurkiewicz](https://github.com/alexjurkiewicz)
-   `datadog_synthetics_test`: Fix synthetics browser test `upload-files` step ([#1219](https://github.com/DataDog/terraform-provider-datadog/pull/1219))
-   `datadog_integration_gcp`: Changed Token URI for GCP Service account ([#1201](https://github.com/DataDog/terraform-provider-datadog/pull/1201)) Thanks [@pbrao08](https://github.com/pbrao08)
-   `datadog_downtime`: Set only one of timestamp or date format for start and end to avoid inconsistent plans ([#1223](https://github.com/DataDog/terraform-provider-datadog/pull/1223))
-   `datadog_security_monitoring_rules`: Fix docs and example for security monitoring default rule ([#1246](https://github.com/DataDog/terraform-provider-datadog/pull/1246))
-   `datadog_logs_index`: Specify 1 filter block ([#1247](https://github.com/DataDog/terraform-provider-datadog/pull/1247)) Thanks [@bendrucker](https://github.com/bendrucker)

NOTES:

-   Update Datadog client to [v1.5.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.5.0)

## 3.4.0 (September 16, 2021)

IMPROVEMENTS:

-   `datadog_dashboard_list`: Expand the example to demonstrate how to use the dashboard list data ([#1148](https://github.com/DataDog/terraform-provider-datadog/pull/1148)) Thanks [@jyee](https://github.com/jyee)
-   `datadog_synthetics_test`: Add support for local variables for browser tests ([#1185](https://github.com/DataDog/terraform-provider-datadog/pull/1185))
-   `datadog_integration_aws`: Document use of `access_key_id` as `account_id` for aws integrations ([#1189](https://github.com/DataDog/terraform-provider-datadog/pull/1189))
-   `datadog_dashboard`: Add available_values property to dashboard template variables ([#1195](https://github.com/DataDog/terraform-provider-datadog/pull/1195))
-   `datadog_user`: Update User roles when re-enabling previously deleted user ([#1174](https://github.com/DataDog/terraform-provider-datadog/pull/1174))

BUGFIXES:

-   `datadog_dashboard_json`: Validate widgets cast in dashboard JSON ([#1197](https://github.com/DataDog/terraform-provider-datadog/pull/1197))

NOTES:

-   Update Datadog client to [v1.4.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.4.0)

## 3.3.0 (August 26, 2021)

IMPROVEMENTS:

-   `datadog_dashboard`: Add audit logs data source to dashboard resource ([#1152](https://github.com/DataDog/terraform-provider-datadog/pull/1152))
-   `datadog_synthetics_test`: Improve consistency by using response from POST/PUT requests directly to save state ([#1117](https://github.com/DataDog/terraform-provider-datadog/pull/1117))
-   `datadog_logs_index`: Add logs index creation ([#1155](https://github.com/DataDog/terraform-provider-datadog/pull/1155))
-   `datadog_synthetics_test`: Add support for `allow_insecure` option in multistep requests ([#1145](https://github.com/DataDog/terraform-provider-datadog/pull/1145))
-   `datadog_synthetics_test`: Add `multistep` API test example ([#1164](https://github.com/DataDog/terraform-provider-datadog/pull/1164))
-   `datadog_synthetics_test`: Do not set useless fields for Synthetics local global variables ([#1175](https://github.com/DataDog/terraform-provider-datadog/pull/1175))
-   `datadog_monitor`: Add `new_group_delay` monitor option ([#1176](https://github.com/DataDog/terraform-provider-datadog/pull/1176))
-   `datadog_synthetics_test`: Add support for restricted roles for global variables ([#1178](https://github.com/DataDog/terraform-provider-datadog/pull/1178))
-   `datadog_dashboard`: Implement formulas and functions support for query table widgets ([#1158](https://github.com/DataDog/terraform-provider-datadog/pull/1158))

FEATURES:

-   `datadog_security_monitoring_filter`: Add security monitoring filter resource ([#1141](https://github.com/DataDog/terraform-provider-datadog/pull/1141))
-   `datadog_security_monitoring_filter`: Add security monitoring filter datasource ([#1142](https://github.com/DataDog/terraform-provider-datadog/pull/1142))
-   `datadog_synthetics_global_variable`: Add synthetics global variable datasource ([#1151](https://github.com/DataDog/terraform-provider-datadog/pull/1151))
-   `datadog_user`: Add datadog user datasource ([#1124](https://github.com/DataDog/terraform-provider-datadog/pull/1124)) Thanks [@tleveque69](https://github.com/tleveque69)
-   `datadog_api_key`: Add datadog api key resource ([#1184](https://github.com/DataDog/terraform-provider-datadog/pull/1184)) Thanks [@bartoszj-bcg](https://github.com/bartoszj-bcg)
-   `datadog_api_key`: Add datadog api key datasource ([#1184](https://github.com/DataDog/terraform-provider-datadog/pull/1184)) Thanks [@bartoszj-bcg](https://github.com/bartoszj-bcg)
-   `datadog_application_key`: Add datadog application key resource ([#1184](https://github.com/DataDog/terraform-provider-datadog/pull/1184)) Thanks [@bartoszj-bcg](https://github.com/bartoszj-bcg)
-   `datadog_application_key`: Add datadog application key datasource ([#1184](https://github.com/DataDog/terraform-provider-datadog/pull/1184)) Thanks [@bartoszj-bcg](https://github.com/bartoszj-bcg)
-   `datadog_child_organization`: Add datadog child organization resource ([#1184](https://github.com/DataDog/terraform-provider-datadog/pull/1184)) Thanks [@bartoszj-bcg](https://github.com/bartoszj-bcg)
-   `datadog_organization_settings`: Add datadog organization settings resource ([#1184](https://github.com/DataDog/terraform-provider-datadog/pull/1184)) Thanks [@bartoszj-bcg](https://github.com/bartoszj-bcg)

BUGFIXES:

-   `datadog_synthetics_test`: Fix missing integer type assertion targets ([#1161](https://github.com/DataDog/terraform-provider-datadog/pull/1161))
-   `datadog_dashboard`: Always set columns attribute when creating log stream widget ([#1163](https://github.com/DataDog/terraform-provider-datadog/pull/1163))
-   `datadog_dashboard_json`: Use custom SendRequest method to get a dashboard ([#1167](https://github.com/DataDog/terraform-provider-datadog/pull/1167))

NOTES:

-   Update Datadog client to [v1.3.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.3.0)

## 3.2.0 (July 14, 2021)

IMPROVEMENTS:

-   `datadog_security_monitoring`: Introduce the detections methods and the new value rules options ([#1116](https://github.com/DataDog/terraform-provider-datadog/pull/1116))
-   `datadog_integration_slack_channel`: Add Slack import example ([#1128](https://github.com/DataDog/terraform-provider-datadog/pull/1128))
-   `datadog_synthetics_test`: Add custom message with a warning on synthetics test resource ([#1133](https://github.com/DataDog/terraform-provider-datadog/pull/1133))
-   `datadog_monitor`: Allow un-setting `restricted_roles` on a monitor resource and add `restricted_roles` to the monitor data source ([#1121](https://github.com/DataDog/terraform-provider-datadog/pull/1121))
-   `datadog_security_monitoring_rules`: Add support for suppress and require filters for the rules resources ([#1131](https://github.com/DataDog/terraform-provider-datadog/pull/1131))
-   `datadog_security_monitoring_rules`: Add support for `has_extended_title` property ([#1130](https://github.com/DataDog/terraform-provider-datadog/pull/1130))

BUGFIXES:

-   `datadog_dashboard`: Fix `is_column_break` issues ([#1140](https://github.com/DataDog/terraform-provider-datadog/pull/1140))
-   `datadog_integration_aws_log_collection` and `datadog_integration_aws_tag_filter`: Fixed Terraform examples ([#1127](https://github.com/DataDog/terraform-provider-datadog/pull/1127))
-   `datadog_synthetics_test`: Update `tick_every` property to use int ([#1119](https://github.com/DataDog/terraform-provider-datadog/pull/1119))
-   `datadog_logs_index`: Fix logs_index update method ([#1126](https://github.com/DataDog/terraform-provider-datadog/pull/1126)
-   `provider`: Fix segfault in `translateclienterror` if `httpresp` is nil ([#1135](https://github.com/DataDog/terraform-provider-datadog/pull/1135))

NOTES:

-   Update Datadog client to [v1.2.0](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.2.0) ([#1143](https://github.com/DataDog/terraform-provider-datadog/pull/1143))
-   Update Terraform plugin SDK to v2.7.0 ([#1132](https://github.com/DataDog/terraform-provider-datadog/pull/1132))

## 3.1.2 (June 24, 2021)

BUGFIXES:

-   `datadog_integration_aws`: Handle all characters for AWS Role Name. ([#1122](https://github.com/DataDog/terraform-provider-datadog/pull/1122))

## 3.1.1 (June 22, 2021)

IMPROVEMENTS:

-   `datadog_integration_aws_tag_filter`: Remove US only constraint from docs. ([#1118](https://github.com/DataDog/terraform-provider-datadog/pull/1118))

BUGFIXES:

-   `datadog_logs_index`: Fix retention_days and daily_limit attributes. ([#1118](https://github.com/DataDog/terraform-provider-datadog/pull/1118))

## 3.1.0 (June 17, 2021)

IMPROVEMENTS:

-   `datadog_logs_index`: Adding missing retention_days and daily_limit parameters. ([#1083](https://github.com/DataDog/terraform-provider-datadog/pull/1083)) Thanks [@DimitryVlasenko](https://github.com/DimitryVlasenko)
-   `datadog_dashboard`: Add support for WidgetCustomLink `is_hidden` and `override_label` properties. ([#1062](https://github.com/DataDog/terraform-provider-datadog/pull/1062))
-   `datadog_synthetics_test`: Add support for monitor name and priority. ([#1104](https://github.com/DataDog/terraform-provider-datadog/pull/1104))
-   `datadog_integration_aws`: Add support for access_key_id and secret_access_key. ([#1101](https://github.com/DataDog/terraform-provider-datadog/pull/1101)).
-   `datadog_dashboard`: Update dashboard examples. ([#1105](https://github.com/DataDog/terraform-provider-datadog/pull/1105))
-   `datadog_synthetics_test`: Add support for global variables in config variables. ([#1106](https://github.com/DataDog/terraform-provider-datadog/pull/1106))
-   `datadog_dashboard_json`: Add dashboard list support. ([#1102](https://github.com/DataDog/terraform-provider-datadog/pull/1102))
-   `datadog_downtime`: Properly handle recurring downtimes definitions. ([#1092](https://github.com/DataDog/terraform-provider-datadog/pull/1092))
-   `datadog_dashboard`: Dashboard RBAC roles. ([#1109](https://github.com/DataDog/terraform-provider-datadog/pull/1109))

BUGFIXES:

-   `datadog_integration_aws`: Properly catch error response from AWS Logs integration. ([#1095](https://github.com/DataDog/terraform-provider-datadog/pull/1095))
-   `datadog_integration_aws`: Handle empty parameters in AWS and Azure integrations. ([#1096](https://github.com/DataDog/terraform-provider-datadog/pull/1096)).
-   `datadog_integration_azure`: Handle empty parameters in AWS and Azure integrations. ([#1096](https://github.com/DataDog/terraform-provider-datadog/pull/1096)).
-   `datadog_monitor`: Re-introduce monitor type diff suppression for query/metric alerts. ([#1099](https://github.com/DataDog/terraform-provider-datadog/pull/1099))
-   `datadog_synthetics_test`: Allow zero value for dns_server_port. ([#1087](https://github.com/DataDog/terraform-provider-datadog/pull/1087))

NOTES:

-   Update Datadog api go client. See [here](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.1.0) for changes.

## 3.0.0 (May 27, 2021)

IMPROVEMENTS:

-   Upgrade terraform-plugin-sdk to v2. See https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html for all the internal changes.

NOTES:

-   `datadog_monitor`: Remove the `threshold` deprecated property.
-   `datadog_monitor`: Remove the `threshold_windows` deprecated property.
-   `datadog_monitor`: Remove the `silenced` deprecated property.
-   `datadog_dashboard`: Remove the `layout` deprecated property from widgets.
-   `datadog_dashboard`: Remove the `time` deprecated property from widgets.
-   `datadog_dashboard`: Remove the `logset` deprecated property from widgets.
-   `datadog_dashboard`: Remove the `count` deprecated property from widgets.
-   `datadog_dashboard`: Remove the `start` deprecated property from widgets.
-   `datadog_dashboard`: Remove the `compute` deprecated property from widgets.
-   `datadog_dashboard`: Remove the `search` deprecated property from widgets.
-   `datadog_integration_pagerduty`: Remove the `services` deprecated property.
-   `datadog_logs_archive`: Remove the `s3` deprecated property.
-   `datadog_logs_archive`: Remove the `azure` deprecated property.
-   `datadog_logs_archive`: Remove the `gcs` deprecated property.
-   `datadog_screenboard`: Remove the deprecated resource
-   `datadog_service_level_objective`: Remove the `monitor_search` deprecated property.
-   `datadog_timeboard`: Remove the deprecated resource.
-   `datadog_synthetics_test`: Remove the `request` deprecated property.
-   `datadog_synthetics_test`: Remove the `assertions` deprecated property.
-   `datadog_synthetics_test`: Remove the `options` deprecated property.
-   `datadog_synthetics_test`: Remove the `step` deprecated property.
-   `datadog_synthetics_test`: Remove the `variable` deprecated property.
-   `datadog_user`: Remove the `handle` deprecated property.
-   `datadog_user`: Remove the `is_admin` deprecated property.
-   `datadog_user`: Remove the `access_role` deprecated property.
-   `datadog_user`: Remove the `role` deprecated property.

## 2.26.1 (May 20, 2021)

BUGFIXES:

-   `datadog_dashboard_json`: Fix `dashboard` attribute retrieval when calling the update method ([#1072](https://github.com/DataDog/terraform-provider-datadog/pull/1072))

## 2.26.0 (May 18, 2021)

IMPROVEMENTS:

-   `datadog_logs_custom_pipeline`: Add mutex to logs custom pipeline resource ([#1069](https://github.com/DataDog/terraform-provider-datadog/pull/1069))
-   `datadog_logs_custom_pipeline`: Use code formatting in description for attribute re-mapper ([#1061](https://github.com/DataDog/terraform-provider-datadog/pull/1061))
-   `datadog_monitor`: Update monitor critical threshold documentation ([#1055](https://github.com/DataDog/terraform-provider-datadog/pull/1055))
-   `datadog_monitor`: Retry on 504's when validating monitors ([#1038](https://github.com/DataDog/terraform-provider-datadog/pull/1038))
-   `datadog_dashboard_json`: Ignore widget IDs for diff on dashboard JSON resource ([#1028](https://github.com/DataDog/terraform-provider-datadog/pull/1028))
-   `datadog_monitor`: Add monitors datasource for multiple monitors ([#1048](https://github.com/DataDog/terraform-provider-datadog/pull/1048))
-   `datadog_synthetics_test`: Add support for setCookie, dnsServerPort, allowFailure and isCritical fields for Synthetics tests ([#1052](https://github.com/DataDog/terraform-provider-datadog/pull/1052))
-   `datadog_dashboard`: Add new properties to group widget, note widget and image widget ([#1044](https://github.com/DataDog/terraform-provider-datadog/pull/1044))
-   `datadog_synthetics_test`: Add support for icmp tests ([#1030](https://github.com/DataDog/terraform-provider-datadog/pull/1030))
-   `datadog_dashboard`: Implement formulas and functions for geomap widgets ([#1043](https://github.com/DataDog/terraform-provider-datadog/pull/1043))
-   `datadog_dashboard`: Formula and Function support for Toplist Widgets in Dashboard resource ([#951](https://github.com/DataDog/terraform-provider-datadog/pull/951))
-   `datadog_dashboard`: Add reflow_type property for dashboards ([#1017](https://github.com/DataDog/terraform-provider-datadog/pull/1017))
-   `datadog_dashboard`: Formula and Function support for Query Value Widgets in Dashboard resource ([#953](https://github.com/DataDog/terraform-provider-datadog/pull/953))

FEATURES:

-   `datadog_service_level_objective`: Add SLO data sources ([#931](https://github.com/DataDog/terraform-provider-datadog/pull/931))

BUGFIXES:

-   `datadog_downtime`: Properly mark active/disabled fields as readonly to avoid diffs ([#1034](https://github.com/DataDog/terraform-provider-datadog/pull/1034))
-   `datadog_integration_aws`: Mark AWS account as non existent if GET returns 400 when AWS integration not installed ([#1047](https://github.com/DataDog/terraform-provider-datadog/pull/1047))

NOTES:

-   Use custom transport for HTTPClient to enable retries on 429 and 5xx http errors ([#1054](https://github.com/DataDog/terraform-provider-datadog/pull/1054))
-   Update Datadog api go client. See [here](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.0.0-beta.22) for changes.

## 2.25.0 (April 15, 2021)

IMPROVEMENTS:

-   `datadog_slo_correction`: Add docs for SLO Correction resource ([#1021](https://github.com/DataDog/terraform-provider-datadog/pull/1021))
-   `datadog_synthetics_test`: Use new API models for api tests ([#1005](https://github.com/DataDog/terraform-provider-datadog/pull/1005))
-   `datadog_monitor`: Improve consistency by using response from POST/PUT requests ([#1015](https://github.com/DataDog/terraform-provider-datadog/pull/1015))
-   `datadog_synthetics_test`: Add `noSavingResponseBody` and `noScreenshot` fields ([#1012](https://github.com/DataDog/terraform-provider-datadog/pull/1012))
-   `datadog_logs_metric`: Add `group_by` block to logs_metric example ([#1010](https://github.com/DataDog/terraform-provider-datadog/pull/1010))

FEATURES:

-   `datadog_dashboard`: Add support for Formula and Function support for Timeseries Widgets ([#892](https://github.com/DataDog/terraform-provider-datadog/pull/892))
-   `datadog_synthetics_test`: Add support for `multi` step synthetics API tests ([#1007](https://github.com/DataDog/terraform-provider-datadog/pull/1007))
-   `datadog_security_monitoring_default_rule`: Add datadog default security monitoring rule filters ([#965](https://github.com/DataDog/terraform-provider-datadog/pull/965))
-   `datadog_synthetics_test`: Add support for global_time_target for SLO widgets ([#1003](https://github.com/DataDog/terraform-provider-datadog/pull/1003))

BUGFIXES:

-   `datadog_synthetics_test`: Set `tick_every` as required and add default value for `example` field ([#1020](https://github.com/DataDog/terraform-provider-datadog/pull/1020))
-   `datadog_monitor`: Fix diff suppression for field `restricted_roles` ([#1011](https://github.com/DataDog/terraform-provider-datadog/pull/1011))
-   `datadog_integration_slack_channel`: Fix `account_id` field not being set on imports ([#1019](https://github.com/DataDog/terraform-provider-datadog/pull/1019))
-   `datadog_synthetics_test`: Fix error when passing empty step param ([#1014](https://github.com/DataDog/terraform-provider-datadog/pull/1014))
-   `datadog_integration_gcp`: Set ForceNew to true on non-updatable GCP resource fields ([#1014](https://github.com/DataDog/terraform-provider-datadog/pull/1007))
-   `datadog_dashboard`: Add retry on 502's when listing dashbaord ([#1006](https://github.com/DataDog/terraform-provider-datadog/pull/1006))

NOTES:

-   Update the underlying Datadog go client to v1.0.0-beta.19. See [here](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.0.0-beta.19) for changes.

## 2.24.0 (March 22, 2021)

IMPROVEMENTS:

-   `datadog_dashboard`: Add `legend_layout` and `legend_columns` to timeseries widget definition ([#992](https://github.com/Datadog/terraform-provider-datadog/pull/992)).

FEATURES:

-   `datadog_metric_tag_configuration` Add new resource ([#960](https://github.com/Datadog/terraform-provider-datadog/pull/960)).

## 2.23.0 (March 16, 2021)

IMPROVEMENTS:

-   `datadog_dashboard`: Implement support for Geomap Dashboard Widget ([#954](https://github.com/Datadog/terraform-provider-datadog/pull/954)).

FEATURES:

-   `datadog_dashboard_json`: Add new dashboard JSON resource ([#950](https://github.com/Datadog/terraform-provider-datadog/pull/950)).

BUGFIXES:

-   `datadog_dashboard`: Add a retry on 504 errors when there is a timeout ([#975](https://github.com/Datadog/terraform-provider-datadog/pull/975)).
-   `datadog_integration_slack_channel`: Fix issue causing slack channels to not be created in some situations ([#981](https://github.com/Datadog/terraform-provider-datadog/pull/981)).
-   `datadog_monitor`: Explicitly check `monitor_id` for `nil` value to fix an issue with terraformer ([#962](https://github.com/Datadog/terraform-provider-datadog/pull/962)).
-   `datadog_security_monitoring_default_rule`: Fix issue that prevented default rule cases notifications to be updated ([#956](https://github.com/Datadog/terraform-provider-datadog/pull/956)).

NOTES:

-   Update the underlying Datadog go client to v1.0.0-beta.17. See [here](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.0.0-beta.17) for changes.

## 2.22.0 (March 3, 2021).

IMPROVEMENTS:

-   `datadog_dashboard`: Improve consistency by using response from POST/PUT requests directly to save state ([#909](https://github.com/Datadog/terraform-provider-datadog/pull/909)).
-   `datadog_downtime`: Improve consistency by using response from POST/PUT requests directly to save state ([#905](https://github.com/Datadog/terraform-provider-datadog/pull/905)).
-   `datadog_ip_ranges`: Add support for reading ipv4/6 prefixes by location for synthetics ([#934](https://github.com/Datadog/terraform-provider-datadog/pull/934)).
-   `datadog_logs_archive_order`: Improve consistency by using response from POST/PUT requests directly to save state ([#912](https://github.com/Datadog/terraform-provider-datadog/pull/912)).
-   `datadog_logs_archive`: Improve consistency by using response from POST/PUT requests directly to save state ([#912](https://github.com/Datadog/terraform-provider-datadog/pull/912)).
-   `datadog_logs_custom_pipeline`: Improve consistency by using response from POST/PUT requests directly to save state ([#913](https://github.com/Datadog/terraform-provider-datadog/pull/913)).
-   `datadog_logs_index_order`: Improve consistency by using response from POST/PUT requests directly to save state ([#915](https://github.com/Datadog/terraform-provider-datadog/pull/915)).
-   `datadog_logs_index`: Improve consistency by using response from POST/PUT requests directly to save state ([#915](https://github.com/Datadog/terraform-provider-datadog/pull/915)).
-   `datadog_logs_integration_pipeline`: Improve consistency by using response from POST/PUT requests directly to save state ([#913](https://github.com/Datadog/terraform-provider-datadog/pull/913)).
-   `datadog_logs_metric`: Improve consistency by using response from POST/PUT requests directly to save state ([#917](https://github.com/Datadog/terraform-provider-datadog/pull/917)).
-   `datadog_logs_pipeline_order`: Improve consistency by using response from POST/PUT requests directly to save state ([#913](https://github.com/Datadog/terraform-provider-datadog/pull/913)).
-   `datadog_metric_metadata`: Improve consistency by using response from POST/PUT requests directly to save state ([#922](https://github.com/Datadog/terraform-provider-datadog/pull/922)).
-   `datadog_monitor`: Add support for `groupby_simple_monitor` option to monitor resource ([#952](https://github.com/Datadog/terraform-provider-datadog/pull/952)).
-   `datadog_monitor`: Improve consistency by using response from POST/PUT requests directly to save state ([#901](https://github.com/Datadog/terraform-provider-datadog/pull/901)).
-   `datadog_role`: Improve consistency by using response from POST/PUT requests directly to save state ([#925](https://github.com/Datadog/terraform-provider-datadog/pull/925)).
-   `datadog_service_level_objective`: Improve consistency by using response from POST/PUT requests directly to save state ([#910](https://github.com/Datadog/terraform-provider-datadog/pull/910)).
-   `datadog_slo_correction`: Improve consistency by using response from POST/PUT requests directly to save state ([#921](https://github.com/Datadog/terraform-provider-datadog/pull/921)).
-   `datadog_user`: Improve consistency by using response from POST/PUT requests directly to save state ([#927](https://github.com/Datadog/terraform-provider-datadog/pull/927)).

FEATURES:

-   `datadog_integration_slack_channel`: Add support for slack channel resource ([#932](https://github.com/Datadog/terraform-provider-datadog/pull/932)).

BUGFIXES:

-   `datadog_dashboard`: Fix template_variable_presets to support optional template_variables ([#944](https://github.com/Datadog/terraform-provider-datadog/pull/944)).

NOTES:

-   `datadog_integration_pagerduty`: Remove deprecation on PagerDuty resource ([#930](https://github.com/Datadog/terraform-provider-datadog/pull/930)).
-   Update the underlying Datadog go client to v1.0.0-beta.16. See [here](https://github.com/DataDog/datadog-api-client-go/releases/tag/v1.0.0-beta.16) for changes.

## 2.21.0 (February 9, 2021)

IMPROVEMENTS:

-   `datadog_integration_aws_filter`: Add new resource ([#881](https://github.com/Datadog/terraform-provider-datadog/pull/881)).
-   `datadog_slo_correction`: Add new resource ([#866](https://github.com/Datadog/terraform-provider-datadog/pull/866)).

FEATURES:

-   `datadog_monitor`: Add restricted roles. ([#883](https://github.com/Datadog/terraform-provider-datadog/pull/883)).
-   `datadog_synthetics_test`: Add parameter to prevent useless diffs for browser tests ([#854](https://github.com/Datadog/terraform-provider-datadog/pull/854)).
-   `datadog_synthetics_test`: Add new `browser_step` field for browser tests ([#849](https://github.com/Datadog/terraform-provider-datadog/pull/849)).

BUGFIXES:

-   `datadog_synthetics_global_variable`: Fix setting `parse_test_options` attribute ([#867](https://github.com/Datadog/terraform-provider-datadog/pull/867)).
-   `datadog_security_monitoring_rule`: Fix enabled attribute retrieval ([#862](https://github.com/Datadog/terraform-provider-datadog/pull/862)).
-   `datadog_metric_metadata`: Fix id retrieval when calling the read function ([#856](https://github.com/Datadog/terraform-provider-datadog/pull/856)).
-   `datadog_logs_custom_pipeline`: Support empty strings for filter query ([#855](https://github.com/Datadog/terraform-provider-datadog/pull/855)).
-   `datadog_dashboard`: Handle crash in `timeseries_definition` ([#863](https://github.com/Datadog/terraform-provider-datadog/pull/863)).
-   `datadog_synthetics_test`: Turn locations into a set ([#864](https://github.com/Datadog/terraform-provider-datadog/pull/864)).

NOTES:

-   `datadog_dashboard`: Deprecate TypeMap complex fields ([#853](https://github.com/Datadog/terraform-provider-datadog/pull/853)).
-   `datadog_synthetics_test`: Deprecate TypeMap field ([#870](https://github.com/Datadog/terraform-provider-datadog/pull/870)).
-   `datadog_monitor` : Include SDK when a tag is a unexpected prefix ([#781](https://github.com/DataDog/terraform-provider-datadog/issues/781)).
-   Backport performance fix to SDK v1.

## 2.20.0 (January 20, 2021)

IMPROVEMENTS:

-   `datadog_logs_metrics`: Add new resource ([#823](https://github.com/Datadog/terraform-provider-datadog/pull/823)).

FEATURES:

-   `datadog_dashboard`: Store dashboard widget IDs ([#815](https://github.com/Datadog/terraform-provider-datadog/pull/815)).
-   `datadog_synthetics_test`: Add support for global variables from test ([#831](https://github.com/Datadog/terraform-provider-datadog/pull/831)).

BUGFIXES:

-   `datadog_synthetics_test`: Store SHA 256 hash of certificates in state instead of the actual cert ([#835](https://github.com/Datadog/terraform-provider-datadog/pull/835)).

NOTES:

-   `datadog_user`: Deprecate `access_role` field ([#834](https://github.com/Datadog/terraform-provider-datadog/pull/834)).
-   `datadog_monitor`: Provide alternative to TypeMap complex fields ([#833](https://github.com/Datadog/terraform-provider-datadog/pull/833)).
-   `datadog_logs_archive`: Provide alternative to TypeMap complex fields ([#838](https://github.com/Datadog/terraform-provider-datadog/pull/838)).

## 2.19.1 (January 8, 2021)

BUGFIXES:

-   `datadog_monitor`: Handle 404 properly with retry ([#824](https://github.com/DataDog/terraform-provider-datadog/pull/824)).
-   `datadog_integration_aws`: Remove incorrect deprecation warning ([#820](https://github.com/DataDog/terraform-provider-datadog/pull/820)).

## 2.19.0 (January 7, 2021)

FEATURES:

-   `datadog_synthetics_test`: Add support for config variables ([#807](https://github.com/DataDog/terraform-provider-datadog/pull/807)).

BUGFIXES:

-   `datadog_user`: Add ability to send user invitations in v2 API ([#814](https://github.com/DataDog/terraform-provider-datadog/pull/814)).
-   `datadog_monitor`: Fix updating priorities. ([#804](https://github.com/DataDog/terraform-provider-datadog/pull/804)).
-   `datadog_monitor`: Add retry on 502 for get and validate ([#816](https://github.com/DataDog/terraform-provider-datadog/pull/816)).
-   `datadog_synthetics_test`: Fix error when setting status code assertion with regex ([#784](https://github.com/DataDog/terraform-provider-datadog/pull/784)).
-   `datadog_logs_index_order`: Enable `UpdateLogsIndexOrder` operation ([#790](https://github.com/DataDog/terraform-provider-datadog/pull/790)).
-   Validate enum values ([#794](https://github.com/DataDog/terraform-provider-datadog/pull/794)).

NOTES:

-   Remove deprecated `ExistsFunc` usage ([#805](https://github.com/DataDog/terraform-provider-datadog/pull/805)).

## 2.18.1 (December 9, 2020)

BUGFIXES:

-   `datadog_user`: Automatically upgrade users when `roles` is set ([#778](https://github.com/DataDog/terraform-provider-datadog/pull/778)).
-   `datadog_dashboard`: Add ForceNew to `layout_type` dashboard attribute ([#774](https://github.com/DataDog/terraform-provider-datadog/pull/774)).

## 2.18.0 (December 8, 2020)

IMPROVEMENTS:

-   `datadog_synthetics_private_location`: Add support for synthetics private locations ([#761](https://github.com/DataDog/terraform-provider-datadog/pull/761)).
-   `datadog_security_monitoring_rule`: Add support for security monitoring rules ([#763](https://github.com/DataDog/terraform-provider-datadog/pull/763)).

FEATURES:

-   `datadog_service_level_objective`: Add `force_delete` attribute, to manage deletion in dashboard references ([#771](https://github.com/DataDog/terraform-provider-datadog/pull/771)).
-   `datadog_synthetics_global_variable`: Add support for secure global variables ([#758](https://github.com/DataDog/terraform-provider-datadog/pull/758)).

BUGFIXES:

-   `datadog_synthetics_test`: Handle numbers in `targetvalue` for synthetics assertions ([#766](https://github.com/DataDog/terraform-provider-datadog/pull/766)).

NOTES:

-   `datadog_user`: Use v2 API. This deprecates several v1 only attributes ([#752](https://github.com/DataDog/terraform-provider-datadog/pull/752)).

## 2.17.0 (November 24, 2020)

FEATURES:

-   `datadog_role`: Add role datasource ([#751](https://github.com/DataDog/terraform-provider-datadog/pull/751))
-   `datadog_role`: Add roles resource and permissions datasource ([#753](https://github.com/DataDog/terraform-provider-datadog/pull/753)).

BUGFIXES:

-   `datadog_dashboard`: Handle multiple dashboards correctly in the datasource ([#759](https://github.com/DataDog/terraform-provider-datadog/pull/759)).
-   `datadog_synthetics_test`: Set client certificate content as sensitive ([#750](https://github.com/DataDog/terraform-provider-datadog/pull/750)).
-   `datadog_monitor`: Fix monitor `no_data_timeframe` import ([#748](https://github.com/DataDog/terraform-provider-datadog/pull/748)).

## 2.16.0 (November 9, 2020)

IMPROVEMENTS:

-   `datadog_dashboard`: Add new data source ([#734](https://github.com/DataDog/terraform-provider-datadog/pull/734)).

BUGFIXES:

-   `datadog_dashboard`: Update go client to get new palette values ([#743](https://github.com/DataDog/terraform-provider-datadog/pull/743)).

## 2.15.0 (November 2, 2020)

IMPROVEMENTS:

-   `datadog_monitor`: Add `priority`field ([#729](https://github.com/DataDog/terraform-provider-datadog/pull/729)).

BUGFIXES:

-   `datadog_synthetics_test`: Handle missing variables field from API response ([#733](https://github.com/DataDog/terraform-provider-datadog/pull/733)).
-   `datadog_monitor`: Handle `0` in `new_host_delay` ([#726](https://github.com/DataDog/terraform-provider-datadog/pull/726)).

NOTES:

-   `provider`: Replace 4d63.com/tz with time/tzdata. It means go 1.15 is required now to build the provider ([#728](https://github.com/DataDog/terraform-provider-datadog/pull/728)).

## 2.14.0 (October 27, 2020)

FEATURES:

-   `datadog_logs_archive_order`: Add a new resource to reorder logs archives ([#694](https://github.com/DataDog/terraform-provider-datadog/pull/694)).
-   `datadog_synthetics_global_variable`: Add a new resource to support global variables in synthetics tests ([#675](https://github.com/DataDog/terraform-provider-datadog/pull/675)).

IMPROVEMENTS:

-   `datadog_dashboard`: Add support for `apm_stats_query` request type in widgets ([#676](https://github.com/DataDog/terraform-provider-datadog/pull/676)).
-   `datadog_dashboard`: Add support for dual y-axis for timeseries widgets ([#685](https://github.com/DataDog/terraform-provider-datadog/pull/685)).
-   `datadog_dashboard`: Add support for `has_search_bar` and `cell_display_mode` properties on widgets ([#686](https://github.com/DataDog/terraform-provider-datadog/pull/686)).
-   `datadog_dashboard`: Add support for `custom_links` property on widgets ([#696](https://github.com/DataDog/terraform-provider-datadog/pull/696)).
-   `datadog_logs_archive`: Add `rehydration_tags` property ([#705](https://github.com/DataDog/terraform-provider-datadog/pull/705)).
-   `datadog_logs_archive`: Add `include_tags` property ([#715](https://github.com/DataDog/terraform-provider-datadog/pull/715)).
-   `datadog_logs_custom_pipeline`: Add `target_format` property to the Logs attribute remapper ([#682](https://github.com/DataDog/terraform-provider-datadog/pull/682)).
-   `datadog_service_level_objective`: Add validate option ([#672](https://github.com/DataDog/terraform-provider-datadog/pull/672)).
-   `datadog_synthetics_test`: Add support for DNS tests ([#673](https://github.com/DataDog/terraform-provider-datadog/pull/673)).
-   `datadog_synthetics_test`: Add support for global variables ([#691](https://github.com/DataDog/terraform-provider-datadog/pull/691)).
-   `datadog_synthetics_test`: Add support for `dns_server` and `request_client_certificate` properties ([#711](https://github.com/DataDog/terraform-provider-datadog/pull/711)).

BUGFIXES:

-   `datadog_synthetics_test`: Don't ignore options diff ([#707](https://github.com/DataDog/terraform-provider-datadog/pull/707)).
-   `datadog_synthetics_test`: Make `tags` property optional ([#712](https://github.com/DataDog/terraform-provider-datadog/pull/712)).
-   `datadog_ip_ranges`: Support EU site ([#713](https://github.com/DataDog/terraform-provider-datadog/pull/713)).

## 2.13.0 (September 16, 2020)

FEATURES:

-   `datadog_dashboard_list`: Add a new datasource for dashboard lists ([#657](https://github.com/DataDog/terraform-provider-datadog/pull/657)).
-   `datadog_synthetics_locations`: Add a new datasource for locations ([#309](https://github.com/DataDog/terraform-provider-datadog/pull/309)).

IMPROVEMENTS:

-   `datadog_dashboard`: A new `dashboard_lists` attribute allows adding dashboard to dashboard lists in the resource itself ([#654](https://github.com/DataDog/terraform-provider-datadog/pull/654)).
-   `datadog_dashboard`: Add support for `multi_compute` attribute ([#629](https://github.com/DataDog/terraform-provider-datadog/pull/629)).
-   `datadog_dashboard`: Add support for `metric` in `conditional_formats` ([#617](https://github.com/DataDog/terraform-provider-datadog/pull/617)).
-   `datadog_dashboard`: Add support for `rum_query` and `security_query` widget requests ([#416](https://github.com/DataDog/terraform-provider-datadog/pull/416)).
-   `datadog_monitor`: Monitors are now validated during plan ([#639](https://github.com/DataDog/terraform-provider-datadog/pull/639)).
-   `datadog_downtime`: Add support for recurrent rules ([#610](https://github.com/DataDog/terraform-provider-datadog/pull/610)).
-   `datadog_synthetics_test`: Add support for steps for browser tests ([#638](https://github.com/DataDog/terraform-provider-datadog/pull/638)).
-   `datadog_synthetics_test`: Add subtype TCP test support for API tests ([#632](https://github.com/DataDog/terraform-provider-datadog/pull/632)).
-   `datadog_synthetics_test`: Add retry and monitor options ([#636](https://github.com/DataDog/terraform-provider-datadog/pull/636)).

BUGFIXES:

-   `datadog_dashboard`: Prevent nil pointer dereference with template variables without prefix ([#630](https://github.com/DataDog/terraform-provider-datadog/pull/630)).
-   `datadog_dashboard`: Don't allow empty content in note widgets ([#607](https://github.com/DataDog/terraform-provider-datadog/pull/607)).
-   `datadog_downtime`: Ignore useless diff on start attribute ([#597](https://github.com/DataDog/terraform-provider-datadog/pull/597)).
-   `datadog_logs_custom_pipeline`: Don't allow empty pipeline filter ([#605](https://github.com/DataDog/terraform-provider-datadog/pull/605)).
-   `provider`: Completely skip creds validation when validate is false ([#641](https://github.com/DataDog/terraform-provider-datadog/pull/641)).

NOTES:

-   `datadog_synthetics_test`: The `options` attribute has been deprecated by `options_list` ([#624](https://github.com/DataDog/terraform-provider-datadog/pull/624)).

## 2.12.1 (July 23, 2020)

This release doesn't contain any user-facing changes. It's done as a required part of process to finalize the transfer of the provider repository under DataDog GitHub organization: https://github.com/DataDog/terraform-provider-datadog.

## 2.12.0 (July 22, 2020)

FEATURES:

-   `datadog_monitor`: Add new datasource for monitors ([#569](https://github.com/DataDog/terraform-provider-datadog/issues/569)), ([#585](https://github.com/DataDog/terraform-provider-datadog/issues/585)).

IMPROVEMENTS:

-   `datadog_synthetics_test`: Enable usage of `validatesJSONPath` operator ([#571](https://github.com/DataDog/terraform-provider-datadog/issues/571)).
-   `datadog_synthetics_test`: Allow usage of the new assertion format ([#571](https://github.com/DataDog/terraform-provider-datadog/issues/571)), ([#582](https://github.com/DataDog/terraform-provider-datadog/issues/582)).
-   `datadog_synthetics_test`: Add support for `basicAuth` and `query` ([#586](https://github.com/DataDog/terraform-provider-datadog/issues/586)).

BUGFIXES:

-   `datadog_downtime`: Replace `time.LoadLocation` by tz.LoadLocation from `4d63.com/tz` package ([#560](https://github.com/DataDog/terraform-provider-datadog/issues/560)).
-   `datadog_downtime`: Use `TypeSet` for monitor tags to avoid unnecessary diffs ([#540](https://github.com/DataDog/terraform-provider-datadog/issues/540)).
-   `provider`: Respect the debug setting in the new Go Datadog client ([#580](https://github.com/DataDog/terraform-provider-datadog/issues/580)).

NOTES:

-   `datadog_integration_pagerduty`: This resource is deprecated. You can use `datadog_integration_pagerduty_service_object` resources directly once the integration is activated ([#584](https://github.com/DataDog/terraform-provider-datadog/issues/584)).

## 2.11.0 (June 29, 2020)

FEATURES:

-   `datadog_logs_archive`: Add `datadog_logs_archive` resource ([#544](https://github.com/DataDog/terraform-provider-datadog/pull/544)).
-   `datadog_integration_azure`: Add `datadog_integration_azure` resource ([#556](https://github.com/DataDog/terraform-provider-datadog/pull/556)).

## 2.10.0 (June 26, 2020)

FEATURES:

-   `datadog_integration_aws`: Add `excluded_regions` parameter ([#549](https://github.com/DataDog/terraform-provider-datadog/pull/549)).
-   `datadog_dashboard`: Add `ServiceMap` widget to dashboard ([#550](https://github.com/DataDog/terraform-provider-datadog/pull/550)).
-   `datadog_dashboard`: Add `show_legend` and `legend_size` fields to Distribution widget ([#551](https://github.com/DataDog/terraform-provider-datadog/pull/551)).
-   `datadog_dashboard`: Add `network_query` and `rum_query` to timeseries widget ([#555](https://github.com/DataDog/terraform-provider-datadog/pull/555)).
-   `datadog_dashboard`: Add `event`, `legend_size` and `show_legend` fields to heatmap widget ([#554](https://github.com/DataDog/terraform-provider-datadog/pull/554)).

IMPROVEMENTS:

-   `datadog_dashboard`: Add readonly url field to dashboard ([#558](https://github.com/DataDog/terraform-provider-datadog/pull/558)).

## 2.9.0 (June 22, 2020)

IMPROVEMENTS:

-   `datadog_monitor`: Add monitor `force_delete` parameter ([#535](https://github.com/DataDog/terraform-provider-datadog/pull/535)) Thanks [@ykyr](https://github.com/ykyr)

BUGFIXES:

-   `datadog_dashboard`: Safely access index field ([#536](https://github.com/DataDog/terraform-provider-datadog/pull/536))
-   `datadog_dashboard`: Set title and title_align properly on heatmap widget ([#539](https://github.com/DataDog/terraform-provider-datadog/pull/539))
-   `datadog_ip_ranges`: Fix data source for IPRanges ([#542](https://github.com/DataDog/terraform-provider-datadog/pull/542))
-   `datadog_monitor`: Fix indent in datadog_monitor docs example ([#543](https://github.com/DataDog/terraform-provider-datadog/pull/543)) Thanks [@nekottyo](https://github.com/nekottyo)

NOTES:

-   `datadog_synthetics_test`: `SyntheticsDeviceID` should accept all allowed values ([#538](https://github.com/DataDog/terraform-provider-datadog/issues/538))
-   Thanks [@razaj92](https://github.com/razaj92) ([#547](https://github.com/DataDog/terraform-provider-datadog/pull/547)) who contributed to this release as well.

## 2.8.0 (June 10, 2020)

FEATURES:

-   `provider`: Add support for `DD_API_KEY`, `DD_APP_KEY` and `DD_HOST` env variables ([#469](https://github.com/DataDog/terraform-provider-datadog/issues/469))
-   `datadog_logs_custom_pipeline`: Add support for lookup processor ([#415](https://github.com/DataDog/terraform-provider-datadog/issues/415))
-   `datadog_integration_aws_lambda_arn`: Add AWS Log Lambda Integration ([#436](https://github.com/DataDog/terraform-provider-datadog/issues/436))
-   `datadog_integration_aws_log_collection`: Add AWS Log collection service resource ([#437](https://github.com/DataDog/terraform-provider-datadog/issues/437)) Thanks [@mhaley-miovision](https://github.com/mhaley-miovision)
-   `datadog_dashboard`: Add support for tags_execution ([#524](https://github.com/DataDog/terraform-provider-datadog/issues/524))
-   `datadog_dashboard`: Add `legend_size` to api request ([#421](https://github.com/DataDog/terraform-provider-datadog/issues/421))
-   `provider`: Add "validate" option that can disable validation ([#474](https://github.com/DataDog/terraform-provider-datadog/issues/474)) Thanks [@bendrucker](https://github.com/bendrucker)

IMPROVEMENTS:

-   `provider`: Harmonized errors across all resources ([#450](https://github.com/DataDog/terraform-provider-datadog/issues/450))
-   `provider`: Add more infos in user agent header ([#455](https://github.com/DataDog/terraform-provider-datadog/issues/455))
-   `provider`: Update the api error message ([#472](https://github.com/DataDog/terraform-provider-datadog/issues/472))
-   `datadog_screenboard`, `datadog_timeboard`: Add deprecation messages ([#496](https://github.com/DataDog/terraform-provider-datadog/issues/496))
-   `provider`: New UserAgent Header ([#455](https://github.com/DataDog/terraform-provider-datadog/issues/455)), ([#510](https://github.com/DataDog/terraform-provider-datadog/issues/510)), ([#511](https://github.com/DataDog/terraform-provider-datadog/issues/511)), and ([#512](https://github.com/DataDog/terraform-provider-datadog/issues/512))
-   `datadog_integration_aws`: Add full AWS Update support ([#521](https://github.com/DataDog/terraform-provider-datadog/issues/521))

BUGFIXES:

-   `datadog_logs_index`: Fail fast if index isn't imported ([#452](https://github.com/DataDog/terraform-provider-datadog/issues/452))
-   `datadog_integration_aws`: Do not set empty structures in request to create aws integration ([#505](https://github.com/DataDog/terraform-provider-datadog/issues/505)) Thanks [@miguelaferreira](https://github.com/miguelaferreira)
-   `datadog_dashboard`: Add default to deprecated `count` field to avoid sending 0 ([#514](https://github.com/DataDog/terraform-provider-datadog/issues/514))
-   `datadog_integration_pagerduty`: Fix perpetual diff in api_token ([#518](https://github.com/DataDog/terraform-provider-datadog/issues/518)) Thanks [@bendrucker](https://github.com/bendrucker)
-   `datadog_dashboard`: Add column revamp properties to dashboard log stream widget ([#517](https://github.com/DataDog/terraform-provider-datadog/issues/517))

NOTES:

-   This release replaces the underlying community driven Datadog API Go client [go-datadog-api](https://github.com/zorkian/go-datadog-api) with the Datadog Official API Go client [datadog-api-client-go](https://github.com/DataDog/datadog-api-client-go) for all resources listed below:
    -   `provider`: Add Datadog Go client API ([#477](https://github.com/DataDog/terraform-provider-datadog/issues/477)) and ([#456](https://github.com/DataDog/terraform-provider-datadog/issues/456))
    -   `datadog_service_level_objective`: Migrate SLO resource with Datadog Go Client ([#490](https://github.com/DataDog/terraform-provider-datadog/issues/490))
    -   `datadog_metric_metadata`: Migrate metric_metadata resource to use Datadog Go client ([#486](https://github.com/DataDog/terraform-provider-datadog/issues/486))
    -   `datadog_integration_aws`: Migrate AWS resource to use Datadog Go client ([#481](https://github.com/DataDog/terraform-provider-datadog/issues/481))
    -   `datadog_integration_gcp`: Migrate GCP resource to use Datadog Go client ([#482](https://github.com/DataDog/terraform-provider-datadog/issues/482))
    -   `datadog_downtime`: Migrate Downtime resource to use Datadog Go client ([#480](https://github.com/DataDog/terraform-provider-datadog/issues/480))
    -   `datadog_ip_ranges`: Migrate IP Range resource with Datadog Go client ([#491](https://github.com/DataDog/terraform-provider-datadog/issues/491))
    -   `datadog_integration_pagerduty_service_object`: Migrate pagerduty_service_object resource to use Datadog Go client ([#488](https://github.com/DataDog/terraform-provider-datadog/issues/488))
    -   `datadog_logs_index`, `datadog_logs_index_order`, `datadog_logs_integration_pipeline`, `datadog_logs_pipeline_order`: Migrate Logs resources to use Datadog Go client ([#483](https://github.com/DataDog/terraform-provider-datadog/issues/483))
    -   `datadog_monitor`: Migrate monitor resource to use Datadog Go client ([#485](https://github.com/DataDog/terraform-provider-datadog/issues/485))
    -   `datadog_dashboard_list`: Migrate Dashboard_list resource to use Datadog Go client ([#479](https://github.com/DataDog/terraform-provider-datadog/issues/479))
    -   `datadog_integration_aws_log_collection`: Migrate aws_log_collection resource to use Datadog Go client ([#501](https://github.com/DataDog/terraform-provider-datadog/issues/501))
    -   `datadog_logs_custom_pipeline`: Migrate Logs custom pipeline resource to utilize Datadog Go client ([#495](https://github.com/DataDog/terraform-provider-datadog/issues/495))
    -   `datadog_synthetics_test`: Migrate synthetics resource to utilize Datadog Go Client ([#499](https://github.com/DataDog/terraform-provider-datadog/issues/499))
    -   `datadog_integration_aws_log_collection`, `datadog_integration_aws_lambda_arn`: Migrate AWS logs to use the Datadog Go Client ([#497](https://github.com/DataDog/terraform-provider-datadog/issues/497))
    -   `datadog_dashboard`: Migrate dashboard resource to use Datadog Go client ([#489](https://github.com/DataDog/terraform-provider-datadog/issues/489))
-   `datadog_screenboard` and `datadog_timeboard` resources are deprecated and should be converted to `datadog_dashboard` resources.
-   Thanks [@NeverTwice](https://github.com/NeverTwice) ([#460](https://github.com/DataDog/terraform-provider-datadog/pull/460)) and [@sepulworld](https://github.com/sepulworld) ([#506](https://github.com/DataDog/terraform-provider-datadog/pull/506)) who contributed to this release as well.

## 2.7.0 (February 10, 2020)

IMPROVEMENTS:

-   `datadog_dashboard`: Add `template_variable_presets` parameter ([#401](https://github.com/DataDog/terraform-provider-datadog/issues/401))
-   `datadog_dashboard`: Add new Monitor Summary widget parameters: `summary_type` and `show_last_triggered` ([#396](https://github.com/DataDog/terraform-provider-datadog/issues/396))
-   `datadog_dashboard`: Hide deprecated Monitor Summary widget parameters: `count` and `start` ([#403](https://github.com/DataDog/terraform-provider-datadog/issues/403))
-   `datadog_monitor`: Improve monitor example with ignoring changes on silenced ([#406](https://github.com/DataDog/terraform-provider-datadog/issues/406))
-   `datadog_service_level_objective`: Fix optional threshold fields handling when updating ([#400](https://github.com/DataDog/terraform-provider-datadog/issues/400))

BUGFIXES:

-   `datadog_downtime`: Gracefully handle recreating downtimes that were canceled manually ([#405](https://github.com/DataDog/terraform-provider-datadog/issues/405))
-   `datadog_screenboard`: Properly set screenboard attributes from client response to not produce non-empty plans ([#404](https://github.com/DataDog/terraform-provider-datadog/issues/404))

NOTES:

-   This is the first release to use the new `terraform-plugin-sdk` ([#346](https://github.com/DataDog/terraform-provider-datadog/issues/346))

## 2.6.0 (January 21, 2020)

FEATURES:

-   `datadog_dashboard`: Add Datadog dashboard SLO widget support ([#355](https://github.com/DataDog/terraform-provider-datadog/issues/355)) Thanks [@mbarrien](https://github.com/mbarrien)

IMPROVEMENTS:

-   `datadog_logs_custom_pipeline`: Support all processors in Logs pipeline ([#357](https://github.com/DataDog/terraform-provider-datadog/pull/357)) Thanks [@tt810](https://github.com/tt810)

BUGFIXES:

-   `datadog_service_level_objective`: Fix slo threshold warning value modified when storing the state ([#352](https://github.com/DataDog/terraform-provider-datadog/pull/352))
-   `datadog_service_level_objective`: `monitor_search` schema removed from the SLO resource as it is not yet supported ([#358](https://github.com/DataDog/terraform-provider-datadog/issues/358)) Thanks [@unclebconnor](https://github.com/unclebconnor)
-   `datadog_monitor`: Resolve non empty diff: "no_data_timeframe = 0 -> 10" on plan diff ([#384](https://github.com/DataDog/terraform-provider-datadog/issues/384)) Thanks [@abicky](https://github.com/abicky)

## 2.5.0 (October 22, 2019)

FEATURES:

-   `datadog_ip_ranges`: New data source for IP ranges ([#298](https://github.com/DataDog/terraform-provider-datadog/issues/298))
-   `datadog_logs_custom_pipeline`: New resource for custom logs pipelines ([#312](https://github.com/DataDog/terraform-provider-datadog/issues/312), [#332](https://github.com/DataDog/terraform-provider-datadog/issues/332))
-   `datadog_logs_index`: New resource for logs indexes ([#326](https://github.com/DataDog/terraform-provider-datadog/issues/326))
-   `datadog_logs_index_order`: New resource for logs index ordering ([#326](https://github.com/DataDog/terraform-provider-datadog/issues/326))
-   `datadog_logs_integration_pipeline`: New resource for integration logs pipelines ([#312](https://github.com/DataDog/terraform-provider-datadog/issues/312), [#332](https://github.com/DataDog/terraform-provider-datadog/issues/332))
-   `datadog_logs_pipeline_order`: New resources for logs pipeline ordering ([#312](https://github.com/DataDog/terraform-provider-datadog/issues/312))

IMPROVEMENTS:

-   `datadog_dashboard`: Added documentation of `event` and `axis` ([#314](https://github.com/DataDog/terraform-provider-datadog/issues/314))
-   `datadog_screenboard`: Added `count` as a valid aggregation method ([#333](https://github.com/DataDog/terraform-provider-datadog/issues/333))

BUGFIXES:

-   `datadog_dashboard`: Fixed parsing of `compute.interval` and `group_by.sort.facet`, mark `group_by.facet` as optional for apm and log queries ([#322](https://github.com/DataDog/terraform-provider-datadog/issues/322), [#325](https://github.com/DataDog/terraform-provider-datadog/issues/325))
-   `datadog_dashboard`: Properly respect `show_legend` ([#329](https://github.com/DataDog/terraform-provider-datadog/issues/329))
-   `datadog_integration_pagerduty`: Add missing exists methods to prevent failing when resource was manually removed outside of Terraform ([#324](https://github.com/DataDog/terraform-provider-datadog/issues/324))
-   `datadog_integration_pagerduty_service_object`: Add missing exists methods to prevent failing when resource was manually removed outside of Terraform ([#324](https://github.com/DataDog/terraform-provider-datadog/issues/324))

## 2.4.0 (September 11, 2019)

FEATURES:

-   `datadog_dashboard_list`: New resource for dashboard lists ([#296](https://github.com/DataDog/terraform-provider-datadog/issues/296))

IMPROVEMENTS:

-   `datadog_dashboard`: Allow specifying `event` and `yaxis` for timeseries definitions ([#282](https://github.com/DataDog/terraform-provider-datadog/issues/282))

## 2.3.0 (August 29, 2019)

IMPROVEMENTS:

-   `datadog-dashboards`: Add resources for log, apm and process query in legacy dashboards ([#272](https://github.com/DataDog/terraform-provider-datadog/issues/272))

BUGFIXES:

-   `datadog_integration_pagerduty`: Make sure PD services don't get removed by updating PD resource ([#304](https://github.com/DataDog/terraform-provider-datadog/issues/304))

## 2.2.0 (August 19, 2019)

FEATURES:

-   `datadog_service_level_objective`: New resource for Service Level Objective (SLO) ([#263](https://github.com/DataDog/terraform-provider-datadog/issues/263))

IMPROVEMENTS:

-   `datadog_dashbaord`: Add support for style block in dashboard widgets. ([#277](https://github.com/DataDog/terraform-provider-datadog/issues/277))
-   `datadog_dashboard`: Add support for metadata block in dashboard widgets ([#278](https://github.com/DataDog/terraform-provider-datadog/issues/278))
-   `datadog_synthetics_test`: Support SSL synthetics tests. ([#279](https://github.com/DataDog/terraform-provider-datadog/issues/279))

BUGFIXES:

-   `datadog_dashboards`: Safely type assert optional fields from log and apm query to avoid a panic if they aren't supplied ([#283](https://github.com/DataDog/terraform-provider-datadog/issues/283))
-   `datadog_synthetics_test`: Fix follow redirects field to properly apply and save in state. ([#256](https://github.com/DataDog/terraform-provider-datadog/issues/256))

## 2.1.0 (July 24, 2019)

FEATURES:

-   `datadog_dashboard`: New Resource combining screenboard and timeboard, allowing a single config to manage all of your Datadog Dashboards. ([#249](https://github.com/DataDog/terraform-provider-datadog/issues/249))
-   `datadog_integration_pagerduty_service_object`: New Resource that allows the configuration of individual pagerduty services for the Datadog Pagerduty Integration. ([#237](https://github.com/DataDog/terraform-provider-datadog/issues/237))

IMPROVEMENTS:

-   `datadog_aws`: Add a mutex around all API operations for this resource. ([#254](https://github.com/DataDog/terraform-provider-datadog/issues/254))
-   `datadog_downtime`: General improvements around allowing the resource to be ran multiple times without sending any unchanged values for the start/end times. Also fixes non empty diff when monitor_tags isn't set. ([#264](https://github.com/DataDog/terraform-provider-datadog/issues/264)] [[#267](https://github.com/DataDog/terraform-provider-datadog/issues/267))
-   `datadog_monitor`: Only add a threshold window if a recovery or trigger window is set. [[#260](https://github.com/DataDog/terraform-provider-datadog/issues/260)] Thanks [@heldersepu](https://github.com/heldersepu)
-   `datadog_user`: Make `is_admin` computed to continue its deprecation path and avoid spurious diffs. ([#251](https://github.com/DataDog/terraform-provider-datadog/issues/251))

NOTES:

-   This release includes Terraform SDK upgrade to 0.12.5. ([#265](https://github.com/DataDog/terraform-provider-datadog/issues/265))

## 2.0.2 (June 26, 2019)

BUGFIXES:

-   `datadog_monitor`: DiffSuppress the difference between `metric alert` and `query alert` no matter what is in the current state and prevent the force recreation of monitors due to this change. ([#247](https://github.com/DataDog/terraform-provider-datadog/issues/247))

## 2.0.1 (June 21, 2019)

BUGFIXES:

-   `datadog_monitor`: Don't force the destruction and recreation of a monitor when the type changes between `metric alert` and `query alert`. ([#242](https://github.com/DataDog/terraform-provider-datadog/issues/242))

## 2.0.0 (June 18, 2019)

NOTES:

-   `datadog_monitor`: The silence attribute is beginning its deprecation process, please use `datadog_downtime` instead ([#221](https://github.com/DataDog/terraform-provider-datadog/issues/221))

IMPROVEMENTS:

-   `datadog_monitor`: Use ForceNew when changing the Monitor type ([#236](https://github.com/DataDog/terraform-provider-datadog/issues/236))
-   `datadog_monitor`: Add default to `no data` timeframe of 10 minutes. ([#212](https://github.com/DataDog/terraform-provider-datadog/issues/212))
-   `datadog_synthetics_test`: Support synthetics monitors in composite monitors. ([#222](https://github.com/DataDog/terraform-provider-datadog/issues/222))
-   `datadog_downtime`: Add validation to tags, add timezone parameter, improve downtime id handling, add descriptions to fields. ([#204](https://github.com/DataDog/terraform-provider-datadog/issues/204))
-   `datadog_screenboard`: Add support for metadata alias in graphs. ([#215](https://github.com/DataDog/terraform-provider-datadog/issues/215))
-   `datadog_screenboard`: Add `custom_bg_color` to graph config. [[#189](https://github.com/DataDog/terraform-provider-datadog/issues/189)] Thanks [@milanvdm](https://github.com/milanvdm)
-   Update the vendored go client to `v2.21.0`. ([#230](https://github.com/DataDog/terraform-provider-datadog/issues/230))

BUGFIXES:

-   `datadog_timeboard`: Fix the `extra_col` from having a non empty plan when there are no changes. ([#231](https://github.com/DataDog/terraform-provider-datadog/issues/231))
-   `datadog_timeboard`: Fix the `precision` from having a non empty plan when there are no changes. ([#228](https://github.com/DataDog/terraform-provider-datadog/issues/228))
-   `datadog_monitor`: Fix the sorting of monitor tags that could lead to a non empty diff. ([#214](https://github.com/DataDog/terraform-provider-datadog/issues/214))
-   `datadog_monitor`: Properly save `query_config` as to avoid to an improper non empty diff. ([#209](https://github.com/DataDog/terraform-provider-datadog/issues/209))
-   `datadog_monitor`: Fix and clarify documentation on unmuting monitor scopes. ([#202](https://github.com/DataDog/terraform-provider-datadog/issues/202))
-   `datadog_screenboard`: Change monitor schema to be of type String instead of Int. [[#154](https://github.com/DataDog/terraform-provider-datadog/issues/154)] Thanks [@mnaboka](https://github.com/mnaboka)

## 1.9.0 (May 09, 2019)

IMPROVEMENTS:

-   `datadog_downtime`: Add `monitor_tags` getting and setting ([#167](https://github.com/DataDog/terraform-provider-datadog/issues/167))
-   `datadog_monitor`: Add support for `enable_logs` in log monitors ([#151](https://github.com/DataDog/terraform-provider-datadog/issues/151))
-   `datadog_monitor`: Add suport for `threshold_windows` attribute ([#131](https://github.com/DataDog/terraform-provider-datadog/issues/131))
-   Support importing dashboards using the new string ID ([#184](https://github.com/DataDog/terraform-provider-datadog/issues/184))
-   Various documentation fixes and improvements ([#152](https://github.com/DataDog/terraform-provider-datadog/issues/152), [#171](https://github.com/DataDog/terraform-provider-datadog/issues/171), [#176](https://github.com/DataDog/terraform-provider-datadog/issues/176), [#178](https://github.com/DataDog/terraform-provider-datadog/issues/178), [#180](https://github.com/DataDog/terraform-provider-datadog/issues/180), [#183](https://github.com/DataDog/terraform-provider-datadog/issues/183))

NOTES:

-   This release includes Terraform SDK upgrade to 0.12.0-rc1. The provider is backwards compatible with Terraform v0.11.X, there should be no significant changes in behavior. Please report any issues to either [Terraform issue tracker](https://github.com/hashicorp/terraform/issues) or to [Terraform Datadog Provider issue tracker](https://github.com/DataDog/terraform-provider-datadog/issues) ([#194](https://github.com/DataDog/terraform-provider-datadog/issues/194), [#198](https://github.com/DataDog/terraform-provider-datadog/issues/198))

## 1.8.0 (April 15, 2019)

INTERNAL:

-   provider: Enable request/response logging in `>=DEBUG` mode ([#153](https://github.com/DataDog/terraform-provider-datadog/issues/153))

IMPROVEMENTS:

-   Add Synthetics API and Browser tests support + update go-datadog-api to latest. ([169](https://github.com/DataDog/terraform-provider-datadog/pull/169))

## 1.7.0 (March 05, 2019)

BUGFIXES:

-   Bump go api client to 2.19.0 to fix TileDefStyle.fillMax type errors. ([143](https://github.com/DataDog/terraform-provider-datadog/pull/143))([144](https://github.com/DataDog/terraform-provider-datadog/pull/144))
-   Fix the usage of `start_date` and `end_data` only being read on the first apply. ([145](https://github.com/DataDog/terraform-provider-datadog/pull/145))

IMPROVEMENTS:

-   Upgrade to Go 1.11. ([141](https://github.com/DataDog/terraform-provider-datadog/pull/141/files))
-   Add AWS Integration resource to the docs. ([146](https://github.com/DataDog/terraform-provider-datadog/pull/146))

FEATURES:

-   **New Resource:** `datadog_integration_pagerduty` ([135](https://github.com/DataDog/terraform-provider-datadog/pull/135))

## 1.6.0 (November 30, 2018)

BUGFIXES:

-   the graph.style.palette_flip field is a boolean but only works if it's passed as a string. ([#29](https://github.com/DataDog/terraform-provider-datadog/issues/29))
-   datadog_monitor - Removal of 'silenced' resource argument has no practical effect. ([#41](https://github.com/DataDog/terraform-provider-datadog/issues/41))
-   datadog_screenboard - widget swapping `x` and `y` parameters. ([#119](https://github.com/DataDog/terraform-provider-datadog/issues/119))
-   datadog_screenboard - panic: interface conversion: interface {} is string, not float64. ([#117](https://github.com/DataDog/terraform-provider-datadog/issues/117))

IMPROVEMENTS:

-   Feature Request: AWS Integration. ([#76](https://github.com/DataDog/terraform-provider-datadog/issues/76))
-   Bump datadog api to v2.18.0 and add support for include units and zero. ([#121](https://github.com/DataDog/terraform-provider-datadog/pull/121))

## 1.5.0 (November 06, 2018)

IMPROVEMENTS:

-   Add Google Cloud Platform integration ([#108](https://github.com/DataDog/terraform-provider-datadog/pull/108))
-   Add new hostmap widget options: `node type`, `fill_min` and `fill_max`. ([#106](https://github.com/DataDog/terraform-provider-datadog/pull/106))
-   Use dates to set downtime interval, improve docs. ([#113](https://github.com/DataDog/terraform-provider-datadog/pull/113))
-   Bump Terraform provider SDK to latest. ([#110](https://github.com/DataDog/terraform-provider-datadog/pull/110))
-   Better document `evaluation_delay` option. ([#112](https://github.com/DataDog/terraform-provider-datadog/pull/112))

## 1.4.0 (October 02, 2018)

IMPROVEMENTS:

-   Pull changes from go-datadog-api v2.14.0 ([#99](https://github.com/DataDog/terraform-provider-datadog/pull/99))
-   Add `api_url` argument to the provider ([#101](https://github.com/DataDog/terraform-provider-datadog/pull/101))

BUGFIXES:

-   Allow `new_host_delay` to be unset ([#100](https://github.com/DataDog/terraform-provider-datadog/issues/100))

## 1.3.0 (September 25, 2018)

IMPROVEMENTS:

-   Add full support for Datadog screenboards ([#91](https://github.com/DataDog/terraform-provider-datadog/pull/91))

BUGFIXES:

-   Do not compute `new_host_delay` ([#88](https://github.com/DataDog/terraform-provider-datadog/pull/88))
-   Remove buggy uptime widget ([#93](https://github.com/DataDog/terraform-provider-datadog/pull/93))

## 1.2.0 (August 27, 2018)

BUG FIXES:

-   Update "monitor type" options in docs ([#81](https://github.com/DataDog/terraform-provider-datadog/pull/81))
-   Fix typo in timeboard documentation ([#83](https://github.com/DataDog/terraform-provider-datadog/pull/83))

IMPROVEMENTS:

-   Update `go-datadog-api` to v.2.11.0 and move vendoring from `gopkg.in/zorkian/go-datadog-api.v2` to `github.com/zorkian/go-datadog-api` ([#84](https://github.com/DataDog/terraform-provider-datadog/pull/84))
-   Deprecate `is_admin` as part of the work needed to add support for `access_role` ([#85](https://github.com/DataDog/terraform-provider-datadog/pull/85))

## 1.1.0 (July 30, 2018)

IMPROVEMENTS:

-   Added more docs detailing expected weird behaviours from the Datadog API. ([#79](https://github.com/DataDog/terraform-provider-datadog/pull/79))
-   Added support for 'unknown' monitor threshold field. ([#45](https://github.com/DataDog/terraform-provider-datadog/pull/45))
-   Deprecated the `role` argument for `User` resources since it's now a noop on the Datadog API. ([#80](https://github.com/DataDog/terraform-provider-datadog/pull/80))

## 1.0.4 (July 06, 2018)

BUG FIXES:

-   Bump `go-datadog-api.v2` to v2.10.0 thus fixing tag removal on monitor updates ([#43](https://github.com/DataDog/terraform-provider-datadog/issues/43))

## 1.0.3 (January 03, 2018)

IMPROVEMENTS:

-   `datadog_downtime`: adding support for setting `monitor_id` ([#18](https://github.com/DataDog/terraform-provider-datadog/issues/18))

## 1.0.2 (December 19, 2017)

IMPROVEMENTS:

-   `datadog_monitor`: Add support for monitor recovery thresholds ([#37](https://github.com/DataDog/terraform-provider-datadog/issues/37))

BUG FIXES:

-   Fix issue with DataDog service converting metric alerts to query alerts ([#16](https://github.com/DataDog/terraform-provider-datadog/issues/16))

## 1.0.1 (December 06, 2017)

BUG FIXES:

-   Fix issue reading resources that have been updated outside of Terraform ([#34](https://github.com/DataDog/terraform-provider-datadog/issues/34))

## 1.0.0 (October 20, 2017)

BUG FIXES:

-   Improved detection of "drift" when graphs are reconfigured outside of Terraform. ([#27](https://github.com/DataDog/terraform-provider-datadog/issues/27))
-   Fixed API response decoding error on graphs. ([#27](https://github.com/DataDog/terraform-provider-datadog/issues/27))

## 0.1.1 (September 26, 2017)

FEATURES:

-   **New Resource:** `datadog_metric_metadata` ([#17](https://github.com/DataDog/terraform-provider-datadog/issues/17))

## 0.1.0 (June 20, 2017)

NOTES:

-   Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
