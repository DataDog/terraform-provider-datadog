package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccMonitorNotificationRule_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorNotificationRule(uniq),
			},
			{
				ResourceName:      "datadog_monitor_notification_rule.r",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMonitorNotificationRule_Create(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorNotificationRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor_notification_rule.r", "name", "A notification rule name"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "recipients.*", "slack-foo"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "recipients.*", "jira-bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", fmt.Sprintf("env:%s", uniq)),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", "host:abc"),
				),
			},
		},
	})
}

func TestAccMonitorNotificationRule_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	uniq_updated := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogMonitorNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorNotificationRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor_notification_rule.r", "name", "A notification rule name"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "recipients.*", "slack-foo"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "recipients.*", "jira-bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", fmt.Sprintf("env:%s", uniq)),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", "host:abc"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorNotificationRule_update(uniq_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor_notification_rule.r", "name", "Updated rule name"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "recipients.*", "jira-foo"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "recipients.*", "jira-bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", fmt.Sprintf("env:%s", uniq_updated)),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", "host:abc"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor_notification_rule.r", "filter.tags.*", "poke:mon"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorNotificationRule(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_monitor_notification_rule" "r" {
    name = "A notification rule name"
    recipients = ["slack-foo", "jira-bar"]
	filter {
	  tags = ["env:%s", "host:abc"]
	}
}`, uniq)
}

func testAccCheckDatadogMonitorNotificationRule_update(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_monitor_notification_rule" "r" {
    name = "Updated rule name"
    recipients = ["jira-foo", "jira-bar"]
	filter {
	  tags = ["env:%s", "host:abc", "poke:mon"]
	}
}`, uniq)
}

func testAccCheckDatadogMonitorNotificationRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := MonitorNotificationRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func MonitorNotificationRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_monitor_notification_rule" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetMonitorsApiV2().GetMonitorNotificationRule(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving MonitorNotificationRule %s", err)}
			}
			return &utils.RetryableError{Prob: "MonitorNotificationRule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogMonitorNotificationRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := monitorNotificationRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func monitorNotificationRuleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_monitor_notification_rule" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetMonitorsApiV2().GetMonitorNotificationRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving MonitorNotificationRule")
		}
	}
	return nil
}
