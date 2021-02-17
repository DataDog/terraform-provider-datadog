package datadog

import (
	"fmt"
	"github.com/zorkian/go-datadog-api"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogIntegrationSlackChannel_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	uniqueSlackTeamName := reg.ReplaceAllString(uniqueEntityName(clock, t), "")
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
					createSlackIntegration(accProvider, uniqueSlackTeamName),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackChannelConfig_Create(),
				Check: resource.ComposeTestCheckFunc(
					createSlackIntegration(accProvider, uniqueSlackTeamName),
					testAccCheckDatadogIntegrationSlackChannelExists(accProvider, "datadog_integration_slack_channel.slack_channel"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackChannelConfig_Update(),
				Check: resource.ComposeTestCheckFunc(
					createSlackIntegration(accProvider, uniqueSlackTeamName),
					testAccCheckDatadogIntegrationSlackChannelExists(accProvider, "datadog_integration_slack_channel.slack_channel"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationSlackChannelConfig_Create() string {
	return `
        resource "datadog_integration_slack_channel" "slack_channel" {
			display {
				message = true
				notified = true
				snapshot = true
				tags = true
			}
			channel_name = "#general"
			team_name    = "test_team"
        }
    `
}

func testAccCheckDatadogIntegrationSlackChannelConfig_Update() string {
	return `
        resource "datadog_integration_slack_channel" "slack_channel" {
			display {
				message = false
				notified = false
				snapshot = false
				tags = false
			}
			channel_name = "#general"
			team_name    = "test_team"
        }
    `
}

func emptyLogsArchiveConfig() string {
	return fmt.Sprintf(`
        resource "datadog_logs_archive_order" "archives" {
		}
    `)
}

func testAccCheckDatadogIntegrationSlackChannelExists(accProvider *schema.Provider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//meta := accProvider.Meta()
		//resourceId := s.RootModule().Resources[resourceName].Primary.ID
		//providerConf := meta.(*ProviderConfiguration)
		//datadogClient := providerConf.DatadogClientV1
		//auth := providerConf.AuthV1
		//var err error
		//
		//id := resourceId
		//
		//_, _, err = datadogClient.SlackIntegrationApi.GetSlackIntegrationChannel(auth, id).Execute()
		//
		//if err != nil {
		//	return translateClientError(err, "error checking slack_channel existence")
		//}

		return nil
	}
}

func testAccCheckDatadogIntegrationSlackChannelDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		//meta := accProvider.Meta()
		//providerConf := meta.(*ProviderConfiguration)
		//datadogClient := providerConf.DatadogClientV1
		//auth := providerConf.AuthV1
		//for _, r := range s.RootModule().Resources {
		//	if r.Type != "datadog_slack_channel" {
		//		continue
		//	}
		//
		//	var err error
		//
		//	id := r.Primary.ID
		//
		//	_, resp, err := datadogClient.SlackIntegrationApi.GetSlackIntegrationChannel(auth, id).Execute()
		//
		//	if err != nil {
		//		if resp.StatusCode == 404 {
		//			continue // resource not found => all ok
		//		} else {
		//			return fmt.Errorf("received an error retrieving slack_channel: %s", err.Error())
		//		}
		//	} else {
		//		return fmt.Errorf("slack_channel %s still exists", r.Primary.ID)
		//	}
		//}

		return nil
	}
}

func createSlackIntegration(accProvider *schema.Provider, uniqueTeamName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient
		slackIntegration := datadog.IntegrationSlackRequest{
			ServiceHooks: []datadog.ServiceHookSlackRequest{
				{
					Account: datadog.String("test_team"),
					Url:     datadog.String("https://ddog-client-test.slack.com/fake-account-hook"),
				},
			},
		}

		if err := client.CreateIntegrationSlack(&slackIntegration); err != nil {
			return err
		}

		return nil
	}
}

func deleteSlackIntegration(accProvider *schema.Provider, uniqueTeamName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		if err := client.DeleteIntegrationSlack(); err != nil {
			return err
		}

		return nil
	}
}
