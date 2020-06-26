package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const pipelineConfigForImport = `
resource "datadog_logs_custom_pipeline" "test_import" {
	name = "imported pipeline"
	is_enabled = false
	filter {
		query = "source:kafka"
	}
	processor {
		arithmetic_processor {
			name = "test arithmetic processor"
			expression = "(time1 - time2)*1000"
			target = "my_arithmetic"
			is_replace_missing = true
		}
	}
	processor {
		attribute_remapper {
			name = "test attribute remapper"
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
		category_processor {
			name = "test category processor"
			target = "redis.severity"
			category {
				name = "debug"
				filter {
					query = "@severity: \".\""
				}
			}
			category {
				name = "verbose"
				filter {
					query = "@severity: \"-\""
				}
			}
		}
	}
	processor {
		date_remapper {
			name = "test date remapper"
			is_enabled = true
			sources = ["date"]
		}
	}
	processor {
		date_remapper {
			name = "2nd date remapper"
			is_enabled = true
			sources = ["other"]
		}
	}
	processor {
		message_remapper {
			name = "test message remapper"
			is_enabled = false
			sources = ["message"]
		}
	}
	processor {
		service_remapper {
			name = "test service remapper"
			sources = ["service"]
		}
	}
	processor {
		status_remapper {
			name = "test status remapper"
			sources = ["status", "extra"]
		}
	}
	processor {
		trace_id_remapper {
			name = "test trace id remapper"
			sources = ["dd.trace_id"]
		}
	}
	processor {
		pipeline {
			name = "nested pipeline"
			filter {
				query = "source:redis"
			}
			processor {
				grok_parser {
					name = "test grok parser"
					source = "message"
					grok {
						support_rules = ""
						match_rules = "Rule %%{word:my_word2} %%{number:my_float2}"
					}
				}
			}
			processor {
				url_parser {
					name = "test url parser"
					sources = ["url", "extra"]
					target = "http_url"
					normalize_ending_slashes = true
				}
			}
		}
	}
	processor {
		user_agent_parser {
			name = "test user agent parser"
			sources = ["user", "agent"]
			target = "http_agent"
			is_encoded = false
		}
	}
}
`

func TestAccLogsCustomPipeline_importBasic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: pipelineConfigForImport,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.test_import", "name", "imported pipeline"),
				),
			},
			{
				ResourceName:      "datadog_logs_custom_pipeline.test_import",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
