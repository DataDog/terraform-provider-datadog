package cassette

import (
	"fmt"
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
		Create:       group.Create,
		Read:         group.Read,
		Update:       group.Update,
		Delete:       group.Delete,
		Selection:    selection,
	}
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
			Expect(scenario.TerraformAddress).To(Equal("datadog_integration_twilio_account.foo"))
			Expect(scenario.TestFilePath).To(ContainSubstring("openapi_example_test.go"))
		})

		// A declared secret must not reach the generated configuration, which is
		// committed; the real value belongs only in a recording run's input.
		It("configures the redacted secret, never the declared one", func() {
			scenario, err := BuildResourceScenario(twilioTarget())
			Expect(err).To(Succeed())
			for _, step := range scenario.Steps {
				for _, value := range step.State.RequestValues {
					Expect(fmt.Sprint(value.Value)).NotTo(ContainSubstring("twilio-basic-auth-secret"))
				}
			}
		})

	})

	Describe("the update step", func() {

	})

	Describe("refresh expansion", func() {

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
			Entry("no create", func(t *ResourceTarget) { t.Create = nil }, "no create operation"),
			Entry("no read", func(t *ResourceTarget) { t.Read = nil }, "no read operation"),
			Entry("no delete", func(t *ResourceTarget) { t.Delete = nil }, "no delete operation"),
		)

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

var _ = Describe("BuildResourceScenario edge paths", func() {
	// An update the description does not fully describe is not fatal: a PATCH
	// body requires nothing, so an empty one materializes, and the step is
	// dropped because it would assert the state the create already asserted.
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

})

var _ = Describe("validation as a scenario stage", func() {
	// attributeSchema walks a JSON:API request schema to one attribute, so a
	// case can contradict exactly one leaf of a real fixture.
	attributeSchema := func(root *model.Schema, name string) *model.Schema {
		Expect(root).NotTo(BeNil())
		data, ok := root.Properties["data"]
		Expect(ok).To(BeTrue(), "fixture request schema has no data member")
		attributes, ok := data.Properties["attributes"]
		Expect(ok).To(BeTrue(), "fixture request schema has no data.attributes")
		attribute, ok := attributes.Properties[name]
		Expect(ok).To(BeTrue(), "fixture declares no attribute %q", name)
		return attribute
	}

	// A real fixture with one leaf's enum narrowed so the example it already
	// declares no longer satisfies it. The description now contradicts itself
	// exactly as a hand-written one would.
	It("fails the scenario when a declared example contradicts its schema", func() {
		target := twilioTarget()
		name := attributeSchema(target.Create.RequestExamples.Schema, "name")
		name.Enum = []string{"a-name-the-example-does-not-use"}

		_, err := BuildResourceScenario(target)
		Expect(err).To(HaveOccurred())

		var conformance *ConformanceError
		Expect(err).To(BeAssignableToTypeOf(conformance))
		Expect(err.Error()).To(ContainSubstring("CreateTwilioIntegrationAccount"))
		Expect(err.Error()).To(ContainSubstring("data.attributes.name"))
	})

	// The contradiction must not be reported with the value in it: the same
	// path could just as easily hold a credential.
	It("names the path without quoting the offending value", func() {
		target := twilioTarget()
		name := attributeSchema(target.Create.RequestExamples.Schema, "name")
		name.Enum = []string{"a-name-the-example-does-not-use"}

		_, err := BuildResourceScenario(target)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).NotTo(ContainSubstring("twilio-prod"))
	})

	// The fixture is the one the stack records end to end, so it must pass
	// untouched. This is the regression guard for wiring validation into a
	// live path: a false violation here would make a working artifact
	// ineligible.
	It("leaves the untouched fixture eligible", func() {
		scenario, err := BuildResourceScenario(twilioTarget())
		Expect(err).To(Succeed())
		Expect(scenario.Steps).NotTo(BeEmpty())
	})
})
