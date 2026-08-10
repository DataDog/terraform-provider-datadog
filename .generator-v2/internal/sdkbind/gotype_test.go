package sdkbind

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// These specs pin the port of formatter.simple_type and of model_oneof.j2's member
// rule against the pinned SDK generator's actual behaviour. They
// deliberately include the four spellings a lookup-based spike inferred wrongly,
// since a derivation that is wrong produces a plausible name rather than nothing.

var _ = Describe("simpleType (port of formatter.simple_type)", func() {
	DescribeTable("primitive Go spellings",
		func(openapiType, format, want string, ok bool) {
			got, derived := simpleType(&model.Schema{Type: openapiType, Format: format})
			Expect(derived).To(Equal(ok))
			Expect(got).To(Equal(want))
		},
		// integer: an unformatted integer is int32, NOT int64.
		Entry("integer, no format", "integer", "", "int32", true),
		Entry("integer/int32", "integer", "int32", "int32", true),
		Entry("integer/int64", "integer", "int64", "int64", true),
		// number: an unformatted number is float; only double is float64.
		Entry("number, no format", "number", "", "float", true),
		Entry("number/double", "number", "double", "float64", true),
		// string: date and date-time both map to time.Time.
		Entry("string, no format", "string", "", "string", true),
		Entry("string/date", "string", "date", "time.Time", true),
		Entry("string/date-time", "string", "date-time", "time.Time", true),
		Entry("string/uuid", "string", "uuid", "uuid.UUID", true),
		Entry("string/binary", "string", "binary", "_io.Reader", true),
		// An unmapped string format falls back to string (.get default).
		Entry("string/email falls back", "string", "email", "string", true),
		Entry("boolean", "boolean", "", "bool", true),
		// Where the Python raises KeyError the SDK cannot generate a type at all.
		Entry("integer with unmapped format", "integer", "int16", "", false),
		Entry("number with unmapped format", "number", "float", "", false),
		Entry("object is not a simple type", "object", "", "", false),
		Entry("untyped is not a simple type", "", "", "", false),
	)

	It("does not spell a nullable alternative as datadog.Nullable", func() {
		// model_oneof.j2 calls get_type(oneOf) with no render_nullable, so the
		// Nullable prefix never reaches a wrapper member. Nullability is carried on
		// OneOfSpec.Nullable and represented by an absent envelope instead.
		got, ok := simpleType(&model.Schema{Type: "string"})
		Expect(ok).To(BeTrue())
		Expect(got).To(Equal("string"))
	})
})

var _ = Describe("memberBinding (port of model_oneof.j2's member rule)", func() {
	It("names a referenced alternative after its component, not its Terraform name", func() {
		// The SDK's get_name wins over the Go type spelling, and acronym casing is
		// the component's own: AWSIntegration, never AwsIntegration.
		name, pointer, err := memberBinding(model.OneOfVariant{
			TFName:  "aws_integration",
			RefName: "AWSIntegration",
			Schema:  &model.Schema{Kind: model.SchemaKindObject, RefName: "AWSIntegration"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("AWSIntegration"))
		Expect(pointer).To(BeTrue())
	})

	It("prefers the component name even where the alternative is a primitive", func() {
		// get_name is checked before the type spelling, so a component that happens
		// to be `type: string` binds to its own name.
		name, _, err := memberBinding(model.OneOfVariant{
			TFName:  "connection_id",
			RefName: "ConnectionId",
			Schema:  &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", RefName: "ConnectionId"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(Equal("ConnectionId"))
	})

	DescribeTable("names an anonymous primitive after its upperfirst Go spelling",
		func(openapiType, format, want string) {
			name, pointer, err := memberBinding(model.OneOfVariant{
				TFName: "value",
				Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: openapiType, Format: format},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal(want))
			Expect(pointer).To(BeTrue())
		},
		Entry("string", "string", "", "String"),
		Entry("boolean", "boolean", "", "Bool"),
		Entry("integer", "integer", "", "Int32"),
		Entry("integer/int64", "integer", "int64", "Int64"),
		Entry("number", "number", "", "Float"),
		Entry("number/double", "number", "double", "Float64"),
	)

	It("leaves a free-form object member unpointered", func() {
		// isAdditionalPropertiesContainer: the SDK emits it as a bare map because a
		// map is already nil-able, so a mapper must not take its address.
		_, pointer, err := memberBinding(model.OneOfVariant{
			TFName:  "metadata",
			RefName: "FreeFormMetadata",
			Schema:  &model.Schema{Kind: model.SchemaKindMap, RefName: "FreeFormMetadata"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(pointer).To(BeFalse())
	})

	It("refuses an anonymous alternative whose Go spelling is not an identifier", func() {
		// upperfirst("time.Time") is "Time.Time", which the SDK cannot declare as a
		// member — so no such member exists and a derived name would be a lie.
		_, _, err := memberBinding(model.OneOfVariant{
			TFName: "date_time",
			Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("time.Time"))
		Expect(err.Error()).To(ContainSubstring("not a valid Go identifier"))
		Expect(err.Error()).To(ContainSubstring("$ref"))
	})

	It("refuses an anonymous enum alternative", func() {
		_, _, err := memberBinding(model.OneOfVariant{
			TFName: "status",
			Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"a", "b"}},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("anonymous enum"))
	})

	It("refuses an anonymous object alternative", func() {
		_, _, err := memberBinding(model.OneOfVariant{
			TFName: "inline",
			Schema: &model.Schema{Kind: model.SchemaKindObject},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no SDK member name"))
		Expect(err.Error()).To(ContainSubstring("object"))
	})

	It("names the offending type and format when a format has no SDK spelling", func() {
		// formatter.simple_type raises KeyError here, so the SDK cannot generate the
		// wrapper. The diagnostic has to say which node, by its OpenAPI spelling.
		_, _, err := memberBinding(model.OneOfVariant{
			TFName: "count",
			Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int16"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("integer/int16"))
	})

	It("refuses an anonymous list alternative", func() {
		// upperfirst("[]Foo") is not a Go identifier, so no such member exists.
		_, _, err := memberBinding(model.OneOfVariant{
			TFName: "items",
			Schema: &model.Schema{Kind: model.SchemaKindArray},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("array"))
	})

	It("reports a missing normalized schema rather than panicking", func() {
		_, _, err := memberBinding(model.OneOfVariant{TFName: "broken"})
		Expect(err).To(MatchError(ContainSubstring("no normalized schema")))
	})
})

var _ = Describe("UpperFirst", func() {
	DescribeTable("upper-cases only the first rune",
		func(in, want string) { Expect(model.UpperFirst(in)).To(Equal(want)) },
		Entry("lower-case type spelling", "string", "String"),
		Entry("already upper", "AWSIntegration", "AWSIntegration"),
		Entry("preserves inner acronym", "int32", "Int32"),
		Entry("empty", "", ""),
	)
})
