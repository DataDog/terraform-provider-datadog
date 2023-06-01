package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogServiceAccountApplicationKey_Create(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	appKeyName := uniqueEntityName(ctx, t)
	svcActResName := "datadog_service_account.foo"
	svcActAppKeyResName := "datadog_service_account_application_key.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSvcAcctAppKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceAccountApplicationKeyConfigRequired(appKeyName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSvcAcctAppKeyExists(accProvider, svcActResName, svcActAppKeyResName),
					resource.TestCheckResourceAttr(svcActAppKeyResName, "name", appKeyName),
					resource.TestCheckResourceAttr(svcActResName, "name", "A Test Service Account"),
					resource.TestCheckResourceAttr(svcActResName, "email", username),
					testAccCheckDatadogSvcAcctAppKeyValueMatches(accProvider, svcActResName, svcActAppKeyResName),
				),
			},
		},
	})
}

func testAccCheckDatadogServiceAccountApplicationKeyConfigRequired(uniq, owner string) string {
	return fmt.Sprintf(`
resource "datadog_service_account" "foo" {
  name = "A Test Service Account"
  email = "%s"
}

resource "datadog_service_account_application_key" "foo" {
  name = "%s"
  service_account_id = datadog_service_account.foo.id
}`, owner, uniq)
}

func testAccCheckDatadogSvcAcctAppKeyExists(accProvider func() (*schema.Provider, error), acctRN, keyRN string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("===== testAccCheckDatadogSvcAcctAppKeyExists")
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogSvcAcctAppKeyExistsHelper(auth, s, apiInstances, acctRN, keyRN); err != nil {
			return err
		}
		return nil
	}
}

func datadogSvcAcctAppKeyExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, acctRN, keyRN string) error {
	svcAcctId := s.RootModule().Resources[acctRN].Primary.ID
	keyId := s.RootModule().Resources[keyRN].Primary.ID

	if _, _, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountApplicationKey(ctx, svcAcctId, keyId); err != nil {
		return fmt.Errorf("received an error retrieving service account application key %s", err)
	}

	return nil
}

func extractOwnerIdFromApplicationKeyRelationshipsHelper(relationships *datadogV2.ApplicationKeyRelationships) string {
	ownedBy := relationships.GetOwnedBy()
	ownerData := ownedBy.GetData()

	return ownerData.GetId()
}

func testAccCheckDatadogSvcAcctAppKeyValueMatches(accProvider func() (*schema.Provider, error), acctRN, keyRN string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogSvcAcctAppKeyValueMatches(auth, s, apiInstances, acctRN, keyRN); err != nil {
			return err
		}
		return nil
	}
}

func datadogSvcAcctAppKeyValueMatches(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, acctRN, keyRN string) error {
	svcAcctResource := s.RootModule().Resources[acctRN].Primary
	keyResource := s.RootModule().Resources[keyRN].Primary

	svcAcctId := svcAcctResource.ID
	keyId := keyResource.ID
	expectedKey := keyResource.Attributes["key"]

	svcAcctAppKeyResp, _, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountApplicationKey(ctx, svcAcctId, keyId)
	if err != nil {
		return fmt.Errorf("received an error retrieving service account application key %s", err)
	}

	appKey := svcAcctAppKeyResp.GetData()
	expectedKeyId := appKey.GetId()

	keyResp, _, err := apiInstances.GetKeyManagementApiV2().GetApplicationKey(ctx, expectedKeyId)
	if err != nil {
		return fmt.Errorf("received an error retrieving service account application key %s", err)
	}

	keyData := keyResp.GetData()
	actualRel := keyData.GetRelationships()
	actualOwnerId := extractOwnerIdFromApplicationKeyRelationshipsHelper(&actualRel)
	actualKeyId := keyData.GetId()

	if expectedKeyId != actualKeyId {
		return fmt.Errorf("service account application key Id does not match")
	}

	if svcAcctId != actualOwnerId {
		return fmt.Errorf("service account application key owners do not match")
	}

	if len(expectedKey) < 1 {
		return fmt.Errorf("service account application key not found in state")
	}
	return nil
}

func testAccCheckDatadogSvcAcctAppKeyDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogSvcAcctAppKeyDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogSvcAcctAppKeyDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_service_account_application_key" {
			continue
		}

		switch r.Type {
		case "datadog_service_account_application_key":
			id := r.Primary.ID
			_, httpResponse, err := apiInstances.GetKeyManagementApiV2().GetApplicationKey(ctx, id)

			if err != nil {
				if httpResponse.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving service account application key %s", err)
			}

			return fmt.Errorf("service account application key still exists")

		case "datadog_service_account":
			id := r.Primary.ID
			_, httpResponse, err := apiInstances.GetUsersApiV2().GetUser(ctx, id)

			if err != nil {
				if httpResponse.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving service account %s", err)
			}

			return fmt.Errorf("service account key still exists")
		default:
			continue
		}
	}

	return nil
}
