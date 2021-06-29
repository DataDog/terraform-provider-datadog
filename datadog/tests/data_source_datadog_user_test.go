package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDatadogUserDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: CreateTestAccDatasourceUserConfig("test@example.com"),
				Check:  resource.TestCheckResourceAttr("datadog_user.foo", "email", "test@example.com"),
			},
			{
				Config: testAccDatasourceUserConfig("test@example.com"),
				Check:  resource.TestCheckResourceAttrSet("data.datadog_user.test", "id"),
			},
		},
	})
}

func TestAccDatadogUserDatasourceError(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceUserConfig("doesntexists@example.com"),
				ExpectError: regexp.MustCompile("didn't found any user mathing this email"),
			},
		},
	})
}

func testAccDatasourceUserConfig(uniq string) string {
	return fmt.Sprintf(`
	data "datadog_user" "test" {
	  filter = "%s"
	}`, uniq)
}

func CreateTestAccDatasourceUserConfig(uniq string) string {
	return fmt.Sprintf(`
  resource "datadog_user" "foo" {
    email = "%s"
  }`, uniq)

}
