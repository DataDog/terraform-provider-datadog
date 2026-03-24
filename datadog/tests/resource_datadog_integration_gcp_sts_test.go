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
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccIntGcpStsMetricNamespaceConfigs(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	t.Run("when mrcs never specified in create and update, uses default", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: accProviders,
			CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					}`, uniq),
					Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"id":       knownvalue.StringExact("prometheus"),
								"disabled": knownvalue.Bool(true),
								"filters":  knownvalue.SetExact(nil),
							}),
						})),
					},
				},
				{
					Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					  is_cspm_enabled                       = "true"
					}`, uniq),
					Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"id":       knownvalue.StringExact("prometheus"),
								"disabled": knownvalue.Bool(true),
								"filters":  knownvalue.SetExact(nil),
							}),
						})),
					},
				},
			},
		})
	})

	t.Run("when mrcs never specified in create, but specified on update, overwrites default", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: accProviders,
			CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					}`, uniq),
					Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"id":       knownvalue.StringExact("prometheus"),
								"disabled": knownvalue.Bool(true),
								"filters":  knownvalue.SetExact(nil),
							}),
						})),
					},
				},
				{
					Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					  metric_namespace_configs = [
						{
						  id       = "aiplatform",
						  disabled = true
                          filters = []
						}
					  ]
					}`, uniq),
					Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"id":       knownvalue.StringExact("aiplatform"),
								"disabled": knownvalue.Bool(true),
								"filters":  knownvalue.SetExact(nil),
							}),
						})),
					},
				},
			},
		})
	})

	t.Run("when mrcs specified in create, but omitted on update, initial configs kept", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: accProviders,
			CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					  metric_namespace_configs = [
						{
						  id       = "aiplatform",
						  disabled = true
                          filters = []
						}
					  ]
					}`, uniq),
					Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"id":       knownvalue.StringExact("aiplatform"),
								"disabled": knownvalue.Bool(true),
								"filters":  knownvalue.SetExact(nil),
							}),
						})),
					},
				},
				{
					Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					}`, uniq),
					Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"id":       knownvalue.StringExact("aiplatform"),
								"disabled": knownvalue.Bool(true),
								"filters":  knownvalue.SetExact(nil),
							}),
						})),
					},
				},
			},
		})
	})
}

func TestAccIntegrationGcpStsBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	cloudRunRevisionFilters := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.StringExact("rev:one"),
		knownvalue.StringExact("rev:two"),
	})
	hostFilters := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.StringExact("host:one"),
	})
	updatedHostFilters := knownvalue.SetExact([]knownvalue.Check{
		knownvalue.StringExact("host:three"),
	})
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					  automute                              = "false"
					  is_cspm_enabled                       = "false"
					  resource_collection_enabled           = "false"
					  is_security_command_center_enabled    = "false"
					  is_resource_change_collection_enabled = "false"
					  is_per_project_quota_enabled          = "false"
					  account_tags                          = ["acct:one", "acct:two", "acct:three"]
					  metric_namespace_configs = [
						{
						  id       = "aiplatform",
						  disabled = true
						  filters = []
						},
						{
						  id       = "pubsub",
						  disabled = false
						  filters = ["snapshot.*", "!*_by_region"]
						}
					  ]
					  cloud_run_revision_filters            = ["rev:one", "rev:two"]
					  host_filters                          = ["host:one"]
					  is_global_location_enabled            = "false"
					  region_filter_configs 				= ["booboi", "lushy"]
					}`, uniq),
				Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("automute"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_cspm_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_per_project_quota_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("account_tags"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("acct:one"),
						knownvalue.StringExact("acct:two"),
						knownvalue.StringExact("acct:three"),
					})),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":       knownvalue.StringExact("aiplatform"),
							"disabled": knownvalue.Bool(true),
							"filters":  knownvalue.SetExact(nil),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":       knownvalue.StringExact("pubsub"),
							"disabled": knownvalue.Bool(false),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("snapshot.*"),
								knownvalue.StringExact("!*_by_region"),
							}),
						}),
					})),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("cloud_run_revision_filters"), cloudRunRevisionFilters),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("host_filters"), hostFilters),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type":    knownvalue.StringExact("cloud_run_revision"),
							"filters": cloudRunRevisionFilters,
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type":    knownvalue.StringExact("gce_instance"),
							"filters": hostFilters,
						}),
					})),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_global_location_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("region_filter_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("booboi"),
						knownvalue.StringExact("lushy")})),
				},
			},
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email                          = "%s@test-project.iam.gserviceaccount.com"
					  automute                              = "true"
					  is_cspm_enabled                       = "true"
					  resource_collection_enabled           = "true"
					  is_security_command_center_enabled    = "true"
					  is_resource_change_collection_enabled = "true"
					  is_per_project_quota_enabled          = "true"
					  metric_namespace_configs = [
						{
						  id       = "appengine",
						  disabled = false
						  filters = ["flex.autoscaler.*"]
						}
					  ]
					  monitored_resource_configs = [
						{
						  type    = "cloud_function"
						  filters = ["func:one", "func:two", "func:three"]
						},
						{
						  type    = "gce_instance"
						  filters = ["host:three"]
						},
					  ]
					  is_global_location_enabled            = "true"
					}`, uniq),
				Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("automute"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_cspm_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_per_project_quota_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("account_tags"), knownvalue.Null()),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":       knownvalue.StringExact("appengine"),
							"disabled": knownvalue.Bool(false),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("flex.autoscaler.*"),
							}),
						}),
					})),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("cloud_run_revision_filters"), knownvalue.SetExact(nil)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("host_filters"), updatedHostFilters),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type": knownvalue.StringExact("cloud_function"),
							"filters": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("func:one"),
								knownvalue.StringExact("func:two"),
								knownvalue.StringExact("func:three"),
							}),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"type":    knownvalue.StringExact("gce_instance"),
							"filters": updatedHostFilters,
						}),
					})),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_global_location_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("region_filter_configs"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccIntegrationGcpStsDefault(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "datadog_integration_gcp_sts" "foo" {
					  client_email = "%s@test-project.iam.gserviceaccount.com"
					}`, uniq),
				Check: resource.ComposeTestCheckFunc(testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("client_email"), knownvalue.StringExact(fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq))),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("automute"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_cspm_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_security_command_center_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("resource_collection_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_resource_change_collection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_per_project_quota_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("account_tags"), knownvalue.Null()),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("metric_namespace_configs"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"id":       knownvalue.StringExact("prometheus"),
							"disabled": knownvalue.Bool(true),
							"filters":  knownvalue.SetExact(nil),
						}),
					})),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("cloud_run_revision_filters"), knownvalue.SetExact(nil)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("host_filters"), knownvalue.SetExact(nil)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("monitored_resource_configs"), knownvalue.SetExact(nil)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("is_global_location_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("datadog_integration_gcp_sts.foo", tfjsonpath.New("region_filter_configs"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckDatadogIntegrationGcpStsDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationGcpStsAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationGcpStsAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_gcp_sts" {
				continue
			}

			resp, httpResp, err := apiInstances.GetGCPIntegrationApiV2().ListGCPSTSAccounts(auth)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving IntegrationGcpStsAccount %s", err)}
			}

			for _, account := range resp.GetData() {
				if account.Id == &r.Primary.ID {
					return &utils.RetryableError{Prob: "IntegrationGcpStsAccount still exists"}
				}
			}

			return nil
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationGcpStsExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationGcpStsAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationGcpStsAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	resp, httpResp, err := apiInstances.GetGCPIntegrationApiV2().ListGCPSTSAccounts(auth)
	if err != nil {
		return utils.TranslateClientError(err, httpResp, "error retrieving Integration Gcp Sts")
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_gcp_sts" {
			continue
		}

		for _, r := range s.RootModule().Resources {
			for _, account := range resp.GetData() {
				if account.GetId() == r.Primary.ID {
					return nil
				}
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving Integration Gcp Sts")
		}
	}
	return nil
}
