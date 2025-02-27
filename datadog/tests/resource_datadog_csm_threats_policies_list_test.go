package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Create a policies_list and update the name and priority of its policy
func TestAccCSMThreatsPoliciesList_CreateAndUpdate(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_csm_threats_policies_list.all"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsPoliciesListDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCSMThreatsPoliciesListConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPoliciesListExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "entries.0.name", "TERRAFORM_POLICY1"),
					resource.TestCheckResourceAttr(resourceName, "entries.0.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "entries.1.name", "TERRAFORM_POLICY2"),
					resource.TestCheckResourceAttr(resourceName, "entries.1.priority", "3"),
				),
			},
			{
				Config: testAccCSMThreatsPoliciesListConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPoliciesListExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "entries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "entries.0.name", "TERRAFORM_POLICY1"),
					resource.TestCheckResourceAttr(resourceName, "entries.0.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "entries.1.name", "TERRAFORM_POLICY2 UPDATED"),
					resource.TestCheckResourceAttr(resourceName, "entries.1.priority", "5"),
				),
			},
		},
	})
}

func testAccCheckCSMThreatsPoliciesListExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in state", resourceName)
		}
		if rs.Type != "datadog_csm_threats_policies_list" {
			return fmt.Errorf(
				"resource %s is not a datadog_csm_threats_policies_list, got: %s",
				resourceName,
				rs.Type,
			)
		}

		if rs.Primary.ID != "policies_list" {
			return fmt.Errorf("expected resource ID to be 'policies_list', got %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCSMThreatsPoliciesListDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_csm_threats_policies_list" {
				continue
			}

			resp, httpResponse, err := apiInstances.GetCSMThreatsApiV2().ListCSMThreatsAgentPolicies(auth)
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode == 404 {
					return nil
				}
				return fmt.Errorf("Received an error while listing the policies: %s", err)
			}

			if len(resp.GetData()) > 1 { // CWS_DD is always present
				return fmt.Errorf("Policies list not empty, some policies are still present")
			}
		}
		return nil
	}
}

func testAccCSMThreatsPoliciesListConfigBasic() string {
	return `
		resource "datadog_csm_threats_policy" "policy1" {
			description = "created with terraform"
			enabled     = false
			tags        = []
		}

		resource "datadog_csm_threats_policy" "policy2" {
			description = "created with terraform 2"
			enabled     = true
			tags        = ["env:staging"]
		}

		resource "datadog_csm_threats_policies_list" "all" {
			entries {
				policy_id = datadog_csm_threats_policy.policy1.id
				name      = "TERRAFORM_POLICY1"
				priority  = 2
			}
			entries {
				policy_id = datadog_csm_threats_policy.policy2.id
				name      = "TERRAFORM_POLICY2"
				priority  = 3
			}
		}
	`
}

func testAccCSMThreatsPoliciesListConfigUpdate() string {
	return `
		resource "datadog_csm_threats_policy" "policy1" {
			description = "created with terraform"
			enabled     = false
			tags        = []
		}

		resource "datadog_csm_threats_policy" "policy2" {
			description = "created with terraform 2"
			enabled     = true
			tags        = ["env:staging"]
		}

		resource "datadog_csm_threats_policies_list" "all" {
			entries {
				policy_id = datadog_csm_threats_policy.policy1.id
				name      = "TERRAFORM_POLICY1"
				priority  = 2
			}
			entries {
				policy_id = datadog_csm_threats_policy.policy2.id
				name      = "TERRAFORM_POLICY2 UPDATED"
				priority  = 5
			}
		}
	`
}
