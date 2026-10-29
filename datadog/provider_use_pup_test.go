package datadog

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/pupauth"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/terraformauth"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const pupTestUser = "33333333-3333-4333-8333-333333333333"

// TestPupHelper is not a test: a stand-in `pup` on PATH runs the test binary
// into it. It prints a JWT for PUP_TEST_ORG, records it in PUP_TEST_LAST, and
// counts its runs in PUP_TEST_COUNT.
func TestPupHelper(t *testing.T) {
	if os.Getenv("PUP_TEST_HELPER") != "1" {
		return
	}
	n := 1
	if b, err := os.ReadFile(os.Getenv("PUP_TEST_COUNT")); err == nil {
		prev, _ := strconv.Atoi(string(b))
		n = prev + 1
	}
	_ = os.WriteFile(os.Getenv("PUP_TEST_COUNT"), []byte(strconv.Itoa(n)), 0o600)
	payload := fmt.Sprintf(`{"exp":%d,"run":%d,"ddauth":{"org_uuid":%q,"user_uuid":%q}}`,
		time.Now().Add(time.Hour).Unix(), n, os.Getenv("PUP_TEST_ORG"), pupTestUser)
	token := "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".c2ln"
	_ = os.WriteFile(os.Getenv("PUP_TEST_LAST"), []byte(token), 0o600)
	fmt.Println(token)
	os.Exit(0)
}

// fakePupOnPath puts a `pup` that runs TestPupHelper first on PATH, logged in to
// workloadIdentityTestOrg, and starts from a fresh pupauth state.
func fakePupOnPath(t *testing.T) (runs func() int) {
	t.Helper()
	clearWorkloadIdentityTestEnv(t)
	for _, name := range utils.UsePupEnvVars {
		t.Setenv(name, "")
	}
	dir := t.TempDir()
	script := fmt.Sprintf("#!/bin/sh\nexec %q -test.run='^TestPupHelper$' -- \"$@\"\n", os.Args[0])
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pup"), []byte(script), 0o700))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("PUP_TEST_HELPER", "1")
	t.Setenv("PUP_TEST_COUNT", filepath.Join(dir, "count"))
	t.Setenv("PUP_TEST_LAST", filepath.Join(dir, "last"))
	t.Setenv("PUP_TEST_ORG", workloadIdentityTestOrg)
	pupauth.ResetForTesting()
	t.Cleanup(pupauth.ResetForTesting)
	return func() int {
		b, err := os.ReadFile(filepath.Join(dir, "count"))
		if os.IsNotExist(err) {
			return 0
		}
		require.NoError(t, err)
		n, err := strconv.Atoi(string(b))
		require.NoError(t, err)
		return n
	}
}

func TestUsePupAuthentication(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		for _, source := range []string{"attribute", "DD_USE_PUP", "DATADOG_USE_PUP"} {
			t.Run(implementation+"/"+source, func(t *testing.T) {
				runs := fakePupOnPath(t)
				t.Setenv("TF_LOG", "DEBUG")
				var logs bytes.Buffer
				previousLogOutput := log.Writer()
				log.SetOutput(&logs)
				t.Cleanup(func() { log.SetOutput(previousLogOutput) })

				requests := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					require.Equal(t, "/api/v2/roles", r.URL.Path)
					last, err := os.ReadFile(os.Getenv("PUP_TEST_LAST"))
					require.NoError(t, err)
					require.Equal(t, "Bearer "+string(last), r.Header.Get("Authorization"))
					require.Empty(t, r.Header.Get("DD-API-KEY"))
					require.Empty(t, r.Header.Get("DD-APPLICATION-KEY"))
					requests++
					_, _ = w.Write([]byte(`{"data":[]}`))
				}))
				defer server.Close()

				values := map[string]interface{}{"api_url": server.URL, "org_uuid": workloadIdentityTestOrg}
				if source == "attribute" {
					values["use_pup"] = true
				} else {
					t.Setenv(source, "true")
				}
				client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, values)
				require.NoError(t, err)
				require.Equal(t, pupauth.Provider, client.GetConfig().DelegatedTokenConfig.Provider)

				roles := datadogV2.NewRolesApi(client)
				for range 2 {
					_, _, err := roles.ListRoles(auth)
					require.NoError(t, err)
				}
				require.Equal(t, 1, runs(), "validation and requests share one run of pup")
				require.Equal(t, 2, requests)

				last, err := os.ReadFile(os.Getenv("PUP_TEST_LAST"))
				require.NoError(t, err)
				require.NotContains(t, logs.String(), string(last))
			})
		}
	}
}

func TestUsePupIsTheLastResort(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		t.Run(implementation+"/API and application keys win", func(t *testing.T) {
			runs := fakePupOnPath(t)
			t.Setenv("DD_API_KEY", "test-api-key")
			t.Setenv("DD_APP_KEY", "test-app-key")
			client, _, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"use_pup": true, "validate": "false",
			})
			require.NoError(t, err)
			require.Nil(t, client.GetConfig().DelegatedTokenConfig)
			require.Equal(t, 0, runs())
		})

		t.Run(implementation+"/a bearer token wins", func(t *testing.T) {
			runs := fakePupOnPath(t)
			t.Setenv("DD_BEARER_TOKEN", "test-bearer")
			client, _, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"use_pup": true, "validate": "false",
			})
			require.NoError(t, err)
			require.Nil(t, client.GetConfig().DelegatedTokenConfig)
			require.Equal(t, 0, runs())
		})

		t.Run(implementation+"/workload identity wins", func(t *testing.T) {
			runs := fakePupOnPath(t)
			t.Setenv(terraformauth.TerraformWorkloadIdentityTokenEnv, workloadIdentityTestToken(t, "https://app.terraform.io"))
			client, _, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"use_pup": true, "validate": "false",
			})
			require.NoError(t, err)
			require.Equal(t, "terraform", client.GetConfig().DelegatedTokenConfig.Provider)
			require.Equal(t, 0, runs())
		})

		t.Run(implementation+"/use_pup = false beats DD_USE_PUP", func(t *testing.T) {
			runs := fakePupOnPath(t)
			t.Setenv("DD_USE_PUP", "true")
			client, _, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"use_pup": false, "validate": "false",
			})
			require.NoError(t, err)
			require.Nil(t, client.GetConfig().DelegatedTokenConfig)
			require.Equal(t, 0, runs())
		})
	}
}

func TestUsePupChecksOrgUUID(t *testing.T) {
	for _, implementation := range []string{"sdkv2", "framework"} {
		t.Run(implementation, func(t *testing.T) {
			fakePupOnPath(t)
			client, auth, err := configureWorkloadIdentityTestProvider(t, implementation, map[string]interface{}{
				"use_pup": true, "org_uuid": "22222222-2222-4222-8222-222222222222", "api_url": "http://127.0.0.1:1",
			})
			// SDKv2 runs pup while validating; the framework defers to the first request.
			if err == nil {
				_, _, err = datadogV2.NewRolesApi(client).ListRoles(auth)
			}
			require.ErrorContains(t, err, "pup is logged in to organization")
		})
	}
}
