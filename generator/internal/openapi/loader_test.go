package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSpec(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{
			name: "valid minimal spec",
			path: "../../testdata/minimal.yaml",
		},
		{
			name:    "missing file",
			path:    "/nonexistent/spec.yaml",
			wantErr: "reading spec file",
		},
		{
			name:    "invalid YAML",
			path:    "invalid.yaml", // created in test
			wantErr: "parsing spec file",
		},
	}

	// Create invalid YAML fixture
	dir := t.TempDir()
	invalidPath := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte("not: [valid: openapi"), 0644); err != nil {
		t.Fatalf("write invalid fixture: %v", err)
	}
	tests[2].path = invalidPath

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := LoadSpec(tt.path)

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
			if model == nil {
				t.Fatal("expected non-nil model")
			}
			if model.Model.Info == nil {
				t.Fatal("expected non-nil info")
			}
			if model.Model.Info.Title != "Test API" {
				t.Errorf("expected title %q, got %q", "Test API", model.Model.Info.Title)
			}
		})
	}
}

func TestLoadSpec_PathCount(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := model.Model.Paths
	if paths == nil {
		t.Fatal("expected non-nil paths")
	}

	count := paths.PathItems.Len()
	if count != 5 {
		t.Errorf("expected 5 paths, got %d", count)
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
