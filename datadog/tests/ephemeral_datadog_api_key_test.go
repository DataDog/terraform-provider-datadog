package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Note: Ephemeral resources don't persist to state, so we cannot use resource.TestCheckResourceAttr
// to verify their computed attributes. These tests verify that:
// 1. The ephemeral resource can be configured and opened without errors
// 2. Error cases are handled correctly

func TestAccDatadogEphemeralApiKey_matchId(t *testing.T) {
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
				Config: testAccEphemeralApiKeyIdConfig(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, "datadog_api_key.api_key_1"),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_1", "name", fmt.Sprintf("%s 1", apiKeyName)),
				),
			},
		},
	})
}

func TestAccDatadogEphemeralApiKey_matchName(t *testing.T) {
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
				Config: testAccEphemeralApiKeyNameConfig(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, "datadog_api_key.api_key_1"),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_1", "name", fmt.Sprintf("%s 1", apiKeyName)),
				),
			},
		},
	})
}

func TestAccDatadogEphemeralApiKey_exactMatchName(t *testing.T) {
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
				Config: testAccEphemeralApiKeyExactMatch(apiKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogApiKeyExists(providers.frameworkProvider, "datadog_api_key.api_key_1"),
					resource.TestCheckResourceAttr("datadog_api_key.api_key_1", "name", apiKeyName),
				),
			},
		},
	})
}

func TestAccDatadogEphemeralApiKey_matchIdError(t *testing.T) {
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
				Config:      testAccEphemeralApiKeyIdOnlyConfig("11111111-2222-3333-4444-555555555555"),
				ExpectError: regexp.MustCompile("error getting api key"),
			},
		},
	})
}

func TestAccDatadogEphemeralApiKey_matchNameError(t *testing.T) {
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
				Config:      testAccEphemeralApiKeyNameOnlyConfig(apiKeyName),
				ExpectError: regexp.MustCompile("your query returned no result, please try a less specific search criteria"),
			},
		},
	})
}

func TestAccDatadogEphemeralApiKey_missingParametersError(t *testing.T) {
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
				Config:      testAccEphemeralApiKeyMissingParametersConfig(),
				ExpectError: regexp.MustCompile("missing id or name parameter"),
			},
		},
	})
}

func testAccEphemeralApiKeyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_api_key" "api_key_1" {
  name = "%s 1"
}

resource "datadog_api_key" "api_key_2" {
  name = "%s 2"
}`, uniq, uniq)
}

func testAccEphemeralApiKeyIdConfig(uniq string) string {
	return fmt.Sprintf(`
%s
ephemeral "datadog_api_key" "api_key" {
  depends_on = [
    datadog_api_key.api_key_1,
    datadog_api_key.api_key_2,
  ]
  id = datadog_api_key.api_key_1.id
}`, testAccEphemeralApiKeyConfig(uniq))
}

func testAccEphemeralApiKeyIdOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
ephemeral "datadog_api_key" "api_key" {
  id = "%s"
}`, uniq)
}

func testAccEphemeralApiKeyNameConfig(uniq string) string {
	return fmt.Sprintf(`
%s
ephemeral "datadog_api_key" "api_key" {
  depends_on = [
    datadog_api_key.api_key_1,
    datadog_api_key.api_key_2,
  ]
  name = datadog_api_key.api_key_1.name
}`, testAccEphemeralApiKeyConfig(uniq))
}

func testAccEphemeralApiKeyNameOnlyConfig(uniq string) string {
	return fmt.Sprintf(`
ephemeral "datadog_api_key" "api_key" {
  name = "%s"
}`, uniq)
}

func testAccEphemeralApiKeyMissingParametersConfig() string {
	return `
ephemeral "datadog_api_key" "api_key" {
}`
}

func testAccEphemeralApiKeyExactMatch(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_api_key" "api_key_1" {
  name = "%s"
}

resource "datadog_api_key" "api_key_2" {
  name = "%s 2"
}
ephemeral "datadog_api_key" "api_key" {
  depends_on = [
    datadog_api_key.api_key_1,
    datadog_api_key.api_key_2,
  ]
  exact_match = true
  name = "%s"
}`, uniq, uniq, uniq)
}
