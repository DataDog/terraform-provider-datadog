package emit

import (
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/cassette"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/parser"
)

var exampleFrozen = time.Date(2026, 6, 25, 8, 30, 50, 0, time.UTC)

// twilioExampleView drives the whole pipeline — parse, select, materialize,
// scenario, view — so these specs assert on what the real chain produces
// rather than on a hand-built scenario.
func twilioExampleView() (exampleTestView, *model.GeneratedTestScenario) {
	spec, err := parser.LoadSpec(filepath.Join(
		"..", "testdata", "mini-oas", "mini-datadog_integration_twilio_account.yaml"))
	Expect(err).To(Succeed())

	var read *model.Operation
	for _, op := range spec.Operations {
		if op.Tracking != nil && op.Tracking.ArtifactName == "integration_twilio_account" {
			read = op
		}
	}
	Expect(read).NotTo(BeNil())

	ops := []*model.Operation{read}
	for _, target := range read.ResolvedGroup.Operations() {
		if target != read {
			ops = append(ops, target)
		}
	}
	selection, err := cassette.Select(ops)
	Expect(err).To(Succeed())

	scenario, err := cassette.BuildResourceScenario(cassette.ResourceTarget{
		ArtifactName: "integration_twilio_account",
		ServerURL:    spec.ServerURL,
		IdStrategy:   model.IdStrategyDataID,
		FreezeTime:   exampleFrozen,
		Create:       read.ResolvedGroup.Create,
		Read:         read.ResolvedGroup.Read,
		Update:       read.ResolvedGroup.Update,
		Delete:       read.ResolvedGroup.Delete,
		Selection:    selection,
	})
	Expect(err).To(Succeed())

	// The scenario comes from the real pipeline, so the materialized paths are
	// genuine. The schema is hand-built because BuildResourceView needs
	// internal/sdkbind to have resolved SDK wrapper types first — the CLI does
	// that, a unit test does not, and this unit is the mapping from scenario
	// plus schema to HCL and checks, not the SDK binding.
	view, err := BuildExampleTestView(scenario, twilioSchemaView(), twilioAPIPaths())
	Expect(err).To(Succeed())
	return view, scenario
}

// twilioSchemaView mirrors the attribute tree tfgen actually emits: leaves in
// Attributes, nested objects in Blocks. That split is the point — a renderer
// walking only Attributes omits every block, leaving a configuration without
// its required arguments.
func twilioSchemaView() SchemaView {
	leaf := func(name string, required bool) AttrView {
		return AttrView{TFName: name, TFType: "schema.StringAttribute",
			Required: required, Optional: !required}
	}
	boolLeaf := func(name string) AttrView {
		return AttrView{TFName: name, TFType: "schema.BoolAttribute", Optional: true}
	}
	return SchemaView{
		Attributes: []AttrView{
			leaf("name", true),
			{TFName: "id", TFType: "schema.StringAttribute", Computed: true},
		},
		Blocks: []AttrView{
			{TFName: "authentication", IsBlock: true, Optional: true,
				Attributes: []AttrView{
					leaf("auth_type", true),
					leaf("username", false),
					{TFName: "password", TFType: "schema.StringAttribute",
						Optional: true, Sensitive: true},
				}},
			{TFName: "settings", IsBlock: true, Optional: true,
				Attributes: []AttrView{leaf("account_sid", true), boolLeaf("censor_logs")}},
			{TFName: "dataflows", IsBlock: true, Optional: true,
				Blocks: []AttrView{
					{TFName: "twilio_messages_logs", IsBlock: true, Optional: true,
						Attributes: []AttrView{boolLeaf("enabled")}},
				}},
		},
	}
}

// twilioAPIPaths is the correspondence APIPathIndex records. The interesting
// entries are the ones a suffix match could never bridge — twilio_messages_logs
// from twilio-messages-logs — and the oneOf wrapper, which advances the
// Terraform path but not the API one.
func twilioAPIPaths() map[string]string {
	const attrs = "data.attributes"
	return map[string]string{
		"name":                     attrs + ".name",
		"id":                       "data.id",
		"authentication.auth_type": attrs + ".authentication.auth_type",
		"authentication.username":  attrs + ".authentication.username",
		"authentication.password":  attrs + ".authentication.password",
		"settings.account_sid":     attrs + ".settings.account_sid",
		"settings.censor_logs":     attrs + ".settings.censor_logs",
		"dataflows.twilio_messages_logs.enabled": attrs +
			".dataflows.twilio-messages-logs.enabled",
	}
}

var _ = Describe("BuildExampleTestView", func() {
	Describe("identity and header", func() {
		It("names the test, cassette and freeze companion consistently", func() {
			view, scenario := twilioExampleView()
			Expect(view.FuncName).To(Equal("TestAccDatadogIntegrationTwilioAccountOpenAPIExample"))
			// The harness derives the cassette from t.Name().
			Expect(view.CassettePath).To(Equal("cassettes/" + scenario.TestFuncName + ".yaml"))
			Expect(view.FreezePath).To(Equal("cassettes/" + scenario.TestFuncName + ".freeze"))
			Expect(view.ResourceType).To(Equal("datadog_integration_twilio_account"))
		})

		// The writer replaces a file carrying the marker and never one without,
		// so the header is what separates regeneration from clobbering a
		// hand-recorded fixture.
		It("carries the ownership marker", func() {
			view, _ := twilioExampleView()
			Expect(view.Marker).To(Equal(model.GeneratedMarker))
			Expect(view.Marker).NotTo(BeEmpty())
		})

		It("reports the interaction count, since a replay failure is usually a mismatch", func() {
			view, scenario := twilioExampleView()
			Expect(view.InteractionCount).To(Equal(len(scenario.Interactions)))
			Expect(view.InteractionCount).To(Equal(7))
		})

		It("derives an unexported config helper", func() {
			view, _ := twilioExampleView()
			Expect(view.ConfigFunc).To(HavePrefix("test"))
			Expect(view.ConfigFunc).NotTo(HavePrefix("Test"))
			Expect(view.ConfigFunc).To(HaveSuffix("Config"))
		})
	})

	Describe("configuration rendering", func() {
		It("renders only attributes the generated schema declares", func() {
			view, _ := twilioExampleView()
			config := view.Steps[0].ConfigBody

			Expect(config).To(HavePrefix(`resource "datadog_integration_twilio_account" "foo" {`))
			Expect(config).To(ContainSubstring(`name = "twilio-prod"`))
			Expect(config).To(ContainSubstring("settings = {"))
			Expect(config).To(ContainSubstring(`account_sid = "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`))
			Expect(config).To(ContainSubstring("censor_logs = true"))
			Expect(config).To(HaveSuffix("}"))

			// The JSON:API envelope is flattened away by the generated schema,
			// so it must not appear in the configuration.
			Expect(config).NotTo(ContainSubstring("data ="))
			Expect(config).NotTo(ContainSubstring("attributes ="))
		})

		// A declared secret must not reach a committed test file.
		It("writes the redacted placeholder, never the declared secret", func() {
			view, _ := twilioExampleView()
			for _, step := range view.Steps {
				Expect(step.ConfigBody).NotTo(ContainSubstring("twilio-basic-auth-secret"))
			}
			Expect(view.Steps[0].ConfigBody).To(ContainSubstring(model.RedactedPlaceholder))
		})

		It("gives each step its own helper when there is more than one", func() {
			view, _ := twilioExampleView()
			Expect(view.Steps).To(HaveLen(2))
			Expect(view.Steps[0].ConfigFunc).To(HaveSuffix("Step1"))
			Expect(view.Steps[1].ConfigFunc).To(HaveSuffix("Step2"))
		})

		// The second step must differ, or it asserts nothing.
		It("renders a distinct configuration for the update step", func() {
			view, _ := twilioExampleView()
			Expect(view.Steps[1].ConfigBody).NotTo(Equal(view.Steps[0].ConfigBody))
			Expect(view.Steps[1].ConfigBody).To(ContainSubstring(`name = "twilio-prod-renamed"`))
			Expect(view.Steps[1].ConfigBody).To(ContainSubstring("censor_logs = false"))
		})

		It("produces identical output on repeated builds", func() {
			first, _ := twilioExampleView()
			second, _ := twilioExampleView()
			Expect(first.Steps[0].ConfigBody).To(Equal(second.Steps[0].ConfigBody))
			Expect(first.Steps[0].Checks).To(Equal(second.Steps[0].Checks))
		})
	})

	Describe("check rendering", func() {
		// Checks read state only: the cassette is the oracle, so a live
		// existence check would add an interaction the scenario never planned.
		It("asserts state and never calls the API", func() {
			view, _ := twilioExampleView()
			checks := strings.Join(view.Steps[0].Checks, "\n")

			Expect(checks).To(ContainSubstring(
				`resource.TestCheckResourceAttrSet("datadog_integration_twilio_account.foo", "id")`))
			Expect(checks).To(ContainSubstring(`"name", "twilio-prod"`))
			Expect(checks).NotTo(ContainSubstring("Exists"))
			Expect(checks).NotTo(ContainSubstring("providers."))
		})

		It("renders a boolean the way Terraform stores it", func() {
			view, _ := twilioExampleView()
			Expect(strings.Join(view.Steps[0].Checks, "\n")).
				To(ContainSubstring(`"settings.censor_logs", "true"`))
		})

		It("orders checks stably", func() {
			view, _ := twilioExampleView()
			checks := view.Steps[0].Checks
			// The id check leads; the rest are lexical by attribute path.
			Expect(checks[0]).To(ContainSubstring("TestCheckResourceAttrSet"))
			var paths []string
			for _, check := range checks[1:] {
				paths = append(paths, check)
			}
			sorted := append([]string(nil), paths...)
			for i := 1; i < len(sorted); i++ {
				for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
					sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
				}
			}
			Expect(paths).To(Equal(sorted))
		})

		It("asserts the placeholder for a replaced secret, which is what state holds", func() {
			view, _ := twilioExampleView()
			checks := strings.Join(view.Steps[0].Checks, "\n")
			Expect(checks).NotTo(ContainSubstring("twilio-basic-auth-secret"))
		})
	})

	It("rejects a scenario that does not validate", func() {
		_, err := BuildExampleTestView(&model.GeneratedTestScenario{}, SchemaView{}, twilioAPIPaths())
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("hclLiteral", func() {
	DescribeTable("renders each value form materialization can produce",
		func(in any, want string) { Expect(hclLiteral(in)).To(Equal(want)) },
		Entry("string", "widget", `"widget"`),
		Entry("string needing escapes", `a"b`, `"a\"b"`),
		Entry("bool", true, "true"),
		Entry("int", 3, "3"),
		Entry("int64", int64(4), "4"),
		Entry("float", 1.5, "1.5"),
		Entry("null", nil, "null"),
		Entry("list", []any{"a", 1, true}, `["a", 1, true]`),
		Entry("nested list", []any{[]any{"a"}}, `[["a"]]`),
		Entry("unexpected type is quoted, not guessed", map[string]any{}, `"map[]"`),
	)
})

var _ = Describe("checkValue", func() {
	DescribeTable("renders values the way Terraform stores them, as strings",
		func(in any, want string) { Expect(checkValue(in)).To(Equal(want)) },
		Entry("string", "widget", "widget"),
		Entry("bool", false, "false"),
		Entry("int", 7, "7"),
		Entry("int64", int64(8), "8"),
		Entry("float", 2.25, "2.25"),
		Entry("other", []any{"a"}, "[a]"),
	)
})

var _ = Describe("lookupValue", func() {
	index := map[string]string{
		"background_color":                       "data.attributes.backgroundColor",
		"dataflows.twilio_messages_logs.enabled": "data.attributes.dataflows.twilio-messages-logs.enabled",
		"name":                                   "data.attributes.name",
		"flat":                                   "flat",
		"nested.name":                            "data.attributes.nested.name",
	}

	// Regression test for the defect this replaced. Matching a Terraform path
	// against an API path by string suffix cannot bridge a name SnakeCase
	// normalized: background_color never suffix-matches backgroundColor, and
	// twilio_messages_logs never matches twilio-messages-logs. The value was
	// silently dropped from the configuration while the cassette still sent it.
	DescribeTable("resolves through the recorded correspondence",
		func(tfPath string, want any) {
			values := map[string]model.MaterializedValue{
				index[tfPath]: {Path: index[tfPath], Value: want},
			}
			got, ok := lookupValue(values, index, strings.Split(tfPath, "."))
			Expect(ok).To(BeTrue())
			Expect(got.Value).To(Equal(want))
		},
		Entry("camelCase", "background_color", "#ffffff"),
		Entry("hyphenated", "dataflows.twilio_messages_logs.enabled", true),
		Entry("already snake_case", "name", "widget"),
		Entry("no envelope at all", "flat", "value"),
	)

	// The old suffix match preferred the shallowest candidate, so a top-level
	// attribute could silently take a nested property's value.
	It("does not resolve a top-level attribute to a nested property of the same name", func() {
		values := map[string]model.MaterializedValue{
			"data.attributes.nested.name": {Path: "data.attributes.nested.name", Value: "INNER"},
		}
		_, ok := lookupValue(values, index, []string{"name"})
		Expect(ok).To(BeFalse())

		got, ok := lookupValue(values, index, []string{"nested", "name"})
		Expect(ok).To(BeTrue())
		Expect(got.Value).To(Equal("INNER"))
	})

	It("reports a Terraform path the index does not know", func() {
		_, ok := lookupValue(map[string]model.MaterializedValue{}, index, []string{"absent"})
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("testHelperName", func() {
	It("lower-cases the leading Test so the helper stays unexported", func() {
		Expect(testHelperName("TestAccDatadogWidgetOpenAPIExample", "Config")).
			To(Equal("testAccDatadogWidgetOpenAPIExampleConfig"))
	})

	// A name off-convention is prefixed rather than sliced, so this cannot panic.
	It("prefixes a name that does not follow the convention", func() {
		Expect(testHelperName("Odd", "Config")).To(Equal("testAccOddConfig"))
	})
})

// edgePaths is the correspondence for the attributes the edge specs declare.
// Each sits directly under the JSON:API attributes member, except the id,
// which the envelope carries itself.
func edgePaths() map[string]string {
	const attrs = "data.attributes"
	paths := map[string]string{"id": "data.id"}
	for _, name := range []string{
		"name", "created_at", "mode", "tags", "note",
		"settings", "settings.unset", "deep", "deep.inner", "deep.inner.also_unset",
	} {
		paths[name] = attrs + "." + name
	}
	return paths
}

var _ = Describe("BuildExampleTestView edge paths", func() {
	scenario := func(values []model.MaterializedValue) *model.GeneratedTestScenario {
		return &model.GeneratedTestScenario{
			ArtifactName: "widget", ArtifactKind: model.ArtifactKindResource,
			TestFuncName:     "TestAccDatadogWidgetOpenAPIExample",
			CassetteBaseName: "TestAccDatadogWidgetOpenAPIExample",
			TerraformAddress: "datadog_widget.foo",
			FreezeTime:       exampleFrozen,
			Steps: []model.ScenarioStep{{
				State: &model.MaterializedConfiguration{RequestValues: values},
			}},
			Interactions: []model.ScenarioInteraction{{
				Index: 0, Role: model.InteractionRoleCreate,
				Request: model.InteractionRequest{Method: "POST", URL: "https://api.datadoghq.com/x"},
			}},
		}
	}

	It("tolerates a step carrying no materialized state", func() {
		s := scenario(nil)
		s.Steps[0].State = nil
		view, err := BuildExampleTestView(s, SchemaView{
			Attributes: []AttrView{{TFName: "name", TFType: "schema.StringAttribute", Optional: true}},
		}, twilioAPIPaths())
		Expect(err).To(Succeed())
		Expect(view.Steps[0].ConfigBody).To(Equal("resource \"datadog_widget\" \"foo\" {\n}"))
		// The id check stands alone when nothing was configured.
		Expect(view.Steps[0].Checks).To(HaveLen(1))
	})

	// An empty block is a claim the scenario never made.
	It("omits a nested block nothing inside it populated", func() {
		view, err := BuildExampleTestView(scenario([]model.MaterializedValue{
			{Path: "data.attributes.name", Value: "widget"},
		}), SchemaView{
			Attributes: []AttrView{
				{TFName: "name",
					TFType: "schema.StringAttribute", Required: true},
				{TFName: "settings", IsBlock: true, Optional: true, Attributes: []AttrView{
					{TFName: "unset", TFType: "schema.StringAttribute", Optional: true},
				}},
				{TFName: "deep", IsBlock: true, Optional: true, Attributes: []AttrView{
					{TFName: "inner", IsBlock: true, Optional: true, Attributes: []AttrView{
						{TFName: "alsoUnset", TFType: "schema.StringAttribute", Optional: true},
					}},
				}},
			},
		}, edgePaths())
		Expect(err).To(Succeed())
		Expect(view.Steps[0].ConfigBody).NotTo(ContainSubstring("settings"))
		Expect(view.Steps[0].ConfigBody).NotTo(ContainSubstring("deep"))
	})

	// A purely computed attribute is never configured, even when the response
	// example happens to carry a value for it.
	It("never configures a purely computed attribute", func() {
		view, err := BuildExampleTestView(scenario([]model.MaterializedValue{
			{Path: "data.attributes.created_at", Value: "2026-06-25T08:30:50Z"},
		}), SchemaView{
			Attributes: []AttrView{
				{TFName: "created_at",
					TFType: "schema.StringAttribute", Computed: true},
			},
		}, edgePaths())
		Expect(err).To(Succeed())
		Expect(view.Steps[0].ConfigBody).NotTo(ContainSubstring("created_at"))
	})

	It("configures an optional-and-computed attribute, which the author may set", func() {
		view, err := BuildExampleTestView(scenario([]model.MaterializedValue{
			{Path: "data.attributes.mode", Value: "fast"},
		}), SchemaView{
			Attributes: []AttrView{
				{TFName: "mode",
					TFType: "schema.StringAttribute", Optional: true, Computed: true},
			},
		}, edgePaths())
		Expect(err).To(Succeed())
		Expect(view.Steps[0].ConfigBody).To(ContainSubstring(`mode = "fast"`))
	})

	// Asserting a list element by index pins an ordering the description does
	// not promise, and a null has nothing meaningful to assert.
	It("configures a list and a null but does not check them", func() {
		view, err := BuildExampleTestView(scenario([]model.MaterializedValue{
			{Path: "data.attributes.tags", Value: []any{"alpha", "beta"}},
			{Path: "data.attributes.note", Value: nil},
			{Path: "data.attributes.name", Value: "widget"},
		}), SchemaView{
			Attributes: []AttrView{
				{TFName: "tags",
					TFType: "schema.ListAttribute", Optional: true},
				{TFName: "note",
					TFType: "schema.StringAttribute", Optional: true},
				{TFName: "name",
					TFType: "schema.StringAttribute", Required: true},
			},
		}, edgePaths())
		Expect(err).To(Succeed())
		Expect(view.Steps[0].ConfigBody).To(ContainSubstring(`tags = ["alpha", "beta"]`))
		Expect(view.Steps[0].ConfigBody).To(ContainSubstring("note = null"))

		checks := strings.Join(view.Steps[0].Checks, "\n")
		Expect(checks).To(ContainSubstring(`"name", "widget"`))
		Expect(checks).NotTo(ContainSubstring("tags"))
		Expect(checks).NotTo(ContainSubstring("note"))
	})

	It("uses one config helper when there is a single step", func() {
		view, err := BuildExampleTestView(scenario(nil), SchemaView{}, twilioAPIPaths())
		Expect(err).To(Succeed())
		Expect(view.Steps).To(HaveLen(1))
		Expect(view.Steps[0].ConfigFunc).NotTo(ContainSubstring("Step"))
	})
})
