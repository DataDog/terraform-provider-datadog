package test

import (
	"context"
	"fmt"
	"testing"

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
						"datadog_reference_table.s3_table", "table_name", fmt.Sprintf("tf_test_s3_%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "source", "S3"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "description", "Test S3 reference table"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "file_metadata.sync_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "file_metadata.access_details.aws_detail.aws_account_id", "123456789000"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "schema.primary_keys.0", "id"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.s3_table", "schema.fields.0.name", "id"),
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

func TestAccReferenceTableGCS_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTableGCS(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.gcs_table", "table_name", fmt.Sprintf("tf_test_gcs_%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.gcs_table", "source", "GCS"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.gcs_table", "file_metadata.access_details.gcp_detail.gcp_project_id", "my-gcp-project"),
				),
			},
		},
	})
}

func TestAccReferenceTableAzure_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogReferenceTableDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogReferenceTableAzure(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.azure_table", "table_name", fmt.Sprintf("tf_test_azure_%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.azure_table", "source", "AZURE"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.azure_table", "file_metadata.access_details.azure_detail.azure_storage_account_name", "datadogstorage"),
				),
			},
		},
	})
}

func TestAccReferenceTable_SchemaEvolution(t *testing.T) {
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
						"datadog_reference_table.evolution", "schema.fields.#", "2"),
				),
			},
			{
				Config: testAccCheckDatadogReferenceTableSchemaAddFields(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogReferenceTableExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.#", "4"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.2.name", "email"),
					resource.TestCheckResourceAttr(
						"datadog_reference_table.evolution", "schema.fields.3.name", "age"),
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
	return fmt.Sprintf(`
resource "datadog_reference_table" "s3_table" {
  table_name  = "tf_test_s3_%s"
  description = "Test S3 reference table"
  source      = "S3"

  file_metadata {
    sync_enabled = true

    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "test-bucket"
        file_path       = "reference-tables/test.csv"
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

  tags = ["test:terraform", "env:test"]
}`, uniq)
}

func testAccCheckDatadogReferenceTableGCS(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "gcs_table" {
  table_name  = "tf_test_gcs_%s"
  description = "Test GCS reference table"
  source      = "GCS"

  file_metadata {
    sync_enabled = true

    access_details {
      gcp_detail {
        gcp_project_id            = "my-gcp-project"
        gcp_bucket_name           = "test-bucket"
        file_path                 = "data/test.csv"
        gcp_service_account_email = "datadog-sa@my-gcp-project.iam.gserviceaccount.com"
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
      name = "name"
      type = "STRING"
    }
  }

  tags = ["test:terraform", "source:gcs"]
}`, uniq)
}

func testAccCheckDatadogReferenceTableAzure(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "azure_table" {
  table_name  = "tf_test_azure_%s"
  description = "Test Azure reference table"
  source      = "AZURE"

  file_metadata {
    sync_enabled = true

    access_details {
      azure_detail {
        azure_tenant_id            = "cccccccc-4444-5555-6666-dddddddddddd"
        azure_client_id            = "aaaaaaaa-1111-2222-3333-bbbbbbbbbbbb"
        azure_storage_account_name = "datadogstorage"
        azure_container_name       = "test-container"
        file_path                  = "data/test.csv"
      }
    }
  }

  schema {
    primary_keys = ["sku"]

    fields {
      name = "sku"
      type = "STRING"
    }

    fields {
      name = "quantity"
      type = "INT32"
    }
  }

  tags = ["test:terraform", "source:azure"]
}`, uniq)
}

func testAccCheckDatadogReferenceTableSchemaInitial(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "evolution" {
  table_name  = "tf_test_evolution_%s"
  description = "Test schema evolution"
  source      = "S3"

  file_metadata {
    sync_enabled = true

    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "test-bucket"
        file_path       = "data/customers.csv"
      }
    }
  }

  schema {
    primary_keys = ["customer_id"]

    fields {
      name = "customer_id"
      type = "STRING"
    }

    fields {
      name = "name"
      type = "STRING"
    }
  }

  tags = ["test:terraform"]
}`, uniq)
}

func testAccCheckDatadogReferenceTableSchemaAddFields(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "evolution" {
  table_name  = "tf_test_evolution_%s"
  description = "Test schema evolution"
  source      = "S3"

  file_metadata {
    sync_enabled = true

    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "test-bucket"
        file_path       = "data/customers.csv"
      }
    }
  }

  schema {
    primary_keys = ["customer_id"]

    fields {
      name = "customer_id"
      type = "STRING"
    }

    fields {
      name = "name"
      type = "STRING"
    }

    # New fields added (additive change)
    fields {
      name = "email"
      type = "STRING"
    }

    fields {
      name = "age"
      type = "INT32"
    }
  }

  tags = ["test:terraform"]
}`, uniq)
}

func testAccCheckDatadogReferenceTableSyncEnabled(uniq string, syncEnabled bool) string {
	return fmt.Sprintf(`
resource "datadog_reference_table" "sync_test" {
  table_name  = "tf_test_sync_%s"
  description = "Test sync_enabled update"
  source      = "S3"

  file_metadata {
    sync_enabled = %t

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

  tags = ["test:terraform"]
}`, uniq, syncEnabled)
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
