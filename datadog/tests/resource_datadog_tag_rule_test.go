package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// regexpCannotCreateBlocking matches the diagnostic the provider raises when a
// config asks for a blocking rule at create time without the opt-in flag.
var regexpCannotCreateBlocking = regexp.MustCompile("Cannot create a blocking tag rule directly")

func TestAccDatadogTagRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagRuleConfigSurfacing(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "name", ruleName),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "source", "logs"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "scope", ruleName),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_key", "env"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "rule_type", "surfacing"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_value_patterns.#", "2"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_value_patterns.0", "prod"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_value_patterns.1", "staging"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "negated", "false"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "required", "false"),
					// hard_delete is provider-only and defaults to true.
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "hard_delete", "true"),
					resource.TestCheckResourceAttrSet("datadog_tag_rule.foo", "id"),
					resource.TestCheckResourceAttrSet("datadog_tag_rule.foo", "version"),
					resource.TestCheckResourceAttrSet("datadog_tag_rule.foo", "created_at"),
					resource.TestCheckResourceAttrSet("datadog_tag_rule.foo", "created_by"),
					resource.TestCheckResourceAttrSet("datadog_tag_rule.foo", "modified_at"),
					resource.TestCheckResourceAttrSet("datadog_tag_rule.foo", "modified_by"),
				),
			},
		},
	})
}

func TestAccDatadogTagRule_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagRuleConfigSurfacing(ruleName),
			},
			{
				ResourceName:      "datadog_tag_rule.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// Provider-only fields have no API representation, so an imported
				// resource cannot reproduce them.
				ImportStateVerifyIgnore: []string{"force_blocking_on_create", "hard_delete"},
			},
		},
	})
}

// TestAccDatadogTagRule_Updated exercises in-place updates. It deliberately stays on
// rule_type = surfacing: promoting a rule to blocking needs elevated permissions the
// integration-test org's keys do not carry (the API returns 403 permission denied), so
// a blocking transition cannot be recorded here.
func TestAccDatadogTagRule_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-updated-%d", clockFromContext(ctx).Now().Unix())
	updatedName := ruleName + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagRuleConfigSurfacing(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "name", ruleName),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "rule_type", "surfacing"),
				),
			},
			{
				Config: testAccCheckDatadogTagRuleConfigUpdated(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "name", updatedName),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "rule_type", "surfacing"),
					// A successful PATCH bumps the server-side version counter.
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "version", "2"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "scope", updatedName+"-alt"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_key", "service"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_value_patterns.#", "1"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "tag_value_patterns.0", "web-*"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "enabled", "false"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "negated", "true"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "required", "true"),
					// source is unchanged, so the rule must not have been replaced.
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "source", "logs"),
				),
			},
		},
	})
}

// TestAccDatadogTagRule_BlockingOnCreate covers the non-atomic create path: the provider
// POSTs the rule as surfacing and then PATCHes it to blocking. Skipped -- see the note in
// the function body.
func TestAccDatadogTagRule_BlockingOnCreate(t *testing.T) {
	// The create POST succeeds as surfacing, but the promoting PATCH returns
	// 403 permission denied with the integration-test org's keys: setting
	// rule_type = blocking requires an elevated governance permission. Recording this
	// path needs credentials that carry it, so it stays unverified for now.
	t.Skip("promoting a tag rule to blocking returns 403 permission denied with test-org credentials")

	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-blocking-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagRuleConfigBlocking(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "rule_type", "blocking"),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "force_blocking_on_create", "true"),
					// The promoting PATCH bumps the server-side version counter past 1.
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "version", "2"),
				),
			},
		},
	})
}

// TestAccDatadogTagRule_BlockingOnCreateWithoutOptIn asserts that asking for a
// blocking rule at create time fails the plan/apply unless the opt-in is set.
func TestAccDatadogTagRule_BlockingOnCreateWithoutOptIn(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-nooptin-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogTagRuleConfigBlockingNoOptIn(ruleName),
				ExpectError: regexpCannotCreateBlocking,
			},
		},
	})
}

// TestAccDatadogTagRule_SourceForcesReplacement asserts that changing source,
// the one attribute absent from the update payload, replaces the resource.
func TestAccDatadogTagRule_SourceForcesReplacement(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-replace-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagRuleConfigSurfacing(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "source", "logs"),
				),
			},
			{
				Config: testAccCheckDatadogTagRuleConfigSourceChanged(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "source", "spans"),
					// A replacement resets the version counter for the new rule.
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "version", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogTagRuleConfigSurfacing(name string) string {
	return fmt.Sprintf(`
resource "datadog_tag_rule" "foo" {
  name               = "%[1]s"
  source             = "logs"
  scope              = "%[1]s"
  tag_key            = "env"
  tag_value_patterns = ["prod", "staging"]
  rule_type          = "surfacing"
}`, name)
}

func testAccCheckDatadogTagRuleConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_tag_rule" "foo" {
  name               = "%[1]s"
  source             = "logs"
  scope              = "%[1]s-alt"
  tag_key            = "service"
  tag_value_patterns = ["web-*"]
  rule_type          = "surfacing"
  enabled            = false
  negated            = true
  required           = true
}`, name)
}

func testAccCheckDatadogTagRuleConfigBlocking(name string) string {
	return fmt.Sprintf(`
resource "datadog_tag_rule" "foo" {
  name               = "%[1]s"
  source             = "logs"
  scope              = "%[1]s"
  tag_key            = "env"
  tag_value_patterns = ["prod", "staging"]
  rule_type          = "blocking"

  force_blocking_on_create = true
}`, name)
}

func testAccCheckDatadogTagRuleConfigBlockingNoOptIn(name string) string {
	return fmt.Sprintf(`
resource "datadog_tag_rule" "foo" {
  name               = "%[1]s"
  source             = "logs"
  scope              = "%[1]s"
  tag_key            = "env"
  tag_value_patterns = ["prod", "staging"]
  rule_type          = "blocking"
}`, name)
}

func testAccCheckDatadogTagRuleConfigSourceChanged(name string) string {
	return fmt.Sprintf(`
resource "datadog_tag_rule" "foo" {
  name               = "%[1]s"
  source             = "spans"
  scope              = "%[1]s"
  tag_key            = "env"
  tag_value_patterns = ["prod", "staging"]
  rule_type          = "surfacing"
}`, name)
}

func testAccCheckDatadogTagRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return tagRuleExistsHelper(auth, s, apiInstances)
	}
}

func testAccCheckDatadogTagRuleDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return tagRuleDestroyHelper(auth, s, apiInstances)
	}
}

func tagRuleExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetTagRulesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_tag_rule" {
			continue
		}

		if _, httpResp, err := api.GetTagRule(ctx, r.Primary.ID); err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving tag rule")
		}
	}
	return nil
}

// tagRuleDestroyHelper assumes the default hard_delete = true, under which a
// destroyed rule is removed permanently and the API returns 404.
func tagRuleDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetTagRulesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_tag_rule" {
			continue
		}

		_, httpResp, err := api.GetTagRule(ctx, r.Primary.ID)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving tag rule")
		}
		return fmt.Errorf("tag rule %s still exists", r.Primary.ID)
	}
	return nil
}
