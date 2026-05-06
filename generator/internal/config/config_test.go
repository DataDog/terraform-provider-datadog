package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid config",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
`,
		},
		{
			name: "valid config with default spec",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    read:
      path: /api/v2/team/{team_id}
      method: get
`,
		},
		{
			name: "empty specs map",
			yaml: `
specs: {}
datasources:
  team:
    read:
      path: /api/v2/team/{team_id}
      method: get
`,
			wantErr: "specs map must not be empty",
		},
		{
			name: "missing specs key",
			yaml: `
datasources:
  team:
    read:
      path: /api/v2/team/{team_id}
      method: get
`,
			wantErr: "specs map must not be empty",
		},
		{
			name: "missing read path",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      method: get
`,
			wantErr: `datasource "team": read.path is required`,
		},
		{
			name: "missing read method",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
`,
			wantErr: `datasource "team": read.method is required`,
		},
		{
			name: "invalid method",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: INVALID
`,
			wantErr: `not a valid HTTP method`,
		},
		{
			name: "unknown spec reference",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v99
    read:
      path: /api/v2/team/{team_id}
      method: get
`,
			wantErr: `spec "v99" not found`,
		},
		{
			name:    "empty file",
			yaml:    "",
			wantErr: "specs map must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("write temp config: %v", err)
			}

			cfg, err := LoadConfig(path)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg == nil {
				t.Fatal("expected non-nil config")
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestResolveSpec(t *testing.T) {
	cfg := &Config{
		Specs: map[string]SpecConfig{
			"v1": {Path: "v1.yaml"},
			"v2": {Path: "v2.yaml"},
		},
	}

	tests := []struct {
		name string
		ds   DataSourceConfig
		want string
	}{
		{
			name: "explicit spec",
			ds:   DataSourceConfig{Spec: "v2"},
			want: "v2",
		},
		{
			name: "default spec (first alphabetically)",
			ds:   DataSourceConfig{},
			want: "v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.ResolveSpec(tt.ds)
			if got != tt.want {
				t.Errorf("ResolveSpec() = %q, want %q", got, tt.want)
			}
		})
	}
}

// T052: Config validation for list block
func TestLoadConfig_ListValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "valid list config",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
    list:
      path: /api/v2/team
      method: get
`,
		},
		{
			name: "list missing path",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
    list:
      method: get
`,
			wantErr: "list.path is required",
		},
		{
			name: "list missing method",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
    list:
      path: /api/v2/team
`,
			wantErr: "list.method is required",
		},
		{
			name: "list invalid method",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
    list:
      path: /api/v2/team
      method: INVALID
`,
			wantErr: "not a valid HTTP method",
		},
		{
			name: "no list block is valid",
			yaml: `
specs:
  v2:
    path: testdata/minimal.yaml
datasources:
  team:
    spec: v2
    read:
      path: /api/v2/team/{team_id}
      method: get
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("write temp config: %v", err)
			}

			_, err := LoadConfig(path)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
