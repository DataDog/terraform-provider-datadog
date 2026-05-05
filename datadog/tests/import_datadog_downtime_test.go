package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDatadogDowntime_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_downtime.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	start := clockFromContext(ctx).Now().Local().Add(time.Hour * time.Duration(1))
	end := start.Add(time.Hour * time.Duration(1))
	startTimestamp := start.Unix()
	endTimestamp := end.Unix()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigImported(startTimestamp, endTimestamp),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogDowntimeConfigImported(start, end int64) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["host:X", "host:Y"]
  start = %d
  end   = %d
  timezone = "UTC"

  message = "Example Datadog downtime message."
  # NOTE: we now ignore monitor_tags on newly created monitors if the attribute
  # value is equal to ["*"], since that's the default and working with TypeList
  # defaults it problematic - see the comment for monitor_tags
  # in resource_datadog_downtime.go
  monitor_tags = ["foo:bar"]
}`, start, end)
}
