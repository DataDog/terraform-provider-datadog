package model

import (
	"fmt"
	"sort"
	"strings"
)

// OneOfVariantNameCandidates contains the stable OpenAPI-derived identifiers
// available for naming one Terraform oneOf variant block. The resolver considers
// them in field order and rejects alternatives with no meaningful candidate.
type OneOfVariantNameCandidates struct {
	DiscriminatorKey string
	RefName          string
	PrimitiveType    string
	PrimitiveFormat  string
}

// OneOfVariantNameResolutionError reports an alternative that has no
// compatibility-stable source for its public Terraform block name.
type OneOfVariantNameResolutionError struct {
	Path        string
	Alternative int
}

func (e *OneOfVariantNameResolutionError) Error() string {
	return fmt.Sprintf(
		"oneOf at %q alternative %d has no stable Terraform variant name; use a discriminator mapping key or a named schema reference",
		e.Path,
		e.Alternative,
	)
}

// ResolveOneOfVariantName returns the stable snake_case Terraform block name for
// a oneOf alternative. Empty or punctuation-only candidates are skipped so the
// next meaningful source in the precedence chain can be used. Anonymous
// non-primitive alternatives fail rather than exposing a content- or
// source-order-derived name as part of the public Terraform schema.
func ResolveOneOfVariantName(path string, alternative int, candidates OneOfVariantNameCandidates) (string, error) {
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
			return name, nil
		}
	}

	return "", &OneOfVariantNameResolutionError{Path: path, Alternative: alternative}
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

// oneOfRoleSuffixes are the CRUD-role markers the Datadog v2 API appends to a
// component name, in both the spellings a name can arrive in: PascalCase for an
// OpenAPI component, snake_case for a Terraform variant block.
var oneOfRoleSuffixes = []string{
	"Response", "_response",
	"Request", "_request",
	"Update", "_update",
	"Create", "_create",
}

// StripOneOfRoleSuffix removes the trailing run of CRUD-role markers from an
// OpenAPI-derived union or alternative name, accepting either the PascalCase
// spelling of a component ("IntegrationAccountBasicAuthRequest") or the
// snake_case spelling of a Terraform variant block
// ("integration_account_basic_auth_request"), and returning the same casing it
// was given.
//
// It exists because one logical union is up to three OpenAPI components — the
// Create body's, the Update body's and the Read response's — and a resource's
// merged schema has to correlate them and then expose exactly one public name
// for the set. Which body won the merge must not be visible in the Terraform
// schema (FR-012e, T099c).
//
// The run is stripped rather than a single suffix because the markers compose:
// downtime spells one alternative "DowntimeScheduleRecurrencesCreateRequest",
// "…RecurrencesUpdateRequest" and "…RecurrencesResponse", which only reduce to
// a common stem once both markers are gone. Stripping stops at whatever it
// cannot remove, so a component actually named "Request" keeps its name.
//
// Stripping is lossy by construction — an alternative legitimately named
// "…Update" reduces the same way — which is why mergeOneOf reaches for it only
// after the bodies have failed to agree on a name by themselves.
func StripOneOfRoleSuffix(name string) string {
stripping:
	for {
		for _, suffix := range oneOfRoleSuffixes {
			if trimmed, ok := strings.CutSuffix(name, suffix); ok && trimmed != "" {
				name = trimmed
				continue stripping
			}
		}
		return name
	}
}
