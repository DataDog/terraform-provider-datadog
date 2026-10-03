package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// Exercise Terraform Core's plan/state consistency checks without requiring an
// enabled live org or fabricating a recorded API cassette.
func TestSyntheticsGlobalVariableEmailTerraformLifecycle(t *testing.T) {
	const address = "7a03c7471a3e4a68e60a.0f7d3f56-17a8-4e90-a342-9b6ae2a709bc@synthetics.dtdg.co"
	var mu sync.Mutex
	var variable map[string]interface{}
	creates, deletes := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if !strings.HasPrefix(req.URL.Path, "/api/v1/synthetics/variables") {
			t.Errorf("unexpected endpoint %s", req.URL.Path)
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case http.MethodPost, http.MethodPut:
			var body map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
				http.Error(w, "invalid request", 400)
				return
			}
			email, _ := body["is_email"].(bool)
			if email {
				if _, ok := body["value"]; ok {
					t.Error("email request included value")
				}
				body["value"] = map[string]interface{}{"value": address, "secure": false}
			}
			if req.Method == http.MethodPost {
				creates++
				body["id"] = fmt.Sprintf("00000000-0000-4000-8000-%012d", creates)
			} else {
				body["id"] = variable["id"]
			}
			variable = body
		case http.MethodGet:
			if variable == nil {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":["not found"]}`))
				return
			}
		case http.MethodDelete:
			deletes++
			variable = nil
			_, _ = w.Write([]byte(`{}`))
			return
		default:
			t.Errorf("unexpected method %s", req.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(variable)
	}))
	defer server.Close()
	config := func(description, attrs string) string {
		return fmt.Sprintf(`
provider "datadog" {
 api_key = "test"
 app_key = "test"
 api_url = %q
 validate = "false"
}
resource "datadog_synthetics_global_variable" "email" {
 name = "PERSISTENT_EMAIL"
 description = %q
 %s
}
`, server.URL, description, attrs)
	}
	emailConfig := config("Created", "is_email = true")
	updatedConfig := config("Updated", "is_email = true")
	const resourceName = "datadog_synthetics_global_variable.email"
	var originalID string
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"datadog": providerserver.NewProtocol6WithError(fwprovider.New())},
		Steps: []resource.TestStep{
			{Config: emailConfig, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "is_email", "true"),
				resource.TestCheckResourceAttr(resourceName, "value", address),
				func(s *terraform.State) error {
					originalID = s.RootModule().Resources[resourceName].Primary.ID
					return nil
				},
			)},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
			{Config: updatedConfig, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
				resource.TestCheckResourceAttr(resourceName, "value", address),
				func(s *terraform.State) error {
					if s.RootModule().Resources[resourceName].Primary.ID != originalID {
						return fmt.Errorf("metadata update replaced the variable")
					}
					return nil
				},
			)},
			{Config: updatedConfig, PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: config("Regular", `value = "text"`), Check: resource.TestCheckResourceAttr(resourceName, "value", "text")},
			{Config: emailConfig, Check: resource.TestCheckResourceAttr(resourceName, "value", address)},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if creates != 3 || deletes != 3 {
		t.Fatalf("expected two kind replacements and final cleanup, got creates=%d deletes=%d", creates, deletes)
	}
}

func TestAccDatadogSyntheticsGlobalVariableEmail_Lifecycle(t *testing.T) {
	if isReplaying() {
		t.Skip("requires a persistent-email-enabled sandbox; the local Terraform lifecycle test covers offline behavior")
	}
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := getUniqueVariableName(ctx, t)
	const resourceName = "datadog_synthetics_global_variable.email"
	config := func(description string) string {
		return fmt.Sprintf(`
resource "datadog_synthetics_global_variable" "email" {
 name = %q
 description = %q
 is_email = true
}
`, name, description)
	}
	var generatedAddress string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsGlobalVariableResourceIsDestroyed(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{Config: config("Created"), Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "is_email", "true"),
				resource.TestCheckResourceAttr(resourceName, "secure", "false"),
				resource.TestMatchResourceAttr(resourceName, "value", regexp.MustCompile(`^[a-f0-9]{20}\.[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}@[^@]+$`)),
				func(s *terraform.State) error {
					generatedAddress = s.RootModule().Resources[resourceName].Primary.Attributes["value"]
					return nil
				},
			)},
			{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
			{Config: config("Updated"), Check: func(s *terraform.State) error {
				if actual := s.RootModule().Resources[resourceName].Primary.Attributes["value"]; actual != generatedAddress {
					return fmt.Errorf("email address changed after metadata update")
				}
				return nil
			}},
			{Config: config("Updated"), PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
