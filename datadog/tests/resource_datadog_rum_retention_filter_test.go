package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// test application of DD Integration Tests org (321813) in us1.prod.dog
const RumRetentionFilterResourceTestAppId = "9ff07c10-11f9-402c-a9d4-9eca42ef4a64"

func TestAccRumRetentionFilterImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_rum_retention_filter.testing_rum_retention_filter_for_import"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumRetentionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: minimalDatadogRumRetentionFilterForImport(uniqueEntityName(ctx, t)),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resources := state.RootModule().Resources
					resourceState := resources[resourceName]
					return RumRetentionFilterResourceTestAppId + ":" + resourceState.Primary.Attributes["id"], nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRumRetentionFilterDecimalSampleRate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumRetentionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: withDecimalSampleRateDatadogRumRetentionFilter(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "application_id", RumRetentionFilterResourceTestAppId),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "name", name),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "event_type", "session"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "sample_rate", "50.5"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "query", "custom_query"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccRumRetentionFilterAttributes(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name1 := uniqueEntityName(ctx, t)
	name2 := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumRetentionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: minimalDatadogRumRetentionFilter(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "application_id", RumRetentionFilterResourceTestAppId),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "name", name1),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "event_type", "session"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "sample_rate", "25"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "query", ""),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "enabled", "true"),
				),
			},
			{
				Config: fullDatadogRumRetentionFilter(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "application_id", RumRetentionFilterResourceTestAppId),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "name", name1),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "event_type", "action"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "sample_rate", "50"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "query", "custom_query_1"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "enabled", "false"),
				),
			},
			{
				Config: withQueryDatadogRumRetentionFilter(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "application_id", RumRetentionFilterResourceTestAppId),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "name", name2),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "event_type", "view"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "sample_rate", "75"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "query", "custom_query_2"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "enabled", "true"),
				),
			},

			{
				Config: withEnabledDatadogRumRetentionFilter(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "application_id", RumRetentionFilterResourceTestAppId),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "name", name2),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "event_type", "session"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "sample_rate", "100"),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "query", ""),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "enabled", "false"),
				),
			},
		},
	})
}

func minimalDatadogRumRetentionFilterForImport(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter_for_import" {
		application_id = %q
		name = %q
	    event_type = "session"
		sample_rate = 25
	}
	`, RumRetentionFilterResourceTestAppId, name)
}

func minimalDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		application_id = %q
		name = %q
	    event_type = "session"
		sample_rate = 25
	}
	`, RumRetentionFilterResourceTestAppId, name)
}

func fullDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		application_id = %q
		name = %q
	    event_type = "action"
		sample_rate = 50
		query = "custom_query_1"
		enabled = false
	}
	`, RumRetentionFilterResourceTestAppId, name)
}

func withQueryDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		application_id = %q
		name = %q
	    event_type = "view"
		sample_rate = 75
		query = "custom_query_2"
	}
	`, RumRetentionFilterResourceTestAppId, name)
}

func withEnabledDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		application_id = %q
		name = %q
	    event_type = "session"
		sample_rate = 100
		enabled = false
	}

	`, RumRetentionFilterResourceTestAppId, name)
}

func withDecimalSampleRateDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		application_id = %q
		name = %q
	    event_type = "session"
		sample_rate = 50.5
		query = "custom_query"
		enabled = true
	}
	`, RumRetentionFilterResourceTestAppId, name)
}

func testAccCheckDatadogRumRetentionFilterDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := RumRetentionFilterDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func RumRetentionFilterDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 5, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_rum_retention_filter" {
				continue
			}

			_, httpResp, err := apiInstances.GetRumRetentionFiltersApiV2().GetRetentionFilter(auth, RumRetentionFilterResourceTestAppId, r.Primary.ID)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}

				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving RumRetentionFilter %s", err)}
			}
			return &utils.RetryableError{Prob: "RumRetentionFilter still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogRumRetentionFilterExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := rumRetentionFilterExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func rumRetentionFilterExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_rum_retention_filter" {
			continue
		}

		res, httpResp, err := apiInstances.GetRumRetentionFiltersApiV2().GetRetentionFilter(auth, RumRetentionFilterResourceTestAppId, r.Primary.ID)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving RumRetentionFilter")
		}

		// Check source is terraform for retention filter created through terraform
		meta, ok := res.Data.AdditionalProperties["meta"].(map[string]interface{})
		if !ok {
			return errors.New("'meta' must exist in `data`")
		}

		source, ok := meta["source"].(string)
		if !ok {
			return errors.New("'source' must be a string in 'meta'")
		}

		if source != "terraform" {
			return errors.New("'source' must be 'terraform'")
		}
	}
	return nil
}
