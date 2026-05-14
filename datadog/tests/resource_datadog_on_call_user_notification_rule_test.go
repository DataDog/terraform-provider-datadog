package test

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_user_email_notification_rule_test.tf
var OnCallUserEmailNotificationRuleTest string

//go:embed resource_datadog_on_call_user_phone_notification_rule_test.tf
var OnCallUserPhoneNotificationRuleTest string

//go:embed resource_datadog_on_call_user_email_notification_channel_only_test.tf
var OnCallUserEmailNotificationChannelOnlyTest string

//go:embed resource_datadog_on_call_user_phone_notification_channel_only_test.tf
var OnCallUserPhoneNotificationChannelOnlyTest string

func TestOnCallUserEmailNotificationRule(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	createConfig := strings.NewReplacer(
		"USER_EMAIL", userEmail,
		"DELAY_MINUTES", "1",
	).Replace(OnCallUserEmailNotificationRuleTest)

	updateConfig := strings.NewReplacer(
		"USER_EMAIL", userEmail,
		"DELAY_MINUTES", "0",
	).Replace(OnCallUserEmailNotificationRuleTest)

	channelOnlyConfig := strings.NewReplacer(
		"USER_EMAIL", userEmail,
	).Replace(OnCallUserEmailNotificationChannelOnlyTest)

	var ruleID, ruleUserID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: createConfig,
				Check: resource.ComposeTestCheckFunc(
					testUserNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_rule.on_call_user_email_rule_test", "category", "high_urgency"),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_rule.on_call_user_email_rule_test", "delay_minutes", "1"),
				),
			},
			{
				Config: updateConfig,
				Check: resource.ComposeTestCheckFunc(
					testUserNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_rule.on_call_user_email_rule_test", "category", "high_urgency"),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_rule.on_call_user_email_rule_test", "delay_minutes", "0"),
					// Capture IDs before the rule is destroyed in the next step
					func(s *terraform.State) error {
						r := s.RootModule().Resources["datadog_on_call_user_notification_rule.on_call_user_email_rule_test"]
						ruleID = r.Primary.ID
						ruleUserID = r.Primary.Attributes["user_id"]
						return nil
					},
				),
			},
			{
				// Remove the notification rule from config — Terraform destroys
				// it while the user is still active, so we can verify via 404.
				Config: channelOnlyConfig,
				Check:  testUserNotificationRuleIsDestroyed(providers.frameworkProvider, &ruleUserID, &ruleID),
			},
		},
	})
}

func TestOnCallUserPhoneNotificationRule(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	userPhone := "+17039389548"

	tfConfig := strings.NewReplacer(
		"USER_EMAIL", userEmail,
		"USER_PHONE", userPhone,
	).Replace(OnCallUserPhoneNotificationRuleTest)

	channelOnlyConfig := strings.NewReplacer(
		"USER_EMAIL", userEmail,
		"USER_PHONE", userPhone,
	).Replace(OnCallUserPhoneNotificationChannelOnlyTest)

	var ruleID, ruleUserID string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					testUserNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_rule.on_call_user_phone_rule_test", "phone.method", "voice"),
					// Capture IDs before the rule is destroyed in the next step
					func(s *terraform.State) error {
						r := s.RootModule().Resources["datadog_on_call_user_notification_rule.on_call_user_phone_rule_test"]
						ruleID = r.Primary.ID
						ruleUserID = r.Primary.Attributes["user_id"]
						return nil
					},
				),
			},
			{
				// Remove the notification rule from config — Terraform destroys
				// it while the user is still active, so we can verify via 404.
				Config: channelOnlyConfig,
				Check:  testUserNotificationRuleIsDestroyed(providers.frameworkProvider, &ruleUserID, &ruleID),
			},
		},
	})
}

func testUserNotificationRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_on_call_user_notification_rule" {
				continue
			}

			id := r.Primary.ID
			userID := r.Primary.Attributes["user_id"]

			_, httpResp, err := apiInstances.GetOnCallApiV2().GetUserNotificationRule(auth, userID, id)
			if err != nil || httpResp.StatusCode != http.StatusOK {
				return utils.TranslateClientError(err, httpResp, "error retrieving user notification rule")
			}

			return nil
		}

		return fmt.Errorf("datadog_on_call_user_notification_rule resource is missing in state")
	}
}

func testUserNotificationRuleIsDestroyed(accProvider *fwprovider.FrameworkProvider, userID, ruleID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return utils.Retry(2*time.Second, 10, func() error {
			_, httpResp, _ := apiInstances.GetOnCallApiV2().GetUserNotificationRule(auth, *userID, *ruleID)
			if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
				return nil
			}
			return &utils.RetryableError{Prob: "notification rule still exists after removal from config"}
		})
	}
}
