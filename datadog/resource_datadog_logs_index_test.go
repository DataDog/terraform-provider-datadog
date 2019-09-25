package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
	"testing"
)

const indexConfigForUpdate = `
resource "datadog_logs_index" "my_index_test" {
	name = "main"
	filter {
		query = "new query"
    }
	exclusion_filter {
		is_enabled = true
		filter {
			query = "source:agent"
			sample_rate = 0.2
		}
	}
	exclusion_filter {
		name = "absent is_enabled"
		filter {
			query = "source:kafka"
			sample_rate = 0.0
		}
	}
	exclusion_filter {
		name = "absent query"
		is_enabled = false
		filter {
			sample_rate = 0.2
		}
	}
	exclusion_filter {
		name = "absent sample_rate"
		is_enabled = false
		filter {
			query = "source:kafka"
		}
	}
}
`

func TestAccDatadogLogsIndex_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIndexDestroyDoNothing,
		Steps: []resource.TestStep{
			{
				Config: indexConfigForUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_logs_index.my_index_test", "name", "main"),
					resource.TestCheckResourceAttr(
						"datadog_logs_index.my_index_test", "exclusion_filter.0.name", ""),
					resource.TestCheckResourceAttr(
						"datadog_logs_index.my_index_test", "exclusion_filter.1.is_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_logs_index.my_index_test", "exclusion_filter.2.filter.0.query", ""),
					resource.TestCheckResourceAttr(
						"datadog_logs_index.my_index_test", "exclusion_filter.3.filter.0.sample_rate", "0"),
				),
			},
		},
	})
}

func testAccCheckIndexDestroyDoNothing(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_index" {
			id := r.Primary.ID
			p, err := client.GetLogsIndex(id)

			if err != nil {
				return fmt.Errorf("received an error when retrieving index, (%s)", err)
			}
			if p == nil {
				return fmt.Errorf("index should not be deleted on API side")
			}
		}

	}
	return nil
}
