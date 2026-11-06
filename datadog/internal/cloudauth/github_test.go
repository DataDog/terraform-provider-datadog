package cloudauth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

func TestGitHubProviderAuthenticate(t *testing.T) {
	const (
		orgUUID      = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
		actionsToken = "actions-request-token"
		oidcToken    = "header.payload.signature"
		accessToken  = "dd-access-token"
	)

	var gotAudience string
	var gotActionsAuth string
	var gotDelegatedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			gotAudience = r.URL.Query().Get("audience")
			gotActionsAuth = r.Header.Get("Authorization")
			if r.Method != http.MethodGet {
				t.Errorf("OIDC method = %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"value":"`+oidcToken+`","count":1}`)
		case "/api/v2/delegated-token":
			gotDelegatedAuth = r.Header.Get("Authorization")
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("content type = %s", r.Header.Get("Content-Type"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data":{"attributes":{"access_token":"`+accessToken+`","expires":"4102444800"}}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv(actionsIDTokenRequestTokenEnv, actionsToken)
	t.Setenv(actionsIDTokenRequestURLEnv, server.URL+"/token?existing=1")

	provider := NewGitHubProvider(GitHubConfig{HTTPClient: server.Client()})
	credentials, err := provider.Authenticate(delegatedTokenContext(server.URL), &datadog.DelegatedTokenConfig{
		OrgUUID:  orgUUID,
		Provider: GitHubProvider,
	})
	if err != nil {
		t.Fatal(err)
	}
	if credentials.DelegatedToken != accessToken {
		t.Fatalf("access token = %q", credentials.DelegatedToken)
	}
	if credentials.OrgUUID != orgUUID {
		t.Fatalf("org uuid = %q", credentials.OrgUUID)
	}
	if gotAudience != "datadog/"+orgUUID {
		t.Fatalf("audience = %q", gotAudience)
	}
	if gotActionsAuth != "Bearer "+actionsToken {
		t.Fatalf("OIDC authorization = %q", gotActionsAuth)
	}
	if gotDelegatedAuth != "Delegated "+oidcToken {
		t.Fatalf("delegated authorization = %q", gotDelegatedAuth)
	}
}

func TestGitHubProviderAuthenticateRejectsMissingRequestConfig(t *testing.T) {
	t.Setenv(actionsIDTokenRequestTokenEnv, "")
	t.Setenv(actionsIDTokenRequestURLEnv, "")

	provider := NewGitHubProvider(GitHubConfig{})
	_, err := provider.Authenticate(context.Background(), &datadog.DelegatedTokenConfig{OrgUUID: "org"})
	if err == nil || !strings.Contains(err.Error(), "id-token: write") {
		t.Fatalf("error = %v", err)
	}
}

func TestGitHubProviderAuthenticateOmitsTokenMaterialFromErrors(t *testing.T) {
	const (
		actionsToken = "actions-request-token"
		oidcToken    = "header.payload.signature"
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"value":"`+oidcToken+`"}`)
		case "/api/v2/delegated-token":
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, "denied "+oidcToken+" "+actionsToken)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv(actionsIDTokenRequestTokenEnv, actionsToken)
	t.Setenv(actionsIDTokenRequestURLEnv, server.URL+"/token")

	provider := NewGitHubProvider(GitHubConfig{HTTPClient: server.Client()})
	_, err := provider.Authenticate(delegatedTokenContext(server.URL), &datadog.DelegatedTokenConfig{OrgUUID: "org"})
	if err == nil {
		t.Fatal("expected exchange error")
	}
	message := err.Error()
	if strings.Contains(message, oidcToken) || strings.Contains(message, actionsToken) || strings.Contains(message, "denied") {
		t.Fatalf("error leaked token material: %s", message)
	}
	if !strings.Contains(message, "403") {
		t.Fatalf("error = %s", message)
	}
}

func TestGitHubProviderAuthenticateRejectsOIDCEndpointFailure(t *testing.T) {
	const actionsToken = "actions-request-token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "bad "+actionsToken)
	}))
	defer server.Close()

	t.Setenv(actionsIDTokenRequestTokenEnv, actionsToken)
	t.Setenv(actionsIDTokenRequestURLEnv, server.URL+"/token")

	provider := NewGitHubProvider(GitHubConfig{HTTPClient: server.Client()})
	_, err := provider.Authenticate(delegatedTokenContext(server.URL), &datadog.DelegatedTokenConfig{OrgUUID: "org"})
	if err == nil {
		t.Fatal("expected OIDC error")
	}
	if strings.Contains(err.Error(), actionsToken) || strings.Contains(err.Error(), "bad ") {
		t.Fatalf("error leaked token material: %s", err)
	}
}

func TestGitHubOIDCRequestURL(t *testing.T) {
	endpoint, err := gitHubOIDCRequestURL("https://token.actions.githubusercontent.example/token?api=1", "datadog/org")
	if err != nil {
		t.Fatal(err)
	}
	parsedURL, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	if parsedURL.URL.Query().Get("audience") != "datadog/org" || parsedURL.URL.Query().Get("api") != "1" {
		t.Fatalf("url = %s", endpoint)
	}

	if _, err := gitHubOIDCRequestURL("/relative", "datadog/org"); err == nil {
		t.Fatal("relative URL accepted")
	}
	if _, err := gitHubOIDCRequestURL("file:///tmp/token", "datadog/org"); err == nil {
		t.Fatal("file URL accepted")
	}
}

func delegatedTokenContext(apiURL string) context.Context {
	parsed, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, datadog.ContextServerIndex, 1)
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{
		"name":     parsed.URL.Host,
		"protocol": parsed.URL.Scheme,
	})
	return ctx
}
