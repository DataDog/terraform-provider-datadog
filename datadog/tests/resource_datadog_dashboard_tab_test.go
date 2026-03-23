package test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const datadogDashboardTabConfig = `
resource "datadog_dashboard" "tab_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"

	widget {
		note_definition {
			content = "Widget 1"
		}
	}

	widget {
		note_definition {
			content = "Widget 2"
		}
	}

	widget {
		note_definition {
			content = "Widget 3"
		}
	}

	tab {
		name       = "Overview"
		widget_ids = ["@1", "@2"]
	}

	tab {
		name       = "Details"
		widget_ids = ["@3"]
	}
}
`

var datadogDashboardTabAsserts = []string{
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.# = 3",
	"tab.# = 2",
	"tab.0.name = Overview",
	"tab.0.widget_ids.# = 2",
	"tab.0.widget_ids.0 = @1",
	"tab.0.widget_ids.1 = @2",
	"tab.1.name = Details",
	"tab.1.widget_ids.# = 1",
	"tab.1.widget_ids.0 = @3",
}

func TestAccDatadogDashboardTab(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTabConfig, "datadog_dashboard.tab_dashboard", datadogDashboardTabAsserts)
}

func TestAccDatadogDashboardTab_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTabConfig, "datadog_dashboard.tab_dashboard")
}

// TestAccDatadogDashboardTabSingleTab verifies a dashboard with one tab containing
// all widgets. This exercises a different topology than the multi-tab tests and
// ensures the @N reverse-mapping works when all widgets belong to one tab.
func TestAccDatadogDashboardTabSingleTab(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	accProvider := testAccProvider(t, accProviders)

	config := replacer.Replace(`
resource "datadog_dashboard" "tab_dashboard" {
	title       = "{{uniq}}"
	description = "Created using the Datadog provider in Terraform"
	layout_type = "ordered"
	widget {
		note_definition { content = "Widget 1" }
	}
	widget {
		note_definition { content = "Widget 2" }
	}
	tab {
		name       = "All Widgets"
		widget_ids = ["@1", "@2"]
	}
}`)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.name", "All Widgets"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.widget_ids.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.widget_ids.0", "@1"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.widget_ids.1", "@2"),
				),
			},
		},
	})
}

// TestAccDatadogDashboardTabUpdate exercises tab update and removal, specifically
// covering the stale-state fix where d.Set("tab", nil) must be called when tabs
// are absent from the API response.
func TestAccDatadogDashboardTabUpdate(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	accProvider := testAccProvider(t, accProviders)

	// Step 1: create with two tabs
	configCreate := replacer.Replace(`
resource "datadog_dashboard" "tab_dashboard" {
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

	// Step 2: rename a tab and move a widget between tabs
	configUpdate := replacer.Replace(`
resource "datadog_dashboard" "tab_dashboard" {
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

	// Step 3: remove all tabs — exercises the stale-state fix
	configRemove := replacer.Replace(`
resource "datadog_dashboard" "tab_dashboard" {
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
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.name", "Overview"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.1.name", "Details"),
				),
			},
			{
				Config: configUpdate,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.name", "Overview (renamed)"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.0.widget_ids.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.1.widget_ids.#", "2"),
				),
			},
			{
				Config: configRemove,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					resource.TestCheckResourceAttr("datadog_dashboard.tab_dashboard", "tab.#", "0"),
				),
			},
		},
	})
}
