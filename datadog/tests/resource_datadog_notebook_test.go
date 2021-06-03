package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogNotebookBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueNotebook := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogNotebookDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogNotebookConfigBasic(uniqueNotebook),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogNotebookExists(accProvider, "datadog_notebook.testing_notebook"),
					// FIXME: add steps to verify attributes were set
				),
			},
		},
	})
}

func testAccCheckDatadogNotebookConfigBasic(uniq string) string {
	// FIXME: implement usage of the `uniq` argument as a title/name/description of the created entity.
	// This ensures uniqueness in case of parallel-running test cases and easier trackability when
	// cleanup fails.
	return fmt.Sprintf(`
        # uniq: %s
        resource "datadog_notebook" "testing_notebook" {

			author {
				created_at = "2222-12-12T12:12:12.123456+0000"
				disabled = true
				email = "example email"
				handle = "example handle"
				icon = "example icon"
				name = "example name"
				status = "example status"
				title = "example title"
				verified = true
			}
			// FIXME: array of objects in Cells
			name = "Example Notebook"
			status = "example status"

			time {
				live_span = "example live_span"
				end = "2021-02-24T20:18:28Z"
				live = true
				start = "2021-02-24T19:18:28Z"
			}
        }
    `, uniq)
}

func testAccCheckDatadogNotebookExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		meta := provider.Meta()
		resourceId := s.RootModule().Resources[resourceName].Primary.ID
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		var err error

		id, err := strconv.ParseInt(resourceId, 10, 64)
		if err != nil {
			return err
		}
		_, _, err = datadogClient.NotebooksApi.GetNotebook(auth, id)

		if err != nil {
			return utils.TranslateClientError(err, "error checking notebook existence")
		}

		return nil
	}
}

func testAccCheckDatadogNotebookDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		meta := provider.Meta()
		providerConf := meta.(*datadog.ProviderConfiguration)
		datadogClient := providerConf.DatadogClientV1
		auth := providerConf.AuthV1
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_notebook" {
				continue
			}

			var err error

			id, err := strconv.ParseInt(r.Primary.ID, 10, 64)
			if err != nil {
				return err
			}
			_, resp, err := datadogClient.NotebooksApi.GetNotebook(auth, id)

			if err != nil {
				if resp != nil && resp.StatusCode == 404 {
					continue // resource not found => all ok
				} else {
					return fmt.Errorf("received an error retrieving notebook: %s", err.Error())
				}
			} else {
				return fmt.Errorf("notebook %s still exists", r.Primary.ID)
			}
		}

		return nil
	}
}
