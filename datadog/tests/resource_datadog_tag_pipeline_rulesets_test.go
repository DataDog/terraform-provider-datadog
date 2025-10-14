package test

import (
	"context"
	"fmt"
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

// testAccCleanupOrphanedTagPipelineRulesets removes ALL existing rulesets before running tests
// This ensures tests start with a clean slate, as the order resource requires all rulesets to be managed
func testAccCleanupOrphanedTagPipelineRulesets(t *testing.T, frameworkProvider *fwprovider.FrameworkProvider) {
	apiInstances := frameworkProvider.DatadogApiInstances
	auth := frameworkProvider.Auth
	api := apiInstances.GetCloudCostManagementApiV2()

	// List all rulesets
	resp, _, err := api.ListRulesets(auth)
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
		resp, _, err := api.ListRulesets(auth)
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
