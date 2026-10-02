package model

import (
	"bytes"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/contracts"
)

func runReportSchema() *jsonschema.Schema {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(contracts.RunReportSchema))
	Expect(err).To(Succeed())
	c := jsonschema.NewCompiler()
	Expect(c.AddResource("run-report.schema.json", doc)).To(Succeed())
	compiled, err := c.Compile("run-report.schema.json")
	Expect(err).To(Succeed())
	return compiled
}

// validateAgainstContract round-trips a value through the generator's own JSON
// encoding before validating, so the test checks the bytes tfgen actually
// writes rather than a hand-built document.
func validateAgainstContract(schema *jsonschema.Schema, value any) error {
	encoded, err := json.Marshal(value)
	Expect(err).To(Succeed())
	decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	Expect(err).To(Succeed())
	return schema.Validate(decoded)
}

var _ = Describe("run report contract", func() {
	var schema *jsonschema.Schema
	BeforeEach(func() { schema = runReportSchema() })

	baseReport := func() *RunReport {
		return &RunReport{
			RunId:            "1b109d83-e778-4ef6-bfcb-e35d0fa9bedc",
			GeneratorVersion: "dev",
			SpecHash:         "42f64bf1f82eb4764304662a1ff2ed90f690e18c8d9b59b2bc83ca90fc951fed",
			StartedAt:        time.Date(2026, 9, 30, 15, 59, 25, 0, time.UTC),
			FinishedAt:       time.Date(2026, 9, 30, 15, 59, 26, 0, time.UTC),
			Artifacts: []ArtifactReportEntry{{
				Name:   "integration_twilio_account",
				Kind:   ArtifactKindResource,
				Status: ArtifactStatusCreated,
				Path:   "datadog/fwprovider/resource_datadog_integration_twilio_account.go",
			}},
			Summary: summarize([]ArtifactReportEntry{{Status: ArtifactStatusCreated}}),
		}
	}

	// The whole point of the optional top-level fields: a run that did not ask
	// for cassettes must produce exactly what it produced before the feature.
	It("accepts a report with no cassette fields at all", func() {
		Expect(validateAgainstContract(schema, baseReport())).To(Succeed())
	})

	// A run that produced no artifacts marshals its nil slice to null, since
	// the field carries no omitempty. Real reports do contain "artifacts":
	// null, so the contract must accept it or it would reject tfgen's own
	// output. Pinned here because it is a wart a future change could
	// accidentally "tidy" into a breaking one.
	It("accepts a null artifacts list, which an empty run really emits", func() {
		report := baseReport()
		report.Artifacts = nil
		report.Summary = summarize(nil)
		encoded, err := json.Marshal(report)
		Expect(err).To(Succeed())
		Expect(string(encoded)).To(ContainSubstring(`"artifacts":null`))
		Expect(validateAgainstContract(schema, report)).To(Succeed())
	})

	It("accepts every artifact status and diagnostic severity the model defines", func() {
		report := baseReport()
		report.Artifacts = nil
		for _, status := range []ArtifactStatus{
			ArtifactStatusCreated, ArtifactStatusUpdated, ArtifactStatusUnchanged,
			ArtifactStatusSkipped, ArtifactStatusFailed, ArtifactStatusRetired,
			ArtifactStatusRetireBlocked, ArtifactStatusRegistrationRetired,
		} {
			report.Artifacts = append(report.Artifacts, ArtifactReportEntry{
				Name: "x", Kind: ArtifactKindDataSource, Status: status, Path: "p",
				Diagnostics: []Diagnostic{
					{Severity: SeverityError, Message: "e"},
					{Severity: SeverityWarning, Message: "w"},
					{Severity: SeverityInfo, Message: "i", Location: "spec:x"},
				},
			})
		}
		Expect(validateAgainstContract(schema, report)).To(Succeed())
	})

	It("accepts both skip reasons", func() {
		report := baseReport()
		report.SkippedOperations = []SkippedOperation{
			{OperationId: "A", Path: "/a", Method: "GET", Reason: SkipReasonTrackingFieldAbsent},
			{OperationId: "B", Path: "/b", Method: "GET", Reason: SkipReasonTrackingFieldSkip},
		}
		Expect(validateAgainstContract(schema, report)).To(Succeed())
	})

	Describe("cassette fields", func() {
		cassetteResult := func() CassetteResult {
			return CassetteResult{
				Name:            "integration_twilio_account",
				Kind:            ArtifactKindResource,
				TestName:        "TestAccDatadogIntegrationTwilioAccountOpenAPIExample",
				Status:          CassetteStatusGenerated,
				WriteAction:     CassetteWriteCreated,
				TestPath:        "datadog/tests/resource_x_openapi_example_test.go",
				CassettePath:    "datadog/tests/cassettes/TestAccX.yaml",
				FreezePath:      "datadog/tests/cassettes/TestAccX.freeze",
				SelectedExample: ScenarioNameDefault,
				Interactions: []InteractionSummary{{
					Index: 0, Role: string(InteractionRoleCreate), OperationId: "CreateX",
					Method: "POST", URL: "https://api.datadoghq.com/api/v2/x", Status: 201,
					SourcePaths: []string{"spec:/api/v2/x.post.requestBody.content.application/json.examples.default"},
				}},
				Diagnostics: []Diagnostic{{Severity: SeverityInfo, Message: "selected example \"default\""}},
			}
		}

		It("accepts a fully populated cassette result and summary", func() {
			report := baseReport()
			report.Cassettes = []CassetteResult{cassetteResult()}
			report.CassetteSummary = &CassetteSummary{Generated: 1}
			Expect(validateAgainstContract(schema, report)).To(Succeed())
		})

		It("accepts every status, write action and interaction role the model defines", func() {
			report := baseReport()
			for _, status := range []CassetteStatus{
				CassetteStatusGenerated, CassetteStatusPreserved, CassetteStatusIneligible,
			} {
				for _, action := range []CassetteWriteAction{
					CassetteWriteNone, CassetteWriteCreated, CassetteWriteUnchanged,
					CassetteWriteReplacedGenerated, CassetteWriteReplacedRecorded,
				} {
					result := cassetteResult()
					result.Status = status
					result.WriteAction = action
					report.Cassettes = append(report.Cassettes, result)
				}
			}
			for i, role := range []InteractionRole{
				InteractionRoleCreate, InteractionRoleRead, InteractionRoleSearch,
				InteractionRoleUpdate, InteractionRoleDelete, InteractionRoleRefresh,
				InteractionRoleDestroyVerification,
			} {
				result := cassetteResult()
				result.Interactions = []InteractionSummary{{
					Index: i, Role: string(role), OperationId: "X",
					Method: "GET", URL: "https://api.datadoghq.com/api/v2/x", Status: 200,
				}}
				report.Cassettes = append(report.Cassettes, result)
			}
			report.CassetteSummary = &CassetteSummary{Generated: 1, Preserved: 2, Ineligible: 3}
			Expect(validateAgainstContract(schema, report)).To(Succeed())
		})

		DescribeTable("rejects a value the model cannot produce",
			func(mutate func(*RunReport), wantMessage string) {
				report := baseReport()
				report.Cassettes = []CassetteResult{cassetteResult()}
				mutate(report)
				err := validateAgainstContract(schema, report)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(wantMessage))
			},
			Entry("an unknown status",
				func(r *RunReport) { r.Cassettes[0].Status = "half_written" }, "status"),
			Entry("an unknown write action",
				func(r *RunReport) { r.Cassettes[0].WriteAction = "clobbered" }, "write_action"),
			Entry("an unknown interaction role",
				func(r *RunReport) { r.Cassettes[0].Interactions[0].Role = "guessing" }, "role"),
			Entry("a negative interaction index",
				func(r *RunReport) { r.Cassettes[0].Interactions[0].Index = -1 }, "index"),
			Entry("a status code outside the HTTP range",
				func(r *RunReport) { r.Cassettes[0].Interactions[0].Status = 42 }, "status"),
			Entry("a negative summary count",
				func(r *RunReport) { r.CassetteSummary = &CassetteSummary{Generated: -1} }, "generated"),
		)
	})
})
