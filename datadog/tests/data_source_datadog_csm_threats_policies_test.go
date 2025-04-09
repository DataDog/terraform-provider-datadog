package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccCSMThreatsPoliciesDataSource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	policyName := uniqueAgentRuleName(ctx)
	dataSourceName := "data.datadog_csm_threats_policies.my_data_source"
	policyConfig := fmt.Sprintf(`
	resource "datadog_csm_threats_policy" "policy_for_data_source_test" {
		name              = "%s"
		enabled           = true
		description       = "im a policy"
		tags              = ["host_name:test_host"]
	}
	`, policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckCSMThreatsPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Create a policy to have at least one
				Config: policyConfig,
				Check:  testAccCheckCSMThreatsPolicyExists(providers.frameworkProvider, "datadog_csm_threats_policy.policy_for_data_source_test"),
			},
			{
				Config: fmt.Sprintf(`
				%s
				data "datadog_csm_threats_policies" "my_data_source" {}
				`, policyConfig),
				Check: checkCSMThreatsPoliciesDataSourceContent(providers.frameworkProvider, dataSourceName, policyName),
			},
		},
	})
}

func checkCSMThreatsPoliciesDataSourceContent(accProvider *fwprovider.FrameworkProvider, dataSourceName string, policyName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("resource missing from state: %s", dataSourceName)
		}

		auth := accProvider.Auth
		apiInstances := accProvider.DatadogApiInstances

		allPoliciesResponse, _, err := apiInstances.GetCSMThreatsApiV2().ListCSMThreatsAgentPolicies(auth)
		if err != nil {
			return err
		}

		// Check the policy we created is in the API response
		resPolicyId := ""
		for _, policy := range allPoliciesResponse.GetData() {
			if policy.Attributes.GetName() == policyName {
				resPolicyId = policy.GetId()
				break
			}
		}
		if resPolicyId == "" {
			return fmt.Errorf("policy with name '%s' not found in API responses", policyName)
		}

		// Check that the data_source fetched is correct
		resourceAttributes := res.Primary.Attributes
		policyIdsCount, err := strconv.Atoi(resourceAttributes["policy_ids.#"])
		if err != nil {
			return err
		}
		policiesCount, err := strconv.Atoi(resourceAttributes["policies.#"])
		if err != nil {
			return err
		}
		if policiesCount != policyIdsCount {
			return fmt.Errorf("the data source contains %d policy IDs but %d policies", policyIdsCount, policiesCount)
		}

		// Find in which position is the policy we created, and check its values
		idx := 0
		for idx < policyIdsCount && resourceAttributes[fmt.Sprintf("policy_ids.%d", idx)] != resPolicyId {
			idx++
		}
		if idx == len(resourceAttributes) {
			return fmt.Errorf("policy with ID '%s' not found in data source", resPolicyId)
		}

		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("policies.%d.name", idx), policyName),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("policies.%d.enabled", idx), "true"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("policies.%d.tags.0", idx), "host_name:test_host"),
			resource.TestCheckResourceAttr(dataSourceName, fmt.Sprintf("policies.%d.description", idx), "im a policy"),
		)(state)
	}
}
