package parser

import (
	"errors"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// trackedOp builds a model.Operation carrying tracking metadata for the given
// artifact name. The kind is irrelevant to duplicate detection.
func trackedOp(path, method, operationId, artifactName string) *model.Operation {
	return &model.Operation{
		Path:        path,
		Method:      method,
		OperationId: operationId,
		Tracking: &model.TrackingFieldMetadata{
			ArtifactKind: model.ArtifactKindResource,
			ArtifactName: artifactName,
		},
	}
}

func TestCheckDuplicateArtifactNamesUnique(t *testing.T) {
	spec := &model.Spec{Operations: []*model.Operation{
		trackedOp("/a", "GET", "GetA", "alpha"),
		trackedOp("/b", "GET", "GetB", "beta"),
		trackedOp("/c", "GET", "GetC", "gamma"),
	}}
	if err := CheckDuplicateArtifactNames(spec); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheckDuplicateArtifactNamesIgnoresUntracked(t *testing.T) {
	spec := &model.Spec{Operations: []*model.Operation{
		trackedOp("/a", "GET", "GetA", "alpha"),
		{Path: "/health", Method: "GET", OperationId: "GetHealth"}, // Tracking nil
		nil, // defensive: nil entries are skipped
		trackedOp("/b", "GET", "GetB", "beta"),
	}}
	if err := CheckDuplicateArtifactNames(spec); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCheckDuplicateArtifactNamesSingleCollision(t *testing.T) {
	spec := &model.Spec{Operations: []*model.Operation{
		trackedOp("/teams", "GET", "ListTeams", "team"),
		trackedOp("/teams/{id}", "GET", "GetTeam", "team"),
	}}
	err := CheckDuplicateArtifactNames(spec)
	var dup *DuplicateArtifactNameError
	if !errors.As(err, &dup) {
		t.Fatalf("error %v (%T) is not a *DuplicateArtifactNameError", err, err)
	}
	if len(dup.Collisions) != 1 {
		t.Fatalf("got %d collisions, want 1", len(dup.Collisions))
	}
	if len(dup.Collisions[0].Sources) != 2 {
		t.Fatalf("got %d sources, want 2", len(dup.Collisions[0].Sources))
	}
	msg := dup.Error()
	for _, want := range []string{"team", "ListTeams", "GetTeam", "/teams", "/teams/{id}"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing %q:\n%s", want, msg)
		}
	}
}

func TestCheckDuplicateArtifactNamesListsAllSources(t *testing.T) {
	spec := &model.Spec{Operations: []*model.Operation{
		trackedOp("/teams", "GET", "ListTeams", "team"),
		trackedOp("/teams/{id}", "GET", "GetTeam", "team"),
		trackedOp("/teams/search", "POST", "SearchTeams", "team"),
	}}
	err := CheckDuplicateArtifactNames(spec)
	var dup *DuplicateArtifactNameError
	if !errors.As(err, &dup) {
		t.Fatalf("error %v (%T) is not a *DuplicateArtifactNameError", err, err)
	}
	if n := len(dup.Collisions[0].Sources); n != 3 {
		t.Fatalf("got %d sources, want all 3 listed", n)
	}
	for _, want := range []string{"ListTeams", "GetTeam", "SearchTeams"} {
		if !strings.Contains(dup.Error(), want) {
			t.Errorf("error message missing source %q", want)
		}
	}
}

func TestCheckDuplicateArtifactNamesMultipleCollisionsSortedByName(t *testing.T) {
	spec := &model.Spec{Operations: []*model.Operation{
		trackedOp("/z", "GET", "GetZ1", "zeta"),
		trackedOp("/z2", "GET", "GetZ2", "zeta"),
		trackedOp("/a", "GET", "GetA1", "alpha"),
		trackedOp("/a2", "GET", "GetA2", "alpha"),
	}}
	err := CheckDuplicateArtifactNames(spec)
	var dup *DuplicateArtifactNameError
	if !errors.As(err, &dup) {
		t.Fatalf("error %v (%T) is not a *DuplicateArtifactNameError", err, err)
	}
	if len(dup.Collisions) != 2 {
		t.Fatalf("got %d collisions, want 2", len(dup.Collisions))
	}
	if dup.Collisions[0].Name != "alpha" || dup.Collisions[1].Name != "zeta" {
		t.Errorf("collisions not sorted by name: %q then %q", dup.Collisions[0].Name, dup.Collisions[1].Name)
	}
}

func TestCheckDuplicateArtifactNamesDeterministic(t *testing.T) {
	// Same collisions, different declaration orders, must yield identical output.
	build := func(ops ...*model.Operation) *model.Spec { return &model.Spec{Operations: ops} }
	a := build(
		trackedOp("/teams", "GET", "ListTeams", "team"),
		trackedOp("/teams/{id}", "GET", "GetTeam", "team"),
		trackedOp("/users", "GET", "ListUsers", "user"),
		trackedOp("/users/{id}", "GET", "GetUser", "user"),
	)
	b := build(
		trackedOp("/users/{id}", "GET", "GetUser", "user"),
		trackedOp("/teams/{id}", "GET", "GetTeam", "team"),
		trackedOp("/users", "GET", "ListUsers", "user"),
		trackedOp("/teams", "GET", "ListTeams", "team"),
	)
	errA, errB := CheckDuplicateArtifactNames(a), CheckDuplicateArtifactNames(b)
	if errA == nil || errB == nil {
		t.Fatal("expected duplicate errors from both specs")
	}
	if errA.Error() != errB.Error() {
		t.Errorf("non-deterministic output:\nA:\n%s\nB:\n%s", errA.Error(), errB.Error())
	}
}
