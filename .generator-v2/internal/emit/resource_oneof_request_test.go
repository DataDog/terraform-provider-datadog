package emit

import (
	"errors"
	"go/format"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/sdkbind"
)

// roleBasicAuth is one role's spelling of an object alternative. The request
// bodies declare a password the response does not, which is the shape that
// makes cloning one body's union wrong (T099b).
func roleBasicAuth(refName string, password bool) *model.Schema {
	properties := map[string]*model.Schema{
		"username": {Kind: model.SchemaKindPrimitive, Type: "string"},
	}
	required := []string{"username"}
	if password {
		properties["password"] = &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", Sensitive: true}
		required = append(required, "password")
	}
	return &model.Schema{Kind: model.SchemaKindObject, RefName: refName, Properties: properties, Required: required}
}

// roleUnion assembles one role's union node. The SDK bindings are left to
// sdkbind (see buildUnionView), which is what keeps the fixture from freezing
// today's wrapper/member/constructor derivation into an assertion.
func roleUnion(unionRef string, alternative *model.Schema) *model.Schema {
	member := alternative.RefName
	if alternative.Kind != model.SchemaKindObject {
		member = model.UpperFirst(alternative.Type)
	}
	tfName := model.SnakeCase(member)
	return &model.Schema{Kind: model.SchemaKindOneOf, RefName: unionRef, OneOf: &model.OneOfSpec{
		Name: unionRef, Path: "resource.data.attributes.auth", RefName: unionRef,
		Variants: []model.OneOfVariant{{
			TFName: tfName, GoName: model.SdkName(tfName), Schema: alternative,
			RefName:      alternative.RefName,
			ValueWrapped: alternative.Kind != model.SchemaKindObject,
		}},
	}}
}

// widgetUnionOperation is widgetResourceOperation plus a required "auth" union
// spelled three ways — the ordinary Datadog integration-account shape.
func widgetUnionOperation(create, update, read *model.Schema) *model.Operation {
	op := widgetResourceOperation()
	group := op.ResolvedGroup
	for _, side := range []struct {
		body  *model.Schema
		union *model.Schema
	}{
		{group.Create.RequestSchema, create},
		{group.Update.RequestSchema, update},
		{group.Read.ResponseSchema, read},
	} {
		attributes := side.body.Properties["data"].Properties["attributes"]
		attributes.Properties["auth"] = side.union
	}
	createAttributes := group.Create.RequestSchema.Properties["data"].Properties["attributes"]
	createAttributes.Required = append(createAttributes.Required, "auth")
	return op
}

// buildUnionView runs the production pipeline over op: sdkbind resolves every
// role's SDK oneOf binding, then the merge and the emit builder run.
func buildUnionView(op *model.Operation) ResourceView {
	GinkgoHelper()
	group := op.ResolvedGroup
	for _, role := range []*model.Operation{group.Create, group.Update, group.Read} {
		Expect(sdkbind.BindOperation(role)).To(Succeed())
	}
	art, err := model.BuildArtifact(op)
	Expect(err).NotTo(HaveOccurred())
	view, err := BuildResourceView(art)
	Expect(err).NotTo(HaveOccurred())
	return view
}

func threeRoleUnionOperation() *model.Operation {
	return widgetUnionOperation(
		roleUnion("WidgetAuthRequest", roleBasicAuth("WidgetBasicAuthRequest", true)),
		roleUnion("WidgetAuthUpdate", roleBasicAuth("WidgetBasicAuthUpdate", true)),
		roleUnion("WidgetAuthResponse", roleBasicAuth("WidgetBasicAuthResponse", false)),
	)
}

var _ = Describe("BuildResourceView oneOf request expansion", func() {
	It("attaches plan-time exactly-one validation to every configurable variant", func() {
		op := threeRoleUnionOperation()
		for _, role := range []*model.Operation{op.ResolvedGroup.Create, op.ResolvedGroup.Update, op.ResolvedGroup.Read} {
			union := role.RequestSchema
			if role == op.ResolvedGroup.Read {
				union = role.ResponseSchema
			}
			auth := union.Properties["data"].Properties["attributes"].Properties["auth"]
			auth.OneOf.Variants = append(auth.OneOf.Variants, model.OneOfVariant{
				TFName: "string", GoName: "String",
				Schema:       &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"},
				ValueWrapped: true,
			})
		}

		view := buildUnionView(op)
		auth := blockByName(view.Schema.Blocks, "auth")
		stringVariant := blockByName(auth.Blocks, "string")
		basicVariant := blockByName(auth.Blocks, "widget_basic_auth")

		Expect(stringVariant.Validators).To(Equal([]string{
			`objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("widget_basic_auth"))`,
		}))
		Expect(basicVariant.Validators).To(Equal([]string{
			`objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("string"))`,
		}))
		Expect(view.UsesObjectValidators).To(BeTrue())

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())
		out := string(src)
		Expect(out).To(ContainSubstring(`"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"`))
		Expect(out).To(ContainSubstring(`Validators: []validator.Object{`))
		Expect(out).To(ContainSubstring(
			`objectvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("widget_basic_auth")),`))
	})

	It("requires the sole variant without inventing a sibling path", func() {
		view := buildUnionView(threeRoleUnionOperation())
		auth := blockByName(view.Schema.Blocks, "auth")
		variant := blockByName(auth.Blocks, "widget_basic_auth")

		Expect(variant.Validators).To(Equal([]string{
			`objectvalidator.ExactlyOneOf()`,
		}))
	})

	It("expands the selected variant into each role's own SDK wrapper, not the merged tree's", func() {
		view := buildUnionView(threeRoleUnionOperation())

		By("Create builds the request wrapper through the request alternative's own constructor")
		create := requestFieldByGoField(view.Create.Envelope.Fields, "Auth").OneOf
		Expect(create).NotTo(BeNil())
		Expect(create.SDKType).To(Equal("datadogV2.WidgetAuthRequest"))
		Expect(requestFieldByGoField(view.Create.Envelope.Fields, "Auth").NullCheck).To(
			Equal("state.Auth != nil"))
		Expect(create.Var).To(Equal("authUnion"))
		Expect(create.MatchVar).To(Equal("authMatches"))
		Expect(create.Variants).To(HaveLen(1))
		Expect(create.Variants[0].TFName).To(Equal("widget_basic_auth"))
		Expect(create.Variants[0].ModelExpr).To(Equal("state.Auth.WidgetBasicAuth"))
		Expect(create.Variants[0].Constructor).To(Equal("datadogV2.NewWidgetBasicAuthRequestWithDefaults()"))
		Expect(create.Variants[0].WrapCall).To(Equal(
			"datadogV2.WidgetBasicAuthRequestAsWidgetAuthRequest(widgetBasicAuthVariant)"))

		By("Update builds a genuinely different wrapper, member and constructor")
		update := requestFieldByGoField(view.Update.Envelope.Fields, "Auth").OneOf
		Expect(update.SDKType).To(Equal("datadogV2.WidgetAuthUpdate"))
		Expect(update.Variants[0].Constructor).To(Equal("datadogV2.NewWidgetBasicAuthUpdateWithDefaults()"))
		Expect(update.Variants[0].WrapCall).To(Equal(
			"datadogV2.WidgetBasicAuthUpdateAsWidgetAuthUpdate(widgetBasicAuthVariant)"))

		By("the request-only password is configurable and sent, which cloning the response would have dropped")
		password := requestFieldByGoField(create.Variants[0].Fields, "Password")
		Expect(password.Required).To(BeTrue())
		Expect(password.ValueExpr).To(Equal("state.Auth.WidgetBasicAuth.Password.ValueString()"))

		By("the selection diagnostic names the union's path and its variants")
		Expect(create.SelectionMessage).To(Equal(
			`resource.data.attributes.auth: exactly one of "widget_basic_auth" must be set, got %d`))
	})

	It("renders a required union as a guarded, counted expansion that fails before the SDK call", func() {
		view := buildUnionView(threeRoleUnionOperation())
		Expect(view.UsesFmt).To(BeTrue())

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())
		_, err = format.Source(src)
		Expect(err).NotTo(HaveOccurred(), "rendered output must already be gofmt-canonical Go:\n%s", src)

		out := string(src)
		for _, want := range []string{
			`if state.Auth != nil {`,
			`var authUnion datadogV2.WidgetAuthRequest`,
			`authMatches := 0`,
			`if state.Auth.WidgetBasicAuth != nil {`,
			`widgetBasicAuthVariant := datadogV2.NewWidgetBasicAuthRequestWithDefaults()`,
			`widgetBasicAuthVariant.SetPassword(state.Auth.WidgetBasicAuth.Password.ValueString())`,
			`authUnion = datadogV2.WidgetBasicAuthRequestAsWidgetAuthRequest(widgetBasicAuthVariant)`,
			`authMatches++`,
			`if authMatches != 1 {`,
			`bodyAttributes.SetAuth(authUnion)`,
			`} else {`,
			`var authUnion datadogV2.WidgetAuthUpdate`,
		} {
			Expect(out).To(ContainSubstring(want), "generated resource missing %q:\n%s", want, out)
		}

		By("the selection check precedes the SDK call in the body it guards")
		check := strings.Index(out, "authMatches != 1")
		Expect(check).To(BeNumerically(">", 0))
		Expect(check).To(BeNumerically("<", strings.Index(out, "r.Api.CreateWidget(")))
	})

	It("preserves request-only fields when the response selects the configured variant", func() {
		view := buildUnionView(threeRoleUnionOperation())
		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())

		out := string(src)
		Expect(out).To(ContainSubstring(`if state.Auth != nil && state.Auth.WidgetBasicAuth != nil {`))
		Expect(out).To(ContainSubstring(`widgetBasicAuthModel = state.Auth.WidgetBasicAuth`))
		Expect(out).To(ContainSubstring(`settingsModel := state.Settings`))
		Expect(out).To(ContainSubstring(`if settingsModel == nil {`))
	})

	It("converts a value-wrapped scalar alternative straight into the SDK member", func() {
		scalar := func(refName string) *model.Schema {
			// A value-wrapped alternative's member is the scalar itself, so the
			// union's own component name is all that differs by role.
			return roleUnion(refName, &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"})
		}
		view := buildUnionView(widgetUnionOperation(
			scalar("WidgetAuthRequest"), scalar("WidgetAuthUpdate"), scalar("WidgetAuthResponse")))

		variant := requestFieldByGoField(view.Create.Envelope.Fields, "Auth").OneOf.Variants[0]
		Expect(variant.TFName).To(Equal("string"))
		Expect(variant.Constructor).To(BeEmpty())
		Expect(variant.Value).NotTo(BeNil())
		Expect(variant.Value.ValueExpr).To(Equal("state.Auth.String.Value.ValueString()"))
		Expect(variant.WrapCall).To(Equal("datadogV2.StringAsWidgetAuthRequest(&stringVariant)"))

		src, err := RenderResource(view)
		Expect(err).NotTo(HaveOccurred())
		_, err = format.Source(src)
		Expect(err).NotTo(HaveOccurred(), "rendered output must already be gofmt-canonical Go:\n%s", src)
		Expect(string(src)).To(ContainSubstring(`stringVariant := state.Auth.String.Value.ValueString()`))
	})

	It("rejects a collection whose element is a union rather than sending an unexpanded list", func() {
		listOf := func(union *model.Schema) *model.Schema {
			return &model.Schema{Kind: model.SchemaKindArray, Items: union}
		}
		op := widgetUnionOperation(
			listOf(roleUnion("WidgetAuthRequest", roleBasicAuth("WidgetBasicAuthRequest", true))),
			listOf(roleUnion("WidgetAuthUpdate", roleBasicAuth("WidgetBasicAuthUpdate", true))),
			listOf(roleUnion("WidgetAuthResponse", roleBasicAuth("WidgetBasicAuthResponse", false))),
		)
		for _, role := range []*model.Operation{op.ResolvedGroup.Create, op.ResolvedGroup.Update, op.ResolvedGroup.Read} {
			Expect(sdkbind.BindOperation(role)).To(Succeed())
		}
		art, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())

		_, err = BuildResourceView(art)
		var unsupported *UnsupportedEmitError
		Expect(errors.As(err, &unsupported)).To(BeTrue())
		Expect(unsupported.Error()).To(ContainSubstring("collection whose element is a oneOf union"))
	})
})
