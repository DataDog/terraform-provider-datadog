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

// enumValues reads one $defs enum straight from the contract bytes. The
// compiled schema does not expose $defs enums, and the raw document is what a
// consumer reads anyway.
func enumValues(def, property string) []any {
	var doc struct {
		Defs map[string]struct {
			Properties map[string]struct {
				Enum []any `json:"enum"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	Expect(json.Unmarshal(contracts.RunReportSchema, &doc)).To(Succeed())
	d, ok := doc.Defs[def]
	Expect(ok).To(BeTrue(), "contract has no $defs/%s", def)
	prop, ok := d.Properties[property]
	Expect(ok).To(BeTrue(), "contract has no $defs/%s/properties/%s", def, property)
	return prop.Enum
}

// declaredProperties reads one $defs property set from the contract bytes.
func declaredProperties(def string) []string {
	var doc struct {
		Defs map[string]struct {
			Properties map[string]any `json:"properties"`
		} `json:"$defs"`
	}
	Expect(json.Unmarshal(contracts.RunReportSchema, &doc)).To(Succeed())
	d, ok := doc.Defs[def]
	Expect(ok).To(BeTrue(), "contract has no $defs/%s", def)
	names := make([]string, 0, len(d.Properties))
	for name := range d.Properties {
		names = append(names, name)
	}
	return names
}

// marshalledKeys is the JSON keys a value actually produces.
func marshalledKeys(value any) []string {
	encoded, err := json.Marshal(value)
	Expect(err).To(Succeed())
	var object map[string]any
	Expect(json.Unmarshal(encoded, &object)).To(Succeed())
	names := make([]string, 0, len(object))
	for name := range object {
		names = append(names, name)
	}
	return names
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

	Describe("diagnostic category", func() {
		// Same drift guard as the cassette enums: every category the model can
		// produce must validate, and the contract must declare no more than
		// the model produces.
		DescribeTable("accepts every category the model can produce",
			func(category DiagnosticCategory) {
				report := baseReport()
				report.Artifacts = []ArtifactReportEntry{{
					Name: "thing", Kind: ArtifactKindResource, Status: ArtifactStatusCreated,
					Diagnostics: []Diagnostic{{
						Severity: SeverityWarning, Message: "m", Category: category,
					}},
				}}
				Expect(validateAgainstContract(schema, report)).To(Succeed())
			},
			Entry("eligibility", DiagnosticCategoryEligibility),
			Entry("selection", DiagnosticCategorySelection),
			Entry("materialization", DiagnosticCategoryMaterialization),
			Entry("validation", DiagnosticCategoryValidation),
			Entry("render", DiagnosticCategoryRender),
			Entry("write", DiagnosticCategoryWrite),
		)

		It("declares no category the model cannot produce", func() {
			Expect(enumValues("diagnostic", "category")).To(HaveLen(6))
		})

		It("rejects a category the model cannot produce", func() {
			report := baseReport()
			report.Artifacts = []ArtifactReportEntry{{
				Name: "thing", Kind: ArtifactKindResource, Status: ArtifactStatusCreated,
				Diagnostics: []Diagnostic{{
					Severity: SeverityWarning, Message: "m", Category: "invented",
				}},
			}}
			err := validateAgainstContract(schema, report)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category"))
		})

		// A diagnostic without a category is the artifact pipeline's shape and
		// must stay valid, which is why the field is omitempty.
		It("accepts a diagnostic carrying no category", func() {
			report := baseReport()
			report.Artifacts = []ArtifactReportEntry{{
				Name: "thing", Kind: ArtifactKindResource, Status: ArtifactStatusCreated,
				Diagnostics: []Diagnostic{{Severity: SeverityWarning, Message: "m"}},
			}}
			Expect(validateAgainstContract(schema, report)).To(Succeed())
		})
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
				SelectedExample: ScenarioNameDefault,
				Diagnostics:     []Diagnostic{{Severity: SeverityInfo, Message: "selected example \"default\""}},
			}
		}

		It("accepts a fully populated cassette result and summary", func() {
			report := baseReport()
			report.Cassettes = []CassetteResult{cassetteResult()}
			report.CassetteSummary = &CassetteSummary{Generated: 1}
			Expect(validateAgainstContract(schema, report)).To(Succeed())
		})

		// Every status and write action the model can produce must validate.
		// The previous table only covered the ones already in the enum, so
		// CassetteWriteUpdated — emitted whenever a run rewrites an existing
		// generated test — drifted out of the contract unnoticed.
		DescribeTable("accepts every write action the model can produce",
			func(action CassetteWriteAction) {
				report := baseReport()
				result := cassetteResult()
				result.WriteAction = action
				report.Cassettes = []CassetteResult{result}
				report.CassetteSummary = &CassetteSummary{Generated: 1}
				Expect(validateAgainstContract(schema, report)).To(Succeed())
			},
			Entry("none", CassetteWriteNone),
			Entry("created", CassetteWriteCreated),
			Entry("unchanged", CassetteWriteUnchanged),
			Entry("updated", CassetteWriteUpdated),
		)

		DescribeTable("accepts every status the model can produce",
			func(status CassetteStatus) {
				report := baseReport()
				result := cassetteResult()
				result.Status = status
				report.Cassettes = []CassetteResult{result}
				report.CassetteSummary = &CassetteSummary{Generated: 1}
				Expect(validateAgainstContract(schema, report)).To(Succeed())
			},
			Entry("generated", CassetteStatusGenerated),
			Entry("skipped", CassetteStatusSkipped),
			Entry("ineligible", CassetteStatusIneligible),
		)

		// Guards the other direction: a value added to the schema but never
		// produced, which the per-value tables above cannot see. Counting is
		// the only way to catch an enum that has grown past the model.
		// additionalProperties:false already fails a Go field missing from the
		// contract. This is the other direction — a contract property the model
		// never populates, which validation cannot see because absent optional
		// fields are valid. cassette_path and freeze_path sat here unpopulated.
		It("declares exactly the summary counters the model produces", func() {
			Expect(declaredProperties("cassetteSummary")).To(
				ConsistOf(marshalledKeys(CassetteSummary{})))
		})

		It("declares exactly the properties a populated result produces", func() {
			full := CassetteResult{
				Name: "n", Kind: ArtifactKindResource, TestName: "t",
				Status: CassetteStatusGenerated, WriteAction: CassetteWriteCreated,
				TestPath: "p", SelectedExample: ScenarioNameDefault,
				Diagnostics: []Diagnostic{{Severity: SeverityInfo, Message: "m"}},
			}
			Expect(declaredProperties("cassetteResult")).To(ConsistOf(marshalledKeys(full)))
		})

		It("declares no write action or status the model cannot produce", func() {
			Expect(enumValues("cassetteResult", "write_action")).To(HaveLen(4))
			Expect(enumValues("cassetteResult", "status")).To(HaveLen(3))
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
			Entry("a negative summary count",
				func(r *RunReport) { r.CassetteSummary = &CassetteSummary{Generated: -1} }, "generated"),
		)
	})
})
