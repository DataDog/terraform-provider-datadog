package cassette

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

func incomplete(err error) *IncompleteError {
	var target *IncompleteError
	ExpectWithOffset(1, errors.As(err, &target)).To(BeTrue(), "want an *IncompleteError, got %v", err)
	return target
}

// materializeRequest selects a target and materializes its create request,
// which is the path every resource scenario depends on.
func materializeRequest(fixture, artifact, operationId string) (MaterializedSet, error) {
	ops := fixtureOperations(fixture, artifact)
	selection, err := Select(ops)
	Expect(err).To(Succeed())

	key := SetKey{OperationId: operationId, Role: SetRoleRequest}
	set, ok := selection.Set(key)
	Expect(ok).To(BeTrue(), "no request set for "+operationId)

	var schema *model.Schema
	for _, op := range ops {
		if op.OperationId == operationId {
			schema = op.RequestExamples.Schema
		}
	}
	return MaterializeSet(set, schema)
}

var _ = Describe("MaterializeSet", func() {
	Describe("a declared whole example", func() {
		It("uses the declared value as the scenario, filtered but not completed", func() {
			got, err := materializeRequest(
				filepath.Join("mini-oas", "mini-datadog_integration_twilio_account.yaml"),
				"integration_twilio_account", "CreateTwilioIntegrationAccount")
			Expect(err).To(Succeed())

			// The normalized request schema is rooted at the JSON:API
			// envelope, so materialized paths carry the data member.
			data := got.Body.(map[string]any)["data"].(map[string]any)
			Expect(data["type"]).To(Equal("integration-account"))
			attributes := data["attributes"].(map[string]any)
			Expect(attributes["name"]).To(Equal("twilio-prod"))
			Expect(attributes["settings"].(map[string]any)["account_sid"]).
				To(Equal("ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))
		})

		// A declared secret must never reach a committed fixture, and the
		// substitution has to be recorded so a writer can prove it happened.
		It("replaces a write-only secret and records the replacement", func() {
			got, err := materializeRequest(
				filepath.Join("mini-oas", "mini-datadog_integration_twilio_account.yaml"),
				"integration_twilio_account", "CreateTwilioIntegrationAccount")
			Expect(err).To(Succeed())

			attributes := got.Body.(map[string]any)["data"].(map[string]any)["attributes"].(map[string]any)
			authentication := attributes["authentication"].(map[string]any)
			Expect(authentication["password"]).To(Equal(model.RedactedPlaceholder))
			Expect(got.SensitiveReplacements).To(HaveKeyWithValue(
				"data.attributes.authentication.password", model.RedactedPlaceholder))
			// The non-secret half of the credential pair is untouched.
			Expect(authentication["username"]).To(Equal("SKxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))

			for _, value := range got.Values {
				Expect(value.Value).NotTo(Equal("twilio-basic-auth-secret"))
			}
		})

		It("records leaf paths and typed values, sorted", func() {
			got, err := materializeRequest(
				filepath.Join("mini-oas", "mini-datadog_integration_twilio_account.yaml"),
				"integration_twilio_account", "CreateTwilioIntegrationAccount")
			Expect(err).To(Succeed())

			var paths []string
			for _, value := range got.Values {
				paths = append(paths, value.Path)
			}
			Expect(paths).To(BeEquivalentTo(sortedCopy(paths)), "leaf values must be stably ordered")
			Expect(paths).To(ContainElement("data.attributes.name"))

			var censor *model.MaterializedValue
			for i := range got.Values {
				if got.Values[i].Path == "data.attributes.settings.censor_logs" {
					censor = &got.Values[i]
				}
			}
			Expect(censor).NotTo(BeNil())
			Expect(censor.Value).To(BeTrue(), "a boolean stays a boolean")
		})

		// The example is the scenario. Completing an omitted optional field
		// would change what the fixture asserts about the API.
		It("does not add a field the example omits", func() {
			schema := &model.Schema{
				Kind:     model.SchemaKindObject,
				Required: []string{"name"},
				Properties: map[string]*model.Schema{
					"name":     {Kind: model.SchemaKindPrimitive, Type: "string"},
					"optional": {Kind: model.SchemaKindPrimitive, Type: "string", HasDefault: true, Default: model.SchemaDefault{Value: model.NewStringDefault("defaulted")}},
				},
			}
			set := SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Candidate: &model.ExampleCandidate{Value: map[string]any{"name": "declared"}},
			}
			got, err := MaterializeSet(set, schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"name": "declared"}))
		})

		It("drops a read-only field from a request but keeps it in a response", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject,
				Properties: map[string]*model.Schema{
					"name":       {Kind: model.SchemaKindPrimitive, Type: "string"},
					"created_at": {Kind: model.SchemaKindPrimitive, Type: "string", ReadOnly: true},
				},
			}
			value := map[string]any{"name": "widget", "created_at": "2026-06-25T08:30:50Z"}

			request, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Candidate: &model.ExampleCandidate{Value: value},
			}, schema)
			Expect(err).To(Succeed())
			Expect(request.Body).To(Equal(map[string]any{"name": "widget"}))

			response, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleResponse, Detail: "201"},
				Candidate: &model.ExampleCandidate{Value: value},
			}, schema)
			Expect(err).To(Succeed())
			Expect(response.Body).To(HaveKeyWithValue("created_at", "2026-06-25T08:30:50Z"))
		})

		It("drops a key the schema does not describe so the request still validates", func() {
			schema := &model.Schema{
				Kind:       model.SchemaKindObject,
				Properties: map[string]*model.Schema{"name": {Kind: model.SchemaKindPrimitive, Type: "string"}},
			}
			got, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Candidate: &model.ExampleCandidate{Value: map[string]any{"name": "a", "stray": "b"}},
			}, schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"name": "a"}))
		})

		It("reports a required field the example omits", func() {
			schema := &model.Schema{
				Kind:     model.SchemaKindObject,
				Required: []string{"name"},
				Properties: map[string]*model.Schema{
					"name": {Kind: model.SchemaKindPrimitive, Type: "string"},
				},
			}
			_, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Candidate: &model.ExampleCandidate{Value: map[string]any{}},
			}, schema)
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("name"))
			Expect(err.Error()).To(ContainSubstring("CreateX request"))
		})

		It("reports an example whose shape disagrees with the schema", func() {
			schema := &model.Schema{
				Kind:       model.SchemaKindObject,
				Properties: map[string]*model.Schema{"name": {Kind: model.SchemaKindPrimitive, Type: "string"}},
			}
			_, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Candidate: &model.ExampleCandidate{Value: "not an object"},
			}, schema)
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("(root)"))
		})
	})

	Describe("assembling property pieces", func() {
		It("assembles a complete value from property examples", func() {
			got, err := materializeRequest(
				filepath.Join("parser", "openapi_examples.yaml"),
				"schema_property_example", "CreateSchemaPropertyExample")
			Expect(err).To(Succeed())

			data := got.Body.(map[string]any)["data"].(map[string]any)
			Expect(data["type"]).To(Equal("widget"))
			attributes := data["attributes"].(map[string]any)
			Expect(attributes["name"]).To(Equal("property-level-name"))
			Expect(attributes["replicas"]).To(Equal(3))
			Expect(attributes["enabled"]).To(BeTrue())
		})

		// This is the case selection deliberately deferred: bad_external
		// resolves via fallback, and completeness is decided here.
		It("reports the required leaves a fallback cannot cover", func() {
			_, err := materializeRequest(
				filepath.Join("parser", "openapi_examples.yaml"),
				"bad_external", "CreateBadExternal")
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("data.attributes.name"))
		})

		It("uses a schema default and a single-member enum, which are not guesses", func() {
			schema := &model.Schema{
				Kind:     model.SchemaKindObject,
				Required: []string{"kind", "retries"},
				Properties: map[string]*model.Schema{
					"kind":    {Kind: model.SchemaKindPrimitive, Type: "string", Enum: []string{"widget"}},
					"retries": {Kind: model.SchemaKindPrimitive, Type: "integer", HasDefault: true, Default: model.SchemaDefault{Value: model.NewInt64Default(3)}},
				},
			}
			got, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Fallbacks: []*model.ExampleCandidate{{Value: "ignored", Location: model.ExampleLocation{PropertyPath: "unrelated"}}},
			}, schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"kind": "widget", "retries": 3}))
		})

		// An empty array or map is a claim about the API, not an absence of
		// information, so it is never invented.
		It("never invents an array or a map", func() {
			fallback := []*model.ExampleCandidate{
				{Value: "x", Location: model.ExampleLocation{PropertyPath: "name"}},
			}
			schema := &model.Schema{
				Kind:     model.SchemaKindObject,
				Required: []string{"name"},
				Properties: map[string]*model.Schema{
					"name":     {Kind: model.SchemaKindPrimitive, Type: "string"},
					"tags":     {Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
					"metadata": {Kind: model.SchemaKindMap},
				},
			}
			got, err := MaterializeSet(SelectedSet{
				Key: SetKey{OperationId: "CreateX", Role: SetRoleRequest}, Fallbacks: fallback,
			}, schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"name": "x"}))
		})

		It("reports a required array it cannot cover", func() {
			schema := &model.Schema{
				Kind:     model.SchemaKindObject,
				Required: []string{"tags"},
				Properties: map[string]*model.Schema{
					"tags": {Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
				},
			}
			_, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Fallbacks: []*model.ExampleCandidate{{Value: "x", Location: model.ExampleLocation{PropertyPath: "other"}}},
			}, schema)
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("tags"))
		})

		It("omits an optional object nothing populated rather than sending it empty", func() {
			schema := &model.Schema{
				Kind:     model.SchemaKindObject,
				Required: []string{"name"},
				Properties: map[string]*model.Schema{
					"name": {Kind: model.SchemaKindPrimitive, Type: "string"},
					"nested": {Kind: model.SchemaKindObject, Properties: map[string]*model.Schema{
						"unset": {Kind: model.SchemaKindPrimitive, Type: "string"},
					}},
				},
			}
			got, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Fallbacks: []*model.ExampleCandidate{{Value: "x", Location: model.ExampleLocation{PropertyPath: "name"}}},
			}, schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"name": "x"}))
		})

		// Choosing a branch would invent a shape the description never
		// committed to.
		It("reports a union offering more than one branch", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject, Required: []string{"auth"},
				Properties: map[string]*model.Schema{
					"auth": {Kind: model.SchemaKindOneOf, Variants: []*model.Schema{
						{Kind: model.SchemaKindPrimitive, Type: "string"},
						{Kind: model.SchemaKindPrimitive, Type: "integer"},
					}},
				},
			}
			_, err := MaterializeSet(SelectedSet{
				Key: SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				// A fallback elsewhere makes the set resolved, so the walk
				// reaches the union rather than stopping at "unresolved".
				Fallbacks: []*model.ExampleCandidate{
					{Value: "x", Location: model.ExampleLocation{PropertyPath: "unrelated"}},
				},
			}, schema)
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Ambiguous).To(ContainElement("auth"))
			Expect(err.Error()).To(ContainSubstring("more than one branch"))
		})

		It("follows a union with exactly one branch, which is no choice at all", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject, Required: []string{"auth"},
				Properties: map[string]*model.Schema{
					"auth": {Kind: model.SchemaKindOneOf, Variants: []*model.Schema{
						{Kind: model.SchemaKindPrimitive, Type: "string"},
					}},
				},
			}
			got, err := MaterializeSet(SelectedSet{
				Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
				Fallbacks: []*model.ExampleCandidate{{Value: "basic", Location: model.ExampleLocation{PropertyPath: "auth"}}},
			}, schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"auth": "basic"}))
		})
	})

	It("reports an unresolved set rather than producing an empty body", func() {
		_, err := MaterializeSet(SelectedSet{Key: SetKey{OperationId: "CreateX", Role: SetRoleRequest}}, nil)
		Expect(err).To(HaveOccurred())
		Expect(incomplete(err).Missing).To(ContainElement("the whole value"))
	})
})

var _ = Describe("IdentityFrom", func() {
	body := map[string]any{
		"data": map[string]any{
			"id":         "953a0060-81ec-4221-aed4-d4733b59cd96",
			"attributes": map[string]any{"id": "attr-id", "uuid": "attr-uuid"},
		},
	}

	DescribeTable("reads the identifier the strategy names",
		func(strategy model.IdStrategy, want string) {
			got, ok := IdentityFrom(body, strategy)
			Expect(ok).To(BeTrue())
			Expect(got).To(Equal(want))
		},
		Entry("data.id", model.IdStrategyDataID, "953a0060-81ec-4221-aed4-d4733b59cd96"),
		Entry("defaulted", model.IdStrategy(""), "953a0060-81ec-4221-aed4-d4733b59cd96"),
		Entry("data.attributes.id", model.IdStrategyDataAttributesID, "attr-id"),
		Entry("data.attributes.uuid", model.IdStrategyDataAttributesUID, "attr-uuid"),
	)

	It("reports a strategy whose value is not in the body", func() {
		_, ok := IdentityFrom(body, model.IdStrategyHeaderLocation)
		Expect(ok).To(BeFalse())
	})

	DescribeTable("reports a body that cannot yield an identity",
		func(in any) {
			_, ok := IdentityFrom(in, model.IdStrategyDataID)
			Expect(ok).To(BeFalse())
		},
		Entry("not an object", "scalar"),
		Entry("no data member", map[string]any{}),
		Entry("data is not an object", map[string]any{"data": "scalar"}),
		Entry("no id member", map[string]any{"data": map[string]any{}}),
		Entry("id is not a string", map[string]any{"data": map[string]any{"id": 7}}),
		Entry("id is empty", map[string]any{"data": map[string]any{"id": ""}}),
	)
})

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

var _ = Describe("MaterializeSet edge paths", func() {
	requestSet := func(value any, fallbacks ...*model.ExampleCandidate) SelectedSet {
		set := SelectedSet{Key: SetKey{OperationId: "CreateX", Role: SetRoleRequest}, Fallbacks: fallbacks}
		if value != nil {
			set.Candidate = &model.ExampleCandidate{Value: value}
		}
		return set
	}

	Describe("declared arrays", func() {
		It("walks each item against the item schema", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject,
				Properties: map[string]*model.Schema{
					"tags": {Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
				},
			}
			got, err := MaterializeSet(requestSet(map[string]any{"tags": []any{"alpha", "beta"}}), schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"tags": []any{"alpha", "beta"}}))

			var paths []string
			for _, value := range got.Values {
				paths = append(paths, value.Path)
			}
			Expect(paths).To(ConsistOf("tags[0]", "tags[1]"))
		})

		It("replaces a secret inside an array item", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject,
				Properties: map[string]*model.Schema{
					"keys": {Kind: model.SchemaKindArray, Items: &model.Schema{
						Kind: model.SchemaKindPrimitive, Type: "string", WriteOnlySecret: true,
					}},
				},
			}
			got, err := MaterializeSet(requestSet(map[string]any{"keys": []any{"s3cr3t"}}), schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"keys": []any{model.RedactedPlaceholder}}))
			Expect(got.SensitiveReplacements).To(HaveKey("keys[0]"))
		})

		It("reports an example that is not an array where the schema says array", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject,
				Properties: map[string]*model.Schema{
					"tags": {Kind: model.SchemaKindArray, Items: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
				},
			}
			_, err := MaterializeSet(requestSet(map[string]any{"tags": "not-a-list"}), schema)
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("tags"))
		})
	})

	Describe("declared unions", func() {
		// With a declared value the branch is already settled, so several
		// alternatives is not ambiguity — unlike when assembling.
		It("passes a declared value through a multi-branch union", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject,
				Properties: map[string]*model.Schema{
					"auth": {Kind: model.SchemaKindOneOf, Variants: []*model.Schema{
						{Kind: model.SchemaKindPrimitive, Type: "string"},
						{Kind: model.SchemaKindPrimitive, Type: "integer"},
					}},
				},
			}
			got, err := MaterializeSet(requestSet(map[string]any{"auth": "basic"}), schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"auth": "basic"}))
		})

		It("reads variants from the normalized OneOf spec, not only the legacy field", func() {
			schema := &model.Schema{
				Kind: model.SchemaKindObject, Required: []string{"auth"},
				Properties: map[string]*model.Schema{
					"auth": {Kind: model.SchemaKindOneOf, OneOf: &model.OneOfSpec{
						Name: "AuthEnvelope",
						Variants: []model.OneOfVariant{
							{TFName: "basic", Schema: &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"}},
						},
					}},
				},
			}
			got, err := MaterializeSet(requestSet(nil,
				&model.ExampleCandidate{Value: "basic", Location: model.ExampleLocation{PropertyPath: "auth"}}), schema)
			Expect(err).To(Succeed())
			Expect(got.Body).To(Equal(map[string]any{"auth": "basic"}))
		})
	})

	Describe("a top-level scalar set", func() {
		It("assembles from a root fallback", func() {
			got, err := MaterializeSet(requestSet(nil,
				&model.ExampleCandidate{Value: "root-value", Location: model.ExampleLocation{PropertyPath: ""}}),
				&model.Schema{Kind: model.SchemaKindPrimitive, Type: "string"})
			// A candidate with no property path is not indexable, so the set
			// has nothing to assemble from and says so.
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("(root)"))
			Expect(got).To(Equal(MaterializedSet{}))
		})

		It("assembles a root scalar from a schema default", func() {
			got, err := MaterializeSet(requestSet(nil,
				&model.ExampleCandidate{Value: "x", Location: model.ExampleLocation{PropertyPath: "ignored"}}),
				&model.Schema{
					Kind: model.SchemaKindPrimitive, Type: "boolean", HasDefault: true,
					Default: model.SchemaDefault{Value: model.NewBoolDefault(true)},
				})
			Expect(err).To(Succeed())
			Expect(got.Body).To(BeTrue())
		})
	})

	Describe("schema defaults", func() {
		DescribeTable("uses each usable scalar default kind",
			func(schema *model.Schema, want any) {
				got, err := MaterializeSet(requestSet(nil,
					&model.ExampleCandidate{Value: "x", Location: model.ExampleLocation{PropertyPath: "other"}}),
					&model.Schema{
						Kind: model.SchemaKindObject, Required: []string{"field"},
						Properties: map[string]*model.Schema{"field": schema},
					})
				Expect(err).To(Succeed())
				Expect(got.Body).To(Equal(map[string]any{"field": want}))
			},
			Entry("string", &model.Schema{Kind: model.SchemaKindPrimitive, Type: "string", HasDefault: true,
				Default: model.SchemaDefault{Value: model.NewStringDefault("s")}}, "s"),
			Entry("bool", &model.Schema{Kind: model.SchemaKindPrimitive, Type: "boolean", HasDefault: true,
				Default: model.SchemaDefault{Value: model.NewBoolDefault(false)}}, false),
			Entry("integer", &model.Schema{Kind: model.SchemaKindPrimitive, Type: "integer", HasDefault: true,
				Default: model.SchemaDefault{Value: model.NewInt64Default(7)}}, 7),
			Entry("number", &model.Schema{Kind: model.SchemaKindPrimitive, Type: "number", HasDefault: true,
				Default: model.SchemaDefault{Value: model.NewFloat64Default(1.5)}}, 1.5),
		)

		// HasDefault stays true for declarations the parser could not decode,
		// so an unusable default must not be mistaken for a value.
		It("ignores a default whose declaration was unusable", func() {
			_, err := MaterializeSet(requestSet(nil,
				&model.ExampleCandidate{Value: "x", Location: model.ExampleLocation{PropertyPath: "other"}}),
				&model.Schema{
					Kind: model.SchemaKindObject, Required: []string{"field"},
					Properties: map[string]*model.Schema{
						"field": {Kind: model.SchemaKindPrimitive, Type: "string", HasDefault: true},
					},
				})
			Expect(err).To(HaveOccurred())
			Expect(incomplete(err).Missing).To(ContainElement("field"))
		})
	})

	// A non-string secret keeps its value here; producing a schema-valid
	// replacement for every type is sanitization's job.
	It("records a non-string secret as sensitive without replacing it", func() {
		schema := &model.Schema{
			Kind: model.SchemaKindObject,
			Properties: map[string]*model.Schema{
				"port": {Kind: model.SchemaKindPrimitive, Type: "integer", Sensitive: true},
			},
		}
		got, err := MaterializeSet(requestSet(map[string]any{"port": 5432}), schema)
		Expect(err).To(Succeed())
		Expect(got.Body).To(Equal(map[string]any{"port": 5432}))
		Expect(got.SensitiveReplacements).To(BeEmpty())
		for _, value := range got.Values {
			if value.Path == "port" {
				Expect(value.Sensitive).To(BeTrue())
			}
		}
	})

	It("is unaffected by a fallback candidate carrying no property path", func() {
		schema := &model.Schema{
			Kind: model.SchemaKindObject, Required: []string{"name"},
			Properties: map[string]*model.Schema{"name": {Kind: model.SchemaKindPrimitive, Type: "string"}},
		}
		_, err := MaterializeSet(SelectedSet{
			Key:       SetKey{OperationId: "CreateX", Role: SetRoleRequest},
			Fallbacks: []*model.ExampleCandidate{nil, {Value: "x"}},
		}, schema)
		Expect(err).To(HaveOccurred())
		Expect(incomplete(err).Missing).To(ContainElement("name"))
	})

	It("passes a value through untouched when the schema is unknown", func() {
		got, err := MaterializeSet(requestSet(map[string]any{"anything": 1}), nil)
		Expect(err).To(Succeed())
		Expect(got.Body).To(Equal(map[string]any{"anything": 1}))
	})
})
