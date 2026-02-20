package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogLogsCustomDestination_basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	destinationWithRequiredFieldsOnly := `
		resource "datadog_logs_custom_destination" "sample_destination" {
			name = "` + name + `"
			http_destination {
				endpoint = "https://example.org"
				basic_auth {
					username = "test-user"
					password = "test-pass"
				}
			}
		}
	`

	path := "datadog_logs_custom_destination.sample_destination"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedLogsCustomDestinations(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: destinationWithRequiredFieldsOnly,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", name),
					resource.TestCheckResourceAttr(path, "query", ""),
					resource.TestCheckResourceAttr(path, "enabled", "true"),
					resource.TestCheckResourceAttr(path, "forward_tags", "true"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list.#", "0"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list_type", "ALLOW_LIST"),

					resource.TestCheckResourceAttr(path, "http_destination.#", "1"),
					resource.TestCheckResourceAttr(path, "http_destination.0.endpoint", "https://example.org"),
					resource.TestCheckResourceAttr(path, "http_destination.0.basic_auth.#", "1"),
					resource.TestCheckResourceAttr(path, "http_destination.0.basic_auth.0.username", "test-user"),
					resource.TestCheckResourceAttr(path, "http_destination.0.basic_auth.0.password", "test-pass"),
				),
			},
		},
	})
}

func TestAccDatadogLogsCustomDestination_forwarder_types(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	nameUpdated := name + "-updated"

	httpWithBasicAuth := `
		http_destination {
			endpoint = "https://example.org"
			basic_auth {
				username = "test-user"
				password = "test-pass"
			}
		}
	`

	httpWithCustomHeaderAuth := `
		http_destination {
			endpoint = "https://example.org"
			custom_header_auth {
				header_name = "test-header-name"
				header_value = "test-header-value"
			}
		}
	`

	splunk := `
		splunk_destination {
			endpoint = "https://example.org"
			access_token = "test-token"
		}
	`

	elasticsearch := `
		elasticsearch_destination {
			endpoint       = "https://example.org"
			index_name     = "test-index"
			index_rotation = "yyyy-'W'ww"
			basic_auth {
				username = "test-user"
				password = "test-pass"
			}
		}
	`

	path := "datadog_logs_custom_destination.sample_destination"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedLogsCustomDestinations(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateLogsCustomDestination(name, httpWithBasicAuth),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", name),
					resource.TestCheckResourceAttr(path, "query", "non-existent-query"),
					resource.TestCheckResourceAttr(path, "enabled", "false"),
					resource.TestCheckResourceAttr(path, "forward_tags", "true"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list.#", "1"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list.0", "a"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list_type", "ALLOW_LIST"),

					resource.TestCheckResourceAttr(path, "http_destination.#", "1"),
					resource.TestCheckResourceAttr(path, "http_destination.0.endpoint", "https://example.org"),
					resource.TestCheckResourceAttr(path, "http_destination.0.basic_auth.#", "1"),
					resource.TestCheckResourceAttr(path, "http_destination.0.basic_auth.0.username", "test-user"),
					resource.TestCheckResourceAttr(path, "http_destination.0.basic_auth.0.password", "test-pass"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateLogsCustomDestination(nameUpdated, httpWithCustomHeaderAuth),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", nameUpdated),
					resource.TestCheckResourceAttr(path, "query", "updated-non-existent-query"),
					resource.TestCheckResourceAttr(path, "enabled", "true"),
					resource.TestCheckResourceAttr(path, "forward_tags", "false"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list.#", "2"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list.0", "a"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list.1", "b"),
					resource.TestCheckResourceAttr(path, "forward_tags_restriction_list_type", "BLOCK_LIST"),

					resource.TestCheckResourceAttr(path, "http_destination.#", "1"),
					resource.TestCheckResourceAttr(path, "http_destination.0.endpoint", "https://example.org"),
					resource.TestCheckResourceAttr(path, "http_destination.0.custom_header_auth.#", "1"),
					resource.TestCheckResourceAttr(path, "http_destination.0.custom_header_auth.0.header_name", "test-header-name"),
					resource.TestCheckResourceAttr(path, "http_destination.0.custom_header_auth.0.header_value", "test-header-value"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateLogsCustomDestination(nameUpdated, splunk),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "http_destination.#", "0"),
					resource.TestCheckResourceAttr(path, "splunk_destination.#", "1"),
					resource.TestCheckResourceAttr(path, "splunk_destination.0.endpoint", "https://example.org"),
					resource.TestCheckResourceAttr(path, "splunk_destination.0.access_token", "test-token"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateLogsCustomDestination(nameUpdated, elasticsearch),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.#", "1"),
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.0.endpoint", "https://example.org"),
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.0.index_name", "test-index"),
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.0.index_rotation", "yyyy-'W'ww"),
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.0.basic_auth.#", "1"),
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.0.basic_auth.0.username", "test-user"),
					resource.TestCheckResourceAttr(path, "elasticsearch_destination.0.basic_auth.0.password", "test-pass"),
				),
			},
		},
	})
}

func testAccCheckDatadogCreateLogsCustomDestination(name string, destination string) string {
	return fmt.Sprintf(`
		resource "datadog_logs_custom_destination" "sample_destination" {
			name                               = "%s"
			query                              = "non-existent-query"
			enabled                            = false
			forward_tags                       = true
			forward_tags_restriction_list      = ["a"]
			forward_tags_restriction_list_type = "ALLOW_LIST"
			%s
		}
	`, name, destination)
}

func testAccCheckDatadogUpdateLogsCustomDestination(name string, destination string) string {
	return fmt.Sprintf(`
		resource "datadog_logs_custom_destination" "sample_destination" {
			name                               = "%s"
			query                              = "updated-non-existent-query"
			enabled                            = true
			forward_tags                       = false
			forward_tags_restriction_list      = ["a", "b"]
			forward_tags_restriction_list_type = "BLOCK_LIST"
			%s
		}
	`, name, destination)
}

// testAccCleanupOrphanedLogsCustomDestinations deletes disabled custom destinations
// that were left behind by previous test runs or external sources, to free up quota.
func testAccCleanupOrphanedLogsCustomDestinations(t *testing.T, frameworkProvider *fwprovider.FrameworkProvider) {
	apiInstances := frameworkProvider.DatadogApiInstances
	auth := frameworkProvider.Auth
	api := apiInstances.GetLogsCustomDestinationsApiV2()

	resp, _, err := api.ListLogsCustomDestinations(auth)
	if err != nil {
		t.Logf("Warning: Could not list custom destinations for cleanup: %v", err)
		return
	}

	destinations := resp.GetData()
	t.Logf("Found %d existing custom destinations, cleaning up disabled ones...", len(destinations))

	for _, dest := range destinations {
		id := dest.GetId()
		attrs, ok := dest.GetAttributesOk()
		if !ok {
			continue
		}
		name := attrs.GetName()
		enabled := attrs.GetEnabled()

		if !enabled {
			t.Logf("Deleting disabled custom destination: %s (ID: %s)", name, id)
			_, err := api.DeleteLogsCustomDestination(auth, id)
			if err != nil {
				t.Logf("Warning: Could not delete custom destination %s: %v", id, err)
			}
		}
	}
}
