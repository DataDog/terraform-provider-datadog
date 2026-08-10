package parser

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("CheckDuplicateArtifactNames", func() {
	trackedOp := func(path, method, operationId, artifactName string) *model.Operation {
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

	kindedOp := func(path, method, operationId, artifactName string, kind model.ArtifactKind) *model.Operation {
		op := trackedOp(path, method, operationId, artifactName)
		op.Tracking.ArtifactKind = kind
		return op
	}

	asDuplicateErr := func(err error) *DuplicateArtifactNameError {
		GinkgoHelper()
		var dup *DuplicateArtifactNameError
		Expect(errors.As(err, &dup)).To(BeTrue(), "error %v (%T) is not a *DuplicateArtifactNameError", err, err)
		return dup
	}

	Context("when every artifact_name is unique within its kind", func() {
		It("returns no error", func() {
			spec := &model.Spec{Operations: []*model.Operation{
				trackedOp("/a", "GET", "GetA", "alpha"),
				trackedOp("/b", "GET", "GetB", "beta"),
				trackedOp("/c", "GET", "GetC", "gamma"),
			}}
			Expect(CheckDuplicateArtifactNames(spec)).To(Succeed())
		})

		It("ignores operations without tracking metadata", func() {
			spec := &model.Spec{Operations: []*model.Operation{
				trackedOp("/a", "GET", "GetA", "alpha"),
				{Path: "/health", Method: "GET", OperationId: "GetHealth"},
				nil,
				trackedOp("/b", "GET", "GetB", "beta"),
			}}
			Expect(CheckDuplicateArtifactNames(spec)).To(Succeed())
		})
	})

	Context("when the same artifact_name appears under different kinds", func() {
		It("does not report a collision", func() {
			spec := &model.Spec{Operations: []*model.Operation{
				kindedOp("/teams", "POST", "CreateTeam", "team", model.ArtifactKindResource),
				kindedOp("/teams/{id}", "GET", "GetTeam", "team", model.ArtifactKindDataSource),
			}}
			Expect(CheckDuplicateArtifactNames(spec)).To(Succeed(),
				"a resource and a data source named %q must be allowed", "team")
		})
	})

	Context("when two operations share an artifact_name within a kind", func() {
		It("returns a single error naming the kind, the name, and both source locations", func() {
			spec := &model.Spec{Operations: []*model.Operation{
				trackedOp("/teams", "GET", "ListTeams", "team"),
				trackedOp("/teams/{id}", "GET", "GetTeam", "team"),
			}}
			dup := asDuplicateErr(CheckDuplicateArtifactNames(spec))
			Expect(dup.Collisions).To(HaveLen(1))
			Expect(dup.Collisions[0].Sources).To(HaveLen(2))
			Expect(dup.Error()).To(SatisfyAll(
				ContainSubstring(string(model.ArtifactKindResource)),
				ContainSubstring("team"),
				ContainSubstring("ListTeams"),
				ContainSubstring("GetTeam"),
				ContainSubstring("/teams"),
				ContainSubstring("/teams/{id}"),
			))
		})

		It("lists every colliding source, not just the first two", func() {
			spec := &model.Spec{Operations: []*model.Operation{
				trackedOp("/teams", "GET", "ListTeams", "team"),
				trackedOp("/teams/{id}", "GET", "GetTeam", "team"),
				trackedOp("/teams/search", "POST", "SearchTeams", "team"),
			}}
			dup := asDuplicateErr(CheckDuplicateArtifactNames(spec))
			Expect(dup.Collisions[0].Sources).To(HaveLen(3))
			Expect(dup.Error()).To(SatisfyAll(
				ContainSubstring("ListTeams"),
				ContainSubstring("GetTeam"),
				ContainSubstring("SearchTeams"),
			))
		})
	})

	Context("when several distinct artifact_names each collide", func() {
		It("reports every collision sorted by name", func() {
			spec := &model.Spec{Operations: []*model.Operation{
				trackedOp("/z", "GET", "GetZ1", "zeta"),
				trackedOp("/z2", "GET", "GetZ2", "zeta"),
				trackedOp("/a", "GET", "GetA1", "alpha"),
				trackedOp("/a2", "GET", "GetA2", "alpha"),
			}}
			dup := asDuplicateErr(CheckDuplicateArtifactNames(spec))
			Expect(dup.Collisions).To(HaveLen(2))
			Expect(dup.Collisions[0].Name).To(Equal("alpha"))
			Expect(dup.Collisions[1].Name).To(Equal("zeta"))
		})

		It("produces identical output regardless of declaration order", func() {
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
			Expect(errA).To(HaveOccurred())
			Expect(errB).To(HaveOccurred())
			Expect(errA.Error()).To(Equal(errB.Error()))
		})
	})
})
