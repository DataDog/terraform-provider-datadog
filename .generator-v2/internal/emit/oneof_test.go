package emit

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/sdkbind"
)

// These specs cover the emit half of the typed oneOf envelope: the schema blocks
// and the generated model structs. The response and request mapping specs belong
// here too and land with the state mapper.

// unionOperation is the incident_type GET-by-id with a union grafted onto its
// attributes, built the way the parser would: a component-backed union whose
// alternatives are one object and one scalar. Bindings are resolved through
// sdkbind, so the fixture cannot drift from the derivation under test elsewhere.
func unionOperation() *model.Operation {
	op := incidentTypeOperation()
	attrs := op.ResponseSchema.Properties["data"].Properties["attributes"]
	attrs.Properties["notification"] = unionSchema("response.data.attributes.notification", "IncidentNotification")
	return op
}

// unionSchema builds a normalized component-backed union with one object
// alternative (webhook) and one string alternative, in the shape parser.normalizeOneOf
// produces: variants sorted by TFName, GoName derived, RefName retained.
func unionSchema(path, component string) *model.Schema {
	webhook := obj(map[string]*model.Schema{
		"url":     prim("string", "Webhook URL."),
		"enabled": prim("boolean", "Whether the webhook is enabled."),
	})
	webhook.RefName = "WebhookNotification"

	return &model.Schema{
		Kind:        model.SchemaKindOneOf,
		RefName:     component,
		Description: "How the incident notifies.",
		OneOf: &model.OneOfSpec{
			Name:    component,
			Path:    path,
			RefName: component,
			Variants: []model.OneOfVariant{
				{
					TFName:       "string",
					GoName:       "String",
					Schema:       prim("string", "A raw notification target."),
					ValueWrapped: true,
				},
				{
					TFName:  "webhook_notification",
					GoName:  "WebhookNotification",
					Schema:  webhook,
					RefName: "WebhookNotification",
				},
			},
		},
	}
}

// unionView binds and builds the view for op, asserting neither step failed.
func unionView(op *model.Operation) DataSourceView {
	GinkgoHelper()
	Expect(sdkbind.BindOperation(op)).To(Succeed())
	art, err := model.BuildArtifact(op)
	Expect(err).NotTo(HaveOccurred())
	view, err := BuildDataSourceView(art)
	Expect(err).NotTo(HaveOccurred())
	return view
}

// modelByName returns the generated struct with the given name.
func modelByName(view DataSourceView, name string) ModelStructView {
	GinkgoHelper()
	for _, m := range view.Models {
		if m.Name == name {
			return m
		}
	}
	Fail("no generated model named " + name)
	return ModelStructView{}
}

// oneOfAssignmentAt finds the oneOf assignment for the union at path, searching
// the state mapper's assignments and recursing through variants.
func oneOfAssignmentAt(view DataSourceView, path string) *OneOfAssignment {
	GinkgoHelper()
	var find func(lists []ListAssignment) *OneOfAssignment
	find = func(lists []ListAssignment) *OneOfAssignment {
		for _, l := range lists {
			if l.Kind == "oneof" && l.OneOf != nil {
				if l.OneOf.Path == path {
					return l.OneOf
				}
				for _, v := range l.OneOf.Variants {
					if got := find(v.Lists); got != nil {
						return got
					}
				}
				continue
			}
			if got := find(l.Lists); got != nil {
				return got
			}
		}
		return nil
	}
	got := find(view.State.Lists)
	if got == nil {
		got = find(view.State.ItemLists)
	}
	Expect(got).NotTo(BeNil(), "no oneOf assignment at %q", path)
	return got
}

// variantAssignment returns the alternative selected by the given SDK member.
func variantAssignment(assign *OneOfAssignment, sdkField string) OneOfVariantAssignment {
	GinkgoHelper()
	for _, v := range assign.Variants {
		if v.SDKField == sdkField {
			return v
		}
	}
	Fail("no variant bound to SDK member " + sdkField)
	return OneOfVariantAssignment{}
}

// blockByName returns the named block from a block list.
func blockByName(blocks []AttrView, name string) AttrView {
	GinkgoHelper()
	for _, b := range blocks {
		if b.TFName == name {
			return b
		}
	}
	Fail("no block named " + name)
	return AttrView{}
}

var _ = Describe("oneOf envelope emission", func() {

	Context("the envelope model", func() {
		It("holds one pointer field per variant, keyed on the Terraform variant name", func() {
			view := unionView(unionOperation())
			envelope := modelByName(view, "IncidentNotificationModel")

			Expect(envelope.Fields).To(Equal([]ModelFieldView{
				{GoField: "String", GoType: "*IncidentNotificationStringModel", TFName: "string"},
				{GoField: "WebhookNotification", GoType: "*IncidentNotificationWebhookNotificationModel", TFName: "webhook_notification"},
			}))
		})

		It("points the parent field at the envelope model, not at types.Object", func() {
			// The projection leaves GoType as types.Object; a generated model field has
			// to be a pointer to the envelope struct or nothing selects a variant.
			view := unionView(unionOperation())
			parent := modelByName(view, "datadogIncidentTypeDataSourceModel")

			var field ModelFieldView
			for _, f := range parent.Fields {
				if f.TFName == "notification" {
					field = f
				}
			}
			Expect(field.GoField).To(Equal("Notification"))
			Expect(field.GoType).To(Equal("*IncidentNotificationModel"))
		})

		It("gives an object variant its own fields and a scalar variant a single value field", func() {
			view := unionView(unionOperation())

			Expect(modelByName(view, "IncidentNotificationWebhookNotificationModel").Fields).To(Equal([]ModelFieldView{
				{GoField: "Enabled", GoType: "types.Bool", TFName: "enabled"},
				{GoField: "Url", GoType: "types.String", TFName: "url"},
			}))
			Expect(modelByName(view, "IncidentNotificationStringModel").Fields).To(Equal([]ModelFieldView{
				{GoField: "Value", GoType: "types.String", TFName: "value"},
			}))
		})

		It("declares the envelope struct before its variant structs", func() {
			view := unionView(unionOperation())
			var envelopeAt, variantAt = -1, -1
			for i, m := range view.Models {
				switch m.Name {
				case "IncidentNotificationModel":
					envelopeAt = i
				case "IncidentNotificationWebhookNotificationModel":
					variantAt = i
				}
			}
			Expect(envelopeAt).To(BeNumerically(">=", 0))
			Expect(variantAt).To(BeNumerically(">", envelopeAt))
		})
	})

	Context("the schema", func() {
		It("renders the union as a single nested block holding one block per variant", func() {
			view := unionView(unionOperation())
			block := blockByName(view.Schema.Blocks, "notification")

			Expect(block.IsBlock).To(BeTrue())
			Expect(block.ListBlock).To(BeFalse())
			Expect(block.Description).To(Equal("How the incident notifies."))

			var variants []string
			for _, v := range block.Blocks {
				variants = append(variants, v.TFName)
				Expect(v.IsBlock).To(BeTrue())
			}
			Expect(variants).To(Equal([]string{"string", "webhook_notification"}))
		})

		It("exposes an object variant's fields as attributes and a scalar variant's as value", func() {
			view := unionView(unionOperation())
			block := blockByName(view.Schema.Blocks, "notification")

			webhook := blockByName(block.Blocks, "webhook_notification")
			var names []string
			for _, a := range webhook.Attributes {
				names = append(names, a.TFName)
			}
			Expect(names).To(Equal([]string{"enabled", "url"}))

			scalar := blockByName(block.Blocks, "string")
			Expect(scalar.Attributes).To(HaveLen(1))
			Expect(scalar.Attributes[0].TFName).To(Equal("value"))
		})
	})

	Context("reuse of one component at two sites", func() {
		It("generates a single model and identical blocks", func() {
			// A reusable union is one generated envelope, not one per use site.
			op := incidentTypeOperation()
			attrs := op.ResponseSchema.Properties["data"].Properties["attributes"]
			attrs.Properties["primary"] = unionSchema("response.data.attributes.primary", "IncidentNotification")
			attrs.Properties["secondary"] = unionSchema("response.data.attributes.secondary", "IncidentNotification")

			view := unionView(op)

			var envelopeModels int
			for _, m := range view.Models {
				if m.Name == "IncidentNotificationModel" {
					envelopeModels++
				}
			}
			Expect(envelopeModels).To(Equal(1), "the component yielded one model per use site")

			Expect(blockByName(view.Schema.Blocks, "primary").Blocks).
				To(Equal(blockByName(view.Schema.Blocks, "secondary").Blocks))

			// The two sites share the variant bodies but not the outer locals, so
			// each unwraps into its own envelope model without colliding.
			primary := oneOfAssignmentAt(view, "response.data.attributes.primary")
			secondary := oneOfAssignmentAt(view, "response.data.attributes.secondary")
			Expect(primary.Variants).To(Equal(secondary.Variants))
			Expect(primary.ModelVar).NotTo(Equal(secondary.ModelVar))
			Expect(primary.LHS).To(Equal("state.Primary"))
			Expect(secondary.LHS).To(Equal("state.Secondary"))
		})
	})

	Context("the assignment handed to the template", func() {
		It("unwraps the wrapper through an ordinary getter and counts every member", func() {
			view := unionView(unionOperation())
			assign := oneOfAssignmentAt(view, "response.data.attributes.notification")

			Expect(assign.SDKType).To(Equal("IncidentNotification"))
			Expect(assign.GoModel).To(Equal("IncidentNotificationModel"))
			Expect(assign.LHS).To(Equal("state.Notification"))
			// The wrapper itself is a normal field on its parent; only its members
			// lack getters.
			Expect(assign.GetterOk).To(Equal("attributes.GetNotificationOk()"))
			Expect(assign.Receiver).To(Equal(assign.Var))
			Expect(assign.MatchVar).NotTo(BeEmpty())
			Expect(assign.Collection).To(BeFalse())
			Expect(assign.Variants).To(HaveLen(2))
		})

		It("reads an object alternative's fields through the SDK member's getters", func() {
			view := unionView(unionOperation())
			webhook := variantAssignment(
				oneOfAssignmentAt(view, "response.data.attributes.notification"), "WebhookNotification")

			Expect(webhook.GoField).To(Equal("WebhookNotification"))
			Expect(webhook.SDKVar).To(Equal("webhookNotificationVariant"))
			Expect(webhook.ModelVar).To(Equal("webhookNotificationModel"))
			Expect(webhook.SDKPointer).To(BeTrue())
			Expect(webhook.Value).To(BeNil())
			Expect(webhook.Scalars).To(ContainElement(StateAssignment{
				Var:      "url",
				GetterOk: "webhookNotificationVariant.GetUrlOk()",
				LHS:      "webhookNotificationModel.Url",
				RHS:      "types.StringValue(*url)",
			}))
		})

		It("dereferences a scalar alternative instead of calling a getter on it", func() {
			// The SDK member of a value-wrapped alternative is a *string, which has no
			// GetValueOk; reading through one would not compile.
			view := unionView(unionOperation())
			scalar := variantAssignment(
				oneOfAssignmentAt(view, "response.data.attributes.notification"), "String")

			Expect(scalar.Scalars).To(BeEmpty())
			Expect(scalar.Value).NotTo(BeNil())
			Expect(*scalar.Value).To(Equal(StateAssignment{
				LHS: "stringModel.Value",
				RHS: "types.StringValue(*stringVariant)",
			}))
		})
	})

	Context("placements the emit path does not represent", func() {
		It("fails the artifact for a union at a map value rather than dropping it", func() {
			// This one is rejected before the envelope branch ever sees it: a map under
			// attributes is unsupported for reasons that predate oneOf. What matters for
			// the generator is that it fails rather than generating an artifact missing a union,
			// so assert the outcome, not which layer produced it.
			op := incidentTypeOperation()
			attrs := op.ResponseSchema.Properties["data"].Properties["attributes"]
			attrs.Properties["by_key"] = &model.Schema{
				Kind:  model.SchemaKindMap,
				Items: unionSchema("response.data.attributes.by_key{}", "IncidentNotification"),
			}

			Expect(sdkbind.BindOperation(op)).To(Succeed())
			art, err := model.BuildArtifact(op)
			Expect(err).NotTo(HaveOccurred())

			_, err = BuildDataSourceView(art)
			Expect(err).To(HaveOccurred())

			var unsupported *UnsupportedEmitError
			Expect(errors.As(err, &unsupported)).To(BeTrue())
			Expect(unsupported.Error()).To(ContainSubstring("response.data.attributes.by_key"))
		})

		DescribeTable("oneOfFieldType accepts the forms a union can wear and rejects the rest",
			func(tfType, want string, ok bool) {
				got, derived := oneOfFieldType(tfType, "FooModel")
				Expect(derived).To(Equal(ok))
				Expect(got).To(Equal(want))
			},
			// A union at its own position: one envelope, held by pointer.
			Entry("single nested block", "schema.SingleNestedBlock", "*FooModel", true),
			Entry("single nested attribute", "schema.SingleNestedAttribute", "*FooModel", true),
			// A collection whose element is a union: one envelope per element.
			Entry("list nested block", "schema.ListNestedBlock", "[]*FooModel", true),
			Entry("list nested attribute", "schema.ListNestedAttribute", "[]*FooModel", true),
			// Anything else has no representation yet, and must fail rather than
			// silently produce no field.
			Entry("map nested attribute", "schema.MapNestedAttribute", "", false),
			Entry("scalar", "schema.StringAttribute", "", false),
		)

		It("names the envelope and the form in the placement diagnostic", func() {
			node := unsupportedOneOfPlacement(
				&model.OneOfEnvelope{Name: "Credentials", Path: "response.data.attributes.by_key{}"},
				"schema.MapNestedAttribute",
			)
			Expect(node.Path).To(Equal("response.data.attributes.by_key{}"))
			Expect(node.Reason).To(ContainSubstring("Credentials"))
			Expect(node.Reason).To(ContainSubstring("schema.MapNestedAttribute"))
			Expect(node.Reason).To(ContainSubstring("does not represent yet"))
		})
	})

	Context("rendering", func() {
		It("produces gofmt-clean Go for an artifact carrying a union", func() {
			// RenderDataSource runs format.Source, so a malformed envelope block or
			// model struct fails here rather than at the provider's build.
			src, err := RenderDataSource(unionView(unionOperation()))
			Expect(err).NotTo(HaveOccurred())

			out := string(src)
			Expect(out).To(ContainSubstring("type IncidentNotificationModel struct {"))
			Expect(out).To(ContainSubstring("WebhookNotification *IncidentNotificationWebhookNotificationModel `tfsdk:\"webhook_notification\"`"))
			Expect(out).To(ContainSubstring(`"notification": schema.SingleNestedBlock{`))
			Expect(out).To(ContainSubstring(`"webhook_notification": schema.SingleNestedBlock{`))
		})

		It("renders deterministically", func() {
			first, err := RenderDataSource(unionView(unionOperation()))
			Expect(err).NotTo(HaveOccurred())
			second, err := RenderDataSource(unionView(unionOperation()))
			Expect(err).NotTo(HaveOccurred())
			Expect(first).To(Equal(second))
		})
	})
})

// unionListOperation puts a list whose element is a union onto the attributes, so
// the collection placement is exercised as well as the union-at-its-own-position one.
func unionListOperation() *model.Operation {
	op := incidentTypeOperation()
	attrs := op.ResponseSchema.Properties["data"].Properties["attributes"]
	attrs.Properties["notifications"] = &model.Schema{
		Kind:        model.SchemaKindArray,
		Description: "Every configured notification.",
		Items:       unionSchema("response.data.attributes.notifications[]", "IncidentNotification"),
	}
	return op
}

var _ = Describe("oneOf envelope in a collection", func() {
	It("holds the envelope by slice and renders a list nested block of variants", func() {
		view := unionView(unionListOperation())

		parent := modelByName(view, "datadogIncidentTypeDataSourceModel")
		var field ModelFieldView
		for _, f := range parent.Fields {
			if f.TFName == "notifications" {
				field = f
			}
		}
		Expect(field.GoType).To(Equal("[]*IncidentNotificationModel"),
			"each element of the list is one envelope")

		block := blockByName(view.Schema.Blocks, "notifications")
		Expect(block.ListBlock).To(BeTrue())
		var variants []string
		for _, v := range block.Blocks {
			variants = append(variants, v.TFName)
		}
		Expect(variants).To(Equal([]string{"string", "webhook_notification"}))
	})

	It("renders gofmt-clean Go", func() {
		src, err := RenderDataSource(unionView(unionListOperation()))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(src)).To(ContainSubstring(`"notifications": schema.ListNestedBlock{`))
	})
})

// oneOfEnvelopeView and oneOfEnvelopeListView back the golden entries in
// templates_test.go, covering a union at its own position and one as a list element.
func oneOfEnvelopeView() DataSourceView     { return unionView(unionOperation()) }
func oneOfEnvelopeListView() DataSourceView { return unionView(unionListOperation()) }

var _ = Describe("oneOf response mapping", func() {
	It("inspects every member rather than taking the first non-nil one", func() {
		// The SDK's own MarshalJSON and GetActualInstance are first-match; the
		// generated mapper deliberately is not.
		src := string(mustRender(unionOperation()))
		Expect(src).To(ContainSubstring("if stringVariant := notification.String; stringVariant != nil {"))
		Expect(src).To(ContainSubstring("if webhookNotificationVariant := notification.WebhookNotification; webhookNotificationVariant != nil {"))
		Expect(src).To(ContainSubstring("notificationMatches++"))
	})

	It("reports an unparsed payload at the union's schema path", func() {
		src := string(mustRender(unionOperation()))
		Expect(src).To(ContainSubstring("case notification.UnparsedObject != nil:"))
		Expect(src).To(ContainSubstring("response.data.attributes.notification: the Datadog API returned a value none of the IncidentNotification alternatives could parse"))
	})

	It("reports multiple populated members instead of choosing one", func() {
		src := string(mustRender(unionOperation()))
		Expect(src).To(ContainSubstring("case notificationMatches > 1:"))
		Expect(src).To(ContainSubstring("alternatives, expected exactly one"))
	})

	It("assigns the envelope only when exactly one member matched", func() {
		src := string(mustRender(unionOperation()))
		Expect(src).To(ContainSubstring("case notificationMatches == 1:\n\t\t\tstate.Notification = notificationEnvelope"))
	})

	It("errors on zero members when the union is required", func() {
		op := unionOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"]
		attrs.Required = []string{"notification"}

		src := string(mustRender(op))
		Expect(src).To(ContainSubstring("matched none of the IncidentNotification alternatives, and the field is not optional"))
	})

	It("accepts zero members when the union is optional", func() {
		// Zero populated members may represent an absent nullable value, so
		// an optional envelope must not raise — it simply stays nil.
		op := unionOperation()
		attrs := op.ResponseSchema.Properties["data"].Properties["attributes"]
		attrs.Properties["notification"].OneOf.Optional = true

		src := string(mustRender(op))
		Expect(src).NotTo(ContainSubstring("matched none of the IncidentNotification alternatives"))
		Expect(src).To(ContainSubstring("case notificationMatches == 1:"))
	})

	It("maps a union nested inside another union's alternative", func() {
		op := unionOperation()
		webhook := op.ResponseSchema.Properties["data"].
			Properties["attributes"].Properties["notification"].OneOf.Variants[1].Schema
		webhook.Properties["target"] = unionSchema(
			"response.data.attributes.notification.webhook_notification.target", "NotificationTarget")

		src := string(mustRender(op))
		// The inner wrapper is reached with an ordinary getter off the outer variant.
		Expect(src).To(ContainSubstring("if target, ok := webhookNotificationVariant.GetTargetOk(); ok && target != nil {"))
		Expect(src).To(ContainSubstring("targetEnvelope := &NotificationTargetModel{}"))
		Expect(src).To(ContainSubstring("targetMatches++"))
	})

	It("imports fmt, which the ambiguous-match diagnostic needs", func() {
		// go/format does not prune imports, so an under-estimate here is a compile
		// error and an over-estimate is an unused import.
		view := unionView(unionOperation())
		Expect(view.UsesFmt).To(BeTrue())
		Expect(string(mustRender(unionOperation()))).To(ContainSubstring("\"fmt\""))
	})
})

// mustRender renders op's data source, asserting every stage succeeded.
func mustRender(op *model.Operation) []byte {
	GinkgoHelper()
	src, err := RenderDataSource(unionView(op))
	Expect(err).NotTo(HaveOccurred())
	return src
}
