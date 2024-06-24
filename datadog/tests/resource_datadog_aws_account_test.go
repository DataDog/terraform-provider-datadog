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

func TestAccAwsAccountBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAwsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAwsAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAwsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_aws_account.foo", "aws_account_id", "123456789012"),
				),
			},
		},
	})
}

func testAccCheckDatadogAwsAccount(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_aws_account" "foo" {
    account_tags = "UPDATE ME"
    auth_config {
    access_key_id = "UPDATE ME"
    external_id = "UPDATE ME"
    role_name = "UPDATE ME"
    secret_access_key = "UPDATE ME"
    }
    aws_account_id = "123456789012"
    aws_regions {
    include_all = "UPDATE ME"
    include_only = "UPDATE ME"
    }
    logs_config {
    lambda_forwarder {
    lambdas = "UPDATE ME"
    sources = "UPDATE ME"
    }
    }
    metrics_config {
    automute_enabled = "UPDATE ME"
    collect_cloudwatch_alarms = "UPDATE ME"
    collect_custom_metrics = "UPDATE ME"
    enabled = "UPDATE ME"
    namespace_filters {
    exclude_all = "UPDATE ME"
    exclude_only = "UPDATE ME"
    include_all = "UPDATE ME"
    include_only = "UPDATE ME"
    }
    tag_filters {
    namespace = "AWS/EC2"
    tags = "UPDATE ME"
    }
    }
    resources_config {
    cloud_security_posture_management_collection = "UPDATE ME"
    extended_collection = "UPDATE ME"
    }
    traces_config {
    xray_services {
    include_all = "UPDATE ME"
    include_only = "UPDATE ME"
    }
    }
}`, uniq)
}

func testAccCheckDatadogAwsAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := AwsAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func AwsAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_aws_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccountv2(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AwsAccount %s", err)}
			}
			return &utils.RetryableError{Prob: "AwsAccount still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogAwsAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := awsAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func awsAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_aws_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccountv2(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving AwsAccount")
		}
	}
	return nil
}
