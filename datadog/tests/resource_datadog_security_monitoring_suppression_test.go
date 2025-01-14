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

// Create a suppression and update its rule query and description without adding a start date and an expiration date
func TestAccSecurityMonitoringSuppression_CreateAndUpdateWithoutDates(t *testing.T) {
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

// Create a suppression without a start date, add one, then remove it
func TestAccSecurityMonitoringSuppression_CreateThenAddAndRemoveStartDate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	suppressionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_monitoring_suppression.suppression_test_with_start_date"

	configWithoutStartDate := fmt.Sprintf(`
	resource "datadog_security_monitoring_suppression" "suppression_test_with_start_date" {
		name              = "%s"
		description       = "suppression for terraform provider test"
		enabled           = true
		rule_query        = "severity:low source:cloudtrail"
		suppression_query = "env:staging"
	}
	`, suppressionName)

	configWithStartDate := func(startDate string) string {
		return fmt.Sprintf(`
		resource "datadog_security_monitoring_suppression" "suppression_test_with_start_date" {
			name              = "%s"
			description       = "suppression for terraform provider test"
			enabled           = true
			rule_query        = "severity:low source:cloudtrail"
			suppression_query = "env:staging"
			start_date   = "%s"
		}
		`, suppressionName, startDate)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringSuppressionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create without start date
			{
				Config: configWithoutStartDate,
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
			// Add start date
			{
				Config: configWithStartDate("2099-01-22T12:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithStartDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2099-01-22T12:00:00Z",
					),
				),
			},
			// Change the timezone of the start date, without changing the value
			{
				Config: configWithStartDate("2099-01-22T13:00:00+01:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithStartDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2099-01-22T13:00:00+01:00",
					),
				),
			},
			// Change the start date
			{
				Config: configWithStartDate("2099-01-22T15:30:00+01:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithStartDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2099-01-22T15:30:00+01:00",
					),
				),
			},
			// Remove start date
			{
				Config: configWithoutStartDate,
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
				Config: configWithExpirationDate("2099-01-22T12:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithExpirationDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2099-01-22T12:00:00Z",
					),
				),
			},
			// Change the timezone of the expiration date, without changing the value
			{
				Config: configWithExpirationDate("2099-01-22T13:00:00+01:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithExpirationDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2099-01-22T13:00:00+01:00",
					),
				),
			},
			// Change the expiration date
			{
				Config: configWithExpirationDate("2099-01-22T15:30:00+01:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithExpirationDate(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						"env:staging",
						"2099-01-22T15:30:00+01:00",
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

// Create a suppression with a suppression query, then replace it with an exclusion query, then add another suppression query
func TestAccSecurityMonitoringSuppression_CreateAndUpdateDataExclusionQuery(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	suppressionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_monitoring_suppression.suppression_test_exclusion"
	dataExclusionQuery := "@account_name:staging"
	suppressionQuery := "@usr.team:internal-security-testing"

	configWithSuppressionQuery := fmt.Sprintf(`
	resource "datadog_security_monitoring_suppression" "suppression_test_exclusion" {
		name              = "%s"
		description       = "suppression for terraform provider test"
		enabled           = true
		rule_query        = "severity:low source:cloudtrail"
		suppression_query = "%s"
	}
	`, suppressionName, suppressionQuery)

	configWithDataExclusionQuery := fmt.Sprintf(`
	resource "datadog_security_monitoring_suppression" "suppression_test_exclusion" {
		name                   = "%s"
		description            = "suppression for terraform provider test"
		enabled                = true
		rule_query             = "severity:low source:cloudtrail"
		data_exclusion_query   = "%s"
	}
	`, suppressionName, dataExclusionQuery)

	configWithBoth := fmt.Sprintf(`
	resource "datadog_security_monitoring_suppression" "suppression_test_exclusion" {
		name                   = "%s"
		description            = "suppression for terraform provider test"
		enabled                = true
		rule_query             = "severity:low source:cloudtrail"
		suppression_query      = "%s"
		data_exclusion_query   = "%s"
	}
	`, suppressionName, suppressionQuery, dataExclusionQuery)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringSuppressionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create with suppression query
			{
				Config: configWithSuppressionQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithDataExclusionQuery(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						&suppressionQuery,
						nil,
					),
				),
			},
			// Replace by data exclusion query
			{
				Config: configWithDataExclusionQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithDataExclusionQuery(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						nil,
						&dataExclusionQuery,
					),
				),
			},
			// Add the suppression query without removing the exclusion query
			{
				Config: configWithBoth,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithDataExclusionQuery(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						&suppressionQuery,
						&dataExclusionQuery,
					),
				),
			},
			// Remove exclusion query
			{
				Config: configWithSuppressionQuery,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringSuppressionExists(providers.frameworkProvider, resourceName),
					checkSecurityMonitoringSuppressionContentWithDataExclusionQuery(
						resourceName,
						suppressionName,
						"suppression for terraform provider test",
						"severity:low source:cloudtrail",
						&suppressionQuery,
						nil,
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
		resource.TestCheckNoResourceAttr(resourceName, "start_date"),
		resource.TestCheckNoResourceAttr(resourceName, "expiration_date"),
		resource.TestCheckNoResourceAttr(resourceName, "data_exclusion_query"),
	)
}

func checkSecurityMonitoringSuppressionContentWithStartDate(resourceName string, name string, description string, ruleQuery string, suppressionQuery string, startDate string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "rule_query", ruleQuery),
		resource.TestCheckResourceAttr(resourceName, "suppression_query", suppressionQuery),
		resource.TestCheckResourceAttr(resourceName, "start_date", startDate),
		resource.TestCheckNoResourceAttr(resourceName, "expiration_date"),
		resource.TestCheckNoResourceAttr(resourceName, "data_exclusion_query"),
	)
}

func checkSecurityMonitoringSuppressionContentWithExpirationDate(resourceName string, name string, description string, ruleQuery string, suppressionQuery string, expirationDate string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "rule_query", ruleQuery),
		resource.TestCheckResourceAttr(resourceName, "suppression_query", suppressionQuery),
		resource.TestCheckNoResourceAttr(resourceName, "start_date"),
		resource.TestCheckResourceAttr(resourceName, "expiration_date", expirationDate),
		resource.TestCheckNoResourceAttr(resourceName, "data_exclusion_query"),
	)
}

func testCheckOptionalResourceAttr(name string, key string, value *string) resource.TestCheckFunc {
	if value == nil {
		return resource.TestCheckNoResourceAttr(name, key)
	} else {
		return resource.TestCheckResourceAttr(name, key, *value)
	}
}

func checkSecurityMonitoringSuppressionContentWithDataExclusionQuery(resourceName string, name string, description string, ruleQuery string, suppressionQuery *string, dataExclusionQuery *string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "rule_query", ruleQuery),
		testCheckOptionalResourceAttr(resourceName, "suppression_query", suppressionQuery),
		testCheckOptionalResourceAttr(resourceName, "data_exclusion_query", dataExclusionQuery),
		resource.TestCheckNoResourceAttr(resourceName, "start_date"),
		resource.TestCheckNoResourceAttr(resourceName, "expiration_date"),
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
