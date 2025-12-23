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

func TestAccDatastoreBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDatastoreDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDatastore(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDatastoreExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_datastore.foo", "description", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_datastore.foo", "name", "datastore-name"),
					resource.TestCheckResourceAttr(
						"datadog_datastore.foo", "org_access", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_datastore.foo", "primary_column_name", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_datastore.foo", "primary_key_generation_strategy", "UPDATE ME"),
				),
			},
		},
	})
}

func testAccCheckDatadogDatastore(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_datastore" "foo" {
    description = "UPDATE ME"
    name = "datastore-name"
    org_access = "UPDATE ME"
    primary_column_name = "UPDATE ME"
    primary_key_generation_strategy = "UPDATE ME"
}`, uniq)
}

func testAccCheckDatadogDatastoreDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := DatastoreDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func DatastoreDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_datastore" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetActionsDatastoresApiV2().GetDatastore(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Datastore %s", err)}
			}
			return &utils.RetryableError{Prob: "Datastore still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogDatastoreExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datastoreExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datastoreExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_datastore" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetActionsDatastoresApiV2().GetDatastore(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Datastore")
		}
	}
	return nil
}
