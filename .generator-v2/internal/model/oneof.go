package model

import (
	"fmt"
	"sort"
	"strings"
)

// OneOfVariantNameCandidates contains the stable OpenAPI-derived identifiers
// available for naming one Terraform oneOf variant block. The resolver considers
// them in field order; StructuralFingerprint is the final fallback.
type OneOfVariantNameCandidates struct {
	DiscriminatorKey string
	RefName          string
	PrimitiveType    string
	PrimitiveFormat  string
	ShaSum           string
}

// ResolveOneOfVariantName returns the stable snake_case Terraform block name for
// a oneOf alternative. Empty or punctuation-only candidates are skipped so the
// next meaningful source in the precedence chain can be used.
func ResolveOneOfVariantName(candidates OneOfVariantNameCandidates) string {
	primitive := candidates.PrimitiveType
	if primitive != "" && candidates.PrimitiveFormat != "" {
		primitive += "_" + candidates.PrimitiveFormat
	}

	for _, candidate := range []string{
		candidates.DiscriminatorKey,
		candidates.RefName,
		primitive,
	} {
		if name := normalizeOneOfVariantName(candidate); name != "" {
			return name
		}
	}

	fallback := "variant"
	if candidates.ShaSum != "" {
		fallback += "_" + candidates.ShaSum
	}
	// "variant" makes this non-empty even when a caller supplies an unusable
	// fingerprint, but normalize it for consistency with every other candidate.
	if name := normalizeOneOfVariantName(fallback); name != "" {
		return name
	}
	return "variant"
}

func normalizeOneOfVariantName(candidate string) string {
	name := strings.Trim(SnakeCase(candidate), "_")
	if name == "" {
		return ""
	}
	if name[0] >= '0' && name[0] <= '9' {
		return "variant_" + name
	}
	return name
}

// OneOfVariantNameCollisionError reports alternatives that normalize to the
// same Terraform block name. Name is deterministic even if the alternatives
// arrive in a different source order.
type OneOfVariantNameCollisionError struct {
	Path string
	Name string
}

func (e *OneOfVariantNameCollisionError) Error() string {
	return fmt.Sprintf("oneOf at %q has colliding variant name %q", e.Path, e.Name)
}

// ValidateOneOfVariantNames rejects post-normalization name collisions. If more
// than one name collides, the lexicographically first duplicate is reported so
// diagnostics do not depend on OpenAPI alternative order.
func ValidateOneOfVariantNames(path string, variants []OneOfVariant) error {
	counts := make(map[string]int, len(variants))
	for _, variant := range variants {
		counts[variant.TFName]++
	}

	var duplicates []string
	for name, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, name)
		}
	}
	if len(duplicates) == 0 {
		return nil
	}
	sort.Strings(duplicates)
	return &OneOfVariantNameCollisionError{Path: path, Name: duplicates[0]}
}
