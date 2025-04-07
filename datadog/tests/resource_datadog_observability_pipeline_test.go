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

func TestAccDatadogObservabilityPipeline_basic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.test"

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
resource "datadog_observability_pipeline" "test" {
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
resource "datadog_observability_pipeline" "test" {
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
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.kafka_test"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineKafkaConfig(),
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

func testAccObservabilityPipelineKafkaConfig() string {
	return `
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
}`
}

func TestAccDatadogObservabilityPipeline_datadogAgentWithTLS(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.agent_tls"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineDatadogAgentTLSConfig(),
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

func testAccObservabilityPipelineDatadogAgentTLSConfig() string {
	return `
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
}`
}

func TestAccDatadogObservabilityPipeline_addFieldsProcessor(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.add_fields"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineAddFieldsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogPipelinesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", "add-fields-pipeline"),
					resource.TestCheckResourceAttr(resourceName, "config.sources.datadog_agent.0.id", "source-1"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.field.0.name", "custom.field"),
					resource.TestCheckResourceAttr(resourceName, "config.processors.add_fields.0.field.0.value", "hello-world"),
					resource.TestCheckResourceAttr(resourceName, "config.destinations.datadog_logs.0.id", "destination-1"),
				),
			},
		},
	})
}

func testAccObservabilityPipelineAddFieldsConfig() string {
	return `
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
		id	  = "add-fields-1"
		inputs = ["source-1"]
		include = "*"
        field {
		  name  = "custom.field"
          value = "hello-world"
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
}`
}

func TestAccDatadogObservabilityPipeline_filterProcessor(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_observability_pipeline.filter"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogPipelinesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityPipelineFilterProcessorConfig(),
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

func testAccObservabilityPipelineFilterProcessorConfig() string {
	return `
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
}`
}
