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
		panic(fmt.Errorf("error creating temporary file: %s", err))
	}

	// delete temp file after test
	defer os.Remove(file.Name())

	// write app json to file
	_, err = file.WriteString(testComplexAppBuilderAppJSON(appName))
	if err != nil {
		panic(fmt.Errorf("error writing to file: %s", err))
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
					resource.TestCheckResourceAttr(resourceName, "app_json", testComplexAppBuilderAppJSON(appName)),
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
		panic(fmt.Errorf("error creating temporary file: %s", err))
	}

	// delete temp file after test
	defer os.Remove(file.Name())

	// write app json to file
	_, err = file.WriteString(testComplexAppBuilderAppJSON(appName))
	if err != nil {
		panic(fmt.Errorf("error writing to file: %s", err))
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
						return errors.New("failed to import resource into state")
					}

					id := s[0].Attributes["id"]
					if err := uuid.Validate(id); err != nil {
						return fmt.Errorf("imported resource ID %s is not a UUID: %s", id, err)
					}

					appJSON := s[0].Attributes["app_json"]
					var js json.RawMessage
					if err := json.Unmarshal([]byte(appJSON), &js); err != nil {
						return fmt.Errorf("imported resource attribute app_json not valid JSON: %s", err)
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
		return fmt.Errorf("error parsing ID as UUID: %s", err)
	}
	if _, _, err := apiInstances.GetAppBuilderApiV2().GetApp(ctx, id); err != nil {
		return fmt.Errorf("error getting app: %s", err)
	}
	return nil
}

func testComplexAppBuilderAppJSON(name string) string {
	return fmt.Sprintf(`
	{
		"queries": [
			{
				"id": "2480671a-04a8-4bd3-91ac-82f723ae145e",
				"name": "getServiceNames",
				"type": "dataTransform",
				"properties": {
					"outputs": "${(() => {// Write the body of your\n// transform function here.\n\n// Remember to return a value. \n\nreturn listServiceDefinitions0.outputs.data.map(output => output.attributes.schema['dd-service']);})()}"
				}
			},
			{
				"id": "4699ffaf-92ca-4e51-8a46-fac54ba1b6c2",
				"name": "getServiceLink",
				"type": "dataTransform",
				"properties": {
					"outputs": "${(() => {// Write the body of your\n// transform function here.\n\n// Remember to return a value. \n\nfunction getPagerDutyLink(ddService, data) {\n    // Iterate over the data array to find the matching dd-service\n    for (let i = 0; i < data.length; i++) {\n        if (data[i].attributes.schema['dd-service'] === ddService) {\n            // Return the corresponding PagerDuty link and exit the function\n            return data[i].attributes.schema.integrations.pagerduty;\n        }\n    }\n    // If no matching service is found, return null or an appropriate message\n    return null;\n}\n\nreturn getPagerDutyLink(region.value,listServiceDefinitions0.outputs.data)\n})()}"
				}
			},
			{
				"id": "e70f70d1-2c52-48a5-80a6-63fd50854f94",
				"name": "getServiceId",
				"type": "dataTransform",
				"properties": {
					"outputs": "${(() => {function parseExample(url) {\n    // Create a URL object from the given string\n    const urlObj = new URL(url);\n\n    // Get the pathname from the URL object\n    const pathname = urlObj.pathname;\n\n    // Split the pathname into parts using '/' as a delimiter\n    const parts = pathname.split('/');\n\n    // The last part of the array will be \"example0\"\n    const example = parts[parts.length - 1];\n\n    return example;\n}\n\n// Example usage\nreturn parseExample(getServiceLink.outputs[0].attributes.schema.integrations.pagerduty)})()}"
				}
			},
			{
				"id": "361a76b5-a868-4998-9bca-5b9dd3a8e287",
				"name": "createOrUpdateFile0",
				"type": "action",
				"properties": {
					"onlyTriggerManually": true,
					"spec": {
						"fqn": "com.datadoghq.github.createOrUpdateFile",
						"inputs": {
							"baseBranch": "main",
							"branchWithChanges": "new-s3-bucket-${bucketName.value}",
							"content": "resource \"aws_s3_bucket\" \"example\" {\n  bucket = \"${bucketName.value}\"\n\n  tags = {\n    Team = \"${owner.value}\"\n    Service = \"${service.value}\"\n    Environment = \"Dev\"\n  }\n}",
							"path": "generated-projects/My-Demo/${bucketName.value}.tf",
							"prDescription": "Auto-generating terraform for new s3 bucket for:\n${justification.value}",
							"prTitle": "new-s3-bucket-${bucketName.value}",
							"repository": "DataDog/app-builder-demo"
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
			},
			{
				"id": "0ca0ea43-3e3a-4071-b6aa-15b72075e0c6",
				"name": "listTeams0",
				"type": "action",
				"properties": {
					"onlyTriggerManually": false,
					"outputs": "${((outputs) => {// Use 'outputs' to reference the query's unformatted output.\n\n// TODO: Apply transformations to the raw query output\n\nreturn outputs.data.map(item => item.attributes.name);})(self.rawOutputs)}",
					"spec": {
						"fqn": "com.datadoghq.dd.teams.listTeams",
						"inputs": {},
						"connectionId": "5afce030-9bbf-437a-aa04-540fa4768635"
					}
				}
			},
			{
				"id": "8c203cf9-77c9-4167-a982-cb920714f3e7",
				"name": "listServiceDefinitions0",
				"type": "action",
				"properties": {
					"outputs": "${((outputs) => {// Use 'outputs' to reference the query's unformatted output.\n\n// TODO: Apply transformations to the raw query output\n\nreturn outputs.data.map(output => output.attributes.schema['dd-service']);})(self.rawOutputs)}",
					"spec": {
						"fqn": "com.datadoghq.dd.service_catalog.listServiceDefinitions",
						"inputs": {
							"limit": 1000
						},
						"connectionId": "5afce030-9bbf-437a-aa04-540fa4768635"
					}
				}
			}
		],
		"id": "e12cdda8-31a5-49f6-9307-dcd4d67cc911",
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
										"name": "grid2",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "gridCell8",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "text4",
																"properties": {
																	"content": "## ✅ You've successfully created a pull request",
																	"contentType": "markdown",
																	"isVisible": true,
																	"textAlign": "center",
																	"verticalAlign": "center"
																},
																"type": "text"
															}
														],
														"isVisible": "true",
														"layout": {
															"default": {
																"height": 5,
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
													"name": "gridCell10",
													"properties": {
														"children": [
															{
																"events": [
																	{
																		"name": "click",
																		"type": "openUrl",
																		"url": "${createOrUpdateFile0?.outputs.data.url}"
																	}
																],
																"name": "button1",
																"properties": {
																	"isBorderless": false,
																	"isDisabled": false,
																	"isLoading": false,
																	"isPrimary": true,
																	"isVisible": true,
																	"label": "Open",
																	"level": "default"
																},
																"type": "button"
															}
														],
														"isVisible": "true",
														"layout": {
															"default": {
																"height": 4,
																"width": 4,
																"x": 4,
																"y": 17
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
																"name": "text1",
																"properties": {
																	"content": "The request is assigned to the infrastructure team for approval. You will receive a slack notification when it is approved with the link to your new S3 bucket",
																	"contentType": "plain_text",
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
																"height": 8,
																"width": 12,
																"x": 0,
																"y": 6
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
								"isVisible": true,
								"size": "sm",
								"title": "Status"
							},
							"type": "modal"
						},
						{
							"events": [],
							"name": "gridCell9",
							"properties": {
								"children": [
									{
										"events": [
											{
												"name": "click",
												"queryName": "createOrUpdateFile0",
												"type": "triggerQuery"
											}
										],
										"name": "button0",
										"properties": {
											"isBorderless": false,
											"isDisabled": false,
											"isLoading": "${createOrUpdateFile0?.isLoading}",
											"isPrimary": true,
											"isVisible": true,
											"label": "Create PR",
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
										"y": 85
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
										"name": "form",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "grid3",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "gridCell2",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "region",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isMultiselect": "false",
																				"isVisible": true,
																				"label": "Region*",
																				"options": "[\n    {\n        \"label\": \"US East (N. Virginia)\",\n        \"value\": \"us-east-1\"\n    },\n    {\n        \"label\": \"US East (Ohio)\",\n        \"value\": \"us-east-2\"\n    },\n    {\n        \"label\": \"US West (N. California)\",\n        \"value\": \"us-west-1\"\n    },\n    {\n        \"label\": \"US West (Oregon)\",\n        \"value\": \"us-west-2\"\n    },\n    {\n        \"label\": \"Africa (Cape Town)\",\n        \"value\": \"af-south-1\"\n    },\n    {\n        \"label\": \"Asia Pacific (Hong Kong)\",\n        \"value\": \"ap-east-1\"\n    },\n    {\n        \"label\": \"Asia Pacific (Hyderabad)\",\n        \"value\": \"ap-south-2\"\n    },\n    {\n        \"label\": \"Asia Pacific (Jakarta)\",\n        \"value\": \"ap-southeast-3\"\n    },\n    {\n        \"label\": \"Asia Pacific (Mumbai)\",\n        \"value\": \"ap-south-1\"\n    },\n    {\n        \"label\": \"Asia Pacific (Osaka)\",\n        \"value\": \"ap-northeast-3\"\n    },\n    {\n        \"label\": \"Asia Pacific (Seoul)\",\n        \"value\": \"ap-northeast-2\"\n    },\n    {\n        \"label\": \"Asia Pacific (Singapore)\",\n        \"value\": \"ap-southeast-1\"\n    },\n    {\n        \"label\": \"Asia Pacific (Sydney)\",\n        \"value\": \"ap-southeast-2\"\n    },\n    {\n        \"label\": \"Asia Pacific (Tokyo)\",\n        \"value\": \"ap-northeast-1\"\n    },\n    {\n        \"label\": \"Canada (Central)\",\n        \"value\": \"ca-central-1\"\n    },\n    {\n        \"label\": \"China (Beijing)\",\n        \"value\": \"cn-north-1\"\n    },\n    {\n        \"label\": \"China (Ningxia)\",\n        \"value\": \"cn-northwest-1\"\n    },\n    {\n        \"label\": \"Europe (Frankfurt)\",\n        \"value\": \"eu-central-1\"\n    },\n    {\n        \"label\": \"Europe (Ireland)\",\n        \"value\": \"eu-west-1\"\n    },\n    {\n        \"label\": \"Europe (London)\",\n        \"value\": \"eu-west-2\"\n    },\n    {\n        \"label\": \"Europe (Milan)\",\n        \"value\": \"eu-south-1\"\n    },\n    {\n        \"label\": \"Europe (Paris)\",\n        \"value\": \"eu-west-3\"\n    },\n    {\n        \"label\": \"Europe (Spain)\",\n        \"value\": \"eu-south-2\"\n    },\n    {\n        \"label\": \"Europe (Stockholm)\",\n        \"value\": \"eu-north-1\"\n    },\n    {\n        \"label\": \"Middle East (Bahrain)\",\n        \"value\": \"me-south-1\"\n    },\n    {\n        \"label\": \"Middle East (UAE)\",\n        \"value\": \"me-central-1\"\n    },\n    {\n        \"label\": \"South America (São Paulo)\",\n        \"value\": \"sa-east-1\"\n    },\n    {\n        \"label\": \"AWS GovCloud (US-East)\",\n        \"value\": \"us-gov-east-1\"\n    },\n    {\n        \"label\": \"AWS GovCloud (US-West)\",\n        \"value\": \"us-gov-west-1\"\n    },\n    {\n        \"label\": \"Israel (Tel Aviv)\",\n        \"value\": \"il-central-1\"\n    }\n]",
																				"placeholder": "Select a region"
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
																			"y": 5
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell4_copy",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "bucketName",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isVisible": true,
																				"label": "Bucket Name*",
																				"placeholder": "Indicate the name of the bucket"
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
																			"y": 13
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
																			"name": "text3",
																			"properties": {
																				"content": "## Bucket Info",
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
																			"y": 0
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell15",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "owner",
																			"properties": {
																				"defaultValue": "alerting",
																				"isDisabled": false,
																				"isMultiselect": "false",
																				"isVisible": true,
																				"label": "Owner*",
																				"options": "${listTeams0?.outputs}",
																				"placeholder": "Select an owning team"
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
																			"y": 21
																		}
																	}
																},
																"type": "gridCell"
															},
															{
																"events": [],
																"name": "gridCell16",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "service",
																			"properties": {
																				"defaultValue": "staging",
																				"isDisabled": false,
																				"isMultiselect": "false",
																				"isVisible": true,
																				"label": "Service (optional)",
																				"options": "${listServiceDefinitions0?.outputs}",
																				"placeholder": "Associate a service"
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
																			"y": 29
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
																			"name": "justification",
																			"properties": {
																				"defaultValue": "",
																				"isDisabled": false,
																				"isVisible": true,
																				"label": "Justification*",
																				"placeholder": "Provide an overview of why you are creating the S3 bucket"
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
																			"y": 37
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
										"height": 56,
										"width": 6,
										"x": 3,
										"y": 28
									}
								}
							},
							"type": "gridCell"
						},
						{
							"events": [],
							"name": "modal0",
							"properties": {
								"children": [
									{
										"events": [],
										"name": "grid4",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "gridCell11",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "text2",
																"properties": {
																	"content": "# Success\nView your terraform [here]()",
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
																"width": 12,
																"x": 0,
																"y": 8
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
								"isVisible": true,
								"size": "sm",
								"title": "View file"
							},
							"type": "modal"
						},
						{
							"events": [],
							"name": "gridCell3",
							"properties": {
								"children": [
									{
										"events": [],
										"name": "text0",
										"properties": {
											"content": "# &nbsp; Create S3 Bucket with Terraform\n&nbsp; Hey ${global?.user.name} 👋,\nfill out the form to generate the terraform for a new S3 bucket. You will receive the output as a PR and receive a slack notification with a link to your new S3 bucket when approved.",
											"contentType": "markdown",
											"isVisible": true,
											"textAlign": "center",
											"verticalAlign": "center"
										},
										"type": "text"
									}
								],
								"isVisible": "true",
								"layout": {
									"default": {
										"height": 14,
										"width": 6,
										"x": 3,
										"y": 14
									}
								}
							},
							"type": "gridCell"
						},
						{
							"events": [],
							"name": "gridCell6",
							"properties": {
								"children": [
									{
										"events": [],
										"name": "container0",
										"properties": {
											"children": [
												{
													"events": [],
													"name": "grid1",
													"properties": {
														"children": [
															{
																"events": [],
																"name": "gridCell7",
																"properties": {
																	"children": [
																		{
																			"events": [],
																			"name": "text5",
																			"properties": {
																				"content": "![](https://app-builder-demo.s3.amazonaws.com/github/githublogo.png)",
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
																			"height": 9,
																			"width": 2,
																			"x": 5,
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
											"hasBackground": false,
											"isVisible": true
										},
										"type": "container"
									}
								],
								"isVisible": "true",
								"layout": {
									"default": {
										"height": 14,
										"width": 4,
										"x": 4,
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
		"description": "Fill out the form to generate the terraform for a new S3 bucket in Github. Created using the Datadog provider in Terraform",
		"favorite": false,
		"name": "%s",
		"rootInstanceName": "grid0",
		"selfService": false,
		"tags": []
	}`, name)
}
