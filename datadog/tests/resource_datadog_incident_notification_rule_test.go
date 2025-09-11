package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIncidentNotificationRule_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("test-rule-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationRuleConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "trigger", "incident_created_trigger"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "handles.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "handles.0", "@team-email@company.com"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "handles.1", "@slack-channel"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "conditions.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "conditions.0.field", "severity"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "conditions.0.values.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "conditions.0.values.0", "SEV-1"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "conditions.0.values.1", "SEV-2"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_rule.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_rule.foo", "incident_type"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_rule.foo", "created"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_notification_rule.foo", "modified"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentNotificationRule_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("test-rule-updated-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationRuleConfig(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "trigger", "incident_created_trigger"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentNotificationRuleConfigUpdated(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "trigger", "incident_saved_trigger"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "handles.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_incident_notification_rule.foo", "handles.0", "@updated-team@company.com"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentNotificationRule_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := fmt.Sprintf("test-rule-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentNotificationRuleConfig(ruleName),
			},
			{
				ResourceName:      "datadog_incident_notification_rule.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentNotificationRuleConfig(ruleName string) string {
	return fmt.Sprintf(`
# First create an incident type to reference
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for notification rule"
}

resource "datadog_incident_notification_rule" "foo" {
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
}`, ruleName)
}

func testAccCheckDatadogIncidentNotificationRuleConfigUpdated(ruleName string) string {
	return fmt.Sprintf(`
# First create an incident type to reference
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for notification rule"
}

resource "datadog_incident_notification_rule" "foo" {
  enabled     = false
  trigger     = "incident_saved_trigger"
  
  handles = [
    "@updated-team@company.com"
  ]
  
  conditions {
    field = "severity"
    values = ["SEV-1", "SEV-2"]
  }
  
  renotify_on = ["status"]
  
  incident_type = datadog_incident_type.test.id
}`, ruleName)
}

func testAccCheckDatadogIncidentNotificationRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := incidentNotificationRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogIncidentNotificationRuleDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := incidentNotificationRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func incidentNotificationRuleExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_notification_rule" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetIncidentNotificationRule(ctx, parseUUID(id))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving incident notification rule")
		}
	}
	return nil
}

func incidentNotificationRuleDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_notification_rule" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetIncidentNotificationRule(ctx, parseUUID(id))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving incident notification rule")
		}
		return fmt.Errorf("incident notification rule still exists")
	}
	return nil
}
