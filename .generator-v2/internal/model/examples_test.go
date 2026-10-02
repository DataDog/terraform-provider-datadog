package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExampleLocation", func() {
	DescribeTable("renders a spec-prefixed anchor matching Diagnostic.Location's form",
		func(in ExampleLocation, want string) { Expect(in.String()).To(Equal(want)) },
		Entry("request body, singular form",
			ExampleLocation{
				Path: "/api/v2/widgets", Method: "POST", Component: ExampleComponentRequestBody,
				MediaType: "application/json", Field: "example",
			},
			"spec:/api/v2/widgets.post.requestBody.content.application/json.example"),
		Entry("response, one name out of a named map",
			ExampleLocation{
				Path: "/api/v2/widgets", Method: "POST", Component: ExampleComponentResponse,
				Detail: "201", MediaType: "application/json", Field: "examples", ExampleName: "default",
			},
			"spec:/api/v2/widgets.post.responses.201.content.application/json.examples.default"),
		Entry("parameter, which is not media-scoped",
			ExampleLocation{
				Path: "/api/v2/widgets/{widget_id}", Method: "GET",
				Component: ExampleComponentParameter, Detail: "widget_id", Field: "example",
			},
			"spec:/api/v2/widgets/{widget_id}.get.parameters.widget_id.example"),
		Entry("schema-property candidate carries the property path",
			ExampleLocation{
				Path: "/api/v2/widgets", Method: "POST", Component: ExampleComponentRequestBody,
				MediaType: "application/json", Field: "example", PropertyPath: "data.attributes.name",
			},
			"spec:/api/v2/widgets.post.requestBody.content.application/json.example (data.attributes.name)"),
	)
})

var _ = Describe("ExampleCandidate", func() {
	It("treats an inline candidate as eligible", func() {
		c := &ExampleCandidate{Value: map[string]any{"name": "widget"}}
		Expect(c.Eligible()).To(BeTrue())
		Expect(c.IneligibleReason()).To(BeEmpty())
	})

	// A declared JSON null is a value the spec chose, not a missing example;
	// absence is the candidate not existing at all.
	It("treats a declared null value as eligible", func() {
		Expect((&ExampleCandidate{Value: nil}).Eligible()).To(BeTrue())
	})

	It("never treats an external candidate as eligible", func() {
		c := &ExampleCandidate{
			External: true,
			Location: ExampleLocation{Path: "/api/v2/widgets", Method: "POST",
				Component: ExampleComponentRequestBody, Field: "examples", ExampleName: "remote"},
		}
		Expect(c.Eligible()).To(BeFalse())
		Expect(c.IneligibleReason()).To(ContainSubstring("external value"))
		Expect(c.IneligibleReason()).To(ContainSubstring("/api/v2/widgets"))
	})

	It("explains a nil candidate rather than panicking", func() {
		var c *ExampleCandidate
		Expect(c.Eligible()).To(BeFalse())
		Expect(c.IneligibleReason()).To(Equal("no example is declared"))
	})

	// A sensitive value must never reach a diagnostic, so the reason string is
	// built from the anchor alone.
	It("keeps a sensitive candidate's value out of its diagnostic", func() {
		c := &ExampleCandidate{
			Value: "plaintext-token-value", Sensitive: true, External: true,
			Location: ExampleLocation{Path: "/api/v2/widgets", Method: "POST"},
		}
		Expect(c.IneligibleReason()).NotTo(ContainSubstring("plaintext-token-value"))
	})
})

var _ = Describe("ExampleSet", func() {
	Describe("SortedNames", func() {
		// Byte-identical regeneration depends on selection never seeing map order.
		It("returns names in lexical order regardless of insertion order", func() {
			s := &ExampleSet{Named: map[string]*ExampleCandidate{
				"zeta": {Name: "zeta"}, "default": {Name: "default"}, "alternate": {Name: "alternate"},
			}}
			Expect(s.SortedNames()).To(Equal([]string{"alternate", "default", "zeta"}))
		})

		It("returns nothing for an unnamed or nil set", func() {
			Expect((&ExampleSet{}).SortedNames()).To(BeEmpty())
			var s *ExampleSet
			Expect(s.SortedNames()).To(BeEmpty())
		})
	})

	Describe("Empty", func() {
		It("is true only when no candidate of any kind is declared", func() {
			var nilSet *ExampleSet
			Expect(nilSet.Empty()).To(BeTrue())
			Expect((&ExampleSet{}).Empty()).To(BeTrue())
			Expect((&ExampleSet{Single: &ExampleCandidate{}}).Empty()).To(BeFalse())
			Expect((&ExampleSet{Named: map[string]*ExampleCandidate{"d": {}}}).Empty()).To(BeFalse())
			Expect((&ExampleSet{SchemaFallback: []*ExampleCandidate{{}}}).Empty()).To(BeFalse())
		})
	})

	Describe("Malformed", func() {
		// OpenAPI makes `example` and `examples` mutually exclusive. Resolving
		// one way would pick a value the description does not determine.
		It("flags a location declaring both the singular and named forms", func() {
			s := &ExampleSet{
				Single: &ExampleCandidate{},
				Named:  map[string]*ExampleCandidate{"default": {}},
				Location: ExampleLocation{Path: "/api/v2/widgets", Method: "POST",
					Component: ExampleComponentRequestBody, MediaType: "application/json"},
			}
			Expect(s.Malformed()).To(BeTrue())
			Expect(s.Validate()).To(MatchError(ContainSubstring("mutually exclusive")))
			Expect(s.Validate()).To(MatchError(ContainSubstring("/api/v2/widgets")))
		})

		It("accepts either form on its own", func() {
			Expect((&ExampleSet{Single: &ExampleCandidate{}}).Malformed()).To(BeFalse())
			Expect((&ExampleSet{Named: map[string]*ExampleCandidate{"d": {}}}).Malformed()).To(BeFalse())
		})

		// Whether a missing example is fatal depends on the interaction needing
		// one, which only the scenario knows — so an empty set validates here.
		It("does not treat an empty set as invalid", func() {
			Expect((&ExampleSet{}).Validate()).To(Succeed())
		})
	})
})

var _ = Describe("parameter serialization defaults", func() {
	DescribeTable("defaults style by location",
		func(in ParameterIn, want ParameterStyle) { Expect(DefaultParameterStyle(in)).To(Equal(want)) },
		Entry("path", ParameterInPath, ParameterStyleSimple),
		Entry("query", ParameterInQuery, ParameterStyleForm),
	)

	DescribeTable("defaults explode by style, which is true only for form",
		func(in ParameterStyle, want bool) { Expect(DefaultParameterExplode(in)).To(Equal(want)) },
		Entry("form", ParameterStyleForm, true),
		Entry("simple", ParameterStyleSimple, false),
		Entry("spaceDelimited", ParameterStyleSpaceDelimited, false),
		Entry("pipeDelimited", ParameterStylePipeDelimited, false),
		Entry("deepObject", ParameterStyleDeepObject, false),
	)
})

var _ = Describe("ResponseExamples", func() {
	Describe("Bodyless", func() {
		// FR-017 turns on this distinction: a 204 declaring no content is a
		// complete contract, not a missing example.
		It("is true for a 204 or 205 declaring no content", func() {
			Expect((&ResponseExamples{Status: "204"}).Bodyless()).To(BeTrue())
			Expect((&ResponseExamples{Status: "205"}).Bodyless()).To(BeTrue())
		})

		It("is false for a 200 that declares no content, which is a gap", func() {
			Expect((&ResponseExamples{Status: "200"}).Bodyless()).To(BeFalse())
		})

		It("is false when a body is declared, whatever the status", func() {
			Expect((&ResponseExamples{Status: "204", BodyPresent: true}).Bodyless()).To(BeFalse())
		})

		It("is false for a nil response", func() {
			var r *ResponseExamples
			Expect(r.Bodyless()).To(BeFalse())
		})
	})

	Describe("Complete", func() {
		It("accepts a legitimately bodyless response with no example", func() {
			Expect((&ResponseExamples{Status: "204"}).Complete()).To(BeTrue())
		})

		It("accepts a body backed by at least one candidate", func() {
			r := &ResponseExamples{Status: "200", BodyPresent: true, MediaType: "application/json",
				Examples: ExampleSet{Single: &ExampleCandidate{Value: map[string]any{"id": "1"}}}}
			Expect(r.Complete()).To(BeTrue())
		})

		It("rejects a declared body with no candidate", func() {
			Expect((&ResponseExamples{Status: "200", BodyPresent: true}).Complete()).To(BeFalse())
		})

		It("rejects a non-bodyless status that declares nothing at all", func() {
			Expect((&ResponseExamples{Status: "200"}).Complete()).To(BeFalse())
		})

		It("rejects a nil response rather than panicking", func() {
			var r *ResponseExamples
			Expect(r.Complete()).To(BeFalse())
		})
	})
})
