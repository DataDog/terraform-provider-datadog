package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogMonitorConfigPoliciesDatasource(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	testName := uniqueEntityName(ctx, t)
	datasource := "data.datadog_monitor_config_policies.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMonitorConfigPolicies(testName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasource, "id", "monitor-config-policies"),
					testAccCheckDatadogMonitorConfigPolicyResourceExists(testName+"-1", "false", "value"),
					testAccCheckDatadogMonitorConfigPolicyResourceExists(testName+"-2", "false", "value"),
				),
			},
		},
	})
}

func testAccDatasourceMonitorConfigPolicies(name string) string {
	return fmt.Sprintf(`
	data "datadog_monitor_config_policies" "test" {
		depends_on   = ["datadog_monitor_config_policy.%[1]s-2"]
	} 
    resource "datadog_monitor_config_policy" "%[1]s-1" {
		policy_type = "tag"
		tag_policy {
			tag_key          = "%[1]s-1"
			tag_key_required = false
			valid_tag_values = ["value"]
		}
	}
    resource "datadog_monitor_config_policy" "%[1]s-2" {
		depends_on   = ["datadog_monitor_config_policy.%[1]s-1"]
		policy_type = "tag"
		tag_policy {
			tag_key          = "%[1]s-2"
			tag_key_required = false
			valid_tag_values = ["value"]
		}
	}`, name)
}

func testAccCheckDatadogMonitorConfigPolicyResourceExists(tag_key string, tag_key_required string, tag_value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_monitor_config_policies" {
				continue
			}
			policy_num, _ := strconv.Atoi(r.Primary.Attributes["monitor_config_policies.#"])
			for i := 0; i < policy_num; i++ {
				tag_key_state := r.Primary.Attributes[fmt.Sprintf("monitor_config_policies.%d.tag_policy.0.tag_key", i)]
				tag_key_required_state := r.Primary.Attributes[fmt.Sprintf("monitor_config_policies.%d.tag_policy.0.tag_key_required", i)]
				tag_value_state := r.Primary.Attributes[fmt.Sprintf("monitor_config_policies.%d.tag_policy.0.valid_tag_values.0", i)]
				if tag_key_state == tag_key && tag_key_required_state == tag_key_required && tag_value_state == tag_value {
					return nil
				}
			}
		}
		return fmt.Errorf("missing monitor config policy")
	}
}
