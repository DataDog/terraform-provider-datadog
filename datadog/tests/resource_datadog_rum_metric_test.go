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

func TestAccRumMetricBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRumMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumMetricExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_metric.foo", "event_type", "session"),
				),
			},
		},
	})
}

func testAccCheckDatadogRumMetric(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_rum_metric" "foo" {
    compute {
    aggregation_type = "distribution"
    include_percentiles = True
    path = "@duration"
    }
    event_type = "session"
    filter {
    query = "@service:web-ui: "
    }
    group_by {
    path = "@browser.name"
    tag_name = "browser_name"
    }
    uniqueness {
    when = "match"
    }
}`, uniq)
}

func testAccCheckDatadogRumMetricDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := RumMetricDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func RumMetricDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_rum_metric" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetRumMetricsApiV2().GetRumMetric(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving RumMetric %s", err)}
			}
			return &utils.RetryableError{Prob: "RumMetric still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogRumMetricExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := rumMetricExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func rumMetricExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_rum_metric" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetRumMetricsApiV2().GetRumMetric(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving RumMetric")
		}
	}
	return nil
}
