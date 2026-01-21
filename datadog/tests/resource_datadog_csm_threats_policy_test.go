package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Create an agent policy and update its description
func TestAccCSMThreatsPolicy_CreateAndUpdate(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	policyName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_policy.policy_test"
	tags := []string{"host_name:test_host"}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "datadog_csm_threats_policy" "policy_test" {
					name              = "%s"
					enabled           = true
					description       = "im a policy"
					tags              = ["host_name:test_host"]
				}
				`, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_test"),
					checkCSMThreatsPolicyContent(
						resourceName,
						policyName,
						"im a policy",
						tags,
					),
				),
			},
			// Update description
			{
				Config: fmt.Sprintf(`
				resource "datadog_csm_threats_policy" "policy_test" {
					name              = "%s"
					enabled           = true
					description       = "updated policy for terraform provider test"
					tags              = ["host_name:test_host"]
				}
				`, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsPolicyContent(
						resourceName,
						policyName,
						"updated policy for terraform provider test",
						tags,
					),
				),
			},
		},
	})
}

// Create an agent policy with host_tags_lists and update its description
func TestAccCSMThreatsPolicy_CreateAndUpdateWithHostTagsLists(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	policyName := uniqueAgentRuleName(ctx)
	resourceName := "datadog_csm_threats_policy.policy_test"
	hostTagsLists := [][]string{
		{"host_name:test_host", "env:prod"},
		{"host_name:test_host2", "env:staging"},
	}
	updatedHostTagsLists := [][]string{
		{"host_name:test", "env:prod"},
		{"host_name:test_host2", "env:test"},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "datadog_csm_threats_policy" "policy_test" {
					name              = "%s"
					enabled           = true
					description       = "im a policy"
					host_tags_lists   = [["host_name:test_host", "env:prod"], ["host_name:test_host2", "env:staging"]]
				}
				`, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_test"),
					checkCSMThreatsPolicyContentWithHostTagsLists(
						resourceName,
						policyName,
						"im a policy",
						hostTagsLists,
					),
				),
			},
			// Update description and tags
			{
				Config: fmt.Sprintf(`
				resource "datadog_csm_threats_policy" "policy_test" {
					name              = "%s"
					enabled           = true
					description       = "updated policy for terraform provider test"
					host_tags_lists   = [["host_name:test", "env:prod"], ["host_name:test_host2", "env:test"]]
				}
				`, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, resourceName),
					checkCSMThreatsPolicyContentWithHostTagsLists(
						resourceName,
						policyName,
						"updated policy for terraform provider test",
						updatedHostTagsLists,
					),
				),
			},
		},
	})
}

func checkCSMThreatsPolicyContent(resourceName string, name string, description string, tags []string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
		resource.TestCheckResourceAttr(resourceName, "tags.0", tags[0]),
	)
}

func checkCSMThreatsPolicyContentWithHostTagsLists(resourceName string, name string, description string, hostTagsLists [][]string) resource.TestCheckFunc {
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", description),
		resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
	}

	// Add checks for each host tags list
	for i, tagList := range hostTagsLists {
		for j, tag := range tagList {
			checks = append(checks, resource.TestCheckResourceAttr(
				resourceName,
				fmt.Sprintf("host_tags_lists.%d.%d", i, j),
				tag,
			))
		}
	}

	return resource.ComposeTestCheckFunc(checks...)
}

func testAccCheckCSMThreatsPolicyExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state %s", resourceName, s.RootModule().Resources)
		}

		if resource.Type != "datadog_csm_threats_policy" {
			return fmt.Errorf("resource %s is not of type datadog_csm_threats_policy, found %s instead", resourceName, resource.Type)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		_, _, err := apiInstances.GetCSMThreatsApiV2().GetCSMThreatsAgentPolicy(auth, resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving policy: %s", err)
		}

		return nil
	}
}

func testAccCheckCSMThreatsPolicyDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_csm_threats_policy" {
				_, httpResponse, err := apiInstances.GetCSMThreatsApiV2().GetCSMThreatsAgentPolicy(auth, resource.Primary.ID)
				if err == nil {
					return errors.New("policy still exists")
				}
				if httpResponse == nil || httpResponse.StatusCode != 404 {
					return fmt.Errorf("received an error while getting the policy: %s", err)
				}
			}
		}

		return nil
	}
}
