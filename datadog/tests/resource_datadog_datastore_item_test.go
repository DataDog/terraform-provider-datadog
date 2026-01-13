package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatastoreItemBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatastoreItemDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatastoreItem(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatastoreItemExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet(
						"datadog_datastore_item.test_item", "datastore_id"),
					resource.TestCheckResourceAttr(
						"datadog_datastore_item.test_item", "item_key", fmt.Sprintf("test-key-%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_datastore_item.test_item", "value.test_field", "test_value"),
				),
			},
		},
	})
}

func testAccCheckDatastoreItem(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_datastore" "test_datastore" {
    name                             = "test-datastore-%s"
    description                      = "Test datastore for items"
    primary_column_name              = "id"
    primary_key_generation_strategy  = "none"
}

resource "datadog_datastore_item" "test_item" {
    datastore_id = datadog_datastore.test_datastore.id
    item_key     = "test-key-%s"
    value = {
        id          = "test-key-%s"
        test_field  = "test_value"
    }
}`, uniq, uniq, uniq)
}

func testAccCheckDatastoreItemDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datastoreItemDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datastoreItemDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_datastore_item" {
				continue
			}

			// Parse composite ID (datastore_id:item_key)
			parts := strings.SplitN(r.Primary.ID, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid ID format: %s", r.Primary.ID)
			}
			datastoreID := parts[0]
			itemKey := parts[1]

			// Try to read the item
			optionalParams := datadogV2.NewListDatastoreItemsOptionalParameters()
			optionalParams.ItemKey = &itemKey

			resp, httpResp, err := apiInstances.GetActionsDatastoresApiV2().ListDatastoreItems(auth, datastoreID, *optionalParams)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Datastore Item: %s", err)}
			}

			// Check if item exists in response
			items := resp.GetData()
			if len(items) == 0 {
				return nil
			}

			return &utils.RetryableError{Prob: "Datastore Item still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatastoreItemExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datastoreItemExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datastoreItemExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_datastore_item" {
			continue
		}

		// Parse composite ID (datastore_id:item_key)
		parts := strings.SplitN(r.Primary.ID, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid ID format: %s", r.Primary.ID)
		}
		datastoreID := parts[0]
		itemKey := parts[1]

		// Try to read the item
		optionalParams := datadogV2.NewListDatastoreItemsOptionalParameters()
		optionalParams.ItemKey = &itemKey

		resp, httpResp, err := apiInstances.GetActionsDatastoresApiV2().ListDatastoreItems(auth, datastoreID, *optionalParams)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Datastore Item")
		}

		// Check if item exists in response
		items := resp.GetData()
		if len(items) == 0 {
			return fmt.Errorf("Datastore Item not found: %s", r.Primary.ID)
		}
	}
	return nil
}
