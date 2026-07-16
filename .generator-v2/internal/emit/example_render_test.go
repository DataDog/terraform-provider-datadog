package emit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("data-source examples", func() {
	It("renders a singular by-ID lookup with the conventional example UUID", func() {
		got := RenderDataSourceExample(incidentTypeView())

		Expect(string(got.Content)).To(Equal(`data "datadog_incident_type" "example" {
  id = "11111111-2222-3333-4444-555555555555"
}
`))
		Expect(got.Diagnostics).To(BeEmpty())
	})

	It("renders deterministic scalar filters for a search-only singular lookup", func() {
		got := RenderDataSourceExample(powerpackSearchView())

		Expect(string(got.Content)).To(Equal(`data "datadog_powerpack" "example" {
  filter_name = "example"
}
`))
		Expect(got.Diagnostics).To(BeEmpty())
	})

	It("renders an unfiltered plural lookup", func() {
		got := RenderDataSourceExample(pluralFixture())

		Expect(string(got.Content)).To(Equal("data \"datadog_teams\" \"example\" {}\n"))
		Expect(got.Diagnostics).To(BeEmpty())
	})

	It("uses type-appropriate, terraform-formatted filter placeholders", func() {
		view := DataSourceView{
			Cardinality: Singular,
			TypeName:    "widgets",
			Searchable:  true,
			Schema: SchemaView{Attributes: []AttrView{
				{TFName: "filter_keyword", TFType: "schema.StringAttribute", Optional: true},
				{TFName: "filter_me", TFType: "schema.BoolAttribute", Optional: true},
				{TFName: "page_size", TFType: "schema.Int64Attribute", Optional: true},
				{TFName: "computed_value", TFType: "schema.StringAttribute", Computed: true},
			}},
		}

		got := RenderDataSourceExample(view)
		Expect(string(got.Content)).To(Equal(`data "datadog_widgets" "example" {
  filter_keyword = "example"
  filter_me      = false
  page_size      = 0
}
`))
		Expect(got.Diagnostics).To(BeEmpty())
	})

	It("renders required scalar inputs for plural shapes", func() {
		view := DataSourceView{
			Cardinality: Plural,
			TypeName:    "widgets",
			Schema: SchemaView{Attributes: []AttrView{
				{TFName: "account_id", TFType: "schema.StringAttribute", Required: true},
				{TFName: "optional_filter", TFType: "schema.StringAttribute", Optional: true},
			}},
		}

		got := RenderDataSourceExample(view)
		Expect(string(got.Content)).To(Equal(`data "datadog_widgets" "example" {
  account_id = "example"
}
`))
		Expect(got.Diagnostics).To(BeEmpty())
	})

	It("reports required inputs it cannot render", func() {
		view := DataSourceView{
			Cardinality: Plural,
			TypeName:    "widgets",
			Schema: SchemaView{
				Attributes: []AttrView{
					{TFName: "account_ids", TFType: "schema.ListAttribute", Required: true},
				},
				Blocks: []AttrView{
					{TFName: "scope", IsBlock: true, Required: true},
				},
			},
		}

		got := RenderDataSourceExample(view)
		Expect(string(got.Content)).To(Equal("data \"datadog_widgets\" \"example\" {}\n"))
		Expect(got.Diagnostics).To(ConsistOf(
			`generated example for "widgets" may be incomplete: required attribute "account_ids" has unsupported type "schema.ListAttribute"`,
			`generated example for "widgets" may be incomplete: required block "scope" cannot be rendered`,
		))
	})

	It("reports a singular shape with no supported lookup strategy", func() {
		got := RenderDataSourceExample(DataSourceView{Cardinality: Singular, TypeName: "widgets"})

		Expect(string(got.Content)).To(Equal("data \"datadog_widgets\" \"example\" {}\n"))
		Expect(got.Diagnostics).To(ConsistOf(
			`generated example for "widgets" may be incomplete: singular lookup has neither by-ID nor searchable resolution`,
		))
	})

	It("renders deterministically", func() {
		view := powerpackSearchView()
		Expect(RenderDataSourceExample(view)).To(Equal(RenderDataSourceExample(view)))
	})
})
