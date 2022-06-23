package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogNotebook_basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: notebookBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.basic_notebook"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "name", name),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "status", "published"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "metadata.0.is_template", "false"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "metadata.0.take_snapshots", "true"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "metadata.0.type", "postmortem"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "time.0.notebook_relative_time.0.live_span", "5m"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "cells.0.attributes.0.markdown_cell.0.definition.0.text", "Description of text"),
				),
			},
			{
				Config: notebookBasicConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.basic_notebook"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "name", name+"-updated"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "status", "published"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "metadata.0.is_template", "false"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "metadata.0.take_snapshots", "false"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "metadata.0.type", "postmortem"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "time.0.notebook_absolute_time.0.start", "2021-02-24T11:18:28Z"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "time.0.notebook_absolute_time.0.end", "2021-02-24T11:20:28Z"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "time.0.notebook_absolute_time.0.live", "false"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.basic_notebook", "cells.0.attributes.0.markdown_cell.0.definition.0.text", "Description of text updated"),
				),
			},
		},
	})
}

func notebookBasicConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "basic_notebook" {  
  cells {
    type = "notebook_cells"
    attributes {
      markdown_cell {
        definition {
          text = "Description of text"
        }
      }
    }
  }
  
  time {
    notebook_relative_time {
      live_span = "5m"
    }
  }
  
  name = "%s"
  status = "published"
  metadata {
    is_template = false
    take_snapshots = true
    type = "postmortem"
  }
}
`, uniq)
}

func notebookBasicConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "basic_notebook" {  
  cells {
    type = "notebook_cells"
    attributes {
      markdown_cell {
        definition {
          text = "Description of text updated"
        }
      }
    }
  }
  
  time {
    notebook_absolute_time {
	  start = "2021-02-24T11:18:28Z"
	  end   = "2021-02-24T11:20:28Z"
      live  = false
    }
  }
  
  name = "%s-updated"
  status = "published"
  metadata {
    is_template = false
    take_snapshots = false
    type = "postmortem"
  }
}
`, uniq)
}

func TestAccDatadogNotebookTimeseriesCell(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: notebookTimeseriesCellConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.timeseries"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "name", name),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.attributes.0.timeseries_cell.0.time.0.notebook_relative_time.0.live_span", "5m"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.attributes.0.timeseries_cell.0.definition.0.request.0.q", "avg:system.cpu.user{app:general} by {env}"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.attributes.0.timeseries_cell.0.split_by.0.keys.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.attributes.0.timeseries_cell.0.split_by.0.tags.#", "2"),
					resource.TestCheckResourceAttr("datadog_notebook.timeseries", "cells.0.attributes.0.timeseries_cell.0.graph_size", "s"),
				),
			},
		},
	})
}

func notebookTimeseriesCellConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "timeseries" {  
  cells {
    type = "notebook_cells"
    attributes {
      timeseries_cell {
        definition {
          request {
            q            = "avg:system.cpu.user{app:general} by {env}"
            display_type = "line"
            style {
              palette    = "warm"
              line_type  = "dashed"
              line_width = "thin"
            }
            metadata {
              expression = "avg:system.cpu.user{app:general} by {env}"
              alias_name = "Alpha"
            }
          }
        }
        split_by {
          keys = ["test"]
          tags = ["test:true", "foo:bar"]
        }
        time {
          notebook_relative_time {
            live_span = "5m"
          }
        }
        graph_size = "s"
      }
    }
  }
  time {
    notebook_absolute_time {
	  start = "2021-02-24T11:18:28Z"
	  end   = "2021-02-24T11:20:28Z"
      live  = false
    }
  }
  
  name = "%s"
  status = "published"
  metadata {
    is_template = false
    take_snapshots = false
    type = "postmortem"
  }
}
`, uniq)
}

func TestAccDatadogNotebookToplistCell(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: notebookToplistCellConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.toplist"),
					resource.TestCheckResourceAttr("datadog_notebook.toplist", "name", name),
					resource.TestCheckResourceAttr("datadog_notebook.toplist", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.toplist", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.toplist", "cells.0.attributes.0.toplist_cell.0.definition.0.request.0.q", "avg:system.cpu.user{app:general} by {env}"),
					resource.TestCheckResourceAttr("datadog_notebook.toplist", "cells.0.attributes.0.toplist_cell.0.graph_size", "m"),
				),
			},
		},
	})
}

func notebookToplistCellConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "toplist" {  
  cells {
    type = "notebook_cells"
    attributes {
	  toplist_cell {
	    definition {
	    request {
		  q = "avg:system.cpu.user{app:general} by {env}"
		  conditional_formats {
		    comparator = "<"
		    value      = "2"
		    palette    = "white_on_green"
		  }
		  conditional_formats {
		    comparator = ">"
		    value      = "2.2"
		    palette    = "white_on_red"
		  }
	    }
	    title = "Widget Title"
	    }

	    graph_size = "m"
	  }
    }
  }
  time {
    notebook_absolute_time {
	  start = "2021-02-24T11:18:28Z"
	  end   = "2021-02-24T11:20:28Z"
      live  = false
    }
  }
  
  name = "%s"
  status = "published"
}
`, uniq)
}

func TestAccDatadogNotebookHeatmapCell(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: notebookHeatmapCellConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.toplist"),
					resource.TestCheckResourceAttr("datadog_notebook.heatmap", "name", name),
					resource.TestCheckResourceAttr("datadog_notebook.heatmap", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.heatmap", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.heatmap", "cells.0.attributes.0.heatmap_cell.0.definition.0.request.0.q", "avg:system.load.1{env:staging} by {account}"),
					resource.TestCheckResourceAttr("datadog_notebook.heatmap", "cells.0.attributes.0.heatmap_cell.0.graph_size", "m"),
				),
			},
		},
	})
}

func notebookHeatmapCellConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "heatmap" {  
  cells {
    attributes {
	  heatmap_cell {
		  definition {
		  request {
		    q = "avg:system.load.1{env:staging} by {account}"
		    style {
		  	  palette = "warm"
		    }
	  	  }
	    }
	    graph_size = "m"
	  }
    }
  }

  time {
    notebook_absolute_time {
	  start = "2021-02-24T11:18:28Z"
	  end   = "2021-02-24T11:20:28Z"
      live  = false
    }
  }
  
  name = "%s"
  status = "published"
}
`, uniq)
}

func TestAccDatadogNotebookDistributionCell(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: notebookDistributionCellConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.distribution"),
					resource.TestCheckResourceAttr("datadog_notebook.distribution", "name", name),
					resource.TestCheckResourceAttr("datadog_notebook.distribution", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.distribution", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.distribution", "cells.0.attributes.0.distribution_cell.0.definition.0.request.0.q", "avg:system.load.1{env:staging} by {account}"),
					resource.TestCheckResourceAttr("datadog_notebook.distribution", "cells.0.attributes.0.distribution_cell.0.graph_size", "m"),
				),
			},
		},
	})
}

func notebookDistributionCellConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "distribution" {
  cells {
    attributes {
	  distribution_cell {
		  definition {
		  request {
		    q = "avg:system.load.1{env:staging} by {account}"
		    style {
		  	  palette = "warm"
		    }
	  	  }
	    }
	    graph_size = "m"
	  }
    }
  }

  time {
    notebook_absolute_time {
	  start = "2021-02-24T11:18:28Z"
	  end   = "2021-02-24T11:20:28Z"
      live  = false
    }
  }
  
  name = "%s"
  status = "published"
}
`, uniq)
}

func TestAccDatadogNotebookLogStreamCell(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: notebookLogStreamCellConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.log_stream"),
					resource.TestCheckResourceAttr("datadog_notebook.log_stream", "name", name),
					resource.TestCheckResourceAttr("datadog_notebook.log_stream", "cells.0.type", "notebook_cells"),
					resource.TestCheckResourceAttr("datadog_notebook.log_stream", "cells.0.attributes.#", "1"),
					resource.TestCheckResourceAttr("datadog_notebook.log_stream", "cells.0.attributes.0.log_stream_cell.0.definition.0.query", "error"),
					resource.TestCheckResourceAttr("datadog_notebook.log_stream", "cells.0.attributes.0.log_stream_cell.0.graph_size", "m"),
				),
			},
		},
	})
}

func notebookLogStreamCellConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_notebook" "log_stream" {
  cells {
    attributes {
	  log_stream_cell {
        definition {
          indexes             = ["main"]
          query               = "error"
          columns             = ["core_host", "core_service", "tag_source"]
          show_date_column    = true
          show_message_column = true
          message_display     = "expanded-md"
          sort {
            column = "time"
            order  = "desc"
          }
        }
        graph_size = "m"
      }
    }
  }

  time {
    notebook_absolute_time {
	  start = "2021-02-24T11:18:28Z"
	  end   = "2021-02-24T11:20:28Z"
      live  = false
    }
  }
  
  name = "%s"
  status = "published"
}
`, uniq)
}

func testAccCheckDatadogNotebookExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			id, err := strconv.ParseInt(r.Primary.ID, 10, 64)
			if err != nil {
				return err
			}
			if _, httpresp, err := datadogClient.NotebooksApi.GetNotebook(auth, id); err != nil {
				return utils.TranslateClientError(err, httpresp, "error checking notebook existence")
			}
		}
		return nil
	}
}

func testAccCheckDatadogNotebookDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		for _, r := range s.RootModule().Resources {
			id, err := strconv.ParseInt(r.Primary.ID, 10, 64)
			if err != nil {
				return err
			}

			_, resp, err := datadogClient.NotebooksApi.GetNotebook(auth, id)
			if err != nil {
				if resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving notebook: %s", err.Error())
				}
			} else {
				return fmt.Errorf("notebook %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
