package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Create a suppression and update its rule query and description without adding an expiration date
func TestAccSecurityMonitoringSuppression_CreateAndUpdateWithoutExpirationDate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	suppressionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_monitoring_suppression.suppression_test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringSuppressionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create suppression
			{
				Config: fmt.Sprintf(`
				resource "datadog_security_monitoring_suppression" "suppression_test" {
					name              = "%s"
					description       = "suppression for terraform provider test"
					enabled           = true
					rule_query        = "severity:low source:cloudtrail"
					suppression_query = "env:staging"
				}
				`, suppressionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContent(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
					),
				),
			},
			// Update description and rule query
			{
				Config: fmt.Sprintf(`
				resource "datadog_security_monitoring_suppression" "suppression_test" {
					name              = "%s"
					description       = "updated suppression for terraform provider test"
					enabled           = true
					rule_query        = "severity:low source:(cloudtrail OR azure)"
					suppression_query = "env:staging"
				}
				`, suppressionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContent(
						resourceName,
						suppressionName,
						"updated suppression for terraform provider test",
						"severity:low source:(cloudtrail OR azure)",
						"env:staging",
					),
				),
			},
		},
	})
}

// Create a suppression without an expiration date, add one, then remove it
func TestAccSecurityMonitoringSuppression_CreateThenAddAndRemoveExpirationDate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	suppressionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_monitoring_suppression.suppression_test_with_expiration_date"

	configWithoutExpirationDate := fmt.Sprintf(`
	resource "datadog_security_monitoring_suppression" "suppression_test_with_expiration_date" {
		name              = "%s"
		description       = "suppression for terraform provider test"
		enabled           = true
		rule_query        = "severity:low source:cloudtrail"
		suppression_query = "env:staging"
	}
	`, suppressionName)

	configWithExpirationDate := func(expirationDate string) string {
		return fmt.Sprintf(`
		resource "datadog_security_monitoring_suppression" "suppression_test_with_expiration_date" {
			name              = "%s"
			description       = "suppression for terraform provider test"
			enabled           = true
			rule_query        = "severity:low source:cloudtrail"
			suppression_query = "env:staging"
			expiration_date   = "%s"
		}
		`, suppressionName, expirationDate)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringSuppressionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create without expiration date
			{
				Config: configWithoutExpirationDate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContent(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
					),
				),
			},
			// Add expiration date
			{
				Config: configWithExpirationDate("2024-01-22T12:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithExpirationDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2024-01-22T12:00:00Z",
					),
				),
			},
			// Change the timezone of the expiration date, without changing the value
			{
				Config: configWithExpirationDate("2024-01-22T13:00:00+01:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithExpirationDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2024-01-22T13:00:00+01:00",
					),
				),
			},
			// Change the expiration date
			{
				Config: configWithExpirationDate("2024-01-22T15:30:00+01:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithExpirationDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2024-01-22T15:30:00+01:00",
					),
				),
			},
			// Remove expiration date
			{
				Config: configWithoutExpirationDate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContent(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
					),
				),
			},
		},
	})
}

func checkSecurityMonitoringSuppressionContent(resourceName string, name string, description string, ruleQuery string, suppressionQuery string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "rule_query", ruleQuery),
		resource.TestCheckResourceAttr(resourceName, "suppression_query", suppressionQuery),
		resource.TestCheckNoResourceAttr(resourceName, "expiration_date"),
	)
}

func checkSecurityMonitoringSuppressionContentWithExpirationDate(resourceName string, name string, description string, ruleQuery string, suppressionQuery string, expirationDate string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "rule_query", ruleQuery),
		resource.TestCheckResourceAttr(resourceName, "suppression_query", suppressionQuery),
		resource.TestCheckResourceAttr(resourceName, "expiration_date", expirationDate),
	)
}

func testAccCheckSecurityMonitoringSuppressionExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state %s", resourceName, s.RootModule().Resources)
		}

		if resource.Type != "datadog_security_monitoring_suppression" {
			return fmt.Errorf("resource %s is not of type datadog_security_monitoring_suppression, found %s instead", resourceName, resource.Type)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		_, _, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringSuppression(auth, resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving suppression: %s", err)
		}

		return nil
	}
}

func testAccCheckSecurityMonitoringSuppressionDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_security_monitoring_suppression" {
				_, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringSuppression(auth, resource.Primary.ID)
				if err == nil {
					return errors.New("suppression still exists")
				}
				if httpResponse == nil || httpResponse.StatusCode != 404 {
					return fmt.Errorf("received an error while getting the suppression: %s", err)
				}
			}
		}

		return nil
	}
}
