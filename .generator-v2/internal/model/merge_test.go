package model

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// jsonAPIBody builds a JSON:API-shaped body — data.attributes.<field> — under
// its own top-level component name, so tests can give the Create request, the
// Update request and the Read response three different enclosing component
// names while still correlating by property position underneath them.
func jsonAPIBody(refName string, attributes map[string]*Schema, required []string) *Schema {
	return &Schema{
		Kind:    SchemaKindObject,
		RefName: refName,
		Properties: map[string]*Schema{
			"data": {
				Kind: SchemaKindObject,
				Properties: map[string]*Schema{
					"attributes": {
						Kind:       SchemaKindObject,
						Properties: attributes,
						Required:   required,
					},
				},
			},
		},
	}
}

func attributesOf(s *Schema) map[string]*Schema {
	return s.Properties["data"].Properties["attributes"].Properties
}

var _ = Describe("MergeResourceSchema", func() {
	It("correlates by property position across three differently-named components and stamps the FR-034a provenance bits", func() {
		createReq := jsonAPIBody("IncidentTypeCreateRequest", map[string]*Schema{
			"name":          {Kind: SchemaKindPrimitive, Type: "string", Description: "name (create)"},
			"is_default":    {Kind: SchemaKindPrimitive, Type: "boolean"},
			"internal_note": {Kind: SchemaKindPrimitive, Type: "string"},
			"priority":      {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"low", "high"}},
			"status":        {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"open", "closed"}},
		}, []string{"name"})

		// A PATCH body marking "priority" required is deliberately unusual: it
		// proves RequestRequired reads the Create body's Required list only.
		updateReq := jsonAPIBody("IncidentTypeUpdateRequest", map[string]*Schema{
			"name":       {Kind: SchemaKindPrimitive, Type: "string"},
			"is_default": {Kind: SchemaKindPrimitive, Type: "boolean"},
			"priority":   {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"low", "high"}},
		}, []string{"priority"})

		readResp := jsonAPIBody("IncidentTypeResponse", map[string]*Schema{
			"name":       {Kind: SchemaKindPrimitive, Type: "string", Description: "The incident type name."},
			"is_default": {Kind: SchemaKindPrimitive, Type: "boolean"},
			"created_at": {Kind: SchemaKindPrimitive, Type: "string", Format: "date-time"},
			"status":     {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"open", "closed", "archived"}, Sensitive: true},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateIncidentType", RequestSchema: createReq},
			Update: &Operation{OperationId: "UpdateIncidentType", RequestSchema: updateReq},
			Read:   &Operation{OperationId: "GetIncidentType", ResponseSchema: readResp},
		}

		merged, diags, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())

		attrs := attributesOf(merged)

		By("required in Create, present in Update, present in response -> Required")
		Expect(attrs["name"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: true}))
		// Cosmetic: descriptions differ, Read response wins.
		Expect(attrs["name"].Description).To(Equal("The incident type name."))

		By("optional in Create, present in Update, present in response -> Optional+Computed")
		Expect(attrs["is_default"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: false, InResponse: true}))

		By("Create-only, absent from Update and the response -> write-only Optional")
		Expect(attrs["internal_note"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: false, InResponse: false}))

		By("response-only -> Computed-only, and Format carries through untouched")
		Expect(attrs["created_at"].Provenance).To(Equal(&SchemaProvenance{InRequest: false, RequestRequired: false, InResponse: true}))
		Expect(attrs["created_at"].Format).To(Equal("date-time"))

		By("Update's Required entry is ignored: RequestRequired reads Create's Required list only")
		Expect(attrs["priority"].Provenance.RequestRequired).To(BeFalse())
		// Enum agrees everywhere it appears, so no cosmetic reconciliation is needed.
		Expect(attrs["priority"].Enum).To(Equal([]string{"high", "low"}))

		By("Enum members union (never intersect) and Sensitive is the disjunction")
		Expect(attrs["status"].Enum).To(Equal([]string{"archived", "closed", "open"}))
		Expect(attrs["status"].Sensitive).To(BeTrue())

		By("the root's own RefName differs across all three bodies; the Read response wins")
		Expect(merged.RefName).To(Equal("IncidentTypeResponse"))

		By("every reconciled cosmetic difference is recorded as an info diagnostic")
		Expect(diags).NotTo(BeEmpty())
		for _, d := range diags {
			Expect(d.Severity).To(Equal(SeverityInfo))
		}
	})

	It("merges with no Update operation at all", func() {
		createReq := jsonAPIBody("XCreateRequest", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
		}, []string{"name"})
		readResp := jsonAPIBody("XResponse", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateX", RequestSchema: createReq},
			Read:   &Operation{OperationId: "GetX", ResponseSchema: readResp},
		}

		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		Expect(attributesOf(merged)["name"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: true}))
	})

	It("never reads group.Search or a Create/Update response body", func() {
		createReq := jsonAPIBody("XCreateRequest", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
		}, nil)
		readResp := jsonAPIBody("XResponse", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{
				OperationId:   "CreateX",
				RequestSchema: createReq,
				// A response-only field must never surface as Computed state.
				ResponseSchema: jsonAPIBody("XCreateResponse", map[string]*Schema{
					"name":            {Kind: SchemaKindPrimitive, Type: "string"},
					"create_response": {Kind: SchemaKindPrimitive, Type: "string"},
				}, nil),
			},
			Read: &Operation{OperationId: "GetX", ResponseSchema: readResp},
			Search: &Operation{
				OperationId: "ListX",
				// A field only the search element carries must not leak in either.
				ResponseSchema: jsonAPIBody("XListItem", map[string]*Schema{
					"name":          {Kind: SchemaKindPrimitive, Type: "string"},
					"search_narrow": {Kind: SchemaKindPrimitive, Type: "string"},
				}, nil),
			},
		}

		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		attrs := attributesOf(merged)
		Expect(attrs).To(HaveKey("name"))
		Expect(attrs).NotTo(HaveKey("create_response"))
		Expect(attrs).NotTo(HaveKey("search_narrow"))
	})

	It("fails with SchemaMergeError naming the path and both spellings on a primitive type conflict", func() {
		createReq := jsonAPIBody("XCreateRequest", map[string]*Schema{
			"count": {Kind: SchemaKindPrimitive, Type: "integer"},
		}, nil)
		readResp := jsonAPIBody("XResponse", map[string]*Schema{
			"count": {Kind: SchemaKindPrimitive, Type: "string"},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateX", RequestSchema: createReq},
			Read:   &Operation{OperationId: "GetX", ResponseSchema: readResp},
		}

		_, _, err := MergeResourceSchema(group)
		Expect(err).To(HaveOccurred())
		var mergeErr *SchemaMergeError
		Expect(errors.As(err, &mergeErr)).To(BeTrue())
		Expect(mergeErr.Path).To(Equal("data.attributes.count"))
		Expect(mergeErr.Aspect).To(Equal("type"))
		Expect([]string{mergeErr.Left, mergeErr.Right}).To(ConsistOf("integer", "string"))
	})

	It("fails with SchemaMergeError on a Kind conflict, at the deeper path when it is inside an array element", func() {
		createReq := jsonAPIBody("XCreateRequest", map[string]*Schema{
			"tags": {
				Kind:  SchemaKindArray,
				Items: &Schema{Kind: SchemaKindPrimitive, Type: "string"},
			},
		}, nil)
		readResp := jsonAPIBody("XResponse", map[string]*Schema{
			"tags": {
				Kind: SchemaKindArray,
				Items: &Schema{
					Kind:       SchemaKindObject,
					Properties: map[string]*Schema{"value": {Kind: SchemaKindPrimitive, Type: "string"}},
				},
			},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateX", RequestSchema: createReq},
			Read:   &Operation{OperationId: "GetX", ResponseSchema: readResp},
		}

		_, _, err := MergeResourceSchema(group)
		Expect(err).To(HaveOccurred())
		var mergeErr *SchemaMergeError
		Expect(errors.As(err, &mergeErr)).To(BeTrue())
		Expect(mergeErr.Path).To(Equal("data.attributes.tags[]"))
		Expect(mergeErr.Aspect).To(Equal("kind"))
		Expect([]string{mergeErr.Left, mergeErr.Right}).To(ConsistOf(string(SchemaKindPrimitive), string(SchemaKindObject)))
	})

	It("requires a resolved Create and Read operation", func() {
		_, _, err := MergeResourceSchema(&ResolvedGroup{})
		Expect(err).To(HaveOccurred())

		_, _, err = MergeResourceSchema(&ResolvedGroup{Create: &Operation{}})
		Expect(err).To(HaveOccurred())
	})
})
