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
			{
				Config: testAccCheckDatadogAwsAccountV2(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsAccountV2Exists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_aws_account_v2.foo", "aws_partition", "aws"),
				),
			},
		},
	})
}

func testAccCheckDatadogAwsAccountV2(accountID, uniq string) string {
	// Update me to make use of the unique value
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
  	}
    resources_config {
		cloud_security_posture_management_collection = false
		extended_collection = true
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
