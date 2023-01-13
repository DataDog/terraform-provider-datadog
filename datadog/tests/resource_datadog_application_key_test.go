package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogApplicationKey_Update(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	applicationKeyName := uniqueEntityName(ctx, t)
	applicationKeyNameUpdate := applicationKeyName + "-2"
	applicationKeyScopes := []string{
		"dashboards_read",
		"dashboards_write",
	}
	resourceName := "datadog_application_key.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApplicationKeyConfigRequired(applicationKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(accProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationKeyName),
					testAccCheckDatadogApplicationKeyValueMatches(accProvider, resourceName),
				),
			},
			{
				Config: testAccCheckDatadogApplicationKeyConfigRequired(applicationKeyNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(accProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationKeyNameUpdate),
					testAccCheckDatadogApplicationKeyValueMatches(accProvider, resourceName),
				),
			},
			{
				Config: testAccCheckDatadogApplicationKeyConfigScopesRequired(applicationKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(accProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scopes.#", fmt.Sprint(len(applicationKeyScopes))),
					testAccCheckDatadogApplicationKeyValueMatches(accProvider, resourceName),
				),
			},
		},
	})
}

func TestDatadogApplicationKey_import(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	resourceName := "datadog_application_key.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApplicationKeyConfigRequired(applicationKeyName),
			},
			{
				Config: testAccCheckDatadogApplicationKeyConfigScopesRequired(applicationKeyName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogApplicationKeyConfigRequired(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_application_key" "foo" {
  name = "%s"
}`, uniq)
}

func testAccCheckDatadogApplicationKeyConfigScopesRequired(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_application_key" "foo" {
  name = "%s"
  scopes = ["dashboards_read", "dashboards_write"]
}`, uniq)
}

func testAccCheckDatadogApplicationKeyExists(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogApplicationKeyExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogApplicationKeyExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := apiInstances.GetKeyManagementApiV2().GetCurrentUserApplicationKey(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving application key %s", err)
	}
	return nil
}

func testAccCheckDatadogApplicationKeyValueMatches(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogApplicationKeyValueMatches(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogApplicationKeyValueMatches(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	primaryResource := s.RootModule().Resources[name].Primary
	id := primaryResource.ID
	expectedKey := primaryResource.Attributes["key"]
	resp, _, err := apiInstances.GetKeyManagementApiV2().GetCurrentUserApplicationKey(ctx, id)
	if err != nil {
		return fmt.Errorf("received an error retrieving application key %s", err)
	}
	actualKey := resp.Data.Attributes.GetKey()
	if expectedKey != actualKey {
		return fmt.Errorf("application key value does not match")
	}
	return nil
}

func testAccCheckDatadogApplicationKeyDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogApplicationKeyDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogApplicationKeyDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_application_key" {
			continue
		}

		id := r.Primary.ID
		_, httpResponse, err := apiInstances.GetKeyManagementApiV2().GetCurrentUserApplicationKey(ctx, id)

		if err != nil {
			if httpResponse.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving application key %s", err)
		}

		return fmt.Errorf("application key still exists")
	}

	return nil
}
