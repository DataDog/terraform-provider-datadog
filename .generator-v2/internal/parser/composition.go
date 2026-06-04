package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// AllOfCollision records a single property that two or more allOf branches
// declared with irreconcilable constraints (incompatible type, or the same
// type with differing format/enum). Branches lists the contributing branch
// refs — "#/components/schemas/<name>" for a $ref branch, or "allOf[i]" for an
// inline one — so the offending composition can be located.
type AllOfCollision struct {
	Property string
	Branches []string
	Detail   string
}

// AllOfCollisionError aggregates every irreconcilable property collision found
// while flattening one allOf, sorted by property name so the message is
// deterministic regardless of branch declaration order. It is returned by
// Flatten and is errors.As-discoverable, mirroring *DuplicateArtifactNameError.
type AllOfCollisionError struct {
	Collisions []AllOfCollision
}

func (e *AllOfCollisionError) Error() string {
	var b strings.Builder
	b.WriteString("parser: incompatible allOf branches:")
	for _, c := range e.Collisions {
		fmt.Fprintf(&b, "\n  property %q: %s (declared by %s)", c.Property, c.Detail, strings.Join(c.Branches, ", "))
	}
	return b.String()
}

// Flatten merges the allOf branches of s into a single object model.Schema:
// properties are the union of all branches, requiredness is OR'd across them,
// and nested allOf is resolved recursively (bounded by WithMaxDepth, default
// DefaultMaxDepth). Properties redeclared with incompatible type/format/enum
// surface as an *AllOfCollisionError.
//
// NOT IMPLEMENTED — this is a placeholder so the package compiles. The behavior
// contract lives in composition_ginkgo_test.go (APIR-2906); replace this body
// with the real flattener.
func Flatten(s *base.Schema, opts ...Option) (*model.Schema, error) {
	return nil, errors.New("parser: Flatten not implemented")
}
