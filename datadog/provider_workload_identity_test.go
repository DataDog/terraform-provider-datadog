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
	"github.com/hashicorp/go-cty/cty"
	framework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/terraformauth"
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
		{terraformauth.TerraformWorkloadIdentityTokenEnv, terraformauth.TerraformWorkloadIdentityTokenFallbackEnv},
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
				t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, token)
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
			for _, tokenEnv := range []string{terraformauth.TerraformWorkloadIdentityTokenEnv, terraformauth.TerraformWorkloadIdentityTokenFallbackEnv} {
				t.Run(implementation+"/"+failure+"/"+tokenEnv, func(t *testing.T) {
					clearWorkloadIdentityTestEnv(t)
					selected := workloadIdentityTestToken(t, "https://app.terraform.io")
					t.Setenv(tokenEnv, selected)
					if tokenEnv == terraformauth.TerraformWorkloadIdentityTokenEnv {
						t.Setenv(terraformauth.TerraformWorkloadIdentityTokenFallbackEnv, workloadIdentityTestToken(t, "https://app.eu.terraform.io"))
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
					t.Setenv("DD_CLOUD_PROVIDER_TYPE", "aws")
					t.Setenv("DD_BEARER_TOKEN", "test-bearer")
					t.Setenv("DD_API_KEY", "test-api-key")
					t.Setenv("DD_APP_KEY", "test-app-key")
					client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
						"api_url":  server.URL,
						"org_uuid": workloadIdentityTestOrg,
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
						t.Setenv(terraformauth.TerraformWorkloadIdentityTokenFallbackEnv, token)
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

func TestTerraformWorkloadIdentityOrganizationBinding(t *testing.T) {
	const otherOrg = "22222222-2222-4222-8222-222222222222"
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, source := range []string{"org_uuid", "DD_ORG_UUID", "DATADOG_ORG_UUID"} {
			for _, tokenEnv := range []string{terraformauth.TerraformWorkloadIdentityTokenEnv, terraformauth.TerraformWorkloadIdentityTokenFallbackEnv} {
				for _, validate := range []string{"true", "false"} {
					t.Run(strings.Join([]string{implementation, source, tokenEnv, validate}, "/"), func(t *testing.T) {
						clearWorkloadIdentityTestEnv(t)
						token := workloadIdentityTestToken(t, "https://app.terraform.io")
						t.Setenv(tokenEnv, token)
						requests := 0
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
							requests++
							w.WriteHeader(http.StatusInternalServerError)
						}))
						defer server.Close()
						t.Setenv("DD_CLOUD_PROVIDER_TYPE", "aws")
						t.Setenv("DD_BEARER_TOKEN", "test-bearer")
						t.Setenv("DD_API_KEY", "test-api-key")
						t.Setenv("DD_APP_KEY", "test-app-key")
						values := map[string]interface{}{
							"api_url":  server.URL,
							"validate": validate,
						}
						if source == "org_uuid" {
							values[source] = otherOrg
						} else {
							t.Setenv(source, otherOrg)
						}
						_, _, err := configureWorkloadIdentityTestProvider(t, implementation, values)
						require.ErrorContains(t, err, "does not match configured org_uuid")
						require.Contains(t, err.Error(), workloadIdentityTestOrg)
						require.Contains(t, err.Error(), otherOrg)
						require.NotContains(t, err.Error(), token)
						require.Zero(t, requests, "organization mismatches must fail before exchange or fallback")
					})
				}
			}
		}
	}
}

func TestTerraformWorkloadIdentityExplicitAuthentication(t *testing.T) {
	const otherOrg = "22222222-2222-4222-8222-222222222222"
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, mode := range []string{"keys", "api-key", "app-key", "bearer", "aws"} {
			t.Run(implementation+"/"+mode, func(t *testing.T) {
				clearWorkloadIdentityTestEnv(t)
				t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
				t.Setenv("DD_CLOUD_PROVIDER_TYPE", "aws")
				t.Setenv("DD_BEARER_TOKEN", "env-bearer")
				t.Setenv("DD_API_KEY", "env-api")
				t.Setenv("DD_APP_KEY", "env-app")
				values := map[string]interface{}{
					"validate": "false",
					"org_uuid": otherOrg,
					// Explicit authentication also bypasses strict token selection.
					"workload_identity_token_tag": "MISSING",
				}
				expectedAPI, expectedApp := "env-api", "env-app"
				switch mode {
				case "keys", "api-key":
					values["api_key"] = "explicit-api"
					expectedAPI = "explicit-api"
					if mode == "keys" {
						values["app_key"] = "explicit-app"
						expectedApp = "explicit-app"
					}
				case "app-key":
					values["app_key"] = "explicit-app"
					expectedApp = "explicit-app"
				case "bearer":
					values["bearer_token"] = "explicit-bearer"
				case "aws":
					values["cloud_provider_type"] = "aws"
				}
				requests := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					require.Equal(t, "/api/v2/roles", r.URL.Path, "explicit authentication must not exchange the ambient WIT")
					if mode == "bearer" {
						require.Equal(t, "Bearer explicit-bearer", r.Header.Get("Authorization"))
						require.Empty(t, r.Header.Get("DD-API-KEY"))
						require.Empty(t, r.Header.Get("DD-APPLICATION-KEY"))
					} else {
						require.Empty(t, r.Header.Get("Authorization"))
						require.Equal(t, expectedAPI, r.Header.Get("DD-API-KEY"))
						require.Equal(t, expectedApp, r.Header.Get("DD-APPLICATION-KEY"))
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"data":[]}`))
				}))
				defer server.Close()
				values["api_url"] = server.URL
				client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, values)
				require.NoError(t, err, "the ambient WIT's different organization must not be checked")
				if mode == "aws" {
					require.Equal(t, "aws", client.GetConfig().DelegatedTokenConfig.Provider)
					require.Equal(t, otherOrg, client.GetConfig().DelegatedTokenConfig.OrgUUID)
					require.IsType(t, &api.AWSAuth{}, client.GetConfig().DelegatedTokenConfig.ProviderAuth)
					require.Zero(t, requests)
					return
				}
				require.Nil(t, client.GetConfig().DelegatedTokenConfig)
				_, _, err = datadogV2.NewRolesApi(client).ListRoles(auth)
				require.NoError(t, err)
				require.Equal(t, 1, requests)
			})
		}
		for _, field := range []string{"api_key", "app_key"} {
			t.Run(implementation+"/incomplete-"+field, func(t *testing.T) {
				clearWorkloadIdentityTestEnv(t)
				t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
				t.Setenv("DD_CLOUD_PROVIDER_TYPE", "aws")
				t.Setenv("DD_BEARER_TOKEN", "env-bearer")
				_, _, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{field: "explicit-key"})
				require.ErrorContains(t, err, "credentials are required")
			})
		}
	}
}

func TestTerraformWorkloadIdentityExplicitTagFailure(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, value := range []string{"", "malformed", "header.eyJhdWQiOiJ2YXVsdCJ9.signature"} {
			t.Run(implementation+"/"+value, func(t *testing.T) {
				clearWorkloadIdentityTestEnv(t)
				t.Setenv("TFC_WORKLOAD_IDENTITY_TOKEN_SELECTED", value)
				t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
				t.Setenv("DD_API_KEY", "env-api")
				t.Setenv("DD_APP_KEY", "env-app")
				_, _, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
					"workload_identity_token_tag": "SELECTED",
					"validate":                    "false",
				})
				require.ErrorContains(t, err, "TFC_WORKLOAD_IDENTITY_TOKEN_SELECTED")
				require.ErrorContains(t, err, "valid datadog/<uuid> audience")
			})
		}
	}
}

func TestTerraformWorkloadIdentityAliasIsolation(t *testing.T) {
	const otherOrg = "22222222-2222-4222-8222-222222222222"
	for _, implementation := range []string{"sdkv2", "framework"} {
		t.Run(implementation, func(t *testing.T) {
			clearWorkloadIdentityTestEnv(t)
			tokenA := workloadIdentityTestToken(t, "https://app.terraform.io")
			payload, err := json.Marshal(map[string]string{"iss": "https://app.terraform.io", "aud": "datadog/" + otherOrg})
			require.NoError(t, err)
			tokenB := "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
			t.Setenv("TFC_WORKLOAD_IDENTITY_TOKEN_ORG_A", tokenA)
			t.Setenv("TFC_WORKLOAD_IDENTITY_TOKEN_ORG_B", tokenB)
			t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, tokenA)
			t.Setenv(terraformauth.TerraformWorkloadIdentityTokenFallbackEnv, tokenA)
			exchanges := map[string]int{}
			requests := map[string]int{}
			rejectA := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/v2/delegated-token":
					name := ""
					switch r.Header.Get("Authorization") {
					case "Delegated " + tokenA:
						name = "a"
					case "Delegated " + tokenB:
						name = "b"
					default:
						t.Errorf("unexpected proof")
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					exchanges[name]++
					if name == "a" && rejectA {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"attributes": map[string]interface{}{
						"access_token": fmt.Sprintf("%s-%d", name, exchanges[name]),
						"expires":      strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10),
					}}})
				case "/api/v2/roles":
					header := r.Header.Get("Authorization")
					switch header {
					case "Bearer a-1", "Bearer a-2":
						requests["a"]++
					case "Bearer b-1":
						requests["b"]++
					case "":
						require.Equal(t, "static-api", r.Header.Get("DD-API-KEY"))
						require.Equal(t, "static-app", r.Header.Get("DD-APPLICATION-KEY"))
						requests["static"]++
					default:
						t.Errorf("unexpected access token %q", header)
					}
					_, _ = w.Write([]byte(`{"data":[]}`))
				default:
					t.Errorf("unexpected path: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			clientA, authA, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"api_url": server.URL, "workload_identity_token_tag": "ORG_A", "org_uuid": workloadIdentityTestOrg,
			})
			require.NoError(t, err)
			clientB, authB, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"api_url": server.URL, "workload_identity_token_tag": "ORG_B", "org_uuid": otherOrg,
			})
			require.NoError(t, err)
			clientStatic, authStatic, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"api_url": server.URL, "validate": "false", "api_key": "static-api", "app_key": "static-app", "org_uuid": otherOrg,
			})
			require.NoError(t, err)
			require.Equal(t, workloadIdentityTestOrg, clientA.GetConfig().DelegatedTokenConfig.OrgUUID)
			require.Equal(t, otherOrg, clientB.GetConfig().DelegatedTokenConfig.OrgUUID)
			for range 2 {
				for _, client := range []struct {
					client *api.APIClient
					auth   context.Context
				}{
					{clientA, authA}, {clientB, authB}, {clientStatic, authStatic},
				} {
					_, _, err := datadogV2.NewRolesApi(client.client).ListRoles(client.auth)
					require.NoError(t, err)
				}
			}
			require.Equal(t, map[string]int{"a": 1, "b": 1}, exchanges)
			require.Equal(t, map[string]int{"a": 2, "b": 2, "static": 2}, requests)
			credentialsA := authA.Value(api.ContextDelegatedToken).(*api.DelegatedTokenCredentials)
			credentialsB := authB.Value(api.ContextDelegatedToken).(*api.DelegatedTokenCredentials)
			require.NotSame(t, credentialsA, credentialsB)
			credentialsA.Expiration = time.Now().Add(-time.Second)
			_, _, err = datadogV2.NewRolesApi(clientA).ListRoles(authA)
			require.NoError(t, err)
			_, _, err = datadogV2.NewRolesApi(clientB).ListRoles(authB)
			require.NoError(t, err)
			require.Equal(t, map[string]int{"a": 2, "b": 1}, exchanges)
			// Refresh rejection for one alias must not select the other token or break its cache.
			rejectA = true
			credentialsA.Expiration = time.Now().Add(-time.Second)
			_, _, err = datadogV2.NewRolesApi(clientA).ListRoles(authA)
			require.ErrorContains(t, err, "401")
			_, _, err = datadogV2.NewRolesApi(clientB).ListRoles(authB)
			require.NoError(t, err)
			require.Equal(t, map[string]int{"a": 3, "b": 1}, exchanges)
			require.Equal(t, map[string]int{"a": 3, "b": 4, "static": 2}, requests)
		})
	}
}

func TestTerraformWorkloadIdentityUnknownTag(t *testing.T) {
	clearWorkloadIdentityTestEnv(t)
	t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
	p := Provider()
	config := terraform.NewResourceConfigRaw(map[string]interface{}{"validate": "false"})
	config.CtyValue = cty.ObjectVal(map[string]cty.Value{"workload_identity_token_tag": cty.UnknownVal(cty.String)})
	diags := p.Configure(context.Background(), config)
	require.True(t, diags.HasError())
	require.Contains(t, fmt.Sprint(diags), "must be known before configuring the provider")

	_, _, err := configureWorkloadIdentityTestProvider(t, "framework", map[string]interface{}{
		"workload_identity_token_tag": tftypes.UnknownValue, "validate": "false",
	})
	require.ErrorContains(t, err, "must be known before configuring the provider")
}

func TestTerraformWorkloadIdentityUnknownExplicitCredentials(t *testing.T) {
	for _, field := range []string{"api_key", "app_key", "bearer_token", "cloud_provider_type"} {
		t.Run(field, func(t *testing.T) {
			clearWorkloadIdentityTestEnv(t)
			t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
			p := Provider()
			config := terraform.NewResourceConfigRaw(map[string]interface{}{"validate": "false"})
			config.CtyValue = cty.ObjectVal(map[string]cty.Value{field: cty.UnknownVal(cty.String)})
			diags := p.Configure(context.Background(), config)
			require.True(t, diags.HasError())
			require.Contains(t, fmt.Sprint(diags), field+" must be known")
			_, _, err := configureWorkloadIdentityTestProvider(t, "framework", map[string]interface{}{
				field: tftypes.UnknownValue, "validate": "false",
			})
			require.ErrorContains(t, err, field+" must be known")
		})
	}
}
