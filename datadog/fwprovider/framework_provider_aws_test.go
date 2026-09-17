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

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFrameworkAWSConfiguration(t *testing.T) {
	for _, mode := range []string{"profile", "explicit", "partial", "api-key"} {
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
				if mode == "explicit" {
					key = "explicit-key"
				}
				if !strings.Contains(headers.Get("Authorization"), "Credential="+key+"/") {
					t.Error("wrong AWS credential source")
				}
				fmt.Fprint(w, `{"data":{"attributes":{"access_token":"test-token"}}}`)
			}))
			defer server.Close()
			config := &ProviderSchema{CloudProviderType: types.StringValue("aws"), OrgUuid: types.StringValue("test-org"), ApiUrl: types.StringValue(server.URL), Validate: types.StringValue("false")}
			switch mode {
			case "explicit":
				config.AWSAccessKeyId = types.StringValue("explicit-key")
				config.AWSSecretAccessKey = types.StringValue("explicit-secret")
			case "partial":
				config.AWSAccessKeyId = types.StringValue("incomplete")
			case "api-key":
				config.CloudProviderType = types.StringNull()
				config.ApiKey = types.StringValue("test-api-key")
				config.AppKey = types.StringValue("test-app-key")
			}
			p := New().(*FrameworkProvider)
			diags := defaultConfigureFunc(p, &provider.ConfigureRequest{}, config)
			if mode == "partial" {
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
				return
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
