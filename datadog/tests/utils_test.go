package test

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func checkRessourceAttributeRegex(name string, key string, pattern string) func(*terraform.State) error {
	return resource.TestCheckResourceAttrWith(name, key, func(attr string) error {
		ok, _ := regexp.MatchString(pattern, attr)
		if !ok {
			return fmt.Errorf("expected %s to match %s", attr, pattern)
		}
		return nil
	})
}
