package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogCurrentUserDatasource_basic(t *testing.T) {
	// TODO: make the CI use "Terraform User"
	if !isRecording() && !isReplaying() {
		t.Skip("datadog_current_user depends on the caller's own credentials and can't be asserted against in a live CI run")
	}
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	check := func(attr, value string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr("data.datadog_current_user.test", attr, value)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "datadog_current_user" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					check("id", "0bd557d2-6ed0-423e-9654-c0e1cd376abc"),
					check("email", "tf-testacccurrentuserdatasource-local@example.com"),
					check("handle", "tf-testacccurrentuserdatasource-local@example.com"),
					check("name", "Terraform User"),
					check("service_account", "false"),
					check("org_id", "4dee724d-00cc-11ea-a77b-570c9d03c6c5"),
					check("org_public_id", "fasjyydbcgwwc2uc"),
					check("org_name", "DD Integration Tests (321813)"),
				),
			},
		},
	})
}
