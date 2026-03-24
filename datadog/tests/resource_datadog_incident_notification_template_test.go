package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIncidentNotificationTemplate_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	templateName := fmt.Sprintf("test-tpl-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentNotificationTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationTemplateConfig(templateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentNotificationTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_template.foo", "name", templateName),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_template.foo", "subject", "SEV-2 Incident: "+templateName),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_template.foo", "category", "alert"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_template.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_template.foo", "created"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_template.foo", "modified"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_template.foo", "incident_type"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentNotificationTemplate_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	templateName := fmt.Sprintf("test-tpl-updated-%d", clockFromContext(ctx).Now().Unix())
	templateNameUpdated := templateName + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentNotificationTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationTemplateConfig(templateName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentNotificationTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_template.foo", "name", templateName),
				),
			},
			{
				Config: testAccCheckDatadogIncidentNotificationTemplateConfig(templateNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentNotificationTemplateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_template.foo", "name", templateNameUpdated),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_template.foo", "subject", "SEV-2 Incident: "+templateNameUpdated),
				),
			},
		},
	})
}

func TestAccDatadogIncidentNotificationTemplate_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	templateName := fmt.Sprintf("test-tpl-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentNotificationTemplateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationTemplateConfig(templateName),
			},
			{
				ResourceName:      "datadog_incident_notification_template.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentNotificationTemplateConfig(templateName string) string {
	return fmt.Sprintf(`
# First create an incident type to reference
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for notification template"
}

resource "datadog_incident_notification_template" "foo" {
  name         = "%s"
  subject      = "SEV-2 Incident: %s"
  content      = "An incident has been declared.\n\nTitle: %s\nSeverity: SEV-2\nAffected Services: web-service, database-service\nStatus: active\n\nPlease join the incident channel for updates."
  category     = "alert"
  incident_type = datadog_incident_type.test.id
}`, templateName, templateName, templateName, templateName)
}

func testAccCheckDatadogIncidentNotificationTemplateExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := incidentNotificationTemplateExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogIncidentNotificationTemplateDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := incidentNotificationTemplateDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func incidentNotificationTemplateExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_notification_template" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetIncidentNotificationTemplate(ctx, parseUUID(id))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving incident notification template")
		}
	}
	return nil
}

func incidentNotificationTemplateDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_notification_template" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetIncidentNotificationTemplate(ctx, parseUUID(id))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving incident notification template")
		}
		return fmt.Errorf("incident notification template still exists")
	}
	return nil
}

func parseUUID(id string) uuid.UUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}
