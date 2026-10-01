package cassette

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

// fixtureOperations loads a fixture and returns the tracked operation for an
// artifact followed by its resolved group, which is the lifecycle order the
// generated test runs in.
func fixtureOperations(fixture, artifact string) []*model.Operation {
	spec, err := parser.LoadSpec(filepath.Join("..", "testdata", fixture))
	Expect(err).To(Succeed())
	for _, op := range spec.Operations {
		if op.Tracking == nil || op.Tracking.ArtifactName != artifact {
			continue
		}
		ops := []*model.Operation{op}
		for _, target := range op.ResolvedGroup.Operations() {
			if target != op {
				ops = append(ops, target)
			}
		}
		return ops
	}
	Fail("no tracked operation for artifact " + artifact)
	return nil
}

func matrixOperation(artifact string) *model.Operation {
	return fixtureOperations(filepath.Join("parser", "openapi_examples.yaml"), artifact)[0]
}

func ineligible(err error) *IneligibleError {
	var target *IneligibleError
	ExpectWithOffset(1, errors.As(err, &target)).To(BeTrue(), "want an *IneligibleError, got %v", err)
	return target
}

var _ = Describe("Select", func() {
	Describe("singular examples", func() {
		// A target with no named set imposes no name, so the scenario is
		// "single" and the singular candidates are used directly.
		It("resolves a singular-only target as the single scenario", func() {
			op := matrixOperation("media_singular")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())
			Expect(selection.ScenarioName).To(Equal(model.ScenarioNameSingle))

			request, ok := selection.Set(SetKey{OperationId: "CreateMediaSingular", Role: SetRoleRequest})
			Expect(ok).To(BeTrue())
			Expect(request.Candidate).NotTo(BeNil())
			Expect(request.Candidate.SourceKind).To(Equal(model.ExampleSourceMedia))
			Expect(request.Candidate.Name).To(BeEmpty())

			response, ok := selection.Set(SetKey{
				OperationId: "CreateMediaSingular", Role: SetRoleResponse, Detail: "201"})
			Expect(ok).To(BeTrue())
			Expect(response.Candidate).NotTo(BeNil())
		})

		It("records provenance for every selected value", func() {
			op := matrixOperation("media_singular")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())
			Expect(selection.Provenance).To(HaveLen(len(selection.Sets)))
			for _, entry := range selection.Provenance {
				Expect(entry.Location.String()).To(HavePrefix("spec:/media-singular"))
			}
		})

		It("exposes the selection in the form the run report carries", func() {
			op := matrixOperation("media_singular")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())

			reported := selection.Model()
			Expect(reported.ScenarioName).To(Equal(model.ScenarioNameSingle))
			Expect(reported.Provenance).To(HaveLen(len(selection.Provenance)))

			// Model must copy, so a caller mutating the report cannot corrupt
			// the selection it was derived from.
			reported.Provenance[0].CandidateName = "mutated"
			Expect(selection.Provenance[0].CandidateName).NotTo(Equal("mutated"))
		})
	})

	Describe("named examples", func() {
		// The Twilio fixture names every example "default", so exactly one
		// common name spans the whole CRUD group.
		It("resolves a target whose named sets share exactly one common name", func() {
			ops := fixtureOperations(
				filepath.Join("mini-oas", "mini-datadog_integration_twilio_account.yaml"),
				"integration_twilio_account")
			selection, err := Select(ops)
			Expect(err).To(Succeed())
			Expect(selection.ScenarioName).To(Equal("default"))

			for _, set := range selection.Sets {
				Expect(set.Resolved()).To(BeTrue(), set.Key.String())
			}
			// Every candidate chosen by name must actually carry that name.
			for _, set := range selection.Sets {
				if set.Candidate != nil && set.Candidate.Name != "" {
					Expect(set.Candidate.Name).To(Equal("default"))
				}
			}
		})

		It("spans the whole lifecycle rather than one operation", func() {
			ops := fixtureOperations(
				filepath.Join("mini-oas", "mini-datadog_integration_twilio_account.yaml"),
				"integration_twilio_account")
			selection, err := Select(ops)
			Expect(err).To(Succeed())

			var operations []string
			for _, set := range selection.Sets {
				operations = append(operations, set.Key.OperationId)
			}
			Expect(operations).To(ContainElements(
				"CreateTwilioIntegrationAccount",
				"GetTwilioIntegrationAccount",
				"UpdateTwilioIntegrationAccount",
			))
		})

		// Preferring "default" and falling back to the lexically smallest
		// common name is a documented rule that belongs with explicit
		// overrides; guessing here would silently pick a scenario.
		It("declines a target offering several common names", func() {
			op := matrixOperation("media_named")
			_, err := Select([]*model.Operation{op})
			Expect(err).To(HaveOccurred())

			reason := ineligible(err)
			Expect(reason.Reason).To(ContainSubstring("several common names"))
			Expect(reason.Error()).To(ContainSubstring("alternate"))
			Expect(reason.Error()).To(ContainSubstring("default"))
			Expect(reason.Sets).NotTo(BeEmpty())
		})
	})

	Describe("incoherent targets", func() {
		// Combining independently named candidates would invent a scenario the
		// description never described.
		It("declines when named sets share no common name", func() {
			op := matrixOperation("media_named")
			// Rename the response's names so request and response disagree
			// entirely, leaving an empty intersection.
			response := op.SuccessResponseExample()
			response.Examples.Named = map[string]*model.ExampleCandidate{
				"other": {Name: "other", Value: map[string]any{}},
			}
			_, err := Select([]*model.Operation{op})
			Expect(err).To(HaveOccurred())
			Expect(ineligible(err).Reason).To(ContainSubstring("share no common name"))
		})

		It("declines a set declaring both example forms", func() {
			op := matrixOperation("bad_both_forms")
			_, err := Select([]*model.Operation{op})
			Expect(err).To(HaveOccurred())
			Expect(ineligible(err).Reason).To(ContainSubstring("mutually exclusive"))
		})

		// An external value is not in the description and can never be
		// reproduced from it, so it is never chosen. It does not by itself
		// doom the target: the schema may still offer usable fallbacks, and
		// whether those are *complete* is materialization's question.
		It("never selects an external candidate, falling back to the schema", func() {
			op := matrixOperation("bad_external")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())

			request, ok := selection.Set(SetKey{OperationId: "CreateBadExternal", Role: SetRoleRequest})
			Expect(ok).To(BeTrue())
			if request.Candidate != nil {
				Expect(request.Candidate.External).To(BeFalse())
				Expect(request.Candidate.SourceKind).NotTo(Equal(model.ExampleSourceMedia))
			}
			for _, fallback := range request.Fallbacks {
				Expect(fallback.External).To(BeFalse())
			}
			// Provenance must never point at a value that was not used.
			for _, entry := range selection.Provenance {
				Expect(entry.CandidateName).NotTo(Equal("remote"))
			}
		})

		It("declines when a set offers only an external candidate and no fallback", func() {
			op := &model.Operation{
				OperationId: "CreateExternalOnly",
				RequestExamples: &model.RequestBodyExamples{
					Present:   true,
					MediaType: "application/json",
					Examples: model.ExampleSet{
						Named: map[string]*model.ExampleCandidate{
							"remote": {Name: "remote", External: true},
						},
					},
				},
			}
			_, err := Select([]*model.Operation{op})
			Expect(err).To(HaveOccurred())
			Expect(ineligible(err).Reason).To(ContainSubstring("no usable example"))
			Expect(ineligible(err).Error()).To(ContainSubstring("CreateExternalOnly"))
		})

		It("declines a target with no example sets at all", func() {
			_, err := Select([]*model.Operation{{OperationId: "Bare"}})
			Expect(err).To(HaveOccurred())
			Expect(ineligible(err).Reason).To(ContainSubstring("declares no example sets"))
		})

		It("tolerates a nil operation in the group", func() {
			_, err := Select([]*model.Operation{nil})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("schema fallbacks", func() {
		// A whole-body schema example is something a human wrote, so it wins
		// over the property pieces the generator would assemble.
		It("prefers a whole-body schema example over property pieces", func() {
			op := matrixOperation("schema_example")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())

			request, ok := selection.Set(SetKey{OperationId: "CreateSchemaExample", Role: SetRoleRequest})
			Expect(ok).To(BeTrue())
			Expect(request.Candidate).NotTo(BeNil())
			Expect(request.Candidate.SourceKind).To(Equal(model.ExampleSourceSchema))
			Expect(request.Fallbacks).To(BeEmpty())
		})

		It("falls back to property pieces when no whole value is declared", func() {
			op := matrixOperation("schema_property_example")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())

			request, ok := selection.Set(SetKey{
				OperationId: "CreateSchemaPropertyExample", Role: SetRoleRequest})
			Expect(ok).To(BeTrue())
			Expect(request.Candidate).To(BeNil())
			Expect(request.Fallbacks).NotTo(BeEmpty())
			for _, fallback := range request.Fallbacks {
				Expect(fallback.SourceKind).To(Equal(model.ExampleSourceSchemaProperty))
				Expect(fallback.Location.PropertyPath).NotTo(BeEmpty())
			}
			// Every assembled piece is traceable, not just the set as a whole.
			Expect(len(selection.Provenance)).To(BeNumerically(">=", len(request.Fallbacks)))
		})
	})

	Describe("required sets", func() {
		// Demanding an example for a declared-bodyless response would make
		// every DELETE ineligible.
		It("does not require an example for a declared bodyless response", func() {
			ops := fixtureOperations(filepath.Join("parser", "openapi_examples.yaml"), "bodyless")
			selection, err := Select(ops)
			Expect(err).To(Succeed())

			// The delete still contributes its path parameter — that is a real
			// input — but contributes no response set, because 204 and 205
			// are complete contracts that need no example.
			for _, set := range selection.Sets {
				if set.Key.OperationId != "DeleteBodyless" {
					continue
				}
				Expect(set.Key.Role).NotTo(Equal(SetRoleResponse),
					"a declared bodyless response needs no example")
			}
		})

		It("requires a path parameter's example but not an unexampled optional query one", func() {
			op := matrixOperation("parameters_case")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())

			_, ok := selection.Set(SetKey{
				OperationId: "GetWithParameters", Role: SetRoleParameter, Detail: "path:widget_id"})
			Expect(ok).To(BeTrue())

			_, ok = selection.Set(SetKey{
				OperationId: "GetWithParameters", Role: SetRoleParameter, Detail: "query:tags"})
			Expect(ok).To(BeTrue())
		})
	})

	Describe("TargetSelection accessors", func() {
		It("reports a key it does not hold", func() {
			op := matrixOperation("media_singular")
			selection, err := Select([]*model.Operation{op})
			Expect(err).To(Succeed())
			_, ok := selection.Set(SetKey{OperationId: "Nope", Role: SetRoleRequest})
			Expect(ok).To(BeFalse())
		})
	})

	Describe("IneligibleError", func() {
		It("renders without set names when it concerns none", func() {
			err := &IneligibleError{Reason: "the target declares no example sets"}
			Expect(err.Error()).To(Equal("the target declares no example sets"))
		})

		It("names the sets it concerns", func() {
			err := &IneligibleError{
				Reason: "no usable example",
				Sets: []SetKey{
					{OperationId: "CreateX", Role: SetRoleRequest},
					{OperationId: "GetX", Role: SetRoleParameter, Detail: "path:id"},
				},
			}
			Expect(err.Error()).To(Equal(
				"no usable example (CreateX request and GetX parameter path:id)"))
		})
	})

	Describe("SetKey", func() {
		It("renders readably for diagnostics", func() {
			Expect(SetKey{OperationId: "CreateX", Role: SetRoleRequest}.String()).
				To(Equal("CreateX request"))
			Expect(SetKey{OperationId: "GetX", Role: SetRoleParameter, Detail: "path:id"}.String()).
				To(Equal("GetX parameter path:id"))
		})
	})
})

var _ = Describe("joinAnd", func() {
	DescribeTable("renders a list as prose",
		func(in []string, want string) { Expect(joinAnd(in)).To(Equal(want)) },
		Entry("empty", []string{}, ""),
		Entry("one", []string{"a"}, "a"),
		Entry("two", []string{"a", "b"}, "a and b"),
		Entry("three", []string{"a", "b", "c"}, "a, b, and c"),
	)
})
