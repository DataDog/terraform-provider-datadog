package test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

//go:embed resource_datadog_on_call_schedule_test.tf
var OnCallScheduleTest string

//go:embed resource_datadog_on_call_schedule_swap_test.tf
var OnCallScheduleSwapTest string

func TestAccOnCallScheduleCreateAndUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	userEmail := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	namePrefix := "team-" + uniq
	handlePrefix := "team-" + uniq

	// The API rejects a schedule with no layer active at the current time, so dates are relative to the test clock.
	now := clockFromContext(ctx).Now().UTC().Truncate(24 * time.Hour)
	firstEffectiveDate := now.AddDate(0, -1, 0).Format("2006-01-02") + "T00:00:00-08:00"
	secondEffectiveDate := now.Format(time.RFC3339)
	endDate := now.AddDate(1, 0, 0).Add(-5 * time.Hour).Format(time.RFC3339)
	rotationStart := now.AddDate(-1, 0, 0).Format("2006-01-02") + "T00:00:00-09:00"
	addedLayerDate := now.AddDate(1, 0, 0).Format(time.RFC3339)

	createConfig := func(effectiveDate string) string {
		return strings.NewReplacer(
			"USER_EMAIL", userEmail,
			"SCHEDULE_NAME", uniq,
			"EFFECTIVE_DATE", effectiveDate,
			"END_DATE", endDate,
			"ROTATION_START", rotationStart,
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
				Config: createConfig(firstEffectiveDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "time_zone", "America/New_York"),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.effective_date", firstEffectiveDate),
				),
			},
			// Update the effective date but the layer id must remain known (not "known after apply")
			{
				Config: createConfig(secondEffectiveDate),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"datadog_on_call_schedule.single_layer",
							tfjsonpath.New("layer").AtSliceIndex(0).AtMapKey("id"),
							knownvalue.NotNull(),
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_on_call_schedule.single_layer", "layer.0.effective_date", secondEffectiveDate),
				),
			},
			// Add a layer on first position
			{
				Config: addLayer(createConfig(secondEffectiveDate), fmt.Sprintf(`
					layer {
						effective_date = "%[1]s"
    					interval {
      						days = 2
    					}
    					rotation_start = "%[1]s"
    					users = [null]
    					name = "Added Layer"
						time_zone = "Asia/Tokyo"
					}
				`, addedLayerDate)),
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
						"datadog_on_call_schedule.single_layer", "layer.1.effective_date", secondEffectiveDate),
				),
			},
		},
	})
}

func TestAccOnCallScheduleLayerIDSwapNames(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	swapConfig := func(layerAName, layerBName string) string {
		return strings.NewReplacer(
			"SWAP_SCHEDULE_NAME", uniq,
			"LAYER_A_NAME", layerAName,
			"LAYER_B_NAME", layerBName,
		).Replace(OnCallScheduleSwapTest)
	}

	var idFirst, idSecond string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOnCallScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: swapConfig("Layer A", "Layer B"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrWith("datadog_on_call_schedule.swap", "layer.0.id", func(v string) error { idFirst = v; return nil }),
					resource.TestCheckResourceAttrWith("datadog_on_call_schedule.swap", "layer.1.id", func(v string) error { idSecond = v; return nil }),
				),
			},
			// Swap layer names: each layer must keep its own ID in the plan, not the other's.
			{
				Config: swapConfig("Layer B", "Layer A"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("datadog_on_call_schedule.swap", tfjsonpath.New("layer").AtSliceIndex(0).AtMapKey("id"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue("datadog_on_call_schedule.swap", tfjsonpath.New("layer").AtSliceIndex(1).AtMapKey("id"), knownvalue.NotNull()),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOnCallScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrWith("datadog_on_call_schedule.swap", "layer.0.id", func(v string) error {
						if v != idSecond {
							return fmt.Errorf("layer 0 (now %q) expected id %q, got %q", "Layer B", idSecond, v)
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("datadog_on_call_schedule.swap", "layer.1.id", func(v string) error {
						if v != idFirst {
							return fmt.Errorf("layer 1 (now %q) expected id %q, got %q", "Layer A", idFirst, v)
						}
						return nil
					}),
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
