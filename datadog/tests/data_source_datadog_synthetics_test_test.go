package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogSyntheticsTest(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := providers.sdkV2Provider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
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
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := providers.sdkV2Provider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
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
		tick_every = 900
	}
	assertion {
		type     = "statusCode"
		operator = "is"
		target   = "200"
	}
	tags = ["env:prod", "foo"]
}`, uniq)
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
