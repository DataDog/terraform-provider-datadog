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

const AppId = "17a2877d-5a77-406e-9039-9da24714936e"

func TestAccRumRetentionFilterImport(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_rum_retention_filter.testing_rum_retention_filter"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumRetentionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: minimalDatadogRumRetentionFilter(uniqueEntityName(ctx, t)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRumRetentionFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: minimalDatadogRumRetentionFilter(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRumRetentionFilterExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_rum_retention_filter.testing_rum_retention_filter", "app_id", AppId),
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
						"datadog_rum_retention_filter.testing_rum_retention_filter", "app_id", AppId),
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
						"datadog_rum_retention_filter.testing_rum_retention_filter", "app_id", AppId),
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
						"datadog_rum_retention_filter.testing_rum_retention_filter", "app_id", AppId),
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

func minimalDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		app_id = %q
		name = %q
	    event_type = "session"
		sample_rate = 25
	}
	`, AppId, name)
}

func fullDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		app_id = %q
		name = %q
	    event_type = "action"
		sample_rate = 50
		query = "custom_query_1"
		enabled = false
	}
	`, AppId, name)
}

func withQueryDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		app_id = %q
		name = %q
	    event_type = "view"
		sample_rate = 75
		query = "custom_query_2"
	}
	`, AppId, name)
}

func withEnabledDatadogRumRetentionFilter(name string) string {
	return fmt.Sprintf(`resource "datadog_rum_retention_filter" "testing_rum_retention_filter" {
		app_id = %q
		name = %q
	    event_type = "session"
		sample_rate = 100
		enabled = false
	}

	`, AppId, name)
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
			appId, retentionFilterId, err := fwprovider.ParseRetentionFilterId(r.Primary.ID)
			if err != nil {
				return err
			}

			_, httpResp, err := apiInstances.GetRumRetentionFiltersApiV2().GetRetentionFilter(auth, appId, retentionFilterId)
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
		appId, retentionFilterId, err := fwprovider.ParseRetentionFilterId(r.Primary.ID)
		if err != nil {
			return err
		}

		res, httpResp, err := apiInstances.GetRumRetentionFiltersApiV2().GetRetentionFilter(auth, appId, retentionFilterId)
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
