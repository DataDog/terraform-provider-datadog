package emit

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
)

// testView is the render context for the acceptance-test template. Like
// DataSourceView it is fully derived in Go so the template carries only layout:
// buildTestView turns a rendered data source into the handful of facts the
// scaffold needs (names, the lookup shape, the filter attributes to seed).
type testView struct {
	// FuncName is the Go test function, e.g. "TestAccDatadogTeamsDatasource". It
	// doubles as the recorded cassette's base name (t.Name()).
	FuncName string
	// ConfigFunc is the config-helper function, e.g. "testAccDatasourceDatadogTeamsConfig".
	ConfigFunc string
	// ResourceType is the provider-prefixed Terraform type, e.g. "datadog_teams";
	// the data source is addressed as data.<ResourceType>.foo.
	ResourceType string
	// CassettePath is where the recorded fixture lands, surfaced in the file's
	// recording instructions.
	CassettePath string

	// Plural adds the list-count check and switches the config wording.
	Plural bool
	// LookupByID emits `id = ...` instead of filter assignments (singular by-id).
	LookupByID bool
	// UseUniq threads uniqueEntityName through the config so a recording produces
	// stable unique names. True when at least one string filter is seeded.
	UseUniq bool
	// Filters are the optional attributes seeded in the data-source block.
	Filters []testFilter
	// CollectionKey is the plural item block's Terraform name, e.g. "teams",
	// checked as "<CollectionKey>.#".
	CollectionKey string
	// ConfigBody is the HCL emitted between the config helper's backticks, built
	// in Go so its whitespace is exact (gofmt does not touch raw-string contents).
	ConfigBody string
}

// testFilter is one optional attribute seeded into the generated config block.
type testFilter struct {
	// TFName is the attribute key, e.g. "filter_keyword".
	TFName string
	// HCLValue is the placeholder written for it: "%[1]q" for a string (bound to
	// uniq via fmt.Sprintf), or a literal "false"/"0" for bool/number.
	HCLValue string
}

// buildTestView derives the acceptance-test scaffold context from a data source.
func buildTestView(v DataSourceView) testView {
	tv := testView{
		FuncName:     "TestAcc" + upperFirst(v.GoName) + "Datasource",
		ConfigFunc:   "testAccDatasource" + upperFirst(v.GoName) + "Config",
		ResourceType: "datadog_" + v.TypeName,
		Plural:       v.Cardinality == Plural,
		// A singular by-id data source resolves on the id; everything else
		// (singular search/both, plural) resolves on optional filters.
		LookupByID: v.Cardinality != Plural && !v.Searchable,
	}
	tv.CassettePath = "datadog/tests/cassettes/" + tv.FuncName + ".yaml"

	for _, a := range v.Schema.Attributes {
		if !a.Optional {
			continue
		}
		value, ok := hclFilterValue(a.TFType)
		if !ok {
			continue // skip non-scalar filters from the scaffold
		}
		tv.Filters = append(tv.Filters, testFilter{TFName: a.TFName, HCLValue: value})
		if value == "%[1]q" {
			tv.UseUniq = true
		}
	}

	if tv.Plural {
		for _, b := range v.Schema.Blocks {
			if b.IsBlock && b.ListBlock {
				tv.CollectionKey = b.TFName
				break
			}
		}
	}

	tv.ConfigBody = buildConfigBody(tv)
	return tv
}

// buildConfigBody assembles the HCL written between the config helper's
// backticks: a TODO header pointing at the seed step, then the data-source block
// with either an id lookup (singular by-id) or the optional filter attributes.
func buildConfigBody(tv testView) string {
	var b strings.Builder
	b.WriteByte('\n')

	if tv.Plural {
		b.WriteString("# TODO(tfgen): create the resource(s) this data source reads")
	} else {
		b.WriteString("# TODO(tfgen): create the resource this data source reads")
	}
	if tv.UseUniq {
		b.WriteString(" (use %[1]q for unique names)")
	}
	if tv.Plural {
		b.WriteString(", then add a depends_on so the read observes them.\n")
	} else {
		b.WriteString(", then reference it below.\n")
	}

	fmt.Fprintf(&b, "data %q \"foo\" {\n", tv.ResourceType)
	switch {
	case tv.LookupByID:
		b.WriteString("\tid = \"REPLACE_ME\" # TODO(tfgen): set to the seeded resource's id, e.g. datadog_x.foo.id\n")
	default:
		for _, f := range tv.Filters {
			fmt.Fprintf(&b, "\t%s = %s\n", f.TFName, f.HCLValue)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

// hclFilterValue returns the HCL placeholder for a scalar filter attribute type
// and whether the type is a supported scalar. Strings bind to uniq via a Sprintf
// verb; bool/number get a literal default the author overrides before recording.
func hclFilterValue(tfType string) (string, bool) {
	switch {
	case strings.Contains(tfType, "String"):
		return "%[1]q", true
	case strings.Contains(tfType, "Bool"):
		return "false", true
	case strings.Contains(tfType, "Int"), strings.Contains(tfType, "Float"):
		return "0", true
	default:
		return "", false
	}
}

// RenderDataSourceTest executes the acceptance-test template for v and returns
// gofmt-canonical Go source. The generated test is a scaffold: it compiles and
// runs, but needs a recorded cassette (and the seed resources / assertions the
// author fills in) to pass in replay mode.
func RenderDataSourceTest(v DataSourceView) ([]byte, error) {
	var buf bytes.Buffer
	if err := dataSourceTemplates.ExecuteTemplate(&buf, "data_source_test", buildTestView(v)); err != nil {
		return nil, fmt.Errorf("emit: executing data source test template for %q: %w", v.TypeName, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("emit: gofmt of generated test %q: %w\n--- raw output ---\n%s", v.TypeName, err, buf.String())
	}
	return dropBlankLineAfterBrace(formatted), nil
}
