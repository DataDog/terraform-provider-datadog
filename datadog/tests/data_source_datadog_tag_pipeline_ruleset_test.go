package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogTagPipelineRulesetDataSource_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	rulesetName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogTagPipelineRulesetDataSourceConfig(rulesetName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "name", rulesetName),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.#", "1"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.name", "Add Team Tag"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.query.query", "service:web-api"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.query.case_insensitivity", "true"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.query.if_tag_exists", "do_not_apply"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.query.addition.key", "team"),
					resource.TestCheckResourceAttr(
						"data.datadog_tag_pipeline_ruleset.foo", "rules.0.query.addition.value", "backend"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesetDataSource_MappingRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTagPipelineRulesetMappingConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-mapping-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.destination_key", "env"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.if_tag_exists", "append"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.source_keys.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.source_keys.*", "environment"),
					resource.TestCheckTypeSetElemAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.mapping.source_keys.*", "stage"),
				),
			},
		},
	})
}

func TestAccDatadogTagPipelineRulesetDataSource_ReferenceTableRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogTagPipelineRulesetDestroy(ctx, providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceTagPipelineRulesetReferenceTableConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "name", fmt.Sprintf("tf-test-ref-table-ruleset-%s", uniq)),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.table_name", "service_mapping"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.case_insensitivity", "true"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.if_tag_exists", "replace"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.source_keys.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.source_keys.*", "service"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.#", "2"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.0.input_column", "service_name"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.0.output_key", "service"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.1.input_column", "team_name"),
					resource.TestCheckResourceAttr("data.datadog_tag_pipeline_ruleset.foo", "rules.0.reference_table.field_pairs.1.output_key", "team"),
				),
			},
		},
	})
}

func testAccCheckDatadogTagPipelineRulesetDataSourceConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name = "%s"
  enabled = true
  rules {
    enabled = true
    name = "Add Team Tag"
    query {
      query = "service:web-api"
      case_insensitivity = true
      if_tag_exists = "do_not_apply"
      addition {
        key = "team"
        value = "backend"
      }
    }
  }
}

data "datadog_tag_pipeline_ruleset" "foo" {
  id = datadog_tag_pipeline_ruleset.foo.id
}`, uniq)
}

func testAccDatasourceTagPipelineRulesetMappingConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_tag_pipeline_ruleset" "foo" {
  name    = "tf-test-mapping-ruleset-%s"
  enabled = true

  rules {
    name    = "mapping-rule"
    enabled = true
    
    mapping {
      destination_key = "env"
      if_tag_exists   = "append"
      source_keys     = ["environment", "stage"]
    }
  }
}

data "datadog_tag_pipeline_ruleset" "foo" {
  id = datadog_tag_pipeline_ruleset.foo.id
}`, uniq)
}

func testAccDatasourceTagPipelineRulesetReferenceTableConfig(uniq string) string {
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
      if_tag_exists      = "replace"
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
}

data "datadog_tag_pipeline_ruleset" "foo" {
  id = datadog_tag_pipeline_ruleset.foo.id
}`, uniq)
}
