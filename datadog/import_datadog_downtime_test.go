package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestDatadogDowntime_import(t *testing.T) {
	resourceName := "datadog_downtime.foo"
	accProviders, _, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigImported,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccCheckDatadogDowntimeConfigImported = `
resource "datadog_downtime" "foo" {
  scope = ["host:X", "host:Y"]
  start = 1735707600
  end   = 1735765200
  timezone = "UTC"

  message = "Example Datadog downtime message."
  # NOTE: we now ignore monitor_tags on newly created monitors if the attribute
  # value is equal to ["*"], since that's the default and working with TypeList
  # defaults it problematic - see the comment for monitor_tags
  # in resource_datadog_downtime.go
  monitor_tags = ["foo:bar"]
}
`
