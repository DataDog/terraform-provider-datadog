package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/google/uuid"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccOpenapiApiBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogOpenapiApiDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOpenapiApi(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogOpenapiApiExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogOpenapiApi(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_openapi_api" "test-api" {
    spec = <<EOT
info:
  contact:
    name: API team
  description: My API description.
  title: %s
openapi: 3.0.0
paths:
  /api/my-api:
    get:
      description: An endpoint
    EOT
}`, uniq)
}

func testAccCheckDatadogOpenapiApiDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := OpenapiApiDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func OpenapiApiDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_openapi_api" {
				continue
			}
			id := r.Primary.ID

			uuid, _ := uuid.Parse(id)
			_, httpResp, err := apiInstances.GetAPIManagementApiV2().GetOpenAPI(auth, uuid)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving OpenapiApi %s", err)}
			}
			return &utils.RetryableError{Prob: "OpenapiApi still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogOpenapiApiExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := openapiApiExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func openapiApiExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_openapi_api" {
			continue
		}
		id := r.Primary.ID

		uuid, _ := uuid.Parse(id)
		_, httpResp, err := apiInstances.GetAPIManagementApiV2().GetOpenAPI(auth, uuid)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving OpenapiApi")
		}
	}
	return nil
}
