package test

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_escalation_policy_test.tf
var OnCallEscalationPolicyTest string

func TestAccOnCallEscalationPolicyCreateAndUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	namePrefix := "team-" + uniq
	handlePrefix := "team-" + uniq

	createConfig := func(effectiveDate string) string {
		return strings.NewReplacer(
			"USER_EMAIL", userEmail,
			"POLICY_NAME", uniq,
			"UNIQ", uniq,
			"TEAM_HANDLE", handlePrefix,
			"TEAM_NAME", namePrefix,
		).Replace(OnCallEscalationPolicyTest)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallEscalationPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createConfig("2025-01-01T00:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallEscalationPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "time_zone", "America/New_York"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.effective_date", "2025-01-01T00:00:00Z"),
					resource.TestCheckResourceAttrWith(
						"datadog_on_call_schedule.single_layer", "layer.0.applied_effective_date", func(value string) error {
							return testAppliedEffectiveDate(value, "2025-01-01T00:00:00Z")
						}),
				),
			},
		},
	})
}

func testAccCheckDatadogOnCallEscalationPolicyExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_on_call_escalation_policy" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetOnCallApiV2().GetOnCallSchedule(auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving OnCallEscalationPolicy")
			}
		}
		return nil
	}
}

func testAccCheckDatadogOnCallEscalationPolicyDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return utils.Retry(2, 10, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "resource_datadog_on_call_escalation_policy" {
					continue
				}
				id := r.Primary.ID

				_, httpResp, err := apiInstances.GetOnCallApiV2().GetOnCallEscalationPolicy(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving OnCallEscalationPolicy %s", err)}
				}
				return &utils.RetryableError{Prob: "OnCallEscalationPolicy still exists"}
			}
			return nil
		})
	}
}
