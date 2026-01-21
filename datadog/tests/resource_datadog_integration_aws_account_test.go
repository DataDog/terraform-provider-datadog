package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccIntegrationAwsAccount_RoleBased(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accountID := uniqueAWSAccountID(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{ // create role-based account with defaults
				Config: testAccCheckDatadogIntegrationAwsAccount_Create_RoleBased(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_regions.include_all", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.sources.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.lambdas.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.source", "s3"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.tags.0", "tag1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.tags.1", "tag2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.automute_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.collect_cloudwatch_alarms", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.collect_custom_metrics", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.namespace_filters.exclude_only.0", "AWS/SQS"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.namespace_filters.exclude_only.1", "AWS/ElasticMapReduce"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.namespace_filters.exclude_only.2", "AWS/Usage"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "resources_config.cloud_security_posture_management_collection", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "resources_config.extended_collection", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "traces_config.xray_services.include_only.#", "0"),
				),
			}, { // update account
				Config: testAccCheckDatadogIntegrationAwsAccount_Update_RoleBased(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_regions.include_only.0", "us-east-1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "auth_config.aws_auth_config_role.role_name", fmt.Sprintf("test-%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.sources.0", "s3"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.lambdas.0", "arn:aws:lambda:us-east-1:123456789123:function:test-fn"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.source", "lambda"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.tags.0", "tag3"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.tags.1", "tag4"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.1.source", "cloudtrail"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.1.tags.0", "tag5"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.1.tags.1", "tag6"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.automute_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.collect_cloudwatch_alarms", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.collect_custom_metrics", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.namespace_filters.include_only.0", "AWS/EC2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.0.namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.0.tags.0", "tag1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.0.tags.1", "tag2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.1.namespace", "AWS/Lambda"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.1.tags.0", "tag3"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.tag_filters.1.tags.1", "tag4"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "resources_config.cloud_security_posture_management_collection", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "resources_config.extended_collection", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "traces_config.xray_services.include_all", "true"),
				),
			}, { // update account and reset nested blocks to default
				Config: testAccCheckDatadogIntegrationAwsAccount_Update_RoleBased_ResetNestedBlocksToDefault(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_regions.include_all", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.sources.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.lambdas.#", "0"),
					// log_source_config should not appear in state when there are no tag_filters
					resource.TestCheckNoResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.automute_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.collect_cloudwatch_alarms", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.collect_custom_metrics", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.namespace_filters.exclude_only.0", "AWS/SQS"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "metrics_config.namespace_filters.exclude_only.1", "AWS/ElasticMapReduce"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "resources_config.cloud_security_posture_management_collection", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "resources_config.extended_collection", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "traces_config.xray_services.include_only.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsAccount_Create_RoleBased(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
        aws_auth_config_role {
            role_name = "test"
        }
    }
    logs_config {
        lambda_forwarder {
            log_source_config {
				tag_filters {
					source = "s3"
					tags = ["tag1", "tag2"]
				}
			}
        }
    }
    metrics_config {
		namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq)
}

func testAccCheckDatadogIntegrationAwsAccount_Update_RoleBased(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["!test-update","*"]
    aws_partition = "aws"
    aws_regions {
      include_only = ["us-east-1"]
    }
    auth_config {
      aws_auth_config_role {
        role_name = "test-%s"
      }
    }
    logs_config {
      lambda_forwarder {
        sources = ["s3", "lambda", "cloudtrail"]
        lambdas = ["arn:aws:lambda:us-east-1:123456789123:function:test-fn"]
		log_source_config {
			tag_filters {
				source = "lambda"
				tags = ["tag3", "tag4"]
			}
			tag_filters {
				source = "cloudtrail"
				tags = ["tag5", "tag6"]
			}
		}
      }
    }
    metrics_config {
      automute_enabled = false
      collect_cloudwatch_alarms = true
      collect_custom_metrics = true
      enabled = false
      namespace_filters {
        include_only = ["AWS/EC2"]
      }
	  tag_filters {
	  	namespace = "AWS/EC2"
	  	tags = ["tag1", "tag2"]
	  }
	  tag_filters {
	  	namespace = "AWS/Lambda"
	  	tags = ["tag3", "tag4"]
	  }
    }
    resources_config {
	    cloud_security_posture_management_collection = true
	    extended_collection = true
    }
    traces_config {
      xray_services {
          include_all = true
      }
    }
}`, accountID, uniq)
}

func testAccCheckDatadogIntegrationAwsAccount_Update_RoleBased_ResetNestedBlocksToDefault(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["!test-update-2","*"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
      aws_auth_config_role {
        role_name = "test-%s"
      }
    }
    logs_config {
        lambda_forwarder {}
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq)
}

func TestAccIntegrationAwsAccountKeyBased(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accountID := uniqueAWSAccountID(ctx, t)
	accessKeyID := uniqueAWSAccessKeyID(ctx, t)
	secretAccessKey := "wJalrXUtnFEMI" + uniq

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{ // create key-based account with defaults
				Config: testAccCheckDatadogIntegrationAwsAccount_Create_KeyBased(accountID, uniq, accessKeyID, secretAccessKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "auth_config.aws_auth_config_keys.access_key_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "auth_config.aws_auth_config_keys.secret_access_key", secretAccessKey),
					// log_source_config should not appear in state when lambda_forwarder is empty
					resource.TestCheckNoResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config"),
				),
			}, { // update account without providing secret key
				Config: testAccCheckDatadogIntegrationAwsAccount_Update_KeyBased(accountID, uniq, accessKeyID, secretAccessKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "auth_config.aws_auth_config_keys.access_key_id", fmt.Sprintf("%sNEW", accessKeyID)),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "auth_config.aws_auth_config_keys.secret_access_key", fmt.Sprintf("%sNEW", secretAccessKey)),
					// log_source_config should not appear in state when lambda_forwarder is empty
					resource.TestCheckNoResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsAccount_Create_KeyBased(accountID, uniq, accessKeyID, secretAccessKey string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
      aws_auth_config_keys {
        access_key_id = "%s"
		secret_access_key = "%s"
      }
    }
    logs_config {
        lambda_forwarder {}
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq, accessKeyID, secretAccessKey)
}

func testAccCheckDatadogIntegrationAwsAccount_Update_KeyBased(accountID, uniq, accessKeyID, secretAccessKey string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
      aws_auth_config_keys {
        access_key_id = "%sNEW"
		secret_access_key = "%sNEW"
      }
    }
    logs_config {
        lambda_forwarder {}
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq, accessKeyID, secretAccessKey)
}

func testAccCheckDatadogIntegrationAwsAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationAwsAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationAwsAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_aws_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccount(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving IntegrationAwsAccount %s", err)}
			}
			return &utils.RetryableError{Prob: "IntegrationAwsAccount still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationAwsAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationAwsAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationAwsAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_aws_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccount(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving IntegrationAwsAccount")
		}
	}
	return nil
}

func TestAccIntegrationAwsAccount_OptionalLogSourceConfig(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accountID := uniqueAWSAccountID(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{ // test that empty log_source_config block is valid and clears filters
				Config: testAccCheckDatadogIntegrationAwsAccount_EmptyLogSourceConfig(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "aws_account_id", accountID),
					// log_source_config should be present but with empty tag_filters
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.#", "0"),
				),
			}, { // test adding tag_filters after having empty config
				Config: testAccCheckDatadogIntegrationAwsAccount_WithLogSourceConfig(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config.tag_filters.0.source", "s3"),
				),
			}, { // test clearing tag_filters by omitting log_source_config entirely
				Config: testAccCheckDatadogIntegrationAwsAccount_NoLogSourceConfig(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountExists(providers.frameworkProvider),
					// log_source_config should not appear in state when omitted
					resource.TestCheckNoResourceAttr(
						"datadog_integration_aws_account.foo", "logs_config.lambda_forwarder.log_source_config"),
				),
			},
		},
	})
}

func TestAccIntegrationAwsAccount_InvalidEmptyTagFilters(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accountID := uniqueAWSAccountID(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogIntegrationAwsAccount_InvalidEmptyTagFilters(accountID, uniq),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsAccount_EmptyLogSourceConfig(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
        aws_auth_config_role {
            role_name = "test"
        }
    }
    logs_config {
        lambda_forwarder {
            log_source_config {}
        }
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq)
}

func testAccCheckDatadogIntegrationAwsAccount_WithLogSourceConfig(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
        aws_auth_config_role {
            role_name = "test"
        }
    }
    logs_config {
        lambda_forwarder {
            log_source_config {
                tag_filters {
                    source = "s3"
                    tags = ["tag1", "tag2"]
                }
            }
        }
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq)
}

func testAccCheckDatadogIntegrationAwsAccount_NoLogSourceConfig(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
        aws_auth_config_role {
            role_name = "test"
        }
    }
    logs_config {
        lambda_forwarder {}
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq)
}

func testAccCheckDatadogIntegrationAwsAccount_InvalidEmptyTagFilters(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = %s
    account_tags = ["tag:%s"]
    aws_partition = "aws"
    aws_regions {}
    auth_config {
        aws_auth_config_role {
            role_name = "test"
        }
    }
    logs_config {
        lambda_forwarder {
            log_source_config {
                tag_filters {
                    source = "s3"
                }
                tag_filters {}
            }
        }
    }
    metrics_config {
        namespace_filters {}
    }
    resources_config {}
    traces_config {
        xray_services {}
    }
}`, accountID, uniq)
}
