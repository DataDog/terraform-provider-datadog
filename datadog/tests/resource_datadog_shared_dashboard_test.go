package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
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
				Config: testAccCheckDatadogSharedDashboard(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSharedDashboardExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogSharedDashboard(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_shared_dashboard" "foo" {
    body {
    author {
    handle = "test@datadoghq.com"
    name = "UPDATE ME"
    }
    created_at = "UPDATE ME"
    dashboard_id = "123-abc-456"
    dashboard_type = "custom_timeboard"
    global_time {
    live_span = "1h"
    }
    global_time_selectable_enabled = "UPDATE ME"
    public_url = "UPDATE ME"
    selectable_template_vars {
    default_value = "UPDATE ME"
    name = "UPDATE ME"
    prefix = "UPDATE ME"
    visible_tags = "UPDATE ME"
    }
    share_list = ["test@datadoghq.com", "test2@email.com"]
    share_type = "UPDATE ME"
    token = "UPDATE ME"
    }
}`, uniq)
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
