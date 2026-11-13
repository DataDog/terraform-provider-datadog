package fwprovider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/cloudauth"
)

func TestFrameworkGitHubConfiguration(t *testing.T) {
	const orgUUID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"

	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "actions-token")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/token":
			if r.URL.Query().Get("audience") != "datadog/"+orgUUID {
				t.Errorf("audience = %q", r.URL.Query().Get("audience"))
			}
			fmt.Fprint(w, `{"value":"header.payload.signature"}`)
		case "/api/v2/delegated-token":
			if r.Header.Get("Authorization") != "Delegated header.payload.signature" {
				t.Errorf("delegated authorization = %q", r.Header.Get("Authorization"))
			}
			fmt.Fprint(w, `{"data":{"attributes":{"access_token":"test-token","expires":"4102444800"}}}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", server.URL+"/token")

	config := &ProviderSchema{
		CloudProviderType: types.StringValue("github"),
		OrgUuid:           types.StringValue(orgUUID),
		ApiUrl:            types.StringValue(server.URL),
		Validate:          types.StringValue("false"),
	}
	p := New().(*FrameworkProvider)
	diags := defaultConfigureFunc(p, &provider.ConfigureRequest{}, config)
	if diags.HasError() {
		t.Fatalf("configure: %v", diags)
	}
	if calls.Load() != 0 {
		t.Fatal("configure performed eager authentication")
	}
	client := p.DatadogApiInstances.HttpClient
	delegated := client.GetConfig().DelegatedTokenConfig
	if delegated == nil || delegated.Provider != cloudauth.GitHubProvider || delegated.OrgUUID != orgUUID {
		t.Fatalf("delegated config = %#v", delegated)
	}
	if _, ok := p.Auth.Value(datadog.ContextDelegatedToken).(*datadog.DelegatedTokenCredentials); !ok {
		t.Fatal("delegated token context missing")
	}
	if p.Auth.Value(datadog.ContextAWSVariables) != nil {
		t.Fatal("AWS credentials context set for GitHub auth")
	}

	token, err := client.GetDelegatedToken(p.Auth)
	if err != nil {
		t.Fatal(err)
	}
	if token.DelegatedToken != "test-token" || calls.Load() != 2 {
		t.Fatalf("token = %q, calls = %d", token.DelegatedToken, calls.Load())
	}
}

func TestFrameworkGitHubConfigurationRejectsUnknownProvider(t *testing.T) {
	config := &ProviderSchema{
		CloudProviderType: types.StringValue("azure"),
		OrgUuid:           types.StringValue("org"),
		Validate:          types.StringValue("false"),
	}
	p := New().(*FrameworkProvider)
	diags := defaultConfigureFunc(p, &provider.ConfigureRequest{}, config)
	if !diags.HasError() {
		t.Fatal("unknown cloud_provider_type accepted")
	}
}
