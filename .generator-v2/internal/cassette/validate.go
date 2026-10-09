package cassette

import (
	"fmt"
	"net/mail"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Scenario validation
//
// Materialization answers whether a description's examples can be assembled
// into a complete value. It does not ask whether the value they assemble is
// one the description itself would accept, and those are different questions:
// an example naming an enum member that does not exist assembles perfectly and
// is still wrong.
//
// That gap matters because of where it surfaces. The generated test is
// committed and replays green against a cassette recorded from the same
// values, so a value the API rejects is only discovered by a recording run
// against a real org — the one step this feature cannot do for you. Checking
// the values against the schemas they came from moves the discovery to
// generate time, where the fix is a change to the description.
//
// Nothing here re-derives structure. Materialization already walked the schema
// and left MaterializedValue paths behind; validation re-reads those paths and
// asks one question per leaf.
//
// Every check is one-sided on purpose. A leaf whose path does not resolve, or
// whose constraint this package cannot evaluate, is passed rather than
// reported. A false violation blocks an artifact whose description was fine
// and trains a reader to ignore the diagnostic; a missed one leaves the
// recording run to catch it, which is where it was caught before.
// ----------------------------------------------------------------------------

// Violation is one materialized value its own schema contradicts.
//
// It carries the path and a reason, never the value. A violating value is
// still a value the description declared, which may be a credential, and a
// diagnostic is the one place it must never appear.
type Violation struct {
	// Path is the dotted API path, e.g. data.attributes.name.
	Path string
	// Reason names the rule in reviewer-facing terms.
	Reason string
}

// ConformanceError reports a set whose materialized values contradict the
// schemas they were built from. Every violation is listed rather than only the
// first, because the remedy is a change to the description and the author
// should see the whole of it at once.
type ConformanceError struct {
	Key        SetKey
	Violations []Violation
}

func (e *ConformanceError) Error() string {
	parts := make([]string, 0, len(e.Violations))
	for _, v := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s %s", v.Path, v.Reason))
	}
	return fmt.Sprintf("%s does not conform to its schema: %s", e.Key, strings.Join(parts, "; "))
}

// ValidateSet checks one materialized set's values against the schema they
// were built from, reporting every leaf the schema contradicts.
//
// Three kinds of leaf are skipped, each because checking it would describe
// something other than the description's own claim:
//
//   - a synthesized leaf, which is schema-valid by construction and already
//     reported on SynthesizedPaths;
//   - a replaced secret, which holds a safe stand-in rather than the declared
//     value, so checking it checks the replacement;
//   - a declared null, since the normalized model does not record nullability
//     and a null would otherwise read as a type violation everywhere.
func ValidateSet(key SetKey, set MaterializedSet, schema *model.Schema) error {
	if schema == nil {
		return nil
	}

	var violations []Violation
	for _, value := range set.Values {
		if value.Sensitive || value.Value == nil {
			continue
		}
		if slices.Contains(set.SynthesizedPaths, value.Path) {
			continue
		}
		if _, replaced := set.SensitiveReplacements[value.Path]; replaced {
			continue
		}
		leaf := value.Schema
		ok := leaf != nil
		if !ok {
			leaf, ok = resolvePath(schema, value.Path)
		}
		if !ok {
			// Materialization's business, not validation's: reporting it here
			// would duplicate an upstream failure with a worse message.
			continue
		}
		if reason, bad := leafViolation(leaf, value.Value); bad {
			violations = append(violations, Violation{Path: value.Path, Reason: reason})
		}
	}

	if len(violations) == 0 {
		return nil
	}
	return &ConformanceError{Key: key, Violations: violations}
}

// leafViolation reports why one value contradicts its schema, or false when it
// does not. Checks run cheapest-first and stop at the first contradiction: a
// value failing its type has nothing useful to say about its format.
func leafViolation(schema *model.Schema, value any) (string, bool) {
	if schema.Kind == model.SchemaKindUnsupported {
		// The normalizer already decided this node has no representation and
		// recorded why. Its reason is the only actionable part, so it is
		// carried through verbatim.
		reason := schema.UnsupportedReason
		if reason == "" {
			reason = "has no representable type or structure"
		}
		return "stands on a schema the normalizer could not represent: " + reason, true
	}
	if reason, bad := typeViolation(schema, value); bad {
		return reason, true
	}
	if reason, bad := enumViolation(schema, value); bad {
		return reason, true
	}
	return formatViolation(schema, value)
}

// typeViolation compares the value's Go type against the schema's declared
// type. JSON decoding and synthesis produce different numeric types for the
// same declaration, so every numeric form is accepted for a numeric schema;
// only a genuine category error is reported.
func typeViolation(schema *model.Schema, value any) (string, bool) {
	switch schema.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Sprintf("is declared a string but materialized as %T", value), true
		}
	case "integer":
		if !isNumeric(value) {
			return fmt.Sprintf("is declared an integer but materialized as %T", value), true
		}
		if !isIntegral(value) {
			return "is declared an integer but materialized with a fractional part", true
		}
	case "number":
		if !isNumeric(value) {
			return fmt.Sprintf("is declared a number but materialized as %T", value), true
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Sprintf("is declared a boolean but materialized as %T", value), true
		}
	}
	return "", false
}

// enumViolation reports a value outside its schema's enumeration. Enum members
// are normalized to strings, so the comparison is on the value's string form;
// the members are named because they come from the description and are what the
// author has to choose between. The value itself is never named.
func enumViolation(schema *model.Schema, value any) (string, bool) {
	if len(schema.Enum) == 0 {
		return "", false
	}
	if slices.Contains(schema.Enum, fmt.Sprint(value)) {
		return "", false
	}
	return fmt.Sprintf("is not one of the declared enum members %s", joinAnd(schema.Enum)), true
}

// formatViolation checks the formats this package can evaluate and passes
// every other one. A format the normalized model does not know how to check
// must not become a refusal: the generator would reject descriptions for using
// a vocabulary it has not learned.
func formatViolation(schema *model.Schema, value any) (string, bool) {
	text, ok := value.(string)
	if !ok || schema.Format == "" {
		return "", false
	}
	valid := true
	switch schema.Format {
	case "date-time":
		_, err := time.Parse(time.RFC3339, text)
		valid = err == nil
	case "date":
		_, err := time.Parse("2006-01-02", text)
		valid = err == nil
	case "uuid":
		valid = isUUID(text)
	case "email":
		_, err := mail.ParseAddress(text)
		valid = err == nil
	case "uri", "url":
		parsed, err := url.Parse(text)
		valid = err == nil && parsed.Scheme != "" && parsed.Host != ""
	default:
		return "", false
	}
	if valid {
		return "", false
	}
	return fmt.Sprintf("is not a valid %s", schema.Format), true
}

// isUUID reports the 8-4-4-4-12 hex shape, without checking version or
// variant: a description declaring format uuid is claiming the shape, and the
// generator is in no position to insist on more.
func isUUID(text string) bool {
	groups := strings.Split(text, "-")
	if len(groups) != 5 {
		return false
	}
	for i, want := range []int{8, 4, 4, 4, 12} {
		if len(groups[i]) != want {
			return false
		}
		if _, err := strconv.ParseUint(groups[i], 16, 64); err != nil {
			return false
		}
	}
	return true
}

func isNumeric(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	return false
}

// isIntegral reports whether a numeric value carries no fractional part. A
// float is how JSON decoding represents every number, including the integers,
// so an integer schema cannot simply reject one.
func isIntegral(value any) bool {
	switch typed := value.(type) {
	case float32:
		return float64(typed) == float64(int64(typed))
	case float64:
		return typed == float64(int64(typed))
	default:
		return true
	}
}

// ----------------------------------------------------------------------------
// Path resolution
// ----------------------------------------------------------------------------

// resolvePath walks a materialized value's path back to the schema node it
// came from, reporting false when the path names nothing the schema describes.
//
// It understands the three shapes materialization emits: dotted object
// properties, "[n]" array indices, and "(root)" for a value at the top. A
// union is searched variant by variant, since a path may address a property
// only one alternative declares.
func resolvePath(schema *model.Schema, path string) (*model.Schema, bool) {
	if schema == nil {
		return nil, false
	}
	if path == "" || path == "(root)" {
		return schema, true
	}
	current := schema
	for _, segment := range strings.Split(path, ".") {
		name, indices := splitIndices(segment)
		if name != "" {
			next, ok := member(current, name)
			if !ok {
				return nil, false
			}
			current = next
		}
		for range indices {
			if current.Items == nil {
				return nil, false
			}
			current = current.Items
		}
	}
	return current, true
}

// member resolves one named step: an object's property, a map's value schema,
// or the same step inside any union alternative that declares it.
func member(schema *model.Schema, name string) (*model.Schema, bool) {
	if schema == nil {
		return nil, false
	}
	switch schema.Kind {
	case model.SchemaKindObject:
		if property, ok := schema.Properties[name]; ok && property != nil {
			return property, true
		}
		return nil, false
	case model.SchemaKindMap:
		// Dynamic keys share one value schema, so any name resolves to it.
		if schema.Items == nil {
			return nil, false
		}
		return schema.Items, true
	case model.SchemaKindOneOf:
		for _, variant := range oneOfVariants(schema) {
			if resolved, ok := member(variant, name); ok {
				return resolved, true
			}
		}
		return nil, false
	default:
		// A property declared on a node with no properties resolves to
		// nothing, which is a skip rather than a violation.
		if property, ok := schema.Properties[name]; ok && property != nil {
			return property, true
		}
		return nil, false
	}
}

// splitIndices separates a path segment's name from its trailing array
// indices, so "tags[0][1]" reads as the name "tags" and two descents.
func splitIndices(segment string) (string, []string) {
	open := strings.Index(segment, "[")
	if open < 0 {
		return segment, nil
	}
	name := segment[:open]
	var indices []string
	for _, part := range strings.Split(segment[open:], "[") {
		if part == "" {
			continue
		}
		indices = append(indices, strings.TrimSuffix(part, "]"))
	}
	return name, indices
}
