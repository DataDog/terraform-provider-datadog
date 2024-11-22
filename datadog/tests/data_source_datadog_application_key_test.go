package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// func TestAccDatadogApplicationKeyDatasource_matchId(t *testing.T) {
// 	if isRecording() || isReplaying() {
// 		t.Skip("This test doesn't support recording or replaying")
// 	}
// 	t.Parallel()
// 	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

// 	applicationKeyName := uniqueEntityName(ctx, t)
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV5ProviderFactories: accProviders,
// 		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDatasourceApplicationKeyIdConfig(applicationKeyName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckDatadogApplicationKeyExists(providers.frameworkProvider, "datadog_application_key.app_key_1"),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_1", "name", fmt.Sprintf("%s 1", applicationKeyName)),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_2", "name", fmt.Sprintf("%s 2", applicationKeyName)),
// 					resource.TestCheckResourceAttr("data.datadog_application_key.app_key", "name", fmt.Sprintf("%s 1", applicationKeyName)),
// 					resource.TestCheckResourceAttrSet("data.datadog_application_key.app_key", "id"),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccDatadogApplicationKeyDatasource_matchName(t *testing.T) {
// 	if isRecording() || isReplaying() {
// 		t.Skip("This test doesn't support recording or replaying")
// 	}
// 	t.Parallel()
// 	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
// 	applicationKeyName := uniqueEntityName(ctx, t)

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV5ProviderFactories: accProviders,
// 		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDatasourceApplicationKeyNameConfig(applicationKeyName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckDatadogApplicationKeyExists(providers.frameworkProvider, "datadog_application_key.app_key_1"),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_1", "name", fmt.Sprintf("%s 1", applicationKeyName)),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_2", "name", fmt.Sprintf("%s 2", applicationKeyName)),
// 					resource.TestCheckResourceAttr("data.datadog_application_key.app_key", "name", fmt.Sprintf("%s 1", applicationKeyName)),
// 					resource.TestCheckResourceAttrSet("data.datadog_application_key.app_key", "id"),
// 				),
// 			},
// 			{
// 				Config: testAccDatasourceAppKeyExactMatch(applicationKeyName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckDatadogApplicationKeyExists(providers.frameworkProvider, "datadog_application_key.app_key_1"),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_1", "name", applicationKeyName),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_2", "name", fmt.Sprintf("%s 2", applicationKeyName)),
// 					resource.TestCheckResourceAttr("data.datadog_application_key.app_key", "name", applicationKeyName),
// 					resource.TestCheckResourceAttrSet("data.datadog_application_key.app_key", "id"),
// 				),
// 			},
// 		},
// 	})
// }

// func TestAccDatadogApplicationKeyDatasource_exactMatchName(t *testing.T) {
// 	if isRecording() || isReplaying() {
// 		t.Skip("This test doesn't support recording or replaying")
// 	}
// 	t.Parallel()
// 	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
// 	applicationKeyName := uniqueEntityName(ctx, t)

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV5ProviderFactories: accProviders,
// 		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccDatasourceAppKeyExactMatch(applicationKeyName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckDatadogApplicationKeyExists(providers.frameworkProvider, "datadog_application_key.app_key_1"),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_1", "name", applicationKeyName),
// 					resource.TestCheckResourceAttr("datadog_application_key.app_key_2", "name", fmt.Sprintf("%s 2", applicationKeyName)),
// 					resource.TestCheckResourceAttr("data.datadog_application_key.app_key", "name", applicationKeyName),
// 					resource.TestCheckResourceAttrSet("data.datadog_application_key.app_key", "id"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccDatadogApplicationKeyDatasource_matchIdError(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
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
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	applicationKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
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
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApplicationKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApplicationKeyMissingParametersConfig(),
				ExpectError: regexp.MustCompile("missing id or name parameter"),
			},
		},
	})
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
	return `
data "datadog_application_key" "app_key" {
}`
}

func testAccDatasourceAppKeyExactMatch(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_application_key" "app_key_1" {
  name = "%s"
}

resource "datadog_application_key" "app_key_2" {
  name = "%s 2"
}
data "datadog_application_key" "app_key" {
  depends_on = [
    datadog_application_key.app_key_1,
    datadog_application_key.app_key_2,
  ]
  exact_match = true
  name = "%s"
}`, uniq, uniq, uniq)
}
