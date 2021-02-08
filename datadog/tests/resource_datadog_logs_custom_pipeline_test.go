package test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"strings"
	"testing"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func pipelineConfigForCreation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "my_pipeline_test" {
	name = "%s"
	is_enabled = true
	filter {
		query = "source:redis"
	}
	processor {
		date_remapper {
			name = "Define date"
			is_enabled = true
			sources = ["verbose"]
		}
	}
	processor {
		arithmetic_processor {
			name = "processor from nested pipeline"
			is_enabled = true
			expression = "(time1-time2)*1000"
			target = "my_arithmetic"
			is_replace_missing = false
		}
	}
	processor {
		category_processor {
			name = "Categorise severity level"
			is_enabled = true
			target = "redis.severity"
			category {
				filter {
					query = "@severity: \"-\""
				}
				name = "verbose"
			}
			category {
				filter {
					query = "@severity: \".\""
				}
				name = "debug"
			}
		}
	}
	processor {
		grok_parser {
			name = "Parsing Stack traces"
			is_enabled = true
			source = "message"
			grok {
				support_rules = "date_parser %%%%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
				match_rules = "rule %%%%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
			}
		}
	}
	processor {
		pipeline {
			name = "my nested pipeline"
			is_enabled = true
			filter {
				query = "source:kafka"
			}
			processor {
				url_parser {
					name = "Define url parser"
					is_enabled = true
					sources = ["http_test"]
					target = "http_test.details"
					normalize_ending_slashes = false
				}
			}
			processor {
				user_agent_parser {
					name = "Define user agent parser"
					is_enabled = true
					sources = ["http_agent"]
					target = "http_agent.details"
					is_encoded = false
				}
			}
		}
	}
	processor {
		geo_ip_parser {
			name = "geo ip parse"
			is_enabled = true
			target = "ip.address"
			sources = ["ip1"]
		}
	}
	processor {
		lookup_processor {
			source = "ip1"
			target = "ip.address"
			lookup_table = ["key,value"]
		}
	}
	processor {
		lookup_processor {
			name = "lookup processor with optional fields"
			is_enabled = true
			source = "ip2"
			target = "ip.address"
			lookup_table = ["key,value"]
			default_lookup = "default"
		}
	}
}`, uniq)
}
func pipelineConfigForUpdate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "my_pipeline_test" {
	name = "%s"
	is_enabled = false
	filter {
		query = "source:kafka"
	}
	processor {
		date_remapper {
			name = "test date remapper"
			is_enabled = true
			sources = ["verbose"]
		}
	}
	processor {
		status_remapper {
			is_enabled = true
			sources = ["redis.severity"]
		}
	}
	processor {
		attribute_remapper {
			name = "Simple attribute remapper to tag target type"
			is_enabled = true
			sources = ["db.instance"]
			source_type = "tag"
		  	target = "db"
			target_type = "tag"
			preserve_source = true
			override_on_conflict = false
		}
	}
	processor {
		attribute_remapper {
			name = "Simple attribute remapper to attribute target type"
			is_enabled = true
			sources = ["db.instance"]
			source_type = "tag"
		  	target = "db"
			target_type = "attribute"
			target_format = "string"
			preserve_source = true
			override_on_conflict = false
		}
	}
	processor {
		grok_parser {
			name = "Parsing Stack traces"
			is_enabled = true
			source = "message"
			samples = ["sample1", "sample2"]
			grok {
				support_rules = "date_parser %%%%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
				match_rules = "rule %%%%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
			}
		}
	}
	processor {
		string_builder_processor {
			name = "string builder"
			is_enabled = true
			template = "%%%%{user.name} is awesome"
			target = "user.name"
			is_replace_missing = true
		}
	}
	processor {
		geo_ip_parser {
			name = "geo ip parse"
			is_enabled = true
			target = "ip.address"
			sources = ["ip1", "ip2"]
		}
	}
	processor {
		lookup_processor {
			source = "ip1"
			target = "ip.address"
			lookup_table = ["key,value", "key2,value2"]
		}
	}
	processor {
		lookup_processor {
			name = "lookup processor with optional fields"
			is_enabled = true
			source = "ip2"
			target = "ip.address"
			lookup_table = ["key,value", "key2,value2"]
			default_lookup = "default"
		}
	}
}`, uniq)
}
func pipelineConfigWithEmptyFilterQuery(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "empty_filter_query_pipeline" {
	name = "%s"
	is_enabled = "true"
	filter {
		query = ""
	}
	processor {
		status_remapper {
			is_enabled = true
			sources = ["redis.severity"]
		}
	}
}`, uniq)
}

func TestAccDatadogLogsPipeline_basic(t *testing.T) {
	rec := initRecorder(t)
	accProviders, clock, cleanup := testAccProviders(t, rec)
	pipelineName := uniqueEntityName(clock, t)
	pipelineName2 := pipelineName + "-updated"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: pipelineConfigForCreation(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "filter.0.query", "source:redis"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.0.date_remapper.0.sources.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.1.arithmetic_processor.0.expression", "(time1-time2)*1000"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.2.category_processor.0.category.0.filter.0.query", "@severity: \"-\""),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.3.grok_parser.0.grok.0.match_rules", "rule %{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.3.grok_parser.0.samples.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.4.pipeline.0.filter.0.query", "source:kafka"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.4.pipeline.0.processor.0.url_parser.0.sources.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.4.pipeline.0.processor.1.user_agent_parser.0.target", "http_agent.details"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.5.geo_ip_parser.0.sources.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.6.lookup_processor.0.lookup_table.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.7.lookup_processor.0.lookup_table.#", "1"),
				),
			}, {
				Config: pipelineConfigForUpdate(pipelineName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineName2),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "filter.0.query", "source:kafka"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.2.attribute_remapper.0.preserve_source", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.3.attribute_remapper.0.target_type", "attribute"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.3.attribute_remapper.0.target_format", "string"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.4.grok_parser.0.samples.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.5.string_builder_processor.0.template", "%{user.name} is awesome"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.6.geo_ip_parser.0.sources.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.7.lookup_processor.0.lookup_table.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.8.lookup_processor.0.lookup_table.#", "2"),
				),
			},
		},
	})
}

func TestAccDatadogLogsPipelineEmptyFilterQuery(t *testing.T) {
	rec := initRecorder(t)
	accProviders, clock, cleanup := testAccProviders(t, rec)
	pipelineName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: pipelineConfigWithEmptyFilterQuery(pipelineName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.empty_filter_query_pipeline", "name", pipelineName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.empty_filter_query_pipeline", "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.empty_filter_query_pipeline", "filter.0.query", ""),
				),
			},
		},
	})
}

func testAccCheckPipelineExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := pipelineExistsChecker(s, authV1, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func pipelineExistsChecker(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_custom_pipeline" {
			id := r.Primary.ID
			if _, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, id).Execute(); err != nil {
				return fmt.Errorf("received an error when retrieving pipeline, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckPipelineDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := pipelineDestroyHelper(s, authV1, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func pipelineDestroyHelper(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_custom_pipeline" {
			id := r.Primary.ID
			p, _, err := datadogClientV1.LogsPipelinesApi.GetLogsPipeline(authV1, id).Execute()

			if err != nil {
				if strings.Contains(err.Error(), "400 Bad Request") {
					continue
				}
				return fmt.Errorf("received an error when retrieving pipeline, (%s)", err)
			}
			if &p != nil {
				return fmt.Errorf("pipeline still exists")
			}
		}

	}
	return nil
}
