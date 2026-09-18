package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSyntheticsTest(t *testing.T) {
	cleanupSyntheticsTests(t)
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := providers.sdkV2Provider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSyntheticsTestConfig(uniq),
				Check:  checkDatadogSyntheticsTest(uniq),
			},
		},
	})
}

func TestAccDatadogSyntheticsTestWithUrl(t *testing.T) {
	cleanupSyntheticsTests(t)
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := providers.sdkV2Provider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsTestIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSyntheticsTestConfigWithUrl(uniq),
				Check:  checkDatadogSyntheticsTest(uniq),
			},
		},
	})
}

func checkDatadogSyntheticsTest(uniq string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "name", uniq),
		resource.TestCheckResourceAttrSet(
			"data.datadog_synthetics_test.data_source_test", "id"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "tags.0", "env:prod"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "tags.1", "foo"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "url", "https://www.example.com"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "type", "api"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "subtype", "http"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "status", "live"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "message", ""),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "locations.#", "1"),
		resource.TestCheckTypeSetElemAttr(
			"data.datadog_synthetics_test.data_source_test", "locations.*", "aws:ap-northeast-1"),
		resource.TestCheckResourceAttrSet(
			"data.datadog_synthetics_test.data_source_test", "monitor_id"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.tick_every", "900"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.min_failure_duration", "120"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.min_location_failed", "1"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.monitor_name", uniq+"-monitor"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.monitor_priority", "3"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.http_version", "any"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.retry.0.count", "2"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.retry.0.interval", "300"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.monitor_options.0.renotify_interval", "60"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.monitor_options.0.renotify_occurrences", "3"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.monitor_options.0.escalation_message", "escalation message"),
		resource.TestCheckResourceAttr(
			"data.datadog_synthetics_test.data_source_test", "options_list.0.monitor_options.0.notification_preset_name", "show_all"),
	)
}

func testAccDatadogSyntheticsTestConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "resource_test" {
	name = "%s"
	status = "live"
	locations = ["aws:ap-northeast-1"]
	type = "api"
	request_definition {
		method = "GET"
		url    = "https://www.example.com"
	}
	options_list {
		tick_every           = 900
		min_failure_duration = 120
		min_location_failed  = 1
		monitor_name         = "%s-monitor"
		monitor_priority     = 3
		retry {
			count    = 2
			interval = 300
		}
		monitor_options {
			renotify_interval        = 60
			renotify_occurrences     = 3
			escalation_message       = "escalation message"
			notification_preset_name = "show_all"
		}
	}
	assertion {
		type     = "statusCode"
		operator = "is"
		target   = "200"
	}
	tags = ["env:prod", "foo"]
}`, uniq, uniq)
}

func testAccCheckDatadogSyntheticsTestConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_synthetics_test" "data_source_test" {
  depends_on = [
    datadog_synthetics_test.resource_test,
  ]
  test_id = datadog_synthetics_test.resource_test.id
}`, testAccDatadogSyntheticsTestConfig(uniq))
}

func testAccCheckDatadogSyntheticsTestConfigWithUrl(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_synthetics_test" "data_source_test" {
  depends_on = [
    datadog_synthetics_test.resource_test,
  ]
  test_id = "https://app.datadoghq.com/synthetics/details/${datadog_synthetics_test.resource_test.id}"
}`, testAccDatadogSyntheticsTestConfig(uniq))
}
