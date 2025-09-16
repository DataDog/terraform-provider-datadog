package test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestCheckWriteOnlyNotInState verifies that write-only attributes are not stored in state
func TestCheckWriteOnlyNotInState(resourceName, attrName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if _, ok := rs.Primary.Attributes[attrName]; ok {
			return fmt.Errorf("Write-only attribute %s should not be in state", attrName)
		}
		return nil
	}
}

// WriteOnlyBasicTestSteps provides the core write-only test pattern
// Each resource implements its own config generation for flexibility
func WriteOnlyBasicTestSteps(
	configGen func(secret, version, uniq string) string,
	resourceName, writeOnlyAttr, versionAttr string,
	uniq string,
) []resource.TestStep {
	return []resource.TestStep{
		// Create with write-only
		{
			Config: configGen("secret123", "1", uniq),
			Check: resource.ComposeTestCheckFunc(
				TestCheckWriteOnlyNotInState(resourceName, writeOnlyAttr),
				resource.TestCheckResourceAttr(resourceName, versionAttr, "1"),
			),
		},
		// Update version (triggers rotation)
		{
			Config: configGen("newsecret456", "v2.0", uniq),
			Check: resource.ComposeTestCheckFunc(
				TestCheckWriteOnlyNotInState(resourceName, writeOnlyAttr),
				resource.TestCheckResourceAttr(resourceName, versionAttr, "v2.0"),
			),
		},
		// Same version (no rotation expected)
		{
			Config: configGen("differentsecret", "v2.0", uniq),
			Check: resource.ComposeTestCheckFunc(
				TestCheckWriteOnlyNotInState(resourceName, writeOnlyAttr),
				resource.TestCheckResourceAttr(resourceName, versionAttr, "v2.0"),
			),
		},
	}
}
