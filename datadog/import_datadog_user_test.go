package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestDatadogUser_import(t *testing.T) {
	resourceName := "datadog_user.foo"
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogUserDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigImported,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccCheckDatadogUserConfigImported = `
resource "datadog_user" "foo" {
  email  = "test@example.com"
  handle = "test@example.com"
  name   = "Test User"
}
`
