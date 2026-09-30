package datadog

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	api "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	framework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const workloadIdentityTestOrg = "11111111-1111-4111-8111-111111111111"

func workloadIdentityTestToken(t *testing.T, issuer string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]interface{}{
		"iss": issuer,
		"aud": "datadog/" + workloadIdentityTestOrg,
		"sub": "organization:example:project:Default Project:workspace:test:run_phase:plan",
	})
	require.NoError(t, err)
	return "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(payload) + ".c2lnbmF0dXJl"
}

func clearWorkloadIdentityTestEnv(t *testing.T) {
	t.Helper()
	for _, names := range [][]string{
		utils.APIKeyEnvVars, utils.APPKeyEnvVars, utils.BearerTokenEnvVars,
		utils.CloudProviderTypeEnvVars, utils.OrgUUIDEnvVars, utils.APIUrlEnvVars,
		{utils.TerraformWorkloadIdentityTokenEnv, utils.TerraformWorkloadIdentityTokenFallbackEnv},
	} {
		for _, name := range names {
			t.Setenv(name, "")
		}
	}
}

// Exercise both public provider configuration paths against the same HTTP oracle.
func configureWorkloadIdentityTestProvider(t *testing.T, implementation string, values map[string]interface{}) (*api.APIClient, context.Context, error) {
	t.Helper()
	ctx := context.Background()
	if implementation == "sdkv2" {
		p := Provider()
		data := schema.TestResourceDataRaw(t, p.Schema, values)
		meta, diags := p.ConfigureContextFunc(ctx, data)
		if diags.HasError() {
			return nil, nil, fmt.Errorf("%v", diags)
		}
		config := meta.(*ProviderConfiguration)
		return config.DatadogApiInstances.HttpClient, config.Auth, nil
	}
	p := fwprovider.New().(*fwprovider.FrameworkProvider)
	var schemaResponse framework.SchemaResponse
	p.Schema(ctx, framework.SchemaRequest{}, &schemaResponse)
	objectType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)
	attributes := make(map[string]tftypes.Value)
	for name, typ := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(typ, values[name])
	}
	var response framework.ConfigureResponse
	p.Configure(ctx, framework.ConfigureRequest{Config: tfsdk.Config{
		Raw: tftypes.NewValue(objectType, attributes), Schema: schemaResponse.Schema,
	}}, &response)
	if response.Diagnostics.HasError() {
		return nil, nil, fmt.Errorf("%v", response.Diagnostics)
	}
	return p.DatadogApiInstances.HttpClient, p.Auth, nil
}

func TestTerraformWorkloadIdentityAuthentication(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, issuer := range []string{"https://app.terraform.io", "https://app.eu.terraform.io", "https://terraform.example.com"} {
			t.Run(implementation+"/"+issuer, func(t *testing.T) {
				clearWorkloadIdentityTestEnv(t)
				t.Setenv("TF_LOG", "DEBUG")
				var logs bytes.Buffer
				previousLogOutput := log.Writer()
				log.SetOutput(&logs)
				t.Cleanup(func() { log.SetOutput(previousLogOutput) })
				token := workloadIdentityTestToken(t, issuer)
				t.Setenv(utils.TerraformWorkloadIdentityTokenEnv, token)
				proof := token
				if issuer == "https://terraform.example.com" {
					proof = "oidc-" + token
				}
				exchanges, requests := 0, 0
				reject := false
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					switch r.URL.Path {
					case "/api/v2/delegated-token":
						exchanges++
						require.Equal(t, http.MethodPost, r.Method)
						require.Equal(t, "Delegated "+proof, r.Header.Get("Authorization"))
						if reject {
							w.WriteHeader(http.StatusUnauthorized)
							_ = json.NewEncoder(w).Encode(map[string]interface{}{"errors": []string{token}})
							return
						}
						_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"attributes": map[string]interface{}{
							"access_token": fmt.Sprintf("delegated-%d", exchanges),
							"expires":      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
						}}})
					case "/api/v2/roles":
						requests++
						require.Equal(t, "Bearer "+fmt.Sprintf("delegated-%d", exchanges), r.Header.Get("Authorization"))
						require.Empty(t, r.Header.Get("DD-API-KEY"))
						require.Empty(t, r.Header.Get("DD-APPLICATION-KEY"))
						_, _ = w.Write([]byte(`{"data":[]}`))
					default:
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()
				client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
					"api_url": server.URL,
					// An org configured for another auth method must not filter the WIT.
					"org_uuid": "22222222-2222-4222-8222-222222222222",
				})
				require.NoError(t, err)
				require.Equal(t, "terraform", client.GetConfig().DelegatedTokenConfig.Provider)
				require.Equal(t, workloadIdentityTestOrg, client.GetConfig().DelegatedTokenConfig.OrgUUID)
				roles := datadogV2.NewRolesApi(client)
				for range 2 {
					_, _, err := roles.ListRoles(auth)
					require.NoError(t, err)
				}
				require.Equal(t, 1, exchanges, "reuse the SDK's cached delegated token")
				require.Equal(t, 2, requests)

				credentials := auth.Value(api.ContextDelegatedToken).(*api.DelegatedTokenCredentials)
				credentials.Expiration = time.Now().Add(-time.Second)
				_, _, err = roles.ListRoles(auth)
				require.NoError(t, err)
				require.Equal(t, 2, exchanges, "refresh through the same WIT exchange adapter")
				require.Equal(t, 3, requests)

				reject = true
				credentials.Expiration = time.Now().Add(-time.Second)
				_, _, err = roles.ListRoles(auth)
				require.Error(t, err)
				require.NotContains(t, err.Error(), token)
				require.Equal(t, 3, exchanges)
				require.Equal(t, 3, requests, "failed refresh must not send an unauthenticated or fallback request")
				require.NotContains(t, logs.String(), token)
				require.NotContains(t, logs.String(), "delegated-1")
				require.NotContains(t, logs.String(), "delegated-2")
			})
		}
	}
}

func TestTerraformWorkloadIdentityPrecedenceAndRejection(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, failure := range []string{"rejected", "malformed response", "empty access token"} {
			for _, tokenEnv := range []string{utils.TerraformWorkloadIdentityTokenEnv, utils.TerraformWorkloadIdentityTokenFallbackEnv} {
				t.Run(implementation+"/"+failure+"/"+tokenEnv, func(t *testing.T) {
					clearWorkloadIdentityTestEnv(t)
					selected := workloadIdentityTestToken(t, "https://app.terraform.io")
					t.Setenv(tokenEnv, selected)
					if tokenEnv == utils.TerraformWorkloadIdentityTokenEnv {
						t.Setenv(utils.TerraformWorkloadIdentityTokenFallbackEnv, workloadIdentityTestToken(t, "https://app.eu.terraform.io"))
					}
					exchanges := 0
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						require.Equal(t, "/api/v2/delegated-token", r.URL.Path, "must not validate static keys or call an API after ETS rejection")
						require.Equal(t, "Delegated "+selected, r.Header.Get("Authorization"), "must not use AWS or a lower-priority WIT")
						exchanges++
						switch failure {
						case "rejected":
							w.WriteHeader(http.StatusUnauthorized)
							_, _ = w.Write([]byte(selected))
						case "malformed response":
							// The SDK's parse error includes the body; it must not
							// leak either the workload or delegated access token.
							_ = json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "secret-access-token", "echo": selected})
						case "empty access token":
							_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"attributes": map[string]interface{}{"access_token": ""}}})
						}
					}))
					defer server.Close()
					client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
						"api_url":               server.URL,
						"cloud_provider_type":   "aws",
						"cloud_provider_region": "us-east-1",
						"org_uuid":              workloadIdentityTestOrg,
						"aws_access_key_id":     "test-aws-id",
						"aws_secret_access_key": "test-aws-secret",
						"bearer_token":          "test-bearer",
						"api_key":               "test-api-key",
						"app_key":               "test-app-key",
					})
					// SDKv2 exchanges during initialization; Framework defers to the
					// first API request because SDKv2 owns mux provider validation.
					if err == nil {
						_, _, err = datadogV2.NewRolesApi(client).ListRoles(auth)
					}
					require.Error(t, err)
					switch failure {
					case "rejected":
						require.Contains(t, err.Error(), "401")
					case "malformed response":
						require.Contains(t, err.Error(), "invalid response")
					case "empty access token":
						require.Contains(t, err.Error(), "empty access token")
					}
					require.NotContains(t, err.Error(), "secret-access-token")
					require.NotContains(t, err.Error(), selected)
					require.Equal(t, 1, exchanges)
				})
			}
		}
	}
}

func TestTerraformWorkloadIdentityFallback(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, mode := range []string{"aws", "bearer", "keys"} {
			for _, tokenKind := range []string{"missing", "unrelated"} {
				t.Run(strings.Join([]string{implementation, mode, tokenKind}, "/"), func(t *testing.T) {
					clearWorkloadIdentityTestEnv(t)
					if tokenKind == "unrelated" {
						token := "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"iss":"https://app.terraform.io","aud":"vault"}`)) + ".c2ln"
						t.Setenv(utils.TerraformWorkloadIdentityTokenFallbackEnv, token)
					}
					values := map[string]interface{}{
						"validate": "false",
						"api_key":  "test-api-key",
						"app_key":  "test-app-key",
						"org_uuid": "22222222-2222-4222-8222-222222222222",
					}
					if mode != "keys" {
						values["bearer_token"] = "test-bearer"
					}
					if mode == "aws" {
						values["cloud_provider_type"] = "aws"
					}
					client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, values)
					require.NoError(t, err)
					switch mode {
					case "aws":
						require.Equal(t, "aws", client.GetConfig().DelegatedTokenConfig.Provider)
						require.NotNil(t, auth.Value(api.ContextDelegatedToken))
						require.Nil(t, auth.Value(api.ContextAccessToken))
						require.Nil(t, auth.Value(api.ContextAPIKeys))
					case "bearer":
						require.Nil(t, client.GetConfig().DelegatedTokenConfig)
						require.Equal(t, "test-bearer", auth.Value(api.ContextAccessToken))
						require.Nil(t, auth.Value(api.ContextAPIKeys))
					case "keys":
						require.Nil(t, client.GetConfig().DelegatedTokenConfig)
						keys := auth.Value(api.ContextAPIKeys).(map[string]api.APIKey)
						require.Equal(t, "test-api-key", keys["apiKeyAuth"].Key)
						require.Equal(t, "test-app-key", keys["appKeyAuth"].Key)
					}
				})
			}
		}
	}
}
