package emit

import (
	"flag"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// updateGolden rewrites the testdata/*.golden files from the current template
// output. Run `go test ./internal/emit/... -update` after an intentional
// template change, then review the diff.
var updateGolden = flag.Bool("update", false, "update golden files")

var _ = Describe("data-source templates", func() {
	// Golden coverage: the views mirror the hand-written
	// datadog/fwprovider/data_source_datadog_{incident_type,teams}.go so each
	// golden is reviewable against a known artifact.
	DescribeTable("render to gofmt-canonical golden output",
		func(golden string, fixture func() DataSourceView) {
			got, err := RenderDataSource(fixture())
			Expect(err).NotTo(HaveOccurred())
			matchGolden(golden, got)
		},
		Entry("singular (incident_type)", "data_source_singular.golden", incidentTypeView),
		Entry("plural (teams)", "data_source_plural.golden", pluralFixture),
	)

	It("renders deterministically across runs", func() {
		for _, v := range []DataSourceView{incidentTypeView(), pluralFixture()} {
			first, err := RenderDataSource(v)
			Expect(err).NotTo(HaveOccurred())
			second, err := RenderDataSource(v)
			Expect(err).NotTo(HaveOccurred())
			Expect(first).To(Equal(second), "render of %q is non-deterministic", v.TypeName)
		}
	})
})

// matchGolden compares got against testdata/<name>, or rewrites it under -update.
func matchGolden(name string, got []byte) {
	GinkgoHelper()
	path := filepath.Join("../testdata/emit", name)

	if *updateGolden {
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, got, 0o644)).To(Succeed())
		return
	}

	want, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred(), "reading golden %s (run with -update to create it)", path)
	Expect(string(got)).To(Equal(string(want)), "rendered output does not match %s", path)
}

// incidentTypeView is the incident_type data source built end-to-end through the
// emit builder, so the golden proves BuildArtifact + BuildDataSourceView rather
// than a hand-written view. The shared incident_type fixture lives in builder_test.go.
func incidentTypeView() DataSourceView {
	GinkgoHelper()
	view, err := BuildDataSourceView(incidentTypeArtifact())
	Expect(err).NotTo(HaveOccurred())
	return view
}

// pluralFixture is the teams data source as a view.
func pluralFixture() DataSourceView {
	return DataSourceView{
		Cardinality: Plural,
		TypeName:    "teams",
		GoName:      "teams",
		Description: "Use this data source to retrieve information about existing teams for use in other resources.",
		SDKPackage:  "datadogV2",
		APIStruct:   "TeamsApi",
		APIAccessor: "GetTeamsApiV2",
		Read: SDKReadView{
			Method:             "ListTeams",
			Paginated:          true,
			ItemType:           "Team",
			OptionalParamsType: "ListTeamsOptionalParameters",
			Filters: []FilterParamView{
				{StateField: "FilterKeyword", ParamField: "FilterKeyword", ValueExpr: "ValueStringPointer()"},
				{StateField: "FilterMe", ParamField: "FilterMe", ValueExpr: "ValueBoolPointer()"},
			},
		},
		Models: []ModelStructView{
			{
				Name: "teamsDataSourceModel",
				Fields: []ModelFieldView{
					{Comment: "Query Parameters", GoField: "FilterKeyword", GoType: "types.String", TFName: "filter_keyword"},
					{GoField: "FilterMe", GoType: "types.Bool", TFName: "filter_me"},
					{Comment: "Results", GoField: "ID", GoType: "types.String", TFName: "id"},
					{GoField: "Teams", GoType: "[]*TeamModel", TFName: "teams"},
				},
			},
			{
				Name: "TeamModel",
				Fields: []ModelFieldView{
					{GoField: "Description", GoType: "types.String", TFName: "description"},
					{GoField: "Handle", GoType: "types.String", TFName: "handle"},
					{GoField: "ID", GoType: "types.String", TFName: "id"},
					{GoField: "LinkCount", GoType: "types.Int64", TFName: "link_count"},
					{GoField: "Name", GoType: "types.String", TFName: "name"},
					{GoField: "Summary", GoType: "types.String", TFName: "summary"},
					{GoField: "UserCount", GoType: "types.Int64", TFName: "user_count"},
				},
			},
		},
		Schema: SchemaView{
			Attributes: []AttrView{
				{TFName: "filter_keyword", TFType: "schema.StringAttribute", Description: "Search query. Can be team name, team handle, or email of team member.", Optional: true},
				{TFName: "filter_me", TFType: "schema.BoolAttribute", Description: "When true, only returns teams the current user belongs to.", Optional: true},
			},
			Blocks: []AttrView{
				{
					TFName:      "teams",
					Description: "List of teams",
					IsBlock:     true,
					ListBlock:   true,
					Attributes: []AttrView{
						{TFName: "description", TFType: "schema.StringAttribute", Description: "Free-form markdown description/content for the team's homepage.", Computed: true},
						{TFName: "handle", TFType: "schema.StringAttribute", Description: "The team's handle.", Computed: true},
						{TFName: "id", TFType: "schema.StringAttribute", Description: "The team's identifier.", Computed: true},
						{TFName: "link_count", TFType: "schema.Int64Attribute", Description: "The number of links belonging to the team.", Computed: true},
						{TFName: "name", TFType: "schema.StringAttribute", Description: "The name of the team.", Computed: true},
						{TFName: "summary", TFType: "schema.StringAttribute", Description: "A brief summary of the team, derived from the `description`.", Computed: true},
						{TFName: "user_count", TFType: "schema.Int64Attribute", Description: "The number of users belonging to the team.", Computed: true},
					},
				},
			},
		},
		State: StateView{
			ItemStruct: "TeamModel",
			ItemField:  "Teams",
			ItemFields: []StateAssignment{
				{LHS: "Description", RHS: "types.StringValue(item.Attributes.GetDescription())"},
				{LHS: "Handle", RHS: "types.StringValue(item.Attributes.GetHandle())"},
				{LHS: "ID", RHS: "types.StringValue(item.GetId())"},
				{LHS: "LinkCount", RHS: "types.Int64Value(int64(item.Attributes.GetLinkCount()))"},
				{LHS: "Name", RHS: "types.StringValue(item.Attributes.GetName())"},
				{LHS: "Summary", RHS: "types.StringValue(item.Attributes.GetSummary())"},
				{LHS: "UserCount", RHS: "types.Int64Value(int64(item.Attributes.GetUserCount()))"},
			},
		},
	}
}
