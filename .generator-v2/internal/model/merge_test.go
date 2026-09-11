package model

import (
	"errors"
	"reflect"
	"strings"

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

// setTestBoolField/testBoolField keep T153 executable before T154 adds the
// write-only metadata fields. The intended red result is a Ginkgo failure that
// names the missing contract, not a Go compilation error that obscures it.
func setTestBoolField(target any, name string, value bool) {
	GinkgoHelper()
	field := reflect.ValueOf(target).Elem().FieldByName(name)
	Expect(field.IsValid()).To(BeTrue(), "%T is missing %s metadata", target, name)
	Expect(field.CanSet()).To(BeTrue(), "%T.%s cannot be set", target, name)
	Expect(field.Kind()).To(Equal(reflect.Bool), "%T.%s must be boolean", target, name)
	field.SetBool(value)
}

func testBoolField(target any, name string) bool {
	GinkgoHelper()
	field := reflect.ValueOf(target).Elem().FieldByName(name)
	Expect(field.IsValid()).To(BeTrue(), "%T is missing %s metadata", target, name)
	Expect(field.Kind()).To(Equal(reflect.Bool), "%T.%s must be boolean", target, name)
	return field.Bool()
}

func secretSchema(writeOnly, sensitive bool) *Schema {
	schema := &Schema{Kind: SchemaKindPrimitive, Type: "string", Sensitive: sensitive}
	if writeOnly {
		setTestBoolField(schema, "WriteOnlySecret", true)
	}
	return schema
}

var _ = Describe("MergeResourceSchema", func() {
	DescribeTable("selects generated write-only handling only from a request-role writeOnly marker",
		func(createWriteOnly, updateWriteOnly, readWriteOnly, requestSensitive, wantWriteOnly, wantSensitive bool) {
			createReq := jsonAPIBody("AccountCreateRequest", map[string]*Schema{
				"password": secretSchema(createWriteOnly, requestSensitive),
			}, nil)
			updateReq := jsonAPIBody("AccountUpdateRequest", map[string]*Schema{
				"password": secretSchema(updateWriteOnly, requestSensitive),
			}, nil)
			readResp := jsonAPIBody("AccountResponse", map[string]*Schema{
				"password": secretSchema(readWriteOnly, readWriteOnly),
			}, nil)

			merged, _, err := MergeResourceSchema(&ResolvedGroup{
				Create: &Operation{OperationId: "CreateAccount", RequestSchema: createReq},
				Update: &Operation{OperationId: "UpdateAccount", RequestSchema: updateReq},
				Read:   &Operation{OperationId: "GetAccount", ResponseSchema: readResp},
			})
			Expect(err).NotTo(HaveOccurred())

			password := attributesOf(merged)["password"]
			Expect(testBoolField(password, "WriteOnlySecret")).To(Equal(wantWriteOnly))
			Expect(password.Sensitive).To(Equal(wantSensitive))
		},
		Entry("Create writeOnly:true selects it", true, false, false, false, true, false),
		Entry("Update writeOnly:true selects it", false, true, false, false, true, false),
		Entry("request x-secret/tracking-sensitive does not select it", false, false, false, true, false, true),
		Entry("response writeOnly:true does not select it", false, false, true, false, false, true),
		Entry("no marker selects neither behavior", false, false, false, false, false, false),
	)

	DescribeTable("records write-only requiredness independently for Create and Update",
		func(createRequired, updateRequired bool) {
			createReq := jsonAPIBody("AccountCreateRequest", map[string]*Schema{
				"password": secretSchema(true, false),
			}, nil)
			updateReq := jsonAPIBody("AccountUpdateRequest", map[string]*Schema{
				"password": secretSchema(true, false),
			}, nil)
			if createRequired {
				createReq.Properties["data"].Properties["attributes"].Required = []string{"password"}
			}
			if updateRequired {
				updateReq.Properties["data"].Properties["attributes"].Required = []string{"password"}
			}

			merged, _, err := MergeResourceSchema(&ResolvedGroup{
				Create: &Operation{OperationId: "CreateAccount", RequestSchema: createReq},
				Update: &Operation{OperationId: "UpdateAccount", RequestSchema: updateReq},
				Read: &Operation{OperationId: "GetAccount", ResponseSchema: jsonAPIBody(
					"AccountResponse", map[string]*Schema{}, nil)},
			})
			Expect(err).NotTo(HaveOccurred())

			password := attributesOf(merged)["password"]
			Expect(testBoolField(password, "SecretRequiredOnCreate")).To(Equal(createRequired))
			Expect(testBoolField(password, "SecretRequiredOnUpdate")).To(Equal(updateRequired))
		},
		Entry("optional in both roles", false, false),
		Entry("required only by Create", true, false),
		Entry("required only by Update", false, true),
		Entry("required by both roles", true, true),
	)

	DescribeTable("rejects write-only fields absent from one request lifecycle role",
		func(createHasSecret, updateHasSecret bool, missingRole string) {
			createAttributes := map[string]*Schema{}
			updateAttributes := map[string]*Schema{}
			if createHasSecret {
				createAttributes["password"] = secretSchema(true, false)
			}
			if updateHasSecret {
				updateAttributes["password"] = secretSchema(true, false)
			}

			_, _, err := MergeResourceSchema(&ResolvedGroup{
				Create: &Operation{OperationId: "CreateAccount", RequestSchema: jsonAPIBody(
					"AccountCreateRequest", createAttributes, nil)},
				Update: &Operation{OperationId: "UpdateAccount", RequestSchema: jsonAPIBody(
					"AccountUpdateRequest", updateAttributes, nil)},
				Read: &Operation{OperationId: "GetAccount", ResponseSchema: jsonAPIBody(
					"AccountResponse", map[string]*Schema{}, nil)},
			})

			Expect(err).To(HaveOccurred())
			var lifecycleErr *WriteOnlyLifecycleError
			Expect(errors.As(err, &lifecycleErr)).To(BeTrue())
			Expect(err.Error()).To(ContainSubstring("data.attributes.password"))
			Expect(err.Error()).To(ContainSubstring(missingRole))
		},
		Entry("Create-only secret", true, false, "Update"),
		Entry("Update-only secret", false, true, "Create"),
	)

	It("rejects a write-only field when the resource has no Update role", func() {
		_, _, err := MergeResourceSchema(&ResolvedGroup{
			Create: &Operation{OperationId: "CreateAccount", RequestSchema: jsonAPIBody(
				"AccountCreateRequest", map[string]*Schema{"password": secretSchema(true, false)}, nil)},
			Read: &Operation{OperationId: "GetAccount", ResponseSchema: jsonAPIBody(
				"AccountResponse", map[string]*Schema{}, nil)},
		})

		var lifecycleErr *WriteOnlyLifecycleError
		Expect(errors.As(err, &lifecycleErr)).To(BeTrue())
		Expect(lifecycleErr.MissingRole).To(Equal("Update"))
	})

	It("suppresses a request write-only field returned by Read and emits one deterministic value-free warning", func() {
		requestPassword := secretSchema(true, false)
		requestPassword.Description = "configured-secret-must-not-appear"
		responsePassword := secretSchema(false, true)
		responsePassword.Description = "observed-secret-must-not-appear"
		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateAccount", RequestSchema: jsonAPIBody(
				"AccountCreateRequest", map[string]*Schema{"password": requestPassword}, []string{"password"})},
			Update: &Operation{OperationId: "UpdateAccount", RequestSchema: jsonAPIBody(
				"AccountUpdateRequest", map[string]*Schema{"password": secretSchema(true, false)}, nil)},
			Read: &Operation{
				OperationId: "GetAccount",
				Tracking: &TrackingFieldMetadata{
					ArtifactKind: ArtifactKindResource,
					ArtifactName: "integration_account",
				},
				ResponseSchema: jsonAPIBody(
					"AccountResponse", map[string]*Schema{"password": responsePassword}, nil),
			},
		}

		merged, firstDiags, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		password := attributesOf(merged)["password"]
		Expect(testBoolField(password, "WriteOnlySecret")).To(BeTrue())
		Expect(password.Provenance.InResponse).To(BeFalse(),
			"a write-only response field must not produce an updateState assignment")

		_, secondDiags, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		warnings := func(diags []Diagnostic) []Diagnostic {
			var out []Diagnostic
			for _, diagnostic := range diags {
				if diagnostic.Severity == SeverityWarning {
					out = append(out, diagnostic)
				}
			}
			return out
		}
		firstWarnings := warnings(firstDiags)
		Expect(firstWarnings).To(HaveLen(1))
		Expect(warnings(secondDiags)).To(Equal(firstWarnings))
		message := firstWarnings[0].Message
		Expect(message).To(ContainSubstring("integration_account"))
		Expect(message).To(ContainSubstring("data.attributes.password"))
		Expect(message).NotTo(ContainSubstring("configured-secret-must-not-appear"))
		Expect(message).NotTo(ContainSubstring("observed-secret-must-not-appear"))
		Expect(strings.ToLower(message)).NotTo(ContainSubstring("value="))
	})

	It("correlates by property position across three differently-named components and stamps the provenance bits", func() {
		createReq := jsonAPIBody("IncidentTypeCreateRequest", map[string]*Schema{
			"name":          {Kind: SchemaKindPrimitive, Type: "string", Description: "name (create)"},
			"is_default":    {Kind: SchemaKindPrimitive, Type: "boolean"},
			"internal_note": {Kind: SchemaKindPrimitive, Type: "string"},
			"secret_token":  {Kind: SchemaKindPrimitive, Type: "string"},
			"priority":      {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"low", "high"}},
			"status":        {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"open", "closed"}},
		}, []string{"name", "secret_token"})

		// A PATCH body marking "priority" required is deliberately unusual: it
		// proves RequestRequired reads the Create body's Required list only.
		updateReq := jsonAPIBody("IncidentTypeUpdateRequest", map[string]*Schema{
			"name":            {Kind: SchemaKindPrimitive, Type: "string"},
			"is_default":      {Kind: SchemaKindPrimitive, Type: "boolean"},
			"priority":        {Kind: SchemaKindPrimitive, Type: "string", Enum: []string{"low", "high"}},
			"extra_on_update": {Kind: SchemaKindPrimitive, Type: "string"},
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

		By("required in Create, present in Update (but not required there), present in response -> Required; requiredness comes from Create alone")
		Expect(attrs["name"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: true}))
		// Cosmetic: descriptions differ, Read response wins.
		Expect(attrs["name"].Description).To(Equal("The incident type name."))

		By("optional in Create, present in Update, present in response -> Optional+Computed")
		Expect(attrs["is_default"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: false, InResponse: true}))

		By("Create-only, absent from Update and the response -> write-only Optional")
		Expect(attrs["internal_note"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: false, InResponse: false}))

		By("required in Create, absent from Update and the response -> write-only Required")
		Expect(attrs["secret_token"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: true, InResponse: false}))

		By("absent from Create, present in Update only -> InRequest true via the Create∪Update union, not required")
		Expect(attrs["extra_on_update"].Provenance).To(Equal(&SchemaProvenance{InRequest: true, RequestRequired: false, InResponse: false}))

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

		By("RequestRefName never prefers Read: the Create body's own name wins instead")
		Expect(merged.RequestRefName).To(Equal("IncidentTypeCreateRequest"))

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

	It("falls RequestRefName back to Update's component name when Create doesn't reach the node, and leaves it empty when neither does", func() {
		createReq := jsonAPIBody("XCreateRequest", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
		}, []string{"name"})
		updateReq := jsonAPIBody("XUpdateRequest", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
			"settings": {
				Kind:    SchemaKindObject,
				RefName: "XSettingsUpdateRequest",
				Properties: map[string]*Schema{
					"enabled": {Kind: SchemaKindPrimitive, Type: "boolean"},
				},
			},
		}, nil)
		readResp := jsonAPIBody("XResponse", map[string]*Schema{
			"name": {Kind: SchemaKindPrimitive, Type: "string"},
			"settings": {
				Kind:    SchemaKindObject,
				RefName: "XSettingsResponse",
				Properties: map[string]*Schema{
					"enabled": {Kind: SchemaKindPrimitive, Type: "boolean"},
				},
			},
			// Present only in the response: RequestRefName must stay empty
			// rather than borrow Read's name for a field the request never sets.
			"audit": {Kind: SchemaKindObject, RefName: "XAuditResponse", Properties: map[string]*Schema{}},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateX", RequestSchema: createReq},
			Update: &Operation{OperationId: "UpdateX", RequestSchema: updateReq},
			Read:   &Operation{OperationId: "GetX", ResponseSchema: readResp},
		}

		merged, _, err := MergeResourceSchema(group)
		Expect(err).NotTo(HaveOccurred())
		attrs := attributesOf(merged)

		Expect(attrs["settings"].RequestRefName).To(Equal("XSettingsUpdateRequest"))
		Expect(attrs["audit"].RequestRefName).To(BeEmpty())
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

		By("a sibling artifact in the same run still builds")
		sibling, _, siblingErr := MergeResourceSchema(siblingGroup())
		Expect(siblingErr).NotTo(HaveOccurred())
		Expect(sibling).NotTo(BeNil())
	})

	It("fails with SchemaMergeError naming the path and both spellings on a primitive format conflict", func() {
		createReq := jsonAPIBody("XCreateRequest", map[string]*Schema{
			"seen_at": {Kind: SchemaKindPrimitive, Type: "string", Format: "date-time"},
		}, nil)
		readResp := jsonAPIBody("XResponse", map[string]*Schema{
			"seen_at": {Kind: SchemaKindPrimitive, Type: "string", Format: "date"},
		}, nil)

		group := &ResolvedGroup{
			Create: &Operation{OperationId: "CreateX", RequestSchema: createReq},
			Read:   &Operation{OperationId: "GetX", ResponseSchema: readResp},
		}

		_, _, err := MergeResourceSchema(group)
		Expect(err).To(HaveOccurred())
		var mergeErr *SchemaMergeError
		Expect(errors.As(err, &mergeErr)).To(BeTrue())
		Expect(mergeErr.Path).To(Equal("data.attributes.seen_at"))
		Expect(mergeErr.Aspect).To(Equal("format"))
		Expect([]string{mergeErr.Left, mergeErr.Right}).To(ConsistOf("date-time", "date"))

		By("a sibling artifact in the same run still builds")
		sibling, _, siblingErr := MergeResourceSchema(siblingGroup())
		Expect(siblingErr).NotTo(HaveOccurred())
		Expect(sibling).NotTo(BeNil())
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

		By("a sibling artifact in the same run still builds")
		sibling, _, siblingErr := MergeResourceSchema(siblingGroup())
		Expect(siblingErr).NotTo(HaveOccurred())
		Expect(sibling).NotTo(BeNil())
	})
})

// siblingGroup is a minimal, always-mergeable group used to prove that a
// structural conflict in one resource's merge leaves no state behind that
// would affect another resource merged afterward in the same run.
func siblingGroup() *ResolvedGroup {
	createReq := jsonAPIBody("SiblingCreateRequest", map[string]*Schema{
		"name": {Kind: SchemaKindPrimitive, Type: "string"},
	}, []string{"name"})
	readResp := jsonAPIBody("SiblingResponse", map[string]*Schema{
		"name": {Kind: SchemaKindPrimitive, Type: "string"},
	}, nil)
	return &ResolvedGroup{
		Create: &Operation{OperationId: "CreateSibling", RequestSchema: createReq},
		Read:   &Operation{OperationId: "GetSibling", ResponseSchema: readResp},
	}
}
