package cassette

import (
	"encoding/json"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

var frozen = time.Date(2026, 6, 25, 8, 30, 50, 0, time.UTC)

// twilioTarget builds a real resource target from the coherent fixture.
func twilioTarget() ResourceTarget {
	ops := fixtureOperations(
		filepath.Join("mini-oas", "mini-datadog_integration_twilio_account.yaml"),
		"integration_twilio_account")
	selection, err := Select(ops)
	Expect(err).To(Succeed())

	var read *model.Operation
	for _, op := range ops {
		if op.Tracking != nil && op.Tracking.ArtifactName == "integration_twilio_account" {
			read = op
		}
	}
	Expect(read).NotTo(BeNil())
	group := read.ResolvedGroup

	return ResourceTarget{
		ArtifactName: "integration_twilio_account",
		ServerURL:    "https://api.datadoghq.com",
		IdStrategy:   model.IdStrategyDataID,
		FreezeTime:   frozen,
		Create:       group.Create,
		Read:         group.Read,
		Update:       group.Update,
		Delete:       group.Delete,
		Selection:    selection,
	}
}

func roles(scenario *model.GeneratedTestScenario) []model.InteractionRole {
	var out []model.InteractionRole
	for _, interaction := range scenario.Interactions {
		out = append(out, interaction.Role)
	}
	return out
}

var _ = Describe("BuildResourceScenario", func() {
	Describe("the Twilio resource", func() {
		It("produces a validated scenario", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			Expect(scenario.Validate()).To(Succeed())
		})

		It("derives the test identity the replay harness expects", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())

			Expect(scenario.TestFuncName).To(Equal("TestAccDatadogIntegrationTwilioAccountOpenAPIExample"))
			// The harness derives the cassette from t.Name(), so these must agree.
			Expect(scenario.CassetteBaseName).To(Equal(scenario.TestFuncName))
			Expect(scenario.TerraformAddress).To(Equal("datadog_integration_twilio_account.foo"))
			Expect(scenario.TestFilePath).To(ContainSubstring("openapi_example_test.go"))
			Expect(scenario.FreezeTime).To(Equal(frozen))
			Expect(scenario.FreezeTime.Location()).To(Equal(time.UTC))
		})

		// The Twilio fixture's update describes a distinct state, so the
		// scenario earns its second step.
		It("follows create, refresh, update, refresh, destroy, verify", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			Expect(roles(scenario)).To(Equal([]model.InteractionRole{
				model.InteractionRoleCreate,
				model.InteractionRoleRead,
				model.InteractionRoleRefresh,
				model.InteractionRoleRefresh,
				model.InteractionRoleUpdate,
				model.InteractionRoleRefresh,
				model.InteractionRoleRefresh,
				model.InteractionRoleDelete,
				model.InteractionRoleDestroyVerification,
			}))
			Expect(scenario.HasUpdateStep()).To(BeTrue())
			Expect(scenario.Steps).To(HaveLen(2))
		})

		// One identity, minted once from the create response. A cassette whose
		// create response and later request targets disagree cannot replay.
		It("mints one identity and reuses it in every later URL", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())

			var created map[string]any
			Expect(json.Unmarshal([]byte(scenario.Interactions[0].Response.Body), &created)).To(Succeed())
			identity := created["data"].(map[string]any)["id"].(string)
			Expect(identity).NotTo(BeEmpty())

			Expect(scenario.Interactions[0].Request.URL).To(Equal(
				"https://api.datadoghq.com/api/v2/integration-interfaces/twilio/accounts"))
			for _, interaction := range scenario.Interactions[1:] {
				Expect(interaction.Request.URL).To(HaveSuffix("/accounts/"+identity),
					"interaction %d addresses a different object", interaction.Index)
			}
		})

		It("sends a create body and no body on a read or delete", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())

			create := scenario.Interactions[0]
			Expect(create.Request.Method).To(Equal("POST"))
			Expect(create.Request.Body).NotTo(BeEmpty())
			Expect(create.Request.ContentLength).To(Equal(len(create.Request.Body)))
			Expect(create.Request.Headers).To(HaveKey("Content-Type"))

			for _, interaction := range scenario.Interactions {
				switch interaction.Role {
				case model.InteractionRoleRead, model.InteractionRoleRefresh,
					model.InteractionRoleDelete, model.InteractionRoleDestroyVerification:
					Expect(interaction.Request.Body).To(BeEmpty())
					Expect(interaction.Request.Headers).NotTo(HaveKey("Content-Type"))
				}
			}
		})

		// A declared secret must not reach the recorded request bytes.
		It("records the redacted secret, never the declared one", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			Expect(scenario.Interactions[0].Request.Body).To(ContainSubstring(model.RedactedPlaceholder))
			for _, interaction := range scenario.Interactions {
				Expect(interaction.Request.Body).NotTo(ContainSubstring("twilio-basic-auth-secret"))
				Expect(interaction.Response.Body).NotTo(ContainSubstring("twilio-basic-auth-secret"))
			}
		})

		It("ends with a delete and a 404 verification", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())

			last := scenario.Interactions[len(scenario.Interactions)-1]
			Expect(last.Role).To(Equal(model.InteractionRoleDestroyVerification))
			Expect(last.Response.StatusCode).To(Equal(404))
			Expect(last.Request.Method).To(Equal("GET"))

			del := scenario.Interactions[len(scenario.Interactions)-2]
			Expect(del.Role).To(Equal(model.InteractionRoleDelete))
			Expect(del.Request.Method).To(Equal("DELETE"))
			// The Twilio delete declares 200, not 204; the matcher compares codes.
			Expect(del.Response.StatusCode).To(Equal(200))
		})

		// Recorded latency would break byte-identical regeneration and changes
		// nothing about replay.
		It("records zero duration and only allowlisted headers", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			for _, interaction := range scenario.Interactions {
				Expect(interaction.Response.Duration).To(BeZero())
				for name := range interaction.Request.Headers {
					Expect(model.RetainedHeaders()).To(ContainElement(name))
				}
				for name := range interaction.Response.Headers {
					Expect(model.RetainedHeaders()).To(ContainElement(name))
				}
			}
		})

		It("carries provenance on the interactions built from examples", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			Expect(scenario.Interactions[0].SourceExamples).NotTo(BeEmpty())
			Expect(scenario.Selection.ScenarioName).To(Equal("default"))
		})

		It("produces identical bytes on repeated builds", func() {
			first, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			second, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())

			encode := func(s *model.GeneratedTestScenario) string {
				out, err := json.Marshal(s.Interactions)
				Expect(err).To(Succeed())
				return string(out)
			}
			Expect(encode(first)).To(Equal(encode(second)))
		})
	})

	Describe("the update step", func() {
		// An update whose response equals the create response asserts nothing,
		// so the scenario keeps the create-only flow instead.
		It("is omitted when the update response matches the create response", func() {
			target := twilioTarget()
			createResponse := target.Create.SuccessResponseExample()
			updateResponse := target.Update.SuccessResponseExample()
			updateResponse.Examples.Named["default"].Value = createResponse.Examples.Named["default"].Value

			scenario, err := BuildResourceScenario(target)
			Expect(err).To(Succeed())
			Expect(scenario.HasUpdateStep()).To(BeFalse())
			Expect(scenario.Steps).To(HaveLen(1))
			Expect(roles(scenario)).To(Equal([]model.InteractionRole{
				model.InteractionRoleCreate,
				model.InteractionRoleRead,
				model.InteractionRoleRefresh,
				model.InteractionRoleDelete,
				model.InteractionRoleDestroyVerification,
			}))
		})

		It("is omitted when the target declares no update operation", func() {
			target := twilioTarget()
			target.Update = nil
			scenario, err := BuildResourceScenario(target)
			Expect(err).To(Succeed())
			Expect(scenario.HasUpdateStep()).To(BeFalse())
			Expect(scenario.Interactions).To(HaveLen(5))
		})
	})

	Describe("refresh expansion", func() {
		// Identical requests are repeated rather than shared, because a
		// replayed interaction is consumed once.
		It("repeats the read once per expected refresh", func() {
			target := twilioTarget()
			target.Update = nil
			target.RefreshesAfterApply = 3

			scenario, err := BuildResourceScenario(target)
			Expect(err).To(Succeed())
			Expect(scenario.Interactions).To(HaveLen(6))

			reads := 0
			for _, interaction := range scenario.Interactions {
				if interaction.Role == model.InteractionRoleRead ||
					interaction.Role == model.InteractionRoleRefresh {
					reads++
				}
			}
			Expect(reads).To(Equal(3))
		})

		It("keeps indexes dense and ordered", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			for i, interaction := range scenario.Interactions {
				Expect(interaction.Index).To(Equal(i))
			}
		})
	})

	Describe("rejected targets", func() {
		DescribeTable("reports a target that cannot describe a lifecycle",
			func(mutate func(*ResourceTarget), want string) {
				target := twilioTarget()
				mutate(&target)
				_, err := BuildResourceScenario(target)
				Expect(err).To(MatchError(ContainSubstring(want)))
			},
			Entry("no artifact name", func(t *ResourceTarget) { t.ArtifactName = "" }, "no artifact name"),
			Entry("no server URL", func(t *ResourceTarget) { t.ServerURL = "" }, "no server URL"),
			Entry("no freeze time", func(t *ResourceTarget) { t.FreezeTime = time.Time{} }, "no freeze time"),
			Entry("no create", func(t *ResourceTarget) { t.Create = nil }, "no create operation"),
			Entry("no read", func(t *ResourceTarget) { t.Read = nil }, "no read operation"),
			Entry("no delete", func(t *ResourceTarget) { t.Delete = nil }, "no delete operation"),
		)

		// Without an identity every later URL would be wrong, so this fails
		// loudly rather than recording an unusable trace.
		It("reports a create response carrying no identity", func() {
			target := twilioTarget()
			target.IdStrategy = model.IdStrategyDataAttributesUID
			_, err := BuildResourceScenario(target)
			Expect(err).To(MatchError(ContainSubstring("no data.attributes.uuid to use as the resource identity")))
		})

		It("reports a selection missing the create request", func() {
			target := twilioTarget()
			var kept []SelectedSet
			for _, set := range target.Selection.Sets {
				if set.Key.OperationId == target.Create.OperationId && set.Key.Role == SetRoleRequest {
					continue
				}
				kept = append(kept, set)
			}
			target.Selection.Sets = kept
			_, err := BuildResourceScenario(target)
			Expect(err).To(MatchError(ContainSubstring("no selected request example")))
		})
	})
})

var _ = Describe("camelCase", func() {
	DescribeTable("converts a snake_case artifact name",
		func(in, want string) { Expect(camelCase(in)).To(Equal(want)) },
		Entry("multi word", "integration_twilio_account", "IntegrationTwilioAccount"),
		Entry("single word", "team", "Team"),
		Entry("doubled separator", "a__b", "AB"),
		Entry("empty", "", ""),
	)
})

var _ = Describe("deleteStatus", func() {
	It("uses the declared success status", func() {
		op := &model.Operation{ResponseExamples: []model.ResponseExamples{{Status: "200", BodyPresent: true}}}
		Expect(deleteStatus(op)).To(Equal(200))
	})

	It("defaults to 204 when none is declared", func() {
		Expect(deleteStatus(&model.Operation{})).To(Equal(204))
	})
})

var _ = Describe("BuildResourceScenario edge paths", func() {
	// An update the description does not fully describe is not fatal: the
	// scenario keeps the create-only flow it can support.
	It("drops the update step when the update examples are incomplete", func() {
		target := twilioTarget()
		updateRequest := target.Update.RequestExamples
		updateRequest.Examples.Named = map[string]*model.ExampleCandidate{
			"default": {Name: "default", Value: map[string]any{}},
		}
		selection, err := Select([]*model.Operation{
			target.Create, target.Read, target.Update, target.Delete,
		})
		if err == nil {
			target.Selection = selection
		}
		scenario, err := BuildResourceScenario(target)
		Expect(err).To(Succeed())
		Expect(scenario.HasUpdateStep()).To(BeFalse())
	})

	It("reports a read declaring no success response", func() {
		target := twilioTarget()
		target.Read = &model.Operation{OperationId: "ReadWithNoSuccess", Method: "GET", Path: "/x"}
		_, err := BuildResourceScenario(target)
		Expect(err).To(MatchError(ContainSubstring("declares no success response")))
	})

	It("reports a selection missing the read response", func() {
		target := twilioTarget()
		var kept []SelectedSet
		for _, set := range target.Selection.Sets {
			if set.Key.OperationId == target.Read.OperationId && set.Key.Role == SetRoleResponse {
				continue
			}
			kept = append(kept, set)
		}
		target.Selection.Sets = kept
		_, err := BuildResourceScenario(target)
		Expect(err).To(MatchError(ContainSubstring("no selected response example")))
	})

	// A non-identity path parameter takes its own declared example, which is
	// how a sub-resource path stays addressable.
	It("substitutes a declared path example rather than the identity", func() {
		ops := fixtureOperations(filepath.Join("parser", "openapi_examples.yaml"), "bodyless")
		selection, err := Select(ops)
		Expect(err).To(Succeed())

		var read *model.Operation
		for _, op := range ops {
			if op.OperationId == "GetBodyless" {
				read = op
			}
		}
		Expect(read).NotTo(BeNil())

		target := ResourceTarget{
			ArtifactName: "bodyless", ServerURL: "https://api.datadoghq.com",
			FreezeTime: frozen, Selection: selection,
			Create: read, Read: read, Delete: read.ResolvedGroup.Delete,
		}
		scenario, err := BuildResourceScenario(target)
		Expect(err).To(Succeed())
		// widget_id is declared on the create path too, so it is user-supplied
		// and takes its own example rather than the minted identity.
		Expect(scenario.Interactions[0].Request.URL).To(HaveSuffix(
			"/bodyless/99999999-9999-9999-9999-999999999999"))
	})

	It("trims a trailing slash from the server origin", func() {
		target := twilioTarget()
		target.ServerURL = "https://api.datadoghq.com/"
		scenario, err := BuildResourceScenario(target)
		Expect(err).To(Succeed())
		Expect(scenario.Interactions[0].Request.URL).To(Equal(
			"https://api.datadoghq.com/api/v2/integration-interfaces/twilio/accounts"))
	})
})

var _ = Describe("encodeBody", func() {
	It("renders nothing for an absent body", func() {
		Expect(encodeBody(nil)).To(BeEmpty())
	})

	// Go's encoder sorts map keys, which is what byte-identical regeneration
	// depends on.
	It("sorts object keys so the same values always produce the same bytes", func() {
		Expect(encodeBody(map[string]any{"b": 1, "a": 2})).To(Equal(`{"a":2,"b":1}`))
	})

	It("renders nothing for a value JSON cannot represent", func() {
		Expect(encodeBody(func() {})).To(BeEmpty())
	})
})

var _ = Describe("identityPathLabel", func() {
	It("names the default strategy when none is declared", func() {
		Expect(identityPathLabel("")).To(Equal("data.id"))
		Expect(identityPathLabel(model.IdStrategyDataAttributesID)).To(Equal("data.attributes.id"))
	})
})
