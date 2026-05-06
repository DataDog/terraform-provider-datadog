package templates

import (
	"strings"
	"unicode"
)

// Common acronyms that should remain uppercase in PascalCase or be treated as single units.
var acronyms = map[string]bool{
	"API": true, "ID": true, "URL": true, "HTTP": true, "HTTPS": true,
	"JSON": true, "XML": true, "HTML": true, "CSS": true, "SQL": true,
	"IP": true, "TCP": true, "UDP": true, "DNS": true, "SSH": true,
	"CPU": true, "GPU": true, "RAM": true, "OS": true, "UI": true,
	"UUID": true, "URI": true, "SDK": true, "CLI": true, "TLS": true,
}

// Go reserved words.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(s string) string {
	var result []rune
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					result = append(result, '_')
				} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					result = append(result, '_')
				}
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// ToCamelCase converts a snake_case or PascalCase string to camelCase.
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if pascal == "" {
		return ""
	}
	// Check if the string starts with an acronym
	for acr := range acronyms {
		if strings.HasPrefix(pascal, acr) {
			return strings.ToLower(acr) + pascal[len(acr):]
		}
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ToPascalCase converts a snake_case or camelCase string to PascalCase.
func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// Split on underscores, hyphens, and spaces
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		upper := strings.ToUpper(part)
		if acronyms[upper] {
			result.WriteString(upper)
			continue
		}
		// Also handle camelCase splits within parts
		subParts := splitCamelCase(part)
		for _, sp := range subParts {
			upper = strings.ToUpper(sp)
			if acronyms[upper] {
				result.WriteString(upper)
			} else {
				runes := []rune(sp)
				runes[0] = unicode.ToUpper(runes[0])
				for i := 1; i < len(runes); i++ {
					runes[i] = unicode.ToLower(runes[i])
				}
				result.WriteString(string(runes))
			}
		}
	}
	return result.String()
}

// ToSDKPascalCase converts a snake_case string to PascalCase using Datadog SDK
// casing conventions. The SDK uses naive PascalCase: capitalize the first letter
// of each word segment and lowercase the rest. It does NOT uppercase acronyms
// like standard Go conventions (e.g., OrgId not OrgID, TeamUrl not TeamURL,
// ApiKey not APIKey).
func ToSDKPascalCase(s string) string {
	if s == "" {
		return ""
	}

	// Normalize: convert to snake_case first to handle camelCase inputs,
	// then split on delimiters. This ensures "createdBy" → "created_by" → "CreatedBy".
	normalized := ToSnakeCase(s)

	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		for i := 1; i < len(runes); i++ {
			runes[i] = unicode.ToLower(runes[i])
		}
		result.WriteString(string(runes))
	}
	return result.String()
}

// splitCamelCase splits a camelCase or PascalCase string into words.
func splitCamelCase(s string) []string {
	var parts []string
	runes := []rune(s)
	start := 0
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) {
			if unicode.IsLower(runes[i-1]) {
				parts = append(parts, string(runes[start:i]))
				start = i
			} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				parts = append(parts, string(runes[start:i]))
				start = i
			}
		}
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

// EscapeGoKeyword appends an underscore to Go reserved words.
func EscapeGoKeyword(s string) string {
	if goKeywords[s] {
		return s + "_"
	}
	return s
}

// SanitizeDescription escapes a string for use in Go source comments or string literals.
func SanitizeDescription(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "`", "'")
	// Collapse multiple lines into single line for schema descriptions
	lines := strings.Split(s, "\n")
	var trimmed []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			trimmed = append(trimmed, line)
		}
	}
	return strings.Join(trimmed, " ")
}
