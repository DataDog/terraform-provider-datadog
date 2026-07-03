package emit

import (
	"flag"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
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
		Entry("singular list-of-string (team)", "data_source_singular_list.golden", teamSingularView),
		Entry("singular nested object arrays (cost_budget)", "data_source_singular_nested.golden", costBudgetView),
		Entry("singular search (powerpack)", "data_source_singular_search.golden", powerpackSearchView),
		Entry("singular both (datastore)", "data_source_singular_both.golden", datastoreBothView),
		Entry("plural (teams)", "data_source_plural.golden", pluralFixture),
		Entry("plural nested object arrays (widgets)", "data_source_plural_nested.golden", pluralNestedView),
		Entry("plural no-params (datastores)", "data_source_plural_no_params.golden", datastoresView),
		Entry("singular nested object (apm_retention_filter)", "data_source_singular_object.golden", retentionFilterView),
		Entry("plural nested object (gizmos)", "data_source_plural_object.golden", pluralObjectView),
	)

	It("renders deterministically across runs", func() {
		for _, v := range []DataSourceView{incidentTypeView(), teamSingularView(), costBudgetView(), powerpackSearchView(), datastoreBothView(), pluralFixture(), pluralNestedView(), datastoresView(), retentionFilterView(), pluralObjectView()} {
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

// teamSingularView is the team singular data source built end-to-end through the
// emit builder; its golden proves collection-of-primitive hoisting — two string
// arrays rendered as schema.ListAttribute and mapped via types.ListValueFrom. The
// shared team fixture lives in builder_test.go.
func teamSingularView() DataSourceView {
	GinkgoHelper()
	return mustView(teamSingularOperation())
}

// costBudgetView is the cost_budget singular data source built end-to-end through
// the emit builder; its golden proves recursive array-of-object hoisting — entries
// rendered as a schema.ListNestedBlock holding a nested tag_filters ListNestedBlock,
// mapped through nested guarded loops. The shared fixture lives in builder_test.go.
func costBudgetView() DataSourceView {
	GinkgoHelper()
	return mustView(costBudgetOperation())
}

// pluralNestedView is the synthetic widgets plural data source built end-to-end;
// its golden proves an object array inside a list item renders as a nested
// ListNestedBlock and maps through a per-element loop after the item literal. The
// shared fixture lives in builder_test.go.
func pluralNestedView() DataSourceView {
	GinkgoHelper()
	return mustView(pluralNestedOperation())
}

// retentionFilterView is the apm retention filter singular data source built
// end-to-end; its golden proves a bare object under attributes renders as a
// schema.SingleNestedBlock with a guarded, recursive object_single state mapping.
func retentionFilterView() DataSourceView {
	GinkgoHelper()
	return mustView(retentionFilterOperation())
}

// pluralObjectView is the synthetic gizmos plural data source built end-to-end; its
// golden proves a bare object inside a list item renders as a SingleNestedBlock and
// maps via an object_single ItemList after the item literal.
func pluralObjectView() DataSourceView {
	GinkgoHelper()
	return mustView(pluralObjectOperation())
}

// datastoresView is the datastores data source built end-to-end through the
// emit builder; its golden proves the no-optional-params, non-paginated,
// zero-filter render path. The shared datastores fixture lives in builder_test.go.
func datastoresView() DataSourceView {
	GinkgoHelper()
	art, err := model.BuildArtifact(datastoresOperation())
	Expect(err).NotTo(HaveOccurred())
	view, err := BuildDataSourceView(art)
	Expect(err).NotTo(HaveOccurred())
	return view
}

// powerpackSearchView is the search-only powerpack data source built end-to-end;
// its golden proves the list→guard→pick render with a paginated, filtered search.
func powerpackSearchView() DataSourceView {
	GinkgoHelper()
	return mustView(powerpackSearchOperation())
}

// datastoreBothView is the id-optional datastore data source built end-to-end; its
// golden proves the by-id-else-search render and is reviewable against the
// hand-written data_source_datadog_datastore.go.
func datastoreBothView() DataSourceView {
	GinkgoHelper()
	return mustView(datastoreBothOperation())
}

// pluralFixture is the teams data source as a view.
func pluralFixture() DataSourceView {
	return DataSourceView{
		Cardinality: Plural,
		TypeName:    "teams",
		GoName:      "datadogTeams",
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
				Name: "datadogTeamsDataSourceModel",
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
					{GoField: "HiddenModules", GoType: "types.List", TFName: "hidden_modules"},
					{GoField: "VisibleModules", GoType: "types.List", TFName: "visible_modules"},
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
						{TFName: "hidden_modules", TFType: "schema.ListAttribute", ElementType: "types.StringType", Description: "Collection of hidden modules for the team.", Computed: true},
						{TFName: "visible_modules", TFType: "schema.ListAttribute", ElementType: "types.StringType", Description: "Collection of visible modules for the team.", Computed: true},
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
			ItemLists: []ListAssignment{
				{Kind: "primitive", LHS: "r.HiddenModules", GetterOk: "item.Attributes.GetHiddenModulesOk()", Var: "hiddenModules", ElementType: "types.StringType"},
				{Kind: "primitive", LHS: "r.VisibleModules", GetterOk: "item.Attributes.GetVisibleModulesOk()", Var: "visibleModules", ElementType: "types.StringType"},
			},
		},
	}
}
