package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIncidentNotificationTemplateDataSource_ByID(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	templateName := fmt.Sprintf("test-ds-id-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationTemplateDataSourceByIdConfig(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "id"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_template.test", "name", templateName),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_template.test", "subject", "SEV-2 Incident: "+templateName),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_template.test", "category", "alert"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "content"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "created"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "modified"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "incident_type"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentNotificationTemplateDataSource_ByName(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	templateName := fmt.Sprintf("test-ds-name-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationTemplateDataSourceByNameConfig(templateName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "id"),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_template.test", "name", templateName),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_template.test", "subject", "SEV-2 Incident: "+templateName),
					resource.TestCheckResourceAttr(
						"data.datadog_incident_notification_template.test", "category", "alert"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "content"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "created"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "modified"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_incident_notification_template.test", "incident_type"),
				),
			},
		},
	})
}

func testAccCheckDatadogIncidentNotificationTemplateDataSourceByIdConfig(templateName string) string {
	return fmt.Sprintf(`
# First create an incident type to reference
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for notification template"
}

resource "datadog_incident_notification_template" "template" {
  name         = "%s"
  subject      = "SEV-2 Incident: %s"
  content      = "An incident has been declared.\n\nTitle: %s\nSeverity: SEV-2\nAffected Services: web-service, database-service\nStatus: active\n\nPlease join the incident channel for updates."
  category     = "alert"
  incident_type = datadog_incident_type.test.id
}

data "datadog_incident_notification_template" "test" {
  id = datadog_incident_notification_template.template.id
}`, templateName, templateName, templateName, templateName)
}

func testAccCheckDatadogIncidentNotificationTemplateDataSourceByNameConfig(templateName string) string {
	return fmt.Sprintf(`
# First create an incident type to reference
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for notification template"
}

resource "datadog_incident_notification_template" "template" {
  name         = "%s"
  subject      = "SEV-2 Incident: %s"
  content      = "An incident has been declared.\n\nTitle: %s\nSeverity: SEV-2\nAffected Services: web-service, database-service\nStatus: active\n\nPlease join the incident channel for updates."
  category     = "alert"
  incident_type = datadog_incident_type.test.id
}

data "datadog_incident_notification_template" "test" {
  name = datadog_incident_notification_template.template.name
}`, templateName, templateName, templateName, templateName)
}
