package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func boolPtr(v bool) *bool { return &v }

var _ = Describe("Operation lifecycle roles", func() {
	// A minimal annotation may name one operationId as both create and read,
	// and several fixtures in testdata/parser do exactly that.
	It("recognizes every role an operation fills", func() {
		op := &Operation{LifecycleRoles: []GroupRole{GroupRoleCreate, GroupRoleRead}}
		Expect(op.HasLifecycleRole(GroupRoleCreate)).To(BeTrue())
		Expect(op.HasLifecycleRole(GroupRoleRead)).To(BeTrue())
		Expect(op.HasLifecycleRole(GroupRoleDelete)).To(BeFalse())
	})

	It("tolerates a nil operation or an unassigned one", func() {
		var op *Operation
		Expect(op.HasLifecycleRole(GroupRoleCreate)).To(BeFalse())
		Expect((&Operation{}).HasLifecycleRole(GroupRoleCreate)).To(BeFalse())
	})
})

var _ = Describe("Operation response examples", func() {
	op := func() *Operation {
		return &Operation{ResponseExamples: []ResponseExamples{
			{Status: "404"},
			{Status: "201", BodyPresent: true, MediaType: "application/json"},
			{Status: "200", BodyPresent: true, MediaType: "application/json"},
			{Status: "429"},
		}}
	}

	Describe("ResponseExampleFor", func() {
		// The post-destroy read expects a 404, so failure outcomes must be
		// addressable and not just the success one.
		It("finds a declared failure outcome", func() {
			Expect(op().ResponseExampleFor("404")).NotTo(BeNil())
			Expect(op().ResponseExampleFor("404").Status).To(Equal("404"))
		})

		It("returns nil for an undeclared status", func() {
			Expect(op().ResponseExampleFor("500")).To(BeNil())
		})

		// Status is compared as written, so a range does not answer for a code.
		It("does not match a range against a concrete code", func() {
			ranged := &Operation{ResponseExamples: []ResponseExamples{{Status: "4XX"}}}
			Expect(ranged.ResponseExampleFor("404")).To(BeNil())
		})

		It("tolerates a nil operation", func() {
			var nilOp *Operation
			Expect(nilOp.ResponseExampleFor("200")).To(BeNil())
		})
	})

	Describe("SuccessResponseExample", func() {
		// Must agree with schema normalization's choice: a cassette built from
		// one response while the provider model came from another would replay
		// values the generated code cannot decode.
		It("selects the lowest-numbered 2xx regardless of declaration order", func() {
			got := op().SuccessResponseExample()
			Expect(got).NotTo(BeNil())
			Expect(got.Status).To(Equal("200"))
		})

		It("selects a 201 when it is the only success outcome", func() {
			only := &Operation{ResponseExamples: []ResponseExamples{
				{Status: "400"}, {Status: "201", BodyPresent: true},
			}}
			Expect(only.SuccessResponseExample().Status).To(Equal("201"))
		})

		It("returns nil when no success outcome is declared", func() {
			none := &Operation{ResponseExamples: []ResponseExamples{{Status: "404"}}}
			Expect(none.SuccessResponseExample()).To(BeNil())
			var nilOp *Operation
			Expect(nilOp.SuccessResponseExample()).To(BeNil())
		})

		It("ignores a non-numeric status rather than failing", func() {
			ranged := &Operation{ResponseExamples: []ResponseExamples{
				{Status: "2XX"}, {Status: "204"},
			}}
			Expect(ranged.SuccessResponseExample().Status).To(Equal("204"))
		})
	})

	Describe("ParameterExampleFor", func() {
		// (name, location) is the merge key, so the same name in two locations
		// is two parameters.
		It("distinguishes the same name in different locations", func() {
			withBoth := &Operation{ParameterExamples: []ParameterExamples{
				{Name: "id", In: ParameterInPath},
				{Name: "id", In: ParameterInQuery},
			}}
			Expect(withBoth.ParameterExampleFor("id", ParameterInPath).In).To(Equal(ParameterInPath))
			Expect(withBoth.ParameterExampleFor("id", ParameterInQuery).In).To(Equal(ParameterInQuery))
			Expect(withBoth.ParameterExampleFor("other", ParameterInQuery)).To(BeNil())
		})

		It("tolerates a nil operation", func() {
			var nilOp *Operation
			Expect(nilOp.ParameterExampleFor("id", ParameterInPath)).To(BeNil())
		})
	})
})

var _ = Describe("QueryParam serialization", func() {
	Describe("ResolvedStyle", func() {
		It("defaults by location when the parameter declares no style", func() {
			Expect(QueryParam{In: ParameterInPath}.ResolvedStyle()).To(Equal(ParameterStyleSimple))
			Expect(QueryParam{In: ParameterInQuery}.ResolvedStyle()).To(Equal(ParameterStyleForm))
		})

		It("honors a declared style", func() {
			p := QueryParam{In: ParameterInQuery, Style: ParameterStylePipeDelimited}
			Expect(p.ResolvedStyle()).To(Equal(ParameterStylePipeDelimited))
		})
	})

	Describe("ResolvedExplode", func() {
		// The pointer exists precisely for this case: form style defaults to
		// true, so a plain bool could not tell "the spec said false" from
		// "the spec said nothing", and the two produce different URLs.
		It("distinguishes an explicit false from an unset flag on form style", func() {
			unset := QueryParam{In: ParameterInQuery}
			Expect(unset.ResolvedExplode()).To(BeTrue())

			explicitFalse := QueryParam{In: ParameterInQuery, Explode: boolPtr(false)}
			Expect(explicitFalse.ResolvedExplode()).To(BeFalse())
		})

		It("defaults to false for a non-form style", func() {
			Expect(QueryParam{In: ParameterInPath}.ResolvedExplode()).To(BeFalse())
			p := QueryParam{In: ParameterInQuery, Style: ParameterStyleSpaceDelimited}
			Expect(p.ResolvedExplode()).To(BeFalse())
		})

		It("honors an explicit true on a style that defaults false", func() {
			p := QueryParam{In: ParameterInPath, Explode: boolPtr(true)}
			Expect(p.ResolvedExplode()).To(BeTrue())
		})
	})
})
