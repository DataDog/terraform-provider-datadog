package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

const tfCloudConfRuleName = "datadog_cloud_configuration_rule.acceptance_test"

func TestAccDatadogCloudConfigurationRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationCreatedCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogCloudConfigurationUpdatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationUpdatedCheck(accProvider, ruleName),
			},
			{
				Config: testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogCloudConfigurationRule_DefaultTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },

		// ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{ // Works with Configuration without a tag attribute
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Config: testAccCheckDatadogCloudConfigurationNoTagsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationNoTagsCheck(accProvider, ruleName),
			},
			{ // New tags are correctly added and duplicates are kept
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Config: testAccCheckDatadogCloudConfigurationDuplicateTagsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationDuplicateTagsCheck(accProvider, ruleName),
			},
			{ // Resource tags take precedence over default tags and duplicates stay
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"foo": "not_bar",
					}),
				},
				Config: testAccCheckDatadogCloudConfigurationDuplicateTagsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationDuplicateTagsPrecedenceCheck(accProvider, ruleName),
			},
			{ // Resource tags take precedence over default tags, but new tags are added
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"foo":  "bar",
						"test": "not_acceptance",
					}),
				},
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationDefaultTagsCheck(accProvider, ruleName),
			},
			{ // Tags without any value work correctly
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"no_value": "",
					}),
				},
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationNoValueTagsCheck(accProvider, ruleName),
			},
			{ // Tags with colons in the value work correctly
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"repo_url": "https://github.com/repo/path",
					}),
				},
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationColonTagsCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogCloudConfigurationRule_MandatoryFieldsOnly(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsConfig(ruleName),
				Check:  testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsCheck(accProvider, ruleName),
			},
		},
	})
}

func TestAccDatadogCloudConfigurationRule_Import(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCloudConfigurationCreatedConfig(ruleName),
			},
			{
				ResourceName:      tfCloudConfRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				Check:             testAccCheckDatadogCloudConfigurationCreatedCheck(accProvider, ruleName),
			},
		},
	})
}

func testAccCheckDatadogCloudConfigurationCreatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule"
  name    = "%s"
  notifications = [ "@channel" ]
  group_by = [ "@resource" ]
  policy = "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"
  resource_type = "gcp_compute_instance"
  related_resource_types = [ "gcp_compute_disk" ]
  severity = "low"
  tags = [
    "terraform:true",
    "test:acceptance",
  ]
  filter {
    action = "suppress"
    query = "resource_id:hel*"
  }
  filter {
    action = "require"
    query = "resource_type:hel*"
  }
}
`, name)
}

func testAccCheckDatadogCloudConfigurationCreatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "2"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "terraform:true"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "test:acceptance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.query", "resource_id:hel*"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.action", "require"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.query", "resource_type:hel*"),
	)
}

func testAccCheckDatadogCloudConfigurationUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = true
  message = "Acceptance test TF rule - updated"
  name    = "%s - updated"
  notifications = [ "@channel-upd" ]
  group_by = [ "@resource", "@resource_type" ]
  policy = "package datadog # updated\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"
  resource_type = "gcp_compute_disk"
  related_resource_types = [ "gcp_compute_instance", "gcp_compute_firewall" ]
  severity = "high"
  tags = [ "test:acceptance-updated" ]
  filter {
    action = "suppress"
    query = "resource_id:updated*"
  }
}
`, name)
}

func testAccCheckDatadogCloudConfigurationUpdatedCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "true"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule - updated"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName+" - updated"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel-upd"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.1", "@resource_type"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog # updated\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.1", "gcp_compute_firewall"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "high"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "1"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "test:acceptance-updated"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.query", "resource_id:updated*"),
	)
}

func testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule - updated again"
  name    = "%s - updated again"
  policy = "package datadog # updated again\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"
  resource_type = "gcp_compute_instance"
  severity = "medium"
  tags = [ "test:acceptance-updated-again" ]
}
`, name)
}

func testAccCheckDatadogCloudConfigurationUpdatedMandatoryFieldsCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule - updated again"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName+" - updated again"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications.0"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog # updated again\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "medium"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "1"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "test:acceptance-updated-again"),
	)
}

func testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule"
  name    = "%s"
  policy  = "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"
  resource_type = "gcp_compute_instance"
  severity = "low"
}
`, name)
}

func testAccCheckDatadogCloudConfigurationCreatedMandatoryFieldsCheck(accProvider func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccCheckDatadogCloudConfRuleExists(accProvider),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.#", "0"),
	)
}

func testAccCheckDatadogCloudConfRuleExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, rule := range s.RootModule().Resources {
			_, _, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, rule.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving security monitoring rule: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogCloudConfigurationRuleDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_cloud_configuration_rule" {
				_, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, resource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == 404 {
						continue
					}
					return fmt.Errorf("received an error deleting cloud configuration rule: %s", err)
				}
				return fmt.Errorf("cloud configuration rule still exists")
			}
		}
		return nil
	}
}

func testAccCheckDatadogCloudConfigurationDuplicateTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule"
  name    = "%s"
  notifications = [ ]
  group_by = [ ]
  policy = "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"
  resource_type = "gcp_compute_instance"
  related_resource_types = [ ]
  severity = "low"
  tags = ["foo:bar", "baz", "foo:thebar"]
  filter {
    action = "suppress"
    query = "resource_id:hel*"
  }
  filter {
    action = "require"
    query = "resource_type:hel*"
  }
}
`, name)
}

func testAccCheckDatadogCloudConfigurationDuplicateTagsPrecedenceCheck(_ func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications.0"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "3"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "baz"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "foo:bar"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "foo:thebar"),
	)
}

func testAccCheckDatadogCloudConfigurationDuplicateTagsCheck(_ func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications.0"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "4"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "baz"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "foo:bar"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "foo:thebar"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "default_key:default_value"),
	)
}

func testAccCheckDatadogCloudConfigurationNoTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_cloud_configuration_rule" "acceptance_test" {
  enabled = false
  message = "Acceptance test TF rule"
  name    = "%s"
  notifications = [ ]
  group_by = [ ]
  policy = "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"
  resource_type = "gcp_compute_instance"
  related_resource_types = [ ]
  severity = "low"
  filter {
    action = "suppress"
    query = "resource_id:hel*"
  }
  filter {
    action = "require"
    query = "resource_type:hel*"
  }
}
`, name)
}

func testAccCheckDatadogCloudConfigurationNoTagsCheck(_ func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "notifications.0"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "group_by.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckNoResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "1"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "default_key:default_value"),
	)
}

func testAccCheckDatadogCloudConfigurationColonTagsCheck(_ func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "3"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "terraform:true"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "test:acceptance"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "repo_url:https://github.com/repo/path"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.query", "resource_id:hel*"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.action", "require"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.query", "resource_type:hel*"),
	)
}

func testAccCheckDatadogCloudConfigurationNoValueTagsCheck(_ func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "3"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "terraform:true"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "test:acceptance"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "no_value"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.query", "resource_id:hel*"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.action", "require"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.query", "resource_type:hel*"),
	)
}

func testAccCheckDatadogCloudConfigurationDefaultTagsCheck(_ func() (*schema.Provider, error), ruleName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "enabled", "false"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "message", "Acceptance test TF rule"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "name", ruleName),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "notifications.0", "@channel"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "group_by.0", "@resource"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "policy", "package datadog\n\nimport data.datadog.output as dd_output\n\nimport future.keywords.contains\nimport future.keywords.if\nimport future.keywords.in\n\nmilliseconds_in_a_day := ((1000 * 60) * 60) * 24\n\neval(iam_service_account_key) = \"skip\" if {\n\tiam_service_account_key.disabled\n} else = \"pass\" if {\n\t(iam_service_account_key.resource_seen_at / milliseconds_in_a_day) - (iam_service_account_key.valid_after_time / milliseconds_in_a_day) <= 90\n} else = \"fail\"\n\n# This part remains unchanged for all rules\nresults contains result if {\n\tsome resource in input.resources[input.main_resource_type]\n\tresult := dd_output.format(resource, eval(resource))\n}\n"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "resource_type", "gcp_compute_instance"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "related_resource_types.0", "gcp_compute_disk"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "severity", "low"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "tags.#", "3"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "terraform:true"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "test:acceptance"),
		resource.TestCheckTypeSetElemAttr(
			tfCloudConfRuleName, "tags.*", "foo:bar"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.action", "suppress"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.0.query", "resource_id:hel*"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.action", "require"),
		resource.TestCheckResourceAttr(
			tfCloudConfRuleName, "filter.1.query", "resource_type:hel*"),
	)
}
