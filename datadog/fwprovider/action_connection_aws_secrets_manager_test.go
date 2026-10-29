package fwprovider

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func actionConnectionTestAWSReference(key string) *actionConnectionAWSSecretReferenceModel {
	return &actionConnectionAWSSecretReferenceModel{SecretID: types.StringValue("example/credentials"), Key: types.StringValue(key)}
}

func actionConnectionTestAWSSettings() *actionConnectionExternalSecretsManagerModel {
	return &actionConnectionExternalSecretsManagerModel{ConnectionID: types.StringValue("00000000-0000-0000-0000-000000000001"), Region: types.StringValue("us-east-1")}
}

func TestActionConnectionAWSSecretsManagerSupportedIntegrations(t *testing.T) {
	ctx := context.Background()
	for _, integration := range additionalActionConnectionSpecs {
		if integration.Name == "azure" {
			continue
		}
		for _, credential := range integration.Credentials {
			t.Run(integration.Name+"/"+credential.Name, func(t *testing.T) {
				literal := connectionResourceModel{connectionModel: connectionModel{Name: types.StringValue("Example connection")}}
				field, _ := lookupTFSDKField(reflect.ValueOf(&literal.connectionModel), integration.Name)
				field.Set(reflect.New(field.Type().Elem()))
				credentialField, _ := lookupTFSDKField(field, credential.Name)
				credentialField.Set(reflect.New(credentialField.Type().Elem()))
				for _, spec := range credential.Fields {
					value, _ := lookupTFSDKField(credentialField, spec.Name)
					value.Set(reflect.ValueOf(types.StringValue("literal")))
				}
				referenced := literal
				referenced.ExternalSecretsManager = actionConnectionTestAWSSettings()
				// Allocate independent credentials to test migration from literals.
				root, _ := lookupTFSDKField(reflect.ValueOf(&referenced.connectionModel), integration.Name)
				root.Set(reflect.New(root.Type().Elem()))
				credentials, _ := lookupTFSDKField(root, credential.Name)
				credentials.Set(reflect.New(credentials.Type().Elem()))
				for _, spec := range credential.Fields {
					ref, _ := lookupTFSDKField(credentials, spec.Name+actionConnectionAWSSecretSuffix)
					ref.Set(reflect.ValueOf(actionConnectionTestAWSReference(spec.Name)))
				}
				var validation resource.ValidateConfigResponse
				validateAdditionalConnectionCredentials(referenced, &validation)
				validateActionConnectionSecretReferences(referenced, &validation)
				if validation.Diagnostics.HasError() {
					t.Fatalf("reference-only credentials are invalid: %v", validation.Diagnostics)
				}
				create, err := connectionModelToCreateApiRequest(ctx, referenced)
				if err != nil {
					t.Fatal(err)
				}
				assertActionConnectionFullyParsed(t, reflect.ValueOf(create))
				if !create.Data.Attributes.HasExternalSecretsManager() {
					t.Fatal("AWS settings missing")
				}
				response := datadogV2.GetActionConnectionResponse{Data: &create.Data}
				read, err := apiResponseToConnectionModel(ctx, response)
				if err != nil {
					t.Fatal(err)
				}
				preserveAdditionalConnectionSecrets(read, literal)
				assertActionConnectionReferenceCount(t, read.connectionModel, len(credential.Fields))
				recreated, err := connectionModelToCreateApiRequest(ctx, *read)
				if err != nil {
					t.Fatal(err)
				}
				assertActionConnectionEqualJSON(t, create.Data.Attributes.Integration, recreated.Data.Attributes.Integration)
				update, err := connectionModelToUpdateApiRequest(ctx, referenced, literal)
				if err != nil {
					t.Fatal(err)
				}
				assertActionConnectionFullyParsed(t, reflect.ValueOf(update))
				if update.Data.Attributes.Integration == nil {
					t.Fatal("migration omitted credential changes")
				}
				assertActionConnectionEqualJSON(t, create.Data.Attributes.Integration, update.Data.Attributes.Integration)
				// Read the reference-only connection into both resource and datasource states.
				var schemaResponse resource.SchemaResponse
				(&actionConnectionResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
				state := tfsdk.State{Schema: schemaResponse.Schema}
				if diags := state.Set(ctx, read); diags.HasError() {
					t.Fatalf("resource state: %v", diags)
				}
				var ds datasource.SchemaResponse
				(&actionConnectionDatasource{}).Schema(ctx, datasource.SchemaRequest{}, &ds)
				dsState := tfsdk.State{Schema: ds.Schema}
				if diags := dsState.Set(ctx, &connectionDatasourceModel{connectionModel: read.connectionModel}); diags.HasError() {
					t.Fatalf("datasource state: %v", diags)
				}
			})
		}
	}
}

func TestActionConnectionAWSSecretsManagerHTTPToken(t *testing.T) {
	ctx := context.Background()
	model := connectionModel{
		Name:                   types.StringValue("Example HTTP connection"),
		ExternalSecretsManager: actionConnectionTestAWSSettings(),
		HTTP: &httpConnectionModel{BaseURL: types.StringValue("https://example.com"), TokenAuth: &httpTokenAuthConnectionModel{
			Tokens:        []*httpConnectionTokenModel{{Type: types.StringValue("SECRET"), Name: types.StringValue("token"), ValueAWSSecretsManager: actionConnectionTestAWSReference("token")}},
			Headers:       []*httpConnectionHeaderModel{{Name: types.StringValue("Header"), Value: types.StringValue("literal-header")}},
			URLParameters: []*httpConnectionUrlParameterModel{{Name: types.StringValue("param"), Value: types.StringValue("literal-parameter")}},
			Body:          &httpConnectionBodyModel{Content: types.StringValue("literal-body"), ContentType: types.StringValue("text/plain")},
		}},
	}
	resourceModel := connectionResourceModel{connectionModel: model}
	create, err := connectionModelToCreateApiRequest(ctx, resourceModel)
	if err != nil {
		t.Fatal(err)
	}
	assertActionConnectionFullyParsed(t, reflect.ValueOf(create))
	read, err := apiResponseToConnectionModel(ctx, datadogV2.GetActionConnectionResponse{Data: &create.Data})
	if err != nil {
		t.Fatal(err)
	}
	assertActionConnectionReferenceCount(t, read.connectionModel, 1)
	recreated, err := connectionModelToCreateApiRequest(ctx, *read)
	if err != nil {
		t.Fatal(err)
	}
	assertActionConnectionEqualJSON(t, create.Data.Attributes.Integration, recreated.Data.Attributes.Integration)
	update, err := connectionModelToUpdateApiRequest(ctx, resourceModel, connectionResourceModel{connectionModel: connectionModel{Name: model.Name}})
	if err != nil {
		t.Fatal(err)
	}
	assertActionConnectionEqualJSON(t, create.Data.Attributes.Integration, update.Data.Attributes.Integration)
}

func TestActionConnectionAWSSecretsManagerSupportedFieldSchema(t *testing.T) {
	ctx := context.Background()
	var resourceSchema resource.SchemaResponse
	(&actionConnectionResource{}).Schema(ctx, resource.SchemaRequest{}, &resourceSchema)
	var dataSourceSchema datasource.SchemaResponse
	(&actionConnectionDatasource{}).Schema(ctx, datasource.SchemaRequest{}, &dataSourceSchema)
	unsupported := []path.Path{
		path.Root("aws").AtName("assume_role").AtName("account_id_aws_secrets_manager"),
		path.Root("aws").AtName("assume_role").AtName("role_aws_secrets_manager"),
		path.Root("azure").AtName("tenant").AtName("app_client_id_aws_secrets_manager"),
		path.Root("azure").AtName("tenant").AtName("client_secret_aws_secrets_manager"),
		path.Root("azure").AtName("tenant").AtName("custom_scopes_aws_secrets_manager"),
		path.Root("azure").AtName("tenant").AtName("tenant_id_aws_secrets_manager"),
		path.Root("http").AtName("base_url_aws_secrets_manager"),
		path.Root("http").AtName("token_auth").AtName("header").AtListIndex(0).AtName("value_aws_secrets_manager"),
		path.Root("http").AtName("token_auth").AtName("url_parameter").AtListIndex(0).AtName("value_aws_secrets_manager"),
		path.Root("http").AtName("token_auth").AtName("body").AtName("content_aws_secrets_manager"),
		path.Root("http").AtName("token_auth").AtName("body").AtName("content_type_aws_secrets_manager"),
	}
	for _, fieldPath := range unsupported {
		if _, diags := resourceSchema.Schema.AttributeAtPath(ctx, fieldPath); !diags.HasError() {
			t.Errorf("unsupported resource reference: %s", fieldPath)
		}
		if _, diags := dataSourceSchema.Schema.AttributeAtPath(ctx, fieldPath); !diags.HasError() {
			t.Errorf("unsupported datasource reference: %s", fieldPath)
		}
	}
	for _, fieldPath := range []path.Path{
		path.Root("gcp").AtName("service_account").AtName("private_key_aws_secrets_manager"),
		path.Root("datadog").AtName("api_key").AtName("datacenter_aws_secrets_manager"),
		path.Root("http").AtName("token_auth").AtName("token").AtListIndex(0).AtName("value_aws_secrets_manager"),
	} {
		if _, diags := resourceSchema.Schema.AttributeAtPath(ctx, fieldPath); diags.HasError() {
			t.Errorf("missing supported resource reference: %s: %v", fieldPath, diags)
		}
		if _, diags := dataSourceSchema.Schema.AttributeAtPath(ctx, fieldPath); diags.HasError() {
			t.Errorf("missing supported datasource reference: %s: %v", fieldPath, diags)
		}
	}
}

func TestActionConnectionAWSSecretsManagerHTTPResponseImport(t *testing.T) {
	ctx := context.Background()
	payload := `{"data":{"id":"00000000-0000-0000-0000-000000000002","type":"action_connection","attributes":{"name":"Example HTTP connection","external_secrets_manager":{"type":"aws_secrets_manager","connectionId":"00000000-0000-0000-0000-000000000001","region":"us-east-1"},"integration":{"type":"HTTP","base_url":"https://example.com","credentials":{"type":"HTTPTokenAuth","http_tokens":[{"name":"token","value":{"type":"aws_secrets_manager","secret_id":"example/http","key":"token","version_stage":"AWSCURRENT"}}]}}}}}`
	var response datadogV2.GetActionConnectionResponse
	if err := json.Unmarshal([]byte(payload), &response); err != nil {
		t.Fatal(err)
	}
	assertActionConnectionFullyParsed(t, reflect.ValueOf(response))
	model, err := apiResponseToConnectionModel(ctx, response)
	if err != nil {
		t.Fatal(err)
	}
	assertActionConnectionReferenceCount(t, model.connectionModel, 1)
	token := model.HTTP.TokenAuth.Tokens[0]
	if token.Type.ValueString() != "SECRET" || token.ValueAWSSecretsManager.VersionStage.ValueString() != "AWSCURRENT" {
		t.Fatalf("import lost HTTP token type or version selector: %v", token)
	}
	request, err := connectionModelToCreateApiRequest(ctx, *model)
	if err != nil {
		t.Fatal(err)
	}
	credentials := request.Data.Attributes.Integration.HTTPIntegration.Credentials.HTTPTokenAuth
	if len(credentials.Tokens) != 1 || len(credentials.HttpTokens) != 0 {
		t.Fatal("requests must use tokens while responses use http_tokens")
	}
	if credentials.Tokens[0].Value.ActionConnectionAWSSecretsManagerReference.GetVersionStage() != "AWSCURRENT" {
		t.Fatal("request lost HTTP token version selector")
	}
}

func TestActionConnectionAWSSecretsManagerValidation(t *testing.T) {
	cases := []struct {
		name      string
		literal   types.String
		settings  *actionConnectionExternalSecretsManagerModel
		reference *actionConnectionAWSSecretReferenceModel
		wantError bool
	}{
		{name: "reference only", settings: actionConnectionTestAWSSettings(), reference: actionConnectionTestAWSReference("token")},
		{name: "literal only", literal: types.StringValue("literal")},
		{name: "settings without a reference", settings: actionConnectionTestAWSSettings(), literal: types.StringValue("literal"), wantError: true},
		{name: "missing settings", reference: actionConnectionTestAWSReference("token"), wantError: true},
		{name: "literal and reference", literal: types.StringValue("literal"), settings: actionConnectionTestAWSSettings(), reference: actionConnectionTestAWSReference("token"), wantError: true},
		{name: "conflicting versions", settings: actionConnectionTestAWSSettings(), reference: &actionConnectionAWSSecretReferenceModel{SecretID: types.StringValue("example"), VersionID: types.StringValue("version"), VersionStage: types.StringValue("AWSCURRENT")}, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn := connectionResourceModel{connectionModel: connectionModel{ExternalSecretsManager: tc.settings, Anthropic: &apiTokenConnectionModel{APIKey: &apiTokenCredentialModel{APIToken: tc.literal, APITokenAWSSecretsManager: tc.reference}}}}
			var response resource.ValidateConfigResponse
			validateActionConnectionSecretReferences(conn, &response)
			if response.Diagnostics.HasError() != tc.wantError {
				t.Fatalf("unexpected diagnostics: %v", response.Diagnostics)
			}
		})
	}
}

func TestActionConnectionAWSSecretsManagerMigrationDoesNotReplace(t *testing.T) {
	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	(&actionConnectionResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	literal := connectionResourceModel{connectionModel: connectionModel{
		Name:    types.StringValue("Example Datadog connection"),
		Datadog: &datadogConnectionModel{APIKey: &datadogAPIKeyCredentialModel{Subdomain: types.StringValue("example")}},
	}}
	referenced := connectionResourceModel{connectionModel: connectionModel{
		Name:                   types.StringValue("Example Datadog connection"),
		ExternalSecretsManager: actionConnectionTestAWSSettings(),
		Datadog:                &datadogConnectionModel{APIKey: &datadogAPIKeyCredentialModel{SubdomainAWSSecretsManager: actionConnectionTestAWSReference("subdomain")}},
	}}
	literal.Tags = types.SetNull(types.StringType)
	literal.EffectiveTags = types.SetNull(types.StringType)
	referenced.Tags = types.SetNull(types.StringType)
	referenced.EffectiveTags = types.SetNull(types.StringType)
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &referenced); diags.HasError() {
		t.Fatal(diags)
	}
	fieldPath := path.Root("datadog").AtName("api_key").AtName("subdomain")
	var literalResponse stringplanmodifier.RequiresReplaceIfFuncResponse
	replaceActionConnectionIfFieldRemoved(ctx, planmodifier.StringRequest{
		Path: fieldPath, Plan: plan, StateValue: types.StringValue("scope"), PlanValue: types.StringNull(),
	}, &literalResponse)
	if literalResponse.Diagnostics.HasError() || literalResponse.RequiresReplace {
		t.Fatalf("literal to reference migration requires replacement: %v", literalResponse.Diagnostics)
	}
	var stateReference types.Object
	if diags := plan.GetAttribute(ctx, fieldPath.ParentPath().AtName("subdomain_aws_secrets_manager"), &stateReference); diags.HasError() {
		t.Fatal(diags)
	}
	if diags := plan.Set(ctx, &literal); diags.HasError() {
		t.Fatal(diags)
	}
	var referenceResponse objectplanmodifier.RequiresReplaceIfFuncResponse
	replaceActionConnectionIfReferenceRemoved(ctx, planmodifier.ObjectRequest{
		Path: fieldPath.ParentPath().AtName("subdomain_aws_secrets_manager"), Plan: plan,
		StateValue: stateReference, PlanValue: types.ObjectNull(stateReference.AttributeTypes(ctx)),
	}, &referenceResponse)
	if referenceResponse.Diagnostics.HasError() || referenceResponse.RequiresReplace {
		t.Fatalf("reference to literal migration requires replacement: %v", referenceResponse.Diagnostics)
	}
}

func assertActionConnectionReferenceCount(t *testing.T, model connectionModel, want int) {
	t.Helper()
	count := 0
	walkActionConnectionReferences(reflect.ValueOf(model), path.Empty(), func(_ reflect.Value, literal types.String, ref *actionConnectionAWSSecretReferenceModel, _ path.Path) {
		if ref != nil {
			count++
			if !literal.IsNull() {
				t.Error("reference has a literal value in state")
			}
		}
	})
	if count != want {
		t.Fatalf("read %d references, want %d", count, want)
	}
}

func assertActionConnectionEqualJSON(t *testing.T, want, got interface{}) {
	t.Helper()
	a, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var expected, actual interface{}
	if err := json.Unmarshal(a, &expected); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("JSON mismatch: want %s, got %s", a, b)
	}
}

func assertActionConnectionFullyParsed(t *testing.T, value reflect.Value) {
	t.Helper()
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !value.IsNil() {
			assertActionConnectionFullyParsed(t, value.Elem())
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			if value.Type().Field(i).Name == "UnparsedObject" && !field.IsNil() {
				t.Fatalf("%s contains UnparsedObject", value.Type())
			}
			assertActionConnectionFullyParsed(t, field)
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			assertActionConnectionFullyParsed(t, value.Index(i))
		}
	}
}
