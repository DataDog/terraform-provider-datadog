package model

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The tree *shape* of a projected envelope is pinned in schema_test.go. These
// specs cover the metadata the emit layer reads off the projection, and the
// failure paths that must not degrade into a dropped field.

var _ = Describe("oneOf envelope metadata", func() {

	// envelopeAt projects schema and returns the OneOfEnvelope carried by the
	// attribute at path.
	envelopeAt := func(schema *Schema, path string) *OneOfEnvelope {
		GinkgoHelper()
		tree, _, err := BuildResponseTree(schema)
		Expect(err).NotTo(HaveOccurred())
		envelope := attrByPath(tree, path).OneOf
		Expect(envelope).NotTo(BeNil(), "no oneOf envelope on the attribute at %q", path)
		return envelope
	}

	It("names the envelope model after the parser's envelope identity", func() {
		envelope := envelopeAt(
			objSchema(map[string]*Schema{
				"choice": oneOfSchema(
					"response.choice",
					"ActionConnectionIntegration",
					primitiveOneOfVariant("boolean", "boolean"),
				),
			}),
			"response.choice",
		)
		Expect(envelope.Name).To(Equal("ActionConnectionIntegration"))
		Expect(envelope.GoModel).To(Equal("ActionConnectionIntegrationModel"))
		Expect(envelope.Path).To(Equal("response.choice"))
	})

	It("gives two uses of one reusable component the same generated model", func() {
		shared := func() *Schema {
			return oneOfSchema("", "Credentials", primitiveOneOfVariant("string", "string"))
		}
		first := shared()
		first.OneOf.Path = "response.a"
		second := shared()
		second.OneOf.Path = "response.b"

		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{"a": first, "b": second}))
		Expect(err).NotTo(HaveOccurred())

		a := attrByPath(tree, "response.a").OneOf
		b := attrByPath(tree, "response.b").OneOf
		Expect(a.GoModel).To(Equal(b.GoModel))
		Expect(a.Variants[0].GoModel).To(Equal(b.Variants[0].GoModel))
		// Only the tree position differs, so a reader can tell the two uses apart.
		Expect(a.Path).NotTo(Equal(b.Path))
	})

	It("hangs the envelope off the collection attribute when the union is an element", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"choices": arrSchema(oneOfSchema(
				"response.choices[]",
				"ChoicesItem",
				primitiveOneOfVariant("string", "string"),
			)),
		}))
		Expect(err).NotTo(HaveOccurred())

		collection := attrByPath(tree, "response.choices")
		Expect(collection.TfType).To(Equal("schema.ListNestedBlock"))
		Expect(collection.OneOf).NotTo(BeNil())
		// The envelope path is the element path, which no attribute occupies.
		Expect(collection.OneOf.Path).To(Equal("response.choices[]"))
		Expect(collection.OneOf.Variants[0].Attribute.Path).To(Equal("response.choices[].string"))
	})

	It("orders variants by Terraform name whatever order the parser hands over", func() {
		envelope := envelopeAt(
			oneOfSchema(
				"response",
				"Choice",
				primitiveOneOfVariant("string", "string"),
				objectOneOfVariant("alpha", map[string]*Schema{"x": primSchema("string")}),
				primitiveOneOfVariant("boolean", "boolean"),
			),
			"response",
		)
		names := make([]string, len(envelope.Variants))
		for i, v := range envelope.Variants {
			names[i] = v.TFName
		}
		Expect(names).To(Equal([]string{"alpha", "boolean", "string"}))
	})

	It("points each variant at the same block the envelope attribute lists as a child", func() {
		tree, _, err := BuildResponseTree(oneOfSchema(
			"response",
			"Choice",
			primitiveOneOfVariant("boolean", "boolean"),
			primitiveOneOfVariant("string", "string"),
		))
		Expect(err).NotTo(HaveOccurred())

		envelope := attrByPath(tree, "response")
		Expect(envelope.OneOf.Variants).To(HaveLen(len(envelope.Children)))
		for i, variant := range envelope.OneOf.Variants {
			Expect(variant.Attribute).To(BeIdenticalTo(envelope.Children[i]),
				"variant %q must alias Children[%d]", variant.TFName, i)
		}
	})

	DescribeTable("value-wraps every alternative that has no fields of its own",
		func(variant OneOfVariant, wantWrapped bool, wantChildPath, wantChildType string) {
			envelope := envelopeAt(oneOfSchema("response", "Choice", variant), "response")
			Expect(envelope.Variants).To(HaveLen(1))
			Expect(envelope.Variants[0].ValueWrapped).To(Equal(wantWrapped))

			child := envelope.Variants[0].Attribute.Children[0]
			Expect(child.Path).To(Equal(wantChildPath))
			Expect(child.TfType).To(Equal(wantChildType))
		},
		Entry("a scalar alternative",
			primitiveOneOfVariant("string", "string"),
			true, "response.string.value", "schema.StringAttribute"),
		Entry("a list alternative",
			OneOfVariant{TFName: "list", GoName: "List", Schema: arrSchema(primSchema("string"))},
			true, "response.list.value", "schema.ListAttribute"),
		Entry("a map alternative",
			OneOfVariant{TFName: "map", GoName: "Map", Schema: mapSchema(primSchema("string"))},
			true, "response.map.value", "schema.MapAttribute"),
		Entry("an object alternative exposes its fields directly instead",
			objectOneOfVariant("object", map[string]*Schema{"name": primSchema("string")}),
			false, "response.object.name", "schema.StringAttribute"),
		Entry("an alternative that is itself a union has no fields to expose either",
			OneOfVariant{
				TFName: "nested",
				GoName: "Nested",
				Schema: oneOfSchema("response.nested.value", "Nested", primitiveOneOfVariant("string", "string")),
			},
			true, "response.nested.value", "schema.SingleNestedBlock"),
	)

	It("derives the variant's Go field and model from the parser's Go name", func() {
		envelope := envelopeAt(
			oneOfSchema("response", "Choice", objectOneOfVariant("aws_integration", nil)),
			"response",
		)
		variant := envelope.Variants[0]
		Expect(variant.GoField).To(Equal("AwsIntegration"))
		Expect(variant.GoModel).To(Equal("ChoiceAwsIntegrationModel"))
	})

	It("carries the SDK binding through untouched rather than deriving it from Terraform names", func() {
		variant := objectOneOfVariant("aws_integration", nil)
		// The SDK keeps the component's acronym casing, which the snake_case
		// Terraform name cannot reproduce — hence the passthrough.
		variant.SDKField = "AWSIntegration"
		variant.SDKConstructor = "AWSIntegrationAsActionConnectionIntegration"

		envelope := envelopeAt(oneOfSchema("response", "Choice", variant), "response")
		Expect(envelope.Variants[0].SDKField).To(Equal("AWSIntegration"))
		Expect(envelope.Variants[0].SDKConstructor).To(Equal("AWSIntegrationAsActionConnectionIntegration"))
	})

	DescribeTable("marks the envelope optional when it may legally be absent",
		func(mutate func(*OneOfSpec), wantOptional bool) {
			schema := oneOfSchema("response", "Choice", primitiveOneOfVariant("string", "string"))
			mutate(schema.OneOf)
			Expect(envelopeAt(schema, "response").Optional).To(Equal(wantOptional))
		},
		Entry("a required, non-nullable union is not optional",
			func(*OneOfSpec) {}, false),
		Entry("an optional containing field makes the envelope optional",
			func(s *OneOfSpec) { s.Optional = true }, true),
		Entry("a nullable union is optional too — null is an absent envelope",
			func(s *OneOfSpec) { s.Nullable = true }, true),
	)

	It("marks a response envelope Computed and a request envelope Required", func() {
		schema := oneOfSchema("response", "Choice", primitiveOneOfVariant("string", "string"))

		Expect(envelopeAt(schema, "response").Computed).To(BeTrue())

		tree, _, err := BuildRequestTree(oneOfSchema("request", "Choice", primitiveOneOfVariant("string", "string")))
		Expect(err).NotTo(HaveOccurred())
		envelope := attrByPath(tree, "request")
		Expect(envelope.OneOf.Computed).To(BeFalse())
		Expect(envelope.Required).To(BeTrue())
		Expect(envelope.Optional).To(BeFalse())
		// A branch is a choice, so it is never itself mandatory; the wrapped value
		// is, once its branch is selected.
		branch := attrByPath(tree, "request.string")
		Expect(branch.Optional).To(BeTrue())
		Expect(branch.Required).To(BeFalse())
		Expect(attrByPath(tree, "request.string.value").Required).To(BeTrue())
	})

	It("leaves an optional request envelope optional rather than required", func() {
		schema := oneOfSchema("request", "Choice", primitiveOneOfVariant("string", "string"))
		schema.OneOf.Optional = true

		tree, _, err := BuildRequestTree(schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(attrByPath(tree, "request").Optional).To(BeTrue())
		Expect(attrByPath(tree, "request").Required).To(BeFalse())
	})
})

var _ = Describe("oneOf envelope projection failures", func() {

	// Every case here must fail the artifact. A union that cannot be projected is
	// never skipped: silently dropping it is not an acceptable outcome.
	expectProjectionError := func(schema *Schema) *OneOfProjectionError {
		GinkgoHelper()
		tree, diags, err := BuildResponseTree(schema)
		Expect(tree).To(BeNil())
		Expect(diags).To(BeEmpty(), "a failed projection must not be reported as a droppable diagnostic")

		var projection *OneOfProjectionError
		Expect(errors.As(err, &projection)).To(BeTrue(), "expected *OneOfProjectionError, got %T: %v", err, err)
		return projection
	}

	// at wraps a union as the "choice" property of a response object, so the
	// projection reports the position the union actually occupies in the tree.
	at := func(union *Schema) *Schema {
		return objSchema(map[string]*Schema{"choice": union})
	}

	It("fails a union with no non-null alternative to select", func() {
		err := expectProjectionError(at(oneOfSchema("response.choice", "Choice")))
		Expect(err.Envelope).To(Equal("Choice"))
		Expect(err.Path).To(Equal("response.choice"))
		Expect(err.Error()).To(ContainSubstring("no non-null alternative"))
	})

	It("fails a union the parser left without an envelope name", func() {
		err := expectProjectionError(at(oneOfSchema("response.choice", "", primitiveOneOfVariant("string", "string"))))
		Expect(err.Path).To(Equal("response.choice"))
		Expect(err.Error()).To(ContainSubstring("no generated envelope name"))
	})

	It("fails an alternative with no stable Terraform name", func() {
		err := expectProjectionError(at(oneOfSchema(
			"response.choice",
			"Choice",
			OneOfVariant{Schema: primSchema("string")},
		)))
		Expect(err.Path).To(Equal("response.choice"))
		Expect(err.Error()).To(ContainSubstring("no stable Terraform variant name"))
	})

	It("names the envelope, alternative and path when an alternative is unrepresentable", func() {
		err := expectProjectionError(objSchema(map[string]*Schema{
			"choice": oneOfSchema(
				"response.choice",
				"Choice",
				primitiveOneOfVariant("string", "string"),
				OneOfVariant{
					TFName: "broken",
					GoName: "Broken",
					Schema: &Schema{Kind: SchemaKindUnsupported, UnsupportedReason: "anyOf has no Terraform equivalent"},
				},
			),
		}))
		Expect(err.Envelope).To(Equal("Choice"))
		Expect(err.Variant).To(Equal("broken"))
		Expect(err.Path).To(Equal("response.choice"))

		// The underlying per-node failure stays reachable, so the reason the
		// alternative could not be represented survives to the run report.
		var unsupported *UnsupportedKindError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
		Expect(unsupported.Path).To(Equal("response.choice.broken.value"))
		Expect(err.Error()).To(ContainSubstring("anyOf has no Terraform equivalent"))
	})

	It("fails a one_of node the parser left without a normalized union", func() {
		tree, _, err := BuildResponseTree(objSchema(map[string]*Schema{
			"choice": {Kind: SchemaKindOneOf},
		}))
		Expect(tree).To(BeNil())

		var unsupported *UnsupportedKindError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
		Expect(unsupported.Path).To(Equal("response.choice"))
		Expect(unsupported.Reason).To(ContainSubstring("no normalized union"))
	})
})
