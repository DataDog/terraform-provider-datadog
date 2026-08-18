package model

import "testing"

// TestSdkName locks in the naive PascalCase translation: split on underscores,
// Title-case each segment (first rune upper, the rest lower), and never
// special-case acronyms. The "no acronym uppercasing" cases (url, uuid, ID)
// are the ones that keep generated code compiling against the SDK.
func TestSdkName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// Basic cases
		{"two snake segments", "org_id", "OrgId"},
		{"api key", "api_key", "ApiKey"},
		{"single word", "url", "Url"},
		{"compound word", "http_endpoint", "HttpEndpoint"},
		{"acronym stays naive", "uuid", "Uuid"},

		// Mixed cases: input casing is normalized, not preserved.
		{"upper acronym segment lowercased", "HTTP_endpoint", "HttpEndpoint"},
		{"mixed segments normalized", "Org_ID", "OrgId"},
		{"trailing digits preserved", "o_auth2", "OAuth2"},
		{"camelCase split on case boundary", "isURL", "IsUrl"},

		// Edge cases.
		{"empty input", "", ""},
		{"stray underscores skipped", "__org__id__", "OrgId"},
		{"leading underscore", "_id", "Id"},

		// Generator parity: snake_case folds \W (spaces, hyphens) to
		// underscores, and acronyms stay naive
		{"whitespace folded", "foo bar", "FooBar"},
		{"hyphen folded", "foo-bar", "FooBar"},
		{"all-caps stays naive", "DASHBOARD_ID", "DashboardId"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SdkName(tt.in); got != tt.want {
				t.Errorf("SdkName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestSdkNameIdempotent asserts SdkName is idempotent: applying it to an
// already-converted name is a no-op, SdkName(SdkName(x)) == SdkName(x). This
// guards the case where a name reaches the translator twice, and covers both
// snake_case inputs and their PascalCase outputs (the latter must be fixed points).
func TestSdkNameIdempotent(t *testing.T) {
	inputs := []string{
		// snake_case inputs
		"org_id", "api_key", "url", "http_endpoint", "uuid", "team_id", "o_auth2",
		// already-cased / mixed inputs
		"HTTP_endpoint", "Org_ID", "isURL", "v2_api",
		// already-PascalCase outputs — these must map to themselves
		"OrgId", "ApiKey", "Url", "HttpEndpoint", "Uuid", "OAuth2", "IsUrl", "HttpServer",
		"", "_id",
	}

	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			once := SdkName(in)
			twice := SdkName(once)
			if once != twice {
				t.Errorf("SdkName not idempotent: SdkName(%q) = %q, SdkName(%q) = %q", in, once, once, twice)
			}
		})
	}
}
