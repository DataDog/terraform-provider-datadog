package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSensitiveDataScannerRuleBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource_name := "datadog_sensitive_data_scanner_rule.sample_rule"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSensitiveDataScannerRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSensitiveDataScannerRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(
						resource_name, "description", "a description"),
					resource.TestCheckResourceAttr(
						resource_name, "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						resource_name, "name", uniq),
					resource.TestCheckResourceAttr(
						resource_name, "pattern", "regex"),
					resource.TestCheckResourceAttr(
						resource_name, "excluded_namespaces.0", "username"),
				),
			},
			{
				Config: testAccCheckDatadogSensitiveDataScannerRuleUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(
						resource_name, "description", "another description"),
					resource.TestCheckResourceAttr(
						resource_name, "is_enabled", "false"),
					resource.TestCheckResourceAttr(
						resource_name, "name", uniq),
					resource.TestCheckResourceAttr(
						resource_name, "pattern", "regex"),
					resource.TestCheckResourceAttr(
						resource_name, "excluded_namespaces.0", "email"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.number_of_chars", "10"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.type", "partial_replacement_from_beginning"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.replacement_string", ""),
					resource.TestCheckResourceAttr(
						resource_name, "included_keyword_configuration.0.keywords.0", "credit card"),
					resource.TestCheckResourceAttr(
						resource_name, "included_keyword_configuration.0.keywords.1", "cc"),
					resource.TestCheckResourceAttr(
						resource_name, "included_keyword_configuration.0.character_count", "20"),
					resource.TestCheckResourceAttr(
						resource_name, "priority", "1"),
				),
			},
			{
				Config: testAccCheckDatadogSensitiveDataScannerRuleChangedGroup(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(
						resource_name, "description", "another description"),
					resource.TestCheckResourceAttr(
						resource_name, "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						resource_name, "name", uniq),
					resource.TestCheckResourceAttr(
						resource_name, "pattern", "regex"),
					resource.TestCheckResourceAttr(
						resource_name, "excluded_namespaces.0", "email"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.number_of_chars", "10"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.type", "partial_replacement_from_beginning"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.replacement_string", ""),
					resource.TestCheckResourceAttr(
						resource_name, "included_keyword_configuration.0.keywords.0", "credit card"),
					resource.TestCheckResourceAttr(
						resource_name, "included_keyword_configuration.0.keywords.1", "cc"),
					resource.TestCheckResourceAttr(
						resource_name, "included_keyword_configuration.0.character_count", "20"),
					resource.TestCheckResourceAttr(
						resource_name, "priority", "1"),
				),
			},
			{
				Config: testAccCheckDatadogSensitiveDataScannerRuleChangedGroupNone(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(
						resource_name, "description", "another description"),
					resource.TestCheckResourceAttr(
						resource_name, "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						resource_name, "name", uniq),
					resource.TestCheckResourceAttr(
						resource_name, "pattern", "regex"),
					resource.TestCheckResourceAttr(
						resource_name, "excluded_namespaces.0", "email"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.type", "none"),
				),
			},
		},
	})
}

func TestAccSensitiveDataScannerRuleWithStandardPattern(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource_name := "datadog_sensitive_data_scanner_rule.another_rule"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogSensitiveDataScannerRuleDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSensitiveDataScannerRuleWithStandardPattern(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider, resource_name),
					resource.TestCheckResourceAttr(
						resource_name, "description", "a description"),
					resource.TestCheckResourceAttr(
						resource_name, "is_enabled", "true"),
					resource.TestCheckResourceAttr(
						resource_name, "name", uniq),
					resource.TestCheckResourceAttr(
						resource_name, "excluded_namespaces.0", "username"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.number_of_chars", "10"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.type", "partial_replacement_from_beginning"),
					resource.TestCheckResourceAttr(
						resource_name, "text_replacement.0.replacement_string", ""),
				),
			},
		},
	})
}

func testAccCheckDatadogSensitiveDataScannerRule(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name = "my group"
	is_enabled = true
	product_list = ["logs"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_rule" "sample_rule" {
	name = "%s"
	description = "a description"
	excluded_namespaces = ["username"]
	is_enabled = true
	group_id = datadog_sensitive_data_scanner_group.sample_group.id
	pattern = "regex"
	tags = ["sensitive_data:true"]
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerRuleNoTags(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name = "my group"
	is_enabled = true
	product_list = ["logs"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_rule" "sample_rule" {
	name = "%s"
	description = "a description"
	excluded_namespaces = ["username"]
	is_enabled = true
	group_id = datadog_sensitive_data_scanner_group.sample_group.id
	pattern = "regex"
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerRuleUpdate(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name = "my group"
	is_enabled = false
	product_list = ["logs"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_rule" "sample_rule" {
	name = "%s"
	description = "another description"
	excluded_namespaces = ["email"]
	is_enabled = false
	group_id = datadog_sensitive_data_scanner_group.sample_group.id
	pattern = "regex"
	tags = ["sensitive_data:true"]
	text_replacement {
		number_of_chars = 10
		replacement_string = ""
		type = "partial_replacement_from_beginning"
	}
	included_keyword_configuration {
		keywords = ["credit card", "cc"]
		character_count = 20
	}
	priority = 1
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerRuleChangedGroup(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name = "my group"
	is_enabled = false
	product_list = ["logs"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_group" "new_group" {
	name = "another group"
	is_enabled = false
	product_list = ["apm"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_rule" "sample_rule" {
	name = "%s"
	description = "another description"
	excluded_namespaces = ["email"]
	is_enabled = true
	group_id = datadog_sensitive_data_scanner_group.new_group.id
	pattern = "regex"
	tags = ["sensitive_data:true"]
	text_replacement {
		number_of_chars = 10
		replacement_string = ""
		type = "partial_replacement_from_beginning"
	}
	included_keyword_configuration {
		keywords = ["credit card", "cc"]
		character_count = 20
	}
	priority = 1
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerRuleChangedGroupNone(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name = "my group"
	is_enabled = false
	product_list = ["logs"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_group" "new_group" {
	name = "another group"
	is_enabled = false
	product_list = ["apm"]
	filter {
		query = "*"
	}
}

resource "datadog_sensitive_data_scanner_rule" "sample_rule" {
	name = "%s"
	description = "another description"
	excluded_namespaces = ["email"]
	is_enabled = true
	group_id = datadog_sensitive_data_scanner_group.new_group.id
	pattern = "regex"
	tags = ["sensitive_data:true"]
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerRuleWithStandardPattern(name string) string {
	return fmt.Sprintf(`
resource "datadog_sensitive_data_scanner_group" "sample_group" {
	name = "my group"
	is_enabled = true
	product_list = ["logs"]
	filter {
		query = "*"
	}
}

data "datadog_sensitive_data_scanner_standard_pattern" "sample_sp" {
	filter = "AWS Access Key ID Scanner"
}

resource "datadog_sensitive_data_scanner_rule" "another_rule" {
	name = "%s"
	description = "a description"
	excluded_namespaces = ["username"]
	is_enabled = true
	group_id = datadog_sensitive_data_scanner_group.sample_group.id
	standard_pattern_id = data.datadog_sensitive_data_scanner_standard_pattern.sample_sp.id
	text_replacement {
		number_of_chars = 10
		replacement_string = ""
		type = "partial_replacement_from_beginning"
	}
}
`, name)
}

func testAccCheckDatadogSensitiveDataScannerRuleDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		for _, resource := range s.RootModule().Resources {
			if resource.Type == "datadog_sensitive_data_scanner_rule" {
				resp, _, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
				if ruleFound := findSensitiveDataScannerRuleHelper(resource.Primary.ID, resp); ruleFound == nil {
					if err != nil {
						return fmt.Errorf("received an error retrieving all scanning groups: %s", err)
					}
					return nil
				}
				return fmt.Errorf("scanning rule still exists")
			}
		}
		return nil
	}
}

func findSensitiveDataScannerRuleHelper(ruleId string, response datadogV2.SensitiveDataScannerGetConfigResponse) *datadogV2.SensitiveDataScannerRuleIncludedItem {
	for _, resource := range response.GetIncluded() {
		if resource.SensitiveDataScannerRuleIncludedItem.GetId() == ruleId {
			return resource.SensitiveDataScannerRuleIncludedItem
		}
	}

	return nil
}

func testAccCheckDatadogSensitiveDataScannerRuleExists(accProvider func() (*schema.Provider, error), name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		ruleId := s.RootModule().Resources[name].Primary.ID
		resp, _, err := apiInstances.GetSensitiveDataScannerApiV2().ListScanningGroups(auth)
		if err != nil {
			return fmt.Errorf("received an error retrieving the list of scanning groups, %s", err)
		}

		if ruleFound := findSensitiveDataScannerRuleHelper(ruleId, resp); ruleFound == nil {
			return fmt.Errorf("received an error retrieving scanning group")
		}

		return nil
	}
}

func TestAccDatadogSensitiveDataScannerRule_DefaultTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{ // New tags are correctly added
				Config: testAccCheckDatadogSensitiveDataScannerRule(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.*", "sensitive_data:true"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.*", "default_key:default_value"),
				),
			},
			{ // Resource tags take precedence over default tags
				Config: testAccCheckDatadogSensitiveDataScannerRule(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"sensitive_data": "false",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.*", "sensitive_data:true"),
				),
			},
			{ // Resource tags take precedence over default tags, but new tags are added
				Config: testAccCheckDatadogSensitiveDataScannerRule(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"sensitive_data": "false",
						"new_tag":        "new_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.*", "sensitive_data:true"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.*", "new_tag:new_value"),
				),
			},
			{ // Works with monitors without a tag attribute
				Config: testAccCheckDatadogSensitiveDataScannerRuleNoTags(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						"datadog_sensitive_data_scanner_rule.sample_rule", "tags.*", "default_key:default_value"),
				),
			},
		},
	})
}
