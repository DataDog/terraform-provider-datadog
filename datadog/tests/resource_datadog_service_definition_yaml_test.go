package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogServiceDefinition_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	uniqUpdated := fmt.Sprintf("%s-updated", uniq)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceDefinitionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceDefinition(uniq),
				Check:  checkServiceDefinitionExists(accProvider),
			},
			{
				Config: testAccCheckDatadogServiceDefinition(uniqUpdated),
				Check:  checkServiceDefinitionExists(accProvider),
			},
		},
	})
}

func testAccCheckDatadogServiceDefinition(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_definition_yaml" "service_definition" {
  service_definition =<<EOF
schema-version: v2
dd-service: %s
team: E Commerce
contacts:
  - name: Support Email
    type: email
    contact: team@shopping.com
  - name: Support Slack
    type: slack
    contact: 'https://www.slack.com/archives/shopping-cart'
repos:
  - name: shopping-cart source code
    provider: github
    url: 'http://github/shopping-cart'
docs:
  - name: shopping-cart architecture
    provider: gdoc
    url: 'https://google.drive/shopping-cart-architecture'
  - name: shopping-cart service Wiki
    provider: wiki
    url: 'https://wiki/shopping-cart'
links:
  - name: shopping-cart runbook
    type: runbook
    url: 'https://runbook/shopping-cart'
tags:
  - 'business-unit:retail'
  - 'cost-center:engineering'
integrations:
  pagerduty: 'https://www.pagerduty.com/service-directory/Pshopping-cart'
extensions:
  datadoghq.com/shopping-cart:
    customField: customValue
EOF
}`, uniq)

}

func checkServiceDefinitionExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		httpClient := providerConf.DatadogApiInstances.HttpClient
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			err := utils.Retry(200*time.Millisecond, 4, func() error {
				if _, _, err := utils.SendRequest(auth, httpClient, "GET", "/api/v2/services/definitions/"+r.Primary.ID, nil); err != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving service %s", err)}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckDatadogServiceDefinitionDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		httpClient := providerConf.DatadogApiInstances.HttpClient
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			err := utils.Retry(200*time.Millisecond, 4, func() error {
				if _, httpResp, err := utils.SendRequest(auth, httpClient, "GET", "/api/v2/services/definitions/"+r.Primary.ID, nil); err != nil {
					if httpResp != nil && httpResp.StatusCode != 404 {
						return &utils.RetryableError{Prob: "service still exists"}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
