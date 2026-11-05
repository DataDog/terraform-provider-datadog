package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	common "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// Exercise Terraform planning, the provider, and the API client against an offline
// HTTP fixture. This models the PATCH contract; it does not verify the live API.
func TestAccAwsCurConfigRemoveAccountFiltersLocal(t *testing.T) {
	t.Parallel()
	const configID = "1266"
	const collectionPath = "/api/v2/cost/aws_cur_config"
	var attributes map[string]json.RawMessage
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if request.URL.Path != collectionPath && request.URL.Path != collectionPath+"/"+configID {
			t.Errorf("unexpected API path: %s", request.URL.Path)
			http.Error(w, "unexpected API path", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost, http.MethodPatch:
			var payload struct {
				Data struct {
					Attributes map[string]json.RawMessage `json:"attributes"`
				} `json:"data"`
			}
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Errorf("decoding request: %v", err)
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if request.Method == http.MethodPost {
				attributes = payload.Data.Attributes
				attributes["created_at"] = json.RawMessage(`"2026-01-30T20:42:44Z"`)
				attributes["updated_at"] = json.RawMessage(`"2026-01-30T20:42:44Z"`)
				attributes["status_updated_at"] = json.RawMessage(`"2026-01-30T20:42:44Z"`)
				attributes["status"] = json.RawMessage(`"active"`)
				attributes["error_messages"] = json.RawMessage(`null`)
				attributes["months"] = json.RawMessage(`15`)
			} else {
				// PATCH preserves omitted fields. An explicit empty object clears filters.
				for name, value := range payload.Data.Attributes {
					attributes[name] = value
				}
				if filters, ok := payload.Data.Attributes["account_filters"]; ok {
					var fields map[string]json.RawMessage
					if err := json.Unmarshal(filters, &fields); err == nil && len(fields) == 0 {
						// The API response includes the nullable include_new_accounts field.
						attributes["account_filters"] = json.RawMessage(`{"include_new_accounts":null}`)
					}
				}
			}
		case http.MethodGet:
			if attributes == nil {
				http.Error(w, `{"errors":["not found"]}`, http.StatusNotFound)
				return
			}
		case http.MethodDelete:
			attributes = nil
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			t.Errorf("unexpected API method: %s", request.Method)
			http.Error(w, "unexpected API method", http.StatusMethodNotAllowed)
			return
		}
		data := map[string]interface{}{"id": configID, "type": "aws_cur_config", "attributes": attributes}
		var responseData interface{} = data
		if request.Method == http.MethodPatch {
			responseData = []interface{}{data}
		}
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"data": responseData}); err != nil {
			t.Errorf("encoding response: %v", err)
		}
	}))
	defer server.Close()

	config := common.NewConfiguration()
	config.Servers = common.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	p := &fwprovider.FrameworkProvider{
		Auth:                context.Background(),
		DatadogApiInstances: &utils.ApiInstances{HttpClient: common.NewAPIClient(config)},
		ConfigureCallbackFunc: func(*fwprovider.FrameworkProvider, *provider.ConfigureRequest, *fwprovider.ProviderSchema) frameworkDiag.Diagnostics {
			return nil
		},
	}
	checkFiltersCleared := func(*terraform.State) error {
		response, _, err := p.DatadogApiInstances.GetCloudCostManagementApiV2().GetCostAWSCURConfig(p.Auth, 1266)
		if err != nil {
			return err
		}
		filters := response.Data.Attributes.GetAccountFilters()
		includeNewAccounts, _ := filters.GetIncludeNewAccountsOk()
		if includeNewAccounts != nil || len(filters.GetIncludedAccounts()) != 0 || len(filters.GetExcludedAccounts()) != 0 {
			return fmt.Errorf("API still has account filters: %+v", filters)
		}
		return nil
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"datadog": providerserver.NewProtocol6WithError(p),
		},
		CheckDestroy: testAccCheckDatadogAwsCurConfigDestroy(p),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAwsCurConfigWithFilters("", true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_aws_cur_config.foo", "account_filters.include_new_accounts", "true"),
					resource.TestCheckResourceAttr("datadog_aws_cur_config.foo", "account_filters.excluded_accounts.0", "123456789012"),
				),
			},
			{
				Config: testAccCheckDatadogAwsCurConfigWithFilters("", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("datadog_aws_cur_config.foo", "account_filters.include_new_accounts", "false"),
					resource.TestCheckResourceAttr("datadog_aws_cur_config.foo", "account_filters.included_accounts.0", "123456789013"),
				),
			},
			{
				Config: testAccCheckDatadogAwsCurConfigBasic(""),
				Check: resource.ComposeTestCheckFunc(
					checkFiltersCleared,
					resource.TestCheckNoResourceAttr("datadog_aws_cur_config.foo", "account_filters.include_new_accounts"),
					resource.TestCheckNoResourceAttr("datadog_aws_cur_config.foo", "account_filters.included_accounts.#"),
					resource.TestCheckNoResourceAttr("datadog_aws_cur_config.foo", "account_filters.excluded_accounts.#"),
				),
			},
			{
				Config:             testAccCheckDatadogAwsCurConfigBasic(""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}
