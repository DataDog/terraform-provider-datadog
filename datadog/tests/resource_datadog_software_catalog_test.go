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

func TestAccDatadogCatalogEntity_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	// uniqUpdated := fmt.Sprintf("%s-updated", uniq)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCatalogEntityDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCatalogEntity(uniq),
				Check:  checkCatalogEntityExists(accProvider),
			},
		},
	})
}

func testAccCheckDatadogCatalogEntity(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_software_catalog" "v3_service" {
  entity =<<EOF
apiVersion: v3
kind: service
metadata:
  name: %s
  displayName: Shopping Cart
  tags:
    - tag:value
  links:
    - name: shopping-cart runbook
      type: runbook
      url: https://runbook/shopping-cart
    - name: shopping-cart architecture
      provider: gdoc
      url: https://google.drive/shopping-cart-architecture
      type: doc
    - name: shopping-cart Wiki
      provider: wiki
      url: https://wiki/shopping-cart
      type: doc
    - name: shopping-cart source code
      provider: github
      url: http://github/shopping-cart
      type: repo
  contacts:
    - name: Support Email
      type: email
      contact: team@shopping.com
    - name: Support Slack
      type: slack
      contact: https://www.slack.com/archives/shopping-cart
  owner: myteam
  additionalOwners:
    - name: opsTeam
      type: operator
integrations:
  pagerduty:
    serviceURL: https://www.pagerduty.com/service-directory/Pshopping-cart
  opsgenie:
    serviceURL: https://www.opsgenie.com/service/shopping-cart
    region: US
extensions:
  datadoghq.com/shopping-cart:
    customField: customValue
spec:
  lifecycle: production
  tier: "1"
  type: web
  languages:
    - go
    - python
  dependsOn:
    - service:serviceA
    - service:serviceB
datadog:
  performanceData:
    tags:
      - 'service:shopping-cart'
      - 'hostname:shopping-cart'
  events:
    - name: "deployment events"
      query: "app:myapp AND type:github"
    - name: "event type B"
      query: "app:myapp AND type:github"
  logs:
    - name: "critical logs"
      query: "app:myapp AND type:github"
    - name: "ops logs"
      query: "app:myapp AND type:github"
  pipelines:
    fingerprints:
      - fp1
      - fp2
  codeLocations:
    - repositoryURL: http://github/shopping-cart.git
      paths:
        - baz/*.c
        - bat/**/*
        - ../plop/*.java
    - repositoryURL: http://github/shopping-cart-2.git
      paths:
        - baz/*.c
        - bat/**/*
        - ../plop/*.java
EOF
}`, uniq)
}

func TestAccDatadogCatalogEntity_Order(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCatalogEntityDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCatalogEntityOrder(uniq),
				Check:  checkCatalogEntityExists(accProvider),
			},
		},
	})
}

func testAccCheckDatadogCatalogEntityOrder(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_software_catalog" "entity" {
  entity =<<EOF
apiVersion: v3
kind: service
metadata: 
	name: %s
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

func checkCatalogEntityExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		httpClient := providerConf.DatadogApiInstances.HttpClient
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			err := utils.Retry(5000*time.Millisecond, 4, func() error {
				if _, _, err := utils.SendRequest(auth, httpClient, "GET", "/api/v2/catalog/entity?include=schema&filter[ref]="+r.Primary.ID, nil); err != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving entity %s", err)}
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

func testAccCheckDatadogCatalogEntityDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		httpClient := providerConf.DatadogApiInstances.HttpClient
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			err := utils.Retry(200*time.Millisecond, 4, func() error {
				if _, httpResp, err := utils.SendRequest(auth, httpClient, "GET", "/api/v2/catalog/entity?filter[ref]="+r.Primary.ID, nil); err != nil {
					if httpResp != nil && httpResp.StatusCode != 404 {
						return &utils.RetryableError{Prob: "entity still exists"}
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
