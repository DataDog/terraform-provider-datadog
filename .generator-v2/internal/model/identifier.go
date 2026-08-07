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

// SnakeCase converts a camelCase or PascalCase identifier to snake_case.
func SnakeCase(value string) string {
	value = patternLeadingAlpha.ReplaceAllString(value, "${1}_${2}")
	value = strings.ToLower(patternFollowingAlpha.ReplaceAllString(value, "${1}_${2}"))
	value = patternWhitespace.ReplaceAllString(value, "_")
	value = strings.TrimRight(value, "_")
	return patternDoubleUnderscore.ReplaceAllString(value, "_")
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
// else unchanged — a port of the SDK generator's
// formatter.escape_reserved_keyword, which is how the SDK itself keeps a property
// named "type" from producing the local `type`.
//
// Reusing the SDK's suffix rather than inventing one means a maintainer who knows
// the SDK reads generated locals the same way in both codebases. It only ever
// fires on a lower-camel identifier: an exported field name starts with an
// upper-case rune, and no Go keyword does.
func EscapeReservedKeyword(word string) string {
	if goKeywords[word] {
		return word + "Var"
	}
	return word
}

// UpperFirst upper-cases the first rune and leaves the rest untouched, a port of
// the SDK generator's utils.upperfirst. It is not a PascalCase conversion: the
// SDK applies it to an already-formed name (a component name, or a Go type
// spelling) when deriving a oneOf wrapper member, so "AWSIntegration" and
// "string" must come back as "AWSIntegration" and "String" respectively.
func UpperFirst(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	return string(unicode.ToUpper(runes[0])) + string(runes[1:])
}

// SdkName translates an OpenAPI identifier into the PascalCase form used by
// datadog-api-client-go. It is the port of the SDK generator's
// utils.camel_case (snake_case, then upperfirst each underscore-delimited part).
//
// OperationIds in the Datadog spec are already PascalCase and serve as SDK
// method anchors directly; SdkName is for snake_case property and parameter names.
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
