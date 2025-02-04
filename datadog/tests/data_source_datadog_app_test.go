package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogAppDatasource_Inline_Basic(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app.test_app_inline_basic"
	datasourceName := "data." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testInlineBasicAppDataSourceConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppExists(providers.frameworkProvider, datasourceName),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestCheckResourceAttrSet(datasourceName, "app_json"),
					resource.TestMatchResourceAttr(datasourceName, "app_json", regexp.MustCompile(`\"name\":\"`+appName+`\"`)),
				),
			},
		},
	})
}

func testInlineBasicAppDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	%s
	data "datadog_app" "test_app_inline_basic" {
		id = datadog_app.test_app_inline_basic.id
	}`, testInlineBasicAppResourceConfig(name))
}

func testInlineBasicAppResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_app" "test_app_inline_basic" {
		app_json = jsonencode(
			%s
		)
	}`, testInlineBasicAppJSON(name))
}

func testInlineBasicAppJSON(name string) string {
	// for the sake of this test, we are only interested in the name
	return fmt.Sprintf(`
		{
			"queries" : [],
			"components" : [
				{
					"events" : [],
					"name" : "grid0",
					"properties" : {
						"children" : []
					},
					"type" : "grid"
				}
			],
			"description" : "Test app created using the Datadog provider in Terraform.",
			"name" : "%s",
			"favorite" : false,
			"rootInstanceName" : "grid0",
			"selfService" : false,
			"tags" : []
		}
	`, name)
}

func testAccCheckDatadogAppDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources[resourceName]
		id, err := uuid.Parse(resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error parsing id as uuid: %s", err)
		}
		_, httpRes, err := apiInstances.GetAppBuilderApiV2().GetApp(auth, id)
		if err != nil {
			if httpRes.StatusCode == 404 {
				return nil
			}
			return err
		}

		return fmt.Errorf("Failed app destroy check")
	}
}
