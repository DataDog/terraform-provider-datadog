package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccSecurityMonitoringSuppressionDataSource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	suppressionName := uniqueEntityName(ctx, t)
	dataSourceName := "data.datadog_security_monitoring_suppressions.my_data_source"

	suppressionConfig := fmt.Sprintf(`
	resource "datadog_security_monitoring_suppression" "suppression_for_data_source_test" {
		name              = "%s"
		enabled           = true
		rule_query        = "source:cloudtrail"
		suppression_query = "env:staging"
	}
	`, suppressionName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringSuppressionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create a suppression to have at least one
				Config: suppressionConfig,
				Check:  testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, "datadog_security_monitoring_suppression.suppression_for_data_source_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s

				data "datadog_security_monitoring_suppressions" "my_data_source" {}
				`, suppressionConfig),
				Check: checkSecurityMonitoringSuppressionsDataSourceContent(providers.frameworkProvider, dataSourceName, suppressionName),
			},
		},
	})
}

func checkSecurityMonitoringSuppressionsDataSourceContent(accProvider *fwprovider.FrameworkProvider, dataSourceName string, suppressionName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("resource missing from state: %s", dataSourceName)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		allSuppressionsResponse, _, err := apiInstances.GetSecurityMonitoringApiV2().ListSecurityMonitoringSuppressions(auth)
		if err != nil {
			return err
		}

		// Check the suppression we created is in the API response
		suppressionId := ""
		for _, supp := range allSuppressionsResponse.GetData() {
			if supp.Attributes.GetName() == suppressionName {
				suppressionId = supp.GetId()
				break
			}
		}

		if suppressionId == "" {
			return fmt.Errorf("suppression with name '%s' not found in API responses", suppressionName)
		}

		// Check the suppression we created is in the data source
		resourceAttributes := res.Primary.Attributes

		suppressionIdsCount, err := strconv.Atoi(resourceAttributes["suppression_ids.#"])
		if err != nil {
			return err
		}
		suppressionsCount, err := strconv.Atoi(resourceAttributes["suppressions.#"])
		if err != nil {
			return err
		}

		if suppressionIdsCount != suppressionsCount {
			return fmt.Errorf("the data source contains %d suppression IDs but %d suppressions", suppressionIdsCount, suppressionsCount)
		}

		// Find in which position is the suppression we created, and check its values
		idx := 0
		for idx < suppressionIdsCount && resourceAttributes[fmt.Sprintf("suppression_ids.%d", idx)] != suppressionId {
			idx++
		}

		if idx == suppressionIdsCount {
			return fmt.Errorf("suppression with ID '%s' not found in data source", suppressionId)
		}

		// Now idx is the index of our suppression, we can check the values in the suppressions field of the data source
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("suppressions.%d.name", idx), suppressionName),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("suppressions.%d.enabled", idx), "true"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("suppressions.%d.description", idx), ""),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("suppressions.%d.rule_query", idx), "source:cloudtrail"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("suppressions.%d.suppression_query", idx), "env:staging"),
		)(state)
	}
}
