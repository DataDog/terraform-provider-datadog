package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestDatadogUser_import(t *testing.T) {
	resourceName := "datadog_user.foo"
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	username := strings.ToLower(uniqueEntityName(clock, t)) + "@example.com"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigImported(username),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"access_role"},
			},
		},
	})
}

func testAccCheckDatadogUserConfigImported(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_user" "foo" {
  email  = "%s"
  handle = "%s"
  name   = "Test User"
}`, uniq, uniq)
}
