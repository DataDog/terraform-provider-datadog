package openapi

import (
	"testing"
)

func TestExtractOperation(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}
	doc := &model.Model

	tests := []struct {
		name           string
		path           string
		method         string
		wantOpID       string
		wantTag        string
		wantPathCount  int
		wantQueryCount int
		wantErr        string
	}{
		{
			name:           "GET team endpoint",
			path:           "/api/v2/team/{team_id}",
			method:         "get",
			wantOpID:       "GetTeam",
			wantTag:        "Teams",
			wantPathCount:  1,
			wantQueryCount: 1,
		},
		{
			name:           "GET simple endpoint",
			path:           "/api/v2/simple/{id}",
			method:         "get",
			wantOpID:       "GetSimple",
			wantTag:        "Simple",
			wantPathCount:  1,
			wantQueryCount: 0,
		},
		{
			name:    "missing path",
			path:    "/api/v2/nonexistent",
			method:  "get",
			wantErr: "not found in spec",
		},
		{
			name:    "missing method",
			path:    "/api/v2/team/{team_id}",
			method:  "post",
			wantErr: "not found for path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := ExtractOperation(doc, tt.path, tt.method)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !containsString(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if op.OperationID != tt.wantOpID {
				t.Errorf("OperationID = %q, want %q", op.OperationID, tt.wantOpID)
			}
			if op.Tag != tt.wantTag {
				t.Errorf("Tag = %q, want %q", op.Tag, tt.wantTag)
			}
			if len(op.PathParams) != tt.wantPathCount {
				t.Errorf("PathParams count = %d, want %d", len(op.PathParams), tt.wantPathCount)
			}
			if len(op.QueryParams) != tt.wantQueryCount {
				t.Errorf("QueryParams count = %d, want %d", len(op.QueryParams), tt.wantQueryCount)
			}
		})
	}
}

func TestExtractOperation_PathParams(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(op.PathParams) != 1 {
		t.Fatalf("expected 1 path param, got %d", len(op.PathParams))
	}

	param := op.PathParams[0]
	if param.Name != "team_id" {
		t.Errorf("path param name = %q, want %q", param.Name, "team_id")
	}
	if !param.Required {
		t.Error("path param should be required")
	}
}

func TestExtractOperation_QueryParams(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(op.QueryParams) != 1 {
		t.Fatalf("expected 1 query param, got %d", len(op.QueryParams))
	}

	param := op.QueryParams[0]
	if param.Name != "filter_keyword" {
		t.Errorf("query param name = %q, want %q", param.Name, "filter_keyword")
	}
	if param.Required {
		t.Error("query param should not be required")
	}
}

// T054: Filter parameter classification
func TestIsFilterParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"filter[keyword]", true},
		{"filter[me]", true},
		{"filter[status]", true},
		{"page[number]", false},
		{"page[size]", false},
		{"sort", false},
		{"include", false},
		{"name", false},
	}

	for _, tt := range tests {
		param := Parameter{Name: tt.name}
		if got := IsFilterParam(param); got != tt.want {
			t.Errorf("IsFilterParam(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsExcludedParam(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"page[number]", true},
		{"page[size]", true},
		{"fields[team]", true},
		{"sort", true},
		{"include", true},
		{"filter[keyword]", false},
		{"name", false},
	}

	for _, tt := range tests {
		param := Parameter{Name: tt.name}
		if got := IsExcludedParam(param); got != tt.want {
			t.Errorf("IsExcludedParam(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// T067/T069: ResponseTypeName extraction from $ref
func TestExtractOperation_ResponseTypeName(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}
	doc := &model.Model

	tests := []struct {
		name     string
		path     string
		method   string
		wantType string
	}{
		{
			name:     "team endpoint with $ref",
			path:     "/api/v2/team/{team_id}",
			method:   "get",
			wantType: "TeamResponse",
		},
		{
			name:     "simple endpoint without $ref",
			path:     "/api/v2/simple/{id}",
			method:   "get",
			wantType: "", // inline schema has no $ref
		},
		{
			name:     "list endpoint with $ref",
			path:     "/api/v2/team",
			method:   "get",
			wantType: "TeamsResponse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := ExtractOperation(doc, tt.path, tt.method)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if op.ResponseTypeName != tt.wantType {
				t.Errorf("ResponseTypeName = %q, want %q", op.ResponseTypeName, tt.wantType)
			}
		})
	}
}

// T054: List endpoint query params extraction
func TestExtractOperation_ListQueryParams(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(op.QueryParams) != 5 {
		t.Fatalf("expected 5 query params, got %d", len(op.QueryParams))
	}

	// Count filter vs excluded params
	var filterCount, excludedCount int
	for _, p := range op.QueryParams {
		if IsFilterParam(p) {
			filterCount++
		}
		if IsExcludedParam(p) {
			excludedCount++
		}
	}

	if filterCount != 2 {
		t.Errorf("expected 2 filter params, got %d", filterCount)
	}
	if excludedCount != 3 {
		t.Errorf("expected 3 excluded params (page[number], page[size], sort), got %d", excludedCount)
	}
}

func TestExtractOperation_ResponseSchemaProxy(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if op.ResponseSchemaProxy == nil {
		t.Fatal("expected non-nil ResponseSchemaProxy")
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building response schema: %v", err)
	}

	if schema.Properties == nil {
		t.Fatal("expected response schema to have properties")
	}

	_, hasData := schema.Properties.Get("data")
	if !hasData {
		t.Error("expected response schema to have 'data' property")
	}
}
