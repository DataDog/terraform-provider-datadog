package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.datadog_agent.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "parser-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "service:my-service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.display_name", "processor group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "parser-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "service:my-service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.display_name", "json parser"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_json.0.field", "message"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "parser-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
				),
			},
			{
				Config: testAccObservabilityPipelineUpdatedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "updated pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "service:updated-service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "service:updated-service"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineBasicConfig() string {
	return `
resource "datadog_observability_pipeline" "basic" {
  name = "test pipeline"

  config {
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "parser-group-1"
      enabled = true
      include = "service:my-service"
      inputs  = ["source-1"]
      display_name = "processor group"

      processor {
        id      = "parser-1"
        enabled = true
        include = "service:my-service"
        display_name = "json parser"
        
        parse_json {
          field = "message"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["parser-group-1"]
      
      datadog_logs {
      }
    }
  }
}`
}

func testAccObservabilityPipelineUpdatedConfig() string {
	return `
resource "datadog_observability_pipeline" "basic" {
  name = "updated pipeline"

  config {

    source {
      id = "source-1" 
      datadog_agent {
      }
    }

    processor_group {
      id      = "parser-group-1"
      enabled = true
      include = "service:updated-service"
      inputs  = ["source-1"]
      
      processor {
        id      = "parser-1"
        enabled = true
        include = "service:updated-service"
        
        parse_json {
          field = "message"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["parser-group-1"]
      datadog_logs {
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
    source {
      id = "kafka-source-1"
      
      kafka {
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
    
    destination {
      id     = "destination-1"
      inputs = ["kafka-source-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "kafka pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "kafka-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.kafka.0.group_id", "consumer-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.kafka.0.topics.0", "topic-a"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.kafka.0.topics.1", "topic-b"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.kafka.0.sasl.0.mechanism", "PLAIN"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "kafka-source-1"),
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
    source {
      id = "source-with-tls"
      
      datadog_agent {
        tls {
          crt_file = "/etc/certs/agent.crt"
          ca_file  = "/etc/certs/ca.crt"
          key_file = "/etc/certs/agent.key"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["source-with-tls"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "agent with tls"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-with-tls"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.datadog_agent.0.tls.0.crt_file", "/etc/certs/agent.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.datadog_agent.0.tls.0.ca_file", "/etc/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.datadog_agent.0.tls.0.key_file", "/etc/certs/agent.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-with-tls"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "filter-group-1"
      enabled = true
      include = "env:prod"
      inputs  = ["source-1"]
      
      processor {
        id      = "filter-1"
        enabled = true
        include = "env:prod"
        
        filter {}
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["filter-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "filter-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "filter-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "filter-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "env:prod"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "rename-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "rename-1"
        enabled = true
        include = "*"
        
        rename_fields {
          field {
            source          = "old.field"
            destination     = "new.field"
            preserve_source = true
          }
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["rename-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "rename-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "rename-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "rename-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.rename_fields.0.field.0.source", "old.field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.rename_fields.0.field.0.destination", "new.field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.rename_fields.0.field.0.preserve_source", "true"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "remove-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "remove-1"
        enabled = true
        include = "*"
        
        remove_fields {
          fields = ["temp.debug", "internal.trace_id"]
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["remove-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "remove-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "remove-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "remove-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.remove_fields.0.fields.0", "temp.debug"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.remove_fields.0.fields.1", "internal.trace_id"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "quota-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "quota-1"
        enabled = true
        include = "*"
        
          quota {
            name    = "limitByHostAndEnv"
            drop_events = true
            ignore_when_missing_partitions = true
            partition_fields = ["host", "env"]

            limit {
              enforce = "events"
              limit   = 1000
            }

            override {
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

            override {
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
    }
    
    destination {
      id     = "destination-1"
      inputs = ["quota-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "quota-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "quota-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "quota-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.name", "limitByHostAndEnv"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.drop_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.limit.0.enforce", "events"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.limit.0.limit", "1000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.0.field.0.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.0.field.0.value", "prod"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.0.field.1.name", "host"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.0.field.1.value", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.0.limit.0.enforce", "events"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.0.limit.0.limit", "500"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.1.field.0.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.1.field.0.value", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.1.field.1.name", "host"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.1.field.1.value", "localhost"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.1.limit.0.enforce", "bytes"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.override.1.limit.0.limit", "300"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "parser-group-1"
      enabled = true
      include = "env:parse"
      inputs  = ["source-1"]
      
      processor {
        id      = "parser-1"
        enabled = true
        include = "env:parse"
        
        parse_json {
          field = "message"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["parser-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "parse-json-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "parser-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "env:parse"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "parser-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_json.0.field", "message"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "add-fields-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "add-fields-1"
        enabled = true
        include = "*"
        
        add_fields {
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
    }
    
    destination {
      id     = "destination-1"
      inputs = ["add-fields-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "add-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "add-fields-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "add-fields-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_fields.0.field.0.name", "custom.field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_fields.0.field.0.value", "hello-world"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_fields.0.field.1.name", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_fields.0.field.1.value", "prod"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "parse-grok-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "parse-grok-1"
        enabled = true
        include = "*"
        
        parse_grok {
          disable_library_rules = true

          rule {
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
    }
    
    destination {
      id     = "destination-1"
      inputs = ["parse-grok-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "parse-grok-test"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "parse-grok-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "parse-grok-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.disable_library_rules", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.source", "message"),

					// Match Rules
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.match_rule.0.name", "match_user"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.match_rule.0.rule", "%{word:user.name}"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.match_rule.1.name", "match_action"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.match_rule.1.rule", "%{word:action}"),

					// Support Rules
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.support_rule.0.name", "word"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.support_rule.0.rule", "\\w+"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.support_rule.1.name", "custom_word"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_grok.0.rule.0.support_rule.1.rule", "[a-zA-Z]+"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "sample-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "sample-1"
        enabled = true
        include = "*"
        
        sample {
          percentage = 10
          group_by   = ["service", "host"]
        }
      }
    }

    processor_group {
      id      = "sample-group-2"
      enabled = false
      include = "*"
      inputs  = ["sample-group-1"]
      
      processor {
        id      = "sample-2"
        enabled = false
        include = "*"
        
        sample {
          percentage = 4.99
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["sample-group-2"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "sample-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "sample-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "sample-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sample.0.percentage", "10"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sample.0.group_by.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sample.0.group_by.0", "service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sample.0.group_by.1", "host"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "sample-group-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "sample-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.sample.0.percentage", "4.99"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.sample.0.group_by.#", "0"),
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
    source {
      id = "fluent-source-1"
      
      fluentd {
        tls {
          crt_file = "/etc/ssl/certs/fluent.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/fluent.key"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["fluent-source-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "fluent-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "fluent-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.fluentd.0.tls.0.crt_file", "/etc/ssl/certs/fluent.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.fluentd.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.fluentd.0.tls.0.key_file", "/etc/ssl/private/fluent.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "fluent-source-1"),
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
    source {
      id = "fluent-source-1"
      
      fluent_bit {
        tls {
          crt_file = "/etc/ssl/certs/fluent.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/fluent.key"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["fluent-source-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "fluent-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "fluent-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.fluent_bit.0.tls.0.crt_file", "/etc/ssl/certs/fluent.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.fluent_bit.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.fluent_bit.0.tls.0.key_file", "/etc/ssl/private/fluent.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "fluent-source-1"),
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
    source {
      id = "http-source-1"
      
      http_server {
        auth_strategy = "plain"
        decoding      = "json"

        tls {
          crt_file = "/etc/ssl/certs/http.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/http.key"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["http-source-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "http-server-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "http-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_server.0.auth_strategy", "plain"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_server.0.decoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_server.0.tls.0.crt_file", "/etc/ssl/certs/http.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_server.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_server.0.tls.0.key_file", "/etc/ssl/private/http.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "http-source-1"),
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
    source {
      id = "s3-source-1"
      
      amazon_s3 {
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
    
    destination {
      id     = "destination-1"
      inputs = ["s3-source-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "s3-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.auth.0.assume_role", "arn:aws:iam::123456789012:role/test-role"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.auth.0.external_id", "external-test-id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.auth.0.session_name", "session-test"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.tls.0.crt_file", "/etc/ssl/certs/s3.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.tls.0.ca_file", "/etc/ssl/certs/s3.ca"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_s3.0.tls.0.key_file", "/etc/ssl/private/s3.key"),
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
    source {
      id = "splunk-hec-source-1"
      
      splunk_hec {
        tls {
          crt_file = "/etc/ssl/certs/splunk.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/splunk.key"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["splunk-hec-source-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "splunk-hec-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.splunk_hec.0.tls.0.crt_file", "/etc/ssl/certs/splunk.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.splunk_hec.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.splunk_hec.0.tls.0.key_file", "/etc/ssl/private/splunk.key"),
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
    source {
      id = "splunk-tcp-source-1"
      
      splunk_tcp {
        tls {
          crt_file = "/etc/ssl/certs/tcp.crt"
          ca_file  = "/etc/ssl/certs/tcp.ca"
          key_file = "/etc/ssl/private/tcp.key"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["splunk-tcp-source-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "splunk-tcp-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.splunk_tcp.0.tls.0.crt_file", "/etc/ssl/certs/tcp.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.splunk_tcp.0.tls.0.ca_file", "/etc/ssl/certs/tcp.ca"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.splunk_tcp.0.tls.0.key_file", "/etc/ssl/private/tcp.key"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "generate-metrics-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "generate-metrics-1"
        enabled = true
        include = "*"
        
        generate_datadog_metrics {
          metric {
            name        = "logs.generated"
            include     = "service:payments"
            metric_type = "count"
            group_by    = ["service", "env"]

            value {
              strategy = "increment_by_field"
              field    = "events.count"
            }
          }

          metric {
            name        = "logs.default_count"
            include     = "service:checkout"
            metric_type = "count"

            value {
              strategy = "increment_by_one"
            }
          }
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["generate-metrics-group-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "generate-metrics-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "generate-metrics-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.generate_datadog_metrics.0.metric.0.name", "logs.generated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.generate_datadog_metrics.0.metric.0.value.0.strategy", "increment_by_field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.generate_datadog_metrics.0.metric.1.value.0.strategy", "increment_by_one"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id = "gcs-destination-1"
      inputs = ["source-1"]
      google_cloud_storage {
        bucket        = "my-gcs-bucket"
        key_prefix    = "logs/"
        storage_class = "NEARLINE"
        acl           = "project-private"

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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "gcs-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.bucket", "my-gcs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.key_prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.auth.0.credentials_file", "/var/secrets/gcp-creds.json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.metadata.0.name", "environment"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.metadata.0.value", "production"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.metadata.1.name", "team"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.metadata.1.value", "platform"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googleCloudStorageDestinationMinimal(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.gcs_dest_minimal"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "gcs_dest_minimal" {
  name = "gcs-destination-minimal-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id = "gcs-destination-1"
      inputs = ["source-1"]
      google_cloud_storage {
        bucket = "my-gcs-bucket"
        storage_class = "NEARLINE"
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.bucket", "my-gcs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_cloud_storage.0.storage_class", "NEARLINE"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id = "splunk-hec-1"
      inputs = ["source-1"]
      splunk_hec {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "splunk-hec-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.splunk_hec.0.auto_extract_timestamp", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.splunk_hec.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.splunk_hec.0.sourcetype", "custom_sourcetype"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.splunk_hec.0.index", "main"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }
    
    destination {
      id     = "sumo-dest-1"
      inputs = ["source-1"]
      
      sumo_logic {

        encoding = "json"
        header_host_name = "host-123"
        header_source_name = "source-name"
        header_source_category = "source-category"

        header_custom_field {
          name  = "X-Sumo-Category"
          value = "my-app-logs"
        }

        header_custom_field {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "sumo-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_host_name", "host-123"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_source_name", "source-name"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_source_category", "source-category"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_custom_field.0.name", "X-Sumo-Category"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_custom_field.0.value", "my-app-logs"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_custom_field.1.name", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sumo_logic.0.header_custom_field.1.value", "debug=true"),
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
    source {
      id = "rsyslog-source-1"
      
      rsyslog {
        mode = "tcp"

        tls {
          crt_file = "/etc/certs/rsyslog.crt"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["rsyslog-source-1"]
        
      datadog_logs {}
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "rsyslog-source-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "rsyslog-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.rsyslog.0.mode", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.rsyslog.0.tls.0.crt_file", "/etc/certs/rsyslog.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "rsyslog-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
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
    source {
      id = "syslogng-source-1"
      syslog_ng {
        mode = "udp"

        tls {
          crt_file = "/etc/certs/syslogng.crt"
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["syslogng-source-1"]
      
      datadog_logs {}
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "syslogng-source-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "syslogng-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.syslog_ng.0.mode", "udp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.syslog_ng.0.tls.0.crt_file", "/etc/certs/syslogng.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "syslogng-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "rsyslog-destination-1"
      inputs = ["source-1"]

      rsyslog {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "rsyslog-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.rsyslog.0.keepalive", "60000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.rsyslog.0.tls.0.crt_file", "/etc/certs/rsyslog.crt"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id = "syslogng-destination-1"
      inputs = ["source-1"]

      syslog_ng {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "syslogng-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.syslog_ng.0.keepalive", "45000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.syslog_ng.0.tls.0.crt_file", "/etc/certs/syslogng.crt"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "elasticsearch-destination-1"
      inputs = ["source-1"]
            
      elasticsearch {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "elasticsearch-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.api_version", "v7"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.bulk_index", "logs-datastream"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "azure-storage-destination-1"
      inputs = ["source-1"]

      azure_storage {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "azure-storage-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.azure_storage.0.container_name", "logs-container"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.azure_storage.0.blob_prefix", "logs/"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "sentinel-dest-1"
      inputs = ["source-1"]

      microsoft_sentinel {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "sentinel-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.microsoft_sentinel.0.client_id", "a1b2c3d4-5678-90ab-cdef-1234567890ab"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.microsoft_sentinel.0.tenant_id", "abcdef12-3456-7890-abcd-ef1234567890"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.microsoft_sentinel.0.dcr_immutable_id", "dcr-uuid-1234"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.microsoft_sentinel.0.table", "CustomLogsTable"),
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
    source {
      id = "source-a"
      datadog_agent {
      }
    }
        
    processor_group {
      id      = "sds-group-1"
      inputs  = ["source-a"]
      enabled = true
      include = "*"

      processor {
        id = "sds-1"
        include = "*"
        enabled = true

        sensitive_data_scanner {

          rule {
            name = "Redact with Exclude"
            tags = ["confidential", "mask"]

            keyword_options {
              keywords  = ["secret", "token"]
              proximity = 6
            }

            pattern {
              custom {
                rule = "\\bsecret-[a-z0-9]+\\b"
                description = "secret"
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

          rule {
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

          rule {
            name = "Partial Default Scope No Tags"
            tags = []

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

          rule {
            name = "Library Empty Tags No Keywords Last Direction"
            tags = []

            pattern {
              library {
                id = "email_address"
                description = "email address"
              }
            }

            scope {
              exclude {
                fields = ["excluded_field"]
              }
            }

            on_match {
              partial_redact {
                characters = 5
                direction  = "last"
              }
            }
          }
        }
      }
    }
    
    destination {
      id     = "sink"
      inputs = ["sds-group-1"]
      
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					// Pipeline metadata
					resource.TestCheckResourceAttr(resourceName, "name", "sds-full-test"),

					// Source
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-a"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.datadog_agent.#", "1"),

					// Processor group metadata
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "sds-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.inputs.0", "source-a"),

					// Processor metadata
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "sds-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),

					// Rule 0 - Custom pattern + keywords + exclude scope + redact action
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.name", "Redact with Exclude"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.tags.0", "confidential"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.tags.1", "mask"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.keyword_options.0.keywords.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.keyword_options.0.keywords.0", "secret"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.keyword_options.0.keywords.1", "token"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.keyword_options.0.proximity", "6"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.pattern.0.custom.0.rule", "\\bsecret-[a-z0-9]+\\b"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.scope.0.exclude.0.fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.scope.0.exclude.0.fields.0", "not_this_field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.0.on_match.0.redact.0.replace", "[REDACTED]"),

					// Rule 1 - Library pattern + use_recommended_keywords + all scope + hash action
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.1.name", "Library Hash"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.1.tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.1.tags.0", "pii"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.1.pattern.0.library.0.id", "ip_address"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.1.pattern.0.library.0.use_recommended_keywords", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.1.scope.0.all", "true"),

					// Rule 2 - Custom pattern + include scope + partial_redact (first)
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.name", "Partial Default Scope No Tags"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.pattern.0.custom.0.rule", "user\\d{3,}"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.scope.0.include.0.fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.scope.0.include.0.fields.0", "this_field_only"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.on_match.0.partial_redact.0.characters", "3"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.2.on_match.0.partial_redact.0.direction", "first"),

					// Rule 3 - Library pattern + no keywords + empty tags + exclude scope + partial_redact (last)
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.name", "Library Empty Tags No Keywords Last Direction"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.pattern.0.library.0.id", "email_address"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.pattern.0.library.0.use_recommended_keywords"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.keyword_options.#"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.scope.0.exclude.0.fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.scope.0.exclude.0.fields.0", "excluded_field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.on_match.0.partial_redact.0.characters", "5"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.sensitive_data_scanner.0.rule.3.on_match.0.partial_redact.0.direction", "last"),

					// Destination
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "sink"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "sds-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_multipleProcessorGroups(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.multi_groups"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "multi_groups" {
  name = "multiple-processor-groups-test"

  config {
    source {
      id = "my-source"
      datadog_agent {
      }
    }

    processor_group {
      id = "filter-group"
      enabled = true
      include = "*"
      inputs = ["my-source"]

      processor {
        id = "filter-1"
        enabled = true
        include = "*"

        filter {}
      }

      processor {
        id = "filter-2"
        enabled = true
        include = "*"

        remove_fields {
          fields = ["old_field", "another_field"]
        }
      }
    }

    processor_group {
      id = "remap-group"
      enabled = true
      include = "*"
      inputs = ["filter-group"]

      processor {
        id = "remap-1"
        enabled = true
        include = "*"

        add_fields {
          field {
            name = "new_field"
            value = ".old_field"
          }
        }
      }
    }

    destination {
      id = "output"
      inputs = ["remap-group"]
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					// Pipeline metadata
					resource.TestCheckResourceAttr(resourceName, "name", "multiple-processor-groups-test"),

					// Source
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "my-source"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.datadog_agent.#", "1"),

					// First processor group (filter) with 2 processors
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "filter-group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.inputs.0", "my-source"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.#", "2"),
					// First processor in group (filter-1)
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "filter-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.filter.#", "1"),
					// Second processor in group (filter-2)
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.id", "filter-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.remove_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.remove_fields.0.fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.remove_fields.0.fields.0", "old_field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.remove_fields.0.fields.1", "another_field"),

					// Second processor group (add_fields) with 1 processor
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "remap-group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.inputs.0", "filter-group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "remap-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.add_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.add_fields.0.field.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.add_fields.0.field.0.name", "new_field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.add_fields.0.field.0.value", ".old_field"),

					// Destination
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "output"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "remap-group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
				),
			},
			// reorder the processors inside the processor group and validate that the pipeline is updated correctly
			{
				Config: `
resource "datadog_observability_pipeline" "multi_groups" {
  name = "multiple-processor-groups-test"

  config {
    source {
      id = "my-source"
      datadog_agent {
      }
    }

    processor_group {
      id = "filter-group"
      enabled = true
      include = "*"
      inputs = ["my-source"]

      processor {
        id = "filter-2"
        enabled = true
        include = "*"

        remove_fields {
          fields = ["old_field", "another_field"]
        }
      }

      processor {
        id = "filter-1"
        enabled = true
        include = "*"

        filter {}
      }
    }

    processor_group {
      id = "remap-group"
      enabled = true
      include = "*"
      inputs = ["filter-group"]

      processor {
        id = "remap-1"
        enabled = true
        include = "*"

        add_fields {
          field {
            name = "new_field"
            value = ".old_field"
          }
        }
      }
    }
    
    destination {
      id = "output"
      inputs = ["remap-group"]
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					// Pipeline metadata
					resource.TestCheckResourceAttr(resourceName, "name", "multiple-processor-groups-test"),

					// First processor group with processors in swapped order
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "filter-group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.#", "2"),
					// First processor is now filter-2 (was second)
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "filter-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.remove_fields.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.remove_fields.0.fields.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.remove_fields.0.fields.0", "old_field"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.remove_fields.0.fields.1", "another_field"),
					// Second processor is now filter-1 (was first)
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.id", "filter-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.1.filter.#", "1"),

					// Second processor group unchanged
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "remap-group"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "remap-1"),

					// Destination unchanged
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "output"),
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
    source {
      id = "sumo-logic-source-1"
      sumo_logic {
      }
    }
    destination {
      id = "logs-1"
      inputs = ["sumo-logic-source-1"]
      datadog_logs {
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "sumo-source-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "sumo-logic-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "sumo-logic-source-1"),
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
    source {
      id = "firehose-source-1"
      amazon_data_firehose {
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
      
    destination {
      id = "destination-1"
      inputs = ["firehose-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "firehose pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "firehose-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_data_firehose.0.auth.0.assume_role", "arn:aws:iam::123456789012:role/ExampleRole"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_data_firehose.0.auth.0.external_id", "external-id-123"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_data_firehose.0.auth.0.session_name", "firehose-session"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.amazon_data_firehose.0.tls.0.crt_file", "/path/to/firehose.crt"),
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
    source {
      id = "http-source-1"
      http_client {
        decoding            = "json"
        scrape_interval_secs = 60
        scrape_timeout_secs  = 10
        auth_strategy       = "basic"
        tls {
          crt_file = "/path/to/http.crt"
        }
      }
    }

    destination {
      id = "destination-1"
      inputs = ["http-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "http-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_client.0.decoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_client.0.scrape_interval_secs", "60"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_client.0.scrape_timeout_secs", "10"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_client.0.auth_strategy", "basic"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.http_client.0.tls.0.crt_file", "/path/to/http.crt"),
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
    source {
      id = "pubsub-source-1"
      google_pubsub {
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

    destination {
      id = "destination-1"
      inputs = ["pubsub-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "pubsub-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.project", "my-gcp-project"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.subscription", "logs-subscription"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.decoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.auth.0.credentials_file", "/secrets/creds.json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.tls.0.crt_file", "/certs/pubsub.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googlePubSubSourceMinimal(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.pubsub_minimal"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "pubsub_minimal" {
  name = "pubsub minimal pipeline"

  config {
    source {
      id = "pubsub-source-1"
      google_pubsub {
        project      = "my-gcp-project"
        subscription = "logs-subscription"
        decoding     = "json"
      }
    }

    destination {
      id = "destination-1"
      inputs = ["pubsub-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "pubsub-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.project", "my-gcp-project"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.subscription", "logs-subscription"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.google_pubsub.0.decoding", "json"),
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
    source {
      id = "logstash-source-1"
      logstash {
        tls {
          crt_file = "/path/to/logstash.crt"
        }
      }
    }

    destination {
      id = "destination-1"
      inputs = ["logstash-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "logstash-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.logstash.0.tls.0.crt_file", "/path/to/logstash.crt"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "dedupe-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "dedupe-match"
        enabled = true
        include = "*"
        
        dedupe {
          fields = ["log.message", "log.tags"]
          mode   = "match"
        }
      }
    }

    processor_group {
      id      = "dedupe-group-2"
      enabled = true
      include = "*"
      inputs  = ["dedupe-group-1"]
      
      processor {
        id      = "dedupe-ignore"
        enabled = true
        include = "*"
        
        dedupe {
          fields = ["log.source", "log.context"]
          mode   = "ignore"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["dedupe-group-2"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "dedupe-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "dedupe-match"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.dedupe.0.mode", "match"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.dedupe.0.fields.0", "log.message"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.dedupe.0.fields.1", "log.tags"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "dedupe-group-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "dedupe-ignore"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.dedupe.0.mode", "ignore"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.dedupe.0.fields.0", "log.source"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.dedupe.0.fields.1", "log.context"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
      
    processor_group {
      id      = "reduce-group-1"
      enabled = true
      include = "env:prod"
      inputs  = ["source-1"]
        
      processor {
        id      = "reduce-1"
        enabled = true
        include = "env:prod"
        
        reduce {
          group_by = ["log.user.id", "log.device.id"]

          merge_strategy {
            path     = "log.user.roles"
            strategy = "flat_unique"
          }

          merge_strategy {
            path     = "log.error.messages"
            strategy = "concat"
          }

          merge_strategy {
            path     = "log.count"
            strategy = "sum"
          }

          merge_strategy {
            path     = "log.status"
            strategy = "retain"
          }
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["reduce-group-1"]

      datadog_logs {}
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "reduce-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "reduce-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.reduce.0.group_by.0", "log.user.id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.reduce.0.group_by.1", "log.device.id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.reduce.0.merge_strategy.0.strategy", "flat_unique"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.reduce.0.merge_strategy.1.strategy", "concat"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.reduce.0.merge_strategy.2.strategy", "sum"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.reduce.0.merge_strategy.3.strategy", "retain"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
        id      = "throttle-group-1"
        enabled = true
        include = "*"
        inputs  = ["source-1"]
        
        processor {
          id      = "throttle-global"
          enabled = true
          include = "*"
          
          throttle {
            threshold = 1000
            window    = 60.0
          }
        }
      }

    processor_group {
      id      = "throttle-group-2"
      enabled = true
      include = "*"
      inputs  = ["throttle-group-1"]
      
      processor {
        id      = "throttle-grouped"
        enabled = true
        include = "*"
        
        throttle {
          threshold = 100
          window    = 10.0
          group_by  = ["log.user.id", "log.level"]
        }
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["throttle-group-2"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "throttle-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "throttle-global"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.throttle.0.threshold", "1000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.throttle.0.window", "60"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "throttle-group-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "throttle-grouped"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.throttle.0.threshold", "100"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.throttle.0.window", "10"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.throttle.0.group_by.0", "log.user.id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.throttle.0.group_by.1", "log.level"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
        id      = "add-env-group-1"
        enabled = true
        include = "*"
        inputs  = ["source-1"]

        processor { 
          id      = "add-env-1"
          enabled = true
          include = "*"

          add_env_vars {
            variable {
              field = "log.environment.region"
              name  = "AWS_REGION"
            }

            variable {
              field = "log.environment.account"
              name  = "AWS_ACCOUNT_ID"
            }
          }
        }
      }

    destination {
      id     = "destination-1"
      inputs = ["add-env-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "add-env-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "add-env-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_env_vars.0.variable.0.name", "AWS_REGION"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_env_vars.0.variable.1.name", "AWS_ACCOUNT_ID"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
        id      = "enrichment-group-1"
        enabled = true
        include = "*"
        inputs  = ["source-1"]

        processor {
          id      = "csv-enrichment"
          enabled = true
          include = "*"

          enrichment_table {
            target = "log.enrichment"

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
        }
      }

      processor_group {
        id      = "enrichment-group-2"
        enabled = true
        include = "*"
        inputs  = ["enrichment-group-1"]

        processor {
          id      = "geoip-enrichment"
          enabled = true
          include = "*"

          enrichment_table {
            target = "log.geo.geoip"

            geoip {
              key_field = "log.source.ip"
              locale    = "en"
              path      = "/etc/geoip/GeoLite2-City.mmdb"
            }
          }
        }
      }
    destination {
      id     = "destination-1"
      inputs = ["enrichment-group-2"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					// CSV enrichment checks
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "enrichment-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "csv-enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.target", "log.enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.path", "/etc/enrichment/lookup.csv"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.encoding.0.type", "csv"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.encoding.0.delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.encoding.0.includes_headers", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.schema.0.column", "region"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.schema.0.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.schema.1.column", "city"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.schema.1.type", "string"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.key.0.column", "user_id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.key.0.comparison", "equals"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.file.0.key.0.field", "log.user.id"),
					// GeoIP enrichment checks
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "enrichment-group-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "geoip-enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enrichment_table.0.target", "log.geo.geoip"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enrichment_table.0.geoip.0.key_field", "log.source.ip"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enrichment_table.0.geoip.0.locale", "en"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.enrichment_table.0.geoip.0.path", "/etc/geoip/GeoLite2-City.mmdb"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_enrichmentTableReferenceTable(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.reference_table_enrichment"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Minimal config
				Config: `
resource "datadog_observability_pipeline" "reference_table_enrichment" {
  name = "reference table enrichment pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "ref-table-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "ref-table-enrichment"
        enabled = true
        include = "*"

        enrichment_table {
          target = "log.enrichment"

          reference_table {
            key_field = "log.user.id"
            table_id  = "550e8400-e29b-41d4-a716-446655440000"
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["ref-table-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "ref-table-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "ref-table-enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.target", "log.enrichment"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.key_field", "log.user.id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.table_id", "550e8400-e29b-41d4-a716-446655440000"),
				),
			},
			{
				// Full config with columns
				Config: `
resource "datadog_observability_pipeline" "reference_table_enrichment" {
  name = "reference table enrichment pipeline updated"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "ref-table-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "ref-table-enrichment"
        enabled = true
        include = "*"

        enrichment_table {
          target = "log.enrichment.updated"

          reference_table {
            key_field = "log.account.id"
            table_id  = "550e8400-e29b-41d4-a716-446655440001"
            columns   = ["region", "country", "city"]
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["ref-table-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "reference table enrichment pipeline updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.target", "log.enrichment.updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.key_field", "log.account.id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.table_id", "550e8400-e29b-41d4-a716-446655440001"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.columns.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.columns.0", "region"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.columns.1", "country"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enrichment_table.0.reference_table.0.columns.2", "city"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googleSecOpsDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.secops"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "secops" {
  name = "secops pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id = "destination-1"
      inputs = ["source-1"]
      google_secops {
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
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.customer_id", "3fa85f64-5717-4562-b3fc-2c963f66afa6"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.log_type", "nginx_logs"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.auth.0.credentials_file", "/secrets/gcp.json"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googleSecOpsDestinationMinimal(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.secops_minimal"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "secops_minimal" {
  name = "secops minimal pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id = "chronicle-dest-1"
      inputs = ["source-1"]
      google_secops {
        customer_id = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
        encoding    = "json"
        log_type    = "nginx_logs"
      }  
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "chronicle-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.customer_id", "3fa85f64-5717-4562-b3fc-2c963f66afa6"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_secops.0.log_type", "nginx_logs"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    destination {
      id     = "destination-1"
      inputs = ["source-1"]
      new_relic {
        region = "us"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.new_relic.0.region", "us"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "destination-1"
      inputs = ["source-1"]
      sentinel_one {
        region = "us"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.sentinel_one.0.region", "us"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "ocsf-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "ocsf-mapper"
        enabled = true
        include = "*"

        ocsf_mapper {
          mapping {
            include         = "source:lib"
            library_mapping = "CloudTrail Account Change"
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["ocsf-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "ocsf-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "ocsf-mapper"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.ocsf_mapper.0.mapping.0.include", "source:lib"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.ocsf_mapper.0.mapping.0.library_mapping", "CloudTrail Account Change"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["source-1"]
      opensearch {
        bulk_index = "logs-datastream"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.opensearch.0.bulk_index", "logs-datastream"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_opensearchDestinationDataStream(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.opensearch_datastream"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Minimal config - only data_stream with dtype
				Config: `
resource "datadog_observability_pipeline" "opensearch_datastream" {
  name = "opensearch-datastream-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "opensearch-destination-1"
      inputs = ["source-1"]
            
      opensearch {
        data_stream {
          dtype = "logs"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "opensearch-datastream-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "opensearch-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.opensearch.0.data_stream.0.dtype", "logs"),
				),
			},
			{
				// Full config - data_stream with all fields
				Config: `
resource "datadog_observability_pipeline" "opensearch_datastream" {
  name = "opensearch-datastream-pipeline-updated"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "opensearch-destination-1"
      inputs = ["source-1"]
            
      opensearch {
        data_stream {
          dtype     = "logs"
          dataset   = "my-application"
          namespace = "production"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "opensearch-datastream-pipeline-updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "opensearch-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.opensearch.0.data_stream.0.dtype", "logs"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.opensearch.0.data_stream.0.dataset", "my-application"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.opensearch.0.data_stream.0.namespace", "production"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id         = "aos-dest-1"
      inputs     = ["source-1"]

      amazon_opensearch {
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

					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "aos-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.bulk_index", "logs-datastream"),

					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.auth.0.strategy", "aws"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.auth.0.aws_region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.auth.0.assume_role", "arn:aws:iam::123456789012:role/example-role"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.auth.0.external_id", "external-id-123"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.auth.0.session_name", "aos-session"),
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
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    destination {
      id     = "destination-1"
      inputs = ["source-1"]
      amazon_opensearch {
        auth {
          strategy = "basic"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_opensearch.0.auth.0.strategy", "basic"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_datadogTagsProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.datadog_tags"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "datadog_tags" {
  name = "datadog tags processor test"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "datadog-tags-group-1"
      enabled = true
      include = "service:my-service"
      inputs  = ["source-1"]

      processor {
        id      = "datadog-tags-processor"
        enabled = true
        include = "service:my-service"

        datadog_tags {
          mode   = "filter"
          action = "include"
          keys   = ["env", "service", "version"]
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["datadog-tags-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "datadog tags processor test"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "datadog-tags-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "datadog-tags-processor"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "service:my-service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.mode", "filter"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.action", "include"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.keys.0", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.keys.1", "service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.keys.2", "version"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
				),
			},
			{
				Config: `
resource "datadog_observability_pipeline" "datadog_tags" {
  name = "datadog tags processor test updated"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "datadog-tags-group-1"
      enabled = true
      include = "service:my-service"
      inputs  = ["source-1"]

      processor {
        id      = "datadog-tags-processor"
        enabled = true
        include = "service:my-service"

        datadog_tags {
          mode   = "filter"
          action = "exclude"
          keys   = ["env", "service"]
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["datadog-tags-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "datadog tags processor test updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.action", "exclude"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.keys.0", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.datadog_tags.0.keys.1", "service"),
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
    source {
      id = "source-1"
      
      datadog_agent {
      }
    }

    processor_group {
      id      = "quota-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "quota-1"
        enabled = true
        include = "*"
        
        quota {
          name        = "MyQuota"
          overflow_action = "drop"

          limit {
            enforce = "events"
            limit   = 1000
          }
        }
      }
    }

    destination {
      id     = "logs-1"
      inputs = ["quota-group-1"]
      
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "quota-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "quota-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.name", "MyQuota"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.overflow_action", "drop"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.limit.0.enforce", "events"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.quota.0.limit.0.limit", "1000"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_opentelemetrySource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.opentelemetry"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "opentelemetry" {
  name = "opentelemetry-pipeline"

  config {
    source {
      id = "opentelemetry-source-1"
      opentelemetry {
        tls {
          crt_file = "/etc/ssl/certs/opentelemetry.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/opentelemetry.key"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["opentelemetry-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "opentelemetry-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "opentelemetry-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.crt_file", "/etc/ssl/certs/opentelemetry.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.key_file", "/etc/ssl/private/opentelemetry.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "opentelemetry-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_opentelemetrySource_noTls(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.opentelemetry_no_tls"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "opentelemetry_no_tls" {
  name = "opentelemetry-pipeline-no-tls"

  config {
    source {
      id = "opentelemetry-source-2"
      opentelemetry {
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["opentelemetry-source-2"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "opentelemetry-pipeline-no-tls"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "opentelemetry-source-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "opentelemetry-source-2"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_opentelemetrySource_update(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.opentelemetry_update"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "opentelemetry_update" {
  name = "opentelemetry-pipeline-update"

  config {
    source {
      id = "opentelemetry-source-3"
      opentelemetry {
        tls {
          crt_file = "/etc/ssl/certs/old.crt"
          ca_file  = "/etc/ssl/certs/old-ca.crt"
          key_file = "/etc/ssl/private/old.key"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["opentelemetry-source-3"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "opentelemetry-pipeline-update"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "opentelemetry-source-3"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.crt_file", "/etc/ssl/certs/old.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.ca_file", "/etc/ssl/certs/old-ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.key_file", "/etc/ssl/private/old.key"),
				),
			},
			{
				Config: `
resource "datadog_observability_pipeline" "opentelemetry_update" {
  name = "opentelemetry-pipeline-updated"

  config {
    source {
      id = "opentelemetry-source-3"
      opentelemetry {
        tls {
          crt_file = "/etc/ssl/certs/new.crt"
          ca_file  = "/etc/ssl/certs/new-ca.crt"
          key_file = "/etc/ssl/private/new.key"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["opentelemetry-source-3"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "opentelemetry-pipeline-updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "opentelemetry-source-3"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.crt_file", "/etc/ssl/certs/new.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.ca_file", "/etc/ssl/certs/new-ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.opentelemetry.0.tls.0.key_file", "/etc/ssl/private/new.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_socketSource_tcp(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.socket_tcp"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "socket_tcp" {
  name = "socket-pipeline"

  config {
    source {
      id = "socket-source-1"
      socket {
        mode = "tcp"

        framing {
          method = "newline_delimited"
        }

        tls {
          crt_file = "/etc/ssl/certs/socket.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/socket.key"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["socket-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "socket-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "socket-source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.mode", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.framing.0.method", "newline_delimited"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.tls.0.crt_file", "/etc/ssl/certs/socket.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.tls.0.key_file", "/etc/ssl/private/socket.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "socket-source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_socketSource_udp(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.socket_udp"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "socket_udp" {
  name = "socket-pipeline-udp"

  config {
    source {
      id = "socket-source-2"
      socket {
        mode = "udp"

        framing {
          method = "character_delimited"
          character_delimited {
            delimiter = "|"
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["socket-source-2"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "socket-pipeline-udp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "socket-source-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.mode", "udp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.framing.0.method", "character_delimited"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.socket.0.framing.0.character_delimited.0.delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "socket-source-2"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_socketDestination_basic(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.socket_dest_basic"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "socket_dest_basic" {
  name = "socket-destination-pipeline-udp"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
      
    destination {
      id       = "socket-dest-2"
      inputs   = ["source-1"]
      
      socket {
        mode     = "udp"
        encoding = "raw_message"

        framing {
          method = "character_delimited"
          character_delimited {
            delimiter = "|"
          }
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "socket-destination-pipeline-udp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "socket-dest-2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.mode", "udp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.encoding", "raw_message"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.framing.0.method", "character_delimited"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.framing.0.character_delimited.0.delimiter", "|"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_socketDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.socket_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "socket_dest" {
  name = "socket-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id       = "socket-dest-1"
      inputs   = ["source-1"]
      socket {
        mode     = "tcp"
        encoding = "json"

        framing {
          method = "newline_delimited"
        }

        tls {
          crt_file = "/etc/ssl/certs/socket.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/socket.key"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "socket-destination-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "socket-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.mode", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.framing.0.method", "newline_delimited"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.tls.0.crt_file", "/etc/ssl/certs/socket.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.socket.0.tls.0.key_file", "/etc/ssl/private/socket.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_socketSource_framingValidation(t *testing.T) {
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "socket_no_framing" {
  name = "socket-pipeline-no-framing"

  config {
    source {
      id = "socket-source-1"
      socket {
        mode = "tcp"
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["socket-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				ExpectError: regexp.MustCompile(`must have a configuration value`),
			},
			{
				Config: `
resource "datadog_observability_pipeline" "socket_invalid_framing" {
  name = "socket-pipeline-invalid-framing"

  config {
    source {
      id = "socket-source-1"
      socket {
        mode = "tcp"

        framing {
          method = "newline_delimited"

          character_delimited {
            delimiter = ","
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["socket-source-1"]
      datadog_logs {
      }
    }
  }
}`,
				ExpectError: regexp.MustCompile(`The 'character_delimited' block must not be specified`),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_cloudPremDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.cloud_prem_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "cloud_prem_dest" {
  name = "cloud-prem-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id     = "cloud-prem-dest-1"
      inputs = ["source-1"]
      cloud_prem {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "cloud-prem-destination-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "cloud-prem-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_kafkaDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.kafka_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "kafka_dest" {
  name = "kafka-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id     = "kafka-dest-1"
      inputs = ["source-1"]
      
      kafka {
        encoding = "json"
        topic    = "logs-topic"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "kafka-destination-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "kafka-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.topic", "logs-topic"),
				),
			},
			{
				Config: `
resource "datadog_observability_pipeline" "kafka_dest" {
  name = "kafka-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id     = "kafka-dest-1"
      inputs = ["source-1"]
      
      kafka {
        encoding               = "json"
        topic                  = "logs-topic"
        compression            = "gzip"
        headers_key            = "headers"
        key_field              = "message_key"
        message_timeout_ms     = 30000
        rate_limit_duration_secs = 1
        rate_limit_num         = 1000
        socket_timeout_ms      = 60000

        sasl {
          mechanism = "PLAIN"
        }

        librdkafka_option {
          name  = "client.id"
          value = "observability-pipeline"
        }

        librdkafka_option {
          name  = "queue.buffering.max.messages"
          value = "15"
        }

        tls {
          ca_file  = "/etc/ssl/certs/ca.crt"
          crt_file = "/etc/ssl/certs/client.crt"
          key_file = "/etc/ssl/private/client.key"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "kafka-destination-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "kafka-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.topic", "logs-topic"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.compression", "gzip"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.headers_key", "headers"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.key_field", "message_key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.message_timeout_ms", "30000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.rate_limit_duration_secs", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.rate_limit_num", "1000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.socket_timeout_ms", "60000"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.sasl.0.mechanism", "PLAIN"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.librdkafka_option.0.name", "client.id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.librdkafka_option.0.value", "observability-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.librdkafka_option.1.name", "queue.buffering.max.messages"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.librdkafka_option.1.value", "15"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.tls.0.crt_file", "/etc/ssl/certs/client.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.kafka.0.tls.0.key_file", "/etc/ssl/private/client.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_customProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.custom_processor"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "custom_processor" {
  name = "remap-vrl-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "remap-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]
      
      processor {
        id      = "remap-processor-1"
        enabled = true
        include = "*"
        
        custom_processor {
          remap {
            include       = "service:web"
            name          = "Parse JSON from message"
            enabled       = true
            source        = ". = parse_json!(string!(.message))"
            drop_on_error = false
          }

          remap {
            include     = "env:prod"
            name        = "Add timestamp"
            enabled     = false
            source      = ".timestamp = now()"
            drop_on_error = false
          }
        }
      }
    }
        
    destination {
      id     = "destination-1"
      inputs = ["remap-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "remap-vrl-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "remap-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "remap-processor-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.0.include", "service:web"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.0.name", "Parse JSON from message"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.0.source", ". = parse_json!(string!(.message))"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.0.drop_on_error", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.1.include", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.1.name", "Add timestamp"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.1.source", ".timestamp = now()"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.1.drop_on_error", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.custom_processor.0.remap.1.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "remap-group-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_amazonS3Destination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.amazon_s3"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "amazon_s3" {
  name = "amazon s3 pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id           = "s3-dest-1"
      inputs       = ["source-1"]

      amazon_s3 {
        bucket       = "my-logs-bucket"
        region       = "us-east-1"
        key_prefix   = "logs/"
        storage_class = "STANDARD"

        auth {
          assume_role  = "arn:aws:iam::123456789012:role/example-role"
          external_id  = "external-id-123"
          session_name = "s3-session"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "s3-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.bucket", "my-logs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.key_prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.storage_class", "STANDARD"),

					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.auth.0.assume_role", "arn:aws:iam::123456789012:role/example-role"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.auth.0.external_id", "external-id-123"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.auth.0.session_name", "s3-session"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_amazonS3Destination_basic(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.amazon_s3_basic"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "amazon_s3_basic" {
  name = "amazon s3 pipeline (minimal)"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id           = "s3-dest-basic-1"
      inputs       = ["source-1"]
      amazon_s3 {
        bucket       = "my-logs-bucket"
        region       = "us-east-1"
        key_prefix   = "logs/"
        storage_class = "STANDARD"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "s3-dest-basic-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.bucket", "my-logs-bucket"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.key_prefix", "logs/"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_s3.0.storage_class", "STANDARD"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_AmazonSecurityLakeDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.amazon_security_lake_dest"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "amazon_security_lake_dest" {
  name = "amazon-security-lake-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id                 = "security-lake-dest-1"
      inputs             = ["source-1"]

			amazon_security_lake {
				bucket             = "my-security-lake-bucket"
				region             = "us-east-1"
				custom_source_name = "my-custom-source"
				tls {
					crt_file = "/path/to/cert.crt"
					ca_file  = "/path/to/ca.crt"
					key_file = "/path/to/key.key"
				}
				auth {
					assume_role  = "arn:aws:iam::123456789012:role/SecurityLakeRole"
					external_id  = "external-id"
					session_name = "session-name"
				}
			}
		}
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "security-lake-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.bucket", "my-security-lake-bucket"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.region", "us-east-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.custom_source_name", "my-custom-source"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.tls.0.crt_file", "/path/to/cert.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.tls.0.ca_file", "/path/to/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.tls.0.key_file", "/path/to/key.key"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.auth.0.assume_role", "arn:aws:iam::123456789012:role/SecurityLakeRole"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.auth.0.external_id", "external-id"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.amazon_security_lake.0.auth.0.session_name", "session-name"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_crowdstrikeNextGenSiemDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.crowdstrike_next_gen_siem_dest"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "crowdstrike_next_gen_siem_dest" {
  name = "crowdstrike-next-gen-siem-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id       = "crowdstrike-dest-1"
      inputs   = ["source-1"]

      crowdstrike_next_gen_siem {
        encoding = "json"
        compression {
          algorithm = "gzip"
          level     = 6
        }
        tls {
          crt_file = "/path/to/cert.crt"
          ca_file  = "/path/to/ca.crt"
          key_file = "/path/to/key.key"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "crowdstrike-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.compression.0.algorithm", "gzip"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.compression.0.level", "6"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.tls.0.crt_file", "/path/to/cert.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.tls.0.ca_file", "/path/to/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.tls.0.key_file", "/path/to/key.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_crowdstrikeNextGenSiemDestination_basic(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.crowdstrike_next_gen_siem_dest_basic"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "crowdstrike_next_gen_siem_dest_basic" {
  name = "crowdstrike-next-gen-siem-destination-pipeline-basic"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id       = "crowdstrike-dest-basic-1"
      inputs   = ["source-1"]
      
      crowdstrike_next_gen_siem {
        encoding = "raw_message"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "crowdstrike-dest-basic-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.crowdstrike_next_gen_siem.0.encoding", "raw_message"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googlePubSubDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.pubsub_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "pubsub_dest" {
  name = "pubsub-destination-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id       = "pubsub-destination-1"
      inputs   = ["source-1"]
        
      google_pubsub {
        project       = "my-gcp-project"
        topic         = "logs-topic"
        encoding      = "json"

        auth {
          credentials_file = "/var/secrets/gcp-creds.json"
        }


        tls {
          crt_file = "/certs/pubsub.crt"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "pubsub-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.project", "my-gcp-project"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.topic", "logs-topic"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.auth.0.credentials_file", "/var/secrets/gcp-creds.json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.tls.0.crt_file", "/certs/pubsub.crt"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_googlePubSubDestinationMinimal(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.pubsub_dest_minimal"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "pubsub_dest_minimal" {
  name = "pubsub-destination-minimal-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
      
    destination {
      id       = "pubsub-destination-1"
      inputs   = ["source-1"]
      google_pubsub {
        project  = "my-gcp-project"
        topic    = "logs-topic"
        encoding = "json"
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.project", "my-gcp-project"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.topic", "logs-topic"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.google_pubsub.0.encoding", "json"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_AddHostname(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.add_hostname_test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineAddHostnameMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "add_hostname pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.add_hostname.#", "1"),
				),
			},
			{
				Config: testAccObservabilityPipelineAddHostnameFull(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "add_hostname pipeline updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "service:updated"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineAddHostnameMinimal() string {
	return `
resource "datadog_observability_pipeline" "add_hostname_test" {
  name = "add_hostname pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {}
    }

    processor_group {
      id      = "group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "add-hostname-proc"
        enabled = true
        include = "service:test"

        add_hostname {}
      }
    }

    destination {
      id     = "dest-1"
      inputs = ["group-1"]
      datadog_logs {}
    }
  }
}
`
}

func testAccObservabilityPipelineAddHostnameFull() string {
	return `
resource "datadog_observability_pipeline" "add_hostname_test" {
  name = "add_hostname pipeline updated"

  config {
    source {
      id = "source-1"
      datadog_agent {}
    }

    processor_group {
      id      = "group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id           = "add-hostname-proc"
        enabled      = true
        include      = "service:updated"
        display_name = "Add Hostname Processor"

        add_hostname {}
      }
    }

    destination {
      id     = "dest-1"
      inputs = ["group-1"]
      datadog_logs {}
    }
  }
}
`
}

func TestAccDatadogObservabilityPipeline_ParseXML(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.parse_xml_test"
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineParseXMLMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "parse_xml pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_xml.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_xml.0.field", "message"),
				),
			},
			{
				Config: testAccObservabilityPipelineParseXMLFull(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "parse_xml pipeline updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_xml.0.field", "xml_data"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_xml.0.include_attr", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_xml.0.parse_number", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_xml.0.attr_prefix", "attr_"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineParseXMLMinimal() string {
	return `
resource "datadog_observability_pipeline" "parse_xml_test" {
  name = "parse_xml pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {}
    }

    processor_group {
      id      = "group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "parse-xml-proc"
        enabled = true
        include = "service:test"

        parse_xml {
          field = "message"
        }
      }
    }

    destination {
      id     = "dest-1"
      inputs = ["group-1"]
      datadog_logs {}
    }
  }
}
`
}

func testAccObservabilityPipelineParseXMLFull() string {
	return `
resource "datadog_observability_pipeline" "parse_xml_test" {
  name = "parse_xml pipeline updated"

  config {
    source {
      id = "source-1"
      datadog_agent {}
    }

    processor_group {
      id      = "group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id           = "parse-xml-proc"
        enabled      = true
        include      = "service:test"
        display_name = "XML Parser"

        parse_xml {
          field                = "xml_data"
          include_attr         = true
          always_use_text_key  = false
          parse_number         = true
          parse_bool           = true
          parse_null           = false
          attr_prefix          = "attr_"
          text_key             = "text"
        }
      }
    }

    destination {
      id     = "dest-1"
      inputs = ["group-1"]
      datadog_logs {}
    }
  }
}
`
}

func TestAccDatadogObservabilityPipeline_SplitArray(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.split_array_test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineSplitArrayMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "split_array pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.split_array.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.split_array.0.array.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.split_array.0.array.0.field", "tags"),
				),
			},
			{
				Config: testAccObservabilityPipelineSplitArrayFull(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "split_array pipeline updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.split_array.0.array.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.split_array.0.array.0.field", "items"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.split_array.0.array.1.field", "events"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineSplitArrayMinimal() string {
	return `
  resource "datadog_observability_pipeline" "split_array_test" {
    name = "split_array pipeline"
  
    config {
      source {
        id = "source-1"
        datadog_agent {}
      }
  
      processor_group {
        id      = "group-1"
        enabled = true
        include = "*"
        inputs  = ["source-1"]
  
        processor {
          id      = "split-array-proc"
          enabled = true
          include = "*"
  
          split_array {
            array {
              include = "*"
              field   = "tags"
            }
          }
        }
      }
  
      destination {
        id     = "dest-1"
        inputs = ["group-1"]
        datadog_logs {}
      }
    }
  }
  `
}

func testAccObservabilityPipelineSplitArrayFull() string {
	return `
  resource "datadog_observability_pipeline" "split_array_test" {
    name = "split_array pipeline updated"
  
    config {
      source {
        id = "source-1"
        datadog_agent {}
      }
  
      processor_group {
        id      = "group-1"
        enabled = true
        include = "*"
        inputs  = ["source-1"]
  
        processor {
          id           = "split-array-proc"
          enabled      = true
          include      = "*"
          display_name = "Array Splitter"
  
          split_array {
            array {
              include = "source:prod"
              field   = "items"
            }
            array {
              include = "source:staging"
              field   = "events"
            }
          }
        }
      }
  
      destination {
        id     = "dest-1"
        inputs = ["group-1"]
        datadog_logs {}
      }
    }
  }
  `
}

func TestAccDatadogObservabilityPipeline_metricsPipelineType(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.metrics_pipeline"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "metrics_pipeline" {
  name = "metrics pipeline"

  config {
    pipeline_type = "metrics"

    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "metric-tags-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "metric-tags-1"
        enabled = true
        include = "*"

        metric_tags {
          rule {
            include = "*"
            mode    = "filter"
            action  = "include"
            keys    = ["env", "service", "version"]
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["metric-tags-group-1"]
      datadog_metrics {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "metrics pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.pipeline_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "metric-tags-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "metric-tags-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.mode", "filter"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.action", "include"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.keys.0", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.keys.1", "service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.keys.2", "version"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_metrics.#", "1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_metricTagsProcessor(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.metric_tags"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "metric_tags" {
  name = "metric tags processor test"

  config {
    pipeline_type = "metrics"

    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "metric-tags-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "metric-tags-processor"
        enabled = true
        include = "*"

        metric_tags {
          rule {
            include = "*"
            mode    = "filter"
            action  = "include"
            keys    = ["env", "service", "version"]
          }

          rule {
            include = "service:web-*"
            mode    = "filter"
            action  = "exclude"
            keys    = ["debug", "internal"]
          }
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["metric-tags-group-1"]
      datadog_metrics {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "metric tags processor test"),
					resource.TestCheckResourceAttr(resourceName, "config.0.pipeline_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "metric-tags-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "metric-tags-processor"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.mode", "filter"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.action", "include"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.keys.0", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.keys.1", "service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.0.keys.2", "version"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.1.include", "service:web-*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.1.mode", "filter"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.1.action", "exclude"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.1.keys.0", "debug"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.metric_tags.0.rule.1.keys.1", "internal"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_datadogMetricsDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.datadog_metrics"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "datadog_metrics" {
  name = "datadog metrics destination test"

  config {
    pipeline_type = "metrics"

    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "filter-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "filter-1"
        enabled = true
        include = "*"

        filter {}
      }
    }

    destination {
      id     = "datadog-metrics-dest-1"
      inputs = ["filter-group-1"]
      datadog_metrics {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "datadog metrics destination test"),
					resource.TestCheckResourceAttr(resourceName, "config.0.pipeline_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "filter-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "datadog-metrics-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "filter-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_metrics.#", "1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_httpClientDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.http_client_dest"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "http_client_dest" {
  name = "http client destination pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id     = "http-client-dest-1"
      inputs = ["source-1"]

      http_client {
        encoding      = "json"
        auth_strategy = "basic"

        compression {
          algorithm = "gzip"
        }

        tls {
          crt_file = "/etc/ssl/certs/http.crt"
          ca_file  = "/etc/ssl/certs/ca.crt"
          key_file = "/etc/ssl/private/http.key"
        }
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "http client destination pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "http-client-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.encoding", "json"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.auth_strategy", "basic"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.compression.0.algorithm", "gzip"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.tls.0.crt_file", "/etc/ssl/certs/http.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.tls.0.ca_file", "/etc/ssl/certs/ca.crt"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.tls.0.key_file", "/etc/ssl/private/http.key"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_httpClientDestinationMinimal(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.http_client_dest_minimal"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "http_client_dest_minimal" {
  name = "http client destination minimal"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }

    destination {
      id     = "http-client-dest-minimal-1"
      inputs = ["source-1"]

      http_client {
        encoding = "json"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "http client destination minimal"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "http-client-dest-minimal-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.http_client.0.encoding", "json"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_metricsFullPipeline(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.metrics_full"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "metrics_full" {
  name = "metrics full pipeline test"

  config {
    pipeline_type = "metrics"

    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "filter-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "filter-1"
        enabled = true
        include = "env:prod"

        filter {}
      }
    }

    processor_group {
      id      = "metric-tags-group-1"
      enabled = true
      include = "*"
      inputs  = ["filter-group-1"]

      processor {
        id      = "metric-tags-1"
        enabled = true
        include = "*"

        metric_tags {
          rule {
            include = "*"
            mode    = "filter"
            action  = "include"
            keys    = ["env", "service"]
          }
        }
      }
    }

    destination {
      id     = "datadog-metrics-dest-1"
      inputs = ["metric-tags-group-1"]
      datadog_metrics {
      }
    }
      destination {
      id     = "http-dest-1"
      inputs = ["metric-tags-group-1"]
      http_client {
        encoding = "json"
        auth_strategy = "none"
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "metrics full pipeline test"),
					resource.TestCheckResourceAttr(resourceName, "config.0.pipeline_type", "metrics"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "filter-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "filter-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.include", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.id", "metric-tags-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.id", "metric-tags-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.metric_tags.0.rule.0.include", "*"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.metric_tags.0.rule.0.keys.0", "env"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.1.processor.0.metric_tags.0.rule.0.keys.1", "service"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "datadog-metrics-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.1.id", "http-dest-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.1.http_client.0.encoding", "json"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_logsPipelineTypeExplicit(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.logs_explicit"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_observability_pipeline" "logs_explicit" {
  name = "logs pipeline explicit"

  config {
    pipeline_type = "logs"

    source {
      id = "source-1"
      datadog_agent {
      }
    }

    processor_group {
      id      = "parser-group-1"
      enabled = true
      include = "*"
      inputs  = ["source-1"]

      processor {
        id      = "parser-1"
        enabled = true
        include = "*"

        parse_json {
          field = "message"
        }
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["parser-group-1"]
      datadog_logs {
      }
    }
  }
}`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "logs pipeline explicit"),
					resource.TestCheckResourceAttr(resourceName, "config.0.pipeline_type", "logs"),
					resource.TestCheckResourceAttr(resourceName, "config.0.source.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.id", "parser-group-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.id", "parser-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.processor_group.0.processor.0.parse_json.0.field", "message"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_elasticsearchDestinationDataStream(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.elasticsearch_datastream"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Minimal config - only data_stream with dtype
				Config: `
resource "datadog_observability_pipeline" "elasticsearch_datastream" {
  name = "elasticsearch-datastream-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "elasticsearch-destination-1"
      inputs = ["source-1"]
            
      elasticsearch {
        api_version = "v8"
        data_stream {
          dtype = "logs"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "elasticsearch-datastream-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "elasticsearch-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.api_version", "v8"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.data_stream.0.dtype", "logs"),
				),
			},
			{
				// Full config - data_stream with all fields
				Config: `
resource "datadog_observability_pipeline" "elasticsearch_datastream" {
  name = "elasticsearch-datastream-pipeline-updated"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "elasticsearch-destination-1"
      inputs = ["source-1"]
            
      elasticsearch {
        api_version = "v8"
        data_stream {
          dtype     = "logs"
          dataset   = "my-application"
          namespace = "production"
        }
      }
    }
  }
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "elasticsearch-datastream-pipeline-updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "elasticsearch-destination-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.api_version", "v8"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.data_stream.0.dtype", "logs"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.data_stream.0.dataset", "my-application"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.elasticsearch.0.data_stream.0.namespace", "production"),
				),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_datadogLogsDestination(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_observability_pipeline.datadog_logs_routes_test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineDatadogLogsDestinationMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "datadog logs routes pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.id", "destination-logs-routes"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "group-logs-routes"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "config.0.destination.0.datadog_logs.0.routes.#"),
				),
			},
			{
				Config: testAccObservabilityPipelineDatadogLogsDestinationFull(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "datadog logs routes pipeline updated"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.inputs.0", "group-logs-routes"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.0.routes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.0.routes.0.route_id", "route-us1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.0.routes.0.include", "service:api"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.0.routes.0.site", "us1"),
					resource.TestCheckResourceAttr(resourceName, "config.0.destination.0.datadog_logs.0.routes.0.api_key_key", "API_KEY_IDENT"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineDatadogLogsDestinationMinimal() string {
	return `
resource "datadog_observability_pipeline" "datadog_logs_routes_test" {
  name = "datadog logs routes pipeline"

  config {
    source {
      id = "source-logs-routes"
      datadog_agent {}
    }

    processor_group {
      id      = "group-logs-routes"
      enabled = true
      include = "*"
      inputs  = ["source-logs-routes"]

      processor {
        id      = "processor-logs-routes"
        enabled = true
        include = "*"

        parse_json {
          field = "message"
        }
      }
    }

    destination {
      id     = "destination-logs-routes"
      inputs = ["group-logs-routes"]

      datadog_logs {}
    }
  }
}
`
}

func testAccObservabilityPipelineDatadogLogsDestinationFull() string {
	return `
resource "datadog_observability_pipeline" "datadog_logs_routes_test" {
  name = "datadog logs routes pipeline updated"

  config {
    source {
      id = "source-logs-routes"
      datadog_agent {}
    }

    processor_group {
      id      = "group-logs-routes"
      enabled = true
      include = "*"
      inputs  = ["source-logs-routes"]

      processor {
        id      = "processor-logs-routes"
        enabled = true
        include = "*"

        parse_json {
          field = "message"
        }
      }
    }

    destination {
      id     = "destination-logs-routes"
      inputs = ["group-logs-routes"]

      datadog_logs {
        routes {
          route_id   = "route-us1"
          include    = "service:api"
          site       = "us1"
          api_key_key = "API_KEY_IDENT"
        }
      }
    }
  }
}
`
}

func TestAccDatadogObservabilityPipeline_elasticsearchValidation(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Both bulk_index and data_stream specified
				Config: `
resource "datadog_observability_pipeline" "elasticsearch_validation" {
  name = "elasticsearch-validation-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "elasticsearch-destination-1"
      inputs = ["source-1"]
            
      elasticsearch {
        api_version = "v7"
        bulk_index  = "logs-index"
        data_stream {
          dtype     = "logs"
          dataset   = "my-app"
          namespace = "production"
        }
      }
    }
  }
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}

func TestAccDatadogObservabilityPipeline_opensearchValidation(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Both bulk_index and data_stream specified
				Config: `
resource "datadog_observability_pipeline" "opensearch_validation" {
  name = "opensearch-validation-pipeline"

  config {
    source {
      id = "source-1"
      datadog_agent {
      }
    }
    
    destination {
      id     = "opensearch-destination-1"
      inputs = ["source-1"]
            
      opensearch {
        bulk_index = "logs-index"
        data_stream {
          dtype     = "logs"
          dataset   = "my-app"
          namespace = "production"
        }
      }
    }
  }
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
