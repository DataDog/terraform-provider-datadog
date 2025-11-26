package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccReferenceTableS3_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTableS3(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "table_name", fmt.Sprintf("tf_test_s3_%s", strings.ToLower(strings.ReplaceAll(uniq, "-", "_")))),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "source", "S3"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "description", "Test S3 reference table"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "file_metadata.sync_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "file_metadata.access_details.aws_detail.aws_account_id", "924305315327"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "schema.primary_keys.0", "a"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "schema.fields.0.name", "a"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "schema.fields.0.type", "STRING"),
					resource.TestCheckResourceAttrSet(
						"datadog_reference_table.s3_table", "id"),
					resource.TestCheckResourceAttrSet(
						"datadog_reference_table.s3_table", "created_by"),
				),
			},
		},
	})
}

func TestAccReferenceTable_SchemaOnCreate(t *testing.T) {
	// Test that schema is set correctly on create
	// Note: Schema updates via PATCH are not supported; schema is derived from the file asynchronously
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTableSchemaInitial(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.#", "3"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.0.name", "a"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.1.name", "b"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.2.name", "c"),
				),
			},
			{
				// Wait for table to be DONE and verify schema is preserved
				Config: testAccCheckDatadogReferenceTableSchemaInitial(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					testAccCheckDatadogReferenceTableStatusDone(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet("datadog_reference_table.evolution", "status"),
					resource.TestCheckResourceAttrSet("datadog_reference_table.evolution", "row_count"),
					// Schema should still have 3 fields after sync completes
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.#", "3"),
				),
			},
		},
	})
}

func TestAccReferenceTable_UpdateSyncEnabled(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTableSyncEnabled(uniq, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.sync_test", "file_metadata.sync_enabled", "true"),
				),
			},
			{
				Config: testAccCheckDatadogReferenceTableSyncEnabled(uniq, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.sync_test", "file_metadata.sync_enabled", "false"),
				),
			},
		},
	})
}

func TestAccReferenceTable_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTableS3(uniq),
			},
			{
				ResourceName:      "datadog_reference_table.s3_table",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogReferenceTableS3(uniq string) string {
	// Sanitize: replace dashes with underscores and convert to lowercase
	sanitized := strings.ToLower(strings.ReplaceAll(uniq, "-", "_"))
	return fmt.Sprintf(`
resource "datadog_reference_table" "s3_table" {
  table_name  = "tf_test_s3_%s"
  description = "Test S3 reference table"
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

  tags = ["test:terraform", "env:test"]
}`, sanitized)
}

func testAccCheckDatadogReferenceTableSchemaInitial(uniq string) string {
	// Sanitize: replace dashes with underscores and convert to lowercase
	sanitized := strings.ToLower(strings.ReplaceAll(uniq, "-", "_"))
	return fmt.Sprintf(`
resource "datadog_reference_table" "evolution" {
  table_name  = "tf_test_evolution_%s"
  description = "Test schema evolution"
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

  tags = ["test:terraform"]
}`, sanitized)
}

func testAccCheckDatadogReferenceTableSyncEnabled(uniq string, syncEnabled bool) string {
	// Sanitize: replace dashes with underscores and convert to lowercase
	sanitized := strings.ToLower(strings.ReplaceAll(uniq, "-", "_"))
	return fmt.Sprintf(`
resource "datadog_reference_table" "sync_test" {
  table_name  = "tf_test_sync_%s"
  description = "Test sync_enabled update"
  source      = "S3"

  file_metadata {
    sync_enabled = %t

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

  tags = ["test:terraform"]
}`, sanitized, syncEnabled)
}

func testAccCheckDatadogReferenceTableDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := ReferenceTableDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func ReferenceTableDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_reference_table" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetReferenceTablesApiV2().GetTable(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving ReferenceTable %s", err)}
			}
			return &utils.RetryableError{Prob: "ReferenceTable still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogReferenceTableExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := referenceTableExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func referenceTableExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_reference_table" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetReferenceTablesApiV2().GetTable(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving ReferenceTable")
		}
	}
	return nil
}

func testAccCheckDatadogReferenceTableStatusDone(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_reference_table" {
				continue
			}
			id := r.Primary.ID

			// Wait for table status to be DONE (or ERROR) before proceeding
			maxRetries := 20
			retryInterval := 3 * time.Second
			for i := 0; i < maxRetries; i++ {
				resp, httpResp, err := apiInstances.GetReferenceTablesApiV2().GetTable(auth, id)
				if err != nil {
					return utils.TranslateClientError(err, httpResp, "error retrieving ReferenceTable")
				}

				if resp.Data != nil {
					attrs := resp.Data.GetAttributes()
					if status, ok := attrs.GetStatusOk(); ok && status != nil {
						statusStr := string(*status)
						if statusStr == "DONE" || statusStr == "ERROR" {
							return nil // Table is ready
						}
						if i < maxRetries-1 {
							time.Sleep(retryInterval)
							continue
						}
						return fmt.Errorf("table status is %s after %d retries, expected DONE or ERROR", statusStr, maxRetries)
					}
				}
				if i < maxRetries-1 {
					time.Sleep(retryInterval)
				}
			}
			return fmt.Errorf("unable to verify table status after %d retries", maxRetries)
		}
		return nil
	}
}

func testAccCheckDatadogReferenceTableSchemaUpdated(accProvider *fwprovider.FrameworkProvider, expectedFieldCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_reference_table" {
				continue
			}
			id := r.Primary.ID

			// Wait for schema to be updated (async operation) - typically completes in ~10 seconds
			maxRetries := 5
			retryInterval := 2 * time.Second
			for i := 0; i < maxRetries; i++ {
				resp, httpResp, err := apiInstances.GetReferenceTablesApiV2().GetTable(auth, id)
				if err != nil {
					return utils.TranslateClientError(err, httpResp, "error retrieving ReferenceTable")
				}

				if resp.Data != nil {
					attrs := resp.Data.GetAttributes()
					if schema, ok := attrs.GetSchemaOk(); ok && schema != nil {
						if fields, ok := schema.GetFieldsOk(); ok && fields != nil {
							actualFieldCount := len(*fields)
							if actualFieldCount == expectedFieldCount {
								return nil // Schema matches expected count
							}
							if i < maxRetries-1 {
								time.Sleep(retryInterval)
								continue
							}
							return fmt.Errorf("schema field count is %d after %d retries, expected %d", actualFieldCount, maxRetries, expectedFieldCount)
						}
					}
				}
				if i < maxRetries-1 {
					time.Sleep(retryInterval)
				}
			}
			return fmt.Errorf("unable to verify schema field count after %d retries", maxRetries)
		}
		return nil
	}
}
