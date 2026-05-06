package templates

import (
	"bytes"
	"fmt"
	"text/template"

	embeddedtmpl "github.com/DataDog/terraform-provider-datadog/generator/templates"
)

// Engine wraps a parsed set of Go templates for rendering.
type Engine struct {
	tmpl *template.Template
}

// NewEngine creates a new template engine, loading all .tmpl files from
// the embedded templates/datasource/ directory.
func NewEngine() (*Engine, error) {
	tmpl, err := template.New("").
		Funcs(FuncMap()).
		ParseFS(embeddedtmpl.DatasourceTemplates, "datasource/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	return &Engine{tmpl: tmpl}, nil
}

// Render executes the named template with the given data and returns
// the rendered bytes.
func (e *Engine) Render(templateName string, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := e.tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return nil, fmt.Errorf("executing template %q: %w", templateName, err)
	}
	return buf.Bytes(), nil
}

// Templates returns the names of all loaded templates.
func (e *Engine) Templates() []string {
	var names []string
	for _, t := range e.tmpl.Templates() {
		if t.Name() != "" {
			names = append(names, t.Name())
		}
	}
	return names
}
