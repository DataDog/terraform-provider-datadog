package datadog

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

const indexForImportConfig = `
resource "datadog_logs_index" "import_index_test" {
	name = "main"
	filter {
		query = "*"
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

const indexOrderForImportConfig = `
resource "datadog_logs_index_order" "import_index_list" {
	name = "main list"
	indexes = [
		"main"
	]
}
`

func TestAccLogsIndex_importBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIndexDestroyDoNothing,
		Steps: []resource.TestStep{
			{
				Config: indexForImportConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_index.import_index_test", "name", "main"),
				),
			},
			{
				ResourceName:      "datadog_logs_index.import_index_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLogsIndexOrder_importBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: indexOrderForImportConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_index_order.import_index_list", "name", "main list"),
				),
			},
			{
				ResourceName: "datadog_logs_index_order.import_index_list",
				ImportState:  true,
			},
		},
	})
}
