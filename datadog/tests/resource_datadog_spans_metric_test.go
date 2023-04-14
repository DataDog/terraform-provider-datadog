package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccSpansMetricBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSpansMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSpansMetric(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSpansMetricExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogSpansMetric(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_spans_metric" "foo" {
    compute {
    aggregation_type = "distribution"
    include_percentiles = "UPDATE ME"
    path = "@duration"
    }
    filter {
    query = "@http.status_code:200 service:my-service"
    }
    group_by {
    path = "resource_name"
    tag_name = "resource_name"
    }
}`, uniq)
}

func testAccCheckDatadogSpansMetricDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := SpansMetricDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func SpansMetricDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_spans_metric" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetSpansMetricsApiV2().GetSpansMetric(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogSpansMetricExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := spansMetricExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func spansMetricExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_spans_metric" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetSpansMetricsApiV2().GetSpansMetric(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving monitor")
		}
	}
	return nil
}
