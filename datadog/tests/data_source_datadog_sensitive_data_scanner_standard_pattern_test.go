package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogStandardPatternDatasourceNameFilter(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	_, accProviders := testAccProviders(context.Background(), t)
	name := "AWS Access Key ID Scanner"

	datasource_name := "data.datadog_sensitive_data_scanner_standard_pattern.sample_sp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceStandardPatternConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						datasource_name, "name", name),
				),
			},
		},
	})
}

func TestAccDatadogStandardPatternDatasourceErrorMultiple(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceStandardPatternConfig("aws"),
				ExpectError: regexp.MustCompile("Couldn't find the standard pattern with name aws"),
			},
		},
	})
}

func TestAccDatadogStandardPatternDatasourceErrorNotFound(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDatasourceStandardPatternConfig("foobarbaz"),
				ExpectError: regexp.MustCompile("Couldn't find the standard pattern with name foobarbaz"),
			},
		},
	})
}

func TestAccDatadogStandardPatternDatasourceIDFilter(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	_, accProviders := testAccProviders(context.Background(), t)

	datasourceName := "data.datadog_sensitive_data_scanner_standard_pattern.by_id"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceStandardPatternConfigByIDReference("AWS Access Key ID Scanner"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", "AWS Access Key ID Scanner"),
				),
			},
		},
	})
}

func TestAccDatadogStandardPatternDatasourceExactNameFilter(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	_, accProviders := testAccProviders(context.Background(), t)
	datasourceName := "data.datadog_sensitive_data_scanner_standard_pattern.sample_sp"
	fullName := "US Tax Identification Number Scanner"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceStandardPatternConfig(fullName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", fullName),
				),
			},
		},
	})
}

func testAccDatasourceStandardPatternConfig(name string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_standard_pattern" "sample_sp" {
  filter = "%s"
}`, name)
}

func testAccDatasourceStandardPatternConfigByID(id string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_standard_pattern" "sample_sp" {
  standard_pattern_id = "%s"
}`, id)
}

func testAccDatasourceStandardPatternConfigByIDReference(name string) string {
	return fmt.Sprintf(`
data "datadog_sensitive_data_scanner_standard_pattern" "by_name" {
  filter = "%s"
}

data "datadog_sensitive_data_scanner_standard_pattern" "by_id" {
  standard_pattern_id = data.datadog_sensitive_data_scanner_standard_pattern.by_name.id
}`, name)
}
