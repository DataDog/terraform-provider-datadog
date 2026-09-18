package fwprovider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFrameworkAWSConfiguration(t *testing.T) {
	for _, mode := range []string{"profile", "explicit", "explicit-temporary", "partial-access", "partial-secret", "partial-token", "api-key"} {
		t.Run(mode, func(t *testing.T) {
			for _, key := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_WEB_IDENTITY_TOKEN_FILE", "AWS_ROLE_ARN", "AWS_REGION", "AWS_DEFAULT_REGION", "AWS_DEFAULT_PROFILE", "AWS_ENDPOINT_URL", "AWS_ENDPOINT_URL_STS", "AWS_CA_BUNDLE"} {
				t.Setenv(key, "")
			}
			t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
			t.Setenv("AWS_PROFILE", "wif-test")
			dir := t.TempDir()
			path := filepath.Join(dir, "credentials")
			if err := os.WriteFile(path, []byte("[wif-test]\naws_access_key_id = profile-key\naws_secret_access_key = profile-secret\n"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)
			t.Setenv("AWS_CONFIG_FILE", filepath.Join(dir, "config"))
			var exchanges atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				exchanges.Add(1)
				if r.URL.Path != "/api/v2/delegated-token" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				parts := strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Delegated "), "|")
				if len(parts) != 4 {
					t.Error("missing signed AWS proof")
					w.WriteHeader(400)
					return
				}
				raw, err := base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					t.Error(err)
					return
				}
				headers := http.Header{}
				if err := json.Unmarshal(raw, &headers); err != nil {
					t.Error(err)
					return
				}
				key := "profile-key"
				if strings.HasPrefix(mode, "explicit") {
					key = "explicit-key"
				}
				if !strings.Contains(headers.Get("Authorization"), "Credential="+key+"/") {
					t.Error("wrong AWS credential source")
				}
				wantSessionToken := ""
				if mode == "explicit-temporary" {
					wantSessionToken = "explicit-token"
				}
				if headers.Get("X-Amz-Security-Token") != wantSessionToken {
					t.Error("wrong AWS session token")
				}
				fmt.Fprint(w, `{"data":{"attributes":{"access_token":"test-token"}}}`)
			}))
			defer server.Close()
			config := &ProviderSchema{CloudProviderType: types.StringValue("aws"), OrgUuid: types.StringValue("test-org"), ApiUrl: types.StringValue(server.URL), Validate: types.StringValue("false")}
			switch mode {
			case "explicit", "explicit-temporary":
				config.AWSAccessKeyId = types.StringValue("explicit-key")
				config.AWSSecretAccessKey = types.StringValue("explicit-secret")
				if mode == "explicit-temporary" {
					config.AWSSessionToken = types.StringValue("explicit-token")
				}
			case "partial-access":
				config.AWSAccessKeyId = types.StringValue("incomplete")
			case "partial-secret":
				config.AWSSecretAccessKey = types.StringValue("incomplete")
			case "partial-token":
				config.AWSSessionToken = types.StringValue("incomplete")
			case "api-key":
				config.CloudProviderType = types.StringNull()
				config.ApiKey = types.StringValue("test-api-key")
				config.AppKey = types.StringValue("test-app-key")
			}
			p := New().(*FrameworkProvider)
			diags := defaultConfigureFunc(p, &provider.ConfigureRequest{}, config)
			if strings.HasPrefix(mode, "partial") {
				if !diags.HasError() {
					t.Fatal("partial credentials accepted")
				}
				return
			}
			if diags.HasError() {
				t.Fatalf("configure: %v", diags)
			}
			if exchanges.Load() != 0 {
				t.Fatal("configure performed eager authentication")
			}
			client := p.DatadogApiInstances.HttpClient
			if mode == "api-key" {
				if client.GetConfig().DelegatedTokenConfig != nil {
					t.Fatal("AWS enabled for API key auth")
				}
				if p.Auth.Value(datadog.ContextAWSVariables) != nil {
					t.Fatal("AWS credentials context set for API key auth")
				}
				keys, ok := p.Auth.Value(datadog.ContextAPIKeys).(map[string]datadog.APIKey)
				if !ok || keys["apiKeyAuth"].Key != "test-api-key" || keys["appKeyAuth"].Key != "test-app-key" {
					t.Fatal("API key context changed")
				}
				return
			}
			credentials, ok := p.Auth.Value(datadog.ContextAWSVariables).(map[string]string)
			if !ok || len(credentials) != 3 {
				t.Fatal("AWS credentials missing from standard authentication context")
			}
			wantAccessKey, wantSecret, wantToken := "", "", ""
			if strings.HasPrefix(mode, "explicit") {
				wantAccessKey, wantSecret = "explicit-key", "explicit-secret"
			}
			if mode == "explicit-temporary" {
				wantToken = "explicit-token"
			}
			if credentials[datadog.AWSAccessKeyIdName] != wantAccessKey ||
				credentials[datadog.AWSSecretAccessKeyName] != wantSecret ||
				credentials[datadog.AWSSessionTokenName] != wantToken {
				t.Fatal("AWS credentials context does not match provider configuration")
			}
			if _, ok := p.Auth.Value(datadog.ContextDelegatedToken).(*datadog.DelegatedTokenCredentials); !ok {
				t.Fatal("delegated token context missing")
			}
			token, err := client.GetDelegatedToken(p.Auth)
			if err != nil {
				t.Fatal(err)
			}
			if token.DelegatedToken != "test-token" || exchanges.Load() != 1 {
				t.Fatal("AWS adapter did not authenticate")
			}
		})
	}
}
