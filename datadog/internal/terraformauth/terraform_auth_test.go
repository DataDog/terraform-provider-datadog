package terraformauth

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTerraformWorkloadIdentityDiscovery(t *testing.T) {
	const org = "11111111-1111-4111-8111-111111111111"
	const otherOrg = "22222222-2222-4222-8222-222222222222"
	token := func(issuer interface{}, audience interface{}) string {
		payload, err := json.Marshal(map[string]interface{}{"iss": issuer, "aud": audience, "exp": 1})
		require.NoError(t, err)
		return "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(payload) + ".c2lnbmF0dXJl"
	}
	us := token("https://app.terraform.io", "datadog/"+org)
	eu := token("https://app.eu.terraform.io", []string{"datadog/" + org})
	tfe := token("https://terraform.example.com", "datadog/"+org)
	tests := []struct {
		name          string
		tagged        string
		untagged      string
		configuredOrg string
		wantError     bool
		wantOrg       string
		wantProof     string
	}{
		{name: "absent"},
		{name: "tagged", tagged: us, wantProof: us},
		{name: "matching configured organization", configuredOrg: org, tagged: us, wantProof: us},
		{name: "mismatched tagged organization is terminal", configuredOrg: otherOrg, tagged: us, untagged: token("https://app.terraform.io", "datadog/"+otherOrg), wantError: true},
		{name: "mismatched untagged organization", configuredOrg: otherOrg, untagged: us, wantError: true},
		{name: "configured organization without WIT allows fallback", configuredOrg: org},
		{name: "configured organization with unrelated WIT allows fallback", configuredOrg: org, tagged: token("https://app.terraform.io", "vault")},
		{name: "configured UUID comparison ignores letter case", configuredOrg: "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", tagged: token("https://app.terraform.io", "datadog/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), wantOrg: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", wantProof: token("https://app.terraform.io", "datadog/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")},
		{name: "tagged wins", tagged: us, untagged: eu, wantProof: us},
		{name: "untagged", untagged: us, wantProof: us},
		{name: "HCP Europe", tagged: eu, wantProof: eu},
		{name: "TFE CustomOIDC", tagged: tfe, wantProof: "oidc-" + tfe},
		{name: "exact issuer routing", tagged: token("https://app.terraform.io.example.com", "datadog/"+org), wantProof: "oidc-" + token("https://app.terraform.io.example.com", "datadog/"+org)},
		{name: "unrelated tagged audience", tagged: token("https://app.terraform.io", "vault"), untagged: us, wantProof: us},
		{name: "unrelated untagged audience", untagged: token("https://app.terraform.io", "vault")},
		{name: "organization comes from tagged audience", tagged: token("https://app.terraform.io", "datadog/"+otherOrg), untagged: us, wantOrg: otherOrg, wantProof: token("https://app.terraform.io", "datadog/"+otherOrg)},
		{name: "malformed tagged token", tagged: "not-a-jwt", untagged: us, wantProof: us},
		{name: "invalid encoding", tagged: "header.!.signature"},
		{name: "invalid JSON", tagged: "header.bm90LWpzb24.signature"},
		{name: "missing issuer reaches ETS", tagged: token(nil, "datadog/"+org), untagged: us, wantProof: token(nil, "datadog/"+org)},
		{name: "empty issuer reaches ETS", tagged: token("", "datadog/"+org), untagged: us, wantProof: token("", "datadog/"+org)},
		{name: "malformed issuer reaches ETS", tagged: token(42, "datadog/"+org), untagged: us, wantProof: token(42, "datadog/"+org)},
		{name: "missing audience", tagged: token("https://app.terraform.io", nil)},
		{name: "invalid audience type", tagged: token("https://app.terraform.io", 42)},
		{name: "empty organization", tagged: token("https://app.terraform.io", "datadog/")},
		{name: "invalid UUID", tagged: token("https://app.terraform.io", "datadog/not-a-uuid")},
		{name: "extra audience path", tagged: token("https://app.terraform.io", "datadog/"+org+"/extra")},
		{name: "ambiguous organizations", tagged: token("https://app.terraform.io", []string{"datadog/" + otherOrg, "datadog/" + org})},
		{name: "invalid Datadog audience before valid one", tagged: token("https://app.terraform.io", []string{"datadog/invalid", "datadog/" + org})},
		{name: "array with unrelated audience", tagged: token("https://app.terraform.io", []string{"vault", "datadog/" + org}), wantProof: token("https://app.terraform.io", []string{"vault", "datadog/" + org})},
		// exp is deliberately in the past: a Datadog WIT must reach ETS, whose
		// rejection is terminal, instead of falling back to another identity.
		{name: "expired WIT is selected for ETS validation", tagged: us, wantProof: us},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(TerraformWorkloadIdentityTokenEnv, tt.tagged)
			t.Setenv(TerraformWorkloadIdentityTokenFallbackEnv, tt.untagged)
			config, err := GetDelegatedTokenConfig(tt.configuredOrg)
			if tt.wantError {
				require.ErrorContains(t, err, "does not match configured org_uuid")
				require.Nil(t, config)
				require.NotContains(t, err.Error(), us)
				return
			}
			require.NoError(t, err)
			if tt.wantProof == "" {
				require.Nil(t, config)
				return
			}
			require.NotNil(t, config)
			wantOrg := tt.wantOrg
			if wantOrg == "" {
				wantOrg = org
			}
			require.Equal(t, wantOrg, config.OrgUUID)
			require.Equal(t, "terraform", config.Provider)
			require.Equal(t, tt.wantProof, config.ProviderAuth.(*terraformWorkloadIdentityAuth).proof)
		})
	}
}
