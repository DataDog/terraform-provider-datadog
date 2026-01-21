package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

type complianceCustomFrameworkModel struct {
	Handle       types.String
	Version      types.String
	Name         types.String
	IconURL      types.String
	Requirements []complianceCustomFrameworkRequirementsModel
}

type complianceCustomFrameworkRequirementsModel struct {
	Name     types.String
	Controls []complianceCustomFrameworkControlsModel
}

type complianceCustomFrameworkControlsModel struct {
	Name    types.String
	RulesID []string
}

func buildCreateFrameworkRequest(model complianceCustomFrameworkModel) *datadogV2.CreateCustomFrameworkRequest {
	req := datadogV2.NewCreateCustomFrameworkRequestWithDefaults()
	iconURL := model.IconURL.ValueString()
	req.SetData(datadogV2.CustomFrameworkData{
		Type: "custom_framework",
		Attributes: datadogV2.CustomFrameworkDataAttributes{
			Handle:  model.Handle.ValueString(),
			Name:    model.Name.ValueString(),
			IconUrl: &iconURL,
			Version: model.Version.ValueString(),
			Requirements: func() []datadogV2.CustomFrameworkRequirement {
				requirements := make([]datadogV2.CustomFrameworkRequirement, len(model.Requirements))
				for i, req := range model.Requirements {
					controls := make([]datadogV2.CustomFrameworkControl, len(req.Controls))
					for j, ctrl := range req.Controls {
						controls[j] = *datadogV2.NewCustomFrameworkControl(ctrl.Name.ValueString(), ctrl.RulesID)
					}
					requirements[i] = *datadogV2.NewCustomFrameworkRequirement(controls, req.Name.ValueString())
				}
				return requirements
			}(),
		},
	})
	return req
}

func TestCustomFramework_CreateBasic(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "icon_url", "test-url"),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9")),
			},
			{
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "security-requirement"),
					resource.TestCheckResourceAttr(path, "requirements.1.name", "compliance-requirement"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "security-control"),
					resource.TestCheckResourceAttr(path, "requirements.1.controls.0.name", "compliance-control"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.0.controls.0.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.0.controls.0.rules_id.*", "def-000-cea"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.1.controls.0.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.1.controls.0.rules_id.*", "def-000-cea"),
				),
			},
			{
				Config: testAccCheckDatadogCreateFrameworkWithoutOptionalFields(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFrameworkWithoutOptionalFields(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "security-requirement"),
					resource.TestCheckResourceAttr(path, "requirements.1.name", "compliance-requirement"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "security-control"),
					resource.TestCheckResourceAttr(path, "requirements.1.controls.0.name", "compliance-control"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.0.controls.0.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.0.controls.0.rules_id.*", "def-000-cea"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.1.controls.0.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.1.controls.0.rules_id.*", "def-000-cea"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "security-requirement"),
					resource.TestCheckResourceAttr(path, "requirements.1.name", "requirement"),
					resource.TestCheckResourceAttr(path, "requirements.2.name", "requirement-2"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "security-control"),
					resource.TestCheckResourceAttr(path, "requirements.1.controls.0.name", "control"),
					resource.TestCheckResourceAttr(path, "requirements.2.controls.0.name", "control-2"),
					resource.TestCheckResourceAttr(path, "requirements.2.controls.1.name", "control-3"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.0.controls.0.rules_id.*", "def-000-cea"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.1.controls.0.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.2.controls.0.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.2.controls.1.rules_id.*", "def-000-cea"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.2.controls.1.rules_id.*", "def-000-be9"),
				),
			},
		},
	})
}

// these handle and version combinations would result in the same state ID however
// terraform would still be able to tell the framework states apart since they have unique resource names (which is required by terraform)
func TestCustomFramework_SameFrameworkID(t *testing.T) {
	handle := "new-terraform"
	version := "framework-1.0"

	handle2 := "new"
	version2 := "terraform-framework-1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	path2 := "datadog_compliance_custom_framework.sample_framework"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
			},
			{
				Config: testAccCheckDatadogCreateSecondFramework(version2, handle2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path2, "handle", handle2),
					resource.TestCheckResourceAttr(path2, "version", version2),
				),
			},
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
				ExpectNonEmptyPlan: false,
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
		ProtoV6ProviderFactories: accProviders,
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
				Config:      testAccCheckDatadogDuplicateRequirementName(version, handle),
				ExpectError: regexp.MustCompile(".*Each Requirement must have a unique name.*"),
			},
			{
				Config:      testAccCheckDatadogDuplicateControlName(version, handle),
				ExpectError: regexp.MustCompile(".*Each Control must have a unique name under the same requirement.*"),
			},
			{
				Config:      testAccCheckDatadogDuplicateRequirement(version, handle),
				ExpectError: regexp.MustCompile(".*Each Requirement must have a unique name.*"),
			},
			{
				Config:      testAccCheckDatadogDuplicateControl(version, handle),
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
			{
				Config:      testAccCheckDatadogCreateFrameworkWithNoRulesId(version, handle),
				ExpectError: regexp.MustCompile("Invalid Attribute Value"),
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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
				),
			},
			{
				PreConfig: func() {
					// Delete the framework directly via API
					_, _, err := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().DeleteCustomFramework(providers.frameworkProvider.Auth, handle, version)
					if err != nil {
						t.Fatalf("Failed to delete framework: %v", err)
					}
				},
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
				),
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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
				),
			},
			{
				PreConfig: func() {
					// Delete the framework directly via API
					_, _, err := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().DeleteCustomFramework(providers.frameworkProvider.Auth, handle, version)
					if err != nil {
						t.Fatalf("Failed to delete framework: %v", err)
					}
				},
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
				),
			},
		},
	})
}

func TestCustomFramework_UpdateIfFrameworkExists(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create the framework directly via API
					_, _, err := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().CreateCustomFramework(providers.frameworkProvider.Auth, *buildCreateFrameworkRequest(complianceCustomFrameworkModel{
						Handle:  types.StringValue(handle),
						Version: types.StringValue(version),
						Name:    types.StringValue("existing-framework"),
						Requirements: []complianceCustomFrameworkRequirementsModel{
							{
								Name: types.StringValue("existing-requirement"),
								Controls: []complianceCustomFrameworkControlsModel{
									{
										Name:    types.StringValue("existing-control"),
										RulesID: []string{"def-000-be9"},
									},
								},
							},
						},
					}))
					if err != nil {
						t.Fatalf("Failed to create framework: %v", err)
					}
				},
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
				),
			},
		},
	})
}

func TestCustomFramework_RecreateOnImmutableFields(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"
	newHandle := "terraform-handle-new"
	newVersion := "2.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
			},
			{
				Config: testAccCheckDatadogCreateFramework(newVersion, newHandle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", newHandle),
					resource.TestCheckResourceAttr(path, "version", newVersion),
					// Verify old resource is deleted
					func(s *terraform.State) error {
						_, httpResp, err := providers.frameworkProvider.DatadogApiInstances.GetSecurityMonitoringApiV2().GetCustomFramework(providers.frameworkProvider.Auth, handle, version)
						if err == nil || (httpResp != nil && httpResp.StatusCode != 400) {
							return fmt.Errorf("old framework with handle %s and version %s still exists", handle, version)
						}
						return nil
					},
				),
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
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDuplicateRulesId(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.#", "1"),
				),
			},
		},
	})
}

func TestCustomFramework_DuplicateHandle(t *testing.T) {
	handle := "terraform-handle"
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_compliance_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
			},
			{
				Config:      testAccCheckDatadogCreateSecondFramework("different-version", handle),
				ExpectError: regexp.MustCompile("Framework with same handle already exists"),
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

func testAccCheckDatadogCreateSecondFramework(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_compliance_custom_framework" "sample_framework" {
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

func testAccCheckDatadogCreateFrameworkWithNoRulesId(version string, handle string) string {
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
					rules_id = []
				}
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

func testAccCheckDatadogDuplicateRequirementName(version string, handle string) string {
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

func testAccCheckDatadogDuplicateRequirement(version string, handle string) string {
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
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogDuplicateControlName(version string, handle string) string {
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

func testAccCheckDatadogDuplicateControl(version string, handle string) string {
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
					rules_id = ["def-000-be9"]
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
