package datadog

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	api "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/commandauth"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/terraformauth"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// TestBearerTokenCommandHelper is not a test: the tests below run the test
// binary as bearer_token_command. It prints "command-token-<run>", or with
// BEARER_TOKEN_COMMAND_TTL a JWT expiring that many seconds ahead, and records
// it in BEARER_TOKEN_COMMAND_LAST. It counts runs in BEARER_TOKEN_COMMAND_COUNT
// and fails when BEARER_TOKEN_COMMAND_FAIL is set.
func TestBearerTokenCommandHelper(t *testing.T) {
	if os.Getenv("BEARER_TOKEN_COMMAND_HELPER") != "1" {
		return
	}
	if os.Getenv("BEARER_TOKEN_COMMAND_FAIL") != "" {
		fmt.Fprint(os.Stderr, "not logged in")
		os.Exit(1)
	}
	counter := os.Getenv("BEARER_TOKEN_COMMAND_COUNT")
	b, _ := os.ReadFile(counter)
	n, _ := strconv.Atoi(string(b))
	n++
	_ = os.WriteFile(counter, []byte(strconv.Itoa(n)), 0o600)
	token := fmt.Sprintf("command-token-%d", n)
	if ttl, err := strconv.Atoi(os.Getenv("BEARER_TOKEN_COMMAND_TTL")); err == nil {
		claims := fmt.Sprintf(`{"exp":%d,"run":%d}`, time.Now().Add(time.Duration(ttl)*time.Second).Unix(), n)
		token = "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(claims)) + ".c2ln"
	}
	_ = os.WriteFile(os.Getenv("BEARER_TOKEN_COMMAND_LAST"), []byte(token), 0o600)
	fmt.Println(token)
	os.Exit(0)
}

// bearerTokenCommandTestArgv returns a command running the helper. The last
// argument keeps each subtest's token cache separate.
func bearerTokenCommandTestArgv(t *testing.T) []string {
	t.Helper()
	t.Setenv("BEARER_TOKEN_COMMAND_HELPER", "1")
	dir := t.TempDir()
	t.Setenv("BEARER_TOKEN_COMMAND_COUNT", filepath.Join(dir, "count"))
	t.Setenv("BEARER_TOKEN_COMMAND_LAST", filepath.Join(dir, "last"))
	return []string{os.Args[0], "-test.run=^TestBearerTokenCommandHelper$", "--", strings.ReplaceAll(t.Name(), " ", "_")}
}

func clearBearerTokenCommandTestEnv(t *testing.T) {
	t.Helper()
	clearWorkloadIdentityTestEnv(t)
	for _, name := range utils.BearerTokenCommandEnvVars {
		t.Setenv(name, "")
	}
}

// bearerTokenCommandTestServer answers ListRoles, checking each request carries
// the command's latest token and no API or application key.
func bearerTokenCommandTestServer(t *testing.T, requests *int) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.Equal(t, "/api/v2/roles", r.URL.Path)
		last, err := os.ReadFile(os.Getenv("BEARER_TOKEN_COMMAND_LAST"))
		require.NoError(t, err)
		require.Equal(t, "Bearer "+string(last), r.Header.Get("Authorization"))
		require.Empty(t, r.Header.Get("DD-API-KEY"))
		require.Empty(t, r.Header.Get("DD-APPLICATION-KEY"))
		*requests++
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(server.Close)
	return server
}

func bearerTokenCommandRuns(t *testing.T) int {
	t.Helper()
	b, err := os.ReadFile(os.Getenv("BEARER_TOKEN_COMMAND_COUNT"))
	if os.IsNotExist(err) {
		return 0
	}
	require.NoError(t, err)
	n, err := strconv.Atoi(string(b))
	require.NoError(t, err)
	return n
}

func TestBearerTokenCommandAuthentication(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, source := range []string{"attribute", "DD_BEARER_TOKEN_COMMAND", "DATADOG_BEARER_TOKEN_COMMAND"} {
			t.Run(implementation+"/"+source, func(t *testing.T) {
				clearBearerTokenCommandTestEnv(t)
				t.Setenv("TF_LOG", "DEBUG")
				var logs bytes.Buffer
				previousLogOutput := log.Writer()
				log.SetOutput(&logs)
				t.Cleanup(func() { log.SetOutput(previousLogOutput) })

				argv := bearerTokenCommandTestArgv(t)
				requests := 0
				server := bearerTokenCommandTestServer(t, &requests)
				values := map[string]interface{}{"api_url": server.URL}
				if source == "attribute" {
					values["bearer_token_command"] = argv
				} else {
					t.Setenv(source, strings.Join(argv, " "))
				}

				client, auth, err := configureBearerTokenCommandTestProvider(t, implementation, values)
				require.NoError(t, err)
				require.Equal(t, commandauth.Provider, client.GetConfig().DelegatedTokenConfig.Provider)

				roles := datadogV2.NewRolesApi(client)
				for range 2 {
					_, _, err := roles.ListRoles(auth)
					require.NoError(t, err)
				}
				require.Equal(t, 1, bearerTokenCommandRuns(t), "SDKv2 validation and the requests share one run")
				require.Equal(t, 2, requests)
				require.NotContains(t, logs.String(), "command-token-1")
			})
		}
	}
}

func TestBearerTokenCommandRefresh(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		t.Run(implementation, func(t *testing.T) {
			clearBearerTokenCommandTestEnv(t)
			argv := bearerTokenCommandTestArgv(t)
			// Expires one second after the one-minute skew, so it goes stale mid-test.
			t.Setenv("BEARER_TOKEN_COMMAND_TTL", "61")
			requests := 0
			server := bearerTokenCommandTestServer(t, &requests)
			client, auth, err := configureBearerTokenCommandTestProvider(t, implementation, map[string]interface{}{
				"api_url":              server.URL,
				"bearer_token_command": argv,
			})
			require.NoError(t, err)
			roles := datadogV2.NewRolesApi(client)
			_, _, err = roles.ListRoles(auth)
			require.NoError(t, err)
			require.Equal(t, 1, bearerTokenCommandRuns(t))

			credentials := auth.Value(api.ContextDelegatedToken).(*api.DelegatedTokenCredentials)
			require.WithinDuration(t, time.Now().Add(time.Second), credentials.Expiration, 2*time.Second, "expiry comes from the JWT exp, less the skew")
			time.Sleep(time.Until(credentials.Expiration) + 100*time.Millisecond)
			_, _, err = roles.ListRoles(auth)
			require.NoError(t, err)
			require.Equal(t, 2, bearerTokenCommandRuns(t), "an expired token reruns the command")
			require.Equal(t, 2, requests)

			t.Setenv("BEARER_TOKEN_COMMAND_FAIL", "1")
			time.Sleep(time.Until(credentials.Expiration) + 100*time.Millisecond)
			_, _, err = roles.ListRoles(auth)
			require.Error(t, err)
			require.Contains(t, err.Error(), "not logged in")
			require.Equal(t, 2, requests, "a failed run must not send an unauthenticated request")
		})
	}
}

func TestBearerTokenCommandPrecedence(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		t.Run(implementation+"/environment command and token conflict", func(t *testing.T) {
			clearBearerTokenCommandTestEnv(t)
			t.Setenv("DD_BEARER_TOKEN", "static-token")
			t.Setenv("DD_BEARER_TOKEN_COMMAND", strings.Join(bearerTokenCommandTestArgv(t), " "))
			_, _, err := configureBearerTokenCommandTestProvider(t, implementation, map[string]interface{}{"validate": "false"})
			require.ErrorContains(t, err, "set only one of DD_BEARER_TOKEN and DD_BEARER_TOKEN_COMMAND")
		})

		t.Run(implementation+"/explicit keys ignore the environment command", func(t *testing.T) {
			clearBearerTokenCommandTestEnv(t)
			t.Setenv("DD_BEARER_TOKEN_COMMAND", strings.Join(bearerTokenCommandTestArgv(t), " "))
			client, _, err := configureBearerTokenCommandTestProvider(t, implementation, map[string]interface{}{
				"api_key": "test-api-key", "app_key": "test-app-key", "validate": "false",
			})
			require.NoError(t, err)
			require.Nil(t, client.GetConfig().DelegatedTokenConfig)
			require.Equal(t, 0, bearerTokenCommandRuns(t))
		})

		t.Run(implementation+"/workload identity wins over the environment command", func(t *testing.T) {
			clearBearerTokenCommandTestEnv(t)
			t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
			t.Setenv("DD_BEARER_TOKEN_COMMAND", strings.Join(bearerTokenCommandTestArgv(t), " "))
			client, _, err := configureBearerTokenCommandTestProvider(t, implementation, map[string]interface{}{"validate": "false"})
			require.NoError(t, err)
			require.Equal(t, "terraform", client.GetConfig().DelegatedTokenConfig.Provider)
		})

		t.Run(implementation+"/the attribute wins over workload identity", func(t *testing.T) {
			clearBearerTokenCommandTestEnv(t)
			t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
			client, _, err := configureBearerTokenCommandTestProvider(t, implementation, map[string]interface{}{
				"bearer_token_command": bearerTokenCommandTestArgv(t), "validate": "false",
			})
			require.NoError(t, err)
			require.Equal(t, commandauth.Provider, client.GetConfig().DelegatedTokenConfig.Provider)
		})
	}

	t.Run("sdkv2/validation fails when the command fails", func(t *testing.T) {
		clearBearerTokenCommandTestEnv(t)
		argv := bearerTokenCommandTestArgv(t)
		t.Setenv("BEARER_TOKEN_COMMAND_FAIL", "1")
		_, _, err := configureBearerTokenCommandTestProvider(t, "sdkv2", map[string]interface{}{"bearer_token_command": argv})
		require.ErrorContains(t, err, "not logged in")
	})
}

// configureBearerTokenCommandTestProvider converts the command list to the
// shape each implementation's test harness expects.
func configureBearerTokenCommandTestProvider(t *testing.T, implementation string, values map[string]interface{}) (*api.APIClient, context.Context, error) {
	t.Helper()
	if argv, ok := values["bearer_token_command"].([]string); ok {
		if implementation == "sdkv2" {
			list := make([]interface{}, len(argv))
			for i, arg := range argv {
				list[i] = arg
			}
			values["bearer_token_command"] = list
		} else {
			list := make([]tftypes.Value, len(argv))
			for i, arg := range argv {
				list[i] = tftypes.NewValue(tftypes.String, arg)
			}
			values["bearer_token_command"] = list
		}
	}
	return configureWorkloadIdentityTestProvider(t, implementation, values)
}
