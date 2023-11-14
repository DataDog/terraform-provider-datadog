package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func datadogSimplePowerpackConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_powerpack" "simple_powerpack" {
	name = "%s"
	tags = ["tag:foo1"]
	description = "Test Powerpack"
	live_span = "1h"

  template_variables {
	defaults = ["defaults"]
	name     = "datacenter"
  }
	widget {
	  iframe_definition {
		url = "https://google.com"
	  }
	}
}`, uniq)
}

var datadogSimplePowerpackConfigAsserts = []string{
	// Powerpack metadata
	"description = Test Powerpack",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	"live_span = 1h",
	// IFrame widget
	"widget.0.iframe_definition.0.url = https://google.com",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpack_update(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	dbName := uniqueEntityName(ctx, t)
	asserts := datadogSimplePowerpackConfigAsserts
	accProvider := testAccProvider(t, accProviders)
	checks := testCheckResourceAttrs("datadog_powerpack.simple_powerpack", checkPowerpackExists(accProvider), asserts)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkPowerpackDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogSimplePowerpackConfig(dbName),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func checkPowerpackExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_powerpack" {
				continue
			}
			if _, _, err := apiInstances.GetPowerpackApiV2().GetPowerpack(auth, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving powerpack %s", err)
			}
		}
		return nil
	}
}

func checkPowerpackDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_powerpack" {
				continue
			}
			err := utils.Retry(2, 10, func() error {

				if _, httpResp, err := apiInstances.GetPowerpackApiV2().GetPowerpack(auth, r.Primary.ID); err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Powerpack %s", err)}
				}
				return &utils.RetryableError{Prob: "Powerpack still exists"}
			})
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}
}

func testAccDatadogPowerpackWidgetUtil(t *testing.T, config string, name string, assertions []string) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	replacer := strings.NewReplacer("{{uniq}}", uniq)
	config = replacer.Replace(config)
	for i := range assertions {
		assertions[i] = replacer.Replace(assertions[i])
	}
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkPowerpackDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs(name, checkPowerpackExists(accProvider), assertions)...,
				),
			},
		},
	})
}
