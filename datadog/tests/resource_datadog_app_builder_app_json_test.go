package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogAppBuilderAppJSONResource_Inline_Basic(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app_json.test_app_inline_basic"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppJSONDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testInlineBasicAppBuilderAppJSONResourceConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppJSONExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "app_json"),
					resource.TestMatchResourceAttr(resourceName, "app_json", regexp.MustCompile(`\"name\":\"`+appName+`\"`)),
				),
			},
		},
	})
}

func TestAccDatadogAppBuilderAppJSONResource_Inline_Basic_Import(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app_json.test_app_inline_basic"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppJSONDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testInlineBasicAppBuilderAppJSONResourceConfig(appName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogAppBuilderAppJSONResource_FromFile_Complex(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app_json.test_app_from_file"

	// create temporary file to hold large app json & avoid ${} interpolation by TF
	file, err := os.CreateTemp("", "test_app_json.*.json")
	if err != nil {
		panic(fmt.Errorf("Error creating temporary file: %s", err))
	}

	// delete temp file after test
	defer os.Remove(file.Name())

	// write app json to file
	_, err = file.WriteString(testComplexAppBuilderAppJSONJSON(appName))
	if err != nil {
		panic(fmt.Errorf("Error writing to file: %s", err))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppJSONDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testLoadFromFileComplexAppBuilderAppJSONResourceConfig(file.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppBuilderAppJSONExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "app_json"),
					resource.TestCheckResourceAttr(resourceName, "app_json", testComplexAppBuilderAppJSONJSON(appName)),
				),
			},
		},
	})
}

func TestAccDatadogAppBuilderAppJSONResource_FromFile_Complex_Import(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	appName := uniqueEntityName(ctx, t)
	resourceName := "datadog_app_builder_app_json.test_app_from_file"

	// create temporary file to hold large app json & avoid ${} interpolation by TF
	file, err := os.CreateTemp("", "test_app_json.*.json")
	if err != nil {
		panic(fmt.Errorf("Error creating temporary file: %s", err))
	}

	// delete temp file after test
	defer os.Remove(file.Name())

	// write app json to file
	_, err = file.WriteString(testComplexAppBuilderAppJSONJSON(appName))
	if err != nil {
		panic(fmt.Errorf("Error writing to file: %s", err))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppBuilderAppJSONDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testLoadFromFileComplexAppBuilderAppJSONResourceConfig(file.Name()),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: func(s []*terraform.InstanceState) (err error) {
					// make sure new state has valid json and expected values
					if len(s) != 1 {
						return errors.New("Failed to import resource into state")
					}

					id := s[0].Attributes["id"]
					if err := uuid.Validate(id); err != nil {
						return fmt.Errorf("Imported resource id %s is not a uuid: %s", id, err)
					}

					appJSON := s[0].Attributes["app_json"]
					var js json.RawMessage
					if err := json.Unmarshal([]byte(appJSON), &js); err != nil {
						return fmt.Errorf("Imported resource attribute app_json not valid json: %s", err)
					}

					return
				},
			},
		},
	})
}

func testLoadFromFileComplexAppBuilderAppJSONResourceConfig(fileName string) string {
	return fmt.Sprintf(`
	resource "datadog_app_builder_app_json" "test_app_from_file" {
		app_json = file("%s")
	}`, fileName)
}

func testAccCheckDatadogAppBuilderAppJSONExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogAppBuilderAppJSONExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogAppBuilderAppJSONExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	idString := s.RootModule().Resources[name].Primary.ID
	id, err := uuid.Parse(idString)
	if err != nil {
		return fmt.Errorf("Error parsing id as uuid: %s", err)
	}
	if _, _, err := apiInstances.GetAppBuilderApiV2().GetApp(ctx, id); err != nil {
		return fmt.Errorf("Error getting app: %s", err)
	}
	return nil
}

func testComplexAppBuilderAppJSONJSON(name string) string {
	return fmt.Sprintf(`
	{
		"queries": [
			{
				"id": "2df060be-0743-45b1-a952-07c70eda8c98",
				"name": "createIssue0",
				"type": "action",
				"properties": {
					"onlyTriggerManually": true,
					"requiresConfirmation": true,
					"showToastOnError": true,
					"spec": {
						"fqn": "com.datadoghq.jira.create_issue",
						"inputs": {
							"description": "${textArea0.value}",
							"issueTypeName": "${Issue_type.value}",
							"priorityName": "${priority.value}",
							"projectKey": "${project.value}",
							"summary": "${title.value}",
							"accountId": "00000000-0000-0000-0000-000000000000"
						}
					}
				},
				"events": [
					{
						"modalName": "confirmation",
						"name": "executionFinished",
						"type": "openModal"
					}
				]
			}
		],
		"components": [
			{
				"events": [],
				"name": "grid0",
				"properties": {
					"backgroundColor": "default",
					"children": [
						{
							"events": [],
							"name": "confirmation",
							"properties": {
								"children": [
									{
										"events": [],
										"name": "grid1",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "gridCell3",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "Backlog",
																"properties": {
																	"content": "## Here is the link to your [ticket](${createIssue0.outputs.issueUrl})\n\n### and you can access the backlog [here](https://example.atlassian.net)",
																	"contentType": "markdown",
																	"isVisible": "true",
																	"textAlign": "center",
																	"verticalAlign": "center"
																},
																"type": "text"
															}
														],
														"isVisible": "true",
														"layout": {
															"default": {
																"height": 31,
																"width": 12,
																"x": 0,
																"y": 0
															}
														}
													},
													"type": "gridCell"
												}
											]
										},
										"type": "grid"
									}
								],
								"isVisible": "true",
								"size": "sm",
								"title": "Ticket submitted!"
							},
							"type": "modal"
						},
						{
							"events": [],
							"name": "gridCell2_copy",
							"properties": {
								"children": [
									{
										"events": [
											{
												"name": "click",
												"queryName": "createIssue0",
												"type": "triggerQuery"
											}
										],
										"name": "button0_copy",
										"properties": {
											"isBorderless": false,
											"isDisabled": false,
											"isLoading": "${createIssue0.isLoading}",
											"isPrimary": true,
											"isVisible": "true",
											"label": "Create Ticket",
											"level": "default"
										},
										"type": "button"
									}
								],
								"isVisible": "true",
								"layout": {
									"default": {
										"height": 4,
										"width": 6,
										"x": 3,
										"y": 81
									}
								}
							},
							"type": "gridCell"
						},
						{
							"events": [],
							"name": "gridCell1",
							"properties": {
								"children": [
									{
										"events": [],
										"name": "container1",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "grid3",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "gridCell1_copy1",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "project",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isMultiselect": false,
																				"isVisible": "true",
																				"label": "Project",
																				"options": "${[\n    {\n        \"label\": \"App Team\",\n        \"value\": \"APPS\"\n    },\n    {\n        \"label\": \"Platform Team\",\n        \"value\": \"PLATFORM\"\n    }\n]}",
																				"placeholder": "App Team"
																			},
																			"type": "select"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 7,
																			"width": 12,
																			"x": 0,
																			"y": 0
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell1_copy2",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "Issue_type",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isMultiselect": "false",
																				"isVisible": "true",
																				"label": "Type",
																				"options": "${[\n    {\n        \"label\": \"Bug\",\n        \"value\": \"Bug\"\n    },\n    {\n        \"label\": \"Task\",\n        \"value\": \"Task\"\n    },\n    {\n        \"label\": \"Sub-task\",\n        \"value\": \"Sub-task\"\n    },\n       {\n        \"label\": \"Story\",\n        \"value\": \"Story\"\n    },\n         {\n        \"label\": \"Epic\",\n        \"value\": \"Epic\"\n    },\n]}",
																				"placeholder": "Bug"
																			},
																			"type": "select"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 7,
																			"width": 12,
																			"x": 0,
																			"y": 8
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell0_copy2",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "title",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isVisible": "true",
																				"label": "Title",
																				"placeholder": "Sample title"
																			},
																			"type": "textInput"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 7,
																			"width": 12,
																			"x": 0,
																			"y": 16
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell1_copy3",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "priority",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isMultiselect": "false",
																				"isVisible": "true",
																				"label": "Priority",
																				"options": "[\n    {\n        \"label\": \"Critical\",\n        \"value\": \"Critical\"\n    },\n    {\n        \"label\": \"High\",\n        \"value\": \"High\"\n    },\n    {\n        \"label\": \"Medium\",\n        \"value\": \"Medium\"\n    },\n    {\n        \"label\": \"Low\",\n        \"value\": \"Low\"\n    }\n]",
																				"placeholder": "Low"
																			},
																			"type": "select"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 7,
																			"width": 12,
																			"x": 0,
																			"y": 24
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell4",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "textArea0",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isVisible": true,
																				"label": "Description",
																				"placeholder": "Sample description"
																			},
																			"type": "textArea"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 13,
																			"width": 12,
																			"x": 0,
																			"y": 35
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell5",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "text1",
																			"properties": {
																				"content": "----",
																				"contentType": "markdown",
																				"isVisible": true,
																				"textAlign": "left",
																				"verticalAlign": "top"
																			},
																			"type": "text"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 4,
																			"width": 12,
																			"x": 0,
																			"y": 31
																		}
																	}
																},
																"type": "gridCell"
															}
														]
													},
													"type": "grid"
												}
											],
											"hasBackground": true,
											"isVisible": true
										},
										"type": "container"
									}
								],
								"isVisible": "true",
								"layout": {
									"default": {
										"height": 53,
										"width": 6,
										"x": 3,
										"y": 27
									}
								}
							},
							"type": "gridCell"
						},
						{
							"events": [],
							"name": "gridCell2",
							"properties": {
								"children": [
									{
										"events": [],
										"name": "container0",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "grid2",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "gridCell0",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "text0",
																			"properties": {
																				"content": "![](https://app-builder-demo.s3.amazonaws.com/jira/jira-software.png)",
																				"contentType": "markdown",
																				"isVisible": true,
																				"textAlign": "center",
																				"verticalAlign": "top"
																			},
																			"type": "text"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 14,
																			"width": 2,
																			"x": 5,
																			"y": 0
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell0",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "text0",
																			"properties": {
																				"content": "![](https://app-builder-demo.s3.amazonaws.com/jira/jira-software.png)",
																				"contentType": "markdown",
																				"isVisible": true,
																				"textAlign": "center",
																				"verticalAlign": "top"
																			},
																			"type": "text"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 14,
																			"width": 2,
																			"x": 5,
																			"y": 0
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell5_copy1",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "header",
																			"properties": {
																				"content": "## Jira Ticket Creator\nHey ${global.user.name} 👋, fill out the form to create a new jira ticket.",
																				"contentType": "markdown",
																				"isVisible": "true",
																				"textAlign": "center",
																				"verticalAlign": "center"
																			},
																			"type": "text"
																		}
																	],
																	"isVisible": "true",
																	"layout": {
																		"default": {
																			"height": 9,
																			"width": 12,
																			"x": 0,
																			"y": 14
																		}
																	}
																},
																"type": "gridCell"
															}
														]
													},
													"type": "grid"
												}
											],
											"hasBackground": false,
											"isVisible": true
										},
										"type": "container"
									}
								],
								"isVisible": "true",
								"layout": {
									"default": {
										"height": 27,
										"width": 6,
										"x": 3,
										"y": 0
									}
								}
							},
							"type": "gridCell"
						}
					]
				},
				"type": "grid"
			}
		],
		"description": "Create new tickets in Jira. Created using the Datadog provider in Terraform",
		"name": "%s",
		"favorite" : false,
		"rootInstanceName" : "grid0",
		"selfService" : false,
		"tags": []
	}`, name)
}
