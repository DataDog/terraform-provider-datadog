package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const datadogDashboardNoteConfig = `
resource "datadog_dashboard" "note_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		note_definition {
			tick_pos= "50%"
			show_tick = true
			tick_edge = "bottom"
			text_align = "center"
			content = "This is a note widget"
			font_size = "18"
			background_color = "green"
		}
	}
}
`

func datadogDashboardNoteConfigNoContent(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard" "note_dashboard" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		note_definition {
			content = ""
		}
	}
}`, uniq)
}

var datadogDashboardNoteAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.note_definition.0.content = This is a note widget",
	"title = {{uniq}}",
	"widget.0.note_definition.0.font_size = 18",
	"widget.0.note_definition.0.text_align = center",
	"widget.0.note_definition.0.show_tick = true",
	"widget.0.note_definition.0.tick_edge = bottom",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.0.note_definition.0.tick_pos = 50%",
	"widget.0.note_definition.0.background_color = green",
}

func TestAccDatadogDashboardNote(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardNoteConfig, "datadog_dashboard.note_dashboard", datadogDashboardNoteAsserts)
}

func TestAccDatadogDashboardNote_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardNoteConfig, "datadog_dashboard.note_dashboard")
}

func TestAccDatadogDashboardNoteContentError(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      datadogDashboardNoteConfigNoContent(uniq),
				ExpectError: regexp.MustCompile("expected \"widget.0.note_definition.0.content\" to not be an empty string"),
			},
		},
	})
}
