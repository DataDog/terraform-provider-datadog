package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogIncidentUserDefinedRole_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := fmt.Sprintf("test-udr-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedRoleDestroy(providers.frameworkProvider, "datadog_incident_user_defined_role.foo"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedRoleConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedRoleExists(providers.frameworkProvider, "datadog_incident_user_defined_role.foo"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "description", "Test user-defined role"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "policy.is_single", "true"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_role.foo", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_incident_user_defined_role.foo", "incident_type_id"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentUserDefinedRole_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := fmt.Sprintf("test-udr-updated-%d", clockFromContext(ctx).Now().Unix())
	uniqUpdated := uniq + "-updated"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedRoleDestroy(providers.frameworkProvider, "datadog_incident_user_defined_role.foo"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedRoleConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedRoleExists(providers.frameworkProvider, "datadog_incident_user_defined_role.foo"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "name", uniq),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "policy.is_single", "true"),
				),
			},
			{
				Config: testAccCheckDatadogIncidentUserDefinedRoleConfigUpdated(uniqUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIncidentUserDefinedRoleExists(providers.frameworkProvider, "datadog_incident_user_defined_role.foo"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "name", uniqUpdated),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "description", "Updated user-defined role"),
					resource.TestCheckResourceAttr(
						"datadog_incident_user_defined_role.foo", "policy.is_single", "false"),
				),
			},
		},
	})
}

func TestAccDatadogIncidentUserDefinedRole_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := fmt.Sprintf("test-udr-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIncidentUserDefinedRoleDestroy(providers.frameworkProvider, "datadog_incident_user_defined_role.foo"),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIncidentUserDefinedRoleConfig(uniq),
			},
			{
				ResourceName:      "datadog_incident_user_defined_role.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogIncidentUserDefinedRoleConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_incident_type" "udr_parent" {
		name = "%s-type"
	}

	resource "datadog_incident_user_defined_role" "foo" {
		name             = "%s"
		description      = "Test user-defined role"
		incident_type_id = datadog_incident_type.udr_parent.id
		policy {
			is_single = true
		}
	}`, name, name)
}

func testAccCheckDatadogIncidentUserDefinedRoleConfigUpdated(name string) string {
	return fmt.Sprintf(`
	resource "datadog_incident_type" "udr_parent" {
		name = "%s-type"
	}

	resource "datadog_incident_user_defined_role" "foo" {
		name             = "%s"
		description      = "Updated user-defined role"
		incident_type_id = datadog_incident_type.udr_parent.id
		policy {
			is_single = false
		}
	}`, name, name)
}

func testAccCheckDatadogIncidentUserDefinedRoleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		id, err := uuid.Parse(s.RootModule().Resources[resourceName].Primary.ID)
		if err != nil {
			return err
		}
		if _, _, err := apiInstances.GetIncidentsApiV2().GetIncidentUserDefinedRole(auth, id); err != nil {
			return fmt.Errorf("received an error retrieving incident user-defined role %s", err)
		}
		return nil
	}
}

func testAccCheckDatadogIncidentUserDefinedRoleDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		res := s.RootModule().Resources[resourceName]
		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return err
		}
		_, httpResp, err := apiInstances.GetIncidentsApiV2().GetIncidentUserDefinedRole(auth, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				return nil
			}
			return err
		}

		return nil
	}
}
