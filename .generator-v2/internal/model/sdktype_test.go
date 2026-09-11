package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These specs pin the port of formatter.simple_type against the pinned SDK
// generator's actual behaviour, including the spellings a lookup-based spike
// inferred wrongly — a derivation that is wrong produces a plausible type
// rather than nothing. They travelled here with the rule itself, which three
// callers now share (T137a).

var _ = Describe("SDKScalarGoType (port of formatter.simple_type)", func() {
	DescribeTable("primitive Go spellings",
		func(openapiType, format, want string, ok bool) {
			got, derived := SDKScalarGoType(&Schema{Type: openapiType, Format: format})
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
		got, ok := SDKScalarGoType(&Schema{Type: "string"})
		Expect(ok).To(BeTrue())
		Expect(got).To(Equal("string"))
	})
})

var _ = Describe("SDKEnumFromValueConstructor (port of model_enum.j2)", func() {
	DescribeTable("the constructor name the SDK declares for an enum component",
		func(goType, want string, wantOK bool) {
			got, ok := SDKEnumFromValueConstructor(goType)
			Expect(ok).To(Equal(wantOK))
			Expect(got).To(Equal(want))
		},
		// The two real path-parameter enums that reached this rule.
		Entry("AssetType", "AssetType", "NewAssetTypeFromValue", true),
		Entry("RumPermanentRetentionFilterID",
			"RumPermanentRetentionFilterID", "NewRumPermanentRetentionFilterIDFromValue", true),
		// The SDK declares New<T>FromValue only for a named component, so every
		// composite spelling has no constructor rather than an invented one.
		Entry("a qualified name", "datadogV2.AssetType", "", false),
		Entry("a slice", "[]AssetType", "", false),
		Entry("a pointer", "*AssetType", "", false),
		Entry("the empty interface", "interface{}", "", false),
		Entry("an empty spelling", "", "", false),
	)
})
