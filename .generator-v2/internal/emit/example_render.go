package emit

import (
	"fmt"
	"strings"
)

const exampleDataSourceID = "11111111-2222-3333-4444-555555555555"

type exampleAttribute struct {
	name  string
	value string
}

// DataSourceExample is the rendered tfplugindocs input plus any limitations the
// generator could not express in HCL. Diagnostics are attached to the main
// artifact's run-report entry so downstream PR generation surfaces them.
type DataSourceExample struct {
	Content     []byte
	Diagnostics []string
}

// RenderDataSourceExample returns a deterministic HCL scaffold for a generated
// data source. By-ID shapes use the repository's conventional example UUID,
// search-only shapes include their scalar filters, and every shape includes
// required scalar attributes. It reports any required input or lookup shape it
// cannot represent rather than silently presenting an incomplete example.
func RenderDataSourceExample(v DataSourceView) DataSourceExample {
	var attributes []exampleAttribute
	var diagnostics []string

	if v.Cardinality == Singular && v.ByID {
		attributes = append(attributes, exampleAttribute{name: "id", value: fmt.Sprintf("%q", exampleDataSourceID)})
	}

	includeOptionalFilters := v.Cardinality == Singular && v.Searchable && !v.ByID
	renderedFilter := false
	for _, attr := range v.Schema.Attributes {
		if !attr.Required && !(includeOptionalFilters && attr.Optional) {
			continue
		}
		value, ok := hclExampleValue(attr.TFType)
		if !ok {
			if attr.Required {
				diagnostics = append(diagnostics, fmt.Sprintf(
					"generated example for %q may be incomplete: required attribute %q has unsupported type %q",
					v.TypeName, attr.TFName, attr.TFType,
				))
			}
			continue
		}
		attributes = append(attributes, exampleAttribute{name: attr.TFName, value: value})
		if includeOptionalFilters {
			renderedFilter = true
		}
	}
	for _, block := range v.Schema.Blocks {
		if block.Required {
			diagnostics = append(diagnostics, fmt.Sprintf(
				"generated example for %q may be incomplete: required block %q cannot be rendered",
				v.TypeName, block.TFName,
			))
		}
	}

	if v.Cardinality == Singular && !v.ByID && !v.Searchable {
		diagnostics = append(diagnostics, fmt.Sprintf(
			"generated example for %q may be incomplete: singular lookup has neither by-ID nor searchable resolution",
			v.TypeName,
		))
	}
	if includeOptionalFilters && !renderedFilter {
		diagnostics = append(diagnostics, fmt.Sprintf(
			"generated example for %q may be incomplete: search-only lookup has no renderable scalar filters",
			v.TypeName,
		))
	}

	typeName := "datadog_" + v.TypeName
	if len(attributes) == 0 {
		return DataSourceExample{
			Content:     []byte(fmt.Sprintf("data %q \"example\" {}\n", typeName)),
			Diagnostics: diagnostics,
		}
	}

	width := 0
	for _, attr := range attributes {
		if len(attr.name) > width {
			width = len(attr.name)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "data %q \"example\" {\n", typeName)
	for _, attr := range attributes {
		fmt.Fprintf(&b, "  %-*s = %s\n", width, attr.name, attr.value)
	}
	b.WriteString("}\n")
	return DataSourceExample{Content: []byte(b.String()), Diagnostics: diagnostics}
}

// hclExampleValue returns a stable HCL literal for a scalar schema attribute.
func hclExampleValue(tfType string) (string, bool) {
	switch {
	case strings.Contains(tfType, "String"):
		return `"example"`, true
	case strings.Contains(tfType, "Bool"):
		return "false", true
	case strings.Contains(tfType, "Int"), strings.Contains(tfType, "Float"):
		return "0", true
	default:
		return "", false
	}
}
