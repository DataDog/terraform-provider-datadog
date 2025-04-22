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
