package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccMSTeamsWorkflowsWebhookHandlesBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	url := "https://fake.url.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMSTeamsWorkflowsWebhookHandles(uniq, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_workflows_webhook_handle.foo", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_workflows_webhook_handle.foo", "url", url),
				),
			},
			{
				Config: testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesUpdated(uniq, url),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_workflows_webhook_handle.foo", "name", fmt.Sprintf("%s-updated", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_workflows_webhook_handle.foo", "url", url),
				),
			},
		},
	})
}

func testAccCheckDatadogMSTeamsWorkflowsWebhookHandles(uniq string, url string) string {
	return fmt.Sprintf(`
resource "datadog_integration_ms_teams_workflows_webhook_handle" "foo" {
    name = "%s"
    url = "%s"
}`, uniq, url)
}

func testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesUpdated(uniq string, url string) string {
	return fmt.Sprintf(`
resource "datadog_integration_ms_teams_workflows_webhook_handle" "foo" {
    name = "%s-updated"
    url = "%s"
}`, uniq, url)
}

func testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := MSTeamsWorkflowsWebhookHandlesDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func MSTeamsWorkflowsWebhookHandlesDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_ms_teams_workflows_webhook_handle" {
			continue
		}
		id := r.Primary.ID

		err := utils.Retry(2, 10, func() error {
			_, httpResp, err := apiInstances.GetMicrosoftTeamsIntegrationApiV2().GetWorkflowsWebhookHandle(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving handle %s", err)}
			}
			return &utils.RetryableError{Prob: "Handle still exists"}
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckDatadogMSTeamsWorkflowsWebhookHandlesExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := msTeamsWorkflowsWebhookHandlesExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func msTeamsWorkflowsWebhookHandlesExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_ms_teams_workflows_webhook_handle" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetMicrosoftTeamsIntegrationApiV2().GetWorkflowsWebhookHandle(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving handle")
		}
	}
	return nil
}
