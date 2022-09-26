package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogApplicationKeyDatasource_matchId(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, accProviders := testAccProviders(context.Background(), t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceApplicationKeyIdConfig(applicationKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(accProvider, "datadog_application_key.app_key_1"),
					resource.TestCheckResourceAttr("datadog_application_key.app_key_1", "name", fmt.Sprintf("%s 1", applicationKeyName)),
					resource.TestCheckResourceAttr("datadog_application_key.app_key_2", "name", fmt.Sprintf("%s 2", applicationKeyName)),
					resource.TestCheckResourceAttr("data.datadog_application_key.app_key", "name", fmt.Sprintf("%s 1", applicationKeyName)),
				),
			},
		},
	})
}

func TestAccDatadogApplicationKeyDatasource_matchName(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, accProviders := testAccProviders(context.Background(), t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceApplicationKeyNameConfig(applicationKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApplicationKeyExists(accProvider, "datadog_application_key.app_key_1"),
					resource.TestCheckResourceAttr("datadog_application_key.app_key_1", "name", fmt.Sprintf("%s 1", applicationKeyName)),
					resource.TestCheckResourceAttr("datadog_application_key.app_key_2", "name", fmt.Sprintf("%s 2", applicationKeyName)),
					resource.TestCheckResourceAttr("data.datadog_application_key.app_key", "name", fmt.Sprintf("%s 1", applicationKeyName)),
				),
			},
		},
	})
}

func TestAccDatadogApplicationKeyDatasource_matchIdError(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApplicationKeyIdOnlyConfig("11111111-2222-3333-4444-555555555555"),
				ExpectError: regexp.MustCompile("error getting application key"),
			},
		},
	})
}

func TestAccDatadogApplicationKeyDatasource_matchNameError(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, accProviders := testAccProviders(context.Background(), t)
	ctx, accProviders := testAccProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApplicationKeyNameOnlyConfig(applicationKeyName),
				ExpectError: regexp.MustCompile("your query returned no result, please try a less specific search criteria"),
			},
		},
	})
}

func TestAccDatadogApplicationKeyDatasource_missingParametersError(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogApplicationKeyDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApplicationKeyMissingParametersConfig(),
				ExpectError: regexp.MustCompile("missing id or name parameter"),
			},
		},
	})
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
	if v, ok := s.RootModule().Resources[name].Primary.Attributes["service_account"]; ok {
		if _, _, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountApplicationKey(ctx, v, id); err != nil {
			return fmt.Errorf("received an error retrieving service account application key %s", err)
		}
	} else {
		if _, _, err := apiInstances.GetKeyManagementApiV2().GetCurrentUserApplicationKey(ctx, id); err != nil {
			return fmt.Errorf("received an error retrieving application key %s", err)
		}
	}
	return nil
}

func testAccApplicationKeyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_application_key" "app_key_1" {
  name = "%s 1"
}

resource "datadog_application_key" "app_key_2" {
  name = "%s 2"
}`, uniq, uniq)
}

func testAccDatasourceApplicationKeyIdConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_application_key" "app_key" {
  depends_on = [
    datadog_application_key.app_key_1,
    datadog_application_key.app_key_2,
  ]
  id = datadog_application_key.app_key_1.id
}`, testAccApplicationKeyConfig(uniq))
}

func testAccDatasourceApplicationKeyIdOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_application_key" "app_key" {
  id = "%s"
}`, uniq)
}

func testAccDatasourceApplicationKeyNameConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_application_key" "app_key" {
  depends_on = [
    datadog_application_key.app_key_1,
    datadog_application_key.app_key_2,
  ]
  name = datadog_application_key.app_key_1.name
}`, testAccApplicationKeyConfig(uniq))
}

func testAccDatasourceApplicationKeyNameOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_application_key" "app_key" {
  name = "%s"
}`, uniq)
}

func testAccDatasourceApplicationKeyMissingParametersConfig() string {
	return fmt.Sprintf(`
data "datadog_application_key" "app_key" {
}`)
}
