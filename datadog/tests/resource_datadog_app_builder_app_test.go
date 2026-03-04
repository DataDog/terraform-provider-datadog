package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogAppBuilderAppResource_Inline_WithOptionalFields(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_inline_optional"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAppBuilderAppInlineWithOptionalFieldsResourceConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "app_json", regexp.MustCompile(fmt.Sprintf(`"name":"%s"`, "This name will be overridden"))),

					resource.TestCheckResourceAttr(resourceName, "action_query_names_to_connection_ids.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "action_query_names_to_connection_ids.listTeams0", "11111111-2222-3333-4444-555555555555"),

					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttr(resourceName, "description", "Created using the Datadog provider in Terraform."),
					resource.TestCheckResourceAttr(resourceName, "root_instance_name", "grid0"),
					resource.TestCheckResourceAttr(resourceName, "published", "true"),
				),
			},
		},
	})
}

func TestAccDatadogAppBuilderAppResource_Inline_WithOptionalFields_Import(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_inline_optional"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAppBuilderAppInlineWithOptionalFieldsResourceConfig(appName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_json"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}

					var importedJSON map[string]any
					if err := json.Unmarshal([]byte(states[0].Attributes["app_json"]), &importedJSON); err != nil {
						return fmt.Errorf("error unmarshaling imported JSON: %v", err)
					}

					// Verify specific fields in the imported JSON
					expectedFields := map[string]any{
						"name":             appName,
						"description":      "Created using the Datadog provider in Terraform.",
						"rootInstanceName": "grid0",
					}

					for field, expected := range expectedFields {
						if actual, ok := importedJSON[field]; !ok {
							return fmt.Errorf("field %q missing from imported JSON", field)
						} else if actual != expected {
							return fmt.Errorf("field %q mismatch, got %v, want %v", field, actual, expected)
						}
					}

					return nil
				},
			},
		},
	})
}

func TestAccDatadogAppBuilderAppResource_Escaped_Interpolated(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_escaped_interpolated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAppBuilderAppEscapedInterpolatedResourceConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "app_json", regexp.MustCompile(fmt.Sprintf(`"name":"%s"`, appName))),

					resource.TestCheckResourceAttr(resourceName, "action_query_names_to_connection_ids.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "action_query_names_to_connection_ids.listTeams0", "11111111-2222-3333-4444-555555555555"),

					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttr(resourceName, "description", "Created using the Datadog provider in Terraform. Variable interpolation."),
					resource.TestCheckResourceAttr(resourceName, "root_instance_name", "grid0"),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
				),
			},
		},
	})
}

func TestAccDatadogAppBuilderAppResource_Escaped_Interpolated_Import(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_escaped_interpolated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAppBuilderAppEscapedInterpolatedResourceConfig(appName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_json"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}

					var importedJSON map[string]any
					if err := json.Unmarshal([]byte(states[0].Attributes["app_json"]), &importedJSON); err != nil {
						return fmt.Errorf("error unmarshaling imported JSON: %v", err)
					}

					// Verify specific fields in the imported JSON
					expectedFields := map[string]any{
						"name":             appName,
						"description":      "Created using the Datadog provider in Terraform. Variable interpolation.",
						"rootInstanceName": "grid0",
					}

					for field, expected := range expectedFields {
						if actual, ok := importedJSON[field]; !ok {
							return fmt.Errorf("field %q missing from imported JSON", field)
						} else if actual != expected {
							return fmt.Errorf("field %q mismatch, got %v, want %v", field, actual, expected)
						}
					}

					return nil
				},
			},
		},
	})
}

func TestAccDatadogAppBuilderAppResource_FromFile(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_from_file"

	exampleFile, err := os.Open("../../examples/resources/datadog_app_builder_app/resource.json")
	if err != nil {
		panic(fmt.Errorf("error opening file: %s", err))
	}
	defer exampleFile.Close()
	exampleFileBytes, err := io.ReadAll(exampleFile)
	if err != nil {
		panic(fmt.Errorf("error reading file: %s", err))
	}
	exampleFileString := string(exampleFileBytes)

	// create temporary file to hold large app json & avoid ${} interpolation by TF
	file, err := os.CreateTemp("", "test_app_json.*.json")
	if err != nil {
		panic(fmt.Errorf("error creating temporary file: %s", err))
	}

	// delete temp file after test
	defer os.Remove(file.Name())

	// write app json to file
	_, err = file.Write(exampleFileBytes)
	if err != nil {
		panic(fmt.Errorf("error writing to file: %s", err))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testLoadFromFileAppBuilderAppResourceConfig(file.Name(), appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "app_json", exampleFileString),

					resource.TestCheckResourceAttr(resourceName, "action_query_names_to_connection_ids.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "action_query_names_to_connection_ids.listTeams0", "11111111-2222-3333-4444-555555555555"),

					resource.TestCheckResourceAttr(resourceName, "name", appName),
					resource.TestCheckResourceAttr(resourceName, "description", "Fill out the form to generate the terraform for a new S3 bucket in Github. Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr(resourceName, "root_instance_name", "grid0"),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
				),
			},
		},
	})
}

func TestAccDatadogAppBuilderAppResource_FromFile_Import(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_from_file"

	exampleFile, err := os.Open("../../examples/resources/datadog_app_builder_app/resource.json")
	if err != nil {
		panic(fmt.Errorf("error opening file: %s", err))
	}
	defer exampleFile.Close()
	exampleFileBytes, err := io.ReadAll(exampleFile)
	if err != nil {
		panic(fmt.Errorf("error reading file: %s", err))
	}

	// create temporary file to hold large app json & avoid ${} interpolation by TF
	file, err := os.CreateTemp("", "test_app_json.*.json")
	if err != nil {
		panic(fmt.Errorf("error creating temporary file: %s", err))
	}

	// delete temp file after test
	defer os.Remove(file.Name())

	// write app json to file
	_, err = file.Write(exampleFileBytes)
	if err != nil {
		panic(fmt.Errorf("error writing to file: %s", err))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testLoadFromFileAppBuilderAppResourceConfig(file.Name(), appName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"app_json", "action_query_names_to_connection_ids"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 state, got %d", len(states))
					}

					var importedJSON map[string]any
					if err := json.Unmarshal([]byte(states[0].Attributes["app_json"]), &importedJSON); err != nil {
						return fmt.Errorf("error unmarshaling imported JSON: %v", err)
					}

					// Verify specific fields in the imported JSON
					expectedFields := map[string]any{
						"name":             appName,
						"description":      "Fill out the form to generate the terraform for a new S3 bucket in Github. Created using the Datadog provider in Terraform",
						"rootInstanceName": "grid0",
					}

					for field, expected := range expectedFields {
						if actual, ok := importedJSON[field]; !ok {
							return fmt.Errorf("field %q missing from imported JSON", field)
						} else if actual != expected {
							return fmt.Errorf("field %q mismatch, got %v, want %v", field, actual, expected)
						}
					}

					return nil
				},
			},
		},
	})
}

func testAppBuilderAppEscapedInterpolatedResourceConfig(name string) string {
	return fmt.Sprintf(`
	variable "name" {
		type    = string
		default = "%s"
	}
	variable "description" {
		type    = string
		default = "Created using the Datadog provider in Terraform. Variable interpolation."
	}
	variable "root_instance_name" {
		type    = string
		default = "grid0"
	}
	variable "connection_id" {
		type    = string
		default = "11111111-2222-3333-4444-555555555555"
	}
	
	resource "datadog_app_builder_app" "test_app_escaped_interpolated" {
		app_json = jsonencode(
			%s
		)
	}`, name, testAppBuilderAppEscapedInterpolatedAppJSON())
}

func testAppBuilderAppEscapedInterpolatedAppJSON() string {
	return `
    {
		"queries" : [
			{
				"id" : "11111111-1111-1111-1111-111111111111",
				"name" : "listTeams0",
				"type" : "action",
				"properties" : {
					"onlyTriggerManually" : false,
					"outputs" : "$${((outputs) => {// Use 'outputs' to reference the query's unformatted output.\n\n// TODO: Apply transformations to the raw query output\n\nreturn outputs.data.map(item => item.attributes.name);})(self.rawOutputs)}",
					"spec" : {
						"fqn" : "com.datadoghq.dd.teams.listTeams",
						"inputs" : {},
						"connectionId" : "${var.connection_id}"
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
		"description" : "${var.description}",
		"favorite" : false,
		"name" : "${var.name}",
		"rootInstanceName" : "${var.root_instance_name}",
		"selfService" : false,
		"tags" : [],
		"connections" : [
			{
				"id" : "${var.connection_id}",
				"name" : "A connection that will be overridden"
			}
		],
		"deployment" : {
			"id" : "11111111-1111-1111-1111-111111111111",
			"app_version_id" : "00000000-0000-0000-0000-000000000000"
		}
    }
	`
}

func testLoadFromFileAppBuilderAppResourceConfig(fileName string, appName string) string {
	return fmt.Sprintf(`
	resource "datadog_app_builder_app" "test_app_from_file" {
		app_json = file("%s")
		action_query_names_to_connection_ids = {
			"listTeams0" = "11111111-2222-3333-4444-555555555555"
		}
		name = "%s"
	}`, fileName, appName)
}

func TestAccDatadogAppBuilderAppResource_PublishUnpublish(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app.test_app_publish"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			// Create app as unpublished
			{
				Config: testAppBuilderAppPublishConfig(appName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "app_json", regexp.MustCompile(fmt.Sprintf(`"name":"%s"`, appName))),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
				),
			},
			// Apply same config again - should handle already unpublished gracefully
			{
				Config: testAppBuilderAppPublishConfig(appName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
				),
			},
			// Update to publish the app
			{
				Config: testAppBuilderAppPublishConfig(appName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "published", "true"),
				),
			},
			// Apply same config again - should handle already published gracefully
			{
				Config: testAppBuilderAppPublishConfig(appName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "published", "true"),
				),
			},
			// Update to unpublish the app
			{
				Config: testAppBuilderAppPublishConfig(appName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
				),
			},
			// Apply unpublish config again to ensure idempotency
			{
				Config: testAppBuilderAppPublishConfig(appName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
				),
			},
			// Publish again to ensure toggle works
			{
				Config: testAppBuilderAppPublishConfig(appName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "published", "true"),
				),
			},
		},
	})
}

func testAppBuilderAppPublishConfig(name string, published bool) string {
	return fmt.Sprintf(`
	resource "datadog_app_builder_app" "test_app_publish" {
		app_json = jsonencode({
			"queries" : [
				{
					"id" : "11111111-1111-1111-1111-111111111111",
					"name" : "listTeams0",
					"type" : "action",
					"properties" : {
						"onlyTriggerManually" : false,
						"outputs" : "$${((outputs) => {return outputs.data.map(item => item.attributes.name);})(self.rawOutputs)}",
						"spec" : {
							"fqn" : "com.datadoghq.dd.teams.listTeams",
							"inputs" : {},
							"connectionId" : "11111111-2222-3333-4444-555555555555"
						}
					}
				}
			],
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
			"description" : "Test app for publish/unpublish",
			"favorite" : false,
			"name" : "%s",
			"rootInstanceName" : "grid0",
			"selfService" : false,
			"tags" : []
		})
		published = %t
	}`, name, published)
}

func testAccCheckDatadogAppBuilderAppExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogAppBuilderAppExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogAppBuilderAppExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	idString := s.RootModule().Resources[name].Primary.ID
	id, err := uuid.Parse(idString)
	if err != nil {
		return fmt.Errorf("error parsing ID as UUID: %s", err)
	}
	if _, _, err := apiInstances.GetAppBuilderApiV2().GetApp(ctx, id); err != nil {
		return fmt.Errorf("error getting app: %s", err)
	}
	return nil
}
