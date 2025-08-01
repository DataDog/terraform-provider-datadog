package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogObservabilityPipelineImport(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_observability_pipeline.basic"
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumMetricDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineBasicConfig(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_basic(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.basic"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "test pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.datadog_agent.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_json.0.id", "parser-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.id", "destination-1"),
				),
			},
			{
				Config: testAccObservabilityPipelineUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "updated pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_json.0.include", "service:updated-service"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineBasicConfig() string {
	return fmt.Sprintf(`
resource "datadog_observability_pipeline" "basic" {
  name = "test pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      parse_json {
        id      = "parser-1"
        include = "service:my-service"
        field   = "message"
        inputs  = ["source-1"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["parser-1"]
      }
    }
  }
}`)
}

func testAccObservabilityPipelineUpdatedConfig() string {
	return `
resource "datadog_observability_pipeline" "basic" {
  name = "updated pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
        tls {
          crt_file = "/path/to/cert.crt"
        }
      }
    }

    processors {
      parse_json {
        id      = "parser-1"
        include = "service:updated-service"
        field   = "message"
        inputs  = ["source-1"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["parser-1"]
      }
    }
  }
}`
}

func testAccCheckDatadogPipelinesDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := PipelinesDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func PipelinesDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_observability_pipeline" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetObsPipelinesV2().GetPipeline(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Observability Pipeline %s", err)}
			}
			return &utils.RetryableError{Prob: "Observability Pipeline still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogPipelinesExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := pipelinesExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func pipelinesExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_observability_pipeline" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetObsPipelinesV2().GetPipeline(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Observability Pipeline")
		}
	}
	return nil
}

func TestAccDatadogObservabilityPipeline_kafka(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.kafka_test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "kafka_test" {
  name = "kafka pipeline"

  config {
    sources {
      kafka {
        id       = "kafka-source-1"
        group_id = "consumer-group-1"
        topics   = ["topic-a", "topic-b"]

        tls {
          crt_file = "/path/to/kafka.crt"
        }

        sasl {
          mechanism = "PLAIN"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["kafka-source-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "kafka pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.kafka.0.id", "kafka-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.kafka.0.group_id", "consumer-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.kafka.0.topics.0", "topic-a"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.kafka.0.topics.1", "topic-b"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.kafka.0.sasl.mechanism", "PLAIN"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.id", "destination-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_datadogAgentWithTLS(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.agent_tls"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "agent_tls" {
  name = "agent with tls"

  config {
    sources {
      datadog_agent {
        id = "source-with-tls"
        tls {
          crt_file = "/etc/certs/agent.crt"
          ca_file  = "/etc/certs/ca.crt"
          key_file = "/etc/certs/agent.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["source-with-tls"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "agent with tls"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.datadog_agent.0.id", "source-with-tls"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.datadog_agent.0.tls.crt_file", "/etc/certs/agent.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.datadog_agent.0.tls.ca_file", "/etc/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.datadog_agent.0.tls.key_file", "/etc/certs/agent.key"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.id", "destination-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_filterProcessor(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.filter"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "filter" {
  name = "filter-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      filter {
        id      = "filter-1"
        include = "env:prod"
        inputs  = ["source-1"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["filter-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "filter-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.filter.0.id", "filter-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.filter.0.include", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.filter.0.inputs.0", "source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_renameFieldsProcessor(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.rename_fields"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "rename_fields" {
  name = "rename-fields-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      rename_fields {
        id      = "rename-1"
        include = "*"
        inputs  = ["source-1"]

        field {
          source          = "old.field"
          destination     = "new.field"
          preserve_source = true
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["rename-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "rename-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.rename_fields.0.id", "rename-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.rename_fields.0.field.0.source", "old.field"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.rename_fields.0.field.0.destination", "new.field"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.rename_fields.0.field.0.preserve_source", "true"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_removeFieldsProcessor(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.remove_fields"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "remove_fields" {
  name = "remove-fields-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      remove_fields {
        id      = "remove-1"
        include = "*"
        inputs  = ["source-1"]
        fields  = ["temp.debug", "internal.trace_id"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["remove-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "remove-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.remove_fields.0.id", "remove-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.remove_fields.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.remove_fields.0.fields.0", "temp.debug"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.remove_fields.0.fields.1", "internal.trace_id"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_quotaProcessor(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.quota"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "quota" {
  name = "quota-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      quota {
        id      = "quota-1"
        include = "*"
        inputs  = ["source-1"]
        name    = "limitByHostAndEnv"
        drop_events = true
        ignore_when_missing_partitions = true
        partition_fields = ["host", "env"]

        limit {
          enforce = "events"
          limit   = 1000
        }

        overrides {
          field {
            name  = "env"
            value = "prod"
          }
          field {
            name  = "host"
            value = "*"
          }
          limit {
            enforce = "events"
            limit   = 500
          }
        }

        overrides {
          field {
            name  = "env"
            value = "*"
          }
          field {
            name  = "host"
            value = "localhost"
          }
          limit {
            enforce = "bytes"
            limit   = 300
          }
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["quota-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "quota-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.id", "quota-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.name", "limitByHostAndEnv"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.drop_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.limit.enforce", "events"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.limit.limit", "1000"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.0.field.0.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.0.field.0.value", "prod"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.0.field.1.name", "host"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.0.field.1.value", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.0.limit.enforce", "events"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.0.limit.limit", "500"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.field.0.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.field.0.value", "*"), resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.field.0.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.field.1.name", "host"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.field.1.value", "localhost"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.limit.enforce", "bytes"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overrides.1.limit.limit", "300"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_parseJsonProcessor(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.parse_json"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "parse_json" {
  name = "parse-json-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      parse_json {
        id      = "parser-1"
        include = "env:parse"
        field   = "message"
        inputs  = ["source-1"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["parser-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "parse-json-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_json.0.id", "parser-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_json.0.include", "env:parse"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_json.0.field", "message"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_json.0.inputs.0", "source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_addFieldsProcessor(t *testing.T) {

	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.add_fields"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "add_fields" {
  name = "add-fields-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      add_fields {
        id      = "add-fields-1"
        include = "*"
        inputs  = ["source-1"]

        field {
          name  = "custom.field"
          value = "hello-world"
        }
        field {
          name  = "env"
          value = "prod"
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["add-fields-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "add-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.id", "add-fields-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.field.0.name", "custom.field"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.field.0.value", "hello-world"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.field.1.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.field.1.value", "prod"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_parseGrokProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.parse_grok"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "parse_grok" {
  name = "parse-grok-test"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      parse_grok {
        id      = "parse-grok-1"
        include = "*"
        inputs  = ["source-1"]
        disable_library_rules = true

        rules {
          source = "message"

          match_rule {
            name = "match_user"
            rule = "%%{word:user.name}"
          }

          match_rule {
            name = "match_action"
            rule = "%%{word:action}"
          }

          support_rule {
            name = "word"
            rule = "\\w+"
          }

          support_rule {
            name = "custom_word"
            rule = "[a-zA-Z]+"
          }
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["parse-grok-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "parse-grok-test"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.id", "parse-grok-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.disable_library_rules", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.source", "message"),

					// Match Rules
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.match_rule.0.name", "match_user"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.match_rule.0.rule", "%{word:user.name}"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.match_rule.1.name", "match_action"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.match_rule.1.rule", "%{word:action}"),

					// Support Rules
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.support_rule.0.name", "word"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.support_rule.0.rule", "\\w+"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.support_rule.1.name", "custom_word"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.parse_grok.0.rules.0.support_rule.1.rule", "[a-zA-Z]+"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_sampleProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.sample"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "sample" {
  name = "sample-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      sample {
        id      = "sample-1"
        include = "*"
        inputs  = ["source-1"]
        rate    = 10
      }

      sample {
        id      = "sample-2"
        include = "*"
        inputs  = ["sample-1"]
        percentage    = 4.99
        group_by = ["env", "service"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["sample-2"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "sample-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.0.id", "sample-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.0.rate", "10"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.1.id", "sample-2"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.1.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.1.percentage", "4.99"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.1.group_by.0", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sample.1.group_by.1", "service"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_fluentdSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.fluentd"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "fluentd" {
  name = "fluent-pipeline"

  config {
    sources {
      fluentd {
        id = "fluent-source-1"
        tls {
          crt_file = "/etc/ssl/certs/fluent.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/fluent.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["fluent-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "fluent-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluentd.0.id", "fluent-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluentd.0.tls.crt_file", "/etc/ssl/certs/fluent.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluentd.0.tls.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluentd.0.tls.key_file", "/etc/ssl/private/fluent.key"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.inputs.0", "fluent-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_fluentBitSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.fluent_bit"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "fluent_bit" {
  name = "fluent-pipeline"

  config {
    sources {
      fluent_bit {
        id = "fluent-source-1"
        tls {
          crt_file = "/etc/ssl/certs/fluent.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/fluent.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["fluent-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "fluent-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluent_bit.0.id", "fluent-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluent_bit.0.tls.crt_file", "/etc/ssl/certs/fluent.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluent_bit.0.tls.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.fluent_bit.0.tls.key_file", "/etc/ssl/private/fluent.key"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.inputs.0", "fluent-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_httpServerSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.http_server"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "http_server" {
  name = "http-server-pipeline"

  config {
    sources {
      http_server {
        id            = "http-source-1"
        auth_strategy = "plain"
        decoding      = "json"

        tls {
          crt_file = "/etc/ssl/certs/http.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/http.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["http-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "http-server-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_server.0.id", "http-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_server.0.auth_strategy", "plain"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_server.0.decoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_server.0.tls.crt_file", "/etc/ssl/certs/http.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_server.0.tls.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_server.0.tls.key_file", "/etc/ssl/private/http.key"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.inputs.0", "http-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_amazonS3Source(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.s3_source"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "s3_source" {
  name = "amazon_s3-source-pipeline"

  config {
    sources {
      amazon_s3 {
        id     = "s3-source-1"
        region = "us-east-1"

        auth {
          assume_role  = "arn:aws:iam::123456789012:role/test-role"
          external_id  = "external-test-id"
          session_name = "session-test"
        }

        tls {
          crt_file = "/etc/ssl/certs/s3.crt"
          ca_file  = "/etc/ssl/certs/s3.ca"
          key_file = "/etc/ssl/private/s3.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["s3-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.id", "s3-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.auth.assume_role", "arn:aws:iam::123456789012:role/test-role"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.auth.external_id", "external-test-id"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.auth.session_name", "session-test"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.tls.crt_file", "/etc/ssl/certs/s3.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.tls.ca_file", "/etc/ssl/certs/s3.ca"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_s3.0.tls.key_file", "/etc/ssl/private/s3.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_splunkHecSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.splunk_hec"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "splunk_hec" {
  name = "splunk-hec-pipeline"

  config {
    sources {
      splunk_hec {
        id = "splunk-hec-source-1"

        tls {
          crt_file = "/etc/ssl/certs/splunk.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/splunk.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["splunk-hec-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_hec.0.id", "splunk-hec-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_hec.0.tls.crt_file", "/etc/ssl/certs/splunk.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_hec.0.tls.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_hec.0.tls.key_file", "/etc/ssl/private/splunk.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_splunkTcpSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.splunk_tcp"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "splunk_tcp" {
  name = "splunk-tcp-pipeline"

  config {
    sources {
      splunk_tcp {
        id = "splunk-tcp-source-1"

        tls {
          crt_file = "/etc/ssl/certs/tcp.crt"
          ca_file  = "/etc/ssl/certs/tcp.ca"
          key_file = "/etc/ssl/private/tcp.key"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["splunk-tcp-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_tcp.0.id", "splunk-tcp-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_tcp.0.tls.crt_file", "/etc/ssl/certs/tcp.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_tcp.0.tls.ca_file", "/etc/ssl/certs/tcp.ca"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.splunk_tcp.0.tls.key_file", "/etc/ssl/private/tcp.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_generateMetricsProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.generate_metrics"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "generate_metrics" {
  name = "generate-metrics-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      generate_datadog_metrics {
        id      = "generate-metrics-1"
        include = "*"
        inputs  = ["source-1"]

        metrics {
          name        = "logs.generated"
          include     = "service:payments"
          metric_type = "count"
          group_by    = ["service", "env"]

          value {
            strategy = "increment_by_field"
            field    = "events.count"
          }
        }

        metrics {
          name        = "logs.default_count"
          include     = "service:checkout"
          metric_type = "count"

          value {
            strategy = "increment_by_one"
          }
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["generate-metrics-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.processors.generate_datadog_metrics.0.id", "generate-metrics-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.generate_datadog_metrics.0.metrics.0.name", "logs.generated"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.generate_datadog_metrics.0.metrics.0.value.strategy", "increment_by_field"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.generate_datadog_metrics.0.metrics.1.value.strategy", "increment_by_one"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googleCloudStorageDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.gcs_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "gcs_dest" {
  name = "gcs-destination-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      google_cloud_storage {
        id            = "gcs-destination-1"
        bucket        = "my-gcs-bucket"
        key_prefix    = "logs/"
        storage_class = "NEARLINE"
        acl           = "project-private"
        inputs        = ["source-1"]

        auth {
          credentials_file = "/var/secrets/gcp-creds.json"
        }

        metadata {
          name  = "environment"
          value = "production"
        }

        metadata {
          name  = "team"
          value = "platform"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.bucket", "my-gcs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.key_prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.auth.credentials_file", "/var/secrets/gcp-creds.json"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.metadata.0.name", "environment"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.metadata.0.value", "production"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.metadata.1.name", "team"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_cloud_storage.0.metadata.1.value", "platform"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_splunkHecDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.splunk_hec_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "splunk_hec_dest" {
  name = "splunk-hec-destination-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      splunk_hec {
        id     = "splunk-hec-1"
        inputs = ["source-1"]

        auto_extract_timestamp = true
        encoding               = "json"
        sourcetype             = "custom_sourcetype"
        index                  = "main"
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.splunk_hec.0.id", "splunk-hec-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.splunk_hec.0.auto_extract_timestamp", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.splunk_hec.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.splunk_hec.0.sourcetype", "custom_sourcetype"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.splunk_hec.0.index", "main"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_sumoLogicDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.sumo"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "sumo" {
  name = "sumo pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      sumo_logic {
        id     = "sumo-dest-1"
        inputs = ["source-1"]
        encoding = "json"
        header_host_name = "host-123"
        header_source_name = "source-name"
        header_source_category = "source-category"

        header_custom_fields {
          name  = "X-Sumo-Category"
          value = "my-app-logs"
        }

        header_custom_fields {
          name  = "X-Custom-Header"
          value = "debug=true"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "sumo pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.id", "sumo-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_host_name", "host-123"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_source_name", "source-name"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_source_category", "source-category"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_custom_fields.0.name", "X-Sumo-Category"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_custom_fields.0.value", "my-app-logs"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_custom_fields.1.name", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sumo_logic.0.header_custom_fields.1.value", "debug=true"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_rsyslogSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.rsyslog_source"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "rsyslog_source" {
  name = "rsyslog-source-pipeline"

  config {
    sources {
      rsyslog {
        id   = "rsyslog-source-1"
        mode = "tcp"

        tls {
          crt_file = "/etc/certs/rsyslog.crt"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["rsyslog-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "rsyslog-source-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.rsyslog.0.id", "rsyslog-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.rsyslog.0.mode", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.rsyslog.0.tls.crt_file", "/etc/certs/rsyslog.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.inputs.0", "rsyslog-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_syslogNgSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.syslogng_source"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "syslogng_source" {
  name = "syslogng-source-pipeline"

  config {
    sources {
      syslog_ng {
        id   = "syslogng-source-1"
        mode = "udp"

        tls {
          crt_file = "/etc/certs/syslogng.crt"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["syslogng-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "syslogng-source-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.syslog_ng.0.id", "syslogng-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.syslog_ng.0.mode", "udp"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.syslog_ng.0.tls.crt_file", "/etc/certs/syslogng.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.inputs.0", "syslogng-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_rsyslogDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.rsyslog_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "rsyslog_dest" {
  name = "rsyslog-dest-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      rsyslog {
        id     = "rsyslog-destination-1"
        inputs = ["source-1"]
        keepalive = 60000

        tls {
          crt_file = "/etc/certs/rsyslog.crt"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "rsyslog-dest-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.rsyslog.0.id", "rsyslog-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.rsyslog.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.rsyslog.0.keepalive", "60000"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.rsyslog.0.tls.crt_file", "/etc/certs/rsyslog.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_syslogNgDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.syslogng_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "syslogng_dest" {
  name = "syslogng-dest-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      syslog_ng {
        id     = "syslogng-destination-1"
        inputs = ["source-1"]
        keepalive = 45000

        tls {
          crt_file = "/etc/certs/syslogng.crt"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "syslogng-dest-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.syslog_ng.0.id", "syslogng-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.syslog_ng.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.syslog_ng.0.keepalive", "45000"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.syslog_ng.0.tls.crt_file", "/etc/certs/syslogng.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_elasticsearchDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.elasticsearch_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "elasticsearch_dest" {
  name = "elasticsearch-dest-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      elasticsearch {
        id          = "elasticsearch-destination-1"
        inputs      = ["source-1"]
        api_version = "v7"
        bulk_index  = "logs-datastream"
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "elasticsearch-dest-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.elasticsearch.0.id", "elasticsearch-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.elasticsearch.0.api_version", "v7"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.elasticsearch.0.bulk_index", "logs-datastream"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.elasticsearch.0.inputs.0", "source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_azureStorageDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.azure_storage_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "azure_storage_dest" {
  name = "azure-storage-dest-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      azure_storage {
        id             = "azure-storage-destination-1"
        inputs         = ["source-1"]
        container_name = "logs-container"
        blob_prefix    = "logs/"
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "azure-storage-dest-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.azure_storage.0.id", "azure-storage-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.azure_storage.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.azure_storage.0.container_name", "logs-container"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.azure_storage.0.blob_prefix", "logs/"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_microsoftSentinelDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.sentinel"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "sentinel" {
  name = "sentinel-pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      microsoft_sentinel {
        id               = "sentinel-dest-1"
        inputs           = ["source-1"]
        client_id        = "a1b2c3d4-5678-90ab-cdef-1234567890ab"
        tenant_id        = "abcdef12-3456-7890-abcd-ef1234567890"
        dcr_immutable_id = "dcr-uuid-1234"
        table            = "CustomLogsTable"
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "sentinel-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.microsoft_sentinel.0.id", "sentinel-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.microsoft_sentinel.0.client_id", "a1b2c3d4-5678-90ab-cdef-1234567890ab"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.microsoft_sentinel.0.tenant_id", "abcdef12-3456-7890-abcd-ef1234567890"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.microsoft_sentinel.0.dcr_immutable_id", "dcr-uuid-1234"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.microsoft_sentinel.0.table", "CustomLogsTable"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_sensitiveDataScannerProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.sds"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "sds" {
  name = "sds-full-test"

  config {
    sources {
      datadog_agent {
        id = "source-a"
      }
    }

    processors {
      sensitive_data_scanner {
        id      = "sds-1"
        include = "*"
        inputs  = ["source-a"]

        rules {
          name = "Redact with Exclude"
          tags = ["confidential", "mask"]

          keyword_options {
            keywords  = ["secret", "token"]
            proximity = 4
          }

          pattern {
            custom {
              rule = "\\bsecret-[a-z0-9]+\\b"
            }
          }

          scope {
            exclude {
              fields = ["not_this_field"]
            }
          }

          on_match {
            redact {
              replace = "[REDACTED]"
            }
          }
        }

        rules {
          name = "Library Hash"
          tags = ["pii"]

          pattern {
            library {
              id = "ip_address"
              use_recommended_keywords = true
            }
          }

          scope {
            all = true
          }

          on_match {
            hash {}
          }
        }

        rules {
          name = "Partial Default Scope"
          tags = ["user", "pii"]

          pattern {
            custom {
              rule = "user\\d{3,}"
            }
          }

		  scope {
            include {
              fields = ["this_field_only"]
            }
          }

          on_match {
            partial_redact {
              characters = 3
              direction  = "first"
            }
          }
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "sink"
        inputs = ["sds-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					// Processor metadata
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.id", "sds-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.inputs.0", "source-a"),

					// Rule 0 - Full custom + exclude + redact
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.name", "Redact with Exclude"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.tags.0", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.tags.1", "mask"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.keyword_options.keywords.0", "secret"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.keyword_options.keywords.1", "token"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.keyword_options.proximity", "4"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.pattern.custom.rule", "\\bsecret-[a-z0-9]+\\b"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.scope.exclude.fields.0", "not_this_field"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.0.on_match.redact.replace", "[REDACTED]"),

					// Rule 1 - Library + all + hash
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.1.name", "Library Hash"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.1.pattern.library.id", "ip_address"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.1.pattern.library.use_recommended_keywords", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.1.scope.all", "true"),

					// Rule 2 - Partial redact
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.2.name", "Partial Default Scope"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.2.pattern.custom.rule", "user\\d{3,}"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.2.on_match.partial_redact.characters", "3"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.2.on_match.partial_redact.direction", "first"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.sensitive_data_scanner.0.rules.2.scope.include.fields.0", "this_field_only"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_sumoLogicSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.sumo_source"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "sumo_source" {
  name = "sumo-source-pipeline"

  config {
    sources {
      sumo_logic {
        id = "sumo-logic-source-1"
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "logs-1"
        inputs = ["sumo-logic-source-1"]
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "sumo-source-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.sumo_logic.0.id", "sumo-logic-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.inputs.0", "sumo-logic-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_amazonDataFirehoseSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.firehose_test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "firehose_test" {
  name = "firehose pipeline"

  config {
    sources {
      amazon_data_firehose {
        id = "firehose-source-1"
        auth {
          assume_role = "arn:aws:iam::123456789012:role/ExampleRole"
          external_id = "external-id-123"
          session_name = "firehose-session"
        }
        tls {
          crt_file = "/path/to/firehose.crt"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["firehose-source-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "firehose pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_data_firehose.0.id", "firehose-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_data_firehose.0.auth.assume_role", "arn:aws:iam::123456789012:role/ExampleRole"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_data_firehose.0.auth.external_id", "external-id-123"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_data_firehose.0.auth.session_name", "firehose-session"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.amazon_data_firehose.0.tls.crt_file", "/path/to/firehose.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_httpClientSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.http_client"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "http_client" {
  name = "http-client pipeline"

  config {
    sources {
      http_client {
        id                  = "http-source-1"
        decoding            = "json"
        scrape_interval_secs = 60
        scrape_timeout_secs  = 10
        auth_strategy       = "basic"
        tls {
          crt_file = "/path/to/http.crt"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["http-source-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_client.0.id", "http-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_client.0.decoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_client.0.scrape_interval_secs", "60"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_client.0.scrape_timeout_secs", "10"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_client.0.auth_strategy", "basic"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.http_client.0.tls.crt_file", "/path/to/http.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googlePubSubSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.pubsub"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "pubsub" {
  name = "pubsub pipeline"

  config {
    sources {
      google_pubsub {
        id           = "pubsub-source-1"
        project      = "my-gcp-project"
        subscription = "logs-subscription"
        decoding     = "json"

        auth {
          credentials_file = "/secrets/creds.json"
        }

        tls {
          crt_file = "/certs/pubsub.crt"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["pubsub-source-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.sources.google_pubsub.0.id", "pubsub-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.google_pubsub.0.project", "my-gcp-project"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.google_pubsub.0.subscription", "logs-subscription"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.google_pubsub.0.decoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.google_pubsub.0.auth.credentials_file", "/secrets/creds.json"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.google_pubsub.0.tls.crt_file", "/certs/pubsub.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_logstashSource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.logstash"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "logstash" {
  name = "logstash pipeline"

  config {
    sources {
      logstash {
        id = "logstash-source-1"
        tls {
          crt_file = "/path/to/logstash.crt"
        }
      }
    }

    processors {}

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["logstash-source-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.sources.logstash.0.id", "logstash-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.logstash.0.tls.crt_file", "/path/to/logstash.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_dedupeProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.dedupe"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "dedupe" {
  name = "dedupe pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      dedupe {
        id      = "dedupe-match"
        include = "*"
        inputs  = ["source-1"]
        fields  = ["log.message", "log.tags"]
        mode    = "match"
      }

      dedupe {
        id      = "dedupe-ignore"
        include = "*"
        inputs  = ["dedupe-match"]
        fields  = ["log.source", "log.context"]
        mode    = "ignore"
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["dedupe-ignore"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.processors.dedupe.0.id", "dedupe-match"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.dedupe.0.mode", "match"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.dedupe.1.id", "dedupe-ignore"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.dedupe.1.mode", "ignore"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.dedupe.0.fields.0", "log.message"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.dedupe.1.fields.1", "log.context"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_reduceProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.reduce"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "reduce" {
  name = "reduce pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      reduce {
        id       = "reduce-1"
        include  = "env:prod"
        inputs   = ["source-1"]
        group_by = ["log.user.id", "log.device.id"]

        merge_strategies {
          path     = "log.user.roles"
          strategy = "flat_unique"
        }

        merge_strategies {
          path     = "log.error.messages"
          strategy = "concat"
        }

        merge_strategies {
          path     = "log.count"
          strategy = "sum"
        }

        merge_strategies {
          path     = "log.status"
          strategy = "retain"
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["reduce-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.id", "reduce-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.include", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.group_by.0", "log.user.id"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.group_by.1", "log.device.id"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.merge_strategies.0.strategy", "flat_unique"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.merge_strategies.1.strategy", "concat"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.merge_strategies.2.strategy", "sum"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.reduce.0.merge_strategies.3.strategy", "retain"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_throttleProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.throttle"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "throttle" {
  name = "throttle pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      throttle {
        id        = "throttle-global"
        include   = "*"
        inputs    = ["source-1"]
        threshold = 1000
        window    = 60.0
      }

      throttle {
        id        = "throttle-grouped"
        include   = "*"
        inputs    = ["throttle-global"]
        threshold = 100
        window    = 10.0
        group_by  = ["log.user.id", "log.level"]
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["throttle-grouped"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.processors.throttle.0.id", "throttle-global"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.throttle.0.threshold", "1000"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.throttle.1.id", "throttle-grouped"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.throttle.1.group_by.0", "log.user.id"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.throttle.1.group_by.1", "log.level"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_addEnvVarsProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.add_env_vars"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "add_env_vars" {
  name = "add-env-vars pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      add_env_vars {
        id      = "add-env-1"
        include = "*"
        inputs  = ["source-1"]

        variables {
          field = "log.environment.region"
          name  = "AWS_REGION"
        }

        variables {
          field = "log.environment.account"
          name  = "AWS_ACCOUNT_ID"
        }

      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["add-env-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_env_vars.0.id", "add-env-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_env_vars.0.variables.0.name", "AWS_REGION"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_env_vars.0.variables.1.name", "AWS_ACCOUNT_ID"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_enrichmentTableProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.enrichment"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "enrichment" {
  name = "enrichment pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      enrichment_table {
        id      = "csv-enrichment"
        include = "*"
        inputs  = ["source-1"]
        target  = "log.geo.csv"

        file {
          path = "/etc/enrichment/lookup.csv"

          encoding {
            type             = "csv"
            delimiter        = ","
            includes_headers = true
          }

          schema {
            column = "region"
            type   = "string"
          }

          schema {
            column = "city"
            type   = "string"
          }

          key {
            column     = "user_id"
            comparison = "equals"
            field      = "log.user.id"
          }
        }
      }

      enrichment_table {
        id      = "geoip-enrichment"
        include = "*"
        inputs  = ["csv-enrichment"]
        target  = "log.geo.geoip"

        geoip {
          key_field = "log.source.ip"
          locale    = "en"
          path      = "/etc/geoip/GeoLite2-City.mmdb"
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["geoip-enrichment"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					// CSV enrichment checks
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.0.id", "csv-enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.0.file.path", "/etc/enrichment/lookup.csv"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.0.file.encoding.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.0.file.schema.0.column", "region"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.0.file.key.0.field", "log.user.id"),
					// GeoIP enrichment checks
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.1.id", "geoip-enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.1.geoip.key_field", "log.source.ip"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.1.geoip.locale", "en"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.enrichment_table.1.geoip.path", "/etc/geoip/GeoLite2-City.mmdb"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googleChronicleDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.chronicle"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "chronicle" {
  name = "chronicle pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

	processors {}

    destinations {
      google_chronicle {
        id          = "chronicle-dest-1"
        inputs      = ["source-1"]
        customer_id = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
        encoding    = "json"
        log_type    = "nginx_logs"

        auth {
          credentials_file = "/secrets/gcp.json"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_chronicle.0.id", "chronicle-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_chronicle.0.customer_id", "3fa85f64-5717-4562-b3fc-2c963f66afa6"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_chronicle.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_chronicle.0.log_type", "nginx_logs"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.google_chronicle.0.auth.credentials_file", "/secrets/gcp.json"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_newRelicDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.newrelic"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "newrelic" {
  name = "newrelic pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      new_relic {
        id     = "newrelic-dest-1"
        inputs = ["source-1"]
        region = "us"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.new_relic.0.id", "newrelic-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.new_relic.0.region", "us"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_sentinelOneDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.sentinelone"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "sentinelone" {
  name = "sentinelone pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

	processors {}

    destinations {
      sentinel_one {
        id     = "sentinelone-dest-1"
        inputs = ["source-1"]
        region = "us"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sentinel_one.0.id", "sentinelone-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.sentinel_one.0.region", "us"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_ocsfMapperLibraryOnly(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.ocsf"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "ocsf" {
  name = "ocsf mapper (library only)"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      ocsf_mapper {
        id      = "ocsf-mapper"
        include = "*"
        inputs  = ["source-1"]

        mapping {
          include         = "source:lib"
          library_mapping = "CloudTrail Account Change"
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "destination-1"
        inputs = ["ocsf-mapper"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.processors.ocsf_mapper.0.id", "ocsf-mapper"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.ocsf_mapper.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.ocsf_mapper.0.mapping.0.include", "source:lib"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.ocsf_mapper.0.mapping.0.library_mapping", "CloudTrail Account Change"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_opensearchDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.opensearch"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "opensearch" {
  name = "opensearch pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      opensearch {
        id         = "opensearch-dest-1"
        inputs     = ["source-1"]
        bulk_index = "logs-datastream"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.opensearch.0.id", "opensearch-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.opensearch.0.bulk_index", "logs-datastream"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.opensearch.0.inputs.0", "source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_amazonOpenSearchDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.amazon_opensearch"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "amazon_opensearch" {
  name = "amazon opensearch pipeline"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      amazon_opensearch {
        id         = "aos-dest-1"
        inputs     = ["source-1"]
        bulk_index = "logs-datastream"

        auth {
          strategy     = "aws"
          aws_region   = "us-east-1"
          assume_role  = "arn:aws:iam::123456789012:role/example-role"
          external_id  = "external-id-123"
          session_name = "aos-session"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.id", "aos-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.bulk_index", "logs-datastream"),

					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.auth.strategy", "aws"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.auth.aws_region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.auth.assume_role", "arn:aws:iam::123456789012:role/example-role"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.auth.external_id", "external-id-123"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.auth.session_name", "aos-session"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_amazonOpenSearchDestination_basic(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.amazon_opensearch_basic"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "amazon_opensearch_basic" {
  name = "amazon opensearch (basic auth)"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {}

    destinations {
      amazon_opensearch {
        id         = "aos-basic-1"
        inputs     = ["source-1"]

        auth {
          strategy = "basic"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.id", "aos-basic-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.amazon_opensearch.0.auth.strategy", "basic"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_quotaProcessor_overflowAction(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.quota"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "quota" {
  name = "quota with overflow_action"

  config {
    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      quota {
        id      = "quota-1"
        include = "*"
        name    = "MyQuota"
        drop_events = false
        inputs  = ["source-1"]

        overflow_action = "drop"

        limit {
          enforce = "events"
          limit   = 1000
        }
      }
    }

    destinations {
      datadog_logs {
        id     = "logs-1"
        inputs = ["quota-1"]
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.id", "quota-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.name", "MyQuota"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.overflow_action", "drop"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.limit.enforce", "events"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.quota.0.limit.limit", "1000"),
				),
			},
		},
	})
}
