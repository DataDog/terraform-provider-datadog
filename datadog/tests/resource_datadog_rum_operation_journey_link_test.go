package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// rum-config does not validate journey ids, so a fixed placeholder avoids depending on DEM journeys.
const rumOperationJourneyLinkTestJourneyId = "tf-test-journey"

func TestAccRumOperationJourneyLink(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	resourceName := "datadog_rum_operation_journey_link.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumOperationJourneyLinkDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccRumOperationJourneyLinkConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "operation_id", rumOperationResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "journey_id", rumOperationJourneyLinkTestJourneyId),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRumOperationJourneyLinkConfig(name string) string {
	return testAccRumOperationConfig(name) + fmt.Sprintf(`

resource "datadog_rum_operation_journey_link" "foo" {
  operation_id = datadog_rum_operation.foo.id
  journey_id   = "%s"
}`, rumOperationJourneyLinkTestJourneyId)
}

func testAccCheckDatadogRumOperationJourneyLinkDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		if err := testAccCheckDatadogRumOperationDestroy(accProvider)(s); err != nil {
			return err
		}
		return utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_rum_operation_journey_link" {
					continue
				}
				resp, _, err := accProvider.DatadogApiInstances.GetRUMOperationsApiV2().ListRUMOperationStrongLinks(accProvider.Auth,
					*datadogV2.NewListRUMOperationStrongLinksOptionalParameters().WithOperationId(r.Primary.Attributes["operation_id"]))
				if err != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error listing RUM operation journey links: %s", err)}
				}
				if len(resp.GetData()) > 0 {
					return &utils.RetryableError{Prob: "RUM operation journey link still exists"}
				}
			}
			return nil
		})
	}
}
