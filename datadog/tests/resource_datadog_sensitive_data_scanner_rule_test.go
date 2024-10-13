package test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"
	"text/template"

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
		}})
}

func TestAccSensitiveDataScannerRuleWithTests(t *testing.T) {
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	ctx, accProviders := testAccProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	cfg := func(ruleCfg string) string {
		var output bytes.Buffer
		_ = template.Must(template.New("config").Parse(`
			resource datadog_sensitive_data_scanner_group {{ .Name }} {
				name = "{{ .Name }}"
				is_enabled = false
				product_list = ["logs"]
				filter {
					query = "*"
				}
			}
			resource datadog_sensitive_data_scanner_rule {{ .Name }} {
				name = "{{ .Name }}"
				group_id = datadog_sensitive_data_scanner_group.{{ .Name }}.id
				{{ .RuleCfg }}
			}
		`)).Execute(&output, map[string]string{"Name": name, "RuleCfg": ruleCfg})
		return output.String()
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: cfg(`
					pattern = "needle"
					pattern_test {
						input = "Find the needle in the haystack"
					}
				`),
			},
			{
				Config: cfg(`
					pattern = "needle"
					pattern_test {
						input = "oops no pattern"
					}
				`),
				ExpectError: regexp.MustCompile(`The pattern_test input "oops no pattern" does not match "needle"`),
			},
			{
				Config: cfg(`
					pattern = "my_secret_token[=:]\w+"
					pattern_test {
						input = "my_secret_token=aaaaaaaaaaa"
					}
					pattern_test {
						input = "my_secret_token:bbbbbbbbbb"
					}
					pattern_test {
						input = "my_secret_token_hash=ccccccccc"
						matches = false
					}
				`),
			},
		}})
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
