package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_user_email_notification_channel_test.tf
var OnCallUserEmailNotificationChannelTest string

//go:embed resource_datadog_on_call_user_phone_notification_channel_test.tf
var OnCallUserPhoneNotificationChannelTest string

func TestOnCallUserEmailNotificationChannel(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"

	tfConfig := strings.NewReplacer(
		"USER_EMAIL", userEmail,
	).Replace(OnCallUserEmailNotificationChannelTest)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallUserNotificationChannelDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					testUserNotificationChannelExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_channel.user_email_channel_test", "email.address", userEmail),
				),
			},
		},
	})
}

func TestOnCallUserPhoneNotificationChannel(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	userPhone := "+17039389548"

	tfConfig := strings.NewReplacer(
		"USER_EMAIL", strings.ToLower(uniqueEntityName(ctx, t))+"@example.com",
		"USER_PHONE", userPhone,
	).Replace(OnCallUserPhoneNotificationChannelTest)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallUserNotificationChannelDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					testUserNotificationChannelExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_on_call_user_notification_channel.user_phone_channel_test", "phone.number", userPhone),
				),
			},
		},
	})
}

func testUserNotificationChannelExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		found := false
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_on_call_user_notification_channel" {
				continue
			}

			id := r.Primary.ID
			userID := r.Primary.Attributes["user_id"]

			channel, httpResp, err := apiInstances.GetOnCallApiV2().GetUserNotificationChannel(auth, userID, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving user notification channel")
			}

			if channel.Data == nil {
				return errors.New("user notification channel not found")
			}

			found = true
		}

		if !found {
			return errors.New("on_call_user_notification_channel resource is missing in state")
		}

		return nil
	}
}

func testAccCheckDatadogOnCallUserNotificationChannelDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_on_call_user_notification_channel" {
					continue
				}
				id := r.Primary.ID
				userID := r.Primary.Attributes["user_id"]

				channel, _, err := apiInstances.GetOnCallApiV2().GetUserNotificationChannel(auth, userID, id)
				if err != nil {
					if channel.Data == nil {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving user notification channel %s", err)}
				}
				return &utils.RetryableError{Prob: "ser notification channel still exists"}
			}
			return nil
		})
	}
}
