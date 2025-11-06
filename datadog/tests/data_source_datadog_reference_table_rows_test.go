package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogReferenceTableRowsDataSource(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDatadogReferenceTableRowsConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "table_id"),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table_rows.test", "row_ids.#", "2"),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table_rows.test", "rows.#", "2"),
					// We can't predict the exact row IDs or values without actual data,
					// but we can check the structure exists
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "rows.0.id"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "rows.0.values.%"),
				),
			},
		},
	})
}

func testAccDataSourceDatadogReferenceTableRowsConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "test" {
  table_name  = "tf_test_ds_rows_%s"
  description = "Test data source for rows"
  source      = "S3"

  file_metadata {
    sync_enabled = true

    access_details {
      aws_detail {
        aws_account_id  = "924305315327"
        aws_bucket_name = "dd-reference-tables-dev-staging"
        file_path       = "test.csv"
      }
    }
  }

  schema {
    primary_keys = ["a"]

    fields {
      name = "a"
      type = "STRING"
    }

    fields {
      name = "b"
      type = "STRING"
    }

    fields {
      name = "c"
      type = "STRING"
    }
  }

  tags = ["test:datasource-rows"]
}

# Note: In a real test scenario, you would populate rows first
# This example assumes some rows exist with these IDs
data "datadog_reference_table_rows" "test" {
  table_id = datadog_reference_table.test.id
  row_ids  = ["row1", "row2"]
}
`, strings.ToLower(strings.ReplaceAll(uniq, "-", "_")))
}
