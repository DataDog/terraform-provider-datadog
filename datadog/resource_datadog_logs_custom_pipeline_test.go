package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

const pipelineConfigForCreation = `
resource "datadog_logs_custom_pipeline" "my_pipeline_test" {
	name = "my first pipeline"
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
				support_rules = "date_parser %%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
				match_rules = "rule %%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
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
}
`
const pipelineConfigForUpdate = `
resource "datadog_logs_custom_pipeline" "my_pipeline_test" {
	name = "updated pipeline"
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
			name = "Simple attribute remapper"
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
		grok_parser {
			name = "Parsing Stack traces"
			is_enabled = true
			source = "message"
			samples = ["sample1", "sample2"]
			grok {
				support_rules = "date_parser %%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
				match_rules = "rule %%{date(\"yyyy-MM-dd HH:mm:ss,SSS\"):timestamp}"
			}
		}
	}
	processor {
		string_builder_processor {
			name = "string builder"
			is_enabled = true
			template = "%%{user.name} is awesome"
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
}
`

func TestAccDatadogLogsPipeline_basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: pipelineConfigForCreation,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider, "datadog_logs_custom_pipeline.my_pipeline_test"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", "my first pipeline"),
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
				),
			}, {
				Config: pipelineConfigForUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider, "datadog_logs_custom_pipeline.my_pipeline_test"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", "updated pipeline"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "filter.0.query", "source:kafka"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.2.attribute_remapper.0.preserve_source", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.3.grok_parser.0.samples.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.4.string_builder_processor.0.template", "%{user.name} is awesome"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.5.geo_ip_parser.0.sources.#", "2"),
				),
			},
		},
	})
}

func testAccCheckPipelineExists(accProvider *schema.Provider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)
		if err := pipelineExistsChecker(s, client); err != nil {
			return err
		}
		return nil
	}
}

func pipelineExistsChecker(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_custom_pipeline" {
			id := r.Primary.ID
			if _, err := client.GetLogsPipeline(id); err != nil {
				return fmt.Errorf("received an error when retrieving pipeline, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckPipelineDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)
		if err := pipelineDestroyHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func pipelineDestroyHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_custom_pipeline" {
			id := r.Primary.ID
			p, err := client.GetLogsPipeline(id)

			if err != nil {
				if strings.Contains(err.Error(), "400 Bad Request") {
					continue
				}
				return fmt.Errorf("received an error when retrieving pipeline, (%s)", err)
			}
			if p != nil {
				return fmt.Errorf("pipeline still exists")
			}
		}

	}
	return nil
}
