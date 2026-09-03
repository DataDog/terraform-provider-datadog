package emit

import (
	"errors"
	"go/format"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("BuildResourceView", func() {
	It("fails rather than rendering an empty schema when the merge has not run", func() {
		op := incidentTypeResourceOperation(true)
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		art.Schema = nil

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("has no schema"))
	})

	It("resolves the CRUD lifecycle, schema and request fields for a resource with an Update role", func() {
		art, err := model.BuildArtifact(incidentTypeResourceOperation(true))
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		Expect(view.TypeName).To(Equal("incident_type"))
		Expect(view.GoName).To(Equal("datadogIncidentType"))
		Expect(view.SDKPackage).To(Equal("datadogV2"))
		Expect(view.APIStruct).To(Equal("IncidentsApi"))
		Expect(view.UpdateUnsupported).To(BeFalse())

		Expect(view.Create).To(Equal(CRUDCallView{
			Method: "CreateIncidentType", GoRequestType: "IncidentTypeCreateRequest", GoResponseType: "IncidentTypeResponse",
			Envelope: &RequestEnvelopeView{
				SDKPackage: "datadogV2", Fields: view.Create.Envelope.Fields,
				DataVar: "bodyData", DataType: "IncidentTypeCreateData",
				TypeExpr:      `datadogV2.IncidentTypeType("incident_types")`,
				AttributesVar: "bodyAttributes", AttributesType: "IncidentTypeAttributes",
			},
		}))
		Expect(view.Read).To(Equal(CRUDCallView{
			Method: "GetIncidentType", GoResponseType: "IncidentTypeResponse",
			Arguments: []SDKArgumentView{{Expression: "state.ID.ValueString()", TFName: "id"}},
		}))
		By("Update builds its own envelope: a PATCH's data and attributes components are distinct types from Create's")
		Expect(view.Update).To(Equal(CRUDCallView{
			Method: "UpdateIncidentType", GoRequestType: "IncidentTypeUpdateRequest", GoResponseType: "IncidentTypeResponse",
			Arguments:    []SDKArgumentView{{Expression: "state.ID.ValueString()", TFName: "id"}},
			BodyIDExpr:   "state.ID.ValueString()",
			BodyIDTarget: "bodyData",
			Envelope: &RequestEnvelopeView{
				SDKPackage: "datadogV2", Fields: view.Create.Envelope.Fields,
				DataVar: "bodyData", DataType: "IncidentTypeUpdateData",
				TypeExpr:      `datadogV2.IncidentTypeType("incident_types")`,
				AttributesVar: "bodyAttributes", AttributesType: "IncidentTypeUpdateAttributes",
			},
		}))
		Expect(view.Delete).To(Equal(CRUDCallView{
			Method:    "DeleteIncidentType",
			Arguments: []SDKArgumentView{{Expression: "state.ID.ValueString()", TFName: "id"}},
		}))

		By("required in Create, present in Update and the response -> Required, request-settable")
		Expect(attrByPath(schemaTree(view), "name")).To(Equal(AttrView{
			TFName: "name", TFType: "schema.StringAttribute", Description: "Name of the incident type.", Required: true,
		}))

		By("optional, present in the response -> Optional+Computed with UseStateForUnknown")
		Expect(attrByPath(schemaTree(view), "description")).To(Equal(AttrView{
			TFName: "description", TFType: "schema.StringAttribute", Description: "Description of the incident type.",
			Optional: true, Computed: true,
			PlanModifiers: []string{"stringplanmodifier.UseStateForUnknown"}, PlanModifierType: "String",
		}))

		By("Create-only, absent from the response -> write-only Optional, no plan modifiers with Update present")
		Expect(attrByPath(schemaTree(view), "internal_note")).To(Equal(AttrView{
			TFName: "internal_note", TFType: "schema.StringAttribute", Description: "Internal note.", Optional: true,
		}))

		By("response-only -> Computed only, not request-settable")
		Expect(attrByPath(schemaTree(view), "last_seen")).To(Equal(AttrView{
			TFName: "last_seen", TFType: "schema.StringAttribute", Description: "Last time this incident type was seen.", Computed: true,
		}))

		const attrsTarget = "bodyAttributes"
		Expect(view.Create.Envelope.Fields).To(ConsistOf(
			RequestFieldView{GoField: "Name", Target: attrsTarget, ValueExpr: "state.Name.ValueString()", Required: true},
			RequestFieldView{
				GoField: "Description", Target: attrsTarget, ValueExpr: "state.Description.ValueString()",
				NullCheck: "!state.Description.IsNull() && !state.Description.IsUnknown()",
			},
			RequestFieldView{
				GoField: "InternalNote", Target: attrsTarget, ValueExpr: "state.InternalNote.ValueString()",
				NullCheck: "!state.InternalNote.IsNull() && !state.InternalNote.IsUnknown()",
			},
		))

		Expect(view.State.ParamType).To(Equal("*datadogV2.IncidentTypeResponse"))
		By("the write-only field never gets a response assignment — its accessor does not exist on the response type")
		for _, a := range view.State.Assignments {
			Expect(a.LHS).NotTo(Equal("state.InternalNote"))
		}
		Expect(view.State.Assignments).To(ContainElement(HaveField("LHS", "state.LastSeen")))
	})

	It("sends the JSON:API type discriminator explicitly when the spec determines it", func() {
		op := incidentTypeResourceOperation(true)
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		By("both roles convert the value through the type property's own component")
		Expect(view.Create.Envelope.TypeExpr).To(Equal(`datadogV2.IncidentTypeType("incident_types")`))
		Expect(view.Update.Envelope.TypeExpr).To(Equal(`datadogV2.IncidentTypeType("incident_types")`))

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(src)).To(ContainSubstring(`bodyData.SetType(datadogV2.IncidentTypeType("incident_types"))`))
	})

	It("leaves the discriminator to the SDK when the value is ambiguous but defaulted", func() {
		By("a two-member enum does not say which value a request carries, so only a spec default can settle it")
		op := incidentTypeResourceOperation(true)
		for _, role := range []*model.Operation{op.ResolvedGroup.Create, op.ResolvedGroup.Update} {
			typeProperty := role.RequestSchema.Properties["data"].Properties["type"]
			typeProperty.Enum = []string{"incident_types", "incident_type"}
			typeProperty.HasDefault = true
		}

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Create.Envelope.TypeExpr).To(BeEmpty())
		Expect(string(mustRenderResource(view))).NotTo(ContainSubstring("bodyData.SetType("))
	})

	DescribeTable("fails the artifact when neither the spec nor the SDK determines the discriminator",
		func(mutate func(*model.Schema), wantMessage string) {
			op := incidentTypeResourceOperation(true)
			mutate(op.ResolvedGroup.Create.RequestSchema.Properties["data"].Properties["type"])

			art, err := model.BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())

			_, err = BuildResourceView(art)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(wantMessage))
		},
		Entry("several allowed values and no default",
			func(s *model.Schema) { s.Enum = []string{"incident_types", "incident_type"} },
			`allows "incident_types" or "incident_type" and declares no default, so the spec does not say which value a request carries`),
		Entry("no allowed values and no default",
			func(s *model.Schema) { s.Enum = nil },
			"constrains no values and declares no default"),
		Entry("an inline type property has no component name to convert through",
			func(s *model.Schema) { s.RefName = "" },
			`the property is inline, so it has no SDK component name to convert the value through`),
		Entry("a default the SDK skips because the property is readOnly",
			func(s *model.Schema) { s.Enum = nil; s.HasDefault = true; s.ReadOnly = true },
			"constrains no values and declares no default"),
	)

	It("fails the artifact when a request envelope level names no SDK component", func() {
		By("an inline data member leaves nothing to construct, and so nothing to set the JSON:API type discriminator on")
		op := incidentTypeResourceOperation(true)
		op.ResolvedGroup.Create.RequestSchema.Properties["data"].RefName = ""

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`Create request body for resource "incident_type" leaves its JSON:API "data" member inline`))

		By("and an inline attributes member, when the body has attributes to set")
		op = incidentTypeResourceOperation(true)
		op.ResolvedGroup.Update.RequestSchema.Properties["data"].Properties["attributes"].RefName = ""

		art, err = model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Update request body for resource \"incident_type\" has settable attributes but leaves data.attributes inline"))
	})

	It("degrades to the forced-replacement stub when the group resolves no Update role", func() {
		art, err := model.BuildArtifact(incidentTypeResourceOperation(false))
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		Expect(view.UpdateUnsupported).To(BeTrue())
		Expect(view.Update).To(Equal(CRUDCallView{}))

		By("every request-settable attribute now also carries RequiresReplace()")
		Expect(attrByPath(schemaTree(view), "internal_note").PlanModifiers).To(Equal(
			[]string{"stringplanmodifier.RequiresReplace"}))
		Expect(attrByPath(schemaTree(view), "description").PlanModifiers).To(Equal(
			[]string{"stringplanmodifier.UseStateForUnknown", "stringplanmodifier.RequiresReplace"}))
	})

	It("fails the artifact on a request-settable nested object rather than silently dropping it", func() {
		op := incidentTypeResourceOperation(true)
		attributes := op.RequestSchema.Properties["data"].Properties["attributes"]
		attributes.Properties["settings"] = &model.Schema{
			Kind: model.SchemaKindObject,
			Properties: map[string]*model.Schema{
				"enabled": {Kind: model.SchemaKindPrimitive, Type: "boolean"},
			},
		}

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		var unsupported *UnsupportedEmitError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
		Expect(unsupported.Nodes).To(ContainElement(HaveField("Path", ContainSubstring("settings"))))
	})

	It("fails when Create and Read resolve different response types", func() {
		op := incidentTypeResourceOperation(true)
		op.ResolvedGroup.Create.ResponseRefName = "SomethingElseResponse"

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("differs from Read's"))
	})
})

var _ = Describe("RenderResource", func() {
	It("renders gofmt-canonical, syntactically valid Go for a full-CRUD resource", func() {
		art, err := model.BuildArtifact(incidentTypeResourceOperation(true))
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())

		_, err = format.Source(src)
		Expect(err).NotTo(HaveOccurred(), "rendered output must already be gofmt-canonical Go:\n%s", src)

		out := string(src)
		Expect(out).To(ContainSubstring(`func (r *datadogIncidentTypeResource) Create(`))
		By("Create builds the envelope innermost-first, so the data component's own constructor sets the JSON:API type discriminator (T138)")
		for _, want := range []string{
			"bodyAttributes := datadogV2.NewIncidentTypeAttributesWithDefaults()",
			"bodyData := datadogV2.NewIncidentTypeCreateDataWithDefaults()",
			"bodyData.SetAttributes(*bodyAttributes)",
			"body := datadogV2.NewIncidentTypeCreateRequestWithDefaults()",
			"body.SetData(*bodyData)",
		} {
			Expect(out).To(ContainSubstring(want))
		}
		By("and never reaches through the wrapper, which would leave that discriminator empty")
		Expect(out).NotTo(ContainSubstring("body.Data.Attributes."))
		Expect(out).To(ContainSubstring(`bodyAttributes.SetName(state.Name.ValueString())`))
		Expect(out).To(ContainSubstring(`if !state.Description.IsNull() && !state.Description.IsUnknown() {`))
		Expect(out).To(ContainSubstring(`resp, _, err := r.Api.CreateIncidentType(r.Auth, *body)`))

		Expect(out).To(ContainSubstring(`func (r *datadogIncidentTypeResource) Read(`))
		Expect(out).To(ContainSubstring(`r.Api.GetIncidentType(r.Auth, state.ID.ValueString())`))
		Expect(out).To(ContainSubstring(`response.State.RemoveResource(ctx)`))

		Expect(out).To(ContainSubstring(`func (r *datadogIncidentTypeResource) Update(`))
		for _, want := range []string{
			"bodyAttributes := datadogV2.NewIncidentTypeUpdateAttributesWithDefaults()",
			"bodyData := datadogV2.NewIncidentTypeUpdateDataWithDefaults()",
			"body := datadogV2.NewIncidentTypeUpdateRequestWithDefaults()",
		} {
			Expect(out).To(ContainSubstring(want))
		}
		Expect(out).To(ContainSubstring(`bodyData.SetId(state.ID.ValueString())`))
		Expect(out).To(ContainSubstring(`r.Api.UpdateIncidentType(r.Auth, state.ID.ValueString(), *body)`))

		Expect(out).To(ContainSubstring(`func (r *datadogIncidentTypeResource) Delete(`))
		Expect(out).To(ContainSubstring(`r.Api.DeleteIncidentType(r.Auth, state.ID.ValueString())`))

		Expect(out).To(ContainSubstring(`resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)`))
		Expect(out).To(ContainSubstring(`"id": utils.ResourceIDAttribute(),`))
	})

	It("stubs Update with an error rather than a request builder when the group resolves no Update role", func() {
		art, err := model.BuildArtifact(incidentTypeResourceOperation(false))
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())

		out := string(src)
		Expect(out).To(ContainSubstring(`response.Diagnostics.AddError("Update should not be called", "Updating this resource should replace it.")`))
		Expect(out).NotTo(ContainSubstring(`NewIncidentTypeUpdateRequestWithDefaults`))
	})
})

// incidentTypeResourceOperation is the incident_type Create operation of a
// full-CRUD resource group, JSON:API-shaped like the real spec: "name" is
// required by Create only, "description" is optional and present in the
// response (server-defaultable), "internal_note" is Create-only (write-only),
// and "last_seen" is response-only. withUpdate selects whether the group
// resolves an Update role.
func incidentTypeResourceOperation(withUpdate bool) *model.Operation {
	attrs := func(required []string, description, internalNote, lastSeen bool, attrsRefName string) *model.Schema {
		props := map[string]*model.Schema{
			"name": prim("string", "Name of the incident type."),
		}
		if description {
			props["description"] = prim("string", "Description of the incident type.")
		}
		if internalNote {
			props["internal_note"] = prim("string", "Internal note.")
		}
		if lastSeen {
			props["last_seen"] = prim("string", "Last time this incident type was seen.")
		}
		s := obj(props)
		s.Required = required
		s.RefName = attrsRefName
		return s
	}
	// The envelope levels carry their own component names, as a real spec's do:
	// the SDK generates a model per $ref, and T138's request mapper constructs
	// each level from that model's own New<Type>WithDefaults().
	body := func(a *model.Schema, dataRefName string) *model.Schema {
		// A real JSON:API envelope carries the "type" discriminator beside
		// "attributes", constrained to a single value by its own component —
		// which is what lets the request mapper send it (T143).
		discriminator := prim("string", "Incident type resource type.")
		discriminator.RefName = "IncidentTypeType"
		discriminator.Enum = []string{"incident_types"}
		data := obj(map[string]*model.Schema{"attributes": a, "type": discriminator})
		data.RefName = dataRefName
		return obj(map[string]*model.Schema{"data": data})
	}

	// idBinding resolves the terminal "{incident_type_id}" path segment as the
	// aliased "id" argument, the same shape a real parsed SDKOperationBinding
	// would carry for a by-id path.
	idBinding := func() *model.SDKOperationBinding {
		return &model.SDKOperationBinding{Required: []model.SDKArgument{
			{Name: "incident_type_id", GoName: "incidentTypeId", GoType: "string", Location: "path", Schema: prim("string", "The incident type ID.")},
		}}
	}

	create := &model.Operation{
		Path: "/api/v2/incidents/config/types", Method: "POST",
		OperationId: "CreateIncidentType", Tag: "Incidents",
		RequestRefName: "IncidentTypeCreateRequest", ResponseRefName: "IncidentTypeResponse",
		RequestSchema: body(attrs([]string{"name"}, true, true, false, "IncidentTypeAttributes"), "IncidentTypeCreateData"),
	}
	read := &model.Operation{
		Path: "/api/v2/incidents/config/types/{incident_type_id}", Method: "GET",
		OperationId: "GetIncidentType", Tag: "Incidents",
		ResponseRefName: "IncidentTypeResponse",
		ResponseSchema:  body(attrs(nil, true, false, true, "IncidentTypeAttributes"), "IncidentTypeData"),
		SDKBinding:      idBinding(),
	}
	del := &model.Operation{
		Path: "/api/v2/incidents/config/types/{incident_type_id}", Method: "DELETE",
		OperationId: "DeleteIncidentType", Tag: "Incidents",
		SDKBinding: idBinding(),
	}
	create.Tracking = &model.TrackingFieldMetadata{
		ArtifactKind:  model.ArtifactKindResource,
		ArtifactName:  "incident_type",
		TfDescription: "Provides a Datadog incident type resource.",
		IdStrategy:    model.IdStrategyDataID,
	}
	create.ResolvedGroup = &model.ResolvedGroup{Create: create, Read: read, Delete: del}
	if withUpdate {
		update := &model.Operation{
			Path: "/api/v2/incidents/config/types/{incident_type_id}", Method: "PATCH",
			OperationId: "UpdateIncidentType", Tag: "Incidents",
			RequestRefName: "IncidentTypeUpdateRequest", ResponseRefName: "IncidentTypeResponse",
			RequestSchema: body(attrs(nil, true, false, false, "IncidentTypeUpdateAttributes"), "IncidentTypeUpdateData"),
			SDKBinding:    idBinding(),
		}
		create.ResolvedGroup.Update = update
	}
	return create
}

// schemaTree flattens a ResourceView's rendered schema attributes for
// attrByPath, which descends AttrView.Attributes/Blocks by TFName.
func schemaTree(v ResourceView) *AttributeTreeLike {
	return &AttributeTreeLike{Attributes: v.Schema.Attributes}
}

// AttributeTreeLike adapts a []AttrView so attrByPath (an *AttributeTree
// helper defined in the model package's schema_test.go) is not needed here —
// this package renders AttrView, not model.Attribute, so it defines its own
// tiny path-lookup instead of depending on a test helper from another package.
type AttributeTreeLike struct {
	Attributes []AttrView
}

func attrByPath(tree *AttributeTreeLike, tfName string) AttrView {
	GinkgoHelper()
	var find func(attrs []AttrView) (AttrView, bool)
	find = func(attrs []AttrView) (AttrView, bool) {
		for _, a := range attrs {
			if a.TFName == tfName {
				return a, true
			}
			if got, ok := find(a.Attributes); ok {
				return got, true
			}
			if got, ok := find(a.Blocks); ok {
				return got, true
			}
		}
		return AttrView{}, false
	}
	got, ok := find(tree.Attributes)
	Expect(ok).To(BeTrue(), "no attribute named %q", tfName)
	return got
}

// mustRenderResource renders a view, failing the spec rather than returning an
// error, for assertions whose subject is the rendered text.
func mustRenderResource(view ResourceView) []byte {
	src, err := RenderResource(view)
	Expect(err).NotTo(HaveOccurred())
	return src
}
