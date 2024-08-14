package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogApplicationKey_Update(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)
	applicationKeyNameUpdate := applicationKeyName + "-2"
	resourceName := "datadog_application_key.foo"
	var keyValue string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApplicationKeyConfigRequired(applicationKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationKeyName),
					resource.TestCheckResourceAttrSet(resourceName, "key"),
					func(s *terraform.State) error {
						resource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Resource not found: %s", resourceName)
						}
						// Store the key value for later comparison after update
						keyValue = resource.Primary.Attributes["key"]
						return nil
					},
				),
			},
			{
				Config: testAccCheckDatadogApplicationKeyConfigRequired(applicationKeyNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", applicationKeyNameUpdate),
					testAccCheckDatadogApplicationKeyNameMatches(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "key"),
					func(s *terraform.State) error {
						resource, ok := s.RootModule().Resources[resourceName]
						if !ok {
							return fmt.Errorf("Resource not found: %s", resourceName)
						}
						stateKeyValue := resource.Primary.Attributes["key"]
						if stateKeyValue != keyValue {
							return fmt.Errorf("application key value %s does not match expected value %s", stateKeyValue, keyValue)
						}
						return nil
					},
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
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogApplicationKeyConfigRequired(applicationKeyName),
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

func testAccCheckDatadogApplicationKeyExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

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

func testAccCheckDatadogApplicationKeyNameMatches(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		if err := datadogApplicationKeyNameMatches(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogApplicationKeyNameMatches(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	primaryResource := s.RootModule().Resources[name].Primary
	id := primaryResource.ID
	resp, _, err := apiInstances.GetKeyManagementApiV2().GetCurrentUserApplicationKey(ctx, id)
	if err != nil {
		return fmt.Errorf("received an error retrieving application key %s", err)
	}
	keyName := resp.Data.Attributes.GetName()
	stateKeyName := primaryResource.Attributes["name"]
	if keyName != stateKeyName {
		return fmt.Errorf("application key name %s in state does not match expected name %s from API", stateKeyName, keyName)
	}
	return nil
}

func testAccCheckDatadogApplicationKeyDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiIntances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogApplicationKeyDestroyHelper(auth, s, apiIntances); err != nil {
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
