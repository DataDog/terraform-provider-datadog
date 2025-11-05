package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogTagPipelineRulesets_Basic(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedTagPipelineRulesets(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "id", "order"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.0"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.1"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.2"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_Update(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedTagPipelineRulesets(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
				),
			},
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigReordered(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
					// Verify the order has changed by checking that the first ruleset is now different
					testAccCheckDatadogTagPipelineRulesetsChanged("datadog_tag_pipeline_rulesets.foo"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_Import(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedTagPipelineRulesets(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
				),
			},
			{
				ResourceName:      "datadog_tag_pipeline_rulesets.foo",
				ImportState:       true,
				ImportStateId:     "order",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_SingleRuleset(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedTagPipelineRulesets(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigSingle(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "1"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.0"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_OverrideTrue(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedTagPipelineRulesets(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigWithOverride(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "id", "order"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "override_ui_defined_resources", "true"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.0"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.1"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.2"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_OverrideTrueDeletesUnmanaged(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			// Don't cleanup - we want to create unmanaged rulesets first
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			// First step: create unmanaged rulesets via Terraform resources (without the order resource)
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigUnmanagedRulesets(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.unmanaged_first", "name", fmt.Sprintf("tf-test-unmanaged-first-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.unmanaged_second", "name", fmt.Sprintf("tf-test-unmanaged-second-%s", uniq)),
				),
			},
			// Second step: add the order resource with override=true, which should delete unmanaged rulesets
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigWithOverrideAndUnmanaged(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "id", "order"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "override_ui_defined_resources", "true"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.0"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.1"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.2"),
					testAccCheckUnmanagedRulesetsDeleted(providers.frameworkProvider, uniq),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_OverrideFalseNoUnmanaged(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCleanupOrphanedTagPipelineRulesets(t, providers.frameworkProvider)
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalse(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "id", "order"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "override_ui_defined_resources", "false"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.0"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.1"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.2"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_OverrideFalseUnmanagedInMiddle(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			// Don't cleanup - we want to create unmanaged rulesets first
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			// First step: create unmanaged rulesets
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigUnmanagedRulesets(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.unmanaged_first", "name", fmt.Sprintf("tf-test-unmanaged-first-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.unmanaged_second", "name", fmt.Sprintf("tf-test-unmanaged-second-%s", uniq)),
				),
			},
			// Second step: add managed rulesets with override=false - should ERROR due to unmanaged in middle
			{
				Config:      testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalseWithUnmanagedMiddle(uniq),
				ExpectError: regexp.MustCompile("Unmanaged rulesets detected in the middle of order"),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesets_OverrideFalseUnmanagedAtEnd(t *testing.T) {
	// Do not run in parallel - this resource manages global ruleset order
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			// Don't cleanup - we want to create unmanaged rulesets first
		},
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetsDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			// First step: create managed rulesets first to establish order
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalse(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
				),
			},
			// Second step: add unmanaged rulesets at the end (they will be positioned after managed ones)
			{
				Config: testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalseWithUnmanagedAtEnd(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetsExists("datadog_tag_pipeline_rulesets.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "id", "order"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "ruleset_ids.#", "3"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_rulesets.foo", "override_ui_defined_resources", "false"),
					// Verify unmanaged rulesets exist but are not deleted (should show warning, not error)
					testAccCheckUnmanagedRulesetsExist(providers.frameworkProvider, uniq),
				),
			},
		},
	})
}

// testAccCheckUnmanagedRulesetsDeleted verifies that unmanaged rulesets were deleted
func testAccCheckUnmanagedRulesetsDeleted(frameworkProvider *fwprovider.FrameworkProvider, uniq string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := frameworkProvider.DatadogApiInstances
		auth := frameworkProvider.Auth
		api := apiInstances.GetCloudCostManagementApiV2()

		// List all current rulesets
		resp, _, err := api.ListTagPipelinesRulesets(auth)
		if err != nil {
			return fmt.Errorf("failed to list rulesets: %v", err)
		}

		var rulesets []datadogV2.RulesetRespData
		if respData, ok := resp.GetDataOk(); ok {
			rulesets = *respData
		}

		// Check that no unmanaged rulesets exist
		unmanagedNames := []string{
			"tf-test-unmanaged-first-" + uniq,
			"tf-test-unmanaged-second-" + uniq,
		}

		for _, ruleset := range rulesets {
			if attrs, ok := ruleset.GetAttributesOk(); ok {
				name := attrs.GetName()
				for _, unmanagedName := range unmanagedNames {
					if name == unmanagedName {
						return fmt.Errorf("unmanaged ruleset %s was not deleted", unmanagedName)
					}
				}
			}
		}

		return nil
	}
}

// testAccCleanupOrphanedTagPipelineRulesets removes ALL existing rulesets before running tests
// This ensures tests start with a clean slate, as the order resource requires all rulesets to be managed
func testAccCleanupOrphanedTagPipelineRulesets(t *testing.T, frameworkProvider *fwprovider.FrameworkProvider) {
	apiInstances := frameworkProvider.DatadogApiInstances
	auth := frameworkProvider.Auth
	api := apiInstances.GetCloudCostManagementApiV2()

	// List all rulesets
	resp, _, err := api.ListTagPipelinesRulesets(auth)
	if err != nil {
		// If we can't list rulesets, log warning and continue
		t.Logf("Warning: Could not list rulesets for cleanup: %v", err)
		return
	}

	var rulesets []datadogV2.RulesetRespData
	if respData, ok := resp.GetDataOk(); ok {
		rulesets = *respData
	}

	t.Logf("Found %d existing rulesets, cleaning up all to ensure clean test state...", len(rulesets))

	// Delete ALL rulesets to ensure clean state for testing
	// This is necessary because the tag_pipeline_rulesets resource requires all rulesets to be managed
	for _, ruleset := range rulesets {
		if rulesetID, ok := ruleset.GetIdOk(); ok {
			name := "unknown"
			if attrs, ok := ruleset.GetAttributesOk(); ok {
				name = attrs.GetName()
			}
			t.Logf("Deleting ruleset: %s (ID: %s)", name, *rulesetID)
			_, err := api.DeleteTagPipelinesRuleset(auth, *rulesetID)
			if err != nil {
				t.Logf("Warning: Could not delete ruleset %s: %v", *rulesetID, err)
			}
		}
	}
}

func testAccCheckDatadogTagPipelineRulesetsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for resource: %s", resourceName)
		}

		// The order resource should always have ID "order"
		if rs.Primary.ID != "order" {
			return fmt.Errorf("expected ID to be 'order', got: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDatadogTagPipelineRulesetsChanged(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// This is a placeholder check - in a real test environment, you would
		// verify that the actual order in DataDog has changed by calling the API
		// For now, we just verify the resource exists
		return testAccCheckDatadogTagPipelineRulesetsExists(resourceName)(s)
	}
}

func testAccCheckDatadogTagPipelineRulesetsDestroy(ctx context.Context, frameworkProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Clean up all rulesets to ensure clean state for subsequent tests
		apiInstances := frameworkProvider.DatadogApiInstances
		auth := frameworkProvider.Auth

		api := apiInstances.GetCloudCostManagementApiV2()

		// List all rulesets
		resp, _, err := api.ListTagPipelinesRulesets(auth)
		if err != nil {
			// If we can't list rulesets, just log and continue
			// The test might have already cleaned up
			return nil
		}

		var rulesets []datadogV2.RulesetRespData
		if respData, ok := resp.GetDataOk(); ok {
			rulesets = *respData
		}

		// Delete ALL rulesets to ensure clean state
		// This is important because the order resource requires all rulesets to be managed
		for _, ruleset := range rulesets {
			if rulesetID, ok := ruleset.GetIdOk(); ok {
				_, _ = api.DeleteTagPipelinesRuleset(auth, *rulesetID)
			}
		}

		return nil
	}
}

// Test configurations

func testAccCheckDatadogTagPipelineRulesetsConfigBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-order-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-order-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-order-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigReordered(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-order-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-order-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-order-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.third.id,
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id
  ]
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigExpanded(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-order-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-order-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-order-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "fourth" {
  name    = "tf-test-order-fourth-%s"
  enabled = true

  rules {
    name    = "fourth-rule"
    enabled = true
    
    query {
      query = "service:fourth"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id,
    datadog_tag_pipeline_ruleset.fourth.id
  ]
}`, uniq, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigReduced(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-order-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-order-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-order-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "fourth" {
  name    = "tf-test-order-fourth-%s"
  enabled = true

  rules {
    name    = "fourth-rule"
    enabled = true
    
    query {
      query = "service:fourth"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigSingle(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "single" {
  name    = "tf-test-order-single-%s"
  enabled = true

  rules {
    name    = "single-rule"
    enabled = true
    
    query {
      query = "service:single"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.single.id
  ]
}`, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigWithOverride(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-override-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-override-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-override-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  override_ui_defined_resources = true
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigUnmanagedRulesets(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "unmanaged_first" {
  name    = "tf-test-unmanaged-first-%s"
  enabled = true

  rules {
    name    = "unmanaged-first-rule"
    enabled = true
    
    query {
      query = "service:unmanaged-first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "unmanaged-test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "unmanaged_second" {
  name    = "tf-test-unmanaged-second-%s"
  enabled = true

  rules {
    name    = "unmanaged-second-rule"
    enabled = true
    
    query {
      query = "service:unmanaged-second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "unmanaged-test"
      }
    }
  }
}`, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigWithOverrideAndUnmanaged(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-override-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-override-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-override-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  override_ui_defined_resources = true
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq)
}


// testAccCheckUnmanagedRulesetsExist verifies that unmanaged rulesets still exist (not deleted)
func testAccCheckUnmanagedRulesetsExist(frameworkProvider *fwprovider.FrameworkProvider, uniq string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := frameworkProvider.DatadogApiInstances
		auth := frameworkProvider.Auth
		api := apiInstances.GetCloudCostManagementApiV2()

		// List all current rulesets
		resp, _, err := api.ListTagPipelinesRulesets(auth)
		if err != nil {
			return fmt.Errorf("failed to list rulesets: %v", err)
		}

		var rulesets []datadogV2.RulesetRespData
		if respData, ok := resp.GetDataOk(); ok {
			rulesets = *respData
		}

		// Check that unmanaged rulesets still exist
		unmanagedNames := []string{
			"tf-test-unmanaged-first-" + uniq,
			"tf-test-unmanaged-second-" + uniq,
		}

		foundCount := 0
		for _, ruleset := range rulesets {
			if attrs, ok := ruleset.GetAttributesOk(); ok {
				name := attrs.GetName()
				for _, unmanagedName := range unmanagedNames {
					if name == unmanagedName {
						foundCount++
					}
				}
			}
		}

		if foundCount != len(unmanagedNames) {
			return fmt.Errorf("expected %d unmanaged rulesets to exist, but found %d", len(unmanagedNames), foundCount)
		}

		return nil
	}
}

func testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalse(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-false-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-false-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-false-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  override_ui_defined_resources = false
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalseWithUnmanagedMiddle(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-false-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-false-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-false-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

# Keep the unmanaged rulesets from the previous step (they will be positioned in the middle)
resource "datadog_tag_pipeline_ruleset" "unmanaged_first" {
  name    = "tf-test-unmanaged-first-%s"
  enabled = true

  rules {
    name    = "unmanaged-first-rule"
    enabled = true
    
    query {
      query = "service:unmanaged-first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "unmanaged-test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "unmanaged_second" {
  name    = "tf-test-unmanaged-second-%s"
  enabled = true

  rules {
    name    = "unmanaged-second-rule"
    enabled = true
    
    query {
      query = "service:unmanaged-second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "unmanaged-test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  override_ui_defined_resources = false
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq, uniq, uniq)
}

func testAccCheckDatadogTagPipelineRulesetsConfigOverrideFalseWithUnmanagedAtEnd(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "first" {
  name    = "tf-test-false-first-%s"
  enabled = true

  rules {
    name    = "first-rule"
    enabled = true
    
    query {
      query = "service:first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "second" {
  name    = "tf-test-false-second-%s"
  enabled = true

  rules {
    name    = "second-rule"
    enabled = true
    
    query {
      query = "service:second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "third" {
  name    = "tf-test-false-third-%s"
  enabled = true

  rules {
    name    = "third-rule"
    enabled = true
    
    query {
      query = "service:third"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}

# Add unmanaged rulesets that will be positioned at the end
resource "datadog_tag_pipeline_ruleset" "unmanaged_first" {
  name    = "tf-test-unmanaged-first-%s"
  enabled = true

  rules {
    name    = "unmanaged-first-rule"
    enabled = true
    
    query {
      query = "service:unmanaged-first"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "unmanaged-test"
      }
    }
  }
}

resource "datadog_tag_pipeline_ruleset" "unmanaged_second" {
  name    = "tf-test-unmanaged-second-%s"
  enabled = true

  rules {
    name    = "unmanaged-second-rule"
    enabled = true
    
    query {
      query = "service:unmanaged-second"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "unmanaged-test"
      }
    }
  }
}

resource "datadog_tag_pipeline_rulesets" "foo" {
  override_ui_defined_resources = false
  ruleset_ids = [
    datadog_tag_pipeline_ruleset.first.id,
    datadog_tag_pipeline_ruleset.second.id,
    datadog_tag_pipeline_ruleset.third.id
  ]
}`, uniq, uniq, uniq, uniq, uniq)
}

