package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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

// TestAccDatadogTagRule_Updated exercises in-place updates, including the
// surfacing -> blocking transition. That transition needs no force_blocking_on_create
// opt-in: the flag gates creation only, because only creation is non-atomic.
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
					// Promoted to blocking through Update, which needs no opt-in flag.
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "rule_type", "blocking"),
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
// POSTs the rule as surfacing and then PATCHes it to blocking.
func TestAccDatadogTagRule_BlockingOnCreate(t *testing.T) {
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

// TestAccDatadogTagRule_SoftDelete covers the hard_delete = false path: destroying the
// resource must leave the rule recoverable (deleted_at set) rather than removing it.
func TestAccDatadogTagRule_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-softdelete-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagRuleSoftDeleted(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagRuleConfigSoftDelete(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_tag_rule.foo", "hard_delete", "false"),
				),
			},
		},
	})
}

// TestAccDatadogTagRule_SoftDeleteDrift covers Read's handling of a rule that was
// soft-deleted out-of-band: the API still returns 200 with deleted_at set, so a refresh
// must treat that as gone rather than looping on the tombstone forever.
func TestAccDatadogTagRule_SoftDeleteDrift(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("tf-test-tag-rule-softdrift-%d", clockFromContext(ctx).Now().Unix())
	config := testAccCheckDatadogTagRuleConfigSurfacing(ruleName)

	// Captured in step 1's Check, read from step 2's PreConfig (no state access there).
	var capturedID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagRuleExists(providers.frameworkProvider),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["datadog_tag_rule.foo"]
						if !ok {
							return fmt.Errorf("resource not found: datadog_tag_rule.foo")
						}
						capturedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				// Soft-delete the rule out-of-band, bypassing Terraform entirely, so the
				// only way the provider can notice is Read's deleted_at check.
				PreConfig: func() {
					api := providers.frameworkProvider.DatadogApiInstances.GetTagRulesApiV2()
					optionalParams := datadogV2.NewDeleteTagRuleOptionalParameters().WithHardDelete(false)
					if _, err := api.DeleteTagRule(providers.frameworkProvider.Auth, capturedID, *optionalParams); err != nil {
						t.Fatalf("failed to soft-delete tag rule out-of-band: %s", err)
					}
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatadogTagRuleConfigSoftDelete(name string) string {
	return fmt.Sprintf(`
resource "datadog_tag_rule" "foo" {
  name               = "%[1]s"
  source             = "logs"
  scope              = "%[1]s"
  tag_key            = "env"
  tag_value_patterns = ["prod", "staging"]
  rule_type          = "surfacing"
  hard_delete        = false
}`, name)
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
  rule_type          = "blocking"
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

// testAccCheckDatadogTagRuleSoftDeleted asserts the opposite of tagRuleDestroyHelper: with
// hard_delete = false, the rule must still exist after destroy, with deleted_at set.
func testAccCheckDatadogTagRuleSoftDeleted(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		api := apiInstances.GetTagRulesApiV2()

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_tag_rule" {
				continue
			}

			resp, httpResp, err := api.GetTagRule(auth, r.Primary.ID)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving soft-deleted tag rule")
			}
			if !resp.Data.Attributes.DeletedAt.IsSet() || resp.Data.Attributes.DeletedAt.Get() == nil {
				return fmt.Errorf("tag rule %s was hard-deleted despite hard_delete = false", r.Primary.ID)
			}
		}
		return nil
	}
}
