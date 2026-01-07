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
				ExpectError: regexp.MustCompile("Your query returned more than one result, please try a more specific search criteria"),
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

// TestAccDatadogStandardPatternDatasourceExactMatch tests that exact match takes priority
// over partial match. This addresses issue #3370 where searching for "US Tax..." would
// incorrectly match "Cyprus Tax..." due to the "us" substring.
func TestAccDatadogStandardPatternDatasourceExactMatch(t *testing.T) {
	t.Parallel()
	if isRecording() || isReplaying() {
		t.Skip("This test doesn't support recording or replaying")
	}

	_, accProviders := testAccProviders(context.Background(), t)

	// Use a pattern name that could potentially match multiple patterns via substring
	// but should return exactly the one we're looking for via exact match
	exactName := "US Tax Identification Number Scanner"
	datasourceName := "data.datadog_sensitive_data_scanner_standard_pattern.sample_sp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceStandardPatternConfig(exactName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", exactName),
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
