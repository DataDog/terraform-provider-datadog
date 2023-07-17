package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccDatadogServiceDefinition_BasicV2_1(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceDefinitionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceDefinitionV2_1(uniq),
				Check:  checkServiceDefinitionExists(accProvider),
			},
		},
	})
}

func TestAccDatadogServiceDefinition_BasicBackstage(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceDefinitionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceDefinitionBackstage(uniq),
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

func testAccCheckDatadogServiceDefinitionV2_1(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_definition_yaml" "service_definition" {
  service_definition =<<EOF
schema-version: v2.1
dd-service: %s
contacts:
  - contact: contact@datadoghq.com
    name: Team Email
    type: email
extensions:
  myorgextension: extensionvalue
integrations:
  opsgenie:
    region: US
    service-url: https://my-org.opsgenie.com/service/123e4567-e89b-12d3-a456-426614174000
  pagerduty:
    service-url: https://my-org.pagerduty.com/service-directory/PMyService
links:
  - name: Architecture
    type: doc
    provider: Gigoogle drivetHub
    url: https://my-runbook
  - name: Runbook
    type: runbook
    url: https://my-runbook
  - name: Source Code
    type: repo
    provider: GitHub
    url: https://github.com/DataDog/schema
tags:
  - my:tag
  - service:tag
team: my-team  
EOF
}`, uniq)
}

func testAccCheckDatadogServiceDefinitionBackstage(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_definition_yaml" "service_definition_backstage" {
  service_definition = <<EOF
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  annotations:
    backstage.io/techdocs-ref: http://a/b/c
    some.annotation: value
  namespace: default
  name: %s
  title: Shopping Cart
  description: A shopping cart service
  tags: ["taga:valuea", "tagb:valueb"]
  links:
    - title: Wiki
      url: https://wiki/shopping-cart
      icon: help
  ignore-attribute: 
    id: 1
    value: "value"
spec:
  type: service
  lifecycle: production
  owner: e-commerce
  system: retail
EOF
}`, uniq)
}

func TestAccDatadogServiceDefinition_Order(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceDefinitionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceDefinitionOrder(uniq),
				Check:  checkServiceDefinitionExists(accProvider),
			},
		},
	})
}

func testAccCheckDatadogServiceDefinitionOrder(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_definition_yaml" "service_definition" {
  service_definition =<<EOF
schema-version: v2
dd-service: %s
contacts:
  - name: AA
    type: slack
    contact: AAA
  - name: BB
    type: email
    contact: BBB@example.com
  - name: AA
    type: email
    contact: AAA@example.com
  - name: BB
    type: email
    contact: AAA@example.com
  - name: AA
    type: email
    contact: BBB@example.com
tags:
  - 'bbb'
  - 'aaa'
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
			err := utils.Retry(5000*time.Millisecond, 4, func() error {
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
