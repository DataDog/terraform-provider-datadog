package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccSharedDashboardBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSharedDashboardDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSharedDashboardOpenDashboard(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSharedDashboardExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_shared_dashboard.shared_dashboard", "share_type", "open"),
					resource.TestCheckResourceAttr("datadog_shared_dashboard.shared_dashboard", "dashboard_type", "custom_timeboard"),
					checkRessourceAttributeRegex("datadog_shared_dashboard.shared_dashboard", "public_url", "https://p.datadoghq.com/.*"),
					resource.TestCheckResourceAttrSet("datadog_shared_dashboard.shared_dashboard", "token"),
					resource.TestCheckResourceAttrSet("datadog_shared_dashboard.shared_dashboard", "author_handle"),
				),
			},
		},
	})
}

func testAccCheckDatadogSharedDashboardOpenDashboard(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_dashboard" "dash_two" {
		title = "%s"
		description  = "Created using the Datadog provider in Terraform"
		layout_type  = "ordered"
		template_variable {
          name    = "datacenter"
          prefix  = "datacenter"
          default = "my-datacenter"
        }

		widget {
			alert_graph_definition {
				alert_id = "1234"
				viz_type = "timeseries"
				title = "Widget Title"
				live_span = "1h"
			}
		}
	}

	resource "datadog_shared_dashboard" "shared_dashboard" {
		dashboard_id = datadog_dashboard.dash_two.id
		dashboard_type = "custom_timeboard"
		global_time {
			live_span = "1h"
		}
		global_time_selectable_enabled = true
		selectable_template_vars {
		  default_value = "*"
		  name          = "datacenter"
		  prefix        = "datacenter"
		  visible_tags  = ["value1", "value2", "value3"]
		}
		share_type = "open"
    }
`, uniq)
}

func TestAccSharedDashboardInviteOnly(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSharedDashboardDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSharedDashboardInviteOnlyDashboard(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSharedDashboardExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_shared_dashboard.shared_dashboard", "share_type", "invite"),
					resource.TestCheckResourceAttr("datadog_shared_dashboard.shared_dashboard", "share_list.#", "3"),
					resource.TestCheckResourceAttr("datadog_shared_dashboard.shared_dashboard", "dashboard_type", "custom_timeboard"),
					checkRessourceAttributeRegex("datadog_shared_dashboard.shared_dashboard", "public_url", "https://p.datadoghq.com/.*"),
					resource.TestCheckResourceAttrSet("datadog_shared_dashboard.shared_dashboard", "token"),
					resource.TestCheckResourceAttrSet("datadog_shared_dashboard.shared_dashboard", "author_handle"),
				),
			},
		},
	})
}

func testAccCheckDatadogSharedDashboardInviteOnlyDashboard(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_dashboard" "dash_two" {
		title = "%s"
		description  = "Created using the Datadog provider in Terraform"
		layout_type  = "ordered"
		template_variable {
          name    = "datacenter"
          prefix  = "datacenter"
          default = "my-datacenter"
        }

		widget {
			alert_graph_definition {
				alert_id = "1234"
				viz_type = "timeseries"
				title = "Widget Title"
				live_span = "1h"
			}
		}
	}

	resource "datadog_shared_dashboard" "shared_dashboard" {
		dashboard_id = datadog_dashboard.dash_two.id
		dashboard_type = "custom_timeboard"
		global_time {
			live_span = "1h"
		}
		global_time_selectable_enabled = true
		selectable_template_vars {
		  default_value = "*"
		  name          = "datacenter"
		  prefix        = "datacenter"
		  visible_tags  = ["value1", "value2", "value3"]
		}
		share_list = [ "user1@org.com", "user2@org.com", "user3@org3.com" ]
		share_type = "invite"
    }
`, uniq)
}

func testAccCheckDatadogSharedDashboardDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := SharedDashboardDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func SharedDashboardDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_shared_dashboard" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetDashboardsApiV1().GetPublicDashboard(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving SharedDashboard %s", err)}
			}
			return &utils.RetryableError{Prob: "SharedDashboard still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogSharedDashboardExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := sharedDashboardExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func sharedDashboardExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_shared_dashboard" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetDashboardsApiV1().GetPublicDashboard(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving SharedDashboard")
		}
	}
	return nil
}
