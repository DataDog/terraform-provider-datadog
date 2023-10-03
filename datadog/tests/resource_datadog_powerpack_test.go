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

func TestAccPowerpackBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPowerpackDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogPowerpack(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPowerpackExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_powerpack.foo", "description", "Powerpack for ABC"),
					resource.TestCheckResourceAttr(
						"datadog_powerpack.foo", "name", "Sample Powerpack"),
				),
			},
		},
	})
}

func testAccCheckDatadogPowerpack(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_powerpack" "foo" {
    description = "Powerpack for ABC"
    group_widget {
    }
    name = "Sample Powerpack"
    tags = ["tag:foo1"]
    template_variables {
    defaults = "UPDATE ME"
    name = "datacenter"
    }
}`, uniq)
}

func testAccCheckDatadogPowerpackDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := PowerpackDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func PowerpackDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_powerpack" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetPowerpackApiV2().GetPowerpack(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Powerpack %s", err)}
			}
			return &utils.RetryableError{Prob: "Powerpack still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogPowerpackExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := powerpackExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func powerpackExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_powerpack" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetPowerpackApiV2().GetPowerpack(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Powerpack")
		}
	}
	return nil
}
