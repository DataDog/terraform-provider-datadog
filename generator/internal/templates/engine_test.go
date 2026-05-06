package templates

import (
	"testing"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/model"
)

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	if engine == nil {
		t.Fatal("expected non-nil engine")
	}

	templates := engine.Templates()
	if len(templates) == 0 {
		t.Fatal("expected loaded templates, got none")
	}

	// Verify key templates are loaded
	expectedTemplates := []string{"main.go.tmpl", "model", "schema", "read", "state", "hooks.go.tmpl"}
	for _, name := range expectedTemplates {
		found := false
		for _, tmplName := range templates {
			if tmplName == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected template %q to be loaded, available: %v", name, templates)
		}
	}
}

func TestEngine_FuncMap(t *testing.T) {
	fm := FuncMap()

	expectedFuncs := []string{"snakeCase", "camelCase", "pascalCase", "sanitizeDescription", "goKeyword", "add", "sub", "join"}
	for _, name := range expectedFuncs {
		if _, ok := fm[name]; !ok {
			t.Errorf("expected FuncMap to have %q", name)
		}
	}
}

func TestEngine_RenderModel(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	data := &model.TemplateData{
		TypeName: "Team",
		Model: model.ModelData{
			StructName: "datadogTeamDataSourceModel",
			Fields: []model.ModelField{
				{GoName: "ID", TfsdkTag: "id", GoType: "types.String"},
				{GoName: "Name", TfsdkTag: "name", GoType: "types.String"},
				{GoName: "UserCount", TfsdkTag: "user_count", GoType: "types.Int64"},
			},
		},
	}

	result, err := engine.Render("model", data)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	output := string(result)

	expectedFragments := []string{
		"datadogTeamDataSourceModel",
		`tfsdk:"id"`,
		`tfsdk:"name"`,
		`tfsdk:"user_count"`,
		"types.String",
		"types.Int64",
	}

	for _, frag := range expectedFragments {
		if !containsStr(output, frag) {
			t.Errorf("expected output to contain %q, got:\n%s", frag, output)
		}
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
