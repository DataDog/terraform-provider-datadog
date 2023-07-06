package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDatadogUser_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_user.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	username := strings.ToLower(uniqueEntityName(ctx, t)) + "@example.com"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogUserV2Destroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogUserConfigImported(username),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_invitation_id", "send_user_invitation"},
			},
		},
	})
}

func testAccCheckDatadogUserConfigImported(uniq string) string {
	return fmt.Sprintf(`%s

resource "datadog_user" "foo" {
  email  = "%s"
  name   = "Test User"
  roles  = [data.datadog_role.st_role.id, data.datadog_role.adm_role.id]
}`, roleDatasources, uniq)
}
