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

func TestAccDatadogAppBuilderAppDataSource_Inline_WithOptionalFields(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_inline_optional"
	datasourceName := "data." + resourceName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAppBuilderAppInlineWithOptionalFieldsDataSourceConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, datasourceName),
					resource.TestCheckResourceAttrSet(datasourceName, "id"),
					resource.TestMatchResourceAttr(datasourceName, "app_json", regexp.MustCompile(fmt.Sprintf(`"name":"%s"`, appName))),

					resource.TestCheckResourceAttr(datasourceName, "action_query_names_to_connection_ids.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "action_query_names_to_connection_ids.listTeams0", "11111111-2222-3333-4444-555555555555"),

					resource.TestCheckResourceAttr(datasourceName, "name", appName),
					resource.TestCheckResourceAttr(datasourceName, "description", "Created using the Datadog provider in Terraform."),
					resource.TestCheckResourceAttr(datasourceName, "root_instance_name", "grid0"),
					resource.TestCheckResourceAttr(datasourceName, "published", "true"),
				),
			},
		},
	})
}

func testAppBuilderAppInlineWithOptionalFieldsDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	%s
	data "datadog_app_builder_app" "test_app_inline_optional" {
		id = datadog_app_builder_app.test_app_inline_optional.id
	}`, testAppBuilderAppInlineWithOptionalFieldsResourceConfig(name))
}

func testAppBuilderAppInlineWithOptionalFieldsResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_app_builder_app" "test_app_inline_optional" {
		name               = "%s"
		description        = "Created using the Datadog provider in Terraform."
		root_instance_name = "grid0"
		published          = true
		
		action_query_names_to_connection_ids = {
			"listTeams0" = "11111111-2222-3333-4444-555555555555"
		}
		
		app_json = jsonencode(
			%s
		)
	}`, name, testAppBuilderAppInlineWithOptionalFieldsAppJSON())
}

func testAppBuilderAppInlineWithOptionalFieldsAppJSON() string {
	return `
    {
		"queries" : [
			{
				"id": "22a65cea-b96f-44fa-a1e2-008885d7bd2e",
				"type": "action",
				"name": "listTeams0",
				"events": [],
				"properties": {
					"spec": {
						"fqn": "com.datadoghq.dd.teams.listTeams",
						"connectionId": "11111111-1111-1111-1111-111111111111",
						"inputs": {}
					}
				}
			}
		],
		"id" : "11111111-2222-3333-4444-555555555555",
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
		"description" : "This description will be overridden",
		"favorite": false,
		"name" : "This name will be overridden",
		"rootInstanceName" : "This rootInstanceName will be overridden",
		"selfService": false,
		"tags" : [],
		"connections": [
			{
				"id": "11111111-1111-1111-1111-111111111111",
				"name": "A connection that will be overridden"
			}
		],
		"deployment" : {
			"id" : "11111111-1111-1111-1111-111111111111",
			"app_version_id" : "00000000-0000-0000-0000-000000000000"
		}
	}
	`
}

func testAccCheckDatadogAppBuilderAppDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources[resourceName]
		id, err := uuid.Parse(resource.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing id as uuid: %s", err)
		}
		_, httpRes, err := apiInstances.GetAppBuilderApiV2().GetApp(auth, id)
		if err != nil {
			if httpRes.StatusCode == 404 {
				return nil
			}
			return err
		}

		return fmt.Errorf("failed app destroy check")
	}
}
