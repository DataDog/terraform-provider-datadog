package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogReferenceTableDataSource(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDatadogReferenceTableConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_id", "table_name", fmt.Sprintf("tf_test_ds_%s", uniq)),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_id", "source", "S3"),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_id", "description", "Test data source"),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_id", "file_metadata.cloud_storage.sync_enabled", "true"),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_id", "schema.primary_keys.0", "id"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table.by_id", "id"),
					resource.TestCheckResourceAttrSet(
						"data.datadog_reference_table.by_id", "created_by"),
					// Test querying by table_name
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_name", "table_name", fmt.Sprintf("tf_test_ds_%s", uniq)),
					resource.TestCheckResourceAttr(
						"data.datadog_reference_table.by_name", "source", "S3"),
				),
			},
		},
	})
}

func testAccDataSourceDatadogReferenceTableConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "test" {
  table_name  = "tf_test_ds_%s"
  description = "Test data source"
  source      = "S3"

  file_metadata {
    sync_enabled = true

    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "test-bucket"
        file_path       = "data/test.csv"
      }
    }
  }

  schema {
    primary_keys = ["id"]

    fields {
      name = "id"
      type = "STRING"
    }

    fields {
      name = "value"
      type = "STRING"
    }
  }

  tags = ["test:datasource"]
}

data "datadog_reference_table" "by_id" {
  id = datadog_reference_table.test.id
}

data "datadog_reference_table" "by_name" {
  table_name = datadog_reference_table.test.table_name
}
`, uniq)
}

