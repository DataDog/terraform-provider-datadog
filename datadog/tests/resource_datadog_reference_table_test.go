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

func TestAccReferenceTableBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTable(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.foo", "description", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.foo", "source", "LOCAL_FILE"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.foo", "table_name", "UPDATE ME"),
				),
			},
		},
	})
}

func testAccCheckDatadogReferenceTable(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_reference_table" "foo" {
    description = "UPDATE ME"
    schema {
    fields {
    name = "UPDATE ME"
    type = "STRING"
    }
    primary_keys = [""]
    }
    source = "LOCAL_FILE"
    table_name = "UPDATE ME"
    tags = "UPDATE ME"
}`, uniq)
}

func testAccCheckDatadogReferenceTableDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := ReferenceTableDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func ReferenceTableDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_reference_table" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetReferenceTablesApiV2().GetTable(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving ReferenceTable %s", err)}
			}
			return &utils.RetryableError{Prob: "ReferenceTable still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogReferenceTableExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := referenceTableExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func referenceTableExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_reference_table" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetReferenceTablesApiV2().GetTable(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving ReferenceTable")
		}
	}
	return nil
}
