package datadog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
)

const config1 = `
resource "datadog_timeboard" "acceptance_test" {
  title = "Acceptance Test Timeboard"
  description = "Created using the Datadog provider in Terraform"
  read_only = true
  graph {
    title = "Top System CPU by Docker container"
    viz = "toplist"
    request {
      q = "top(avg:docker.cpu.system{*} by {container_name}, 10, 'mean', 'desc')"
    }
    style = {
      palette_flip = false
    }
  }
  graph {
    title = "Top System CPU by Docker container, flipped"
    viz = "toplist"
    request {
      q = "top(avg:docker.cpu.system{*} by {container_name}, 10, 'mean', 'desc')"
    }
    style = {
      palette_flip = true
    }
  }

}
`

const config2 = `
resource "datadog_timeboard" "acceptance_test" {
  title = "Acceptance Test Timeboard"
  description = "Created using the Datadog provider in Terraform"
  graph {
    title = "Redis latency (ms)"
    viz = "timeseries"
    request {
      q = "avg:redis.info.latency_ms{$host}"
      metadata_json = jsonencode({
        "avg:redis.info.latency_ms{$host}": {
          "alias": "Redis latency"
        }
      })
    }
  }
  graph {
    title = "Redis memory usage"
    viz = "timeseries"
    request {
      q = "avg:redis.mem.used{$host} - avg:redis.mem.lua{$host}, avg:redis.mem.lua{$host}"
      aggregator = "sum"
      stacked = true
    }
    request {
      q = "avg:redis.mem.rss{$host}"
    }
    request {
      q = "avg:redis.mem.rss{$host}"
      type = "bars"
      style = {
        palette = "warm"
      }
      aggregator = "max"
    }
  }
  template_variable {
    name = "host"
    prefix = "host"
  }
}
`

const config3 = `
resource "datadog_timeboard" "acceptance_test" {
  title = "Acceptance Test Timeboard"
  description = "Created using the Datadog provider in Terraform"
  graph {
    title = "Redis latency (ms)"
    viz = "timeseries"
    request {
      q = "avg:redis.info.latency_ms{$host}"
    }
    events = ["sources:capistrano"]

    marker {
      label = "High Latency"
      type = "error solid"
      value = "y > 100"
    }
    yaxis = {
      max = "50"
      scale = "sqrt"
            include_zero = true
            include_units = true
    }
  }
  graph {
    title = "ELB Requests"
    viz = "query_value"
    request {
      q = "sum:aws.elb.request_count{*}.as_count()"
      type = "line"
      aggregator = "min"
      conditional_format {
        comparator = ">"
        value = "1000"
        palette = "white_on_red"
      }
      conditional_format {
        comparator = "<="
        value = "1000"
        palette = "white_on_green"
      }
    }
    custom_unit = "hits"
    precision = "*"
    text_align = "left"
  }
  template_variable {
    name = "host"
    prefix = "host"
  }
}
`

const config4 = `
resource "datadog_timeboard" "acceptance_test" {
  title = "Acceptance Test Timeboard"
  description = "Created using the Datadog provider in Terraform"
  graph {
    title = "Widget with Multiple Queries"
    viz = "timeseries"
    request {
      q = "avg:system.cpu.user{*}"
      type = "line"
      style = {
        palette = "purple"
          type    = "dashed"
          width   = "thin"
        }
    }
    request {
      log_query {
        index = "mcnulty"
        compute {
          aggregation = "avg"
          facet = "@duration"
          interval = 5000
        }
        search {
          query = "status:info"
        }
        group_by {
          facet = "host"
          limit = 10
          sort {
            aggregation = "avg"
            order = "desc"
          }
        }
      }
      type = "area"
    }
    request {
      apm_query {
        index = "apm-search"
        compute {
          aggregation = "avg"
          facet = "@duration"
          interval = 5000
        }
        search {
          query = "type:web"
        }
        group_by {
          facet = "resource_name"
          limit = 50
          sort {
            aggregation = "avg"
            order = "desc"
            facet = "@string_query.interval"
          }
        }
      }
      type = "bars"
    }
    request {
      process_query {
        metric = "process.stat.cpu.total_pct"
        search_by = "error"
        filter_by = ["active"]
        limit = 50
      }
      type = "area"
    }
  }
}
`

func TestAccDatadogTimeboard_update(t *testing.T) {

	step0 := resource.TestStep{
		Config: config1,
		Check: resource.ComposeTestCheckFunc(
			checkExists,
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "title", "Acceptance Test Timeboard"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "description", "Created using the Datadog provider in Terraform"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "read_only", "true"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.title", "Top System CPU by Docker container"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.viz", "toplist"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.q", "top(avg:docker.cpu.system{*} by {container_name}, 10, 'mean', 'desc')"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.style.palette_flip", "false"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.style.palette_flip", "true"),
		),
	}

	step1 := resource.TestStep{
		Config: config2,
		Check: resource.ComposeTestCheckFunc(
			checkExists,
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "title", "Acceptance Test Timeboard"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "description", "Created using the Datadog provider in Terraform"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.title", "Redis latency (ms)"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.viz", "timeseries"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.q", "avg:redis.info.latency_ms{$host}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.metadata_json",
				"{\"avg:redis.info.latency_ms{$host}\":{\"alias\":\"Redis latency\"}}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.title", "Redis memory usage"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.viz", "timeseries"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.q", "avg:redis.mem.used{$host} - avg:redis.mem.lua{$host}, avg:redis.mem.lua{$host}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.aggregator", "sum"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.stacked", "true"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.1.q", "avg:redis.mem.rss{$host}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "template_variable.0.name", "host"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "template_variable.0.prefix", "host"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.2.type", "bars"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.2.q", "avg:redis.mem.rss{$host}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.2.aggregator", "max"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.2.style.palette", "warm"),
		),
	}

	step2 := resource.TestStep{
		Config: config3,
		Check: resource.ComposeTestCheckFunc(
			checkExists,
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "title", "Acceptance Test Timeboard"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "description", "Created using the Datadog provider in Terraform"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.title", "Redis latency (ms)"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.viz", "timeseries"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.q", "avg:redis.info.latency_ms{$host}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.events.#", "1"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.marker.0.label", "High Latency"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.marker.0.type", "error solid"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.marker.0.value", "y > 100"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.yaxis.max", "50"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.yaxis.scale", "sqrt"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.yaxis.include_units", "true"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.yaxis.include_zero", "true"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.title", "ELB Requests"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.viz", "query_value"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.q", "sum:aws.elb.request_count{*}.as_count()"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.aggregator", "min"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.type", "line"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.conditional_format.0.comparator", ">"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.conditional_format.0.value", "1000"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.conditional_format.0.palette", "white_on_red"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.conditional_format.1.comparator", "<="),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.conditional_format.1.value", "1000"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.request.0.conditional_format.1.palette", "white_on_green"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.custom_unit", "hits"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.precision", "*"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.1.text_align", "left"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "template_variable.0.name", "host"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "template_variable.0.prefix", "host"),
		),
	}

	step3 := resource.TestStep{
		Config: config4,
		Check: resource.ComposeTestCheckFunc(
			checkExists,
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "title", "Acceptance Test Timeboard"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "description", "Created using the Datadog provider in Terraform"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.title", "Widget with Multiple Queries"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.viz", "timeseries"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.q", "avg:system.cpu.user{*}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.type", "line"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.style.%", "3"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.style.palette", "purple"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.style.type", "dashed"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.style.width", "thin"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.0.metadata_json",
				"{\"avg:system.cpu.user{*}\":{\"alias\":\"Avg CPU user\"}}"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.index", "mcnulty"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.compute.0.aggregation", "avg"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.compute.0.facet", "@duration"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.compute.0.interval", "5000"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.search.0.query", "status:info"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.group_by.#", "1"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.group_by.0.facet", "host"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.group_by.0.limit", "10"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.group_by.0.sort.0.aggregation", "avg"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.group_by.0.sort.0.facet", "@duration"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.log_query.0.group_by.0.sort.0.order", "desc"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.1.type", "area"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.index", "apm-search"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.compute.0.aggregation", "avg"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.compute.0.facet", "@duration"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.compute.0.interval", "5000"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.search.0.query", "type:web"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.group_by.#", "1"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.group_by.0.facet", "resource_name"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.group_by.0.limit", "50"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.group_by.0.sort.0.aggregation", "avg"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.group_by.0.sort.0.facet", "@string_query.interval"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.apm_query.0.group_by.0.sort.0.order", "desc"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.2.type", "bars"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.3.process_query.0.metric", "process.stat.cpu.total_pct"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.3.process_query.0.search_by", "error"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.3.process_query.0.filter_by.#", "1"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.3.process_query.0.filter_by.0", "active"),
			resource.TestCheckResourceAttr("datadog_timeboard.acceptance_test", "graph.0.request.3.process_query.0.limit", "50"),
		),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkDestroy,
		Steps:        []resource.TestStep{step0, step1, step2, step3},
	})
}

func checkExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetDashboard(i); err != nil {
			return fmt.Errorf("Received an error retrieving monitor %s", err)
		}
	}
	return nil
}

func checkDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetDashboard(i); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving timeboard %s", err)
		}
		return fmt.Errorf("Timeboard still exists")
	}
	return nil
}

func TestValidateAggregatorMethod(t *testing.T) {
	validMethods := []string{
		"avg",
		"max",
		"min",
		"sum",
	}
	for _, v := range validMethods {
		_, errors := validateAggregatorMethod(v, "request")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid aggregator method: %q", v, errors)
		}
	}

	invalidMethods := []string{
		"average",
		"suM",
		"m",
		"foo",
	}
	for _, v := range invalidMethods {
		_, errors := validateAggregatorMethod(v, "request")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid aggregator method", v)
		}
	}

}
