package test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// A separate gauge fixture is ingested without creating any tag or aggregation
// configuration. Recording or live execution requires an owned sandbox metric.
func TestAccMetricHistoricalIngestion(t *testing.T) {
	skipIfNoCassette(t)
	ctx, providers, factories := testAccFrameworkMuxProviders(context.Background(), t)
	name := "tf.historical." + strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")
	address := "datadog_metric_historical_ingestion.test"
	config := fmt.Sprintf(`resource "datadog_metric_historical_ingestion" "test" { metric_name = %q }`, name)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: factories,
		CheckDestroy:             historicalIngestionCheckDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{PreConfig: submitTestMetricPreConfig(t, name, "gauge"), Config: config, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(address, "id", name),
				resource.TestCheckResourceAttr(address, "metric_name", name),
			)},
			{ResourceName: address, ImportState: true, ImportStateVerify: true},
			{Config: config, PlanOnly: true},
			// Simulate an external disable and wait for a definitive absence before
			// refresh; the next apply must recreate the resource.
			{PreConfig: func() { historicalIngestionDisableForDrift(t, providers.frameworkProvider, name) }, Config: config,
				Check: resource.TestCheckResourceAttr(address, "id", name)},
			{Config: config, PlanOnly: true},
		},
	})
}

func historicalIngestionDisableForDrift(t *testing.T, p *fwprovider.FrameworkProvider, name string) {
	t.Helper()
	api := p.DatadogApiInstances.GetMetricsApiV2()
	h, err := api.DeleteHistoricalMetricsConfiguration(p.Auth, name)
	if err != nil {
		t.Fatalf("disable fixture: %v", utils.TranslateClientError(err, h, "disable historical ingestion"))
	}
	err = retry.RetryContext(p.Auth, 60*time.Second, func() *retry.RetryError {
		_, h, err := api.GetHistoricalMetricsConfiguration(p.Auth, name)
		if err != nil {
			if h != nil && h.StatusCode == http.StatusNotFound {
				return nil
			}
			return retry.NonRetryableError(utils.TranslateClientError(err, h, "verify drift"))
		}
		return retry.RetryableError(fmt.Errorf("historical ingestion remains visible"))
	})
	if err != nil {
		t.Fatalf("verify fixture drift: %v", err)
	}
}

func historicalIngestionCheckDestroy(p *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "datadog_metric_historical_ingestion" {
				continue
			}
			_, h, err := p.DatadogApiInstances.GetMetricsApiV2().GetHistoricalMetricsConfiguration(p.Auth, rs.Primary.ID)
			if err != nil && h != nil && h.StatusCode == http.StatusNotFound {
				continue
			}
			if err != nil {
				return utils.TranslateClientError(err, h, "check historical ingestion destruction")
			}
			return fmt.Errorf("historical ingestion still exists for %s", rs.Primary.ID)
		}
		return nil
	}
}
