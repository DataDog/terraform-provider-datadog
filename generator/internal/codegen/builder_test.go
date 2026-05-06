package codegen

import (
	"strings"
	"testing"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/openapi"
)

func TestBuildTemplateData_SDKNames(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Description:      "Get a team by ID.",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
		PathParams: []openapi.Parameter{
			{Name: "team_id", Description: "The team ID.", Required: true},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString, Description: "Team name."},
			{Name: "handle", Type: openapi.FieldTypeString, Description: "Team handle."},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.SDKImportPath != "datadogV2" {
		t.Errorf("SDKImportPath = %q, want %q", data.SDKImportPath, "datadogV2")
	}
	if data.SDKApiType != "TeamsApi" {
		t.Errorf("SDKApiType = %q, want %q", data.SDKApiType, "TeamsApi")
	}
	if data.SDKApiAccessor != "GetTeamsApiV2" {
		t.Errorf("SDKApiAccessor = %q, want %q", data.SDKApiAccessor, "GetTeamsApiV2")
	}
	if data.SDKReadMethod != "GetTeam" {
		t.Errorf("SDKReadMethod = %q, want %q", data.SDKReadMethod, "GetTeam")
	}
	if data.TypeName != "Team" {
		t.Errorf("TypeName = %q, want %q", data.TypeName, "Team")
	}
	if data.SDKResponseType != "TeamResponse" {
		t.Errorf("SDKResponseType = %q, want %q", data.SDKResponseType, "TeamResponse")
	}
}

func TestBuildTemplateData_ModelFields(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
		PathParams: []openapi.Parameter{
			{Name: "team_id", Required: true},
		},
		QueryParams: []openapi.Parameter{
			{Name: "filter_keyword", Required: false},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
			{Name: "user_count", Type: openapi.FieldTypeInt64},
			{Name: "is_active", Type: openapi.FieldTypeBool},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: team_id (path), filter_keyword (query), id, name, user_count, is_active
	expectedFields := map[string]string{
		"TeamID":        "types.String",
		"FilterKeyword": "types.String",
		"ID":            "types.String",
		"Name":          "types.String",
		"UserCount":     "types.Int64",
		"IsActive":      "types.Bool",
	}

	if len(data.Model.Fields) != len(expectedFields) {
		t.Fatalf("expected %d model fields, got %d", len(expectedFields), len(data.Model.Fields))
	}

	for _, f := range data.Model.Fields {
		expected, ok := expectedFields[f.GoName]
		if !ok {
			t.Errorf("unexpected field %q", f.GoName)
			continue
		}
		if f.GoType != expected {
			t.Errorf("field %q: GoType = %q, want %q", f.GoName, f.GoType, expected)
		}
	}
}

func TestBuildTemplateData_SchemaAttributes(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
		PathParams: []openapi.Parameter{
			{Name: "team_id", Required: true, Description: "The team ID."},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString, Description: "Team name."},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// team_id should be Required, name should be Computed
	var teamIDAttr, nameAttr *bool
	for _, attr := range data.Schema.Attributes {
		if attr.Name == "team_id" {
			r := attr.Required
			teamIDAttr = &r
		}
		if attr.Name == "name" {
			c := attr.Computed
			nameAttr = &c
		}
	}

	if teamIDAttr == nil || !*teamIDAttr {
		t.Error("team_id should be Required")
	}
	if nameAttr == nil || !*nameAttr {
		t.Error("name should be Computed")
	}
}

func TestBuildTemplateData_NestedObject(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetComplex",
		Tag:              "Complex",
		Path:             "/api/v2/complex/{id}",
		Method:           "get",
		ResponseTypeName: "ComplexResponse",
		PathParams: []openapi.Parameter{
			{Name: "id", Required: true},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{
				Name: "config",
				Type: openapi.FieldTypeObject,
				Children: &openapi.SchemaObject{
					Fields: []openapi.SchemaField{
						{Name: "key", Type: openapi.FieldTypeString},
						{Name: "threshold", Type: openapi.FieldTypeFloat32},
					},
				},
			},
		},
	}

	data, err := BuildTemplateData("complex", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check block was created
	if len(data.Schema.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(data.Schema.Blocks))
	}
	if data.Schema.Blocks[0].Type != "SingleNestedBlock" {
		t.Errorf("block type = %q, want %q", data.Schema.Blocks[0].Type, "SingleNestedBlock")
	}

	// Check nested struct was created
	if len(data.Model.NestedStructs) != 1 {
		t.Fatalf("expected 1 nested struct, got %d", len(data.Model.NestedStructs))
	}
	if data.Model.NestedStructs[0].StructName != "ComplexConfigModel" {
		t.Errorf("nested struct name = %q, want %q", data.Model.NestedStructs[0].StructName, "ComplexConfigModel")
	}
}

func TestBuildTemplateData_StateMapping(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
			{Name: "link_count", Type: openapi.FieldTypeInt32},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.State.FieldMappings) != 2 {
		t.Fatalf("expected 2 field mappings, got %d", len(data.State.FieldMappings))
	}

	// Check name mapping
	nameMapping := data.State.FieldMappings[0]
	if nameMapping.ModelField != "state.Name" {
		t.Errorf("name ModelField = %q, want %q", nameMapping.ModelField, "state.Name")
	}
	if nameMapping.TypeConverter != "types.StringValue" {
		t.Errorf("name TypeConverter = %q, want %q", nameMapping.TypeConverter, "types.StringValue")
	}

	// Check link_count mapping (int32 needs cast)
	linkMapping := data.State.FieldMappings[1]
	if linkMapping.NeedsCast != "int64" {
		t.Errorf("link_count NeedsCast = %q, want %q", linkMapping.NeedsCast, "int64")
	}
}

// T067/T069: Response type from $ref
func TestBuildTemplateData_ResponseTypeFromRef(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.SDKResponseType != "TeamResponse" {
		t.Errorf("SDKResponseType = %q, want %q", data.SDKResponseType, "TeamResponse")
	}
}

func TestBuildTemplateData_MissingRefFallsBack(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "", // No $ref
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fall back to OperationID + "Response"
	if data.SDKResponseType != "GetTeamResponse" {
		t.Errorf("SDKResponseType = %q, want %q", data.SDKResponseType, "GetTeamResponse")
	}
}

// T070/T072: Nullable field handling
func TestBuildTemplateData_NullableFields(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString, Nullable: false},
			{Name: "description", Type: openapi.FieldTypeString, Nullable: true},
			{Name: "count", Type: openapi.FieldTypeInt64, Nullable: true},
			{Name: "score", Type: openapi.FieldTypeFloat64, Nullable: true},
			{Name: "enabled", Type: openapi.FieldTypeBool, Nullable: true},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		fieldName    string
		wantNullable bool
		wantNullVal  string
	}{
		{"state.Name", false, ""},
		{"state.Description", true, "types.StringNull()"},
		{"state.Count", true, "types.Int64Null()"},
		{"state.Score", true, "types.Float64Null()"},
		{"state.Enabled", true, "types.BoolNull()"},
	}

	for _, tt := range tests {
		var found bool
		for _, m := range data.State.FieldMappings {
			if m.ModelField == tt.fieldName {
				found = true
				if m.Nullable != tt.wantNullable {
					t.Errorf("%s: Nullable = %v, want %v", tt.fieldName, m.Nullable, tt.wantNullable)
				}
				if m.NullValue != tt.wantNullVal {
					t.Errorf("%s: NullValue = %q, want %q", tt.fieldName, m.NullValue, tt.wantNullVal)
				}
				if tt.wantNullable && m.SDKOkAccessor == "" {
					t.Errorf("%s: SDKOkAccessor should be set for nullable field", tt.fieldName)
				}
				if !tt.wantNullable && m.SDKOkAccessor != "" {
					t.Errorf("%s: SDKOkAccessor should be empty for non-nullable field", tt.fieldName)
				}
			}
		}
		if !found {
			t.Errorf("field mapping %q not found", tt.fieldName)
		}
	}
}

// T073/T074: Date-time field exclusion
func TestBuildTemplateData_DateTimeExclusion(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
			{Name: "created_at", Type: openapi.FieldTypeString, Format: "date-time"},
			{Name: "modified_at", Type: openapi.FieldTypeString, Format: "date-time"},
			{Name: "website", Type: openapi.FieldTypeString, Format: "uri"},
			{Name: "count", Type: openapi.FieldTypeInt64},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Model fields should not contain created_at or modified_at
	for _, f := range data.Model.Fields {
		if f.TfsdkTag == "created_at" || f.TfsdkTag == "modified_at" {
			t.Errorf("model should not contain date-time field %q", f.TfsdkTag)
		}
	}

	// Schema attributes should not contain created_at or modified_at
	for _, a := range data.Schema.Attributes {
		if a.Name == "created_at" || a.Name == "modified_at" {
			t.Errorf("schema should not contain date-time field %q", a.Name)
		}
	}

	// State mappings should not contain created_at or modified_at
	for _, m := range data.State.FieldMappings {
		if m.ModelField == "state.CreatedAt" || m.ModelField == "state.ModifiedAt" {
			t.Errorf("state should not contain date-time field %q", m.ModelField)
		}
	}

	// Non-date-time fields should still be present
	foundName, foundWebsite, foundCount := false, false, false
	for _, m := range data.State.FieldMappings {
		switch m.ModelField {
		case "state.Name":
			foundName = true
		case "state.Website":
			foundWebsite = true
		case "state.Count":
			foundCount = true
		}
	}
	if !foundName {
		t.Error("name field should be present in state")
	}
	if !foundWebsite {
		t.Error("website (format: uri) field should be present in state")
	}
	if !foundCount {
		t.Error("count field should be present in state")
	}
}

// T075/T077: List field conversion
func TestBuildTemplateData_ListFields(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
			{Name: "tags", Type: openapi.FieldTypeArrayOfStrings},
			{Name: "counts", Type: openapi.FieldTypeArrayOfInts},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		modelField   string
		wantIsList   bool
		wantElemType string
	}{
		{"state.Name", false, ""},
		{"state.Tags", true, "types.StringType"},
		{"state.Counts", true, "types.Int64Type"},
	}

	for _, tt := range tests {
		var found bool
		for _, m := range data.State.FieldMappings {
			if m.ModelField == tt.modelField {
				found = true
				if m.IsList != tt.wantIsList {
					t.Errorf("%s: IsList = %v, want %v", tt.modelField, m.IsList, tt.wantIsList)
				}
				if m.ListElementType != tt.wantElemType {
					t.Errorf("%s: ListElementType = %q, want %q", tt.modelField, m.ListElementType, tt.wantElemType)
				}
			}
		}
		if !found {
			t.Errorf("field mapping %q not found", tt.modelField)
		}
	}
}

// T056/T057: Filter-fallback builder
func TestBuildTemplateData_WithListOp(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
		PathParams: []openapi.Parameter{
			{Name: "team_id", Required: true, Description: "The team ID."},
		},
	}

	listOp := &openapi.ParsedOperation{
		OperationID: "ListTeams",
		Tag:         "Teams",
		Path:        "/api/v2/team",
		Method:      "get",
		QueryParams: []openapi.Parameter{
			{Name: "filter[keyword]", Description: "Search query."},
			{Name: "filter[me]", Description: "Only my teams."},
			{Name: "page[number]", Description: "Page number."},
			{Name: "page[size]", Description: "Page size."},
			{Name: "sort", Description: "Sort order."},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, listOp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify HasListFallback
	if !data.HasListFallback {
		t.Error("HasListFallback should be true")
	}
	if data.SDKListMethod != "ListTeams" {
		t.Errorf("SDKListMethod = %q, want %q", data.SDKListMethod, "ListTeams")
	}
	if data.SDKListOptParams != "ListTeamsOptionalParameters" {
		t.Errorf("SDKListOptParams = %q, want %q", data.SDKListOptParams, "ListTeamsOptionalParameters")
	}

	// Verify filter params in ReadData
	if len(data.Read.FilterParams) != 2 {
		t.Fatalf("expected 2 filter params, got %d", len(data.Read.FilterParams))
	}
	if data.Read.FilterParams[0].GoName != "FilterKeyword" {
		t.Errorf("filter param[0] GoName = %q, want %q", data.Read.FilterParams[0].GoName, "FilterKeyword")
	}
	if data.Read.FilterParams[1].GoName != "FilterMe" {
		t.Errorf("filter param[1] GoName = %q, want %q", data.Read.FilterParams[1].GoName, "FilterMe")
	}

	// Verify path param becomes Optional+Computed
	var teamIDAttr *struct{ optional, computed bool }
	for _, attr := range data.Schema.Attributes {
		if attr.Name == "team_id" {
			teamIDAttr = &struct{ optional, computed bool }{attr.Optional, attr.Computed}
		}
	}
	if teamIDAttr == nil {
		t.Fatal("team_id attribute not found in schema")
	}
	if !teamIDAttr.optional {
		t.Error("team_id should be Optional with list fallback")
	}
	if !teamIDAttr.computed {
		t.Error("team_id should be Computed with list fallback")
	}

	// Verify filter params appear as Optional schema attributes
	var foundFilterKeyword, foundFilterMe bool
	for _, attr := range data.Schema.Attributes {
		if attr.Name == "filter_keyword" {
			foundFilterKeyword = true
			if !attr.Optional {
				t.Error("filter_keyword should be Optional")
			}
		}
		if attr.Name == "filter_me" {
			foundFilterMe = true
			if !attr.Optional {
				t.Error("filter_me should be Optional")
			}
		}
	}
	if !foundFilterKeyword {
		t.Error("filter_keyword not found in schema attributes")
	}
	if !foundFilterMe {
		t.Error("filter_me not found in schema attributes")
	}

	// Verify filter params appear in model fields
	var foundModelFK, foundModelFM bool
	for _, f := range data.Model.Fields {
		if f.GoName == "FilterKeyword" {
			foundModelFK = true
		}
		if f.GoName == "FilterMe" {
			foundModelFM = true
		}
	}
	if !foundModelFK {
		t.Error("FilterKeyword not found in model fields")
	}
	if !foundModelFM {
		t.Error("FilterMe not found in model fields")
	}
}

// Multi-word API tags (e.g., "Cloud Cost Management") produce valid Go identifiers
func TestBuildTemplateData_MultiWordTag(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetCostBudget",
		Tag:              "Cloud Cost Management",
		Description:      "Get a cost budget.",
		Path:             "/api/v2/cost/budget/{budget_id}",
		Method:           "get",
		ResponseTypeName: "CostBudgetResponse",
		PathParams: []openapi.Parameter{
			{Name: "budget_id", Description: "The budget ID.", Required: true},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString, Description: "Budget name."},
		},
	}

	data, err := BuildTemplateData("cost_budget", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.SDKApiType != "CloudCostManagementApi" {
		t.Errorf("SDKApiType = %q, want %q", data.SDKApiType, "CloudCostManagementApi")
	}
	if data.SDKApiAccessor != "GetCloudCostManagementApiV2" {
		t.Errorf("SDKApiAccessor = %q, want %q", data.SDKApiAccessor, "GetCloudCostManagementApiV2")
	}
}

// T084: Block type classification and NeedsNestedObject
func TestBuildTemplateData_BlockTypes(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetComplex",
		Tag:              "Complex",
		Path:             "/api/v2/complex/{id}",
		Method:           "get",
		ResponseTypeName: "ComplexResponse",
		PathParams: []openapi.Parameter{
			{Name: "id", Required: true},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{
				Name:        "config",
				Type:        openapi.FieldTypeObject,
				Description: "Configuration settings.",
				Children: &openapi.SchemaObject{
					Fields: []openapi.SchemaField{
						{Name: "key", Type: openapi.FieldTypeString, Description: "Config key."},
					},
				},
			},
			{
				Name:        "endpoints",
				Type:        openapi.FieldTypeArrayOfObjects,
				Description: "List of endpoints.",
				Children: &openapi.SchemaObject{
					Fields: []openapi.SchemaField{
						{Name: "url", Type: openapi.FieldTypeString, Description: "Endpoint URL."},
						{Name: "weight", Type: openapi.FieldTypeInt64, Description: "Endpoint weight."},
					},
				},
			},
		},
	}

	data, err := BuildTemplateData("complex", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data.Schema.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(data.Schema.Blocks))
	}

	// config -> SingleNestedBlock (object, not array)
	configBlock := data.Schema.Blocks[0]
	if configBlock.Name != "config" {
		t.Errorf("block[0] name = %q, want %q", configBlock.Name, "config")
	}
	if configBlock.Type != "SingleNestedBlock" {
		t.Errorf("config block type = %q, want %q", configBlock.Type, "SingleNestedBlock")
	}
	if configBlock.NeedsNestedObject() {
		t.Error("SingleNestedBlock should NOT need NestedObject wrapper")
	}
	if len(configBlock.Attributes) == 0 {
		t.Error("config block should have attributes")
	}

	// endpoints -> ListNestedBlock (array of objects)
	endpointsBlock := data.Schema.Blocks[1]
	if endpointsBlock.Name != "endpoints" {
		t.Errorf("block[1] name = %q, want %q", endpointsBlock.Name, "endpoints")
	}
	if endpointsBlock.Type != "ListNestedBlock" {
		t.Errorf("endpoints block type = %q, want %q", endpointsBlock.Type, "ListNestedBlock")
	}
	if !endpointsBlock.NeedsNestedObject() {
		t.Error("ListNestedBlock SHOULD need NestedObject wrapper")
	}
	if len(endpointsBlock.Attributes) == 0 {
		t.Error("endpoints block should have attributes")
	}
}

// T089: SDK accessor casing uses ToSDKPascalCase (trailing acronyms lowered)
func TestBuildTemplateData_SDKCasing(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "org_id", Type: openapi.FieldTypeString},
			{Name: "name", Type: openapi.FieldTypeString},
			{Name: "team_url", Type: openapi.FieldTypeString, Nullable: true},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify SDK accessors use SDK PascalCase (lowercased trailing acronyms)
	sdkTests := []struct {
		modelField        string
		wantSDKAccessor   string
		wantSDKOkAccessor string
	}{
		{"state.OrgID", "attributes.GetOrgId()", ""},
		{"state.Name", "attributes.GetName()", ""},
		{"state.TeamURL", "attributes.GetTeamUrl()", "attributes.GetTeamUrlOk()"},
	}

	for _, tt := range sdkTests {
		var found bool
		for _, m := range data.State.FieldMappings {
			if m.ModelField == tt.modelField {
				found = true
				if m.SDKAccessor != tt.wantSDKAccessor {
					t.Errorf("%s: SDKAccessor = %q, want %q", tt.modelField, m.SDKAccessor, tt.wantSDKAccessor)
				}
				if tt.wantSDKOkAccessor != "" && m.SDKOkAccessor != tt.wantSDKOkAccessor {
					t.Errorf("%s: SDKOkAccessor = %q, want %q", tt.modelField, m.SDKOkAccessor, tt.wantSDKOkAccessor)
				}
			}
		}
		if !found {
			t.Errorf("field mapping %q not found", tt.modelField)
		}
	}

	// Verify model struct fields use standard PascalCase (NOT SDK casing)
	modelGoNames := make(map[string]bool)
	for _, f := range data.Model.Fields {
		modelGoNames[f.GoName] = true
	}
	if !modelGoNames["OrgID"] {
		t.Error("model field should use standard PascalCase: OrgID (not OrgId)")
	}
	if !modelGoNames["TeamURL"] {
		t.Error("model field should use standard PascalCase: TeamURL (not TeamUrl)")
	}
}

// T105: Builder handles map-typed fields from composition
func TestBuildTemplateData_MapFields(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetComposed",
		Tag:              "Composed",
		Path:             "/api/v2/composed/{id}",
		Method:           "get",
		ResponseTypeName: "ComposedResponse",
		PathParams: []openapi.Parameter{
			{Name: "id", Required: true},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString, Description: "Name."},
			{Name: "labels", Type: openapi.FieldTypeMapOfStrings, Description: "Labels."},
			{Name: "counts", Type: openapi.FieldTypeMapOfInts, Description: "Counts."},
		},
	}

	data, err := BuildTemplateData("composed", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check state field mappings for map fields
	tests := []struct {
		modelField     string
		wantIsMap      bool
		wantMapElemTyp string
	}{
		{"state.Name", false, ""},
		{"state.Labels", true, "types.StringType"},
		{"state.Counts", true, "types.Int64Type"},
	}

	for _, tt := range tests {
		var found bool
		for _, m := range data.State.FieldMappings {
			if m.ModelField == tt.modelField {
				found = true
				if m.IsMap != tt.wantIsMap {
					t.Errorf("%s: IsMap = %v, want %v", tt.modelField, m.IsMap, tt.wantIsMap)
				}
				if m.MapElementType != tt.wantMapElemTyp {
					t.Errorf("%s: MapElementType = %q, want %q", tt.modelField, m.MapElementType, tt.wantMapElemTyp)
				}
			}
		}
		if !found {
			t.Errorf("field mapping %q not found", tt.modelField)
		}
	}

	// Check schema attributes include map attributes with ElementType
	var foundLabels bool
	for _, attr := range data.Schema.Attributes {
		if attr.Name == "labels" {
			foundLabels = true
			if attr.Type != "schema.MapAttribute" {
				t.Errorf("labels schema type = %q, want schema.MapAttribute", attr.Type)
			}
			if attr.ElementType != "types.StringType" {
				t.Errorf("labels ElementType = %q, want types.StringType", attr.ElementType)
			}
		}
	}
	if !foundLabels {
		t.Error("labels attribute not found in schema")
	}

	// Check model fields include map as types.Map
	var foundLabelsModel bool
	for _, f := range data.Model.Fields {
		if f.GoName == "Labels" {
			foundLabelsModel = true
			if f.GoType != "types.Map" {
				t.Errorf("Labels GoType = %q, want types.Map", f.GoType)
			}
		}
	}
	if !foundLabelsModel {
		t.Error("Labels not found in model fields")
	}
}

// T105: Non-map field has IsMap=false (backward compat)
func TestBuildTemplateData_NonMapFieldIsMapFalse(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, m := range data.State.FieldMappings {
		if m.IsMap {
			t.Errorf("non-map field %q should have IsMap=false", m.ModelField)
		}
		if m.MapElementType != "" {
			t.Errorf("non-map field %q should have empty MapElementType", m.ModelField)
		}
	}
}

// F1/F8: Empty tag validation — FR-5 requires error when operation has no tag
func TestBuildTemplateData_EmptyTag(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetWidget",
		Tag:              "", // empty tag
		Path:             "/api/v2/widget/{id}",
		Method:           "get",
		ResponseTypeName: "WidgetResponse",
		PathParams:       []openapi.Parameter{{Name: "id", Required: true}},
	}
	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
		},
	}
	_, err := BuildTemplateData("widget", op, schema, false, nil)
	if err == nil {
		t.Fatal("expected error for empty tag, got nil")
	}
	if !strings.Contains(err.Error(), "no tag") {
		t.Errorf("error should mention missing tag, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "/api/v2/widget/{id}") {
		t.Errorf("error should identify the operation path, got: %s", err.Error())
	}
}

// T057: Builder without listOp produces same output as before (backward compatibility)
func TestBuildTemplateData_WithoutListOp(t *testing.T) {
	op := &openapi.ParsedOperation{
		OperationID:      "GetTeam",
		Tag:              "Teams",
		Path:             "/api/v2/team/{team_id}",
		Method:           "get",
		ResponseTypeName: "TeamResponse",
		PathParams: []openapi.Parameter{
			{Name: "team_id", Required: true, Description: "The team ID."},
		},
	}

	schema := &openapi.SchemaObject{
		Fields: []openapi.SchemaField{
			{Name: "name", Type: openapi.FieldTypeString},
		},
	}

	data, err := BuildTemplateData("team", op, schema, true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.HasListFallback {
		t.Error("HasListFallback should be false without listOp")
	}
	if data.Read.HasListFallback {
		t.Error("Read.HasListFallback should be false without listOp")
	}
	if data.Schema.HasListFallback {
		t.Error("Schema.HasListFallback should be false without listOp")
	}

	// Path param should be Required
	for _, attr := range data.Schema.Attributes {
		if attr.Name == "team_id" {
			if !attr.Required {
				t.Error("team_id should be Required without list fallback")
			}
			if attr.Optional {
				t.Error("team_id should not be Optional without list fallback")
			}
		}
	}
}
