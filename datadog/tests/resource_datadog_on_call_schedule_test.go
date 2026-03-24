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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createConfig("2025-01-01T00:00:00-08:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "time_zone", "America/New_York"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.effective_date", "2025-01-01T00:00:00-08:00"),
				),
			},
			// Update the effective date
			{
				Config: createConfig("2025-02-01T00:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.effective_date", "2025-02-01T00:00:00Z"),
				),
			},
			// Add a layer on first position
			{
				Config: addLayer(createConfig("2025-02-01T00:00:00Z"), `
					layer {
						effective_date = "2026-01-01T00:00:00Z"
    					interval {
      						days = 2
    					}
    					rotation_start = "2026-01-01T00:00:00Z"
    					users = [null]
    					name = "Added Layer"
						time_zone = "Asia/Tokyo"
					}
				`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.name", "Added Layer"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.time_zone", "Asia/Tokyo"),
					// Existing layer is not modified
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.1.name", "Primary On-Call Layer"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.1.effective_date", "2025-02-01T00:00:00Z"),
				),
			},
		},
	})
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

		if httpResp.StatusCode == 404 {
			return errors.New("OnCallSchedule does not exist")
		}
	}
	return nil
}
