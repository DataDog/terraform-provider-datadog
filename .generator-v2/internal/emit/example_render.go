package emit

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

const exampleDataSourceID = "11111111-2222-3333-4444-555555555555"

// maxPluralAutoFilters caps how many optional scalar filters the renderer will
// auto-select for an all-optional plural shape. One keeps the example minimal,
// illustrative, and byte-deterministic.
const maxPluralAutoFilters = 1

type exampleAttribute struct {
	name  string
	value string
}

// DataSourceExample is the rendered tfplugindocs input plus any limitations the
// generator could not express in HCL. Diagnostics are attached to the main
// artifact's run-report entry so downstream PR generation surfaces them; a
// warning marks an example a human must complete by hand.
type DataSourceExample struct {
	Content     []byte
	Diagnostics []model.Diagnostic
}

// RenderDataSourceExample returns a deterministic HCL scaffold for a generated
// data source. By-ID shapes use the repository's conventional example UUID,
// search-only shapes include their scalar filters, all-optional plural shapes
// auto-select a representative filter, and every shape includes required scalar
// attributes. It reports any input or lookup shape it cannot represent rather
// than silently presenting an incomplete example.
func RenderDataSourceExample(v DataSourceView) DataSourceExample {
	var attributes []exampleAttribute
	var diagnostics []model.Diagnostic

	if v.Cardinality == Singular && v.ByID {
		attributes = append(attributes, exampleAttribute{name: "id", value: fmt.Sprintf("%q", exampleDataSourceID)})
	}

	includeOptionalFilters := v.Cardinality == Singular && v.Searchable && !v.ByID
	renderedFilter := false
	requiredInputSeen := false
	for _, attr := range v.Schema.Attributes {
		if attr.Required {
			requiredInputSeen = true
		}
		if !attr.Required && !(includeOptionalFilters && attr.Optional) {
			continue
		}
		value, ok := hclExampleValue(attr.TFType)
		if !ok {
			if attr.Required {
				diagnostics = append(diagnostics, incompleteExample(v.TypeName, fmt.Sprintf(
					"required attribute %q has unsupported type %q", attr.TFName, attr.TFType)))
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
			requiredInputSeen = true
			diagnostics = append(diagnostics, incompleteExample(v.TypeName, fmt.Sprintf(
				"required block %q cannot be rendered", block.TFName)))
		}
	}

	// A plural shape usually exposes only optional query parameters, so the
	// passes above collect nothing and the example would fall through to an
	// empty block. Auto-select a stable, capped subset of the optional scalar
	// filters and flag it so a reviewer confirms the choice is representative.
	if v.Cardinality == Plural && len(attributes) == 0 && !requiredInputSeen {
		picked := 0
		for _, attr := range v.Schema.Attributes {
			if picked >= maxPluralAutoFilters {
				break
			}
			if !attr.Optional {
				continue
			}
			value, ok := hclExampleValue(attr.TFType)
			if !ok {
				continue
			}
			attributes = append(attributes, exampleAttribute{name: attr.TFName, value: value})
			picked++
		}
		if picked > 0 {
			diagnostics = append(diagnostics, model.Diagnostic{
				Severity: model.SeverityInfo,
				Message: fmt.Sprintf(
					"generated example for %q uses an auto-selected optional filter; confirm it is representative and add others as needed",
					v.TypeName),
			})
		} else {
			diagnostics = append(diagnostics, incompleteExample(v.TypeName,
				"all inputs are optional and none is a renderable scalar; a usable filter must be added by hand"))
		}
	}

	if v.Cardinality == Singular && !v.ByID && !v.Searchable {
		diagnostics = append(diagnostics, incompleteExample(v.TypeName,
			"singular lookup has neither by-ID nor searchable resolution"))
	}
	if includeOptionalFilters && !renderedFilter {
		diagnostics = append(diagnostics, incompleteExample(v.TypeName,
			"search-only lookup has no renderable scalar filters"))
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

// incompleteExample builds a warning-severity diagnostic for an example the
// generator could not fully render, so CI can flag that a human must supply one.
func incompleteExample(typeName, detail string) model.Diagnostic {
	return model.Diagnostic{
		Severity: model.SeverityWarning,
		Message:  fmt.Sprintf("generated example for %q may be incomplete: %s", typeName, detail),
	}
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
