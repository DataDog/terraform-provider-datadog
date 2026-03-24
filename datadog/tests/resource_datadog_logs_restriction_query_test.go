package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogLogsRestrictionQuery_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogLogsRestrictionQueryDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogLogsRestrictionQuery(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsRestrictionQueryExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "restriction_query", fmt.Sprintf("service:test-%s", name)),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "id"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "modified_at"),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "role_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDatadogLogsRestrictionQuery_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogLogsRestrictionQueryDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogLogsRestrictionQuery(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsRestrictionQueryExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "restriction_query", fmt.Sprintf("service:test-%s", name)),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "id"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "modified_at"),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "role_ids.#", "0"),
				),
			},
			{
				ResourceName:      "datadog_logs_restriction_query.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckDatadogLogsRestrictionQueryUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsRestrictionQueryExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "restriction_query", fmt.Sprintf("service:updated-%s", name)),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "role_ids.#", "1"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "id"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "modified_at"),
				),
			},
			{
				Config: testAccCheckDatadogLogsRestrictionQueryRemoveRoles(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsRestrictionQueryExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "restriction_query", fmt.Sprintf("service:final-%s", name)),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "role_ids.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDatadogLogsRestrictionQuery(name string) string {
	return fmt.Sprintf(`
		resource "datadog_logs_restriction_query" "test" {
			restriction_query = "service:test-%s"
		}
	`, name)
}

func testAccCheckDatadogLogsRestrictionQueryUpdate(name string) string {
	return fmt.Sprintf(`
		resource "datadog_role" "test" {
			name = "tf-test-role-%s"

			lifecycle {
				ignore_changes = [permission]
			}
		}

		resource "datadog_logs_restriction_query" "test" {
			restriction_query = "service:updated-%s"
			role_ids = [datadog_role.test.id]
		}
	`, name, name)
}

func testAccCheckDatadogLogsRestrictionQueryRemoveRoles(name string) string {
	return fmt.Sprintf(`
		resource "datadog_logs_restriction_query" "test" {
			restriction_query = "service:final-%s"
		}
	`, name)
}

func testAccCheckDatadogLogsRestrictionQueryExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_logs_restriction_query" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetLogsRestrictionQueriesApiV2().GetRestrictionQuery(auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving logs restriction query")
			}
		}
		return nil
	}
}

func testAccCheckDatadogLogsRestrictionQueryDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		err := utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_logs_restriction_query" {
					continue
				}
				id := r.Primary.ID

				_, httpResp, err := apiInstances.GetLogsRestrictionQueriesApiV2().GetRestrictionQuery(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
						return nil
					}
					return utils.TranslateClientError(err, httpResp, "error retrieving logs restriction query")
				}
				return fmt.Errorf("logs restriction query still exists")
			}
			return nil
		})
		return err
	}
}

func TestAccDatadogLogsRestrictionQuery_WithMultipleRoles(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogLogsRestrictionQueryDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogLogsRestrictionQueryMultipleRoles(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsRestrictionQueryExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "restriction_query", fmt.Sprintf("service:test-%s", name)),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "role_ids.#", "2"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "id"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_logs_restriction_query.test", "modified_at"),
				),
			},
			{
				Config: testAccCheckDatadogLogsRestrictionQueryUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogLogsRestrictionQueryExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_logs_restriction_query.test", "role_ids.#", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogLogsRestrictionQueryMultipleRoles(name string) string {
	return fmt.Sprintf(`
		resource "datadog_role" "test1" {
			name = "tf-test-role1-%s"

			lifecycle {
				ignore_changes = [permission]
			}
		}

		resource "datadog_role" "test2" {
			name = "tf-test-role2-%s"

			lifecycle {
				ignore_changes = [permission]
			}
		}

		resource "datadog_logs_restriction_query" "test" {
			restriction_query = "service:test-%s"
			role_ids = [
				datadog_role.test1.id,
				datadog_role.test2.id
			]
		}
	`, name, name, name)
}
