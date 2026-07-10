package fwprovider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestActionConnectionAdditionalIntegrationSmoke(t *testing.T) {
	cases := []struct {
		integrationType string
		credentialType  string
		fields          map[string]types.String
	}{
		{"Anthropic", "AnthropicAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"Asana", "AsanaAccessToken", map[string]types.String{"access_token": types.StringValue("secret")}},
		{"Azure", "AzureTenant", map[string]types.String{"app_client_id": types.StringValue("id"), "client_secret": types.StringValue("secret"), "tenant_id": types.StringValue("id")}},
		{"CircleCI", "CircleCIAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"Clickup", "ClickupAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"Cloudflare", "CloudflareAPIToken", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"Cloudflare", "CloudflareGlobalAPIToken", map[string]types.String{"auth_email": types.StringValue("a@b.com"), "global_api_key": types.StringValue("secret")}},
		{"ConfigCat", "ConfigCatSDKKey", map[string]types.String{"api_password": types.StringValue("secret"), "api_username": types.StringValue("user"), "sdk_key": types.StringValue("secret")}},
		{"Datadog", "DatadogAPIKey", map[string]types.String{"api_key": types.StringValue("secret"), "app_key": types.StringValue("secret"), "datacenter": types.StringValue("us1")}},
		{"Fastly", "FastlyAPIKey", map[string]types.String{"api_key": types.StringValue("secret")}},
		{"Freshservice", "FreshserviceAPIKey", map[string]types.String{"api_key": types.StringValue("secret"), "domain": types.StringValue("example")}},
		{"GCP", "GCPServiceAccount", map[string]types.String{"private_key": types.StringValue("secret"), "service_account_email": types.StringValue("a@b.com")}},
		{"Gemini", "GeminiAPIKey", map[string]types.String{"api_key": types.StringValue("secret")}},
		{"Gitlab", "GitlabAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"GreyNoise", "GreyNoiseAPIKey", map[string]types.String{"api_key": types.StringValue("secret")}},
		{"LaunchDarkly", "LaunchDarklyAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"Notion", "NotionAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"Okta", "OktaAPIToken", map[string]types.String{"api_token": types.StringValue("secret"), "domain": types.StringValue("example")}},
		{"OpenAI", "OpenAIAPIKey", map[string]types.String{"api_token": types.StringValue("secret")}},
		{"ServiceNow", "ServiceNowBasicAuth", map[string]types.String{"instance": types.StringValue("example"), "password": types.StringValue("secret"), "username": types.StringValue("user")}},
		{"Split", "SplitAPIKey", map[string]types.String{"api_key": types.StringValue("secret")}},
		{"Statsig", "StatsigAPIKey", map[string]types.String{"api_key": types.StringValue("secret")}},
		{"VirusTotal", "VirusTotalAPIKey", map[string]types.String{"api_key": types.StringValue("secret")}},
	}

	for _, tc := range cases {
		t.Run(tc.integrationType+tc.credentialType, func(t *testing.T) {
			data, err := json.Marshal(actionConnectionIntegrationData(tc.integrationType, tc.credentialType, tc.fields))
			if err != nil {
				t.Fatal(err)
			}
			var create datadogV2.ActionConnectionIntegration
			if err := json.Unmarshal(data, &create); err != nil {
				t.Fatal(err)
			}
			if create.UnparsedObject != nil {
				t.Fatalf("create integration was not parsed: %s", data)
			}
			model := &connectionResourceModel{}
			if err := setAdditionalConnectionModelFromAPI(model, create); err != nil {
				t.Fatal(err)
			}
			if _, err := additionalCreateActionConnectionIntegration(*model); err != nil {
				t.Fatal(err)
			}
			if _, err := additionalUpdateActionConnectionIntegration(*model); err != nil {
				t.Fatal(err)
			}
		})
	}

	var schemaResponse resource.SchemaResponse
	(&actionConnectionResource{}).Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	if diags := schemaResponse.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("invalid schema: %v", diags)
	}
	state := tfsdk.State{Schema: schemaResponse.Schema}
	model := connectionResourceModel{
		ID:   types.StringValue("id"),
		Name: types.StringValue("name"),
		Anthropic: &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{
			APIToken: types.StringValue("secret"),
		}},
	}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("model does not match resource schema: %v", diags)
	}

	var dataSourceSchemaResponse datasource.SchemaResponse
	(&actionConnectionDatasource{}).Schema(context.Background(), datasource.SchemaRequest{}, &dataSourceSchemaResponse)
	if diags := dataSourceSchemaResponse.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("invalid data source schema: %v", diags)
	}
	dataSourceState := tfsdk.State{Schema: dataSourceSchemaResponse.Schema}
	if diags := dataSourceState.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("model does not match data source schema: %v", diags)
	}
}
