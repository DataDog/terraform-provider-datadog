package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// fixes this issue: https://github.com/DataDog/terraform-provider-datadog/issues/3435
func Test_ReplacingHostFiltersWithMRC(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             checkIntegrationGCPDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
					  project_id                 = "%s"
					  private_key_id             = "1234567890123456789012345678901234567890"
					  private_key                = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
					  client_email               = "%s@awesome-project-id.iam.gserviceaccount.com"
					  client_id                  = "123456789012345678901"
					  host_filters               = "host:one"
					}`, uniq, uniq),
				Check: resource.ComposeTestCheckFunc(checkIntegrationGCPExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("project_id"), knownvalue.StringExact(uniq)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key_id"), knownvalue.StringExact("1234567890123456789012345678901234567890")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key"), knownvalue.StringExact("-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_id"), knownvalue.StringExact("123456789012345678901")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("automute"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("cspm_resource_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("host_filters"), knownvalue.StringExact("host:one")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type": knownvalue.StringExact("gce_instance"),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("host:one"),
							}),
						}),
					})),
				},
			},
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
					  project_id     = "%s"
					  private_key_id = "1234567890123456789012345678901234567890"
					  private_key    = "-----BEGIN PRIVATE KEY-----\n key updated \n-----END PRIVATE KEY-----\n"
					  client_email   = "%s@awesome-project-id.iam.gserviceaccount.com"
					  client_id      = "123456789012345678901"
					  automute       = true
					  monitored_resource_configs = [
						{
						  type    = "gce_instance"
						  filters = ["host:one"]
						},
					  ]
					}`, uniq, uniq),
				Check: resource.ComposeTestCheckFunc(checkIntegrationGCPExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("project_id"), knownvalue.StringExact(uniq)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key_id"), knownvalue.StringExact("1234567890123456789012345678901234567890")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key"), knownvalue.StringExact("-----BEGIN PRIVATE KEY-----\n key updated \n-----END PRIVATE KEY-----\n")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_id"), knownvalue.StringExact("123456789012345678901")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("automute"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("cspm_resource_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("host_filters"), knownvalue.StringExact("host:one")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type": knownvalue.StringExact("gce_instance"),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("host:one"),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestAccDatadogIntegrationGCP(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	cloudRunRevisionFilters := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.StringExact("rev:one"),
		knownvalue.StringExact("rev:two"),
	})
	updatedCloudRunRevisionFilters := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.StringExact("rev:three"),
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             checkIntegrationGCPDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
					  project_id                 = "%s"
					  private_key_id             = "1234567890123456789012345678901234567890"
					  private_key                = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
					  client_email               = "%s@awesome-project-id.iam.gserviceaccount.com"
					  client_id                  = "123456789012345678901"
					  cloud_run_revision_filters = ["rev:one", "rev:two"]
					  host_filters               = "host:one"
					}`, uniq, uniq),
				Check: resource.ComposeTestCheckFunc(checkIntegrationGCPExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("project_id"), knownvalue.StringExact(uniq)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key_id"), knownvalue.StringExact("1234567890123456789012345678901234567890")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key"), knownvalue.StringExact("-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_id"), knownvalue.StringExact("123456789012345678901")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("automute"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("cspm_resource_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("cloud_run_revision_filters"), cloudRunRevisionFilters),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("host_filters"), knownvalue.StringExact("host:one")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type":    knownvalue.StringExact("cloud_run_revision"),
							"filters": cloudRunRevisionFilters,
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type": knownvalue.StringExact("gce_instance"),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("host:one"),
							}),
						}),
					})),
				},
			},
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
					  project_id     = "%s"
					  private_key_id = "1234567890123456789012345678901234567890"
					  private_key    = "-----BEGIN PRIVATE KEY-----\n key updated \n-----END PRIVATE KEY-----\n"
					  client_email   = "%s@awesome-project-id.iam.gserviceaccount.com"
					  client_id      = "123456789012345678901"
					  automute       = true
					  monitored_resource_configs = [
						{
						  type    = "cloud_function"
						  filters = ["func:one", "func:two", "func:three"]
						},
						{
						  type    = "cloud_run_revision"
						  filters = ["rev:three"]
						},
					  ]
					}`, uniq, uniq),
				Check: resource.ComposeTestCheckFunc(checkIntegrationGCPExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("project_id"), knownvalue.StringExact(uniq)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key_id"), knownvalue.StringExact("1234567890123456789012345678901234567890")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("private_key"), knownvalue.StringExact("-----BEGIN PRIVATE KEY-----\n key updated \n-----END PRIVATE KEY-----\n")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("client_id"), knownvalue.StringExact("123456789012345678901")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("automute"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("cspm_resource_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("cloud_run_revision_filters"), updatedCloudRunRevisionFilters),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("host_filters"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue("datadog_integration_gcp.awesome_gcp_project_integration", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type": knownvalue.StringExact("cloud_function"),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("func:one"),
								knownvalue.StringExact("func:two"),
								knownvalue.StringExact("func:three"),
							}),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type":    knownvalue.StringExact("cloud_run_revision"),
							"filters": updatedCloudRunRevisionFilters,
						}),
					})),
				},
			},
		},
	})
}

func checkIntegrationGCPExists(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		integrations, _, err := apiInstances.GetGCPIntegrationApiV1().ListGCPIntegration(auth)
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			projectID := r.Primary.ID
			for _, integration := range integrations {
				if integration.GetProjectId() == projectID {
					return nil
				}
			}
			return fmt.Errorf("the Google Cloud Platform integration doesn't exist: projectID=%s", projectID)
		}
		return nil
	}
}

func checkIntegrationGCPDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		integrations, _, err := apiInstances.GetGCPIntegrationApiV1().ListGCPIntegration(auth)
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			projectID := r.Primary.ID
			for _, integration := range integrations {
				if integration.GetProjectId() == projectID {
					return fmt.Errorf("the Google Cloud Platform integration still exist: projectID=%s", projectID)
				}
			}
		}
		return nil
	}
}
