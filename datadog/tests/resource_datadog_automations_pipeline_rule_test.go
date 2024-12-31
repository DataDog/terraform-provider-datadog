package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogAutomationsPipelineRule_complex(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	inboxConfig := `
		resource "datadog_automations_pipeline_rule" "foo" {
			name    = "` + name + `"
			enabled = false
			inbox {
				rule {
					issue_type = "vulnerability"
					rule_types = ["misconfiguration"]
					rule_ids   = ["tdl-pxj-hqb"]
					severities = ["critical"]
					query      = "nice"
				}
				action {
					reason_description = "Scheduled maintenance window"
				}
			}
		}
	`

	muteConfig := `
		resource "datadog_automations_pipeline_rule" "foo" {
			name    = "` + name + `"
			enabled = false
			mute {
				rule {
					issue_type = "vulnerability"
					rule_types = ["misconfiguration"]
					rule_ids   = ["tdl-pxj-hqb"]
					severities = ["critical"]
					query      = "nice"
				}
				action {
					reason_description = "Scheduled maintenance window"
					reason             = "risk_accepted"
					enabled_until      = 2000000000000
				}
			}
		}
	`

	dueDateConfig := `
		resource "datadog_automations_pipeline_rule" "foo" {
			name    = "` + name + `"
			enabled = false
			due_date {
				rule {
					issue_type = "vulnerability"
					rule_types = ["misconfiguration"]
					rule_ids   = ["tdl-pxj-hqb"]
					severities = []
					query      = "nice"
				}
				action {
					notify_before_due = "P12D"
					due_time_per_severity {
						severity = "critical"
						time     = "P3D"
					}
					due_time_per_severity {
						severity = "low"
						time     = "P6D"
					}
				}
			}
		}
	`

	path := "datadog_automations_pipeline_rule.foo"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: inboxConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", name),
					resource.TestCheckResourceAttr(path, "enabled", "false"),
					resource.TestCheckResourceAttr(path, "inbox.rule.issue_type", "vulnerability"),
					resource.TestCheckResourceAttr(path, "inbox.rule.rule_types.#", "1"),
					resource.TestCheckResourceAttr(path, "inbox.rule.rule_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(path, "inbox.rule.rule_ids.#", "1"),
					resource.TestCheckResourceAttr(path, "inbox.rule.rule_ids.0", "tdl-pxj-hqb"),
					resource.TestCheckResourceAttr(path, "inbox.rule.severities.#", "1"),
					resource.TestCheckResourceAttr(path, "inbox.rule.severities.0", "critical"),
					resource.TestCheckResourceAttr(path, "inbox.rule.query", "nice"),
					resource.TestCheckResourceAttr(path, "inbox.action.reason_description", "Scheduled maintenance window"),
				),
			},
			{
				Config: muteConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", name),
					resource.TestCheckResourceAttr(path, "enabled", "false"),
					resource.TestCheckResourceAttr(path, "mute.rule.issue_type", "vulnerability"),
					resource.TestCheckResourceAttr(path, "mute.rule.rule_types.#", "1"),
					resource.TestCheckResourceAttr(path, "mute.rule.rule_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(path, "mute.rule.rule_ids.#", "1"),
					resource.TestCheckResourceAttr(path, "mute.rule.rule_ids.0", "tdl-pxj-hqb"),
					resource.TestCheckResourceAttr(path, "mute.rule.severities.#", "1"),
					resource.TestCheckResourceAttr(path, "mute.rule.severities.0", "critical"),
					resource.TestCheckResourceAttr(path, "mute.rule.query", "nice"),
					resource.TestCheckResourceAttr(path, "mute.action.reason_description", "Scheduled maintenance window"),
					resource.TestCheckResourceAttr(path, "mute.action.reason", "risk_accepted"),
					resource.TestCheckResourceAttr(path, "mute.action.enabled_until", "2000000000000"),
				),
			},
			{
				Config: dueDateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "name", name),
					resource.TestCheckResourceAttr(path, "enabled", "false"),
					resource.TestCheckResourceAttr(path, "due_date.rule.issue_type", "vulnerability"),
					resource.TestCheckResourceAttr(path, "due_date.rule.rule_types.#", "1"),
					resource.TestCheckResourceAttr(path, "due_date.rule.rule_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(path, "due_date.rule.rule_ids.#", "1"),
					resource.TestCheckResourceAttr(path, "due_date.rule.rule_ids.0", "tdl-pxj-hqb"),
					resource.TestCheckResourceAttr(path, "due_date.rule.severities.#", "0"),
					resource.TestCheckResourceAttr(path, "due_date.rule.query", "nice"),
					resource.TestCheckResourceAttr(path, "due_date.action.notify_before_due", "P12D"),
					resource.TestCheckResourceAttr(path, "due_date.action.due_time_per_severity.#", "2"),
					resource.TestCheckResourceAttr(path, "due_date.action.due_time_per_severity.0.severity", "critical"),
					resource.TestCheckResourceAttr(path, "due_date.action.due_time_per_severity.0.time", "P3D"),
					resource.TestCheckResourceAttr(path, "due_date.action.due_time_per_severity.1.severity", "low"),
					resource.TestCheckResourceAttr(path, "due_date.action.due_time_per_severity.1.time", "P6D"),
				),
			},
		},
	})
}
