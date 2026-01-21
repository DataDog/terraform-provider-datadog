package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIncidentNotificationRuleDataSource_ByID(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("test-ds-id-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationRuleDataSourceByIdConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_rule.test", "id"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "trigger", "incident_created_trigger"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "handles.#", "2"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "handles.0", "@team-email@company.com"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "handles.1", "@slack-channel"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "conditions.#", "1"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "conditions.0.field", "severity"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "conditions.0.values.#", "2"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "conditions.0.values.0", "SEV-1"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_rule.test", "conditions.0.values.1", "SEV-2"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_rule.test", "incident_type"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_rule.test", "created"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_rule.test", "modified"),
				),
			},
		},
	})
}

func testAccCheckDatadogIncidentNotificationRuleDataSourceByIdConfig(ruleName string) string {
	return fmt.Sprintf(`
# First create an incident type to reference
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for notification rule"
}

resource "datadog_incident_notification_rule" "rule" {
  enabled     = true
  trigger     = "incident_created_trigger"
  
  handles = [
    "@team-email@company.com",
    "@slack-channel"
  ]
  
  conditions {
    field = "severity"
    values = ["SEV-1", "SEV-2"]
  }
  
  renotify_on = ["status", "severity"]
  
  incident_type = datadog_incident_type.test.id
}

data "datadog_incident_notification_rule" "test" {
  id = datadog_incident_notification_rule.rule.id
}`, ruleName)
}
