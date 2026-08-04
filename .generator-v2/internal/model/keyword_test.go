package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Kept out of identifier_test.go, which is still one of the bare-testing files
// slated for Ginkgo conversion; new specs go in Ginkgo form from the start.
var _ = Describe("EscapeReservedKeyword", func() {
	DescribeTable("appends Var to a Go keyword and leaves anything else alone",
		func(in, want string) { Expect(EscapeReservedKeyword(in)).To(Equal(want)) },
		// The case that motivated this: a JSON:API "type" property.
		Entry("type", "type", "typeVar"),
		Entry("range", "range", "rangeVar"),
		Entry("func", "func", "funcVar"),
		Entry("interface", "interface", "interfaceVar"),
		Entry("var", "var", "varVar"),
		// Not keywords, however close they look.
		Entry("typeName", "typeName", "typeName"),
		Entry("types", "types", "types"),
		Entry("name", "name", "name"),
		Entry("empty", "", ""),
		// Exported spellings can never collide: no Go keyword is capitalized.
		Entry("Type", "Type", "Type"),
	)

	It("covers Go's whole reserved-word set", func() {
		// Mirrors formatter.KEYWORDS in the SDK generator; a short list here would
		// mean some property name still emits invalid Go.
		Expect(goKeywords).To(HaveLen(25))
	})
})
