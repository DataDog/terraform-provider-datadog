package test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	// Sample UUID for testing (sink org)
	sinkOrgID = "01234567-8901-2345-6789-012345678901"
)

func TestAccDatadogOrgConnection_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	_ = ctx

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOrgConnectionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgConnection(sinkOrgID, []string{"logs"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgConnectionExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_org_connection.foo", "sink_org_id", sinkOrgID),
					resource.TestCheckResourceAttr(
						"datadog_org_connection.foo", "connection_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_org_connection.foo", "connection_types.*", "logs"),
					resource.TestCheckResourceAttrSet(
						"datadog_org_connection.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_org_connection.foo", "source_org_id"),
					resource.TestCheckResourceAttrSet(
						"datadog_org_connection.foo", "source_org_name"),
					resource.TestCheckResourceAttrSet(
						"datadog_org_connection.foo", "sink_org_name"),
					resource.TestCheckResourceAttrSet(
						"datadog_org_connection.foo", "created_at"),
				),
			},
		},
	})
}

func TestAccDatadogOrgConnection_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	_ = ctx

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOrgConnectionDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrgConnection(sinkOrgID, []string{"logs"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgConnectionExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_org_connection.foo", "connection_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_org_connection.foo", "connection_types.*", "logs"),
				),
			},
			{
				Config: testAccCheckDatadogOrgConnectionUpdate(sinkOrgID, []string{"logs", "metrics"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOrgConnectionExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_org_connection.foo", "connection_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_org_connection.foo", "connection_types.*", "logs"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_org_connection.foo", "connection_types.*", "metrics"),
				),
			},
		},
	})
}

func TestAccDatadogOrgConnection_InvalidInput(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	_ = ctx

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogOrgConnectionInvalidUUID("invalid-uuid", []string{"logs"}),
				ExpectError: regexp.MustCompile("must be a valid UUID"),
			},
			{
				Config:      testAccCheckDatadogOrgConnectionEmptyConnectionTypes(sinkOrgID),
				ExpectError: regexp.MustCompile("Attribute connection_types set must contain at least 1 elements"),
			},
			{
				Config:      testAccCheckDatadogOrgConnectionEmptyStringInConnectionTypes(sinkOrgID),
				ExpectError: regexp.MustCompile(`Attribute connection_types\[.*\] string length must be at least 1`),
			},
		},
	})
}

func testAccCheckDatadogOrgConnection(sinkOrgID string, connectionTypes []string) string {
	return fmt.Sprintf(`
		resource "datadog_org_connection" "foo" {
			sink_org_id = "%s"
			connection_types = %s
		}`, sinkOrgID, formatStringSliceForTerraform(connectionTypes))
}

func testAccCheckDatadogOrgConnectionUpdate(sinkOrgID string, connectionTypes []string) string {
	return fmt.Sprintf(`
		resource "datadog_org_connection" "foo" {
			sink_org_id = "%s"
			connection_types = %s
		}`, sinkOrgID, formatStringSliceForTerraform(connectionTypes))
}

func testAccCheckDatadogOrgConnectionInvalidUUID(sinkOrgID string, connectionTypes []string) string {
	return fmt.Sprintf(`
		resource "datadog_org_connection" "foo" {
			sink_org_id = "%s"
			connection_types = %s
		}`, sinkOrgID, formatStringSliceForTerraform(connectionTypes))
}

func testAccCheckDatadogOrgConnectionEmptyConnectionTypes(sinkOrgID string) string {
	return fmt.Sprintf(`
		resource "datadog_org_connection" "foo" {
			sink_org_id = "%s"
			connection_types = []
		}`, sinkOrgID)
}

func testAccCheckDatadogOrgConnectionEmptyStringInConnectionTypes(sinkOrgID string) string {
	return fmt.Sprintf(`
		resource "datadog_org_connection" "foo" {
			sink_org_id = "%s"
			connection_types = [""]
		}`, sinkOrgID)
}

func testAccCheckDatadogOrgConnectionExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_org_connection" {
				continue
			}
			sinkOrgID := r.Primary.Attributes["sink_org_id"]

			queryParams := datadogV2.ListOrgConnectionsOptionalParameters{}
			queryParams.WithSinkOrgId(sinkOrgID)
			resp, httpResp, err := apiInstances.GetOrgConnectionsApiV2().ListOrgConnections(auth, queryParams)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving org connection")
			}
			if len(resp.GetData()) == 0 {
				return fmt.Errorf("org connection not found")
			}
		}
		return nil
	}
}

func testAccCheckDatadogOrgConnectionDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		err := utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_org_connection" {
					continue
				}
				sinkOrgID := r.Primary.Attributes["sink_org_id"]

				queryParams := datadogV2.ListOrgConnectionsOptionalParameters{}
				queryParams.WithSinkOrgId(sinkOrgID)
				resp, httpResp, err := apiInstances.GetOrgConnectionsApiV2().ListOrgConnections(auth, queryParams)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
						return nil
					}
					return utils.TranslateClientError(err, httpResp, "error retrieving org connection")
				}
				if len(resp.GetData()) > 0 {
					return fmt.Errorf("org connection still exists")
				}
			}
			return nil
		})
		return err
	}
}

// Helper function to format a string slice for Terraform configuration
func formatStringSliceForTerraform(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	result := "["
	for i, item := range slice {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf(`"%s"`, item)
	}
	result += "]"
	return result
}
