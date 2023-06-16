package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	communityClient "github.com/zorkian/go-datadog-api"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogIntegrationSlackChannel_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	uniqueChannelAccountName := reg.ReplaceAllString(uniqueEntityName(ctx, t), "")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationSlackChannelDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Workaround to ensure we create the slack integration before running the tests.
				Config: emptyLogsArchiveConfig(),
				Check: resource.ComposeTestCheckFunc(
					createSlackIntegration(providers.frameworkProvider),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackChannelConfigCreate(uniqueChannelAccountName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelExists(providers.frameworkProvider, "datadog_integration_slack_channel.slack_channel"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "channel_name", "#"+uniqueChannelAccountName),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.message", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.notified", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.snapshot", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.tags", "true"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackChannelConfigUpdate(uniqueChannelAccountName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelExists(providers.frameworkProvider, "datadog_integration_slack_channel.slack_channel"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "channel_name", "#"+uniqueChannelAccountName),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.message", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.notified", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.snapshot", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "display.0.tags", "false"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationSlackChannel_Import(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	uniqueChannelAccountName := reg.ReplaceAllString(uniqueEntityName(ctx, t), "")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationSlackChannelDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// Workaround to ensure we create the slack integration before running the tests.
				Config: emptyLogsArchiveConfig(),
				Check: resource.ComposeTestCheckFunc(
					createSlackIntegration(providers.frameworkProvider),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackChannelConfigCreate(uniqueChannelAccountName),
			},
			{
				ResourceName:      "datadog_integration_slack_channel.slack_channel",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIntegrationSlackChannelConfigCreate(uniq string) string {
	return fmt.Sprintf(`
       resource "datadog_integration_slack_channel" "slack_channel" {
			display {
				message = true
				notified = true
				snapshot = true
				tags = true
			}
			channel_name = "#%s"
			account_name    = "test_account"
       }
   `, uniq)
}

func testAccCheckDatadogIntegrationSlackChannelConfigUpdate(uniq string) string {
	return fmt.Sprintf(`
       resource "datadog_integration_slack_channel" "slack_channel" {
			display {
				message = false
				notified = false
				snapshot = false
				tags = false
			}
			channel_name = "#%s"
			account_name    = "test_account"
       }
   `, uniq)
}

func emptyLogsArchiveConfig() string {
	return `
		data "datadog_ip_ranges" "test" {
		}
   `
}

func testAccCheckDatadogIntegrationSlackChannelExists(provider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		accountName := s.RootModule().Resources[resourceName].Primary.Attributes["account_name"]
		channelName := s.RootModule().Resources[resourceName].Primary.Attributes["channel_name"]

		_, httpresp, err := apiInstances.GetSlackIntegrationApiV1().GetSlackIntegrationChannel(auth, accountName, channelName)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking slack_channel existence")
		}

		return nil
	}
}

func testAccCheckDatadogIntegrationSlackChannelDestroy(provider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := provider.DatadogApiInstances
		auth := provider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_slack_channel" {
				continue
			}

			accountName := r.Primary.Attributes["account_name"]
			channelName := r.Primary.Attributes["channel_name"]

			_, resp, err := apiInstances.GetSlackIntegrationApiV1().GetSlackIntegrationChannel(auth, accountName, channelName)

			if err != nil {
				if resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving slack_channel: %s", err.Error())
				}
			} else {
				return fmt.Errorf("slack_channel %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}

func createSlackIntegration(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := provider.CommunityClient
		slackIntegration := communityClient.IntegrationSlackRequest{
			ServiceHooks: []communityClient.ServiceHookSlackRequest{
				{
					Account: communityClient.String("test_account"),
					Url:     communityClient.String("https://ddog-client-test.slack.com/fake-account-hook"),
				},
			},
		}

		if err := client.CreateIntegrationSlack(&slackIntegration); err != nil {
			return err
		}

		return nil
	}
}
