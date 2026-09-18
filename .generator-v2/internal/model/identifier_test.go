package model

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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

// The sdkClassName table is a Ginkgo spec in an otherwise stdlib-testing file:
// the package's bootstrap already runs both, and colocating it with the function
// it covers matters more than uniformity here (research §9 tracks migrating the
// rest of this file).
var _ = Describe("SdkClassName", func() {
	// Every expectation is the output datadog-api-client-go's utils.class_name
	// produces for the same input, so a maintainer can verify against the Python.
	DescribeTable("strips non-alphanumerics and appends Api without re-capitalizing",
		func(tag, want string) { Expect(SdkClassName(tag)).To(Equal(want)) },
		Entry("a single word gains only the suffix", "Incidents", "IncidentsApi"),
		Entry("a PascalCase multi-word tag loses its spaces", "Org Groups", "OrgGroupsApi"),
		Entry("an acronym is preserved, not folded", "APM", "APMApi"),
		Entry("a lower-cased word stays lower-cased", "org groups", "orggroupsApi"),
		Entry("punctuation is stripped without creating a capital",
			"cloud-cost.management", "cloudcostmanagementApi"),
		Entry("digits survive", "AWS Logs Integration v2", "AWSLogsIntegrationv2Api"),
		Entry("an empty tag still yields the suffix", "", "Api"),
	)
})
