package emit

import (
	"errors"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/sdkbind"
)

const testSecretDescription = "Secret password or private key."

func operationAttributes(body *model.Schema) *model.Schema {
	return body.Properties["data"].Properties["attributes"]
}

func addRootSecret(op *model.Operation, name string, schema *model.Schema, requiredOnCreate, requiredOnUpdate bool) {
	create := operationAttributes(op.ResolvedGroup.Create.RequestSchema)
	update := operationAttributes(op.ResolvedGroup.Update.RequestSchema)
	create.Properties[name] = model.CloneSchema(schema)
	update.Properties[name] = model.CloneSchema(schema)
	if requiredOnCreate {
		create.Required = append(create.Required, name)
	}
	if requiredOnUpdate {
		update.Required = append(update.Required, name)
	}
}

func rootWriteOnlyOperation(requiredOnCreate, requiredOnUpdate bool) *model.Operation {
	op := widgetResourceOperation()
	addRootSecret(op, "apiToken", &model.Schema{
		Kind:            model.SchemaKindPrimitive,
		Type:            "string",
		Description:     "API token.",
		Sensitive:       true,
		WriteOnlySecret: true,
	}, requiredOnCreate, requiredOnUpdate)
	return op
}

func oneLevelWriteOnlyOperation() *model.Operation {
	op := widgetResourceOperation()
	for _, body := range []*model.Schema{
		op.ResolvedGroup.Create.RequestSchema,
		op.ResolvedGroup.Update.RequestSchema,
	} {
		url := operationAttributes(body).Properties["settings"].Properties["url"]
		url.Description = testSecretDescription
		url.Sensitive = true
		url.WriteOnlySecret = true
	}
	operationAttributes(op.ResolvedGroup.Read.ResponseSchema).Properties["settings"].Properties["url"].Description =
		"response-description-must-not-enter-warning"
	return op
}

func renameAuthProperty(body *model.Schema, request bool) *model.Schema {
	attributes := operationAttributes(body)
	union := attributes.Properties["auth"]
	delete(attributes.Properties, "auth")
	attributes.Properties["authentication"] = union
	for index, name := range attributes.Required {
		if name == "auth" {
			attributes.Required[index] = "authentication"
		}
	}
	if request {
		union.OneOf.Path = "request.data.attributes.authentication"
	} else {
		union.OneOf.Path = "response.data.attributes.authentication"
	}
	return union.OneOf.Variants[0].Schema
}

func recursiveWriteOnlyOperation(requiredOnUpdate bool) *model.Operation {
	op := widgetUnionOperation(
		roleUnion("IntegrationAccountAuthenticationRequest", roleBasicAuth("IntegrationAccountBasicAuthRequest", true)),
		roleUnion("IntegrationAccountAuthenticationUpdate", roleBasicAuth("IntegrationAccountBasicAuthUpdate", true)),
		roleUnion("IntegrationAccountAuthenticationResponse", roleBasicAuth("IntegrationAccountBasicAuthResponse", false)),
	)
	createBasic := renameAuthProperty(op.ResolvedGroup.Create.RequestSchema, true)
	updateBasic := renameAuthProperty(op.ResolvedGroup.Update.RequestSchema, true)
	responseBasic := renameAuthProperty(op.ResolvedGroup.Read.ResponseSchema, false)

	for _, basic := range []*model.Schema{createBasic, updateBasic} {
		basic.Properties["password"].Description = testSecretDescription
		basic.Properties["password"].Sensitive = true
		basic.Properties["password"].WriteOnlySecret = true
		basic.Properties["username"].Description = "Readable username."
		basic.Required = nil
	}
	createBasic.Required = []string{"password"}
	if requiredOnUpdate {
		updateBasic.Required = []string{"password"}
	}
	responseBasic.Properties["username"].Description = "Readable username."
	responseBasic.Required = nil
	return op
}

func buildResourceForWriteOnly(op *model.Operation) (*model.Artifact, ResourceView) {
	GinkgoHelper()
	for _, role := range []*model.Operation{
		op.ResolvedGroup.Create,
		op.ResolvedGroup.Update,
		op.ResolvedGroup.Read,
	} {
		Expect(sdkbind.BindOperation(role)).To(Succeed())
	}
	artifact, err := model.BuildArtifact(op)
	Expect(err).NotTo(HaveOccurred())
	view, err := BuildResourceView(artifact)
	Expect(err).NotTo(HaveOccurred())
	return artifact, view
}

func renderWriteOnlyResource(op *model.Operation) (*model.Artifact, ResourceView, string) {
	GinkgoHelper()
	artifact, view := buildResourceForWriteOnly(op)
	return artifact, view, string(mustRenderResource(view))
}

var _ = Describe("generated resource write-only-only contract", func() {
	DescribeTable("replaces the original Terraform field at every supported static depth",
		func(build func() *model.Operation, original, goName, parentBlocks, description string) {
			_, _, source := renderWriteOnlyResource(build())

			Expect(source).To(ContainSubstring(`"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"`))
			Expect(source).To(ContainSubstring("fwutils.CreateWriteOnlySecretAttributes"))
			Expect(source).To(ContainSubstring(`WriteOnlyAttr: "` + original + `_wo"`))
			Expect(source).To(ContainSubstring(`TriggerAttr: "` + original + `_wo_version"`))
			Expect(regexp.MustCompile(`ParentBlocks:\s+\[\]string\{` + regexp.QuoteMeta(parentBlocks) + `\}`).MatchString(source)).To(BeTrue())
			Expect(source).To(ContainSubstring("Mode: fwutils.WriteOnlySecretModeOnly"))
			Expect(source).To(ContainSubstring(`WriteOnlyDescription: "` + description + ` This write-only value is not stored in Terraform state."`))
			Expect(source).To(ContainSubstring(`TriggerDescription: "Version trigger for ` + original + `_wo rotation."`))

			Expect(regexp.MustCompile(goName + `Wo\s+types\.String\s+` + "`tfsdk:\"" + original + `_wo"` + "`").MatchString(source)).To(BeTrue())
			Expect(regexp.MustCompile(goName + `WoVersion\s+types\.String\s+` + "`tfsdk:\"" + original + `_wo_version"` + "`").MatchString(source)).To(BeTrue())
			Expect(source).NotTo(ContainSubstring(goName + ` types.String ` + "`tfsdk:\"" + original + `\"` + "`"))
			Expect(strings.Count(source, "`tfsdk:\""+original+"_wo\"`")).To(Equal(1))
			Expect(strings.Count(source, "`tfsdk:\""+original+"_wo_version\"`")).To(Equal(1))
		},
		Entry("resource root", func() *model.Operation { return rootWriteOnlyOperation(true, false) }, "api_token", "ApiToken", "", "API token."),
		Entry("one single-nested ancestor", oneLevelWriteOnlyOperation, "url", "Url", `"settings"`, testSecretDescription),
		Entry("recursive single-nested path through oneOf", func() *model.Operation { return recursiveWriteOnlyOperation(false) }, "password", "Password", `"authentication", "integration_account_basic_auth"`, testSecretDescription),
	)

	DescribeTable("routes the original SDK setter through handlers with role-specific Update requiredness",
		func(requiredOnUpdate bool) {
			_, _, source := renderWriteOnlyResource(recursiveWriteOnlyOperation(requiredOnUpdate))
			Expect(source).To(ContainSubstring("fwutils.WriteOnlySecretHandler"))
			Expect(source).To(ContainSubstring("SecretRequiredOnUpdate: " + map[bool]string{false: "false", true: "true"}[requiredOnUpdate]))
			Expect(source).To(ContainSubstring("GetSecretForCreate(ctx, &request.Config)"))
			Expect(source).To(ContainSubstring("GetSecretForUpdate(ctx, &request.Config, &request)"))
			Expect(source).To(ContainSubstring("SetPassword("))
			Expect(source).NotTo(ContainSubstring("SetPasswordWo("))
			Expect(source).NotTo(ContainSubstring("state.Password.ValueString()"))
		},
		Entry("Update may omit the secret", false),
		Entry("Update requires the secret", true),
	)

	It("keeps nested ancestors configuration-owned and readable siblings independently computed", func() {
		_, view := buildResourceForWriteOnly(recursiveWriteOnlyOperation(false))
		authentication := attrByPath(schemaTree(view), "authentication")
		basic := attrByPath(schemaTree(view), "integration_account_basic_auth")
		username := attrByPath(schemaTree(view), "username")
		for _, ancestor := range []AttrView{authentication, basic} {
			Expect(ancestor.Computed).To(BeFalse())
			Expect(ancestor.PlanModifiers).NotTo(ContainElement("objectplanmodifier.UseStateForUnknown"))
		}
		Expect(username.Optional).To(BeTrue())
		Expect(username.Computed).To(BeTrue())
		Expect(username.PlanModifiers).To(ContainElement("stringplanmodifier.UseStateForUnknown"))
	})

	It("suppresses a same-path Read assignment and emits one value-free warning", func() {
		artifact, _, source := renderWriteOnlyResource(oneLevelWriteOnlyOperation())
		Expect(source).NotTo(ContainSubstring("state.Settings.Url ="))
		var warnings []model.Diagnostic
		for _, diagnostic := range artifact.Diagnostics {
			if diagnostic.Severity == model.SeverityWarning {
				warnings = append(warnings, diagnostic)
			}
		}
		Expect(warnings).To(HaveLen(1))
		Expect(warnings[0].Message).To(ContainSubstring("data.attributes.settings.url"))
		Expect(warnings[0].Message).NotTo(ContainSubstring("response-description-must-not-enter-warning"))
	})

	DescribeTable("does not select write-only handling from display sensitivity",
		func(_ string) {
			op := widgetResourceOperation()
			addRootSecret(op, "credential", &model.Schema{
				Kind: model.SchemaKindPrimitive, Type: "string", Description: "Credential.", Sensitive: true,
			}, false, false)
			_, _, source := renderWriteOnlyResource(op)
			Expect(source).To(ContainSubstring(`"credential": schema.StringAttribute{`))
			credentialSchema := source[strings.Index(source, `"credential": schema.StringAttribute{`):]
			Expect(regexp.MustCompile(`Sensitive:\s+true`).MatchString(credentialSchema[:strings.Index(credentialSchema, "\n\t\t\t},")])).To(BeTrue())
			Expect(regexp.MustCompile("types\\.String\\s+`tfsdk:\"credential\"`").MatchString(source)).To(BeTrue())
			Expect(source).NotTo(ContainSubstring("credential_wo"))
			Expect(source).NotTo(ContainSubstring("WriteOnlySecretHandler"))
			Expect(source).NotTo(ContainSubstring("datadog/internal/fwutils"))
		},
		Entry("x-secret:true only", "x-secret"),
		Entry("tracking sensitive:true only", "tracking-sensitive"),
	)

	It("keeps response/data-source sensitivity ordinary and never emits write-only companions", func() {
		op := incidentTypeOperation()
		secret := &model.Schema{
			Kind: model.SchemaKindPrimitive, Type: "string", Description: "Observed secret.",
			Sensitive: true, WriteOnlySecret: true,
		}
		operationAttributes(op.ResponseSchema).Properties["observed_secret"] = secret
		artifact, err := model.BuildArtifact(op)
		Expect(err).NotTo(HaveOccurred())
		view, err := BuildDataSourceView(artifact)
		Expect(err).NotTo(HaveOccurred())
		source, err := RenderDataSource(view)
		Expect(err).NotTo(HaveOccurred())
		out := string(source)
		Expect(out).To(ContainSubstring(`"observed_secret": schema.StringAttribute{`))
		Expect(regexp.MustCompile(`Sensitive:\s+true`).MatchString(out)).To(BeTrue())
		Expect(out).NotTo(ContainSubstring("observed_secret_wo"))
		Expect(out).NotTo(ContainSubstring("WriteOnlySecretHandler"))
	})

	It("contains no private-state hash, digest comparison, keepers, or generated secret cleanup", func() {
		_, _, source := renderWriteOnlyResource(recursiveWriteOnlyOperation(false))
		lower := strings.ToLower(source)
		for _, forbidden := range []string{"private.set", "sha256", "secret hash", "digest", "keepers", "passwordwo = types.stringnull"} {
			Expect(lower).NotTo(ContainSubstring(forbidden))
		}
	})
})

func collectionNestedSecret(kind model.SchemaKind, refName string) *model.Schema {
	secret := &model.Schema{
		Kind: model.SchemaKindPrimitive, Type: "string", Description: "Nested token.", WriteOnlySecret: true,
	}
	item := &model.Schema{
		Kind: model.SchemaKindObject, RefName: refName,
		Properties: map[string]*model.Schema{"token": secret},
	}
	return &model.Schema{Kind: kind, Items: item, Description: "Credentials collection."}
}

func addUnsupportedWriteOnly(op *model.Operation, name string, schema *model.Schema) {
	for _, body := range []*model.Schema{
		op.ResolvedGroup.Create.RequestSchema,
		op.ResolvedGroup.Update.RequestSchema,
	} {
		operationAttributes(body).Properties[name] = model.CloneSchema(schema)
	}
}

func modelAttributeByPath(attributes []*model.Attribute, path string) *model.Attribute {
	for _, attribute := range attributes {
		if attribute.Path == path {
			return attribute
		}
		if found := modelAttributeByPath(attribute.Children, path); found != nil {
			return found
		}
	}
	return nil
}

var _ = Describe("generated write-only support boundary", func() {
	DescribeTable("fails unsupported shapes before template rendering with their path and category",
		func(build func() *model.Artifact, wantPath, wantCategory string) {
			artifact := build()
			_, err := BuildResourceView(artifact)
			Expect(err).To(HaveOccurred())
			var unsupported *UnsupportedEmitError
			Expect(errors.As(err, &unsupported)).To(BeTrue(), "expected UnsupportedEmitError, got %v", err)
			Expect(strings.ToLower(err.Error())).To(ContainSubstring(strings.ToLower(wantPath)))
			Expect(strings.ToLower(err.Error())).To(ContainSubstring(strings.ToLower(wantCategory)))
		},
		Entry("string nested below a list element",
			func() *model.Artifact {
				op := widgetResourceOperation()
				addUnsupportedWriteOnly(op, "credentials", collectionNestedSecret(model.SchemaKindArray, "CredentialRequest"))
				artifact, err := model.BuildArtifact(op)
				Expect(err).NotTo(HaveOccurred())
				return artifact
			},
			"resource.data.attributes.credentials[].token", "list"),
		Entry("string nested below a map element",
			func() *model.Artifact {
				op := widgetResourceOperation()
				addUnsupportedWriteOnly(op, "credentials", collectionNestedSecret(model.SchemaKindMap, "CredentialRequest"))
				artifact, err := model.BuildArtifact(op)
				Expect(err).NotTo(HaveOccurred())
				return artifact
			},
			"resource.data.attributes.credentials{}.token", "map"),
		Entry("set containment",
			func() *model.Artifact {
				op := widgetResourceOperation()
				addUnsupportedWriteOnly(op, "credentials", collectionNestedSecret(model.SchemaKindArray, "CredentialRequest"))
				artifact, err := model.BuildArtifact(op)
				Expect(err).NotTo(HaveOccurred())
				container := modelAttributeByPath(artifact.Schema.Attributes, "resource.data.attributes.credentials")
				Expect(container).NotTo(BeNil())
				container.TfType = "schema.SetNestedAttribute"
				return artifact
			},
			"resource.data.attributes.credentials[].token", "set"),
		Entry("non-string value",
			func() *model.Artifact {
				op := widgetResourceOperation()
				addUnsupportedWriteOnly(op, "enabled_secret", &model.Schema{
					Kind: model.SchemaKindPrimitive, Type: "boolean", Description: "Unsupported boolean secret.", WriteOnlySecret: true,
				})
				artifact, err := model.BuildArtifact(op)
				Expect(err).NotTo(HaveOccurred())
				return artifact
			},
			"resource.data.attributes.enabled_secret", "string"),
		Entry("_wo companion collision",
			func() *model.Artifact {
				op := rootWriteOnlyOperation(false, false)
				addRootSecret(op, "apiTokenWo", prim("string", "Colliding field."), false, false)
				artifact, err := model.BuildArtifact(op)
				Expect(err).NotTo(HaveOccurred())
				return artifact
			},
			"resource.data.attributes.api_token", "collision"),
		Entry("_wo_version companion collision",
			func() *model.Artifact {
				op := rootWriteOnlyOperation(false, false)
				addRootSecret(op, "apiTokenWoVersion", prim("string", "Colliding field."), false, false)
				artifact, err := model.BuildArtifact(op)
				Expect(err).NotTo(HaveOccurred())
				return artifact
			},
			"resource.data.attributes.api_token", "collision"),
	)
})
