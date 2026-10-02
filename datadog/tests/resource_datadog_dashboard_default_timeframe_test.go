package test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const datadogDashboardDefaultTimeframeConfig = `
resource "datadog_dashboard" "default_timeframe_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"
  description = "Created using the Datadog provider in Terraform"

  default_timeframe {
    live {
      unit  = "week"
      value = 1
    }
  }

  widget {
    note_definition {
      content = "Widget 1"
    }
  }
}
`

var datadogDashboardDefaultTimeframeAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"default_timeframe.# = 1",
	"default_timeframe.0.live.# = 1",
	"default_timeframe.0.live.0.unit = week",
	"default_timeframe.0.live.0.value = 1",
	"widget.# = 1",
}

// defaultTimeframe tests use VCR cassettes: the public v1 dashboard API used by RECORD=none
// integration runs rejects default_timeframe in the request body, so these tests are skipped
// there (same pattern as datadog_dashboard_v2 widget tests).
func testAccDatadogDashboardDefaultTimeframeUtil(t *testing.T, config string, name string, assertions []string) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	if os.Getenv("RECORD") == "none" {
		t.Skip("datadog_dashboard default_timeframe tests require cassettes; skipped when RECORD=none")
	}
	t.Parallel()
	ctx, accProviders := testAccProvidersWithCassette(context.Background(), t, t.Name())
	accProvider := testAccProvider(t, accProviders)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)
	for i := range assertions {
		assertions[i] = replacer.Replace(assertions[i])
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs(name, checkDashboardExists(accProvider), assertions)...,
				),
			},
		},
	})
}

func testAccDatadogDashboardDefaultTimeframeUtilImport(t *testing.T, config string, name string) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}
	if os.Getenv("RECORD") == "none" {
		t.Skip("datadog_dashboard default_timeframe tests require cassettes; skipped when RECORD=none")
	}
	t.Parallel()
	ctx, accProviders := testAccProvidersWithCassette(context.Background(), t, t.Name())
	accProvider := testAccProvider(t, accProviders)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogDashboardDefaultTimeframe(t *testing.T) {
	testAccDatadogDashboardDefaultTimeframeUtil(t, datadogDashboardDefaultTimeframeConfig, "datadog_dashboard.default_timeframe_dashboard", datadogDashboardDefaultTimeframeAsserts)
}

func TestAccDatadogDashboardDefaultTimeframe_import(t *testing.T) {
	testAccDatadogDashboardDefaultTimeframeUtilImport(t, datadogDashboardDefaultTimeframeConfig, "datadog_dashboard.default_timeframe_dashboard")
}
