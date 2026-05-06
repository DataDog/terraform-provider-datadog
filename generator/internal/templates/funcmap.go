package templates

import (
	"strings"
	"text/template"
)

// FuncMap returns the template function map for use in Go templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"snakeCase":           ToSnakeCase,
		"camelCase":           ToCamelCase,
		"pascalCase":          ToPascalCase,
		"sanitizeDescription": SanitizeDescription,
		"goKeyword":           EscapeGoKeyword,
		"add":                 func(a, b int) int { return a + b },
		"sub":                 func(a, b int) int { return a - b },
		"join":                strings.Join,
	}
}
