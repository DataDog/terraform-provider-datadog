package fwprovider

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestActionConnectionPlanPreservesTagsOnlySetOutsideTerraform(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uiTags, diags := types.SetValueFrom(ctx, types.StringType, []string{"shimu:test", "test:shimu"})
	if diags.HasError() {
		t.Fatalf("building UI tags: %v", diags.Errors())
	}
	emptyTags, diags := types.SetValueFrom(ctx, types.StringType, []string{})
	if diags.HasError() {
		t.Fatalf("building empty tags: %v", diags.Errors())
	}

	tests := []struct {
		name        string
		tags        types.Set
		defaultTags map[string]string
		wantTags    []string
	}{
		{name: "no resource or provider tags keeps UI tags", tags: types.SetNull(types.StringType), wantTags: []string{"shimu:test", "test:shimu"}},
		{name: "explicit empty tags clear UI tags", tags: emptyTags, wantTags: []string{}},
		{name: "provider default tags own the tag set", tags: types.SetNull(types.StringType), defaultTags: map[string]string{"team": "workflow-automation"}, wantTags: []string{"team:workflow-automation"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &actionConnectionResource{DefaultTags: tt.defaultTags}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

			stateModel := connectionResourceModel{
				connectionModel: connectionModel{ID: types.StringValue("connection-id"), Name: types.StringValue("connection"), Tags: types.SetNull(types.StringType)},
				EffectiveTags:   uiTags,
			}
			planModel := stateModel
			planModel.Tags = tt.tags
			planModel.EffectiveTags = types.SetUnknown(types.StringType)

			state := tfsdk.State{Schema: schemaResponse.Schema}
			if diags := state.Set(ctx, &stateModel); diags.HasError() {
				t.Fatalf("building state: %v", diags.Errors())
			}
			plan := tfsdk.Plan{Schema: schemaResponse.Schema}
			if diags := plan.Set(ctx, &planModel); diags.HasError() {
				t.Fatalf("building plan: %v", diags.Errors())
			}

			response := resource.ModifyPlanResponse{Plan: plan}
			r.ModifyPlan(ctx, resource.ModifyPlanRequest{Plan: plan, State: state}, &response)
			if response.Diagnostics.HasError() {
				t.Fatalf("modifying plan: %v", response.Diagnostics.Errors())
			}

			var effectiveTags types.Set
			if diags := response.Plan.GetAttribute(ctx, path.Root("effective_tags"), &effectiveTags); diags.HasError() {
				t.Fatalf("reading effective tags: %v", diags.Errors())
			}
			var gotTags []string
			if diags := effectiveTags.ElementsAs(ctx, &gotTags, false); diags.HasError() {
				t.Fatalf("reading effective tag values: %v", diags.Errors())
			}
			slices.Sort(gotTags)
			if !slices.Equal(gotTags, tt.wantTags) {
				t.Fatalf("effective tags = %v, want %v", gotTags, tt.wantTags)
			}
		})
	}
}

func TestActionConnectionMetadataUpdatePreservesCredentials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tags, diags := types.SetValueFrom(ctx, types.StringType, []string{"team:workflow"})
	if diags.HasError() {
		t.Fatalf("building tags: %v", diags.Errors())
	}

	tests := []struct {
		name            string
		connection      connectionModel
		changeName      bool
		unknownAWSData  bool
		wantIntegration bool
	}{
		{
			name: "HTTP tag update after credentials were entered outside Terraform",
			connection: connectionModel{
				Name: types.StringValue("HTTP connection"),
				HTTP: &httpConnectionModel{TokenAuth: &httpTokenAuthConnectionModel{
					Tokens: []*httpConnectionTokenModel{{
						Type: types.StringValue("SECRET"), Name: types.StringValue("apiKey"), Value: types.StringValue("PLACEHOLDER_FILL_IN_UI"),
					}},
				}},
			},
		},
		{
			name: "AWS tag update includes integration",
			connection: connectionModel{
				Name: types.StringValue("AWS connection"),
				AWS: &awsConnectionModel{AssumeRole: &awsAssumeRoleConnectionModel{
					AccountID: types.StringValue("123456789012"), Role: types.StringValue("role"),
					ExternalID: types.StringValue("external-id"), PrincipalID: types.StringValue("principal-id"),
				}},
			},
			unknownAWSData:  true,
			wantIntegration: true,
		},
		{
			name: "Datadog tag update",
			connection: connectionModel{
				Name: types.StringValue("Datadog connection"),
				Datadog: &datadogConnectionModel{APIKey: &datadogAPIKeyCredentialModel{
					APIKey: types.StringValue("placeholder-api-key"), AppKey: types.StringValue("placeholder-app-key"),
					Datacenter: types.StringValue("us1"),
				}},
			},
		},
		{
			name: "HTTP name update",
			connection: connectionModel{
				Name: types.StringValue("HTTP connection"),
				HTTP: &httpConnectionModel{TokenAuth: &httpTokenAuthConnectionModel{
					Tokens: []*httpConnectionTokenModel{{
						Type: types.StringValue("SECRET"), Name: types.StringValue("apiKey"), Value: types.StringValue("PLACEHOLDER_FILL_IN_UI"),
					}},
				}},
			},
			changeName: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prior := connectionResourceModel{connectionModel: tt.connection, EffectiveTags: types.SetNull(types.StringType)}
			plan := prior
			if tt.unknownAWSData {
				plan.AWS = &awsConnectionModel{AssumeRole: &awsAssumeRoleConnectionModel{
					AccountID: prior.AWS.AssumeRole.AccountID, Role: prior.AWS.AssumeRole.Role,
					ExternalID: types.StringUnknown(), PrincipalID: types.StringUnknown(),
				}}
			}
			if tt.changeName {
				plan.Name = types.StringValue("renamed HTTP connection")
			} else {
				plan.EffectiveTags = tags
			}

			request, err := connectionModelToUpdateApiRequest(ctx, plan, prior)
			if err != nil {
				t.Fatalf("building update: %v", err)
			}
			payload, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("marshaling update: %v", err)
			}
			var body struct {
				Data struct {
					Attributes map[string]json.RawMessage `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(payload, &body); err != nil {
				t.Fatalf("parsing update: %v", err)
			}
			if _, exists := body.Data.Attributes["integration"]; exists != tt.wantIntegration {
				t.Fatalf("integration present = %t, want %t", exists, tt.wantIntegration)
			}
			if _, exists := body.Data.Attributes["tags"]; exists == tt.changeName {
				t.Fatalf("tags present = %t, want %t", exists, !tt.changeName)
			}
			if _, exists := body.Data.Attributes["name"]; !exists {
				t.Fatal("metadata PATCH omitted name")
			}
		})
	}
}

func TestActionConnectionIntegrationUpdateIncludesChangedCredentials(t *testing.T) {
	t.Parallel()

	prior := connectionResourceModel{connectionModel: connectionModel{
		Name: types.StringValue("HTTP connection"),
		HTTP: &httpConnectionModel{TokenAuth: &httpTokenAuthConnectionModel{
			Tokens: []*httpConnectionTokenModel{{
				Type: types.StringValue("SECRET"), Name: types.StringValue("apiKey"), Value: types.StringValue("old token"),
			}},
		}},
	}}
	plan := prior
	plan.HTTP = &httpConnectionModel{TokenAuth: &httpTokenAuthConnectionModel{
		Tokens: []*httpConnectionTokenModel{{
			Type: types.StringValue("SECRET"), Name: types.StringValue("apiKey"), Value: types.StringValue("new token"),
		}},
	}}

	request, err := connectionModelToUpdateApiRequest(context.Background(), plan, prior)
	if err != nil {
		t.Fatalf("building update: %v", err)
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshaling update: %v", err)
	}
	var body struct {
		Data struct {
			Attributes map[string]json.RawMessage `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatalf("parsing update: %v", err)
	}
	if _, exists := body.Data.Attributes["integration"]; !exists {
		t.Fatal("credential change did not include integration in PATCH")
	}
}

func TestActionConnectionHTTPUpdateOmitsUnchangedTokens(t *testing.T) {
	t.Parallel()

	oldAuth := &httpTokenAuthConnectionModel{
		Tokens: []*httpConnectionTokenModel{
			{Type: types.StringValue("SECRET"), Name: types.StringValue("apiKey"), Value: types.StringValue("PLACEHOLDER_FILL_IN_UI")},
			{Type: types.StringValue("SECRET"), Name: types.StringValue("appKey"), Value: types.StringValue("PLACEHOLDER_FILL_IN_UI")},
		},
		Headers: []*httpConnectionHeaderModel{{Name: types.StringValue("X-Test"), Value: types.StringValue("old")}},
	}
	oldHTTP := &httpConnectionModel{BaseURL: types.StringValue("https://example.com"), TokenAuth: oldAuth}
	newHeaderAuth := *oldAuth
	newHeaderAuth.Headers = []*httpConnectionHeaderModel{{Name: types.StringValue("X-Test"), Value: types.StringValue("new")}}
	newTokenAuth := *oldAuth
	newTokenAuth.Tokens = []*httpConnectionTokenModel{
		{Type: types.StringValue("SECRET"), Name: types.StringValue("apiKey"), Value: types.StringValue("new key")},
		oldAuth.Tokens[1],
	}
	deletedTokenAuth := *oldAuth
	deletedTokenAuth.Tokens = oldAuth.Tokens[1:]

	tests := []struct {
		name            string
		plannedHTTP     *httpConnectionModel
		wantBaseURL     bool
		wantCredentials bool
		wantHeaders     bool
		wantTokens      int
		wantDeleted     bool
	}{
		{name: "base URL", plannedHTTP: &httpConnectionModel{BaseURL: types.StringValue("https://new.example.com"), TokenAuth: oldAuth}, wantBaseURL: true, wantCredentials: true},
		{name: "header", plannedHTTP: &httpConnectionModel{BaseURL: oldHTTP.BaseURL, TokenAuth: &newHeaderAuth}, wantCredentials: true, wantHeaders: true},
		{name: "one token changed", plannedHTTP: &httpConnectionModel{BaseURL: oldHTTP.BaseURL, TokenAuth: &newTokenAuth}, wantCredentials: true, wantTokens: 1},
		{name: "one token deleted", plannedHTTP: &httpConnectionModel{BaseURL: oldHTTP.BaseURL, TokenAuth: &deletedTokenAuth}, wantCredentials: true, wantTokens: 1, wantDeleted: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prior := connectionResourceModel{connectionModel: connectionModel{Name: types.StringValue("connection"), HTTP: oldHTTP}}
			plan := connectionResourceModel{connectionModel: connectionModel{Name: prior.Name, HTTP: tt.plannedHTTP}}
			request, err := connectionModelToUpdateApiRequest(context.Background(), plan, prior)
			if err != nil {
				t.Fatalf("building update: %v", err)
			}
			payload, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("marshaling update: %v", err)
			}
			var body struct {
				Data struct {
					Attributes struct {
						Integration map[string]json.RawMessage `json:"integration"`
					} `json:"attributes"`
				} `json:"data"`
			}
			if err := json.Unmarshal(payload, &body); err != nil {
				t.Fatalf("parsing update: %v", err)
			}
			integration := body.Data.Attributes.Integration
			if _, exists := integration["base_url"]; exists != tt.wantBaseURL {
				t.Fatalf("base_url present = %t, want %t", exists, tt.wantBaseURL)
			}
			credentialsJSON, exists := integration["credentials"]
			if exists != tt.wantCredentials {
				t.Fatalf("credentials present = %t, want %t", exists, tt.wantCredentials)
			}
			if !exists {
				return
			}
			var credentials map[string]json.RawMessage
			if err := json.Unmarshal(credentialsJSON, &credentials); err != nil {
				t.Fatalf("parsing credentials: %v", err)
			}
			if string(credentials["type"]) != `"HTTPTokenAuth"` {
				t.Fatal("HTTP update did not include credential type")
			}
			if _, exists := credentials["headers"]; exists != tt.wantHeaders {
				t.Fatalf("headers present = %t, want %t", exists, tt.wantHeaders)
			}
			var tokens []json.RawMessage
			if raw, exists := credentials["tokens"]; exists {
				if err := json.Unmarshal(raw, &tokens); err != nil {
					t.Fatalf("parsing tokens: %v", err)
				}
			}
			if len(tokens) != tt.wantTokens {
				t.Fatalf("sent %d tokens, want %d", len(tokens), tt.wantTokens)
			}
			if len(tokens) == 1 {
				var token struct {
					Name    string `json:"name"`
					Deleted bool   `json:"deleted"`
				}
				if err := json.Unmarshal(tokens[0], &token); err != nil {
					t.Fatalf("parsing token update: %v", err)
				}
				if token.Name != "apiKey" || token.Deleted != tt.wantDeleted {
					t.Fatalf("unexpected token update: name %q, deleted %t", token.Name, token.Deleted)
				}
			}
		})
	}
}
