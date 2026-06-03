package model

import (
	"regexp"
	"strings"
	"unicode"
)

// Mirroring the generator's own rule is what makes the names we emit match
// datadog-api-client-go, the SDK uses naive PascalCase with
// no acronym uppercasing ("org_id" → "OrgId","url" → "Url", "uuid" → "Uuid").
var (
	patternLeadingAlpha     = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	patternFollowingAlpha   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	patternWhitespace       = regexp.MustCompile(`\W`)
	patternDoubleUnderscore = regexp.MustCompile(`__+`)
)

func snakeCase(value string) string {
	value = patternLeadingAlpha.ReplaceAllString(value, "${1}_${2}")
	value = strings.ToLower(patternFollowingAlpha.ReplaceAllString(value, "${1}_${2}"))
	value = patternWhitespace.ReplaceAllString(value, "_")
	value = strings.TrimRight(value, "_")
	return patternDoubleUnderscore.ReplaceAllString(value, "_")
}

// SdkName translates an OpenAPI identifier into the PascalCase form used by
// datadog-api-client-go.
//
// OperationIds in the Datadog spec are already PascalCase and serve as SDK
// method anchors directly; SdkName is for snake_case property and parameter names.
func SdkName(openapiName string) string {
	var b strings.Builder
	for _, part := range strings.Split(snakeCase(openapiName), "_") {
		if part == "" {
			continue
		}
		runes := []rune(part)
		b.WriteRune(unicode.ToUpper(runes[0]))
		b.WriteString(string(runes[1:]))
	}
	return b.String()
}
