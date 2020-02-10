package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

func TestAccDatadogIntegrationSlack_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationSlackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationSlackConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelsExists("test1", "#test1_chan1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "account", "test1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "url", "http:///myslack/url1"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationSlack_TwoServices(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationSlackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationSlackConfig_TwoChannels,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelsExists("test2", "#channel1"),
					testAccCheckDatadogIntegrationSlackChannelsExists("test2", "#channel2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "account", "test2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo", "url", "http:///myslack/url2"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationSlack_RenameAccount(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogIntegrationSlackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationSlackConfig_TwoAccount,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelsExists("test3_1", "#test3_1_chan1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo1", "account", "test3_1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo1", "url", "http:///myslack/url3_1"),
					testAccCheckDatadogIntegrationSlackChannelsExists("test3_2", "#test3_2_chan1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo2", "account", "test3_2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo2", "url", "http:///myslack/url3_2"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationSlackConfig_TwoAccountRenamed,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationSlackChannelsExists("test3_1", "#test3_1_chan1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo1", "account", "test3_1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo1", "url", "http:///myslack/url3_1"),
					testAccCheckDatadogIntegrationSlackChannelsNotExists("test3_2", "#test3_2_chan1"), // Chan should'nt exist anymore
					testAccCheckDatadogIntegrationSlackChannelsExists("test3_newname", "#test3_2_chan1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo2", "account", "test3_newname"),
					resource.TestCheckResourceAttr(
						"datadog_integration_slack.foo2", "url", "http:///myslack/url3_2"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationSlackChannelsExists(accountName, channelName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		slackIntegration, err := client.GetIntegrationSlack()
		if err != nil {
			return fmt.Errorf("Received an error retrieving integration slack %s", err)
		}
		for _, channel := range slackIntegration.Channels {
			if channel.GetChannelName() == channelName && channel.GetAccount() == accountName {
				return nil
			}
		}
		return fmt.Errorf("Didn't find channel (%s) in integration slack", channelName)
	}
}

func testAccCheckDatadogIntegrationSlackChannelsNotExists(accountName, channelName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		slackIntegration, err := client.GetIntegrationSlack()
		if err != nil {
			return fmt.Errorf("Received an error retrieving integration slack %s", err)
		}
		for _, channel := range slackIntegration.Channels {
			if channel.GetChannelName() == channelName && channel.GetAccount() == accountName {
				return fmt.Errorf("find bad channel (%s) in integration slack", channelName)
			}
		}
		return nil
	}
}

func testAccCheckDatadogIntegrationSlackDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	_, err := client.GetIntegrationSlack()
	if err != nil {
		if strings.Contains(err.Error(), "slack not found") {
			return nil
		}

		return fmt.Errorf("Received an error retrieving integration slack %s", err)
	}

	return fmt.Errorf("Integration slack is not properly destroyed")
}

const testAccCheckDatadogIntegrationSlackConfig = `
resource "datadog_integration_slack" "foo" {
	account = "test1"
	url = "http:///myslack/url1"
   
   channels {
	   channel_name = "#test1_chan1"
	   transfer_all_user_comments = false       
   }
}
 `

const testAccCheckDatadogIntegrationSlackConfig_TwoChannels = `
 locals {
	 slack_channels = [
		 "#channel1", "#channel2"
	 ]
 }
 resource "datadog_integration_slack" "foo" {
	account = "test2"
	url = "http:///myslack/url2"
   

  dynamic "channels" {
		for_each = local.slack_channels
		content {
			channel_name = channels.value
	   		transfer_all_user_comments = false       
		}
	}
}
`
const testAccCheckDatadogIntegrationSlackConfig_TwoAccount = `
resource "datadog_integration_slack" "foo1" {
	account = "test3_1"
	url = "http:///myslack/url3_1"
   
   channels {
	   channel_name = "#test3_1_chan1"
	   transfer_all_user_comments = false       
   }
}
resource "datadog_integration_slack" "foo2" {
	account = "test3_2"
	url = "http:///myslack/url3_2"
   
   channels {
	   channel_name = "#test3_2_chan1"
	   transfer_all_user_comments = false       
   }
}
`
const testAccCheckDatadogIntegrationSlackConfig_TwoAccountRenamed = `
resource "datadog_integration_slack" "foo1" {
	account = "test3_1"
	url = "http:///myslack/url3_1"
   
   channels {
	   channel_name = "#test3_1_chan1"
	   transfer_all_user_comments = false       
   }
}
resource "datadog_integration_slack" "foo2" {
	account = "test3_newname"
	url = "http:///myslack/url3_2"
   
   channels {
	   channel_name = "#test3_2_chan1"
	   transfer_all_user_comments = false       
   }
}
`
