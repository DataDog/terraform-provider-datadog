package model

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// passwordShape is how one role declares the credential the three
// IntegrationAccountBasicAuth components disagree about: the request requires
// it, the update accepts it, the response does not have it at all. A single
// parameter rather than two booleans, because "absent but required" is not a
// state any body can be in.
type passwordShape int

const (
	noPassword passwordShape = iota
	optionalPassword
	requiredPassword
)

// basicAuth is one role's spelling of the shared "basic auth" alternative,
// modelled on the pinned spec's IntegrationAccountBasicAuth{Request,Update,
// Response}.
func basicAuth(refName string, password passwordShape) *Schema {
	properties := map[string]*Schema{
		"auth_type": {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"basic"}},
		"username":  {Kind: SchemaKindPrimitive, Type: "string"},
	}
	required := []string{"auth_type", "username"}
	if password != noPassword {
		properties["password"] = &Schema{Kind: SchemaKindPrimitive, Type: "string", Sensitive: true}
		if password == requiredPassword {
			required = append(required, "password")
		}
	}
	return &Schema{Kind: SchemaKindObject, RefName: refName, Properties: properties, Required: required}
}

// authUnion is one role's spelling of the authentication union, already
// carrying the SDK binding internal/sdkbind would have resolved for that role.
func authUnion(unionRefName, altRefName string, alternative *Schema, optional bool) *Schema {
	tfName := SnakeCase(altRefName)
	return &Schema{
		Kind:    SchemaKindOneOf,
		RefName: unionRefName,
		OneOf: &OneOfSpec{
			Name:     unionRefName,
			Path:     "data.attributes.authentication",
			RefName:  unionRefName,
			SDKType:  unionRefName,
			Optional: optional,
			Variants: []OneOfVariant{{
				TFName:         tfName,
				GoName:         SdkName(tfName),
				Schema:         alternative,
				RefName:        altRefName,
				SDKField:       altRefName,
				SDKConstructor: altRefName + "As" + unionRefName,
				SDKPointer:     true,
			}},
		},
	}
}

func authGroup(create, update, read *Schema) *ResolvedGroup {
	return &ResolvedGroup{
		Create: &Operation{OperationId: "CreateAccount", RequestSchema: jsonAPIBody(
			"AccountCreateRequest", map[string]*Schema{"authentication": create}, []string{"authentication"})},
		Update: &Operation{OperationId: "UpdateAccount", RequestSchema: jsonAPIBody(
			"AccountUpdateRequest", map[string]*Schema{"authentication": update}, nil)},
		Read: &Operation{OperationId: "GetAccount", ResponseSchema: jsonAPIBody(
			"AccountResponse", map[string]*Schema{"authentication": read}, nil)},
	}
}

// threeRoleAuthGroup is the shape T099 was widened for: one logical union
// spelled as three components, each wrapping its own spelling of one logical
// alternative.
func threeRoleAuthGroup() *ResolvedGroup {
	return authGroup(
		authUnion("AccountAuthenticationRequest", "IntegrationAccountBasicAuthRequest",
			basicAuth("IntegrationAccountBasicAuthRequest", requiredPassword), false),
		authUnion("AccountAuthenticationUpdate", "IntegrationAccountBasicAuthUpdate",
			basicAuth("IntegrationAccountBasicAuthUpdate", optionalPassword), true),
		authUnion("AccountAuthenticationResponse", "IntegrationAccountBasicAuthResponse",
			basicAuth("IntegrationAccountBasicAuthResponse", noPassword), true),
	)
}

var _ = Describe("MergeResourceSchema over a oneOf", func() {
	It("unions each correlated alternative's own properties instead of cloning the preferred body's", func() {
		merged, _, err := MergeResourceSchema(threeRoleAuthGroup())
		Expect(err).NotTo(HaveOccurred())

		union := attributesOf(merged)["authentication"]
		Expect(union.Kind).To(Equal(SchemaKindOneOf))
		Expect(union.OneOf.Variants).To(HaveLen(1))
		alternative := union.OneOf.Variants[0].Schema

		By("the request-only password survives, which cloning the Read response would have dropped")
		Expect(alternative.Properties).To(HaveKey("password"))
		Expect(alternative.Properties["password"].Provenance).To(Equal(
			&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: false}))

		By("a property every body carries is stamped as such")
		Expect(alternative.Properties["username"].Provenance).To(Equal(
			&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: true}))

		By("required-ness comes from the Create alternative alone, as everywhere else")
		Expect(alternative.Required).To(Equal([]string{"auth_type", "password", "username"}))

		By("the union node itself is stamped")
		Expect(union.Provenance).To(Equal(
			&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: true}))
	})

	It("publishes one role-independent name for the union and for each variant", func() {
		merged, _, err := MergeResourceSchema(threeRoleAuthGroup())
		Expect(err).NotTo(HaveOccurred())

		union := attributesOf(merged)["authentication"]
		Expect(union.OneOf.Name).To(Equal("AccountAuthentication"))
		Expect(union.OneOf.Variants[0].TFName).To(Equal("integration_account_basic_auth"))
		Expect(union.OneOf.Variants[0].GoName).To(Equal("IntegrationAccountBasicAuth"))
	})

	It("keeps the Read response's SDK binding on the merged node, leaving the request roles to read their own", func() {
		group := threeRoleAuthGroup()
		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())

		union := attributesOf(merged)["authentication"]
		Expect(union.OneOf.SDKType).To(Equal("AccountAuthenticationResponse"))
		Expect(union.OneOf.Variants[0].SDKConstructor).To(
			Equal("IntegrationAccountBasicAuthResponseAsAccountAuthenticationResponse"))

		By("the three per-role bindings are untouched by the merge, which is where emit reads them from")
		createUnion := group.Create.RequestSchema.Properties["data"].Properties["attributes"].Properties["authentication"]
		Expect(createUnion.OneOf.SDKType).To(Equal("AccountAuthenticationRequest"))
		Expect(createUnion.OneOf.Variants[0].SDKConstructor).To(
			Equal("IntegrationAccountBasicAuthRequestAsAccountAuthenticationRequest"))
		updateUnion := group.Update.RequestSchema.Properties["data"].Properties["attributes"].Properties["authentication"]
		Expect(updateUnion.OneOf.SDKType).To(Equal("AccountAuthenticationUpdate"))
	})

	It("permits absence when any body permits it, and OR-s nullability the same way", func() {
		group := threeRoleAuthGroup()
		group.Read.ResponseSchema.Properties["data"].
			Properties["attributes"].Properties["authentication"].OneOf.Nullable = true

		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		union := attributesOf(merged)["authentication"].OneOf
		// Create requires it and is not nullable; Update and Read say otherwise.
		Expect(union.Optional).To(BeTrue())
		Expect(union.Nullable).To(BeTrue())
	})

	It("publishes the bodies' own name verbatim when they already agree on it", func() {
		// Every body spelling the alternative identically means the name is
		// already role-independent, so stripping it could only lose
		// information — an alternative legitimately ending in "Update"
		// survives exactly here.
		group := authGroup(
			authUnion("AccountAuthentication", "BasicAuthUpdate", basicAuth("BasicAuthUpdate", requiredPassword), false),
			authUnion("AccountAuthentication", "BasicAuthUpdate", basicAuth("BasicAuthUpdate", optionalPassword), true),
			authUnion("AccountAuthentication", "BasicAuthUpdate", basicAuth("BasicAuthUpdate", noPassword), true),
		)
		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		Expect(attributesOf(merged)["authentication"].OneOf.Variants[0].TFName).To(Equal("basic_auth_update"))
	})

	It("strips a whole run of role markers, not just the last one", func() {
		// downtime spells one alternative three ways, and only removing both
		// markers reduces them to a common stem (T099c).
		group := authGroup(
			authUnion("ScheduleCreateRequest", "RecurrencesCreateRequest",
				basicAuth("RecurrencesCreateRequest", requiredPassword), false),
			authUnion("ScheduleUpdateRequest", "RecurrencesUpdateRequest",
				basicAuth("RecurrencesUpdateRequest", optionalPassword), true),
			authUnion("ScheduleResponse", "RecurrencesResponse",
				basicAuth("RecurrencesResponse", noPassword), true),
		)
		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		union := attributesOf(merged)["authentication"].OneOf
		Expect(union.Name).To(Equal("Schedule"))
		Expect(union.Variants[0].TFName).To(Equal("recurrences"))
	})

	It("merges a union only one body reaches without asking for a correlation", func() {
		// The request bodies carry an unrelated field, so they reach the
		// attributes object but not the union.
		plain := &Schema{Kind: SchemaKindPrimitive, Type: "string"}
		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateAccount", RequestSchema: jsonAPIBody(
				"AccountCreateRequest", map[string]*Schema{"name": plain}, nil)},
			Update: &Operation{OperationId: "UpdateAccount", RequestSchema: jsonAPIBody(
				"AccountUpdateRequest", map[string]*Schema{"name": plain}, nil)},
			Read: &Operation{OperationId: "GetAccount", ResponseSchema: jsonAPIBody(
				"AccountResponse", map[string]*Schema{
					"name": plain,
					"authentication": authUnion(
						"AccountAuthenticationResponse", "IntegrationAccountBasicAuthResponse",
						basicAuth("IntegrationAccountBasicAuthResponse", noPassword), true),
				}, nil)},
		}

		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		union := attributesOf(merged)["authentication"]
		// One body cannot disagree with itself, so its own name stands.
		Expect(union.OneOf.Variants[0].TFName).To(Equal("integration_account_basic_auth_response"))
		Expect(union.Provenance).To(Equal(
			&SchemaProvenance{InRequest: false, RequestRequired: false, InResponse: true}))
	})

	It("fails at the union's path when the bodies list different alternatives", func() {
		group := threeRoleAuthGroup()
		readUnion := group.Read.ResponseSchema.Properties["data"].Properties["attributes"].Properties["authentication"]
		readUnion.OneOf.Variants = append(readUnion.OneOf.Variants, OneOfVariant{
			TFName: "token_auth", GoName: "TokenAuth",
			Schema: &Schema{Kind: SchemaKindObject, RefName: "TokenAuth"},
		})

		_, _, err := MergeResourceSchema(group)
		var conflict *OneOfMergeError
		Expect(errors.As(err, &conflict)).To(BeTrue())
		Expect(conflict.Path).To(Equal("data.attributes.authentication"))
		By("the bodies are quoted as they spell themselves, not as the correlation reduced them")
		Expect(conflict.Create).To(Equal([]string{"integration_account_basic_auth_request"}))
		Expect(conflict.Read).To(Equal([]string{"integration_account_basic_auth_response", "token_auth"}))
		Expect(conflict.Error()).To(ContainSubstring("do not list the same alternatives"))
	})

	It("fails when one body's own alternatives collapse onto a single role-independent name", func() {
		group := threeRoleAuthGroup()
		createUnion := group.Create.RequestSchema.Properties["data"].Properties["attributes"].Properties["authentication"]
		createUnion.OneOf.Variants = append(createUnion.OneOf.Variants, OneOfVariant{
			// "…_update" strips to the same stem as "…_request".
			TFName: "integration_account_basic_auth_update", GoName: "IntegrationAccountBasicAuthUpdate",
			Schema: basicAuth("IntegrationAccountBasicAuthUpdate", optionalPassword),
		})

		_, _, err := MergeResourceSchema(group)
		var conflict *OneOfMergeError
		Expect(errors.As(err, &conflict)).To(BeTrue())
		Expect(conflict.Error()).To(ContainSubstring(`role-independent name is "integration_account_basic_auth"`))
	})

	It("reports an alternative no body gives a normalized schema, rather than cloning nil", func() {
		group := threeRoleAuthGroup()
		for _, body := range []*Schema{
			group.Create.RequestSchema, group.Update.RequestSchema, group.Read.ResponseSchema,
		} {
			union := body.Properties["data"].Properties["attributes"].Properties["authentication"]
			union.OneOf.Variants[0].Schema = nil
		}

		_, _, err := MergeResourceSchema(group)
		var conflict *OneOfMergeError
		Expect(errors.As(err, &conflict)).To(BeTrue())
		Expect(conflict.Error()).To(ContainSubstring("no normalized schema in any body"))
	})

	It("hands a union carrying no normalized spec straight through, so the projection reports it", func() {
		group := authGroup(
			&Schema{Kind: SchemaKindOneOf, RefName: "AccountAuthenticationRequest"},
			&Schema{Kind: SchemaKindOneOf, RefName: "AccountAuthenticationUpdate"},
			&Schema{Kind: SchemaKindOneOf, RefName: "AccountAuthenticationResponse"},
		)
		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		Expect(attributesOf(merged)["authentication"].OneOf).To(BeNil())
	})
})

var _ = Describe("StripOneOfRoleSuffix", func() {
	DescribeTable("removes a trailing role suffix in either casing",
		func(in, want string) { Expect(StripOneOfRoleSuffix(in)).To(Equal(want)) },
		Entry("PascalCase request", "IntegrationAccountBasicAuthRequest", "IntegrationAccountBasicAuth"),
		Entry("PascalCase update", "AccountAuthenticationUpdate", "AccountAuthentication"),
		Entry("PascalCase response", "AccountAuthenticationResponse", "AccountAuthentication"),
		Entry("PascalCase create", "AccountAuthenticationCreate", "AccountAuthentication"),
		Entry("snake_case", "integration_account_basic_auth_response", "integration_account_basic_auth"),
		Entry("no suffix", "AwsIntegration", "AwsIntegration"),
		Entry("suffix is the whole name", "Request", "Request"),
		Entry("snake suffix is the whole name", "request", "request"),
		Entry("empty", "", ""),
	)
})
