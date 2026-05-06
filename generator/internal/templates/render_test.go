package templates

import (
	"strings"
	"testing"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/model"
)

// T085: Template rendering test for NestedObject wrapper
func TestRenderSchemaBlock_NestedObject(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	data := &model.TemplateData{
		TypeName:    "Complex",
		Description: "A complex resource.",
		Schema: model.SchemaData{
			Attributes: []model.AttributeData{
				{Name: "title", Type: "schema.StringAttribute", Description: "Title.", Computed: true},
			},
			Blocks: []model.BlockData{
				{
					Name:        "config",
					Type:        "SingleNestedBlock",
					Description: "Config settings.",
					Attributes: []model.AttributeData{
						{Name: "key", Type: "schema.StringAttribute", Description: "Config key.", Computed: true},
					},
				},
				{
					Name:        "endpoints",
					Type:        "ListNestedBlock",
					Description: "Endpoint list.",
					Attributes: []model.AttributeData{
						{Name: "url", Type: "schema.StringAttribute", Description: "URL.", Computed: true},
						{Name: "weight", Type: "schema.Int64Attribute", Description: "Weight.", Computed: true},
					},
				},
			},
		},
	}

	output, err := engine.Render("schema", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)

	// ListNestedBlock should have NestedObject wrapper
	if !strings.Contains(s, "schema.ListNestedBlock{") {
		t.Error("output should contain schema.ListNestedBlock{")
	}
	if !strings.Contains(s, "NestedObject: schema.NestedBlockObject{") {
		t.Error("ListNestedBlock should have NestedObject: schema.NestedBlockObject{")
	}

	// SingleNestedBlock should NOT have NestedObject wrapper
	if !strings.Contains(s, "schema.SingleNestedBlock{") {
		t.Error("output should contain schema.SingleNestedBlock{")
	}

	// Verify that SingleNestedBlock section does not contain NestedObject
	singleIdx := strings.Index(s, "schema.SingleNestedBlock{")
	listIdx := strings.Index(s, "schema.ListNestedBlock{")
	if singleIdx >= 0 && listIdx >= 0 && singleIdx < listIdx {
		singleSection := s[singleIdx:listIdx]
		if strings.Contains(singleSection, "NestedObject") {
			t.Error("SingleNestedBlock section should NOT contain NestedObject wrapper")
		}
	}

	// Both blocks should contain Attributes maps
	// Top-level attributes + config block attributes + endpoints block attributes (inside NestedObject)
	attrCount := strings.Count(s, "Attributes: map[string]schema.Attribute{")
	if attrCount < 3 {
		t.Errorf("expected at least 3 Attributes maps (top-level + 2 blocks), got %d", attrCount)
	}
}

// T107: Template rendering tests for map attributes
func TestRenderSchemaBlock_MapAttribute(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	data := &model.TemplateData{
		TypeName:    "Composed",
		Description: "A composed resource.",
		Schema: model.SchemaData{
			Attributes: []model.AttributeData{
				{Name: "name", Type: "schema.StringAttribute", Description: "Name.", Computed: true},
				{Name: "labels", Type: "schema.MapAttribute", Description: "Labels.", Computed: true, ElementType: "types.StringType"},
			},
		},
	}

	output, err := engine.Render("schema", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)

	if !strings.Contains(s, "schema.MapAttribute") {
		t.Error("output should contain schema.MapAttribute")
	}
	if !strings.Contains(s, "ElementType: types.StringType") {
		t.Error("map attribute should have ElementType: types.StringType")
	}
}

func TestRenderState_MapValueFrom(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	data := &model.TemplateData{
		TypeName:        "Composed",
		SDKImportPath:   "datadogV2",
		SDKResponseType: "ComposedResponse",
		Read: model.ReadData{
			IsJSONAPI: true,
		},
		Model: model.ModelData{
			StructName: "datadogComposedDataSourceModel",
		},
		State: model.StateData{
			FieldMappings: []model.FieldMapping{
				{
					ModelField:     "state.Labels",
					SDKAccessor:    "attributes.GetLabels()",
					TypeConverter:  "types.StringValue",
					IsMap:          true,
					MapElementType: "types.StringType",
				},
				{
					ModelField:    "state.Name",
					SDKAccessor:   "attributes.GetName()",
					TypeConverter: "types.StringValue",
				},
			},
		},
	}

	output, err := engine.Render("state", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)

	if !strings.Contains(s, "types.MapValueFrom(ctx, types.StringType, attributes.GetLabels())") {
		t.Error("map field should use types.MapValueFrom")
	}
	// Non-map field should use regular assignment
	if !strings.Contains(s, "state.Name = types.StringValue(attributes.GetName())") {
		t.Error("non-map field should use regular assignment")
	}
}

func TestRenderModel_MapField(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}

	data := &model.TemplateData{
		Model: model.ModelData{
			StructName: "datadogComposedDataSourceModel",
			Fields: []model.ModelField{
				{GoName: "ID", TfsdkTag: "id", GoType: "types.String"},
				{GoName: "Labels", TfsdkTag: "labels", GoType: "types.Map"},
			},
		},
	}

	output, err := engine.Render("model", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)

	if !strings.Contains(s, "Labels types.Map") {
		t.Error("model should contain Labels types.Map")
	}
}

// T114b: Schema template calls ModifySchema hook
func TestRenderSchema_ModifySchemaHook(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}
	data := &model.TemplateData{
		TypeName:    "Team",
		Description: "Team data source.",
		Schema: model.SchemaData{
			Attributes: []model.AttributeData{
				{Name: "name", Type: "schema.StringAttribute", Description: "Name.", Computed: true},
			},
		},
	}
	output, err := engine.Render("schema", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)
	if !strings.Contains(s, "d.Hooks.ModifySchema(&resp.Schema)") {
		t.Error("schema template should call d.Hooks.ModifySchema")
	}
}

// T115b: Read template calls BeforeRead and AfterRead hooks
func TestRenderRead_Hooks(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}
	data := &model.TemplateData{
		TypeName:        "Team",
		SDKImportPath:   "datadogV2",
		SDKReadMethod:   "GetTeam",
		SDKResponseType: "TeamResponse",
		Model:           model.ModelData{StructName: "datadogTeamDataSourceModel"},
		Read: model.ReadData{
			PathParams: []model.ParamData{{Name: "team_id", GoName: "TeamId", ValueMethod: "ValueString()"}},
			IsJSONAPI:  true,
		},
	}
	output, err := engine.Render("read", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)
	if !strings.Contains(s, "d.Hooks.BeforeRead(ctx, req, &state)") {
		t.Error("read template should call d.Hooks.BeforeRead")
	}
	if !strings.Contains(s, "d.Hooks.AfterRead(ctx, &state)") {
		t.Error("read template should call d.Hooks.AfterRead")
	}
}

// F6: State template renders variant probing for all variants (anyOf multi-variant support)
func TestRenderState_VariantProbing(t *testing.T) {
	engine, err := NewEngine()
	if err != nil {
		t.Fatalf("NewEngine() error: %v", err)
	}
	data := &model.TemplateData{
		TypeName:        "Composed",
		SDKImportPath:   "datadogV2",
		SDKResponseType: "ComposedResponse",
		Read:            model.ReadData{IsJSONAPI: true},
		Model:           model.ModelData{StructName: "datadogComposedDataSourceModel"},
		State: model.StateData{
			FieldMappings: []model.FieldMapping{
				{ModelField: "state.Name", SDKAccessor: "attributes.GetName()", TypeConverter: "types.StringValue"},
			},
			VariantBlocks: []model.VariantBlockMapping{
				{
					ParentField: "Config",
					Variants: []model.VariantMapping{
						{BlockName: "config_type_a", SDKOkAccessor: "attributes.GetConfigTypeAOk()", HelperFunc: "updateConfigTypeAState", ModelField: "ConfigTypeA"},
						{BlockName: "config_type_b", SDKOkAccessor: "attributes.GetConfigTypeBOk()", HelperFunc: "updateConfigTypeBState", ModelField: "ConfigTypeB"},
						{BlockName: "config_type_c", SDKOkAccessor: "attributes.GetConfigTypeCOk()", HelperFunc: "updateConfigTypeCState", ModelField: "ConfigTypeC"},
					},
				},
			},
		},
	}
	output, err := engine.Render("state", data)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	s := string(output)

	// All 3 variants should have probing code (anyOf: all non-nil variants populated)
	if !strings.Contains(s, "attributes.GetConfigTypeAOk()") {
		t.Error("state template should probe variant ConfigTypeA")
	}
	if !strings.Contains(s, "attributes.GetConfigTypeBOk()") {
		t.Error("state template should probe variant ConfigTypeB")
	}
	if !strings.Contains(s, "attributes.GetConfigTypeCOk()") {
		t.Error("state template should probe variant ConfigTypeC")
	}

	// Each variant should call its helper when non-nil
	if !strings.Contains(s, "updateConfigTypeAState(ctx, v)") {
		t.Error("state template should call updateConfigTypeAState helper")
	}

	// Each variant should set nil when absent
	nilCount := strings.Count(s, "= nil")
	if nilCount < 3 {
		t.Errorf("expected at least 3 nil assignments for absent variants, got %d", nilCount)
	}

	// Comment should mention oneOf/anyOf
	if !strings.Contains(s, "oneOf/anyOf variant probing") {
		t.Error("state template should include variant probing comment")
	}
}
