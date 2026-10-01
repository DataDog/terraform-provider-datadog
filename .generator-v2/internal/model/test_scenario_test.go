package model

import (
	"time"

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
		CassetteBaseName: "TestAccDatadogIntegrationTwilioAccountOpenAPIExample",
		FreezeTime:       time.Date(2026, 6, 25, 8, 30, 50, 0, time.UTC),
		Steps:            []ScenarioStep{{Config: `resource "datadog_x" "foo" {}`}},
		Interactions: []ScenarioInteraction{
			{Index: 0, Role: InteractionRoleCreate, OperationId: "CreateX",
				Request: InteractionRequest{Method: "POST", URL: "https://api.datadoghq.com/api/v2/x"}},
		},
	}
}

var _ = Describe("RetainedHeaders", func() {
	It("exposes the allowlist without letting a caller mutate it", func() {
		Expect(RetainedHeaders()).To(Equal([]string{"Accept", "Content-Type"}))
		RetainedHeaders()[0] = "Authorization"
		Expect(RetainedHeaders()).To(Equal([]string{"Accept", "Content-Type"}))
	})
})

var _ = Describe("FilterRetainedHeaders", func() {
	// The provider harness filters everything outside this allowlist, so
	// retaining more would commit bytes that cannot affect matching — and
	// could commit an authorization header.
	It("keeps only Accept and Content-Type, canonicalizing names", func() {
		got := FilterRetainedHeaders(map[string][]string{
			"accept":        {"application/json"},
			"CONTENT-TYPE":  {"application/json"},
			"Authorization": {"Bearer super-secret"},
			"Dd-Api-Key":    {"abc123"},
		})
		Expect(got).To(HaveLen(2))
		Expect(got).To(HaveKeyWithValue("Accept", []string{"application/json"}))
		Expect(got).To(HaveKeyWithValue("Content-Type", []string{"application/json"}))
	})

	It("drops an allowlisted header with no values", func() {
		Expect(FilterRetainedHeaders(map[string][]string{"Accept": {}})).To(BeNil())
	})

	It("returns nil for empty or wholly disallowed input", func() {
		Expect(FilterRetainedHeaders(nil)).To(BeNil())
		Expect(FilterRetainedHeaders(map[string][]string{"Authorization": {"x"}})).To(BeNil())
	})

	It("does not alias the caller's slices", func() {
		in := map[string][]string{"Accept": {"application/json"}}
		got := FilterRetainedHeaders(in)
		in["Accept"][0] = "mutated"
		Expect(got["Accept"][0]).To(Equal("application/json"))
	})
})

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

		// The replay harness derives the cassette name from t.Name(), so a
		// mismatch replays the wrong fixture or none at all.
		It("rejects a cassette basename that does not match the test function", func() {
			s := validScenario()
			s.CassetteBaseName = "TestAccSomethingElse"
			Expect(s.Validate()).To(MatchError(ContainSubstring("does not match test function")))
		})

		It("rejects a freeze time that is absent or not UTC", func() {
			s := validScenario()
			s.FreezeTime = time.Time{}
			Expect(s.Validate()).To(MatchError(ContainSubstring("no freeze time")))

			s = validScenario()
			s.FreezeTime = time.Date(2026, 6, 25, 8, 30, 50, 0, time.FixedZone("CEST", 2*60*60))
			Expect(s.Validate()).To(MatchError(ContainSubstring("not UTC")))
		})

		// Replay interactions are single-use and consumed in order, so a
		// sparse or reordered index set silently misaligns the whole trace.
		It("rejects interaction indexes that are not dense and ordered", func() {
			s := validScenario()
			s.Interactions = append(s.Interactions, ScenarioInteraction{
				Index: 5, Role: InteractionRoleRead,
				Request: InteractionRequest{Method: "GET", URL: "https://api.datadoghq.com/api/v2/x/1"},
			})
			Expect(s.Validate()).To(MatchError(ContainSubstring("dense and ordered")))
		})

		It("rejects an interaction with no request method or URL", func() {
			s := validScenario()
			s.Interactions[0].Request.URL = ""
			Expect(s.Validate()).To(MatchError(ContainSubstring("no request method or URL")))
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
			Entry("no interactions", func(s *GeneratedTestScenario) { s.Interactions = nil }, "no interactions"),
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

var _ = Describe("CassetteBundle", func() {
	complete := func() *CassetteBundle {
		return &CassetteBundle{
			TestPath: "t_test.go", TestContent: []byte("package test"),
			CassettePath: "c.yaml", CassetteContent: []byte("version: 2"),
			FreezePath: "c.freeze", FreezeContent: []byte("2026-06-25T08:30:50Z"),
		}
	}

	It("reports its three paths in a stable order", func() {
		Expect(complete().Paths()).To(Equal([]string{"t_test.go", "c.yaml", "c.freeze"}))
		var b *CassetteBundle
		Expect(b.Paths()).To(BeNil())
	})

	// A half-written bundle replays as a confusing failure rather than an
	// obvious absence, so the writer refuses to touch disk without all three.
	It("is complete only when all three members carry content", func() {
		Expect(complete().Complete()).To(BeTrue())

		missingBody := complete()
		missingBody.FreezeContent = nil
		Expect(missingBody.Complete()).To(BeFalse())

		missingPath := complete()
		missingPath.CassettePath = ""
		Expect(missingPath.Complete()).To(BeFalse())

		var nilBundle *CassetteBundle
		Expect(nilBundle.Complete()).To(BeFalse())
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
