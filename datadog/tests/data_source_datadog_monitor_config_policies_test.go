package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogMonitorConfigPoliciesDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	datasource := "data.datadog_monitor_config_policies.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckDatadogMonitorConfigPolicyDestroy(accProvider),
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMonitorConfigPolicies(),
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

func testAccDatasourceMonitorConfigPolicies() string {
	return fmt.Sprintf(`
	data "datadog_monitor_config_policies" "test" {}
    %s 
    %s`,
		testAccCheckDatadogMonitorConfigPolicyConfig("test1", "tagKey1"),
		testAccCheckDatadogMonitorConfigPolicyConfig("test2", "tagKey2"))
}
