package datadog

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DeprecatedDocumentation(description string, replacedResource *string) string {
	if replacedResource != nil {
		return fmt.Sprintf("%s%s", WarningCallout(fmt.Sprintf("This resource is deprecated - use the `%s` resource instead.", *replacedResource)), description)
	}
	return fmt.Sprintf("%s%s", WarningCallout("This resource is deprecated"), description)
}

func WarningCallout(message string) string {
	return fmt.Sprintf("!>%s\n\n", message)
}

func Ptr[T any](v T) *T {
	return &v
}

func resourceDiffRawConfigValue(diff *schema.ResourceDiff, attribute string) (cty.Value, bool) {
	value, diags := diff.GetRawConfigAt(cty.GetAttrPath(attribute))
	if diags.HasError() || !value.IsKnown() || value.IsNull() {
		return cty.NilVal, false
	}
	return value, true
}

func isResourceDiffAttributeConfigured(diff *schema.ResourceDiff, attribute string) bool {
	_, configured := resourceDiffRawConfigValue(diff, attribute)
	return configured
}

func isResourceDiffOptionalBoolFalse(diff *schema.ResourceDiff, attribute string) bool {
	value, configured := resourceDiffRawConfigValue(diff, attribute)
	return configured && value.Type().Equals(cty.Bool) && value.False()
}
