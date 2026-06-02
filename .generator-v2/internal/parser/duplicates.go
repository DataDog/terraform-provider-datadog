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
// artifact_name, when more than one did.
type ArtifactNameCollision struct {
	Name    string
	Sources []OperationLocation
}

// DuplicateArtifactNameError aggregates every artifact_name collision found
// across the spec into a single error, naming all source locations for each.
// Collisions are sorted by name and their sources by (path, method), so the
// message is deterministic regardless of input order.
type DuplicateArtifactNameError struct {
	Collisions []ArtifactNameCollision
}

func (e *DuplicateArtifactNameError) Error() string {
	var b strings.Builder
	b.WriteString("parser: duplicate artifact_name across operations:")
	for _, c := range e.Collisions {
		fmt.Fprintf(&b, "\n  %q declared by:", c.Name)
		for _, s := range c.Sources {
			fmt.Fprintf(&b, "\n    - %s %s (operationId %q)", s.Method, s.Path, s.OperationId)
		}
	}
	return b.String()
}

// CheckDuplicateArtifactNames reports every artifact_name claimed by more than
// one operation. Operations without tracking metadata are ignored (they
// generate no artifact and so cannot collide). It returns a single aggregated
// *DuplicateArtifactNameError naming every collision and all of its sources, or
// nil when all names are unique.
func CheckDuplicateArtifactNames(spec *model.Spec) error {
	byName := make(map[string][]OperationLocation)
	for _, op := range spec.Operations {
		if op == nil || op.Tracking == nil {
			continue
		}
		byName[op.Tracking.ArtifactName] = append(byName[op.Tracking.ArtifactName], OperationLocation{
			Path:        op.Path,
			Method:      op.Method,
			OperationId: op.OperationId,
		})
	}

	var collisions []ArtifactNameCollision
	for name, locs := range byName {
		if len(locs) < 2 {
			continue
		}
		sort.Slice(locs, func(i, j int) bool {
			if locs[i].Path != locs[j].Path {
				return locs[i].Path < locs[j].Path
			}
			return locs[i].Method < locs[j].Method
		})
		collisions = append(collisions, ArtifactNameCollision{Name: name, Sources: locs})
	}
	if len(collisions) == 0 {
		return nil
	}
	sort.Slice(collisions, func(i, j int) bool { return collisions[i].Name < collisions[j].Name })
	return &DuplicateArtifactNameError{Collisions: collisions}
}
