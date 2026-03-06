package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogDashboardSecureEmbeddedDashboard_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	title := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashboardSecureEmbeddedDashboardDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardSecureEmbeddedDashboardBasic(title),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "title", title),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "status", "active"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "global_time_live_span", "1h"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "global_time_selectable", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "viewing_preferences_theme", "system"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "viewing_preferences_high_density", "false"),
					resource.TestCheckResourceAttrSet("datadog_dashboard_secure_embedded_dashboard.foo", "token"),
					resource.TestCheckResourceAttrSet("datadog_dashboard_secure_embedded_dashboard.foo", "url"),
					resource.TestCheckResourceAttrSet("datadog_dashboard_secure_embedded_dashboard.foo", "credential"),
					resource.TestCheckResourceAttrSet("datadog_dashboard_secure_embedded_dashboard.foo", "id"),
				),
			},
		},
	})
}

func TestAccDatadogDashboardSecureEmbeddedDashboard_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	title := uniqueEntityName(ctx, t)
	updatedTitle := title + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashboardSecureEmbeddedDashboardDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardSecureEmbeddedDashboardBasic(title),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "title", title),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "status", "active"),
				),
			},
			{
				Config: testAccCheckDatadogDashboardSecureEmbeddedDashboardUpdated(title, updatedTitle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "title", updatedTitle),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "status", "paused"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "global_time_live_span", "4h"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "viewing_preferences_theme", "dark"),
					// Token and credential must be stable across updates
					resource.TestCheckResourceAttrSet("datadog_dashboard_secure_embedded_dashboard.foo", "token"),
					resource.TestCheckResourceAttrSet("datadog_dashboard_secure_embedded_dashboard.foo", "credential"),
				),
			},
		},
	})
}

func TestAccDatadogDashboardSecureEmbeddedDashboard_WithTemplateVars(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	title := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashboardSecureEmbeddedDashboardDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashboardSecureEmbeddedDashboardWithTemplateVars(title),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "title", title),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "selectable_template_vars.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "selectable_template_vars.0.name", "env"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "selectable_template_vars.0.prefix", "env"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "selectable_template_vars.0.default_values.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard_secure_embedded_dashboard.foo", "selectable_template_vars.0.default_values.0", "prod"),
				),
			},
		},
	})
}

// CheckDestroy verifies the secure embed is deleted after the test.
func testAccCheckDatadogDashboardSecureEmbeddedDashboardDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_dashboard_secure_embedded_dashboard" {
				continue
			}
			dashboardID := r.Primary.Attributes["dashboard_id"]
			token := r.Primary.Attributes["token"]
			path := fmt.Sprintf("/api/v2/dashboard/%s/shared/secure-embed/%s", dashboardID, token)

			err := utils.Retry(200*time.Millisecond, 4, func() error {
				_, httpResp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", path, nil)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: "secure embed still exists or error reading: " + err.Error()}
				}
				return &utils.RetryableError{Prob: "secure embed still exists"}
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// minimalDashboardConfig returns a shared config block creating a simple dashboard.
// The dashboard ID is referenced via datadog_dashboard.test.id.
func minimalDashboardConfig(title string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "test" {
  title        = "%s"
  layout_type  = "ordered"
  is_read_only = false

  widget {
    note_definition {
      content          = "placeholder"
      background_color = "white"
      font_size        = "14"
      text_align       = "left"
      has_padding      = true
      show_tick        = false
    }
  }
}
`, title)
}

func testAccCheckDatadogDashboardSecureEmbeddedDashboardBasic(title string) string {
	return minimalDashboardConfig(title) + fmt.Sprintf(`
resource "datadog_dashboard_secure_embedded_dashboard" "foo" {
  dashboard_id = datadog_dashboard.test.id
  title        = "%s"
}
`, title)
}

func testAccCheckDatadogDashboardSecureEmbeddedDashboardUpdated(dashTitle, embedTitle string) string {
	return minimalDashboardConfig(dashTitle) + fmt.Sprintf(`
resource "datadog_dashboard_secure_embedded_dashboard" "foo" {
  dashboard_id          = datadog_dashboard.test.id
  title                 = "%s"
  status                = "paused"
  global_time_live_span = "4h"
  viewing_preferences_theme = "dark"
}
`, embedTitle)
}

func testAccCheckDatadogDashboardSecureEmbeddedDashboardWithTemplateVars(title string) string {
	return minimalDashboardConfig(title) + fmt.Sprintf(`
resource "datadog_dashboard_secure_embedded_dashboard" "foo" {
  dashboard_id = datadog_dashboard.test.id
  title        = "%s"

  selectable_template_vars {
    name           = "env"
    prefix         = "env"
    default_values = ["prod"]
    visible_tags   = ["prod", "staging"]
  }
}
`, title)
}
