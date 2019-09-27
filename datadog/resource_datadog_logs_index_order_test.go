package datadog

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

const indexorderConfig = `
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
resource "datadog_logs_index_order" "my_index_list" {
	name = "main list"
	depends_on = [
		"datadog_logs_index.my_index_test"
	]
	indexes = [
		"${datadog_logs_index.my_index_test.id}"
	]
}
`

func TestAccDatadogLogsIndexOrder_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIndexDestroyDoNothing,
		Steps: []resource.TestStep{
			{
				Config: indexorderConfig,
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
					resource.TestCheckResourceAttr(
						"datadog_logs_index_order.my_index_list", "name", "main list"),
					resource.TestCheckResourceAttr(
						"datadog_logs_index_order.my_index_list", "indexes.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_index_order.my_index_list", "indexes.0", "main"),
				),
			},
		},
	})
}
