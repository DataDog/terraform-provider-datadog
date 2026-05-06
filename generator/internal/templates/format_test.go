package templates

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PascalCase", "pascal_case"},
		{"camelCase", "camel_case"},
		{"snake_case", "snake_case"},
		{"APIKey", "api_key"},
		{"teamID", "team_id"},
		{"GetTeamURL", "get_team_url"},
		{"HTTPSEnabled", "https_enabled"},
		{"simpleWord", "simple_word"},
		{"A", "a"},
		{"AB", "ab"},
		{"", ""},
		{"already_snake", "already_snake"},
		{"UserCount", "user_count"},
		{"linkCount", "link_count"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"team_id", "TeamID"},
		{"user_count", "UserCount"},
		{"api_key", "APIKey"},
		{"simple", "Simple"},
		{"snake_case_name", "SnakeCaseName"},
		{"already_pascal", "AlreadyPascal"},
		{"url", "URL"},
		{"http_url", "HTTPURL"},
		{"", ""},
		{"id", "ID"},
		{"camelCase", "CamelCase"},
		{"Cloud Cost Management", "CloudCostManagement"},
		{"APM Retention Filters", "ApmRetentionFilters"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"team_id", "teamID"},
		{"user_count", "userCount"},
		{"PascalCase", "pascalCase"},
		{"APIKey", "apiKey"},
		{"simple", "simple"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToCamelCase(tt.input)
			if got != tt.want {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeGoKeyword(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"type", "type_"},
		{"range", "range_"},
		{"map", "map_"},
		{"var", "var_"},
		{"name", "name"},
		{"count", "count"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EscapeGoKeyword(tt.input)
			if got != tt.want {
				t.Errorf("EscapeGoKeyword(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSDKPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"org_id", "OrgId"},
		{"team_url", "TeamUrl"},
		{"api_key", "ApiKey"},
		{"name", "Name"},
		{"description", "Description"},
		{"hidden_modules", "HiddenModules"},
		{"user_id", "UserId"},
		{"http", "Http"},
		{"team_api", "TeamApi"},
		{"json_config", "JsonConfig"},
		{"", ""},
		{"team_ip", "TeamIp"},
		{"my_uri", "MyUri"},
		{"ssh_key", "SshKey"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToSDKPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("ToSDKPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeDescription(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple",
			input: "A simple description.",
			want:  "A simple description.",
		},
		{
			name:  "multi-line",
			input: "Line one.\nLine two.\nLine three.",
			want:  "Line one. Line two. Line three.",
		},
		{
			name:  "quotes",
			input: `Has "quotes" inside.`,
			want:  `Has \"quotes\" inside.`,
		},
		{
			name:  "backticks",
			input: "Has `backticks` inside.",
			want:  "Has 'backticks' inside.",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "whitespace lines",
			input: "  First  \n  \n  Last  ",
			want:  "First Last",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeDescription(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}
