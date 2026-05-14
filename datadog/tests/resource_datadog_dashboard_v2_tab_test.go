package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Tab — basic create with two tabs referencing widgets by @N position
func TestAccDatadogDashboardV2Tab(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardTabConfig, "datadog_dashboard.tab_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardTab", config, name, datadogDashboardTabAsserts)
}

func TestAccDatadogDashboardV2Tab_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardTabConfig, "datadog_dashboard.tab_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardTab_import", config, name)
}

// TabUpdate — multi-step: rename tab, move widgets between tabs, remove all tabs.
// Uses its own cassette (cannot reuse v1 cassette because v2 includes widget IDs
// in PUT bodies, causing request body mismatch).
func TestAccDatadogDashboardV2TabUpdate(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)

	v2Name := "datadog_dashboard_v2.tab_dashboard"

	configCreate := replacer.Replace(`
resource "datadog_dashboard_v2" "tab_dashboard" {
	title       = "{{uniq}}"
	description = "Created using the Datadog provider in Terraform"
	layout_type = "ordered"
	widget {
		note_definition { content = "Widget 1" }
	}
	widget {
		note_definition { content = "Widget 2" }
	}
	widget {
		note_definition { content = "Widget 3" }
	}
	tab {
		name       = "Overview"
		widget_ids = ["@1", "@2"]
	}
	tab {
		name       = "Details"
		widget_ids = ["@3"]
	}
}`)

	configUpdate := replacer.Replace(`
resource "datadog_dashboard_v2" "tab_dashboard" {
	title       = "{{uniq}}"
	description = "Created using the Datadog provider in Terraform"
	layout_type = "ordered"
	widget {
		note_definition { content = "Widget 1" }
	}
	widget {
		note_definition { content = "Widget 2" }
	}
	widget {
		note_definition { content = "Widget 3" }
	}
	tab {
		name       = "Overview (renamed)"
		widget_ids = ["@1"]
	}
	tab {
		name       = "Details"
		widget_ids = ["@2", "@3"]
	}
}`)

	configRemove := replacer.Replace(`
resource "datadog_dashboard_v2" "tab_dashboard" {
	title       = "{{uniq}}"
	description = "Created using the Datadog provider in Terraform"
	layout_type = "ordered"
	widget {
		note_definition { content = "Widget 1" }
	}
	widget {
		note_definition { content = "Widget 2" }
	}
	widget {
		note_definition { content = "Widget 3" }
	}
}`)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: configCreate,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(v2Name, "tab.#", "2"),
					resource.TestCheckResourceAttr(v2Name, "tab.0.name", "Overview"),
					resource.TestCheckResourceAttr(v2Name, "tab.1.name", "Details"),
				),
			},
			{
				Config: configUpdate,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(v2Name, "tab.#", "2"),
					resource.TestCheckResourceAttr(v2Name, "tab.0.name", "Overview (renamed)"),
					resource.TestCheckResourceAttr(v2Name, "tab.0.widget_ids.#", "1"),
					resource.TestCheckResourceAttr(v2Name, "tab.1.widget_ids.#", "2"),
				),
			},
			{
				Config: configRemove,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr(v2Name, "tab.#", "0"),
				),
			},
		},
	})
}
