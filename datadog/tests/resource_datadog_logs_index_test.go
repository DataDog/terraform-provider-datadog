package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogLogsIndex_Basic(t *testing.T) {

	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateLogsIndexConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit", "200000"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit_reset.0.reset_time", "10:00"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit_reset.0.reset_utc_offset", "+02:00"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit_warning_threshold_percentage", "70"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "disable_daily_limit", "false"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "retention_days", "15"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "flex_retention_days", "180"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "filter.#", "1"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "filter.0.query", "non-existent-query"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.#", "0"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateLogsIndexConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit", "20000"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit_reset.0.reset_time", "12:00"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit_reset.0.reset_utc_offset", "-02:00"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "daily_limit_warning_threshold_percentage", "99.99"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "disable_daily_limit", "false"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "retention_days", "15"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "filter.0.query", "test:query"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.name", "Filter coredns logs"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.is_enabled", "true"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.filter.0.query", "app:coredns"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.0.filter.0.sample_rate", "0.97"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateLogsIndexDisableDailyLimitConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "disable_daily_limit", "true"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "retention_days", "15"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "exclusion_filter.#", "0"),
				),
			},
			{
				Config: testAccCheckDatadogUpdateLogsIndexZeroRetentionConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "retention_days", "0"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_index", "flex_retention_days", "360"),
				),
			},
			{
				Config: testAccCheckDatadogCreateFlexOnlyLogsIndexConfig(uniq + "-flex"),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_flex_index", "name", uniq+"-flex"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_flex_index", "retention_days", "0"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_flex_index", "flex_retention_days", "360"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_flex_index", "filter.#", "1"),
					resource.TestCheckResourceAttr("datadog_logs_index.sample_flex_index", "filter.0.query", "non-existent-flex-query"),
				),
			},
		},
	})
}

func sleep() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if isReplaying() {
			return nil
		}
		time.Sleep(10 * time.Second)
		return nil
	}
}

func testAccCheckDatadogCreateLogsIndexConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name           = "%s"
  daily_limit    = 200000
  daily_limit_reset {
	reset_time = "10:00"
	reset_utc_offset = "+02:00"
  }
  daily_limit_warning_threshold_percentage = 70
  retention_days = 15
  flex_retention_days = 180
  filter {
    query = "non-existent-query"
  }
}
`, name)
}

func testAccCheckDatadogCreateFlexOnlyLogsIndexConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_flex_index" {
  name           = "%s"
  flex_retention_days = 360
  filter {
    query = "non-existent-flex-query"
  }
}
`, name)
}

func testAccCheckDatadogUpdateLogsIndexConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name                   = "%s"
  daily_limit            = 20000
  daily_limit_reset {
	reset_time = "12:00"
	reset_utc_offset = "-02:00"
  }
  daily_limit_warning_threshold_percentage = 99.99
  disable_daily_limit    = false
  retention_days         = 15
  flex_retention_days = 180
  filter {
    query                = "test:query"
  }
  exclusion_filter {
    name                 = "Filter coredns logs"
    is_enabled           = true
    filter {
      query              = "app:coredns"
      sample_rate        = 0.97
    }
  }
}
`, name)
}

func testAccCheckDatadogUpdateLogsIndexDisableDailyLimitConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name                   = "%s"
  daily_limit            = 20000
  daily_limit_reset {
	reset_time = "10:00"
	reset_utc_offset = "+02:00"
  }
  daily_limit_warning_threshold_percentage = 70
  disable_daily_limit    = true
  retention_days         = 15
  flex_retention_days = 180
  filter {
    query                = "test:query"
  }
}
`, name)
}

func testAccCheckDatadogUpdateLogsIndexZeroRetentionConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "sample_index" {
  name                   = "%s"
  daily_limit            = 20000
  daily_limit_reset {
	reset_time = "10:00"
	reset_utc_offset = "+02:00"
  }
  daily_limit_warning_threshold_percentage = 70
  disable_daily_limit    = true
  retention_days         = 0
  flex_retention_days = 360
  filter {
    query                = "test:query"
  }
}
`, name)
}

func TestAccDatadogLogsIndex_WithTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogLogsIndexWithTagsConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.tagged_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.tagged_index", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("datadog_logs_index.tagged_index", "tags.*", "team:backend"),
					resource.TestCheckTypeSetElemAttr("datadog_logs_index.tagged_index", "tags.*", "env:test"),
				),
			},
			{
				Config: testAccCheckDatadogLogsIndexWithTagsUpdatedConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.tagged_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.tagged_index", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr("datadog_logs_index.tagged_index", "tags.*", "team:frontend"),
					resource.TestCheckTypeSetElemAttr("datadog_logs_index.tagged_index", "tags.*", "env:staging"),
					resource.TestCheckTypeSetElemAttr("datadog_logs_index.tagged_index", "tags.*", "service:api"),
				),
			},
			{
				Config: testAccCheckDatadogLogsIndexWithoutTagsConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					sleep(),
					resource.TestCheckResourceAttr("datadog_logs_index.tagged_index", "name", uniq),
					resource.TestCheckResourceAttr("datadog_logs_index.tagged_index", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDatadogLogsIndexWithTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "tagged_index" {
  name           = "%s"
  retention_days = 15
  filter {
    query = "non-existent-tags-query"
  }
  tags = ["team:backend", "env:test"]
}
`, name)
}

func testAccCheckDatadogLogsIndexWithTagsUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "tagged_index" {
  name           = "%s"
  retention_days = 15
  filter {
    query = "non-existent-tags-query"
  }
  tags = ["team:frontend", "env:staging", "service:api"]
}
`, name)
}

func testAccCheckDatadogLogsIndexWithoutTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_logs_index" "tagged_index" {
  name           = "%s"
  retention_days = 15
  filter {
    query = "non-existent-tags-query"
  }
}
`, name)
}
