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

func TestAccApmRetentionFilter(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	query := "error_code:123 service:my-service"
	rate := 0.1
	updatedQuery := "error_code:123"
	updatedRate := 1
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApmRetentionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: buildApmRetentionFilterResourceConfig(uniq, query, rate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApmRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_apm_retention_filter.test", "name", uniq),
					resource.TestCheckResourceAttr("datadog_apm_retention_filter.test", "filter.query", query),
					resource.TestCheckResourceAttr("datadog_apm_retention_filter.test", "rate", fmt.Sprintf("%.1f", rate))),
			},
			{
				Config: buildApmRetentionFilterResourceConfig(uniq, updatedQuery, float64(updatedRate)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApmRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_apm_retention_filter.test", "name", uniq),
					resource.TestCheckResourceAttr("datadog_apm_retention_filter.test", "filter.query", updatedQuery),
					resource.TestCheckResourceAttr("datadog_apm_retention_filter.test", "rate", fmt.Sprintf("%d", updatedRate))),
			},
		},
	})
}

func testAccCheckDatadogApmRetentionFilterDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := ApmRetentionFilterDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func ApmRetentionFilterDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_apm_retention_filter" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetApmRetentionFiltersApiV2().GetApmRetentionFilter(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving retention filter %s", err)}
			}
			return &utils.RetryableError{Prob: "retention filter still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogApmRetentionFilterExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := ApmRetentionFilterExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func ApmRetentionFilterExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_apm_retention_filter" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetApmRetentionFiltersApiV2().GetApmRetentionFilter(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving the retention filter")
		}
	}
	return nil
}

func buildApmRetentionFilterResourceConfig(name, query string, rate float64) string {
	return fmt.Sprintf(`
		resource "datadog_apm_retention_filter" "test" {
			name = "%s"
			rate = %f
			filter {
				query = "%s"
			}
			filter_type = "spans-sampling-processor"
			enabled = true
		}
	`, name, rate, query)

}
