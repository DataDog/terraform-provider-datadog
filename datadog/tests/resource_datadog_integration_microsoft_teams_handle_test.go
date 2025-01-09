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

func TestAccMSTeamsTenantBasedHandlesBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	apiInstances := providers.frameworkProvider.DatadogApiInstances
	auth := providers.frameworkProvider.Auth
	resp, _, err := apiInstances.GetMicrosoftTeamsIntegrationApiV2().ListTenantBasedHandles(auth)
	// Return if there are no available channels to test
	if err != nil || len(resp.Data) == 0 {
		return
	}

	tenantName := *resp.Data[0].Attributes.TenantName
	teamName := *resp.Data[0].Attributes.TeamName
	channelName := *resp.Data[0].Attributes.ChannelName

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMSTeamsTenantBasedHandlesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMSTeamsTenantBasedHandles(uniq, tenantName, teamName, channelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMSTeamsTenantBasedHandlesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "tenant_name", tenantName),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "team_name", teamName),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "channel_name", channelName),
				),
			},
			{
				Config: testAccCheckDatadogMSTeamsTenantBasedHandlesUpdated(uniq, tenantName, teamName, channelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMSTeamsTenantBasedHandlesExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "name", fmt.Sprintf("%s-updated", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "tenant_name", tenantName),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "team_name", teamName),
					resource.TestCheckResourceAttr(
						"datadog_integration_ms_teams_tenant_based_handle.foo", "channel_name", channelName),
				),
			},
		},
	})
}

func testAccCheckDatadogMSTeamsTenantBasedHandles(uniq string, tenantName string, teamName string, channelName string) string {
	return fmt.Sprintf(`
resource "datadog_integration_ms_teams_tenant_based_handle" "foo" {
    name = "%s"
    tenant_name = "%s"
	team_name = "%s"
	channel_name = "%s"
}`, uniq, tenantName, teamName, channelName)
}

func testAccCheckDatadogMSTeamsTenantBasedHandlesUpdated(uniq string, tenantName string, teamName string, channelName string) string {
	return fmt.Sprintf(`
resource "datadog_integration_ms_teams_tenant_based_handle" "foo" {
    name = "%s-updated"
    tenant_name = "%s"
	team_name = "%s"
	channel_name = "%s"
}`, uniq, tenantName, teamName, channelName)
}

func testAccCheckDatadogMSTeamsTenantBasedHandlesDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := MSTeamsTenantBasedHandlesDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func MSTeamsTenantBasedHandlesDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_ms_teams_tenant_based_handle" {
			continue
		}
		id := r.Primary.ID

		err := utils.Retry(2, 10, func() error {
			_, httpResp, err := apiInstances.GetMicrosoftTeamsIntegrationApiV2().GetTenantBasedHandle(auth, id)
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

func testAccCheckDatadogMSTeamsTenantBasedHandlesExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := msTeamsTenantBasedHandlesExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func msTeamsTenantBasedHandlesExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_ms_teams_tenant_based_handle" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetMicrosoftTeamsIntegrationApiV2().GetTenantBasedHandle(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving handle")
		}
	}
	return nil
}
