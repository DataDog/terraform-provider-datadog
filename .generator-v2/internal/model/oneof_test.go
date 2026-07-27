package model

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("normalized oneOf model", func() {

	It("keeps the envelope identity, presence semantics, and discriminator metadata", func() {
		union := &Schema{
			Kind: SchemaKindOneOf,
			OneOf: &OneOfSpec{
				Name:     "ActionQueryMockedOutputsEnabled",
				Path:     "request.data.attributes.enabled",
				Optional: true,
				Nullable: true,
				Discriminator: &OneOfDiscriminator{
					PropertyName: "type",
					Mapping: map[string]string{
						"boolean": "#/components/schemas/BooleanEnabled",
						"string":  "#/components/schemas/StringEnabled",
					},
				},
			},
		}

		Expect(union.Kind).To(Equal(SchemaKindOneOf))
		Expect(union.OneOf.Name).To(Equal("ActionQueryMockedOutputsEnabled"))
		Expect(union.OneOf.Path).To(Equal("request.data.attributes.enabled"))
		Expect(union.OneOf.Optional).To(BeTrue())
		Expect(union.OneOf.Nullable).To(BeTrue())
		Expect(union.OneOf.Discriminator.PropertyName).To(Equal("type"))
		Expect(union.OneOf.Discriminator.Mapping).To(Equal(map[string]string{
			"boolean": "#/components/schemas/BooleanEnabled",
			"string":  "#/components/schemas/StringEnabled",
		}))
	})

	It("carries non-null alternatives and their Terraform/SDK bindings", func() {
		booleanSchema := &Schema{Kind: SchemaKindPrimitive, Type: "boolean"}
		objectSchema := &Schema{
			Kind: SchemaKindObject,
			Properties: map[string]*Schema{
				"name": {Kind: SchemaKindPrimitive, Type: "string"},
			},
		}
		union := &OneOfSpec{
			Name:     "MockedOutput",
			Path:     "response.data.attributes.output",
			Nullable: true,
			Variants: []OneOfVariant{
				{
					TFName:         "boolean",
					GoName:         "Boolean",
					Schema:         booleanSchema,
					RefName:        "BooleanMockedOutput",
					SDKField:       "Bool",
					SDKConstructor: "BoolAsActionQueryMockedOutputs",
					ValueWrapped:   true,
				},
				{
					TFName:         "object",
					GoName:         "Object",
					Schema:         objectSchema,
					RefName:        "ObjectMockedOutput",
					SDKField:       "ObjectMockedOutput",
					SDKConstructor: "ObjectMockedOutputAsActionQueryMockedOutputs",
					ValueWrapped:   false,
				},
			},
		}

		Expect(union.Variants).To(HaveLen(2))
		Expect(union.Variants[0]).To(Equal(OneOfVariant{
			TFName:         "boolean",
			GoName:         "Boolean",
			Schema:         booleanSchema,
			RefName:        "BooleanMockedOutput",
			SDKField:       "Bool",
			SDKConstructor: "BoolAsActionQueryMockedOutputs",
			ValueWrapped:   true,
		}))
		Expect(union.Variants[1]).To(Equal(OneOfVariant{
			TFName:         "object",
			GoName:         "Object",
			Schema:         objectSchema,
			RefName:        "ObjectMockedOutput",
			SDKField:       "ObjectMockedOutput",
			SDKConstructor: "ObjectMockedOutputAsActionQueryMockedOutputs",
			ValueWrapped:   false,
		}))
	})

	It("retains the legacy kind alias during parser migration", func() {
		Expect(SchemaKindVariant).To(Equal(SchemaKindOneOf))
	})
})

var _ = Describe("deterministic oneOf variant naming", func() {

	DescribeTable("uses the first meaningful candidate in the stable precedence order",
		func(candidates OneOfVariantNameCandidates, want string) {
			Expect(ResolveOneOfVariantName(candidates)).To(Equal(want))
		},
		Entry("discriminator mapping key wins",
			OneOfVariantNameCandidates{
				DiscriminatorKey: "HTTP Result",
				RefName:          "ReferencedResult",
				PrimitiveType:    "string",
				PrimitiveFormat:  "uuid",
				ShaSum:           "deadbeef",
			},
			"http_result",
		),
		Entry("referenced schema name wins over primitive type",
			OneOfVariantNameCandidates{
				RefName:       "ReferencedResult",
				PrimitiveType: "string",
				ShaSum:        "deadbeef",
			},
			"referenced_result",
		),
		Entry("primitive type includes its format",
			OneOfVariantNameCandidates{
				PrimitiveType:   "string",
				PrimitiveFormat: "uuid",
				ShaSum:          "deadbeef",
			},
			"string_uuid",
		),
		Entry("structural fingerprint is the final fallback",
			OneOfVariantNameCandidates{ShaSum: "deadbeef"},
			"variant_deadbeef",
		),
		Entry("an unusable reference name falls through to the primitive name",
			OneOfVariantNameCandidates{
				RefName:       "!!!",
				PrimitiveType: "boolean",
				ShaSum:        "deadbeef",
			},
			"boolean",
		),
		Entry("a numeric-leading name receives a Terraform-safe prefix",
			OneOfVariantNameCandidates{DiscriminatorKey: "123 result"},
			"variant_123_result",
		),
		Entry("empty candidates still produce a stable name",
			OneOfVariantNameCandidates{},
			"variant",
		),
	)

	It("rejects names that collide after normalization", func() {
		first := ResolveOneOfVariantName(OneOfVariantNameCandidates{DiscriminatorKey: "Foo Bar"})
		second := ResolveOneOfVariantName(OneOfVariantNameCandidates{DiscriminatorKey: "foo_bar"})
		Expect(first).To(Equal("foo_bar"))
		Expect(second).To(Equal(first))

		err := ValidateOneOfVariantNames("request.choice", []OneOfVariant{
			{TFName: first},
			{TFName: second},
		})
		Expect(err).To(HaveOccurred())

		var collision *OneOfVariantNameCollisionError
		Expect(errors.As(err, &collision)).To(BeTrue())
		Expect(collision.Path).To(Equal("request.choice"))
		Expect(collision.Name).To(Equal("foo_bar"))
	})

	It("reports the same collision regardless of alternative order", func() {
		forward := []OneOfVariant{
			{TFName: "zeta"}, {TFName: "zeta"},
			{TFName: "alpha"}, {TFName: "alpha"},
		}
		reversed := []OneOfVariant{
			{TFName: "alpha"}, {TFName: "alpha"},
			{TFName: "zeta"}, {TFName: "zeta"},
		}

		var first, second *OneOfVariantNameCollisionError
		Expect(errors.As(ValidateOneOfVariantNames("request", forward), &first)).To(BeTrue())
		Expect(errors.As(ValidateOneOfVariantNames("request", reversed), &second)).To(BeTrue())
		Expect(first).To(Equal(second))
		Expect(first.Name).To(Equal("alpha"))
	})
})
