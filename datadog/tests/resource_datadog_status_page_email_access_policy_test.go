package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogStatusPageEmailAccessPolicy_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-eap-basic-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageEmailAccessPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageEmailAccessPolicyConfig(pageName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageEmailAccessPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page.foo", "type", "private"),
					resource.TestCheckResourceAttrPair(
						"datadog_status_page_email_access_policy.allow", "page_id", "datadog_status_page.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.allow", "pattern", "@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.allow", "effect", "allow"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_email_access_policy.allow", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_email_access_policy.allow", "created_at"),
					resource.TestCheckResourceAttrPair(
						"datadog_status_page_email_access_policy.deny", "page_id", "datadog_status_page.foo", "id"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.deny", "pattern", "contractor@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.deny", "effect", "deny"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_email_access_policy.deny", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_status_page_email_access_policy.deny", "created_at"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageEmailAccessPolicy_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-eap-updated-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageEmailAccessPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageEmailAccessPolicySingleConfig(pageName, "@example.com", "allow"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageEmailAccessPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.foo", "pattern", "@example.com"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.foo", "effect", "allow"),
				),
			},
			{
				// The API has no update operation, so changing a policy replaces it.
				Config: testAccCheckDatadogStatusPageEmailAccessPolicySingleConfig(pageName, "@*.example.com", "deny"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("datadog_status_page_email_access_policy.foo", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogStatusPageEmailAccessPolicyExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.foo", "pattern", "@*.example.com"),
					resource.TestCheckResourceAttr(
						"datadog_status_page_email_access_policy.foo", "effect", "deny"),
				),
			},
		},
	})
}

func TestAccDatadogStatusPageEmailAccessPolicy_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	pageName := fmt.Sprintf("test-sp-eap-import-%d", clockFromContext(ctx).Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogStatusPageEmailAccessPolicyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogStatusPageEmailAccessPolicySingleConfig(pageName, "@example.com", "allow"),
			},
			{
				ResourceName:      "datadog_status_page_email_access_policy.foo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccStatusPageEmailAccessPolicyImportStateIDFunc("datadog_status_page_email_access_policy.foo"),
			},
		},
	})
}

func testAccStatusPageEmailAccessPolicyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}
		return fmt.Sprintf("%s:%s", r.Primary.Attributes["page_id"], r.Primary.ID), nil
	}
}

func testAccCheckDatadogStatusPageEmailAccessPolicyConfig(pageName string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "private"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"
}

resource "datadog_status_page_email_access_policy" "allow" {
  page_id = datadog_status_page.foo.id
  pattern = "@example.com"
  effect  = "allow"
}

resource "datadog_status_page_email_access_policy" "deny" {
  page_id = datadog_status_page.foo.id
  pattern = "contractor@example.com"
  effect  = "deny"
}`, pageName)
}

func testAccCheckDatadogStatusPageEmailAccessPolicySingleConfig(pageName, pattern, effect string) string {
	return fmt.Sprintf(`
resource "datadog_status_page" "foo" {
  name               = "%[1]s"
  type               = "private"
  domain_prefix      = "%[1]s"
  visualization_type = "bars_and_uptime_percentage"
}

resource "datadog_status_page_email_access_policy" "foo" {
  page_id = datadog_status_page.foo.id
  pattern = "%[2]s"
  effect  = "%[3]s"
}`, pageName, pattern, effect)
}

func testAccCheckDatadogStatusPageEmailAccessPolicyExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageEmailAccessPolicyExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogStatusPageEmailAccessPolicyDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := statusPageEmailAccessPolicyDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

// The API has no endpoint to get a single email access policy, so both helpers
// look the policy up in its status page's list of policies.
func statusPageEmailAccessPolicyExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_email_access_policy" {
			continue
		}

		resp, httpResp, err := api.ListEmailAccessPolicies(ctx, parseUUID(r.Primary.Attributes["page_id"]))
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error listing email access policies")
		}
		found := false
		for _, policy := range resp.GetData() {
			if policy.GetId() == parseUUID(r.Primary.ID) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("status page email access policy %s not found", r.Primary.ID)
		}
	}
	return nil
}

func statusPageEmailAccessPolicyDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	api := apiInstances.GetStatusPagesApiV2()
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_status_page_email_access_policy" {
			continue
		}

		resp, httpResp, err := api.ListEmailAccessPolicies(ctx, parseUUID(r.Primary.Attributes["page_id"]))
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			return utils.TranslateClientError(err, httpResp, "error listing email access policies")
		}
		for _, policy := range resp.GetData() {
			if policy.GetId() == parseUUID(r.Primary.ID) {
				return fmt.Errorf("status page email access policy still exists")
			}
		}
	}
	return nil
}
