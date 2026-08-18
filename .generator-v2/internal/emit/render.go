package emit

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates/*.go.tmpl
var templateFS embed.FS

// funcMap holds template helpers. By design these are presentation helpers,
// not arbitrary derivation: a View arrives fully populated. The type-derived
// helpers (hashExpr, fmtVerb) are the exception — they compute the non-pointer
// accessor and printf verb from a FilterParamView.ValueExpr, keeping those two
// redundant fields out of the struct.
var funcMap = template.FuncMap{
	"title": upperFirst,
	// hashExpr strips "Pointer" from a ValueExpr, e.g. "ValueStringPointer()" → "ValueString()".
	"hashExpr": func(valueExpr string) string {
		return strings.Replace(valueExpr, "Pointer()", "()", 1)
	},
	// fmtVerb returns the printf verb for a ValueExpr, e.g. "ValueStringPointer()" → "%s".
	"fmtVerb": func(valueExpr string) string {
		switch {
		case strings.HasPrefix(valueExpr, "ValueString"):
			return "%s"
		case strings.HasPrefix(valueExpr, "ValueBool"):
			return "%t"
		case strings.HasPrefix(valueExpr, "ValueInt64"), strings.HasPrefix(valueExpr, "ValueInt32"):
			return "%d"
		case strings.HasPrefix(valueExpr, "ValueFloat64"):
			return "%f"
		default:
			return "%v"
		}
	},
}

// dataSourceTemplates is the parsed template set: the singular and plural roots
// plus the shared partials in data_source_common.go.tmpl.
var dataSourceTemplates = template.Must(
	template.New("data_source").Funcs(funcMap).ParseFS(templateFS, "templates/*.go.tmpl"),
)

// RenderDataSource executes the singular or plural data-source template for v
// and returns gofmt-canonical Go source.
//
// NOTE: this is a minimal execution harness so the templates are runnable and
// testable on their own.
func RenderDataSource(v DataSourceView) ([]byte, error) {
	if err := checkDuplicateFields(v.Models); err != nil {
		return nil, fmt.Errorf("emit: data source %q: %w", v.TypeName, err)
	}

	name := "data_source_singular"
	if v.Cardinality == Plural {
		name = "data_source_plural"
	}

	var buf bytes.Buffer
	if err := dataSourceTemplates.ExecuteTemplate(&buf, name, v); err != nil {
		return nil, fmt.Errorf("emit: executing %s template for %q: %w", name, v.TypeName, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("emit: gofmt of generated data source %q: %w\n--- raw output ---\n%s", v.TypeName, err, buf.String())
	}
	return dropBlankLineAfterBrace(formatted), nil
}

// checkDuplicateFields rejects duplicate Go model names and a model struct that
// declares the same Terraform name twice. Either shape produces generated code
// that does not compile, so failing here keeps a broken artifact from being
// written and reported as created.
func checkDuplicateFields(models []ModelStructView) error {
	modelNames := make(map[string]bool, len(models))
	for _, m := range models {
		if modelNames[m.Name] {
			return fmt.Errorf("model %s is declared twice", m.Name)
		}
		modelNames[m.Name] = true

		seen := make(map[string]bool, len(m.Fields))
		for _, f := range m.Fields {
			if seen[f.TFName] {
				return fmt.Errorf("model %s declares tfsdk:%q twice", m.Name, f.TFName)
			}
			seen[f.TFName] = true
		}
	}
	return nil
}

// dropBlankLineAfterBrace removes blank lines that immediately follow a line
// ending in "{". gofmt does not strip these, but idiomatic Go never opens a
// block with a blank line; centralizing the rule here keeps the templates free
// of whitespace-control noise. Blank lines after any other line (e.g. the
// group separator before a "// Results" comment) are preserved.
func dropBlankLineAfterBrace(src []byte) []byte {
	lines := bytes.Split(src, []byte("\n"))
	out := make([][]byte, 0, len(lines))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 && len(out) > 0 {
			prev := bytes.TrimRight(out[len(out)-1], " \t")
			if n := len(prev); n > 0 && prev[n-1] == '{' {
				continue
			}
		}
		out = append(out, line)
	}
	return bytes.Join(out, []byte("\n"))
}

// upperFirst upper-cases the first rune of s.
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
