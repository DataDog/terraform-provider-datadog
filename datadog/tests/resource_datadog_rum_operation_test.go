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

const rumOperationResourceName = "datadog_rum_operation.foo"

func TestAccRumOperation(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumOperationDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccRumOperationConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumOperationExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(rumOperationResourceName, "name", name),
					resource.TestCheckResourceAttr(rumOperationResourceName, "display_name", name),
					resource.TestCheckResourceAttr(rumOperationResourceName, "application_id", RumRetentionFilterResourceTestAppId),
					resource.TestCheckResourceAttr(rumOperationResourceName, "description", "Checkout flow"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "category", "form_submission"),
					resource.TestCheckTypeSetElemAttr(rumOperationResourceName, "tags.*", "team:rum"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.#", "2"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.0.type", "start"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.0.nodes.0.query", "@view.name:/checkout"),
				),
			},
			{
				Config: testAccRumOperationConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumOperationExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(rumOperationResourceName, "display_name", "Checkout"),
					resource.TestCheckNoResourceAttr(rumOperationResourceName, "description"),
					resource.TestCheckNoResourceAttr(rumOperationResourceName, "category"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.#", "3"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.2.type", "error"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.2.composite.kind", "in_order"),
					resource.TestCheckResourceAttr(rumOperationResourceName, "steps.2.composite.predicates.#", "2"),
				),
			},
			{
				ResourceName:      rumOperationResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRumOperationConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_rum_operation" "foo" {
  name           = "%s"
  application_id = "%s"
  description    = "Checkout flow"
  category       = "form_submission"
  tags           = ["team:rum"]

  steps = [
    {
      type  = "start"
      nodes = [{ query = "@view.name:/checkout" }]
    },
    {
      type  = "stop"
      nodes = [{ query = "@view.name:/checkout/complete" }]
    },
  ]
}`, name, RumRetentionFilterResourceTestAppId)
}

func testAccRumOperationConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_rum_operation" "foo" {
  name           = "%s"
  application_id = "%s"
  display_name   = "Checkout"

  steps = [
    {
      type  = "start"
      nodes = [{ query = "@view.name:/checkout" }]
    },
    {
      type  = "stop"
      nodes = [{ query = "@view.name:/checkout/complete" }]
    },
    {
      type = "error"
      composite = {
        kind          = "in_order"
        max_window_ms = 60000
        predicates = [
          { query = "@type:action @action.name:pay" },
          { query = "@type:error @error.message:*declined*" },
        ]
      }
    },
  ]
}`, name, RumRetentionFilterResourceTestAppId)
}

func testAccCheckDatadogRumOperationDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		return utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_rum_operation" {
					continue
				}
				_, httpResp, err := accProvider.DatadogApiInstances.GetRUMOperationsApiV2().GetRUMOperation(accProvider.Auth, r.Primary.ID)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						continue
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving RUM operation: %s", err)}
				}
				return &utils.RetryableError{Prob: "RUM operation still exists"}
			}
			return nil
		})
	}
}

func testAccCheckDatadogRumOperationExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_rum_operation" {
				continue
			}
			_, httpResp, err := accProvider.DatadogApiInstances.GetRUMOperationsApiV2().GetRUMOperation(accProvider.Auth, r.Primary.ID)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving RUM operation")
			}
		}
		return nil
	}
}
