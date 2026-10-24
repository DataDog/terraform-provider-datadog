package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogFeatureFlag_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFeatureFlagDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogFeatureFlagConfig(uniq, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogFeatureFlagExists(providers.frameworkProvider),
					testAccCheckDatadogFeatureFlagEnvironmentExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_feature_flag.foo", "name", "Feature flag acceptance"),
					resource.TestCheckResourceAttr("datadog_feature_flag.foo", "variants.#", "2"),
					resource.TestCheckResourceAttrSet("datadog_feature_flag_environment.foo", "key"),
					resource.TestCheckResourceAttr("datadog_feature_flag_environment_config.foo", "enabled", "true"),
					resource.TestCheckResourceAttr("datadog_feature_flag_allocations.foo", "allocations.#", "1"),
					resource.TestCheckResourceAttrSet("datadog_feature_flag_allocations.foo", "allocations.0.id"),
				),
			},
			{
				Config: testAccDatadogFeatureFlagConfig(uniq, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_feature_flag.foo", "name", "Feature flag acceptance updated"),
					resource.TestCheckResourceAttr("datadog_feature_flag_environment_config.foo", "enabled", "false"),
					resource.TestCheckResourceAttr("datadog_feature_flag_allocations.foo", "allocations.0.name", "Staff updated"),
				),
			},
			{
				ResourceName:      "datadog_feature_flag.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "datadog_feature_flag_environment.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "datadog_feature_flag_environment_config.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "datadog_feature_flag_allocations.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccDatadogFeatureFlagConfig(uniq string, updated bool) string {
	name := "Feature flag acceptance"
	allocationName := "Staff"
	enabled := true
	if updated {
		name += " updated"
		allocationName += " updated"
		enabled = false
	}
	return fmt.Sprintf(`
resource "datadog_feature_flag" "foo" {
  key                 = "%[1]s"
  name                = "%[2]s"
  description         = "%[2]s description"
  value_type          = "BOOLEAN"
  default_variant_key = "off"
  variants = [
    { key = "off", name = "Off", value = "false" },
    { key = "on",  name = "On",  value = "true"  },
  ]
}

resource "datadog_feature_flag_environment" "foo" {
  name    = "%[1]s-env"
  queries = ["env:acceptance"]
}

resource "datadog_feature_flag_environment_config" "foo" {
  feature_flag_id = datadog_feature_flag.foo.id
  environment_id  = datadog_feature_flag_environment.foo.id
  enabled         = %[3]t
}

resource "datadog_feature_flag_allocations" "foo" {
  feature_flag_id = datadog_feature_flag.foo.id
  environment_id  = datadog_feature_flag_environment.foo.id
  depends_on      = [datadog_feature_flag_environment_config.foo]

  allocations = [{
    key  = "%[1]s-staff"
    name = "%[4]s"
    type = "FEATURE_GATE"
    targeting_rules = [{
      conditions = [{
        attribute = "user.tier"
        operator  = "EQUALS"
        value     = ["staff"]
      }]
    }]
    variant_weights = [{
      variant_key = "on"
      value       = 100
    }]
  }]
}
`, uniq, name, enabled, allocationName)
}

func testAccCheckDatadogFeatureFlagExists(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "datadog_feature_flag" {
				continue
			}
			id, err := uuid.Parse(item.Primary.ID)
			if err != nil {
				return err
			}
			flag, httpResponse, err := provider.DatadogApiInstances.GetFeatureFlagsApiV2().GetFeatureFlag(provider.Auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResponse, "error reading feature flag")
			}
			if flag.Data.Attributes.ArchivedAt.IsSet() && flag.Data.Attributes.ArchivedAt.Get() != nil {
				return fmt.Errorf("feature flag %s is archived", id)
			}
		}
		return nil
	}
}

func testAccCheckDatadogFeatureFlagEnvironmentExists(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			if item.Type != "datadog_feature_flag_environment" {
				continue
			}
			id, err := uuid.Parse(item.Primary.ID)
			if err != nil {
				return err
			}
			_, httpResponse, err := provider.DatadogApiInstances.GetFeatureFlagsApiV2().GetFeatureFlagsEnvironment(provider.Auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResponse, "error reading feature flag environment")
			}
		}
		return nil
	}
}

func testAccCheckDatadogFeatureFlagDestroy(provider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(state *terraform.State) error {
		for _, item := range state.RootModule().Resources {
			switch item.Type {
			case "datadog_feature_flag":
				id, err := uuid.Parse(item.Primary.ID)
				if err != nil {
					return err
				}
				flag, httpResponse, err := provider.DatadogApiInstances.GetFeatureFlagsApiV2().GetFeatureFlag(provider.Auth, id)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
						continue
					}
					return utils.TranslateClientError(err, httpResponse, "error reading archived feature flag")
				}
				if !flag.Data.Attributes.ArchivedAt.IsSet() || flag.Data.Attributes.ArchivedAt.Get() == nil {
					return fmt.Errorf("feature flag %s was not archived", id)
				}
			case "datadog_feature_flag_environment":
				id, err := uuid.Parse(item.Primary.ID)
				if err != nil {
					return err
				}
				_, httpResponse, err := provider.DatadogApiInstances.GetFeatureFlagsApiV2().GetFeatureFlagsEnvironment(provider.Auth, id)
				if err == nil {
					return fmt.Errorf("feature flag environment %s still exists", id)
				}
				if httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound {
					return utils.TranslateClientError(err, httpResponse, "error reading deleted feature flag environment")
				}
			}
		}
		return nil
	}
}
