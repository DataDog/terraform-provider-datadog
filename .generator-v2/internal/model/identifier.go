package model

import (
	"regexp"
	"strings"
	"unicode"
)

// These patterns reproduce the SDK generator's own casing rule: naive
// PascalCase with no acronym uppercasing ("org_id" → "OrgId", "url" → "Url",
// "uuid" → "Uuid").
var (
	patternLeadingAlpha     = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	patternFollowingAlpha   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	patternWhitespace       = regexp.MustCompile(`\W`)
	patternDoubleUnderscore = regexp.MustCompile(`__+`)
	patternNonAlphanumeric  = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

// SnakeCase converts a camelCase or PascalCase identifier to snake_case, a port
// of utils.snake_case from datadog-api-client-go's bundled generator, applying
// the same five operations in the same order.
func SnakeCase(value string) string {
	value = patternLeadingAlpha.ReplaceAllString(value, "${1}_${2}")
	value = strings.ToLower(patternFollowingAlpha.ReplaceAllString(value, "${1}_${2}"))
	value = patternWhitespace.ReplaceAllString(value, "_")
	value = strings.TrimRight(value, "_")
	return patternDoubleUnderscore.ReplaceAllString(value, "_")
}

// SdkClassName ports utils.class_name from datadog-api-client-go's bundled
// generator:
//
//	value = re.sub(r'[^a-zA-Z0-9]', '', value)
//	return value + "Api"
//
// It strips non-alphanumerics and does not re-capitalize, relying on tags
// already being PascalCase: "Org Groups" → "OrgGroupsApi", "APM" → "APMApi",
// while "org groups" → "orggroupsApi" — capitalizing that would name a Go type
// the SDK never generated.
func SdkClassName(tag string) string {
	return patternNonAlphanumeric.ReplaceAllString(tag, "") + "Api"
}

// goKeywords is Go's full reserved-word set, the same list the SDK generator
// carries as formatter.KEYWORDS.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// EscapeReservedKeyword appends "Var" to a Go reserved word and returns anything
// else unchanged, a port of the SDK generator's
// formatter.escape_reserved_keyword — how the SDK keeps a property named "type"
// from producing the local `type`. It only ever fires on a lower-camel
// identifier: no Go keyword starts with an upper-case rune.
func EscapeReservedKeyword(word string) string {
	if goKeywords[word] {
		return word + "Var"
	}
	return word
}

// UpperFirst upper-cases the first rune and leaves the rest untouched, a port of
// the SDK generator's utils.upperfirst. Not a PascalCase conversion: it applies
// to an already-formed name, so "AWSIntegration" and "string" come back as
// "AWSIntegration" and "String".
func UpperFirst(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	return string(unicode.ToUpper(runes[0])) + string(runes[1:])
}

// SdkName translates an OpenAPI identifier into the PascalCase form used by
// datadog-api-client-go, a port of the SDK generator's utils.camel_case
// (snake_case, then upperfirst each underscore-delimited part). It is for
// snake_case property and parameter names; operationIds are already PascalCase.
func SdkName(openapiName string) string {
	var b strings.Builder
	for _, part := range strings.Split(SnakeCase(openapiName), "_") {
		if part == "" {
			continue
		}
		runes := []rune(part)
		b.WriteRune(unicode.ToUpper(runes[0]))
		b.WriteString(string(runes[1:]))
	}
	return b.String()
}
