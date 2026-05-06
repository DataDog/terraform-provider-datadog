package openapi

import (
	"testing"
)

func TestIsJSONAPIEnvelope(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		method string
		want   bool
	}{
		{
			name:   "JSON:API team endpoint",
			path:   "/api/v2/team/{team_id}",
			method: "get",
			want:   true,
		},
		{
			name:   "flat simple endpoint",
			path:   "/api/v2/simple/{id}",
			method: "get",
			want:   false,
		},
		{
			name:   "JSON:API complex endpoint",
			path:   "/api/v2/complex/{id}",
			method: "get",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := ExtractOperation(&model.Model, tt.path, tt.method)
			if err != nil {
				t.Fatalf("extracting operation: %v", err)
			}

			schema, err := op.ResponseSchemaProxy.BuildSchema()
			if err != nil {
				t.Fatalf("building schema: %v", err)
			}

			got := IsJSONAPIEnvelope(schema)
			if got != tt.want {
				t.Errorf("IsJSONAPIEnvelope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnwrapJSONAPI(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, typeName, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attrsProxy == nil {
		t.Fatal("expected non-nil attributes proxy")
	}

	if typeName != "team" {
		t.Errorf("typeName = %q, want %q", typeName, "team")
	}

	// Verify the attributes schema has the expected properties
	attrsSchema, err := attrsProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building attributes schema: %v", err)
	}

	if attrsSchema.Properties == nil {
		t.Fatal("attributes schema has no properties")
	}

	expectedFields := []string{"name", "handle", "description", "user_count", "link_count", "is_active", "status"}
	for _, field := range expectedFields {
		if _, ok := attrsSchema.Properties.Get(field); !ok {
			t.Errorf("expected attributes schema to have %q property", field)
		}
	}
}

func TestUnwrapJSONAPI_NonEnvelope(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/simple/{id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	_, _, err = UnwrapJSONAPI(schema)
	if err == nil {
		t.Fatal("expected error when unwrapping non-JSON:API schema")
	}
}

func TestIsJSONAPIEnvelope_NilSchema(t *testing.T) {
	if IsJSONAPIEnvelope(nil) {
		t.Error("expected false for nil schema")
	}
}
