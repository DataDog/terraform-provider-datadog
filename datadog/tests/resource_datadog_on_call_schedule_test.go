package test

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_schedule_test.tf
var OnCallScheduleTest string

func TestAccOnCallScheduleCreateAndUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	namePrefix := "team-" + uniq
	handlePrefix := "team-" + uniq

	createConfig := func(effectiveDate string) string {
		return strings.NewReplacer(
			"USER_EMAIL", userEmail,
			"SCHEDULE_NAME", uniq,
			"EFFECTIVE_DATE", effectiveDate,
			"TEAM_HANDLE", handlePrefix,
			"TEAM_NAME", namePrefix,
		).Replace(OnCallScheduleTest)
	}

	addLayer := func(source string, layer string) string {
		return strings.NewReplacer(
			"layer {", layer+"layer {",
		).Replace(source)
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createConfig("2025-01-01T00:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
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
			// Update the effective date
			{
				Config: createConfig("2025-02-01T00:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.effective_date", "2025-02-01T00:00:00Z"),
					resource.TestCheckResourceAttrWith(
						"datadog_on_call_schedule.single_layer", "layer.0.applied_effective_date", func(value string) error {
							return testAppliedEffectiveDate(value, "2025-01-01T00:00:00Z")
						}),
				),
			},
			// Add a layer
			{
				Config: addLayer(createConfig("2025-02-01T00:00:00Z"), `
					layer {
						effective_date = "2026-01-01T00:00:00Z"
    					interval {
      						days = 2
    					}
    					rotation_start = "2026-01-01T00:00:00Z"
    					member {}
    					name = "Added Layer"
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.name", "Added Layer"),
					// Existing layer is not modified
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.1.name", "Primary On-Call Layer"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.1.effective_date", "2025-02-01T00:00:00Z"),
					resource.TestCheckResourceAttrWith(
						"datadog_on_call_schedule.single_layer", "layer.1.applied_effective_date", func(value string) error {
							return testAppliedEffectiveDate(value, "2025-01-01T00:00:00Z")
						}),
				),
			},
		},
	})
}

func testAppliedEffectiveDate(appliedEffectiveDate string, effectiveDate string) error {
	appliedEffectiveTime, err := time.Parse(time.RFC3339, appliedEffectiveDate)
	if err != nil {
		return fmt.Errorf("failed to parse effective_date: %s", err)
	}

	effectiveTime, err := time.Parse(time.RFC3339, effectiveDate)
	if err != nil {
		return fmt.Errorf("failed to parse effective_date: %s", err)
	}

	if appliedEffectiveTime.Before(effectiveTime) {
		return fmt.Errorf("applied_effective_date is before effective_date")
	}
	return nil
}

func testAccCheckDatadogOnCallScheduleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := OnCallScheduleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func OnCallScheduleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_on_call_schedule" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetOnCallApiV2().GetOnCallSchedule(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving OnCallSchedule %s", err)}
			}
			return &utils.RetryableError{Prob: "OnCallSchedule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogOnCallScheduleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := onCallScheduleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func onCallScheduleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_on_call_schedule" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetOnCallApiV2().GetOnCallSchedule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving OnCallSchedule")
		}
	}
	return nil
}
