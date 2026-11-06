package cloudauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

const (
	// GitHubProvider is the delegated-token provider name for GitHub Actions OIDC.
	GitHubProvider = "github"

	actionsIDTokenRequestTokenEnv = "ACTIONS_ID_TOKEN_REQUEST_TOKEN" // #nosec G101 -- environment variable name
	actionsIDTokenRequestURLEnv   = "ACTIONS_ID_TOKEN_REQUEST_URL"
	gitHubAudiencePrefix          = "datadog/"
	maximumTokenResponseSize      = 1 << 20
)

// GitHubConfig contains the Terraform provider settings that influence GitHub
// Actions delegated authentication.
type GitHubConfig struct {
	HTTPClient *http.Client
}

type gitHubProvider struct {
	httpClient *http.Client
}

// NewGitHubProvider creates a delegated-token provider that exchanges a GitHub
// Actions OIDC token for a Datadog access token. The OIDC token is requested
// when authentication runs, from ACTIONS_ID_TOKEN_REQUEST_URL, using the
// audience datadog/<org_uuid>.
func NewGitHubProvider(config GitHubConfig) datadog.DelegatedTokenProvider {
	client := config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &gitHubProvider{httpClient: client}
}

// Authenticate requests a GitHub Actions OIDC token and exchanges it at the
// Datadog delegated-token endpoint. The OIDC token is the delegated proof and
// is not retained beyond the exchange.
func (provider *gitHubProvider) Authenticate(ctx context.Context, config *datadog.DelegatedTokenConfig) (*datadog.DelegatedTokenCredentials, error) {
	if config == nil || config.OrgUUID == "" {
		return nil, errors.New("missing org UUID in delegated token configuration")
	}

	proof, err := provider.requestOIDCToken(ctx, config.OrgUUID)
	if err != nil {
		return nil, err
	}
	tokenURL, err := datadog.GetDelegatedTokenUrl(ctx)
	if err != nil {
		return nil, redactSecrets(fmt.Errorf("resolving Datadog delegated token URL: %w", err), proof)
	}
	credentials, err := provider.exchangeDelegatedToken(ctx, tokenURL, config.OrgUUID, proof)
	if err != nil {
		return nil, redactSecrets(fmt.Errorf("exchanging GitHub OIDC token for Datadog delegated token: %w", err), proof)
	}
	return credentials, nil
}

func (provider *gitHubProvider) requestOIDCToken(ctx context.Context, orgUUID string) (string, error) {
	actionsToken := os.Getenv(actionsIDTokenRequestTokenEnv)
	actionsURL := os.Getenv(actionsIDTokenRequestURLEnv)
	if actionsToken == "" || actionsURL == "" {
		return "", errors.New("GitHub Actions OIDC token request is not configured; grant id-token: write so ACTIONS_ID_TOKEN_REQUEST_TOKEN and ACTIONS_ID_TOKEN_REQUEST_URL are set")
	}

	audience := gitHubAudiencePrefix + orgUUID
	endpoint, err := gitHubOIDCRequestURL(actionsURL, audience)
	if err != nil {
		return "", err
	}
	log.Printf("[INFO] Datadog GitHub delegated auth requesting an OIDC token for audience %q", audience)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", redactSecrets(fmt.Errorf("creating GitHub OIDC token request: %w", err), actionsToken, actionsURL)
	}
	request.Header.Set("Authorization", "Bearer "+actionsToken)
	request.Header.Set("Accept", "application/json")

	response, err := provider.httpClient.Do(request)
	if err != nil {
		return "", redactSecrets(fmt.Errorf("requesting GitHub OIDC token: %w", err), actionsToken, actionsURL)
	}
	defer response.Body.Close()

	body, err := readLimitedBody(response.Body)
	if err != nil {
		return "", redactSecrets(fmt.Errorf("reading GitHub OIDC token response: %w", err), actionsToken, actionsURL)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub OIDC token endpoint returned %s", response.Status)
	}

	var payload struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.Value == "" {
		return "", errors.New("GitHub OIDC token response did not include a token")
	}
	return payload.Value, nil
}

func gitHubOIDCRequestURL(requestURL, audience string) (string, error) {
	parsed, err := url.Parse(requestURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("ACTIONS_ID_TOKEN_REQUEST_URL must be an absolute URL")
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("ACTIONS_ID_TOKEN_REQUEST_URL must use http or https")
	}
	query := parsed.Query()
	query.Set("audience", audience)
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String(), nil
}

func (provider *gitHubProvider) exchangeDelegatedToken(ctx context.Context, tokenURL, orgUUID, proof string) (*datadog.DelegatedTokenCredentials, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating Datadog delegated token request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Delegated "+proof)

	response, err := provider.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting Datadog delegated token: %w", err)
	}
	defer response.Body.Close()

	body, err := readLimitedBody(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Datadog delegated token response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Datadog delegated token endpoint returned %s", response.Status)
	}

	credentials, err := datadog.ParseDelegatedTokenResponse(body, orgUUID, proof)
	if err != nil {
		// The core parser may include the response body in its error. Never expose
		// token material from a malformed response in logs or diagnostics.
		return nil, errors.New("parsing Datadog delegated token response: invalid response")
	}
	if credentials.DelegatedToken == "" {
		return nil, errors.New("Datadog delegated token endpoint returned an empty token")
	}
	return credentials, nil
}

func readLimitedBody(body io.Reader) ([]byte, error) {
	limited, err := io.ReadAll(io.LimitReader(body, maximumTokenResponseSize+1))
	if err != nil {
		return nil, err
	}
	if len(limited) > maximumTokenResponseSize {
		return nil, errors.New("response exceeds 1 MiB")
	}
	return limited, nil
}

func redactSecrets(err error, secrets ...string) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	redacted := message
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		redacted = strings.ReplaceAll(redacted, secret, "[REDACTED]")
	}
	if redacted == message {
		return err
	}
	return errors.New(redacted)
}
