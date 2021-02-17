package test

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	communityClient "github.com/zorkian/go-datadog-api"
)

func TestAccDatadogIntegrationSlackChannel_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	uniqueChannelTeamName := reg.ReplaceAllString(uniqueEntityName(clock, t), "")
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogIntegrationSlackChannelDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				// Workaround to ensure we create the slack integration before running the tests.
				Config: emptyLogsArchiveConfig(),
				Check: resource.ComposeTestCheckFunc(
					createSlackIntegration(accProvider),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackChannelConfig_Create(uniqueChannelTeamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelExists(accProvider, "datadog_integration_slack_channel.slack_channel"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "channel_name", "#"+uniqueChannelTeamName),
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
				Config: testAccCheckDatadogIntegrationSlackChannelConfig_Update(uniqueChannelTeamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelExists(accProvider, "datadog_integration_slack_channel.slack_channel"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack_channel.slack_channel", "channel_name", "#"+uniqueChannelTeamName),
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

func testAccCheckDatadogIntegrationSlackChannelConfig_Create(uniq string) string {
	return fmt.Sprintf(`
       resource "datadog_integration_slack_channel" "slack_channel" {
			display {
				message = true
				notified = true
				snapshot = true
				tags = true
			}
			channel_name = "#%s"
			team_name    = "test_team"
       }
   `, uniq)
}

func testAccCheckDatadogIntegrationSlackChannelConfig_Update(uniq string) string {
	return fmt.Sprintf(`
       resource "datadog_integration_slack_channel" "slack_channel" {
			display {
				message = false
				notified = false
				snapshot = false
				tags = false
			}
			channel_name = "#%s"
			team_name    = "test_team"
       }
   `, uniq)
}

func emptyLogsArchiveConfig() string {
	return fmt.Sprintf(`
       resource "datadog_logs_archive_order" "archives" {
		}
   `)
}

func testAccCheckDatadogIntegrationSlackChannelExists(accProvider *schema.Provider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1

		teamName := s.RootModule().Resources[resourceName].Primary.Attributes["team_name"]
		channelName := s.RootModule().Resources[resourceName].Primary.Attributes["channel_name"]

		_, _, err := datadogClient.SlackIntegrationApi.GetSlackIntegrationChannel(auth, teamName, channelName).Execute()
		if err != nil {
			return utils.TranslateClientError(err, "error checking slack_channel existence")
		}

		return nil
	}
}

func testAccCheckDatadogIntegrationSlackChannelDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		meta := accProvider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_slack_channel" {
				continue
			}

			teamName := r.Primary.Attributes["team_name"]
			channelName := r.Primary.Attributes["channel_name"]

			_, resp, err := datadogClient.SlackIntegrationApi.GetSlackIntegrationChannel(auth, teamName, channelName).Execute()

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

func createSlackIntegration(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		client := providerConf.CommunityClient
		slackIntegration := communityClient.IntegrationSlackRequest{
			ServiceHooks: []communityClient.ServiceHookSlackRequest{
				{
					Account: communityClient.String("test_team"),
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
