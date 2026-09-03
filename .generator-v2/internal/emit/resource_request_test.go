package emit

import (
	"errors"
	"go/format"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var _ = Describe("BuildResourceView request mapping", func() {
	It("builds nested objects (recursively), an enum, and format leaves, and renders valid Go", func() {
		art, err := model.BuildArtifact(widgetResourceOperation())
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		By("a top-level required enum casts to its named SDK type")
		priority := requestFieldByGoField(view.RequestFields, "Priority")
		Expect(priority.Required).To(BeTrue())
		Expect(priority.Nested).To(BeNil())
		Expect(priority.ValueExpr).To(Equal("datadogV2.WidgetPriority(state.Priority.ValueString())"))

		By("a date-time leaf parses before the setter, guarded on the field's own null check")
		expiresAt := requestFieldByGoField(view.RequestFields, "ExpiresAt")
		Expect(expiresAt.Required).To(BeFalse())
		Expect(expiresAt.NullCheck).To(Equal("!state.ExpiresAt.IsNull() && !state.ExpiresAt.IsUnknown()"))
		Expect(expiresAt.ParsedVar).To(Equal("expiresAtParsed"))
		Expect(expiresAt.ParseCall).To(Equal("time.Parse(time.RFC3339, state.ExpiresAt.ValueString())"))
		Expect(expiresAt.ValueExpr).To(Equal("expiresAtParsed"))

		By("a uuid leaf parses the same way")
		externalID := requestFieldByGoField(view.RequestFields, "ExternalId")
		Expect(externalID.ParsedVar).To(Equal("externalIdParsed"))
		Expect(externalID.ParseCall).To(Equal("uuid.Parse(state.ExternalId.ValueString())"))

		By("an int32-formatted integer casts down from the model's int64")
		retries := requestFieldByGoField(view.RequestFields, "Retries")
		Expect(retries.ValueExpr).To(Equal("int32(state.Retries.ValueInt64())"))

		By("a nested object builds its own SDK value via WithDefaults(), guarded on the model's own nil check")
		settings := requestFieldByGoField(view.RequestFields, "Settings")
		Expect(settings.Required).To(BeFalse())
		Expect(settings.NullCheck).To(Equal("state.Settings != nil"))
		Expect(settings.Nested).NotTo(BeNil())
		Expect(settings.Nested.Constructor).To(Equal("datadogV2.NewWidgetSettingsCreateRequestWithDefaults()"))
		Expect(settings.Nested.Var).To(Equal("settingsValue"))
		Expect(settings.Nested.ModelExpr).To(Equal("state.Settings"))

		By("the nested object's own required leaf reads through the model's nested pointer")
		url := requestFieldByGoField(settings.Nested.Fields, "Url")
		Expect(url.Required).To(BeTrue())
		Expect(url.ValueExpr).To(Equal("state.Settings.Url.ValueString()"))

		By("a doubly-nested object resolves relative to its immediate parent, not the top-level state")
		limits := requestFieldByGoField(settings.Nested.Fields, "Limits")
		Expect(limits.Nested).NotTo(BeNil())
		Expect(limits.Nested.Constructor).To(Equal("datadogV2.NewWidgetLimitsRequestWithDefaults()"))
		Expect(limits.Nested.ModelExpr).To(Equal("state.Settings.Limits"))
		maxField := requestFieldByGoField(limits.Nested.Fields, "Max")
		Expect(maxField.ValueExpr).To(Equal("int32(state.Settings.Limits.Max.ValueInt64())"))

		Expect(view.UsesUUID).To(BeTrue())
		Expect(view.UsesTime).To(BeTrue())

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())
		_, err = format.Source(src)
		Expect(err).NotTo(HaveOccurred(), "rendered output must already be gofmt-canonical Go:\n%s", src)

		out := string(src)
		for _, want := range []string{
			`"time"`,
			`"github.com/google/uuid"`,
			`body.Data.Attributes.SetPriority(datadogV2.WidgetPriority(state.Priority.ValueString()))`,
			`expiresAtParsed, err := time.Parse(time.RFC3339, state.ExpiresAt.ValueString())`,
			`externalIdParsed, err := uuid.Parse(state.ExternalId.ValueString())`,
			`body.Data.Attributes.SetRetries(int32(state.Retries.ValueInt64()))`,
			`if state.Settings != nil {`,
			`settingsValue := datadogV2.NewWidgetSettingsCreateRequestWithDefaults()`,
			`settingsValue.SetUrl(state.Settings.Url.ValueString())`,
			`limitsValue := datadogV2.NewWidgetLimitsRequestWithDefaults()`,
			`limitsValue.SetMax(int32(state.Settings.Limits.Max.ValueInt64()))`,
			`settingsValue.SetLimits(*limitsValue)`,
			`body.Data.Attributes.SetSettings(*settingsValue)`,
		} {
			Expect(out).To(ContainSubstring(want), "generated resource missing %q:\n%s", want, out)
		}
	})

	It("fails a request-settable oneOf with a diagnostic naming it, distinct from the generic unsupported-shape message", func() {
		op := widgetResourceOperation()
		attributes := op.RequestSchema.Properties["data"].Properties["attributes"]
		attributes.Properties["auth"] = oneOfSchemaForTest()

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		var unsupported *UnsupportedEmitError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
		found := false
		for _, n := range unsupported.Nodes {
			if strings.Contains(n.Path, "auth") {
				Expect(n.Reason).To(ContainSubstring("oneOf"))
				found = true
			}
		}
		Expect(found).To(BeTrue())
	})

	It("fails a request-settable array whose element has a format, naming the element path", func() {
		op := widgetResourceOperation()
		attributes := op.RequestSchema.Properties["data"].Properties["attributes"]
		attributes.Properties["tags"] = arrSchema(&model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"})

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		Expect(err).To(HaveOccurred())
		var unsupported *UnsupportedEmitError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
		found := false
		for _, n := range unsupported.Nodes {
			if strings.Contains(n.Path, "tags[]") {
				Expect(n.Reason).To(ContainSubstring("date-time"))
				found = true
			}
		}
		Expect(found).To(BeTrue())
	})

	It("builds a primitive list, a primitive map and a list of objects, and renders valid Go", func() {
		op := widgetResourceOperation()
		attributes := op.RequestSchema.Properties["data"].Properties["attributes"]
		attributes.Properties["tags"] = arrSchema(prim("string", ""))
		attributes.Properties["limits_by_region"] = &model.Schema{
			Kind:  model.SchemaKindMap,
			Items: prim("integer", "int64"),
		}
		attributes.Properties["recipients"] = &model.Schema{
			Kind: model.SchemaKindArray,
			Items: &model.Schema{
				Kind:     model.SchemaKindObject,
				RefName:  "RecipientRequest",
				Required: []string{"email"},
				Properties: map[string]*model.Schema{
					"email": prim("string", ""),
					"name":  prim("string", ""),
				},
			},
		}

		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		view, err := BuildResourceView(art)
		Expect(err).NotTo(HaveOccurred())

		By("a primitive list decodes via ElementsAs into a native Go slice")
		tags := requestFieldByGoField(view.RequestFields, "Tags")
		Expect(tags.Required).To(BeFalse())
		Expect(tags.Collection).NotTo(BeNil())
		Expect(tags.Collection.Kind).To(Equal("primitive"))
		Expect(tags.Collection.ConvertType).To(Equal("[]string"))
		Expect(tags.Collection.ConvertCall).To(Equal("state.Tags.ElementsAs(ctx, &tagsElements, false)"))

		By("a primitive map decodes into a native Go map")
		limits := requestFieldByGoField(view.RequestFields, "LimitsByRegion")
		Expect(limits.Collection.Kind).To(Equal("primitive"))
		Expect(limits.Collection.ConvertType).To(Equal("map[string]int64"))

		By("a list of objects builds one request element per already-decoded state element")
		recipients := requestFieldByGoField(view.RequestFields, "Recipients")
		Expect(recipients.Collection.Kind).To(Equal("object"))
		Expect(recipients.Collection.RangeExpr).To(Equal("state.Recipients"))
		Expect(recipients.Collection.Constructor).To(Equal("datadogV2.NewRecipientRequestWithDefaults()"))
		email := requestFieldByGoField(recipients.Collection.Fields, "Email")
		Expect(email.Required).To(BeTrue())
		Expect(email.ValueExpr).To(Equal("recipientsItem.Email.ValueString()"))

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())
		_, err = format.Source(src)
		Expect(err).NotTo(HaveOccurred(), "rendered output must already be gofmt-canonical Go:\n%s", src)

		out := string(src)
		for _, want := range []string{
			`var tagsElements []string`,
			`state.Tags.ElementsAs(ctx, &tagsElements, false)`,
			`body.Data.Attributes.SetTags(tagsElements)`,
			`var limitsByRegionElements map[string]int64`,
			`body.Data.Attributes.SetLimitsByRegion(limitsByRegionElements)`,
			`var recipientsElements []datadogV2.RecipientRequest`,
			`for _, recipientsItem := range state.Recipients {`,
			`recipientsElement := datadogV2.NewRecipientRequestWithDefaults()`,
			`recipientsElement.SetEmail(recipientsItem.Email.ValueString())`,
			`recipientsElements = append(recipientsElements, *recipientsElement)`,
			`body.Data.Attributes.SetRecipients(recipientsElements)`,
		} {
			Expect(out).To(ContainSubstring(want), "generated resource missing %q:\n%s", want, out)
		}
	})
})

// requestFieldByGoField finds the field named goField in fields, failing the
// spec if absent.
func requestFieldByGoField(fields []RequestFieldView, goField string) RequestFieldView {
	GinkgoHelper()
	for _, f := range fields {
		if f.GoField == goField {
			return f
		}
	}
	Fail("no request field named " + goField)
	return RequestFieldView{}
}

func arrSchema(item *model.Schema) *model.Schema {
	return &model.Schema{Kind: model.SchemaKindArray, Items: item}
}

func oneOfSchemaForTest() *model.Schema {
	return &model.Schema{
		Kind: model.SchemaKindOneOf,
		OneOf: &model.OneOfSpec{
			Name: "Auth",
			Path: "resource.auth",
			Variants: []model.OneOfVariant{
				{TFName: "basic", GoName: "Basic", Schema: prim("string", ""), ValueWrapped: true},
			},
		},
	}
}

// widgetResourceOperation is a synthetic full-CRUD resource exercising every
// request-mapping shape BuildResourceView supports: a required top-level
// enum, optional date-time/uuid/int32-formatted leaves, and a
// server-defaulted ("settings") nested object one level deep whose own
// "limits" field nests a second level, mirroring the real
// elastic_cloud_account fixture's "dataflows" shape. Create and Update name
// their own nested components distinctly from the Read response's, proving
// the request side never borrows the response's name (FR-034c only governs
// RefName, not RequestRefName).
func widgetResourceOperation() *model.Operation {
	limits := func(refName string) *model.Schema {
		return &model.Schema{
			Kind:    model.SchemaKindObject,
			RefName: refName,
			Properties: map[string]*model.Schema{
				"max": {Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int32"},
			},
		}
	}
	settings := func(refName, limitsRefName string) *model.Schema {
		return &model.Schema{
			Kind:    model.SchemaKindObject,
			RefName: refName,
			Properties: map[string]*model.Schema{
				"url":    {Kind: model.SchemaKindPrimitive, Type: "string"},
				"limits": limits(limitsRefName),
			},
			Required: []string{"url"},
		}
	}
	attrs := func(required []string, settingsRefName, limitsRefName string) *model.Schema {
		return &model.Schema{
			Kind:     model.SchemaKindObject,
			Required: required,
			Properties: map[string]*model.Schema{
				"priority": {
					Kind: model.SchemaKindPrimitive, Type: "string", RefName: "WidgetPriority",
					Enum: []string{"low", "high"},
				},
				"expires_at":  {Kind: model.SchemaKindPrimitive, Type: "string", Format: "date-time"},
				"external_id": {Kind: model.SchemaKindPrimitive, Type: "string", Format: "uuid"},
				"retries":     {Kind: model.SchemaKindPrimitive, Type: "integer", Format: "int32"},
				"settings":    settings(settingsRefName, limitsRefName),
			},
		}
	}
	body := func(a *model.Schema) *model.Schema {
		return &model.Schema{Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
			"data": {Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{"attributes": a}},
		}}
	}
	idBinding := func() *model.SDKOperationBinding {
		return &model.SDKOperationBinding{Required: []model.SDKArgument{
			{Name: "widget_id", GoName: "widgetId", GoType: "string", Location: "path", Schema: prim("string", "")},
		}}
	}

	create := &model.Operation{
		Path: "/api/v2/widgets", Method: "POST",
		OperationId: "CreateWidget", Tag: "Widgets",
		RequestRefName: "WidgetCreateRequest", ResponseRefName: "WidgetResponse",
		RequestSchema: body(attrs([]string{"priority"}, "WidgetSettingsCreateRequest", "WidgetLimitsRequest")),
	}
	read := &model.Operation{
		Path: "/api/v2/widgets/{widget_id}", Method: "GET",
		OperationId: "GetWidget", Tag: "Widgets",
		ResponseRefName: "WidgetResponse",
		ResponseSchema:  body(attrs(nil, "WidgetSettingsResponse", "WidgetLimitsResponse")),
		SDKBinding:      idBinding(),
	}
	del := &model.Operation{
		Path: "/api/v2/widgets/{widget_id}", Method: "DELETE",
		OperationId: "DeleteWidget", Tag: "Widgets",
		SDKBinding: idBinding(),
	}
	update := &model.Operation{
		Path: "/api/v2/widgets/{widget_id}", Method: "PATCH",
		OperationId: "UpdateWidget", Tag: "Widgets",
		RequestRefName: "WidgetUpdateRequest", ResponseRefName: "WidgetResponse",
		RequestSchema: body(attrs(nil, "WidgetSettingsUpdateRequest", "WidgetLimitsRequest")),
		SDKBinding:    idBinding(),
	}

	create.Tracking = &model.TrackingFieldMetadata{
		ArtifactKind:  model.ArtifactKindResource,
		ArtifactName:  "widget",
		TfDescription: "A test-only resource exercising nested/enum/format request mapping.",
		IdStrategy:    model.IdStrategyDataID,
	}
	create.ResolvedGroup = &model.ResolvedGroup{Create: create, Read: read, Update: update, Delete: del}
	return create
}
