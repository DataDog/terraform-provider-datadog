package model

import (
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

	It("keeps ordered non-null alternatives and their Terraform/SDK bindings", func() {
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

	It("keeps null separate from the non-null variant list", func() {
		union := &OneOfSpec{
			Nullable: true,
			Variants: []OneOfVariant{
				{TFName: "boolean", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "boolean"}},
				{TFName: "string", Schema: &Schema{Kind: SchemaKindPrimitive, Type: "string"}},
			},
		}

		Expect(union.Nullable).To(BeTrue())
		Expect(union.Variants).To(HaveLen(2))
		Expect(union.Variants).To(ConsistOf(
			HaveField("TFName", "boolean"),
			HaveField("TFName", "string"),
		))
	})

	It("retains the legacy kind alias during parser migration", func() {
		Expect(SchemaKindVariant).To(Equal(SchemaKindOneOf))
	})
})
