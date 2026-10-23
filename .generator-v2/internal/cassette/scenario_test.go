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

})
