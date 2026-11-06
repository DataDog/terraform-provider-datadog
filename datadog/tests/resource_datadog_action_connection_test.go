package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	testConnectionBaseURL = "https://catfact.ninja"
)

func TestAccDatadogActionConnectionResource_AWSSecretsManager_Clickup(t *testing.T) {
	testAccActionConnectionAWSSecretsManager(t, "Clickup")
}

func TestAccDatadogActionConnectionResource_AWSSecretsManager_GCP(t *testing.T) {
	testAccActionConnectionAWSSecretsManager(t, "GCP")
}

func TestAccDatadogActionConnectionResource_AWSSecretsManager_HTTP(t *testing.T) {
	testAccActionConnectionAWSSecretsManager(t, "HTTP")
}

func testAccActionConnectionAWSSecretsManager(t *testing.T, integration string) {
	t.Helper()
	cases := []struct {
		name     string
		template string
		fields   []actionConnectionAWSSecretTestField
	}{
		{
			name:     "Clickup",
			template: "clickup {\napi_key {\n%s\n}\n}",
			fields:   []actionConnectionAWSSecretTestField{{"api_token", "clickup.api_key.api_token", "synthetic-token"}},
		},
		{
			name:     "GCP",
			template: "gcp {\nservice_account {\n%s\n}\n}",
			fields: []actionConnectionAWSSecretTestField{
				{"private_key", "gcp.service_account.private_key", "synthetic-private-key"},
				{"service_account_email", "gcp.service_account.service_account_email", "sa@example.iam.gserviceaccount.com"},
			},
		},
		{
			name: "HTTP",
			template: `http {
  base_url = "https://example.com"
  token_auth {
    token {
      name = "token"
      type = "SECRET"
      %s
    }
  }
}`,
			fields: []actionConnectionAWSSecretTestField{{"value", "http.token_auth.token.0.value", "synthetic-token"}},
		},
	}
	for _, tc := range cases {
		if tc.name != integration {
			continue
		}
		ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
		name := uniqueEntityName(ctx, t)
		resourceName := "datadog_action_connection.secret_reference"
		var connectionID string
		config := func(reference bool, stage string) string {
			fields := make([]string, 0, len(tc.fields))
			for _, field := range tc.fields {
				if reference {
					fields = append(fields, fmt.Sprintf(`%s_aws_secrets_manager = {
  secret_id = "terraform-provider/synthetic/action-connection"
  key = %q
  version_stage = %q
}`, field.name, field.name, stage))
				} else {
					fields = append(fields, fmt.Sprintf("%s = %q", field.name, field.literal))
				}
			}
			return fmt.Sprintf(`
%s
resource "datadog_action_connection" "secret_reference" {
  name = %q
  external_secrets_manager = {
    connection_id = datadog_action_connection.aws_assume_role_conn.id
    region = "us-east-1"
  }
  %s
}`, testAWSAssumeRoleConnectionResourceConfig(name+"-reader"), name, fmt.Sprintf(tc.template, strings.Join(fields, "\n")))
		}
		check := func(reference bool, stage string) resource.TestCheckFunc {
			checks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttrWith(resourceName, "id", func(id string) error {
					if connectionID == "" {
						connectionID = id
					} else if connectionID != id {
						return fmt.Errorf("credential migration replaced connection %s with %s", connectionID, id)
					}
					return nil
				}),
				resource.TestCheckResourceAttrPair(resourceName, "external_secrets_manager.connection_id", "datadog_action_connection.aws_assume_role_conn", "id"),
				resource.TestCheckResourceAttr(resourceName, "external_secrets_manager.region", "us-east-1"),
			}
			for _, field := range tc.fields {
				if reference {
					checks = append(checks,
						resource.TestCheckNoResourceAttr(resourceName, field.path),
						resource.TestCheckResourceAttr(resourceName, field.path+"_aws_secrets_manager.secret_id", "terraform-provider/synthetic/action-connection"),
						resource.TestCheckResourceAttr(resourceName, field.path+"_aws_secrets_manager.key", field.name),
						resource.TestCheckResourceAttr(resourceName, field.path+"_aws_secrets_manager.version_stage", stage),
					)
				} else {
					checks = append(checks,
						resource.TestCheckResourceAttr(resourceName, field.path, field.literal),
						resource.TestCheckNoResourceAttr(resourceName, field.path+"_aws_secrets_manager.secret_id"),
					)
				}
			}
			return resource.ComposeTestCheckFunc(checks...)
		}
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: accProviders,
			CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
			Steps: []resource.TestStep{
				{Config: config(true, "AWSCURRENT"), Check: check(true, "AWSCURRENT")},
				{ResourceName: resourceName, ImportState: true, ImportStateVerify: true},
				{Config: config(true, "AWSPREVIOUS"), Check: check(true, "AWSPREVIOUS")},
				{Config: config(false, ""), Check: check(false, "")},
				{Config: config(true, "AWSCURRENT"), Check: check(true, "AWSCURRENT")},
			},
		})
	}
}

type actionConnectionAWSSecretTestField struct {
	name    string
	path    string
	literal string
}

func TestAccDatadogActionConnectionResource_AWS_AssumeRole(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.aws_assume_role_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAWSAssumeRoleConnectionResourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "aws.assume_role.account_id", testAWSAccountID),
					resource.TestCheckResourceAttr(resourceName, "aws.assume_role.role", testAWSRole),
					resource.TestCheckResourceAttrSet(resourceName, "aws.assume_role.principal_id"),
					resource.TestCheckResourceAttrSet(resourceName, "aws.assume_role.external_id"),
				),
			},
		},
	})
}

func TestAccDatadogActionConnectionResource_Tags(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.aws_assume_role_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAWSAssumeRoleConnectionResourceConfigWithTags(connectionName, `["env:test", "team:action-platform"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "env:test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "team:action-platform"),
				),
			},
			{
				Config: testAWSAssumeRoleConnectionResourceConfigWithTags(connectionName, `["env:prod", "team:action-platform"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "env:prod"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "team:action-platform"),
				),
			},
		},
	})
}

func TestAccDatadogActionConnectionResource_DefaultTags(t *testing.T) {
	t.Parallel()

	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.aws_assume_role_conn"

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAWSAssumeRoleConnectionResourceConfigWithTags(connectionName, `["env:test"]`),
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"env":        "default",
						"managed_by": "terraform",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "env:test"),
					resource.TestCheckResourceAttr(resourceName, "effective_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "effective_tags.*", "env:test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "effective_tags.*", "managed_by:terraform"),
				),
			},
			{
				Config: testAWSAssumeRoleConnectionResourceConfigWithTags(connectionName, `["env:test"]`),
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"datadog": withDefaultTagsFw(ctx, providers, map[string]string{
						"env":        "default",
						"managed_by": "opentofu",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "env:test"),
					resource.TestCheckResourceAttr(resourceName, "effective_tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "effective_tags.*", "env:test"),
					resource.TestCheckTypeSetElemAttr(resourceName, "effective_tags.*", "managed_by:opentofu"),
				),
			},
		},
	})
}

func TestAccDatadogActionConnectionResource_HTTP_TokenAuth(t *testing.T) {
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.http_token_auth_conn"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testHTTPTokenAuthConnectionResourceConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "http.base_url", testConnectionBaseURL),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.token.0.type", "SECRET"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.token.0.name", "token1"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.token.0.value", "secret value 1"), // retrieved from state
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.header.0.name", "header-name"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.header.0.value", "header-value"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.url_parameter.0.name", "urlParamName"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.url_parameter.0.value", "urlParamValue"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.body.content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "http.token_auth.body.content", "{\"key\":\"{{ token1 }}\"}"),
				),
			},
		},
	})
}

func testHTTPTokenAuthConnectionResourceConfig(name string) string {
	return fmt.Sprintf(`
	resource "datadog_action_connection" "http_token_auth_conn" {
		name = "%s"

		http {
			base_url = "%s"

			token_auth {
				token {
					type = "SECRET"
					name = "token1"
					value = "secret value 1"
				}

				header {
					name = "header-name"
					value = "header-value"
				}

				url_parameter {
					name = "urlParamName"
					value = "urlParamValue"
				}

				body {
					content_type = "application/json"
					content = jsonencode({
						key = "{{ token1 }}"
					})
				}
			}
		}
	}`, name, testConnectionBaseURL)
}

func testAccCheckDatadogConnectionExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogConnectionExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogConnectionExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := apiInstances.GetActionConnectionApiV2().GetActionConnection(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving connection: %s", err)
	}
	return nil
}

func testAccActionConnectionIntegration(t *testing.T, integration string, attrs map[string]string) {
	t.Helper()
	t.Parallel()

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	connectionName := uniqueEntityName(ctx, t)
	resourceName := "datadog_action_connection.test"

	checks := []resource.TestCheckFunc{
		testAccCheckDatadogConnectionExists(providers.frameworkProvider, resourceName),
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", connectionName),
	}
	for attr, value := range attrs {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, attr, value))
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogConnectionDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_action_connection" "test" {
	name = "%s"
%s
}`, connectionName, integration),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
		},
	})
}

func TestAccDatadogActionConnectionResource_Anthropic(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	anthropic {
		api_key {
			api_token = "anthropic-token"
		}
	}`, map[string]string{"anthropic.api_key.api_token": "anthropic-token"})
}

func TestAccDatadogActionConnectionResource_Asana(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	asana {
		access_token {
			access_token = "asana-access-token"
		}
	}`, map[string]string{"asana.access_token.access_token": "asana-access-token"})
}

func TestAccDatadogActionConnectionResource_Azure(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	azure {
		tenant {
			app_client_id = "azure-client-id"
			client_secret = "azure-client-secret"
			custom_scopes = "https://management.azure.com/.default"
			tenant_id     = "azure-tenant-id"
		}
	}`, map[string]string{
		"azure.tenant.app_client_id": "azure-client-id",
		"azure.tenant.client_secret": "azure-client-secret",
		"azure.tenant.custom_scopes": "https://management.azure.com/.default",
		"azure.tenant.tenant_id":     "azure-tenant-id",
	})
}

func TestAccDatadogActionConnectionResource_CircleCI(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	circle_ci {
		api_key {
			api_token = "circle-ci-token"
		}
	}`, map[string]string{"circle_ci.api_key.api_token": "circle-ci-token"})
}

func TestAccDatadogActionConnectionResource_Clickup(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	clickup {
		api_key {
			api_token = "clickup-token"
		}
	}`, map[string]string{"clickup.api_key.api_token": "clickup-token"})
}

func TestAccDatadogActionConnectionResource_Cloudflare_APIToken(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	cloudflare {
		api_token {
			api_token = "cloudflare-token"
		}
	}`, map[string]string{"cloudflare.api_token.api_token": "cloudflare-token"})
}

func TestAccDatadogActionConnectionResource_Cloudflare_GlobalAPIToken(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	cloudflare {
		global_api_token {
			auth_email     = "user@example.com"
			global_api_key = "cloudflare-global-key"
		}
	}`, map[string]string{
		"cloudflare.global_api_token.auth_email":     "user@example.com",
		"cloudflare.global_api_token.global_api_key": "cloudflare-global-key",
	})
}

func TestAccDatadogActionConnectionResource_ConfigCat(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	config_cat {
		sdk_key {
			api_password = "configcat-password"
			api_username = "configcat-user"
			sdk_key      = "configcat-sdk-key"
		}
	}`, map[string]string{
		"config_cat.sdk_key.api_password": "configcat-password",
		"config_cat.sdk_key.api_username": "configcat-user",
		"config_cat.sdk_key.sdk_key":      "configcat-sdk-key",
	})
}

func TestAccDatadogActionConnectionResource_Datadog(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	datadog {
		api_key {
			api_key    = "datadog-api-key"
			app_key    = "datadog-app-key"
			datacenter = "us1.datadoghq.com"
			subdomain  = "example"
		}
	}`, map[string]string{
		"datadog.api_key.api_key":    "datadog-api-key",
		"datadog.api_key.app_key":    "datadog-app-key",
		"datadog.api_key.datacenter": "us1.datadoghq.com",
		"datadog.api_key.subdomain":  "example",
	})
}

func TestAccDatadogActionConnectionResource_Fastly(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	fastly {
		api_key {
			api_key = "fastly-api-key"
		}
	}`, map[string]string{"fastly.api_key.api_key": "fastly-api-key"})
}

func TestAccDatadogActionConnectionResource_Freshservice(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	freshservice {
		api_key {
			api_key = "freshservice-api-key"
			domain  = "example.freshservice.com"
		}
	}`, map[string]string{
		"freshservice.api_key.api_key": "freshservice-api-key",
		"freshservice.api_key.domain":  "example.freshservice.com",
	})
}

func TestAccDatadogActionConnectionResource_GCP(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	gcp {
		service_account {
			private_key           = "gcp-private-key"
			service_account_email = "sa@example.iam.gserviceaccount.com"
		}
	}`, map[string]string{
		"gcp.service_account.private_key":           "gcp-private-key",
		"gcp.service_account.service_account_email": "sa@example.iam.gserviceaccount.com",
	})
}

func TestAccDatadogActionConnectionResource_Gemini(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	gemini {
		api_key {
			api_key = "gemini-api-key"
		}
	}`, map[string]string{"gemini.api_key.api_key": "gemini-api-key"})
}

func TestAccDatadogActionConnectionResource_Gitlab(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	gitlab {
		api_key {
			api_token = "gitlab-token"
		}
	}`, map[string]string{"gitlab.api_key.api_token": "gitlab-token"})
}

func TestAccDatadogActionConnectionResource_GreyNoise(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	grey_noise {
		api_key {
			api_key = "greynoise-api-key"
		}
	}`, map[string]string{"grey_noise.api_key.api_key": "greynoise-api-key"})
}

func TestAccDatadogActionConnectionResource_LaunchDarkly(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	launch_darkly {
		api_key {
			api_token = "launchdarkly-token"
		}
	}`, map[string]string{"launch_darkly.api_key.api_token": "launchdarkly-token"})
}

func TestAccDatadogActionConnectionResource_Notion(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	notion {
		api_key {
			api_token = "notion-token"
		}
	}`, map[string]string{"notion.api_key.api_token": "notion-token"})
}

func TestAccDatadogActionConnectionResource_Okta(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	okta {
		api_token {
			api_token = "okta-token"
			domain    = "example.okta.com"
		}
	}`, map[string]string{
		"okta.api_token.api_token": "okta-token",
		"okta.api_token.domain":    "example.okta.com",
	})
}

func TestAccDatadogActionConnectionResource_OpenAI(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	openai {
		api_key {
			api_token = "openai-token"
		}
	}`, map[string]string{"openai.api_key.api_token": "openai-token"})
}

func TestAccDatadogActionConnectionResource_ServiceNow(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	service_now {
		basic_auth {
			instance = "example-instance"
			password = "service-now-password"
			username = "service-now-user"
		}
	}`, map[string]string{
		"service_now.basic_auth.instance": "example-instance",
		"service_now.basic_auth.password": "service-now-password",
		"service_now.basic_auth.username": "service-now-user",
	})
}

func TestAccDatadogActionConnectionResource_Split(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	split {
		api_key {
			api_key = "split-api-key"
		}
	}`, map[string]string{"split.api_key.api_key": "split-api-key"})
}

func TestAccDatadogActionConnectionResource_Statsig(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	statsig {
		api_key {
			api_key = "statsig-api-key"
		}
	}`, map[string]string{"statsig.api_key.api_key": "statsig-api-key"})
}

func TestAccDatadogActionConnectionResource_VirusTotal(t *testing.T) {
	testAccActionConnectionIntegration(t, `
	virus_total {
		api_key {
			api_key = "virustotal-api-key"
		}
	}`, map[string]string{"virus_total.api_key.api_key": "virustotal-api-key"})
}
