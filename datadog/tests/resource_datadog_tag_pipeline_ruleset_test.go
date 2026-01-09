package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogTagPipelineRuleset_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-tag-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "enabled", "true"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_ruleset.foo", "id"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_ruleset.foo", "position"),
					resource.TestCheckResourceAttrSet("datadog_tag_pipeline_ruleset.foo", "version"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.#", "1"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.name", "test-rule"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRuleset_MappingRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigMapping(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-mapping-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.destination_key", "env"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.if_not_exists", "true"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.source_keys.#", "2"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.source_keys.*", "environment"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.source_keys.*", "stage"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRuleset_QueryRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigQuery(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-query-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.query.query", "source:app AND env:prod"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.query.case_insensitivity", "true"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.query.if_not_exists", "false"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.query.addition.key", "team"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.query.addition.value", "backend"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRuleset_ReferenceTableRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigReferenceTable(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-ref-table-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.table_name", "service_mapping"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.case_insensitivity", "true"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.if_not_exists", "false"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.source_keys.#", "1"),
					resource.TestCheckTypeSetElemAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.source_keys.*", "service"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.#", "2"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.0.input_column", "service_name"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.0.output_key", "service"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.1.input_column", "team_name"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.1.output_key", "team"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRuleset_MultipleRules(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigMultiple(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-multi-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.#", "3"),
					// First rule (mapping)
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.name", "mapping-rule"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.destination_key", "env"),
					// Second rule (query)
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.1.name", "query-rule"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.1.query.query", "service:web"),
					// Third rule (reference table)
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.2.name", "ref-table-rule"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.2.reference_table.table_name", "lookup_table"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRuleset_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.#", "1"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.name", "test-rule"),
				),
			},
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-updated-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "enabled", "false"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.#", "2"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.0.name", "updated-rule"),
					resource.TestCheckResourceAttr("datadog_tag_pipeline_ruleset.foo", "rules.1.name", "new-rule"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRuleset_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetConfigBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogTagPipelineRulesetExists("datadog_tag_pipeline_ruleset.foo"),
				),
			},
			{
				ResourceName:            "datadog_tag_pipeline_ruleset.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"position"},
			},
		},
	})
}

func testAccCheckDatadogTagPipelineRulesetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}

func testAccCheckDatadogTagPipelineRulesetDestroy(ctx context.Context, frameworkProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := frameworkProvider.DatadogApiInstances
		auth := frameworkProvider.Auth

		if err := datadogTagPipelineRulesetDestroyHelper(ctx, auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogTagPipelineRulesetDestroyHelper(ctx context.Context, auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetCloudCostManagementApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_tag_pipeline_ruleset" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetTagPipelinesRuleset(auth, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving tag pipeline ruleset: %s", err.Error())
		}
		return fmt.Errorf("tag pipeline ruleset still exists")
	}
	return nil
}

func testAccCheckDatadogTagPipelineRulesetConfigBasic(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-tag-ruleset-%s"
  enabled = true

  rules {
    name    = "test-rule"
    enabled = true
    
    query {
      query = "service:test"
      if_not_exists = false
      
      addition {
        key   = "env"
        value = "test"
      }
    }
  }
}`, uniq)
}

func testAccCheckDatadogTagPipelineRulesetConfigMapping(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-mapping-ruleset-%s"
  enabled = true

  rules {
    name    = "mapping-rule"
    enabled = true
    
    mapping {
      destination_key = "env"
      if_not_exists   = true
      source_keys     = ["environment", "stage"]
    }
  }
}`, uniq)
}

func testAccCheckDatadogTagPipelineRulesetConfigQuery(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-query-ruleset-%s"
  enabled = true

  rules {
    name    = "query-rule"
    enabled = true
    
    query {
      query             = "source:app AND env:prod"
      case_insensitivity = true
      if_not_exists     = false
      
      addition {
        key   = "team"
        value = "backend"
      }
    }
  }
}`, uniq)
}

func testAccCheckDatadogTagPipelineRulesetConfigReferenceTable(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-ref-table-ruleset-%s"
  enabled = true

  rules {
    name    = "ref-table-rule"
    enabled = true
    
    reference_table {
      table_name         = "service_mapping"
      case_insensitivity = true
      if_not_exists      = false
      source_keys        = ["service"]
      
      field_pairs {
        input_column = "service_name"
        output_key   = "service"
      }
      
      field_pairs {
        input_column = "team_name"
        output_key   = "team"
      }
    }
  }
}`, uniq)
}

func testAccCheckDatadogTagPipelineRulesetConfigMultiple(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-multi-ruleset-%s"
  enabled = true

  rules {
    name    = "mapping-rule"
    enabled = true
    
    mapping {
      destination_key = "env"
      if_not_exists   = true
      source_keys     = ["environment"]
    }
  }
  
  rules {
    name    = "query-rule" 
    enabled = true
    
    query {
      query         = "service:web"
      if_not_exists = false
      
      addition {
        key   = "tier"
        value = "frontend"
      }
    }
  }
  
  rules {
    name    = "ref-table-rule"
    enabled = true
    
    reference_table {
      table_name         = "lookup_table"
      case_insensitivity = false
      if_not_exists      = false
      source_keys        = ["app"]
      
      field_pairs {
        input_column = "application"
        output_key   = "app_name"
      }
    }
  }
}`, uniq)
}

func testAccCheckDatadogTagPipelineRulesetConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-updated-ruleset-%s"
  enabled = false

  rules {
    name    = "updated-rule"
    enabled = true
    
    mapping {
      destination_key = "environment"
      if_not_exists   = false
      source_keys     = ["env"]
    }
  }
  
  rules {
    name    = "new-rule"
    enabled = false
    
    query {
      query         = "service:updated-service"
      if_not_exists = false
      
      addition {
        key   = "tier"
        value = "backend"
      }
    }
  }
}`, uniq)
}
