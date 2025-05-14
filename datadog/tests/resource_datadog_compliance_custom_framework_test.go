package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestCustomFramework_Create(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "icon_url", "test-url"),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
				),
			},
			{
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "compliance-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "security-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "compliance-control",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "security-control",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
			},
			{
				Config: testAccCheckDatadogCreateFrameworkWithoutOptionalFields(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
				),
			},
		},
	})
}

func TestCustomFramework_CreateWithoutIconURL(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFrameworkWithoutOptionalFields(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
				),
			},
		},
	})
}

func TestCustomFramework_CreateAndUpdateMultipleRequirements(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "compliance-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "security-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "compliance-control",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "security-control",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "security-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement-2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control-2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control-3",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
			},
		},
	})
}

func TestCustomFramework_SameConfigNoUpdate(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"

	// First config with one order of requirements
	config1 := fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
			requirements  {
				name = "requirement2"
				controls {
					name = "control2.2"
					rules_id = ["def-000-be9"]
				}
				controls {
					name = "control2"
					rules_id = ["def-000-cea"]
				}
			}
			requirements  {
				name = "requirement3"
				controls {
					name = "control3"
					rules_id = ["def-000-be9", "def-000-cea"]
				}
			}
		}
	`, version, handle)

	// Second config with different order of requirements
	// test switching order of requirements, controls, and rules_id
	config2 := fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			icon_url      = "test url"
			requirements  {
				name = "requirement3"
				controls {
					name = "control3"
					rules_id = ["def-000-cea", "def-000-be9"]
				}
			}
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
			requirements  {
				name = "requirement2"
				controls {
					name = "control2"
					rules_id = ["def-000-cea"]
				}
				controls {
					name = "control2.2"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "icon_url", "test url"),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control3",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "icon_url", "test url"),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control3",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
				// This step should not trigger an update since only the order is different
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// There is no way to validate the duplicate rule IDs in the config before they are removed from the state in Terraform
// because the state model converts the rules_id into sets, the duplicate rule IDs are removed
// This test validates that the duplicate rule IDs are removed from the state and only one rule ID is present
func TestCustomFramework_DuplicateRuleIds(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDuplicateRulesId(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.#", "1"),
				),
			},
		},
	})
}

func TestCustomFramework_InvalidCreate(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogCreateInvalidFrameworkName(version, handle),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			{
				Config:      testAccCheckDatadogCreateFrameworkRuleIdsInvalid(version, handle),
				ExpectError: regexp.MustCompile("400 Bad Request"),
			},
			{
				Config:      testAccCheckDatadogCreateFrameworkNoRequirements(version, handle),
				ExpectError: regexp.MustCompile("Invalid Block"),
			},
			{
				Config:      testAccCheckDatadogCreateFrameworkWithNoControls(version, handle),
				ExpectError: regexp.MustCompile("Invalid Block"),
			},
			{
				Config:      testAccCheckDatadogCreateEmptyHandle(version),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			{
				Config:      testAccCheckDatadogCreateEmptyVersion(handle),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			{
				Config:      testAccCheckDatadogDuplicateRequirements(version, handle),
				ExpectError: regexp.MustCompile(".*Each Requirement must have a unique name.*"),
			},
			{
				Config:      testAccCheckDatadogDuplicateControls(version, handle),
				ExpectError: regexp.MustCompile(".*Each Control must have a unique name under the same requirement.*"),
			},
			{
				Config:      testAccCheckDatadogEmptyRequirementName(version, handle),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
			{
				Config:      testAccCheckDatadogEmptyControlName(version, handle),
				ExpectError: regexp.MustCompile("Invalid Attribute Value Length"),
			},
		},
	})
}

func TestCustomFramework_RecreateAfterAPIDelete(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				// First create the framework
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
			},
			{
				// Simulate framework being deleted in UI
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					// Delete the framework in the UI
					func(s *terraform.State) error {
						_, _, err := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().DeleteCustomFramework(providers.frameworkProvider.Auth, handle, version)
						return err
					},
				),
				// Expect a non-empty plan since the framework was deleted
				ExpectNonEmptyPlan: true,
			},
			{
				// Update the framework -
				// The read would be able to tell that the framework was deleted in UI so then it delete the local terraform state of the framework
				// this should trigger a create since it was deleted in UI
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					// Verify the framework was recreated with the new requirements
					resource.TestCheckResourceAttr(path, "requirements.#", "2"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestCustomFramework_DeleteAfterAPIDelete(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				// First create the framework in terraform
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
			},
			{
				// Simulate framework being deleted in UI
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					// Delete the framework in the UI
					func(s *terraform.State) error {
						_, _, err := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().DeleteCustomFramework(providers.frameworkProvider.Auth, handle, version)
						return err
					},
				),
				// Expect a non-empty plan since the framework was deleted
				ExpectNonEmptyPlan: true,
			},
			{
				// Try to remove the resource from terraform
				// Since the framework was deleted in UI, terraform should just remove it from state
				// The read in terraform will return a 400 error because the framework handle and version no longer exist
				// This will delete the framework from terraform
				Config: "# Empty config to simulate removing the resource",
				Check:  resource.ComposeTestCheckFunc(
				// No checks needed since we're removing the resource
				),
				// Expect no changes needed since the resource was already deleted
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestCustomFramework_CreateConflict(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				// First create the framework using the API directly
				Config: "# Empty config since we're creating via API",
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// Create framework using API
						api := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
						auth := providers.frameworkProvider.Auth

						// Check if framework exists and delete it if it does
						_, httpRes, err := api.GetCustomFramework(auth, handle, version)
						if err == nil && httpRes.StatusCode == 200 {
							// Framework exists, delete it
							_, _, err = api.DeleteCustomFramework(auth, handle, version)
							if err != nil {
								return fmt.Errorf("failed to delete existing framework: %v", err)
							}
						}

						// Create a basic framework that matches the config we'll try to create
						createRequest := datadogV2.NewCreateCustomFrameworkRequestWithDefaults()
						iconURL := "test url"
						createRequest.SetData(datadogV2.CustomFrameworkData{
							Type: "custom_framework",
							Attributes: datadogV2.CustomFrameworkDataAttributes{
								Handle:  handle,
								Name:    "new-framework-terraform",
								IconUrl: &iconURL,
								Version: version,
								Requirements: []datadogV2.CustomFrameworkRequirement{
									*datadogV2.NewCustomFrameworkRequirement(
										[]datadogV2.CustomFrameworkControl{
											*datadogV2.NewCustomFrameworkControl("control1", []string{"def-000-be9"}),
										},
										"requirement1",
									),
								},
							},
						})

						// Create the framework
						_, _, err = api.CreateCustomFramework(auth, *createRequest)
						if err != nil {
							return fmt.Errorf("failed to create framework: %v", err)
						}
						return nil
					},
				),
			},
			{
				// Try to create the same framework through Terraform
				Config:      testAccCheckDatadogCreateFramework(version, handle),
				ExpectError: regexp.MustCompile("409 Conflict"),
			},
		},
	})
}

func testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-name"
			icon_url      = "test url"
			requirements  {
				name = "security-requirement"
				controls {
					name = "security-control"
					rules_id = ["def-000-cea", "def-000-be9"]
				}
			}
			requirements  {
				name = "compliance-requirement"
				controls {
					name = "compliance-control"
					rules_id = ["def-000-be9", "def-000-cea"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogUpdateFrameworkWithMultipleRequirements(version, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-name"
			icon_url      = "test url"
			requirements  {
				name = "security-requirement"
				controls {
					name = "security-control"
					rules_id = ["def-000-cea"]
				}
			}
			requirements {
				name = "requirement"
				controls {
					name = "control"
					rules_id = ["def-000-be9"]
				}
			}
			requirements {
				name = "requirement-2"
				controls {
					name = "control-2"
					rules_id = ["def-000-be9"]
				}
				controls {
					name = "control-3"
					rules_id = ["def-000-cea", "def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogCreateFramework(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			icon_url      = "test-url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogCreateFrameworkWithoutOptionalFields(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogCreateFrameworkRuleIdsInvalid(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["invalid-rule-id"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogCreateFrameworkWithNoControls(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
			}
		}
	`, version, handle)
}

func testAccCheckDatadogCreateFrameworkNoRequirements(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			icon_url      = "test url"
		}
	`, version, handle)
}

func testAccCheckDatadogCreateInvalidFrameworkName(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "" 
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogEmptyRequirementName(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "name" 
			icon_url      = "test url"
			requirements  {
				name = ""
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogEmptyControlName(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "name" 
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = ""
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogCreateEmptyHandle(version string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = ""
			name          = "framework-name" 
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version)
}

func testAccCheckDatadogCreateEmptyVersion(handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = ""
			handle        = "%s"
			name          = "framework-name" 
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, handle)
}

func testAccCheckDatadogDuplicateRequirements(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "framework-name"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
			requirements  {
				name = "requirement1"
				controls {
					name = "control2"
					rules_id = ["def-000-cea"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogDuplicateControls(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "framework-name"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
			requirements  {
				name = "requirement2"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
				controls {
					name = "control1"
					rules_id = ["def-000-cea"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogDuplicateRulesId(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "framework-name"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9", "def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogFrameworkDestroy(ctx context.Context, accProvider *fwprovider.FrameworkProvider, resourceName string, version string, handle string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		resource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return nil
		}
		if resource.Primary == nil {
			return nil
		}
		handle := resource.Primary.Attributes["handle"]
		version := resource.Primary.Attributes["version"]
		_, httpRes, err := apiInstances.GetSecurityMonitoringApiV2().GetCustomFramework(ctx, handle, version)
		if err != nil {
			if httpRes != nil && httpRes.StatusCode == 400 {
				return nil
			}
			return err
		}

		return fmt.Errorf("framework destroy check failed")
	}
}
