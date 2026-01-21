package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogReferenceTableRowsDataSource(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying")
	}
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Step 1: Create the table (rows will sync asynchronously)
				Config: testAccDataSourceDatadogReferenceTableRowsConfigStep1(uniq),
				Check: resource.ComposeTestCheckFunc(
					// Verify the table was created
					resource.TestCheckResourceAttrSet(
						"datadog_reference_table.test", "id"),
					// Note: We don't check row_count here because sync is asynchronous
					// Step 2 will wait for rows to be available
				),
			},
			{
				// Step 2: Query rows from the existing synced table
				Config: testAccDataSourceDatadogReferenceTableRowsConfigStep2(uniq),
				Check: resource.ComposeTestCheckFunc(
					// Verify the data source configuration
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "table_id"),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table_rows.test", "row_ids.#", "2"),
					// Verify rows were retrieved
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table_rows.test", "rows.#", "2"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "rows.0.id"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "rows.0.values.%"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "rows.1.id"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table_rows.test", "rows.1.values.%"),
				),
			},
		},
	})
}

func testAccDataSourceDatadogReferenceTableRowsConfigStep1(uniq string) string {
	sanitized := strings.ToLower(strings.ReplaceAll(uniq, "-", "_"))
	return fmt.Sprintf(`
# Step 1: Create the table and wait for it to sync
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
`, sanitized)
}

func testAccDataSourceDatadogReferenceTableRowsConfigStep2(uniq string) string {
	sanitized := strings.ToLower(strings.ReplaceAll(uniq, "-", "_"))
	return fmt.Sprintf(`
# Step 1: Create the table and wait for it to sync
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

# Step 2: Query rows from the existing synced table
# test.csv contains rows with primary key values "1" and "2"
data "datadog_reference_table_rows" "test" {
  table_id = datadog_reference_table.test.id
  row_ids  = ["1", "2"]
}
`, sanitized)
}
