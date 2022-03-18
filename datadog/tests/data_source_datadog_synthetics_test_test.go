package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccDatatogSyntheticsTest(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToUpper(strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_"))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsResourceIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAcccheckDatatogSyntheticsTestConfig(uniq),
				Check:  checkDatatogSyntheticsTest(accProvider, uniq),
			},
		},
	})
}

func checkDatatogSyntheticsTest(accProvider func() (*schema.Provider, error), uniq string) resource.TestCheckFunc {
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

func testAccDatatogSyntheticsTestConfig(uniq string) string {
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

func testAcccheckDatatogSyntheticsTestConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_synthetics_test" "data_source_test" {
  depends_on = [
    datadog_synthetics_test.resource_test,
  ]
  test_id = datadog_synthetics_test.resource_test.id
}`, testAccDatatogSyntheticsTestConfig(uniq))
}
