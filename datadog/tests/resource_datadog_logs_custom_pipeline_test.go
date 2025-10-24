package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func pipelineConfigForCreation(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "my_pipeline_test" {
	name = "%s"
	is_enabled = true
	tags = ["key1:value1", "key2:value2"]
	description = "Pipeline description"
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
	processor {
		reference_table_lookup_processor {
			name           = "reftablelookup"
			is_enabled     = true
			source         = "sourcefield"
			target         = "targetfield"
			lookup_enrichment_table   = "test_reference_table_do_not_delete"
		}
	}
 
    processor {
      span_id_remapper {
			name   = "span_id_remapper"
			is_enabled = true
			sources = [ "dd.span_id" ]
		}
    }
    processor {
        array_processor {
            name = "array append operation"
            is_enabled = true
            operation {
                append {
                    source = "network.client.ip"
                    target = "sourceIps"
                    preserve_source = true
                }
            }
        }
    }
    processor {
        array_processor {
            name = "array length operation"
            is_enabled = true
            operation {
                length {
                    source = "tags"
                    target = "tagCount"
                }
            }
        }
    }
    processor {
        array_processor {
            name = "array select operation"
            is_enabled = true
            operation {
                select {
                    source = "httpRequest.headers"
                    target = "referrer"
                    filter = "name:Referrer"
                    value_to_extract = "value"
                }
            }
        }
    }
    processor {
        decoder_processor {
            name = "decoder operation"
            is_enabled = true
            source = "encoded_message"
            target = "decoded_messsage"
            binary_to_text_encoding = "base64"
            input_representation = "utf_8"
        }
    }
    processor {
        schema_processor {
            name = "Map to OCSF Schema"
            is_enabled = true

            schema {
                schema_type = "ocsf"
                version = "1.5.0"
                class_name = "Account Change"
                class_uid = 3001
                profiles = ["cloud"]
            }

            mappers {
                schema_remapper {
                    name = "Map ocsf.user to ocsf.user"
                    sources = ["ocsf.user"]
                    target = "ocsf.user"
                }

                schema_remapper {
                    name = "Map ocsf.time to ocsf.time"
                    sources = ["ocsf.time"]
                    target = "ocsf.time"
                }

                schema_remapper {
                    name = "Map ocsf.cloud to ocsf.cloud"
                    sources = ["ocsf.cloud"]
                    target = "ocsf.cloud"
                }

                schema_category_mapper {
                    name = "Map to OCSF severity"

                    targets {
                        name = "ocsf.severity"
                        id = "ocsf.severity_id"
                    }

                    categories {
                        name = "Informational"
                        id = 1
                        filter {
                            query = "@eventName:*"
                        }
                    }
                }

                schema_category_mapper {
                    name = "Map to OCSF categories"

                    targets {
                        name = "ocsf.activity_name"
                        id = "ocsf.activity_id"
                    }

                    categories {
                        name = "Create"
                        id = 1
                        filter {
                            query = "eventName:CreateUser"
                        }
                    }

                    fallback {
                        values = {
                            "ocsf.activity_name" = "Other"
                            "ocsf.activity_id" = "99"
                        }
                        sources = {
                            "ocsf.activity_name" = "[\"eventName\"]"
                        }
                    }
                }
            }
        }
    }

}`, uniq)
}

func pipelineConfigWithoutTagsAndDescriptionForCreation(uniq string) string {
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
}`, uniq)
}

func pipelineConfigForUpdate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "my_pipeline_test" {
	name = "%s"
	is_enabled = false
	tags = ["key1:value1", "key2:value2"]
	description = "Pipeline description"
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
	processor {
		reference_table_lookup_processor {
			name           = "reftablelookup"
			is_enabled     = true
			source         = "sourcefield"
			target         = "targetfield"
			lookup_enrichment_table   = "test_reference_table_do_not_delete"
		}
	}
 
    processor {
      span_id_remapper {
			name   = "span_id_remapper"
			is_enabled = true
			sources = [ "dd.span_id" ]
		}
    }
    processor {
        array_processor {
            name = "array append operation updated"
            is_enabled = false
            operation {
                append {
                    source = "network.server.ip"
                    target = "destinationIps"
                    preserve_source = true
                }
            }
        }
    }
    processor {
        array_processor {
            name = "array length operation updated"
            is_enabled = true
            operation {
                length {
                    source = "headers"
                    target = "headerCount"
                }
            }
        }
    }
    processor {
        array_processor {
            name = "array select operation updated"
            is_enabled = true
            operation {
                select {
                    source = "httpResponse.headers"
                    target = "contentType"
                    filter = "name:Content-Type"
                    value_to_extract = "value"
                }
            }
        }
    }
    processor {
        decoder_processor {
            name = "decoder operation updated"
            is_enabled = false
            source = "hex_encoded_data"
            target = "decoded_messsage"
            binary_to_text_encoding = "base16"
            input_representation = "integer"
        }
    }
processor {
        schema_processor {
            name = "Map to OCSF Schema updated"
            is_enabled = true

            schema {
                schema_type = "ocsf"
                version = "1.5.0"
                class_name = "Authentication"
                class_uid = 3002
            }

            mappers {
                schema_remapper {
                    name = "Map ocsf.user to ocsf.user"
                    sources = ["ocsf.user"]
                    target = "ocsf.user"
                }

                schema_remapper {
                    name = "Map http_request to ocsf.http.request"
                    sources = ["http_request"]
                    target = "ocsf.http_request"
                }

                schema_remapper {
                    name = "Map ocsf.time to ocsf.time"
                    sources = ["ocsf.time"]
                    target = "ocsf.time"
                }

                schema_category_mapper {
                    name = "Map to OCSF severity"

                    targets {
                        name = "ocsf.severity"
                        id = "ocsf.severity_id"
                    }

                    categories {
                        name = "Informational"
                        id = 1
                        filter {
                            query = "@eventName:*"
                        }
                    }
                }

                schema_category_mapper {
                    name = "Map to OCSF categories"

                    targets {
                        name = "ocsf.activity_name"
                        id = "ocsf.activity_id"
                    }

                    categories {
                        name = "Create"
                        id = 1
                        filter {
                            query = "eventName:CreateUser"
                        }
                    }

                    fallback {
                        values = {
                            "ocsf.activity_name" = "Other"
                            "ocsf.activity_id" = "99"
                        }
                        sources = {
                            "ocsf.activity_name" = "[\"eventName\"]"
                        }
                    }
                }
            }
        }
    }

}`, uniq)
}

func pipelineConfigWithoutTagsAndDescriptionForUpdate(uniq string) string {
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

	processor {
		category_processor {
		  target = "foo.severity"
		  category {
			name = "debug"
			filter {
			  query = ""
			}
		  }
		  name       = "sample category processor"
		  is_enabled = true
		}
	  }

	processor {
		pipeline {
			is_enabled = true
			name       = "Nginx"
			filter {
				query = ""
			}

		}
	}
}`, uniq)
}

func TestAccDatadogLogsPipeline_basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	pipelineCreatedName := uniqueEntityName(ctx, t)
	pipelineUpdatedName := pipelineCreatedName + "-updated"
	pipelineWithoutTagsAndDescriptionCreatedName := pipelineCreatedName + "-without-tags-and-description"
	pipelineWithoutTagsAndDescriptionUpdatedName := pipelineUpdatedName + "-without-tags-and-description"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: pipelineConfigForCreation(pipelineCreatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineCreatedName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.0", "key1:value1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.1", "key2:value2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "description", "Pipeline description"),
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
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.8.reference_table_lookup_processor.0.lookup_enrichment_table", "test_reference_table_do_not_delete"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.9.span_id_remapper.0.sources.#", "1"),
				),
			}, {
				Config: pipelineConfigForUpdate(pipelineUpdatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineUpdatedName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.0", "key1:value1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.1", "key2:value2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "description", "Pipeline description"),
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
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.9.reference_table_lookup_processor.0.lookup_enrichment_table", "test_reference_table_do_not_delete"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "processor.10.span_id_remapper.0.sources.#", "1"),
				),
			}, { // for backward compatibility as tags and description were not always present and clients have Terraform files without them
				Config: pipelineConfigWithoutTagsAndDescriptionForCreation(pipelineWithoutTagsAndDescriptionCreatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineWithoutTagsAndDescriptionCreatedName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "filter.0.query", "source:redis"),
				),
			}, {
				Config: pipelineConfigWithoutTagsAndDescriptionForUpdate(pipelineWithoutTagsAndDescriptionUpdatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineWithoutTagsAndDescriptionUpdatedName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "is_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "filter.0.query", "source:kafka"),
				),
			},
		},
	})
}

func TestAccDatadogLogsPipeline_import(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	pipelineName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckPipelineDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: pipelineConfigForCreation(pipelineName),
			},
			{
				ResourceName:      "datadog_logs_custom_pipeline.my_pipeline_test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogLogsPipelineEmptyFilterQuery(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	pipelineName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckPipelineDestroy(accProvider),
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
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.empty_filter_query_pipeline", "processor.1.category_processor.0.category.0.filter.0.query", ""),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.empty_filter_query_pipeline", "processor.2.pipeline.0.filter.0.query", ""),
				),
			},
		},
	})
}

func TestAccDatadogLogsPipelineDefaultTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	pipelineName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{ // Default tags are correctly added
				Config: pipelineConfigForCreation(pipelineName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.*", "key1:value1"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.*", "key2:value2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.*", "default_key:default_value"),
				),
			},
			{ // Resource tags take precedence over default tags and duplicates stay
				Config: pipelineConfigForCreation(pipelineName),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"key1": "new_value1",
						"key2": "new_value2",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "name", pipelineName),
					resource.TestCheckResourceAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.*", "key1:value1"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_logs_custom_pipeline.my_pipeline_test", "tags.*", "key2:value2"),
				),
			},
		},
	})
}

func testAccCheckPipelineExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := pipelineExistsChecker(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func pipelineExistsChecker(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_custom_pipeline" {
			id := r.Primary.ID
			if _, _, err := apiInstances.GetLogsPipelinesApiV1().GetLogsPipeline(ctx, id); err != nil {
				return fmt.Errorf("received an error when retrieving pipeline, (%s)", err)
			}
		}
	}
	return nil
}

func testAccCheckPipelineDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := pipelineDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func pipelineDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_logs_custom_pipeline" {
			err := utils.Retry(2, 5, func() error {
				id := r.Primary.ID
				_, _, err := apiInstances.GetLogsPipelinesApiV1().
					GetLogsPipeline(ctx, id)
				if err != nil {
					if strings.Contains(err.Error(), "400 Bad Request") {
						return nil
					}
					return &utils.FatalError{Prob: fmt.Sprintf("received an error when retrieving pipeline, (%s)", err)}
				}
				return &utils.RetryableError{Prob: "pipeline still exists"}
			})
			return err
		}
	}
	return nil
}
