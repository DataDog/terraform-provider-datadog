package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_user_email_notification_rule_test.tf
var OnCallUserEmailNotificationRuleTest string

//go:embed resource_datadog_on_call_user_phone_notification_rule_test.tf
var OnCallUserPhoneNotificationRuleTest string

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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallUserNotificationRuleDestroy(providers.frameworkProvider),
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
				),
			},
		},
	})
}

func TestOnCallUserPhoneNotificationRule(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	userPhone := "+17039389548"

	tfConfig := strings.NewReplacer(
		"USER_EMAIL", strings.ToLower(uniqueEntityName(ctx, t))+"@example.com",
		"USER_PHONE", userPhone,
		"USER_PHONE", userPhone,
	).Replace(OnCallUserPhoneNotificationRuleTest)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallUserNotificationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					testUserNotificationRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_rule.on_call_user_phone_rule_test", "phone.method", "voice"),
				),
			},
		},
	})
}

func testUserNotificationRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		found := false
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

			found = true
		}

		if !found {
			return errors.New("datadog_on_call_user_notification_rule resource is missing in state")
		}

		return nil
	}
}

func testAccCheckDatadogOnCallUserNotificationRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_on_call_user_notification_rule" {
					continue
				}
				id := r.Primary.ID
				userID := r.Primary.Attributes["user_id"]

				_, httpResp, err := apiInstances.GetOnCallApiV2().GetUserNotificationRule(auth, userID, id)
				if httpResp.StatusCode == http.StatusNotFound {
					return nil
				}
				if err != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving user notification rule %s", err)}
				}
				return &utils.RetryableError{Prob: "user notification rule still exists"}
			}
			return nil
		})
	}
}
