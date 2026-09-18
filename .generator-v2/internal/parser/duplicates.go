package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// OperationLocation identifies one operation participating in an artifact_name
// collision.
type OperationLocation struct {
	Path        string
	Method      string
	OperationId string
}

// ArtifactNameCollision records every operation that declared a given
// (kind, artifact_name) pair, when more than one did. Uniqueness is scoped per
// kind: Terraform resources and data sources occupy separate namespaces, so a
// datadog_team resource and a datadog_team data source may coexist.
type ArtifactNameCollision struct {
	Kind    model.ArtifactKind
	Name    string
	Sources []OperationLocation
}

// DuplicateArtifactNameError aggregates every (kind, artifact_name) collision
// found across the spec into a single error, naming all source locations for
// each. Collisions are sorted by (name, kind) and their sources by
// (path, method), so the message is deterministic regardless of input order.
type DuplicateArtifactNameError struct {
	Collisions []ArtifactNameCollision
}

func (e *DuplicateArtifactNameError) Error() string {
	var b strings.Builder
	b.WriteString("parser: duplicate artifact_name across operations:")
	for _, c := range e.Collisions {
		fmt.Fprintf(&b, "\n  %s %q declared by:", c.Kind, c.Name)
		for _, s := range c.Sources {
			fmt.Fprintf(&b, "\n    - %s %s (operationId %q)", s.Method, s.Path, s.OperationId)
		}
	}
	return b.String()
}

// CheckDuplicateArtifactNames reports every (kind, artifact_name) pair claimed
// by more than one operation. Uniqueness is scoped per kind — Terraform
// resources and data sources are separate namespaces, so the same name under
// different kinds is allowed. Operations without tracking metadata are ignored
// (they generate no artifact and so cannot collide). It returns a single
// aggregated *DuplicateArtifactNameError naming every collision and all of its
// sources, or nil when every name is unique within its kind.
func CheckDuplicateArtifactNames(spec *model.Spec) error {
	// Key on (kind, name) rather than name alone. Keying on the kind value
	// itself — not an `== "resource"` special case — keeps this correct if new
	// artifact kinds are ever introduced.
	type artifactKey struct {
		Kind model.ArtifactKind
		Name string
	}
	byKey := make(map[artifactKey][]OperationLocation)
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil {
			continue
		}
		key := artifactKey{Kind: op.Tracking.ArtifactKind, Name: op.Tracking.ArtifactName}
		byKey[key] = append(byKey[key], OperationLocation{
			Path:        op.Path,
			Method:      op.Method,
			OperationId: op.OperationId,
		})
	}

	var collisions []ArtifactNameCollision
	for key, locs := range byKey {
		if len(locs) < 2 {
			continue
		}
		sort.Slice(locs, func(i, j int) bool {
			if locs[i].Path != locs[j].Path {
				return locs[i].Path < locs[j].Path
			}
			return locs[i].Method < locs[j].Method
		})
		collisions = append(collisions, ArtifactNameCollision{Kind: key.Kind, Name: key.Name, Sources: locs})
	}
	if len(collisions) == 0 {
		return nil
	}
	sort.Slice(collisions, func(i, j int) bool {
		if collisions[i].Name != collisions[j].Name {
			return collisions[i].Name < collisions[j].Name
		}
		return collisions[i].Kind < collisions[j].Kind
	})
	return &DuplicateArtifactNameError{Collisions: collisions}
}
