package datadog

import "fmt"

func DeprecatedDocumentation(description string, replacedResource *string) string {
	if replacedResource != nil {
		return fmt.Sprintf("!>This resource is deprecated - use the `%s` resource instead.\n\n%s", *replacedResource, description)
	} else {
		return fmt.Sprintf("!>This resource is deprecated \n\n %s", description)
	}
}

func Ptr[T any](v T) *T {
	return &v
}
