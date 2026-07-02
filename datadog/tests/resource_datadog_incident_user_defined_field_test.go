package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIncidentUserDefinedField_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	fieldName := fmt.Sprintf("test_udf_basic_%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedFieldDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedFieldConfig(fieldName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedFieldExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "name", fieldName),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "type", "1"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "display_name", "Root Cause"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "category", "what_happened"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "valid_values.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "valid_values.0.value", "service_bug"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "attached_to", "incidents"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_field.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_field.foo", "incident_type"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_field.foo", "created"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentUserDefinedField_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	fieldName := fmt.Sprintf("test_udf_updated_%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedFieldDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedFieldConfig(fieldName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedFieldExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "display_name", "Root Cause"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "category", "what_happened"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentUserDefinedFieldConfigUpdated(fieldName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedFieldExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "display_name", "Updated Root Cause"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "category", "why_it_happened"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_field.foo", "valid_values.#", "1"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentUserDefinedField_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	fieldName := fmt.Sprintf("test_udf_import_%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedFieldDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedFieldConfig(fieldName),
			},
			{
				ResourceName:      "datadog_incident_user_defined_field.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentUserDefinedFieldConfig(fieldName string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for user-defined field"
}

resource "datadog_incident_user_defined_field" "foo" {
  name          = "%s"
  display_name  = "Root Cause"
  type          = 1
  category      = "what_happened"
  incident_type = datadog_incident_type.test.id

  valid_values {
    display_name = "Service Bug"
    value        = "service_bug"
    description  = "A bug in the service code."
  }

  valid_values {
    display_name = "Human Error"
    value        = "human_error"
  }
}`, fieldName, fieldName)
}

func testAccCheckDatadogIncidentUserDefinedFieldConfigUpdated(fieldName string) string {
	return fmt.Sprintf(`
resource "datadog_incident_type" "test" {
  name        = "%s-it"
  description = "Test incident type for user-defined field"
}

resource "datadog_incident_user_defined_field" "foo" {
  name          = "%s"
  display_name  = "Updated Root Cause"
  type          = 1
  category      = "why_it_happened"
  incident_type = datadog_incident_type.test.id

  valid_values {
    display_name = "Service Bug"
    value        = "service_bug"
    description  = "A bug in the service code."
  }
}`, fieldName, fieldName)
}

func testAccCheckDatadogIncidentUserDefinedFieldExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return incidentUserDefinedFieldExistsHelper(auth, s, apiInstances)
	}
}

func testAccCheckDatadogIncidentUserDefinedFieldDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return incidentUserDefinedFieldDestroyHelper(auth, s, apiInstances)
	}
}

func incidentUserDefinedFieldExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_user_defined_field" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetIncidentUserDefinedField(ctx, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving incident user-defined field")
		}
	}
	return nil
}

func incidentUserDefinedFieldDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetIncidentsApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_incident_user_defined_field" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := api.GetIncidentUserDefinedField(ctx, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving incident user-defined field")
		}
		return fmt.Errorf("incident user-defined field still exists")
	}
	return nil
}
