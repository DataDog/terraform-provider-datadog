package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttr(datasource, "monitor_config_policies.#", "2"),
					resource.TestCheckResourceAttrSet(datasource, "monitor_config_policies.0.id"),
					resource.TestCheckResourceAttr(datasource, "monitor_config_policies.0.tag_policy.0.tag_key", "tagKey1"),
					resource.TestCheckResourceAttr(datasource, "monitor_config_policies.0.tag_policy.0.tag_key_required", "false"),
					resource.TestCheckResourceAttr(datasource, "monitor_config_policies.0.tag_policy.0.valid_tag_values.#", "1"),
					resource.TestCheckResourceAttr(datasource, "monitor_config_policies.0.tag_policy.0.valid_tag_values.0", "value"),

					resource.TestCheckResourceAttrSet(datasource, "monitor_config_policies.1.id"),
					resource.TestCheckResourceAttr(datasource, "monitor_config_policies.1.tag_policy.0.tag_key", "tagKey2"),
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
			tag_key          = "tagKey1"
			tag_key_required = false
			valid_tag_values = ["value"]
		}
	}
    resource "datadog_monitor_config_policy" "%[1]s-2" {
		depends_on   = ["datadog_monitor_config_policy.%[1]s-1"]
		policy_type = "tag"
		tag_policy {
			tag_key          = "tagKey2"
			tag_key_required = false
			valid_tag_values = ["value"]
		}
	}`, name)
}
