package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogApiKeyDatasource_matchId(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceApiKeyIdConfig(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, "datadog_api_key.api_key_1"),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_1", "name", fmt.Sprintf("%s 1", apiKeyName)),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_2", "name", fmt.Sprintf("%s 2", apiKeyName)),
					resource.TestCheckResourceAttr("data.datadog_api_key.api_key", "name", fmt.Sprintf("%s 1", apiKeyName)),
				),
			},
		},
	})
}

func TestAccDatadogApiKeyDatasource_matchName(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceApiKeyNameConfig(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, "datadog_api_key.api_key_1"),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_1", "name", fmt.Sprintf("%s 1", apiKeyName)),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_2", "name", fmt.Sprintf("%s 2", apiKeyName)),
					resource.TestCheckResourceAttr("data.datadog_api_key.api_key", "name", fmt.Sprintf("%s 1", apiKeyName)),
				),
			},
		},
	})
}

func TestAccDatadogApiKeyDatasource_exactMatchName(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceApiKeyExactMatch(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, "datadog_api_key.api_key_1"),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_1", "name", apiKeyName),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_2", "name", fmt.Sprintf("%s 2", apiKeyName)),
					resource.TestCheckResourceAttr("data.datadog_api_key.api_key", "name", apiKeyName),
					resource.TestCheckResourceAttrPair("data.datadog_api_key.api_key", "id", "datadog_api_key.api_key_1", "id"),
				),
			},
		},
	})
}

func TestAccDatadogApiKeyDatasource_matchIdError(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApiKeyIdOnlyConfig("11111111-2222-3333-4444-555555555555"),
				ExpectError: regexp.MustCompile("error getting api key"),
			},
		},
	})
}

func TestAccDatadogApiKeyDatasource_matchNameError(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	apiKeyName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApiKeyNameOnlyConfig(apiKeyName),
				ExpectError: regexp.MustCompile("your query returned no result, please try a less specific search criteria"),
			},
		},
	})
}

func TestAccDatadogApiKeyDatasource_missingParametersError(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogApiKeyDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceApiKeyMissingParametersConfig(),
				ExpectError: regexp.MustCompile("missing id or name parameter"),
			},
		},
	})
}

func testAccApiKeyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_api_key" "api_key_1" {
  name = "%s 1"
}

resource "datadog_api_key" "api_key_2" {
  name = "%s 2"
}`, uniq, uniq)
}

func testAccDatasourceApiKeyIdConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_api_key" "api_key" {
  depends_on = [
    datadog_api_key.api_key_1,
    datadog_api_key.api_key_2,
  ]
  id = datadog_api_key.api_key_1.id
}`, testAccApiKeyConfig(uniq))
}

func testAccDatasourceApiKeyIdOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_api_key" "api_key" {
  id = "%s"
}`, uniq)
}

func testAccDatasourceApiKeyNameConfig(uniq string) string {
	return fmt.Sprintf(`
%s
data "datadog_api_key" "api_key" {
  depends_on = [
    datadog_api_key.api_key_1,
    datadog_api_key.api_key_2,
  ]
  name = datadog_api_key.api_key_1.name
}`, testAccApiKeyConfig(uniq))
}

func testAccDatasourceApiKeyNameOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_api_key" "api_key" {
  name = "%s"
}`, uniq)
}

func testAccDatasourceApiKeyMissingParametersConfig() string {
	return `
data "datadog_api_key" "api_key" {
}`
}

func testAccDatasourceApiKeyExactMatch(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_api_key" "api_key_1" {
  name = "%s"
}

resource "datadog_api_key" "api_key_2" {
  name = "%s 2"
}
data "datadog_api_key" "api_key" {
  depends_on = [
    datadog_api_key.api_key_1,
    datadog_api_key.api_key_2,
  ]
  exact_match = true
  name = "%s"
}`, uniq, uniq, uniq)
}
