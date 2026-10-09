package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	datadogClient "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// Exercise real Terraform import, refresh, update and plan against an API that
// never persists secure on global references and redacts secure local values.
func TestAccDatadogSyntheticsAPITest_globalVariableSecureLifecycle(t *testing.T) {
	var mu sync.Mutex
	var remote map[string]interface{}
	var importRemote map[string]interface{}
	updates := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/synthetics/tests/delete":
			importRemote = remote
			remote = nil
			_, err := w.Write([]byte(`{"deleted_tests":[]}`))
			assert.NoError(t, err)
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/synthetics/tests/api",
			r.Method == http.MethodPut && r.URL.Path == "/api/v1/synthetics/tests/api/abc-def-ghi":
			if err := json.NewDecoder(r.Body).Decode(&remote); !assert.NoError(t, err) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.Method == http.MethodPut {
				updates++
			}
			remote["public_id"] = "abc-def-ghi"
			config := remote["config"].(map[string]interface{})
			for _, raw := range config["configVariables"].([]interface{}) {
				variable := raw.(map[string]interface{})
				if variable["type"] == "global" {
					assert.NotContains(t, variable, "secure", "global security must not be sent to the API")
					// Reproduce the API response that caused false -> true drift.
					variable["secure"] = false
				} else if variable["secure"] == true {
					delete(variable, "example")
					delete(variable, "pattern")
				}
			}
		case r.Method == http.MethodGet && (r.URL.Path == "/api/v1/synthetics/tests/abc-def-ghi" || r.URL.Path == "/api/v1/synthetics/tests/api/abc-def-ghi"):
			if remote == nil {
				w.WriteHeader(http.StatusNotFound)
				_, err := w.Write([]byte(`{"errors":["not found"]}`))
				assert.NoError(t, err)
				return
			}
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		assert.NoError(t, json.NewEncoder(w).Encode(remote))
	}))
	defer server.Close()

	clientConfig := datadogClient.NewConfiguration()
	clientConfig.Servers = datadogClient.ServerConfigurations{{URL: server.URL}}
	clientConfig.OperationServers = nil
	clientConfig.HTTPClient = server.Client()
	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"datadog_synthetics_test": datadog.Provider().ResourcesMap["datadog_synthetics_test"],
		},
		ConfigureContextFunc: func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return &datadog.ProviderConfiguration{
				Auth: context.Background(),
				DatadogApiInstances: &utils.ApiInstances{
					HttpClient: datadogClient.NewAPIClient(clientConfig),
				},
			}, nil
		},
	}
	config := func(name, firstSecure, secondSecure, thirdSecure string) string {
		return fmt.Sprintf(`
resource "datadog_synthetics_test" "global_secure" {
  name = %q
  type = "api"
  subtype = "http"
  status = "paused"
  locations = ["aws:us-east-1"]
  request_definition {
    method = "GET"
    url = "https://example.com"
  }
  assertion {
    type = "statusCode"
    operator = "is"
    target = "200"
  }
  options_list {
    tick_every = 900
  }
  config_variable {
    name = "GLOBAL_TRUE"
    type = "global"
    id = "global-true-id"
    %s
  }
  config_variable {
    name = "GLOBAL_FALSE"
    type = "global"
    id = "global-false-id"
    %s
  }
  config_variable {
    name = "GLOBAL_OMITTED"
    type = "global"
    id = "global-omitted-id"
    %s
  }
  config_variable {
    name = "LOCAL_SECURE"
    type = "text"
    secure = true
    pattern = "local-secret"
    example = "secret-example"
  }
  config_variable {
    name = "LOCAL_PLAIN"
    type = "text"
    secure = false
    pattern = "plain-value"
    example = "plain-example"
  }
}`, name, firstSecure, secondSecure, thirdSecure)
	}
	const resourceName = "datadog_synthetics_test.global_secure"
	check := func(first, second, third string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "config_variable.0.secure", first),
			resource.TestCheckResourceAttr(resourceName, "config_variable.1.secure", second),
			resource.TestCheckResourceAttr(resourceName, "config_variable.2.secure", third),
			resource.TestCheckResourceAttr(resourceName, "config_variable.3.secure", "true"),
			resource.TestCheckResourceAttr(resourceName, "config_variable.3.pattern", "local-secret"),
			resource.TestCheckResourceAttr(resourceName, "config_variable.4.secure", "false"),
			resource.TestCheckResourceAttr(resourceName, "config_variable.4.pattern", "plain-value"),
		)
	}
	initial := config("initial", "secure = true", "secure = false", "")
	firstUpdate := config("first update", "secure = true", "secure = false", "")
	secondUpdate := config("second update", "secure = true", "secure = false", "")
	toggled := config("second update", "secure = false", "secure = true", "secure = true")
	checkImport := func(states []*terraform.InstanceState) error {
		for i := 0; i < 3; i++ {
			if got := states[0].Attributes[fmt.Sprintf("config_variable.%d.secure", i)]; got != "false" {
				return fmt.Errorf("imported global variable %d secure = %s, want false", i, got)
			}
		}
		if got := states[0].Attributes["config_variable.3.secure"]; got != "true" {
			return fmt.Errorf("imported local variable secure = %s, want true", got)
		}
		return nil
	}
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"datadog": provider},
		Steps: []resource.TestStep{
			{Config: initial, Check: check("true", "false", "false")},
			{Config: initial, PlanOnly: true},
			{Config: firstUpdate, Check: check("true", "false", "false")},
			{Config: firstUpdate, PlanOnly: true},
			{Config: secondUpdate, Check: check("true", "false", "false")},
			{Config: secondUpdate, PlanOnly: true},
			{Config: toggled, Check: check("false", "true", "true")},
			{Config: toggled, PlanOnly: true},
			{Config: initial, Check: check("true", "false", "false")},
			{
				ResourceName: resourceName, ImportState: true, ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"config_variable.0.secure", "config_variable.3.pattern", "config_variable.3.example"},
				ImportStateCheck:        checkImport,
			},
		},
	})

	// Seed an existing API test and import into an empty Terraform state, then
	// verify that setting secure=true converges with a single update.
	mu.Lock()
	remote = importRemote
	mu.Unlock()
	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{"datadog": provider},
		Steps: []resource.TestStep{
			{
				Config: initial, ResourceName: resourceName, ImportState: true,
				ImportStateId: "abc-def-ghi", ImportStatePersist: true, ImportStateCheck: checkImport,
			},
			// Applying the configured value after import must converge after one update.
			{Config: initial, Check: check("true", "false", "false")},
			{Config: initial, PlanOnly: true},
			{Config: initial, PlanOnly: true},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 5, updates, "only intentional changes and post-import reconciliation should update")
}
