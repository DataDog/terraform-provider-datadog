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
func TestAccCSMThreatsPolicies_CreateAndUpdate(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_csm_threats_policies.all_policies"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsPoliciesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCSMThreatsPoliciesConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPoliciesExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policies.0.name", "terraform_policy"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.description", "description"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.tags.0", "env:staging"),
				),
			},
			{
				Config: testAccCSMThreatsPoliciesConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPoliciesExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policies.0.name", "terraform_policy updated"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.description", "new description"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "policies.0.tags.0", "foo:bar"),
				),
			},
		},
	})
}

func testAccCheckCSMThreatsPoliciesExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in state", resourceName)
		}
		if rs.Type != "datadog_csm_threats_policies" {
			return fmt.Errorf(
				"resource %s is not a datadog_csm_threats_policies, got: %s",
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

func testAccCheckCSMThreatsPoliciesDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_csm_threats_policies" {
				continue
			}

			if _, ok := s.RootModule().Resources[r.Primary.ID]; ok {
				return fmt.Errorf("Resource %s still exists in state", r.Primary.ID)
			}
		}
		return nil
	}
}

func testAccCSMThreatsPoliciesConfig() string {
	return `
		resource "datadog_csm_threats_policies" "all_policies" {
			policies {
				policy_label = "policy"
				name         = "terraform_policy"
				description  = "description"
				enabled      = false
				tags         = ["env:staging"]
			}
		}
	`
}

func testAccCSMThreatsPoliciesConfigUpdate() string {
	return `
		resource "datadog_csm_threats_policies" "all_policies" {
			policies {
				policy_label = "policy"
				name         = "terraform_policy updated"
				description  = "new description"
				enabled      = true
				tags         = ["foo:bar"]
			}
		}
	`
}
