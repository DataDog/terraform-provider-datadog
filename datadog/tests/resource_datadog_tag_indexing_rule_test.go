package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// metricSafeUniq converts uniqueEntityName output to a string safe for use in
// metric name patterns: metric names and glob patterns only allow alphanumeric
// characters, dots, underscores, and asterisks (no hyphens).
func metricSafeUniq(uniq string) string {
	return strings.ReplaceAll(uniq, "-", ".")
}

func TestAccDatadogTagIndexingRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	mUniq := metricSafeUniq(uniq)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagIndexingRuleDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagIndexingRuleConfigBasic(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "name", fmt.Sprintf("tf-test-tag-indexing-rule-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "metric_name_matches.#", "1"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_indexing_rule.foo", "metric_name_matches.*", fmt.Sprintf("tf.test.%s.*", mUniq)),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "exclude_tags_mode", "false"),
					resource.TestCheckResourceAttrSet("datadog_tag_indexing_rule.foo", "id"),
					resource.TestCheckResourceAttrSet("datadog_tag_indexing_rule.foo", "rule_order"),
					resource.TestCheckResourceAttrSet("datadog_tag_indexing_rule.foo", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_tag_indexing_rule.foo", "modified_at"),
				),
			},
			{
				ResourceName:            "datadog_tag_indexing_rule.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				// modified_at advances between create and import read (server-side clock skew).
				ImportStateVerifyIgnore: []string{"modified_at"},
			},
		},
	})
}

func TestAccDatadogTagIndexingRule_WithOptions(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	mUniq := metricSafeUniq(uniq)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagIndexingRuleDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagIndexingRuleConfigWithOptions(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.bar", "name", fmt.Sprintf("tf-test-tag-indexing-options-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.bar", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_indexing_rule.bar", "tags.*", "env"),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.bar", "exclude_tags_mode", "true"),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.bar", "options.version", "1"),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.bar", "options.data.manage_preexisting_metrics", "true"),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.bar", "options.data.override_previous_rules", "false"),
				),
			},
			{
				ResourceName:            "datadog_tag_indexing_rule.bar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"modified_at"},
			},
		},
	})
}

func TestAccDatadogTagIndexingRule_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	mUniq := metricSafeUniq(uniq)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagIndexingRuleDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagIndexingRuleConfigBasic(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "name", fmt.Sprintf("tf-test-tag-indexing-rule-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "tags.#", "0"),
				),
			},
			{
				Config: testAccCheckDatadogTagIndexingRuleConfigUpdated(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "name", fmt.Sprintf("tf-test-tag-indexing-rule-updated-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_indexing_rule.foo", "tags.*", "env"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_indexing_rule.foo", "tags.*", "service"),
				),
			},
		},
	})
}

func testAccCheckDatadogTagIndexingRuleDestroy(ctx context.Context, frameworkProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := frameworkProvider.DatadogApiInstances
		auth := frameworkProvider.Auth
		return datadogTagIndexingRuleDestroyHelper(ctx, auth, s, apiInstances)
	}
}

func datadogTagIndexingRuleDestroyHelper(_ context.Context, auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetMetricsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_tag_indexing_rule" {
			continue
		}
		_, httpResp, err := api.GetTagIndexingRule(auth, r.Primary.ID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving tag indexing rule: %s", err.Error())
		}
		return fmt.Errorf("tag indexing rule still exists")
	}
	return nil
}

func testAccCheckDatadogTagIndexingRuleConfigBasic(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "foo" {
  name                = "tf-test-tag-indexing-rule-%s"
  metric_name_matches = ["tf.test.%s.*"]
  tags                = []
  exclude_tags_mode   = false
}`, uniq, mUniq)
}

func testAccCheckDatadogTagIndexingRuleConfigUpdated(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "foo" {
  name                = "tf-test-tag-indexing-rule-updated-%s"
  metric_name_matches = ["tf.test.%s.*"]
  tags                = ["env", "service"]
  exclude_tags_mode   = false
}`, uniq, mUniq)
}

func testAccCheckDatadogTagIndexingRuleConfigWithOptions(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "bar" {
  name                = "tf-test-tag-indexing-options-%s"
  metric_name_matches = ["tf.test.options.%s.*"]
  tags                = ["env"]
  exclude_tags_mode   = true

  options = {
    version = 1
    data = {
      manage_preexisting_metrics = true
      override_previous_rules    = false
    }
  }
}`, uniq, mUniq)
}
