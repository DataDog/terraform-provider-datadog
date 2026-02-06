package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccMonitor_Fwprovider_Create(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitor(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "type", "metric alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "message", "Monitor triggered. Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Update(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitor(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "type", "metric alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "message", "Monitor triggered. Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"),
				),
			},
			{
				Config: testAccCheckDatadogMonitor_update(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "type", "metric alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "message", "updated message"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.r", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Basic(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_group_delay", "500"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "draft_status", "published"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Assets(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorAssetsConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "assets.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_monitor.foo", "assets.*", map[string]string{
						"name":          "Datadog Runbook",
						"url":           "/notebook/1234",
						"category":      "runbook",
						"resource_key":  "1234",
						"resource_type": "notebook",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_monitor.foo", "assets.*", map[string]string{
						"name":     "Confluence Runbook",
						"url":      "https://datadoghq.atlassian.net/wiki/spaces/ENG/pages/12345/Runbook",
						"category": "runbook",
					}),
				),
			},
			{
				Config: testAccCheckDatadogMonitorAssetsConfigUpdated(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "assets.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_monitor.foo", "assets.*", map[string]string{
						"name":          "Datadog Runbook 2",
						"url":           "/notebook/5678",
						"category":      "runbook",
						"resource_key":  "5678",
						"resource_type": "notebook",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_monitor.foo", "assets.*", map[string]string{
						"name":     "Confluence Runbook",
						"url":      "https://datadoghq.atlassian.net/wiki/spaces/ENG/pages/12345/Runbook",
						"category": "runbook",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("datadog_monitor.foo", "assets.*", map[string]string{
						"name":     "Google Doc Runbook",
						"url":      "https://docs.google.com/runbook",
						"category": "runbook",
					}),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_ServiceCheck_Basic(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorServiceCheckConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "service check"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", `"custom.check".by("environment:foo").last(2).count_by_status()`),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "600"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.unknown", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "3"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_BasicNoTresholdOrPriority(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholdsOrPriority(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "bar:baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "priority"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Updated(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	monitorNameUpdated := monitorName + "-updated"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_group_delay", "500"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "300"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "draft_status", "published"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "a different message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated! @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "900"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "800"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "no_data_timeframe", "20"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notification_preset_name", "show_all"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "40"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "0.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "3.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "2.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "10"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz:qux"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "quux"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "draft_status", "draft"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "metric alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_UpdatedToRemoveTags(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	monitorNameUpdated := monitorName + "-updated"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notification_preset_name", "hide_query"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_group_delay", "500"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "300"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:bar"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdatedWithPriorityRemoved(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "a different message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated! @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notification_preset_name", "show_all"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "900"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "800"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "no_data_timeframe", "20"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "40"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "3.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "2.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "10"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "priority"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "tags"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "0"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "metric alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_TrimWhiteSpace(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigWhitespace(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel\n"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2\n"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Basic_float_int(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigInts(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigIntsMixed(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "2"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_Log(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigLogAlert(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "log alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").by(\"source\").last(\"5m\") > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "on_missing_data"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "group_retention_duration"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "enable_logs_sample", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "groupby_simple_monitor", "true"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_LogMultiAlert(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigLogMultiAlert(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "log alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").by(\"source,status\").last(\"5m\") > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "on_missing_data", "show_no_data"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "group_retention_duration", "2d"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "enable_logs_sample", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "groupby_simple_monitor", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_by.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_by.0", "status"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_NoThresholdWindows(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholdWindows(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "message", "test"),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_ThresholdWindows(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigThresholdWindows(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query",
						"avg(last_1h):anomalies(avg:system.cpu.system{name:cassandra}, 'basic', 3, direction='above', alert_window='last_5m', interval=20, count_default_zero='true') >= 1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.25"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_threshold_windows.0.recovery_window", "last_5m"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_threshold_windows.0.trigger_window", "last_5m"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_ComposeWithSyntheticsTest(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorComposeWithSyntheticsTest(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "name", monitorName),
					resource.TestCheckResourceAttrSet(
						"datadog_monitor.bar", "query"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_FormulaFunction(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorFormulaFunction(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "variables.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "variables.0.event_query.#", "2"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_FormulaFunction_Cost(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCostMonitorFormulaFunction(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "variables.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "variables.0.cloud_cost_query.#", "1"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_ZeroDelay(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigZeroDelay(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "0"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_RestrictedRoles(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigRestrictedRoles(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "restricted_roles.#", "1"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigRestrictedRolesRemoved(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "restricted_roles.#", "0"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_SchedulingOptions(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorWithSchedulingOptions(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.evaluation_window.0.day_starts", "04:00"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.evaluation_window.0.month_starts", "1"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_EmptySchedulingOptions(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorWithEmptySchedulingOptions(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_SchedulingOptionsHourStart(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorWithSchedulingOptionsHourStart(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.evaluation_window.0.hour_starts", "0"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_SchedulingOptionsCustomSchedule(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorWithSchedulingOptionsCustomSchedule(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.timezone", "America/New_York"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.start", "2023-11-10T12:31:00"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_SchedulingOptionsCustomScheduleNoStart(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorWithSchedulingOptionsCustomSchedule(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.timezone", "America/New_York"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_DefaultTags(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckDatadogMonitorDestroyFwprovider(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{ // New tags are correctly added and duplicates are kept
				Config: testAccCheckDatadogMonitorConfigDuplicateTags(monitorName),
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "4"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:thebar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "default_key:default_value"),
				),
			},
			{ // Resource tags take precedence over default tags and duplicates stay
				Config: testAccCheckDatadogMonitorConfigDuplicateTags(monitorName),
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"foo": "not_bar",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:thebar"),
				),
			},
			{ // Resource tags take precedence over default tags, but new tags are added
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"foo":     "not_bar",
						"new_tag": "new_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "new_tag:new_value"),
				),
			},
			{ // Tags without any value work correctly
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"no_value": "",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "no_value"),
				),
			},
			{ // Tags with colons in the value work correctly
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"repo_url": "https://github.com/repo/path",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "effective_tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "repo_url:https://github.com/repo/path"),
				),
			},
			{ // Works with monitors without a tag attribute
				Config: testAccCheckDatadogMonitorConfigNoTag(monitorName),
				ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "tags"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "0"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "effective_tags.*", "default_key:default_value"),
				),
			},
		},
	})
}

func TestAccMonitor_Fwprovider_WithRestrictionPolicy(t *testing.T) {
	t.Setenv("TERRAFORM_MONITOR_FRAMEWORK_PROVIDER", "true")
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorWithRestrictionRoleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorWithRestrictionPolicy(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.bar", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_restriction_policy.baz", "bindings.0.principals.#", "2"),
					verifyRestrictedRolesSize(providers.frameworkProvider, 1),
				),
			},
			{ // A monitor resource update should not clear restricted_roles when the field is not defined
				Config: testAccCheckDatadogMonitorWithRestrictionPolicyUpdated(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.bar", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_restriction_policy.baz", "bindings.0.principals.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.bar", "restricted_roles.#", "1"),
					verifyRestrictedRolesSize(providers.frameworkProvider, 1),
				),
			},
			{ // Removing restriction policy resource should wipe out restricted_roles
				Config: testAccCheckDatadogMonitorWithRestrictionPolicyDestroyed(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExistsFwprovider(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.bar", "name", monitorName),
					verifyRestrictedRolesSize(providers.frameworkProvider, 0),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitor(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_monitor" "r" {
	name               = "%s"
    type               = "metric alert"
    message            = "Monitor triggered. Notify: @hipchat-channel"
    query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
}`, uniq)
}

func testAccCheckDatadogMonitor_update(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_monitor" "r" {
	name               = "%s"
    type               = "metric alert"
    message            = "updated message"
    query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
}`, uniq)
}

func testAccCheckDatadogMonitorAssetsConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name    = "%s"
  type    = "query alert"
  message = "some message Notify: @hipchat-channel"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
	warning  = "1.0"
	critical = "2.0"
  }

  assets {
	name               = "Datadog Runbook"
	url                = "/notebook/1234"
	category           = "runbook"
	resource_key       = "1234"
	resource_type      = "notebook"
  }

  assets {
	name               = "Confluence Runbook"
	url                = "https://datadoghq.atlassian.net/wiki/spaces/ENG/pages/12345/Runbook"
	category           = "runbook"
  }
}`, uniq)
}

func testAccCheckDatadogMonitorDestroyFwprovider(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := MonitorDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func MonitorDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_monitor" {
				continue
			}
			id, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			_, httpResp, err := apiInstances.GetMonitorsApiV1().GetMonitor(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogMonitorExistsFwprovider(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := monitorExistsHelperFwprovider(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func monitorExistsHelperFwprovider(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_monitor" {
			continue
		}
		id, _ := strconv.ParseInt(r.Primary.ID, 10, 64)

		_, httpResp, err := apiInstances.GetMonitorsApiV1().GetMonitor(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Monitor")
		}
	}
	return nil
}
