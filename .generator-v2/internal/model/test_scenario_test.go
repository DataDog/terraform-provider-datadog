package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// validScenario is the minimum sound scenario; specs mutate one field from it
// so each failure reason is asserted in isolation.
func validScenario() *GeneratedTestScenario {
	return &GeneratedTestScenario{
		ArtifactName:     "integration_twilio_account",
		ArtifactKind:     ArtifactKindResource,
		TestFilePath:     "datadog/tests/resource_datadog_integration_twilio_account_openapi_example_test.go",
		TestFuncName:     "TestAccDatadogIntegrationTwilioAccountOpenAPIExample",
		TerraformAddress: "datadog_integration_twilio_account.foo",
		Steps:            []ScenarioStep{{Config: `resource "datadog_x" "foo" {}`}},
	}
}

var _ = Describe("ExampleSelection", func() {
	It("reports an explicit name for an operation that has one", func() {
		s := &ExampleSelection{ExplicitNames: map[string]string{"CreateX": "alternate"}}
		name, ok := s.Explicit("CreateX")
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("alternate"))

		_, ok = s.Explicit("CreateY")
		Expect(ok).To(BeFalse())
	})

	It("orders explicit operations lexically so diagnostics never vary", func() {
		s := &ExampleSelection{ExplicitNames: map[string]string{
			"UpdateX": "a", "CreateX": "b", "DeleteX": "c",
		}}
		Expect(s.SortedExplicitOperations()).To(Equal([]string{"CreateX", "DeleteX", "UpdateX"}))
	})
})

var _ = Describe("MaterializedConfiguration", func() {
	It("orders replaced sensitive paths lexically", func() {
		c := &MaterializedConfiguration{SensitiveReplacements: map[string]string{
			"data.attributes.token":    RedactedPlaceholder,
			"data.attributes.password": RedactedPlaceholder,
		}}
		Expect(c.SortedSensitivePaths()).To(Equal([]string{
			"data.attributes.password", "data.attributes.token",
		}))
	})

})

var _ = Describe("GeneratedTestScenario", func() {
	Describe("Validate", func() {
		It("accepts a sound scenario", func() {
			Expect(validScenario().Validate()).To(Succeed())
		})

		DescribeTable("rejects a structurally incomplete scenario",
			func(mutate func(*GeneratedTestScenario), want string) {
				s := validScenario()
				mutate(s)
				Expect(s.Validate()).To(MatchError(ContainSubstring(want)))
			},
			Entry("no artifact name", func(s *GeneratedTestScenario) { s.ArtifactName = "" }, "no artifact name"),
			Entry("no test function", func(s *GeneratedTestScenario) { s.TestFuncName = "" }, "no test function name"),
			Entry("no steps", func(s *GeneratedTestScenario) { s.Steps = nil }, "no Terraform steps"),
			Entry("no Terraform address", func(s *GeneratedTestScenario) { s.TerraformAddress = "" }, "no Terraform address"),
		)

		It("rejects a nil scenario rather than panicking", func() {
			var s *GeneratedTestScenario
			Expect(s.Validate()).To(MatchError(ContainSubstring("nil")))
		})
	})

	Describe("HasUpdateStep", func() {
		// An update whose response equals the create response asserts nothing,
		// so the step exists only when examples give a distinct state.
		It("is true only when a second step was materialized", func() {
			Expect(validScenario().HasUpdateStep()).To(BeFalse())
			s := validScenario()
			s.Steps = append(s.Steps, ScenarioStep{Config: "updated"})
			Expect(s.HasUpdateStep()).To(BeTrue())
			var nilScenario *GeneratedTestScenario
			Expect(nilScenario.HasUpdateStep()).To(BeFalse())
		})
	})
})

var _ = Describe("NewCassetteDiagnostic", func() {
	// Locations are addresses and always safe to commit; candidate values are
	// not. Passing a location keeps the anchor machine-readable and the
	// message value-free.
	It("anchors the diagnostic at the example location", func() {
		d := NewCassetteDiagnostic(SeverityWarning, "no usable response example",
			ExampleLocation{
				Path: "/api/v2/widgets", Method: "POST", Component: ExampleComponentResponse,
				Detail: "201", MediaType: "application/json", Field: "examples", ExampleName: "default",
			})
		Expect(d.Severity).To(Equal(SeverityWarning))
		Expect(d.Message).To(Equal("no usable response example"))
		Expect(d.Location).To(Equal(
			"spec:/api/v2/widgets.post.responses.201.content.application/json.examples.default"))
	})
})

var _ = Describe("Redact", func() {
	It("replaces every occurrence of each sensitive value", func() {
		got := Redact("token=s3cr3t and again s3cr3t", []string{"s3cr3t"})
		Expect(got).To(Equal("token=" + RedactedPlaceholder + " and again " + RedactedPlaceholder))
		Expect(got).NotTo(ContainSubstring("s3cr3t"))
	})

	It("handles several values and ignores empty ones", func() {
		got := Redact("a=one b=two", []string{"one", "", "two"})
		Expect(got).To(Equal("a=" + RedactedPlaceholder + " b=" + RedactedPlaceholder))
	})

	It("leaves a message with nothing to redact untouched", func() {
		Expect(Redact("all clear", nil)).To(Equal("all clear"))
	})
})
