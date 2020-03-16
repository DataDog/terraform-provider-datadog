package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
	"testing"
)

func TestAccountAndRoleFromID(t *testing.T) {
	cases := map[string]struct {
		id        string
		accountID string
		roleName  string
		err       error
	}{
		"basic":        {"1234:qwe", "1234", "qwe", nil},
		"underscores":  {"1234:qwe_rty_asd", "1234", "qwe_rty_asd", nil},
		"no delimeter": {"1234", "", "", fmt.Errorf("error extracting account ID and Role name from an Amazon Web Services integration id: 1234")},
	}
	for name, tc := range cases {
		accountID, roleName, err := accountAndRoleFromID(tc.id)

		if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("%s: erros should be '%s', not `%s`", name, tc.err.Error(), err.Error())
		} else if err != nil && tc.err == nil {
			t.Errorf("%s: erros should be nil, not `%s`", name, err.Error())
		} else if err == nil && tc.err != nil {
			t.Errorf("%s: erros should be '%s', not nil", name, tc.err.Error())
		}

		if accountID != tc.accountID {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountID, tc.accountID)
		}
		if roleName != tc.roleName {
			t.Errorf("%s: role name '%s' didn't match `%s`", name, roleName, tc.roleName)
		}
	}
}

const testAccDatadogIntegrationAWSConfig = `
resource "datadog_integration_aws" "account" {
  account_id                       = "1234567888"
  role_name                        = "testacc-datadog-integration-role"
}
`

func TestAccDatadogIntegrationAWS(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkIntegrationAWSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists,
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_id", "1234567888"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"role_name", "testacc-datadog-integration-role"),
				),
			},
		},
	})
}

func checkIntegrationAWSExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	integrations, err := client.GetIntegrationAWS()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		for _, integration := range *integrations {
			if *integration.AccountID == accountId {
				return nil
			}
		}
		return fmt.Errorf("The AWS integration does not exists for account: accountId=%s", accountId)
	}
	return nil
}

func checkIntegrationAWSDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	integrations, err := client.GetIntegrationAWS()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		for _, integration := range *integrations {
			if *integration.AccountID == accountId {
				return fmt.Errorf("The AWS integration still exists for account: accountId=%s", accountId)
			}
		}
	}
	return nil
}
