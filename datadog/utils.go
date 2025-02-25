package datadog

import "fmt"

func DeprecatedDocumentation(description string, replacedResource *string) string {
	if replacedResource != nil {
		return fmt.Sprintf("%s%s", WarningCallout(fmt.Sprintf("This resource is deprecated - use the `%s` resource instead.", *replacedResource)), description)
	} else {
		return fmt.Sprintf("%s%s", WarningCallout("This resource is deprecated"), description)
	}
}

func WarningCallout(message string) string {
	return fmt.Sprintf("!>%s\n\n", message)
}

func Ptr[T any](v T) *T {
	return &v
}
