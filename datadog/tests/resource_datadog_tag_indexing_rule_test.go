package test

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

// skipIfNoCassette skips the test in cassette-replay mode (RECORD=false) when no
// cassette file has been recorded yet. Run `make cassettes` to record them.
func skipIfNoCassette(t *testing.T) {
	t.Helper()
	if isReplaying() {
		if _, err := os.Stat(fmt.Sprintf("cassettes/%s.yaml", t.Name())); os.IsNotExist(err) {
			t.Skipf("cassette not yet recorded; run: RECORD=true TF_ACC=1 gotestsum --packages ./datadog/tests/... -- -run %s", t.Name())
		}
	}
}

func TestAccDatadogTagIndexingRule_Basic(t *testing.T) {
	skipIfNoCassette(t)
	cleanupTagIndexingRules(t)
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
				ResourceName:      "datadog_tag_indexing_rule.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// modified_at advances between create and import read (server-side clock skew).
				// options.* appear after import because the API always returns default options
				// even when the rule was created without configuring them explicitly.
				ImportStateVerifyIgnore: []string{
					"modified_at",
					"options.%",
					"options.data.%",
					"options.data.manage_preexisting_metrics",
					"options.data.override_previous_rules",
					"options.version",
				},
			},
		},
	})
}

func TestAccDatadogTagIndexingRule_WithOptions(t *testing.T) {
	skipIfNoCassette(t)
	cleanupTagIndexingRules(t)
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
	skipIfNoCassette(t)
	cleanupTagIndexingRules(t)
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

func TestAccDatadogTagIndexingRule_ExcludeMode(t *testing.T) {
	skipIfNoCassette(t)
	cleanupTagIndexingRules(t)
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
				Config: testAccCheckDatadogTagIndexingRuleConfigExcludeUsage(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.exclude", "exclude_tags_mode", "true"),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.exclude", "options.data.dynamic_tags.exclude_not_queried_window_seconds", "604800"),
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.exclude", "options.data.dynamic_tags.exclude_not_used_in_assets", "true"),
				),
			},
			{
				ResourceName:            "datadog_tag_indexing_rule.exclude",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"modified_at"},
			},
		},
	})
}

// TestAccDatadogTagIndexingRule_ExcludeMode_UsageFieldIndividuallySet sets only
// exclude_not_used_in_assets, leaving exclude_not_queried_window_seconds null. The framework's
// automatic post-apply plan check fails the test if the null-preserving flatten in updateState were
// to normalize the unset window to a zero value ("inconsistent result after apply").
func TestAccDatadogTagIndexingRule_ExcludeMode_UsageFieldIndividuallySet(t *testing.T) {
	skipIfNoCassette(t)
	cleanupTagIndexingRules(t)
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
				Config: testAccCheckDatadogTagIndexingRuleConfigExcludeUsageBoolOnly(uniq, mUniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_indexing_rule.exclude_bool_only", "options.data.dynamic_tags.exclude_not_used_in_assets", "true"),
					resource.TestCheckNoResourceAttr("datadog_tag_indexing_rule.exclude_bool_only", "options.data.dynamic_tags.exclude_not_queried_window_seconds"),
				),
			},
		},
	})
}

// TestAccDatadogTagIndexingRule_ExcludeMode_Update locks in that buildUpdateRequest always sends
// exclude_tags_mode on update: the API rejects (400) an update touching the exclude_not_* fields
// unless exclude_tags_mode is explicit in the request body, so this step failing would signal a
// regression to a conditional SetExcludeTagsMode call.
func TestAccDatadogTagIndexingRule_ExcludeMode_Update(t *testing.T) {
	skipIfNoCassette(t)
	cleanupTagIndexingRules(t)
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
				Config: testAccCheckDatadogTagIndexingRuleConfigExcludeUsage(uniq, mUniq),
				Check:  resource.TestCheckResourceAttr("datadog_tag_indexing_rule.exclude", "options.data.dynamic_tags.exclude_not_queried_window_seconds", "604800"),
			},
			{
				Config: testAccCheckDatadogTagIndexingRuleConfigExcludeUsageUpdated(uniq, mUniq),
				Check:  resource.TestCheckResourceAttr("datadog_tag_indexing_rule.exclude", "options.data.dynamic_tags.exclude_not_queried_window_seconds", "1209600"),
			},
		},
	})
}

// TestAccDatadogTagIndexingRule_ExcludeMode_ValidateConfig proves the plan-time ValidateConfig
// block fires (with no API call) when an exclude_not_* usage field is set but exclude_tags_mode
// is left at its false default.
func TestAccDatadogTagIndexingRule_ExcludeMode_ValidateConfig(t *testing.T) {
	skipIfNoCassette(t)
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	mUniq := metricSafeUniq(uniq)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogTagIndexingRuleConfigExcludeUsageWithoutMode(uniq, mUniq),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`require exclude_tags_mode to be true`),
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

func testAccCheckDatadogTagIndexingRuleConfigExcludeUsage(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "exclude" {
  name                = "tf-test-tag-indexing-exclude-%s"
  metric_name_matches = ["tf.test.exclude.%s.*"]
  tags                = ["env"]
  exclude_tags_mode   = true

  options = {
    version = 1
    data = {
      dynamic_tags = {
        exclude_not_queried_window_seconds = 604800
        exclude_not_used_in_assets         = true
      }
    }
  }
}`, uniq, mUniq)
}

func testAccCheckDatadogTagIndexingRuleConfigExcludeUsageUpdated(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "exclude" {
  name                = "tf-test-tag-indexing-exclude-%s"
  metric_name_matches = ["tf.test.exclude.%s.*"]
  tags                = ["env"]
  exclude_tags_mode   = true

  options = {
    version = 1
    data = {
      dynamic_tags = {
        exclude_not_queried_window_seconds = 1209600
        exclude_not_used_in_assets         = true
      }
    }
  }
}`, uniq, mUniq)
}

func testAccCheckDatadogTagIndexingRuleConfigExcludeUsageBoolOnly(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "exclude_bool_only" {
  name                = "tf-test-tag-indexing-exclude-bool-only-%s"
  metric_name_matches = ["tf.test.exclude.bool.only.%s.*"]
  exclude_tags_mode   = true

  options = {
    version = 1
    data = {
      dynamic_tags = {
        exclude_not_used_in_assets = true
      }
    }
  }
}`, uniq, mUniq)
}

func testAccCheckDatadogTagIndexingRuleConfigExcludeUsageWithoutMode(uniq, mUniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_indexing_rule" "exclude_invalid" {
  name                = "tf-test-tag-indexing-exclude-invalid-%s"
  metric_name_matches = ["tf.test.exclude.invalid.%s.*"]

  options = {
    version = 1
    data = {
      dynamic_tags = {
        exclude_not_used_in_assets = true
      }
    }
  }
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
