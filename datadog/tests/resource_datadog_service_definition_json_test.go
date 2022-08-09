package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func TestAccDatadogServiceDefinitionJSONBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))
	uniqUpdated := fmt.Sprintf("%s-updated", uniq)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogServiceDefinitionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogServiceDefinitionJSON(uniq),
				Check:  checkServiceDefinitionExists(accProvider),
			},
			{
				Config: testAccCheckDatadogServiceDefinitionJSON(uniqUpdated),
				Check:  checkServiceDefinitionExists(accProvider),
			},
		},
	})
}

func testAccCheckDatadogServiceDefinitionJSON(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_service_definition_json" "service_definition_json" {
  definition =<<EOF
{
    "schema-version": "v2",
    "dd-service": "%s",
    "team": "Team A",
    "contacts": [],
    "repos": [],
    "tags": [],
    "integrations": {},
    "dd-team": "team-a",
    "docs": [],
    "extensions": {},
    "links": []
}
EOF
}`, uniq)

}

func checkServiceDefinitionExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			err := utils.Retry(200*time.Millisecond, 4, func() error {
				if _, _, err := utils.SendRequest(authV1, datadogClientV1, "GET", "/api/v2/services/definitions/"+r.Primary.ID, nil); err != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving service %s", err)}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccCheckDatadogServiceDefinitionDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			err := utils.Retry(200*time.Millisecond, 4, func() error {
				if _, httpResp, err := utils.SendRequest(authV1, datadogClientV1, "GET", "/api/v2/services/definitions/"+r.Primary.ID, nil); err != nil {
					if httpResp != nil && httpResp.StatusCode != 404 {
						return &utils.RetryableError{Prob: "service still exists"}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
