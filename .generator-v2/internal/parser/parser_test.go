package parser

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// scrambledSpecPath is a spec whose paths and methods are declared out of
// (path, method) order on purpose, so ordering tests fail if LoadSpec stops sorting.
var scrambledSpecPath = filepath.Join("../testdata/parser", "scrambled_ordering.yaml")

func operationOrder(s *model.Spec) []string {
	out := make([]string, len(s.Operations))
	for i, op := range s.Operations {
		out[i] = op.Method + " " + op.Path
	}
	return out
}

var _ = Describe("LoadSpec", func() {

	Context("operation ordering", func() {
		It("sorts operations by path then method regardless of source order", func() {
			spec, err := LoadSpec(scrambledSpecPath)
			Expect(err).To(Succeed())
			Expect(operationOrder(spec)).To(Equal([]string{
				"DELETE /alpha",
				"GET /alpha",
				"POST /alpha",
				"PUT /mid",
				"GET /zebra",
			}))
		})

		It("returns the same order across repeated loads", func() {
			first, err := LoadSpec(scrambledSpecPath)
			Expect(err).To(Succeed())
			baseline := operationOrder(first)
			Expect(baseline).NotTo(BeEmpty())

			for i := range 5 {
				got, err := LoadSpec(scrambledSpecPath)
				Expect(err).To(Succeed(), "run %d", i)
				Expect(operationOrder(got)).To(Equal(baseline), "run %d diverged", i)
			}
		})
	})

	Context("cycle detection", func() {
		It("returns a *RefCycleError for a self-referential $ref", func() {
			_, err := LoadSpec(filepath.Join("../testdata/parser", "cycle_self.yaml"))
			Expect(err).To(HaveOccurred())

			var cycleErr *RefCycleError
			Expect(errors.As(err, &cycleErr)).To(BeTrue(), "got %T: %v", err, err)
			Expect(cycleErr.Cycles).NotTo(BeEmpty())
			Expect(cycleErr.Cycles[0].Ref).To(Equal("#/components/schemas/Node"))
		})

		It("returns a *RefCycleError for an indirect A->B->A cycle", func() {
			_, err := LoadSpec(filepath.Join("../testdata/parser", "cycle_indirect.yaml"))
			var cycleErr *RefCycleError
			Expect(errors.As(err, &cycleErr)).To(BeTrue(), "got %T: %v", err, err)
		})
	})

	Context("depth limiting", func() {
		var deepSpec = filepath.Join("../testdata/parser", "deep_chain.yaml")

		It("validates an 8-deep chain at the default max-depth", func() {
			_, err := LoadSpec(deepSpec)
			Expect(err).To(Succeed())
		})

		It("validates an 8-deep chain at an explicit matching max-depth", func() {
			_, err := LoadSpec(deepSpec, WithMaxDepth(8))
			Expect(err).To(Succeed())
		})

		It("returns a *MaxDepthError when depth is exceeded", func() {
			_, err := LoadSpec(deepSpec, WithMaxDepth(4))
			Expect(err).To(HaveOccurred())

			var cycleErr *RefCycleError
			Expect(errors.As(err, &cycleErr)).To(BeFalse(), "deep-but-acyclic refs must not be reported as a cycle")

			var depthErr *MaxDepthError
			Expect(errors.As(err, &depthErr)).To(BeTrue(), "got %T: %v", err, err)
		})
	})

	Context("tracking metadata", func() {
		It("populates tracking on flagged operations and leaves it nil on unflagged ones", func() {
			spec, err := LoadSpec(filepath.Join("../testdata/parser", "tracking_valid.yaml"))
			Expect(err).To(Succeed())

			byID := make(map[string]*model.Operation, len(spec.Operations))
			for _, op := range spec.Operations {
				byID[op.OperationId] = op
			}

			res := byID["CreateIncidentType"]
			Expect(res).NotTo(BeNil())
			Expect(res.Tracking).NotTo(BeNil())
			Expect(res.Tracking.ArtifactKind).To(Equal(model.ArtifactKindResource))
			Expect(res.Tracking.ArtifactName).To(Equal("incident_type"))

			ds := byID["GetTeam"]
			Expect(ds).NotTo(BeNil())
			Expect(ds.Tracking).NotTo(BeNil())
			Expect(ds.Tracking.ArtifactKind).To(Equal(model.ArtifactKindDataSource))

			h := byID["GetHealth"]
			Expect(h).NotTo(BeNil())
			Expect(h.Tracking).To(BeNil(), "unflagged operation must have nil tracking")
		})

		It("returns a *DuplicateArtifactNameError naming both sources", func() {
			_, err := LoadSpec(filepath.Join("../testdata/parser", "tracking_duplicate.yaml"))
			var dup *DuplicateArtifactNameError
			Expect(errors.As(err, &dup)).To(BeTrue(), "got %T: %v", err, err)
			Expect(dup.Error()).To(SatisfyAll(
				ContainSubstring("ListTeams"),
				ContainSubstring("GetTeam"),
			))
		})

		It("returns a *TrackingError with the correct path for a malformed extension", func() {
			_, err := LoadSpec(filepath.Join("../testdata/parser", "tracking_malformed.yaml"))
			var te *TrackingError
			Expect(errors.As(err, &te)).To(BeTrue(), "got %T: %v", err, err)
			Expect(te.Path).To(Equal("/widgets"))
		})

		It("allows a resource and data source to share an artifact_name", func() {
			spec, err := LoadSpec(filepath.Join("../testdata/parser", "tracking_cross_kind_names.yaml"))
			Expect(err).To(Succeed(), "a resource and data source sharing a name must load without error")

			byID := make(map[string]*model.Operation, len(spec.Operations))
			for _, op := range spec.Operations {
				byID[op.OperationId] = op
			}

			r := byID["CreateTeam"]
			Expect(r).NotTo(BeNil())
			Expect(r.Tracking).NotTo(BeNil())
			Expect(r.Tracking.ArtifactKind).To(Equal(model.ArtifactKindResource))

			d := byID["GetTeam"]
			Expect(d).NotTo(BeNil())
			Expect(d.Tracking).NotTo(BeNil())
			Expect(d.Tracking.ArtifactKind).To(Equal(model.ArtifactKindDataSource))
		})
	})
})
