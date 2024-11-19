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

func TestAccAwsAccountV2Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	accountID := uniqueAWSAccountID(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAwsAccountV2Destroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{ // create account
				Config: testAccCheckDatadogAwsAccountV2(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsAccountV2Exists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_regions.aws_regions_include_only.include_only.0", "us-east-1"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "auth_config.aws_auth_config_role.role_name", "test"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "logs_config.lambda_forwarder.sources.0", "s3"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "logs_config.lambda_forwarder.lambdas.0", "arn:aws:lambda:us-east-1:123456789123:function:test-fn"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.automute_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.collect_cloudwatch_alarms", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.collect_custom_metrics", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.namespace_filters.aws_namespace_filters_include_only.include_only.0", "AWS/EC2"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "resources_config.cloud_security_posture_management_collection", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "resources_config.extended_collection", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "traces_config.xray_services.x_ray_services_include_all.include_all", "true"),
				),
			}, { // update account
				Config: testAccCheckDatadogAwsAccountV2UpdateConfig(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsAccountV2Exists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_partition", "aws"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_regions.aws_regions_include_all.include_all", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "auth_config.aws_auth_config_role.role_name", fmt.Sprintf("test-%s", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "logs_config.lambda_forwarder.sources.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "logs_config.lambda_forwarder.lambdas.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.automute_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.collect_cloudwatch_alarms", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.collect_custom_metrics", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "metrics_config.namespace_filters.aws_namespace_filters_include_only.exclude_only.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "resources_config.cloud_security_posture_management_collection", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "resources_config.extended_collection", "true"),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "traces_config.xray_services.x_ray_services_include_only.include_only.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDatadogAwsAccountV2(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_aws_account_v2" "foo" {
        aws_account_id = %s
        account_tags = ["tag:%s"]
        aws_partition = "aws"
    aws_regions {
        aws_regions_include_only {
            include_only = ["us-east-1"]
        }
    }
    auth_config {
        aws_auth_config_role {
            role_name = "test"
        }
    }
    logs_config {
        lambda_forwarder {
            sources = ["s3", "lambda"]
            lambdas = ["arn:aws:lambda:us-east-1:123456789123:function:test-fn"]
        }
    }
        metrics_config {
            automute_enabled = true
            collect_cloudwatch_alarms = true
            collect_custom_metrics = true
            enabled = true
        namespace_filters {
        aws_namespace_filters_include_only {
            include_only = ["AWS/EC2"]
        }
        }

        }
        resources_config {
        cloud_security_posture_management_collection = false
        extended_collection = false
        }
    traces_config {
        xray_services {
            x_ray_services_include_all {
                include_all = true
            }
        }
    }
}`, accountID, uniq)
}

func testAccCheckDatadogAwsAccountV2UpdateConfig(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_aws_account_v2" "foo" {
    aws_account_id = %s
    account_tags = []
    aws_partition = "aws"
  aws_regions {
    aws_regions_include_all {
      include_all = true
    }
  }
  auth_config {
    aws_auth_config_role {
      role_name = "test-%s"
    }
  }
  logs_config {
    lambda_forwarder {
      sources = []
      lambdas = []
    }
  }
    metrics_config {
      automute_enabled = false
      collect_cloudwatch_alarms = false
      collect_custom_metrics = false
      enabled = false
    namespace_filters {
    aws_namespace_filters_exclude_only {
      exclude_only = []
    }
    }

    }
    resources_config {
    cloud_security_posture_management_collection = true
    extended_collection = true
    }
  traces_config {
    xray_services {
      x_ray_services_include_only {
        include_only = []
      }
    }
  }
}`, accountID, uniq)
}

func testAccCheckDatadogAwsAccountV2Destroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AwsAccountV2DestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AwsAccountV2DestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_aws_account_v2" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccount(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AwsAccountV2 %s", err)}
			}
			return &utils.RetryableError{Prob: "AwsAccountV2 still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAwsAccountV2Exists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := awsAccountV2ExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func awsAccountV2ExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_aws_account_v2" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccount(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AwsAccountV2")
		}
	}
	return nil
}
