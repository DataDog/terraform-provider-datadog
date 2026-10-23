package parser

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v4"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// matrixOperation finds a tracked operation in the example-form matrix fixture
// by artifact name, so each spec names the case it is about.
func matrixOperation(artifact string) *model.Operation {
	spec := loadSpecMust("openapi_examples.yaml")
	for _, op := range spec.Operations {
		if op.Tracking != nil && op.Tracking.ArtifactName == artifact {
			return op
		}
	}
	Fail("no tracked operation for artifact " + artifact)
	return nil
}

var _ = Describe("ExtractExamples", func() {
	Describe("media-type forms", func() {
		It("reads a singular example and anchors it at the example field", func() {
			op := matrixOperation("media_singular")
			set := op.RequestExamples.Examples
			Expect(op.RequestExamples.Present).To(BeTrue())
			Expect(op.RequestExamples.MediaType).To(Equal("application/json"))
			Expect(set.Single).NotTo(BeNil())
			Expect(set.Named).To(BeEmpty())

			Expect(set.Single.SourceKind).To(Equal(model.ExampleSourceMedia))
			Expect(set.Single.Name).To(BeEmpty(), "a singular example carries no name")
			Expect(set.Single.Location.String()).To(Equal(
				"spec:/media-singular.post.requestBody.content.application/json.example"))

			body, ok := set.Single.Value.(map[string]any)
			Expect(ok).To(BeTrue())
			data := body["data"].(map[string]any)
			Expect(data["type"]).To(Equal("widget"))
			Expect(data["attributes"].(map[string]any)["name"]).To(Equal("singular-request"))
		})

		It("reads every named example and anchors each at its own key", func() {
			op := matrixOperation("media_named")
			set := op.RequestExamples.Examples
			Expect(set.Single).To(BeNil())
			Expect(set.SortedNames()).To(Equal([]string{"alternate", "default"}))

			Expect(set.Named["default"].Name).To(Equal("default"))
			Expect(set.Named["default"].Location.String()).To(Equal(
				"spec:/media-named.post.requestBody.content.application/json.examples.default"))
			Expect(set.Named["alternate"].Location.String()).To(ContainSubstring(".examples.alternate"))
		})

		// OpenAPI makes the two forms mutually exclusive, so a location
		// declaring both has no determinate value. Extraction records both
		// rather than resolving one way, and Malformed is what makes the
		// target ineligible later.
		It("records both forms when a location declares both, and reports it malformed", func() {
			op := matrixOperation("bad_both_forms")
			set := op.RequestExamples.Examples
			Expect(set.Single).NotTo(BeNil())
			Expect(set.Named).To(HaveKey("default"))
			Expect(set.Malformed()).To(BeTrue())
			Expect(set.Validate()).To(MatchError(ContainSubstring("mutually exclusive")))
		})
	})

	Describe("reusable references", func() {
		// FR-003 and US1 AS3: a referenced example must be represented
		// exactly as a directly declared one would be.
		It("resolves a local reference to the value it names", func() {
			op := matrixOperation("local_reference")
			candidate := op.RequestExamples.Examples.Named["default"]
			Expect(candidate).NotTo(BeNil())
			Expect(candidate.External).To(BeFalse())
			Expect(candidate.Eligible()).To(BeTrue())

			body := candidate.Value.(map[string]any)["data"].(map[string]any)
			Expect(body["attributes"].(map[string]any)["name"]).To(Equal("reusable-request"))
		})

		// An external value is not in the description, so it can never be
		// reproduced from it. The candidate is retained so a diagnostic can
		// name it, but it is never eligible and its value stays nil.
		It("retains an external example as ineligible without fetching it", func() {
			op := matrixOperation("bad_external")
			candidate := op.RequestExamples.Examples.Named["remote"]
			Expect(candidate).NotTo(BeNil())
			Expect(candidate.External).To(BeTrue())
			Expect(candidate.Value).To(BeNil())
			Expect(candidate.Eligible()).To(BeFalse())
			Expect(candidate.IneligibleReason()).To(ContainSubstring("external value"))
		})
	})

	Describe("schema fallbacks", func() {
		It("reads a whole-body schema example", func() {
			op := matrixOperation("schema_example")
			set := op.RequestExamples.Examples
			Expect(set.Single).To(BeNil(), "no media example is declared")
			Expect(set.Named).To(BeEmpty())

			var whole *model.ExampleCandidate
			for _, candidate := range set.SchemaFallback {
				if candidate.SourceKind == model.ExampleSourceSchema {
					whole = candidate
				}
			}
			Expect(whole).NotTo(BeNil())
			body := whole.Value.(map[string]any)["data"].(map[string]any)
			Expect(body["attributes"].(map[string]any)["name"]).To(Equal("schema-level-request"))
		})

		// A property example is only meaningful assembled with its siblings,
		// so the dotted path is what makes it usable.
		It("reads each scalar property example keyed by its dotted path", func() {
			op := matrixOperation("schema_property_example")
			byPath := map[string]any{}
			for _, candidate := range op.RequestExamples.Examples.SchemaFallback {
				if candidate.SourceKind == model.ExampleSourceSchemaProperty {
					byPath[candidate.Location.PropertyPath] = candidate.Value
				}
			}
			Expect(byPath).To(HaveKeyWithValue("data.attributes.name", "property-level-name"))
			Expect(byPath).To(HaveKeyWithValue("data.attributes.replicas", 3))
			Expect(byPath).To(HaveKeyWithValue("data.attributes.enabled", true))
		})

		It("marks a write-only secret property sensitive", func() {
			op := matrixOperation("sensitive_case")
			var token *model.ExampleCandidate
			for _, candidate := range op.RequestExamples.Examples.SchemaFallback {
				if candidate.Location.PropertyPath == "data.attributes.api_token" {
					token = candidate
				}
			}
			Expect(token).NotTo(BeNil())
			Expect(token.Sensitive).To(BeTrue())
		})
	})

	Describe("responses", func() {
		It("records every declared outcome, not only the success one", func() {
			op := matrixOperation("media_singular")
			var statuses []string
			for _, response := range op.ResponseExamples {
				statuses = append(statuses, response.Status)
			}
			Expect(statuses).To(ContainElement("201"))

			success := op.ResponseExampleFor("201")
			Expect(success).NotTo(BeNil())
			Expect(success.BodyPresent).To(BeTrue())
			Expect(success.Examples.Single).NotTo(BeNil())
			// Only the success response carries the normalized schema schema
			// normalization built the provider model from.
			Expect(success.Schema).NotTo(BeNil())
		})

		// FR-017: a 204 or 205 declaring no content is a complete contract,
		// not a missing example.
		It("treats a declared bodyless response as complete without an example", func() {
			op := matrixOperation("bodyless")
			del := op.ResolvedGroup.Delete
			Expect(del).NotTo(BeNil())

			noContent := del.ResponseExampleFor("204")
			Expect(noContent).NotTo(BeNil())
			Expect(noContent.BodyPresent).To(BeFalse())
			Expect(noContent.Bodyless()).To(BeTrue())
			Expect(noContent.Complete()).To(BeTrue())
			Expect(noContent.Examples.Empty()).To(BeTrue())

			reset := del.ResponseExampleFor("205")
			Expect(reset).NotTo(BeNil())
			Expect(reset.Bodyless()).To(BeTrue())
		})
	})

	Describe("parameters", func() {
		It("reads path and query parameter examples with their serialization", func() {
			op := matrixOperation("parameters_case")

			tags := op.ParameterExampleFor("tags", model.ParameterInQuery)
			Expect(tags).NotTo(BeNil())
			Expect(tags.Examples.Single.Value).To(Equal([]any{"alpha", "beta"}))
			Expect(tags.Examples.Single.SourceKind).To(Equal(model.ExampleSourceParameter))
			Expect(tags.Examples.Single.Location.String()).To(Equal(
				"spec:/parameters/{widget_id}.get.parameters.tags.example"))

			archived := op.ParameterExampleFor("include_archived", model.ParameterInQuery)
			Expect(archived.Examples.Single.Value).To(Equal(true))

			reserved := op.ParameterExampleFor("reserved_filter", model.ParameterInQuery)
			Expect(reserved.Examples.Single.Value).To(Equal("env:prod AND name:a/b c"))

			// Declared on the path item, inherited by the operation.
			id := op.ParameterExampleFor("widget_id", model.ParameterInPath)
			Expect(id).NotTo(BeNil())
			Expect(id.Required).To(BeTrue())
			Expect(id.Examples.Single.Value).To(Equal("99999999-9999-9999-9999-999999999999"))
		})

		// Serialization lives on QueryParam alone, which resolves the
		// location- and style-dependent defaults. Asserting it here rather
		// than on the example contract keeps one authority for what a
		// parameter serializes to.
		It("leaves serialization to the normalized parameter", func() {
			op := matrixOperation("parameters_case")
			byName := map[string]model.QueryParam{}
			for _, p := range append(op.QueryParams, op.PathParams...) {
				byName[p.Name] = p
			}

			Expect(byName["tags"].ResolvedStyle()).To(Equal(model.ParameterStyleForm))
			Expect(byName["tags"].ResolvedExplode()).To(BeFalse(), "the fixture declares explode: false")
			// An omitted explode defaults to true for form style, which is why
			// the flag is a pointer rather than a plain bool.
			Expect(byName["include_archived"].ResolvedExplode()).To(BeTrue())
			Expect(byName["widget_id"].ResolvedStyle()).To(Equal(model.ParameterStyleSimple))
			Expect(byName["widget_id"].ResolvedExplode()).To(BeFalse())
		})

		// Reading only the operation's own parameters would drop the path
		// parameter naming the object it acts on, making the URL unbuildable.
		It("merges path-item parameters with the operation's own, operation winning", func() {
			op := matrixOperation("parameter_merge")

			shared := op.ParameterExampleFor("shared", model.ParameterInQuery)
			Expect(shared).NotTo(BeNil())
			Expect(shared.Examples.Single.Value).To(Equal("from-operation"))

			inherited := op.ParameterExampleFor("only_path_item", model.ParameterInQuery)
			Expect(inherited).NotTo(BeNil())
			Expect(inherited.Examples.Single.Value).To(Equal("path-item-only"))

			Expect(op.ParameterExampleFor("widget_id", model.ParameterInPath)).NotTo(BeNil())
		})
	})

	Describe("the coherent Twilio resource fixture", func() {
		It("extracts a full CRUD contract whose create request and response agree", func() {
			spec, err := LoadSpec("../testdata/mini-oas/mini-datadog_integration_twilio_account.yaml")
			Expect(err).To(Succeed())

			var read *model.Operation
			for _, op := range spec.Operations {
				if op.Tracking != nil && op.Tracking.ArtifactName == "integration_twilio_account" {
					read = op
				}
			}
			Expect(read).NotTo(BeNil())
			group := read.ResolvedGroup
			Expect(group.Create).NotTo(BeNil())
			Expect(group.Update).NotTo(BeNil())
			Expect(group.Delete).NotTo(BeNil())

			request := group.Create.RequestExamples
			Expect(request).NotTo(BeNil())
			Expect(request.Required).To(BeTrue())
			// This fixture uses the named form throughout, so the candidates
			// live under "default" rather than in Single.
			Expect(request.Examples.Single).To(BeNil())
			requestAttrs := request.Examples.Named["default"].
				Value.(map[string]any)["data"].(map[string]any)["attributes"].(map[string]any)

			response := group.Create.ResponseExampleFor("201")
			Expect(response).NotTo(BeNil())
			responseAttrs := response.Examples.Named["default"].
				Value.(map[string]any)["data"].(map[string]any)["attributes"].(map[string]any)

			// The fixture exists to be coherent: shared attributes must agree,
			// or a replayed create would fail Terraform's post-apply
			// consistency check.
			Expect(responseAttrs["name"]).To(Equal(requestAttrs["name"]))
			Expect(responseAttrs["settings"].(map[string]any)["account_sid"]).
				To(Equal(requestAttrs["settings"].(map[string]any)["account_sid"]))

			// writeOnly: the password is sent and never returned.
			Expect(requestAttrs["authentication"].(map[string]any)).To(HaveKey("password"))
			Expect(responseAttrs["authentication"].(map[string]any)).NotTo(HaveKey("password"))

			// The update must describe a distinct state, else an update step
			// would assert nothing.
			updateResponse := group.Update.ResponseExampleFor("200")
			updatedAttrs := updateResponse.Examples.Named["default"].
				Value.(map[string]any)["data"].(map[string]any)["attributes"].(map[string]any)
			Expect(updatedAttrs["name"]).NotTo(Equal(responseAttrs["name"]))
		})
	})

	Describe("untracked operations", func() {
		It("leaves an operation no tracking group reaches unfilled", func() {
			spec := loadSpecMust("tracking_valid.yaml")
			for _, op := range spec.Operations {
				if op.Tracking != nil {
					continue
				}
				reached := false
				for _, other := range spec.Operations {
					if other.Tracking == nil || other.ResolvedGroup == nil {
						continue
					}
					for _, target := range other.ResolvedGroup.Operations() {
						if target == op {
							reached = true
						}
					}
				}
				if !reached {
					Expect(op.RequestExamples).To(BeNil())
					Expect(op.ResponseExamples).To(BeEmpty())
					Expect(op.ParameterExamples).To(BeEmpty())
				}
			}
		})
	})
})

var _ = Describe("decodeExampleValue", func() {
	node := func(source string) *yaml.Node {
		var doc yaml.Node
		Expect(yaml.Unmarshal([]byte(source), &doc)).To(Succeed())
		return doc.Content[0]
	}

	DescribeTable("decodes the supported semantic value set",
		func(source string, want any) {
			got, err := decodeExampleValue(node(source))
			Expect(err).To(Succeed())
			Expect(got).To(Equal(want))
		},
		Entry("string", `"twilio-prod"`, "twilio-prod"),
		Entry("integer", `3`, 3),
		Entry("float", `1.5`, 1.5),
		Entry("boolean", `true`, true),
		Entry("sequence", `[alpha, beta]`, []any{"alpha", "beta"}),
		Entry("mapping", `{name: widget, enabled: false}`,
			map[string]any{"name": "widget", "enabled": false}),
		Entry("nested", `{data: {attributes: {tags: [a]}}}`,
			map[string]any{"data": map[string]any{"attributes": map[string]any{"tags": []any{"a"}}}}),
	)

	// A declared null is a value the description chose, distinct from the
	// candidate not existing at all.
	It("decodes an explicit null to a nil value without erroring", func() {
		got, err := decodeExampleValue(node(`null`))
		Expect(err).To(Succeed())
		Expect(got).To(BeNil())
	})

	It("reports an absent node rather than panicking", func() {
		_, err := decodeExampleValue(nil)
		Expect(err).To(MatchError(ContainSubstring("no example node")))
	})

	// A value the renderer cannot represent would surface as a replay
	// mismatch rather than a generation error, so it is rejected up front.
	It("rejects a non-string mapping key, naming where it appears", func() {
		_, err := decodeExampleValue(node(`{data: {1: alpha}}`))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported example value type"))
		Expect(err.Error()).To(ContainSubstring("data"))
	})

	It("rejects an unsupported value nested inside a sequence", func() {
		_, err := decodeExampleValue(node(`{items: [{2: two}]}`))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("items[0]"))
	})

	It("names the example root when the top-level value is unsupported", func() {
		err := checkSupportedValue(map[int]any{1: "x"}, "")
		Expect(err).To(MatchError(ContainSubstring("the example root")))
	})
})
