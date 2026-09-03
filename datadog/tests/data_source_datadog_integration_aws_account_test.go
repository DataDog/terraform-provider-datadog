package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogIntegrationAwsAccountDatasource(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAwsAccount(accountID, uniq),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.datadog_integration_aws_account.foo", "id",
						"datadog_integration_aws_account.foo", "id"),
					resource.TestCheckResourceAttr(
						"data.datadog_integration_aws_account.foo", "aws_account_id", accountID),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAwsAccount(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
    aws_account_id = "%s"
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
}

data "datadog_integration_aws_account" "foo" {
    aws_account_id = datadog_integration_aws_account.foo.aws_account_id
}`, accountID, uniq)
}
