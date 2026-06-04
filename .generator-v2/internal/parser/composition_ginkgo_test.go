package parser

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// Each Context below pins exactly one acceptance criterion of the allOf
// flattener (APIR-2906); the It sentences are the criteria restated as
// behavior. Fixtures are real OpenAPI fragments so branch $refs resolve the way
// they do in a parsed spec — which is what lets a collision name its refs.
var _ = Describe("Flatten", func() {
	// schemaFrom parses the body of components.schemas and returns the named
	// component's resolved schema, so every fixture reads as the YAML a real
	// spec would carry.
	schemaFrom := func(schemasYAML, name string) *base.Schema {
		GinkgoHelper()
		spec := "openapi: 3.0.0\ninfo:\n  title: t\n  version: \"1\"\ncomponents:\n  schemas:\n" + schemasYAML
		doc, err := libopenapi.NewDocument([]byte(spec))
		Expect(err).NotTo(HaveOccurred())
		v3, err := doc.BuildV3Model()
		Expect(err).NotTo(HaveOccurred())
		proxy := v3.Model.Components.Schemas.GetOrZero(name)
		Expect(proxy).NotTo(BeNil(), "fixture has no component schema %q", name)
		return proxy.Schema()
	}

	asCollisionErr := func(err error) *AllOfCollisionError {
		GinkgoHelper()
		var collision *AllOfCollisionError
		Expect(errors.As(err, &collision)).To(BeTrue(), "error %v (%T) is not a *AllOfCollisionError", err, err)
		return collision
	}

	// ---- AC1: union of properties, Required sorted ascending -----------------

	Context("when two branches declare disjoint properties", func() {
		const fixture = `
    Base:
      type: object
      properties:
        name: {type: string}
        id: {type: string}
    Extra:
      type: object
      properties:
        age: {type: integer}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`

		It("flattens to one object schema holding the union of every branch's properties", func() {
			flat, err := Flatten(schemaFrom(fixture, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			Expect(flat.Kind).To(Equal(model.SchemaKindObject))
			Expect(flat.Properties).To(HaveLen(3))
			Expect(flat.Properties).To(HaveKey("name"))
			Expect(flat.Properties).To(HaveKey("id"))
			Expect(flat.Properties).To(HaveKey("age"))
		})

		It("returns the required list sorted ascending, independent of the order each branch declared it", func() {
			withRequired := `
    Base:
      type: object
      required: [name, id]
      properties:
        name: {type: string}
        id: {type: string}
    Extra:
      type: object
      required: [age]
      properties:
        age: {type: integer}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`
			flat, err := Flatten(schemaFrom(withRequired, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			Expect(flat.Required).To(Equal([]string{"age", "id", "name"}))
		})
	})

	// ---- AC2: identical redeclaration merges silently ------------------------

	Context("when several branches redeclare the same property identically", func() {
		const fixture = `
    Base:
      type: object
      properties:
        id: {type: string, format: uuid}
    Extra:
      type: object
      properties:
        id: {type: string, format: uuid}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`

		It("merges the duplicate into a single property, preserving its type and constraints, without error", func() {
			flat, err := Flatten(schemaFrom(fixture, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			Expect(flat.Properties).To(HaveLen(1))
			Expect(flat.Properties["id"].Type).To(Equal("string"))
			Expect(flat.Properties["id"].Format).To(Equal("uuid"))
		})
	})

	// ---- AC3: requiredness is OR'd across branches ---------------------------

	Context("when branches agree on a property but disagree on its required flag", func() {
		const fixture = `
    Base:
      type: object
      required: [id]
      properties:
        id: {type: string}
        note: {type: string}
    Extra:
      type: object
      properties:
        id: {type: string}
        note: {type: string}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`

		It("marks a property required when any branch requires it, and optional when none does, without error", func() {
			flat, err := Flatten(schemaFrom(fixture, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			Expect(flat.Required).To(ContainElement("id"))    // Base required it
			Expect(flat.Required).NotTo(ContainElement("note")) // neither branch did
		})
	})

	// ---- AC4: incompatible types are a typed collision -----------------------

	Context("when two branches declare the same property with incompatible types", func() {
		const fixture = `
    Base:
      type: object
      properties:
        id: {type: string}
    Extra:
      type: object
      properties:
        id: {type: integer}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`

		It("returns a typed error naming the property, both types, and both contributing branch refs", func() {
			_, err := Flatten(schemaFrom(fixture, "Composed"))
			collision := asCollisionErr(err)
			Expect(collision.Collisions).To(HaveLen(1))
			Expect(collision.Collisions[0].Property).To(Equal("id"))
			Expect(collision.Collisions[0].Branches).To(HaveLen(2))
			Expect(collision.Error()).To(SatisfyAll(
				ContainSubstring("id"),
				ContainSubstring("string"),
				ContainSubstring("integer"),
				ContainSubstring("Base"),
				ContainSubstring("Extra"),
			))
		})
	})

	// ---- AC5: same type, differing constraints are a typed collision ---------

	Context("when two branches declare a property with the same type but differing constraints", func() {
		It("errors on conflicting formats, naming the property and both branch refs", func() {
			fixture := `
    Base:
      type: object
      properties:
        id: {type: string, format: uuid}
    Extra:
      type: object
      properties:
        id: {type: string, format: int64}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`
			_, err := Flatten(schemaFrom(fixture, "Composed"))
			collision := asCollisionErr(err)
			Expect(collision.Collisions[0].Property).To(Equal("id"))
			Expect(collision.Error()).To(SatisfyAll(
				ContainSubstring("id"),
				ContainSubstring("uuid"),
				ContainSubstring("int64"),
				ContainSubstring("Base"),
				ContainSubstring("Extra"),
			))
		})

		It("errors on conflicting enums, naming the property and both branch refs", func() {
			fixture := `
    Base:
      type: object
      properties:
        status: {type: string, enum: [active, paused]}
    Extra:
      type: object
      properties:
        status: {type: string, enum: [active, deleted]}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`
			_, err := Flatten(schemaFrom(fixture, "Composed"))
			collision := asCollisionErr(err)
			Expect(collision.Collisions[0].Property).To(Equal("status"))
			Expect(collision.Collisions[0].Branches).To(HaveLen(2))
			Expect(collision.Error()).To(SatisfyAll(
				ContainSubstring("status"),
				ContainSubstring("Base"),
				ContainSubstring("Extra"),
			))
		})

		It("aggregates every colliding property into one error, sorted ascending by property name", func() {
			fixture := `
    Base:
      type: object
      properties:
        id: {type: string}
        name: {type: string, format: uuid}
    Extra:
      type: object
      properties:
        id: {type: integer}
        name: {type: string, format: int64}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`
			_, err := Flatten(schemaFrom(fixture, "Composed"))
			collision := asCollisionErr(err)
			Expect(collision.Collisions).To(HaveLen(2))
			Expect(collision.Collisions[0].Property).To(Equal("id"))
			Expect(collision.Collisions[1].Property).To(Equal("name"))
		})
	})

	// ---- AC6: nested allOf flattens recursively ------------------------------

	Context("when an allOf branch itself contains an allOf", func() {
		It("produces the same result as the equivalent flat composition", func() {
			flatComposition := `
    A: {type: object, properties: {a: {type: string}}}
    B: {type: object, properties: {b: {type: string}}}
    C: {type: object, properties: {c: {type: string}}}
    Root:
      allOf:
        - $ref: '#/components/schemas/A'
        - $ref: '#/components/schemas/B'
        - $ref: '#/components/schemas/C'
`
			nestedComposition := `
    A: {type: object, properties: {a: {type: string}}}
    B: {type: object, properties: {b: {type: string}}}
    C: {type: object, properties: {c: {type: string}}}
    Root:
      allOf:
        - allOf:
            - $ref: '#/components/schemas/A'
            - $ref: '#/components/schemas/B'
        - $ref: '#/components/schemas/C'
`
			flat, err := Flatten(schemaFrom(flatComposition, "Root"))
			Expect(err).NotTo(HaveOccurred())
			nested, err := Flatten(schemaFrom(nestedComposition, "Root"))
			Expect(err).NotTo(HaveOccurred())
			Expect(nested).To(Equal(flat))
		})
	})

	// ---- AC7: max-depth bounds the recursion ---------------------------------

	Context("when nested allOf is deeper than the configured max depth", func() {
		const deeplyNested = `
    Leaf: {type: object, properties: {x: {type: string}}}
    Root:
      allOf:
        - allOf:
            - allOf:
                - $ref: '#/components/schemas/Leaf'
`

		It("errors when nesting exceeds the bound, and the message names the depth limit", func() {
			_, err := Flatten(schemaFrom(deeplyNested, "Root"), WithMaxDepth(1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("1"))
		})

		It("flattens successfully when the same nesting is within the default bound", func() {
			flat, err := Flatten(schemaFrom(deeplyNested, "Root"))
			Expect(err).NotTo(HaveOccurred())
			Expect(flat.Properties).To(HaveKey("x"))
		})
	})

	// ---- AC8: deterministic, order-independent output ------------------------
	// (Literal idempotency Flatten(Flatten(s)) is ill-typed here — input is a
	// *base.Schema, output a *model.Schema — so determinism is the property the
	// golden-file pipeline actually depends on.)

	Context("when the same composition is flattened more than once", func() {
		const fixture = `
    Base:
      type: object
      properties:
        name: {type: string}
        id: {type: string}
    Extra:
      type: object
      properties:
        age: {type: integer}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Base'
        - $ref: '#/components/schemas/Extra'
`

		It("returns deeply-equal output on repeated runs over the same schema", func() {
			first, err := Flatten(schemaFrom(fixture, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			second, err := Flatten(schemaFrom(fixture, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			Expect(second).To(Equal(first))
		})

		It("returns the same output regardless of the order the branches are declared", func() {
			swapped := `
    Base:
      type: object
      properties:
        name: {type: string}
        id: {type: string}
    Extra:
      type: object
      properties:
        age: {type: integer}
    Composed:
      allOf:
        - $ref: '#/components/schemas/Extra'
        - $ref: '#/components/schemas/Base'
`
			original, err := Flatten(schemaFrom(fixture, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			reordered, err := Flatten(schemaFrom(swapped, "Composed"))
			Expect(err).NotTo(HaveOccurred())
			Expect(reordered).To(Equal(original))
		})
	})
})
