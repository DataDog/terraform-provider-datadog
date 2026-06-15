package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogTagIndexingRuleOrder_Basic(t *testing.T) {
	skipIfNoCassette(t)
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	_ = providers
	uniq := uniqueEntityName(ctx, t)
	mUniq := metricSafeUniq(uniq)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagIndexingRuleOrderConfig(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule_order.order", "name", fmt.Sprintf("tf-test-order-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule_order.order", "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair("datadog_tag_indexing_rule_order.order", "rule_ids.0", "datadog_tag_indexing_rule.first", "id"),
					resource.TestCheckResourceAttrPair("datadog_tag_indexing_rule_order.order", "rule_ids.1", "datadog_tag_indexing_rule.second", "id"),
				),
			},
			{
				Config: testAccCheckDatadogTagIndexingRuleOrderConfigSwapped(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule_order.order", "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair("datadog_tag_indexing_rule_order.order", "rule_ids.0", "datadog_tag_indexing_rule.second", "id"),
					resource.TestCheckResourceAttrPair("datadog_tag_indexing_rule_order.order", "rule_ids.1", "datadog_tag_indexing_rule.first", "id"),
				),
			},
		},
	})
}

func testAccCheckDatadogTagIndexingRuleOrderConfig(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "first" {
  name                = "tf-test-order-first-%s"
  metric_name_matches = ["tf.test.order.first.%s.*"]
  tags                = []
  exclude_tags_mode   = false
}

resource "datadog_tag_indexing_rule" "second" {
  name                = "tf-test-order-second-%s"
  metric_name_matches = ["tf.test.order.second.%s.*"]
  tags                = []
  exclude_tags_mode   = false
}

resource "datadog_tag_indexing_rule_order" "order" {
  name     = "tf-test-order-%s"
  rule_ids = [
    datadog_tag_indexing_rule.first.id,
    datadog_tag_indexing_rule.second.id,
  ]
}`, uniq, mUniq, uniq, mUniq, uniq)
}

func testAccCheckDatadogTagIndexingRuleOrderConfigSwapped(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "first" {
  name                = "tf-test-order-first-%s"
  metric_name_matches = ["tf.test.order.first.%s.*"]
  tags                = []
  exclude_tags_mode   = false
}

resource "datadog_tag_indexing_rule" "second" {
  name                = "tf-test-order-second-%s"
  metric_name_matches = ["tf.test.order.second.%s.*"]
  tags                = []
  exclude_tags_mode   = false
}

resource "datadog_tag_indexing_rule_order" "order" {
  name     = "tf-test-order-%s"
  rule_ids = [
    datadog_tag_indexing_rule.second.id,
    datadog_tag_indexing_rule.first.id,
  ]
}`, uniq, mUniq, uniq, mUniq, uniq)
}
