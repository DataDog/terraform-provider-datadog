package terraformauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/google/uuid"
)

const (
	// TerraformWorkloadIdentityTokenEnv is the Datadog-specific token injected by HCP Terraform/TFE.
	TerraformWorkloadIdentityTokenEnv = "TFC_WORKLOAD_IDENTITY_TOKEN_DATADOG" // #nosec G101 -- environment variable name
	// TerraformWorkloadIdentityTokenFallbackEnv is the untagged Terraform workload identity token.
	TerraformWorkloadIdentityTokenFallbackEnv = "TFC_WORKLOAD_IDENTITY_TOKEN" // #nosec G101 -- environment variable name
)

// GetDelegatedTokenConfig discovers a Terraform workload identity token
// addressed to a Datadog organization. A nil configuration without an error
// lets the caller retain its existing authentication selection. Claims are inspected only for selection and
// routing. A configured organization must match the selected token's audience.
// ETS is responsible for signature, issuer, expiry and mapping validation.
func GetDelegatedTokenConfig(orgUUID string) (*datadog.DelegatedTokenConfig, error) {
	for _, name := range []string{TerraformWorkloadIdentityTokenEnv, TerraformWorkloadIdentityTokenFallbackEnv} {
		token := os.Getenv(name)
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			continue
		}
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			continue
		}
		var claims struct {
			Issuer   json.RawMessage `json:"iss"`
			Audience json.RawMessage `json:"aud"`
		}
		if err := json.Unmarshal(payload, &claims); err != nil {
			continue
		}
		tokenOrgUUID := terraformAudienceOrgUUID(claims.Audience)
		if tokenOrgUUID == "" {
			continue
		}
		if orgUUID != "" && !strings.EqualFold(orgUUID, tokenOrgUUID) {
			return nil, fmt.Errorf("%s audience %q does not match configured org_uuid %q", name, "datadog/"+tokenOrgUUID, orgUUID)
		}

		// Inspect the issuer only to route TFE through CustomOIDC. Missing or
		// malformed issuers still reach ETS and must not cause credential fallback.
		proof := token
		var issuer string
		if err := json.Unmarshal(claims.Issuer, &issuer); err == nil && issuer != "" &&
			issuer != "https://app.terraform.io" && issuer != "https://app.eu.terraform.io" {
			proof = "oidc-" + token
		}
		return &datadog.DelegatedTokenConfig{
			OrgUUID:      tokenOrgUUID,
			Provider:     "terraform",
			ProviderAuth: &terraformWorkloadIdentityAuth{proof: proof},
		}, nil
	}
	return nil, nil
}

// terraformAudienceOrgUUID accepts the JWT string and array forms of aud. It
// requires an unambiguous Datadog organization so provider selection cannot
// disagree with ETS about which audience determines the organization.
func terraformAudienceOrgUUID(raw json.RawMessage) string {
	var audiences []string
	var audience string
	if err := json.Unmarshal(raw, &audience); err == nil {
		audiences = []string{audience}
	} else if err := json.Unmarshal(raw, &audiences); err != nil {
		return ""
	}
	var orgUUID string
	for _, audience := range audiences {
		if !strings.HasPrefix(audience, "datadog/") {
			continue
		}
		candidate := strings.TrimPrefix(audience, "datadog/")
		parsed, err := uuid.Parse(candidate)
		if err != nil || len(candidate) != 36 || !strings.EqualFold(candidate, parsed.String()) {
			return ""
		}
		if orgUUID != "" && orgUUID != candidate {
			return ""
		}
		orgUUID = candidate
	}
	return orgUUID
}

// terraformWorkloadIdentityAuth supplies the same SDK exchange/refresh interface
// as AWSAuth. Retain the selected proof for the lifetime of this provider: an ETS
// rejection must never trigger a different token or authentication method.
type terraformWorkloadIdentityAuth struct {
	proof string
}

func (a *terraformWorkloadIdentityAuth) Authenticate(ctx context.Context, config *datadog.DelegatedTokenConfig) (*datadog.DelegatedTokenCredentials, error) {
	credentials, err := datadog.GetDelegatedToken(ctx, config.OrgUUID, a.proof)
	if err != nil {
		// SDK response-decoding errors can include the response body. Avoid
		// exposing credentials from a malformed exchange response in diagnostics.
		if strings.HasPrefix(err.Error(), "failed to parse token response:") {
			return nil, errors.New("terraform workload identity token exchange returned an invalid response")
		}
		return nil, fmt.Errorf("terraform workload identity token exchange failed: %s", strings.ReplaceAll(err.Error(), a.proof, "[REDACTED]"))
	}
	if credentials == nil || credentials.DelegatedToken == "" {
		return nil, errors.New("terraform workload identity token exchange returned an empty access token")
	}
	return credentials, nil
}
