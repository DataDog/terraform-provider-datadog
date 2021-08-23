package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDatadogOrganizationSettings_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_organization_settings.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueEntity := uniqueEntityName(ctx, t)
	organizationName := fmt.Sprint(uniqueEntity[len(uniqueEntity)-30:])

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogOrganizationSettingsConfig_Required(organizationName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
