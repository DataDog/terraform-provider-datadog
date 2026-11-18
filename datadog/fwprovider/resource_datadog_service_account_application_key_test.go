package fwprovider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const testAppKeyRawCreatedAt = "2026-11-18T13:37:52.051772Z"

func newTestServiceAccountAppKeyResource(t *testing.T, getStatus int) (*serviceAccountApplicationKeyResource, resource.SchemaResponse) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "/application_keys"):
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"key-id","type":"application_keys","attributes":{"key":"secret-value","last4":"alue","name":"k","created_at":"` + testAppKeyRawCreatedAt + `"}}}`))
		case req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "/application_keys/key-id"):
			w.WriteHeader(getStatus)
			if getStatus == http.StatusOK {
				_, _ = w.Write([]byte(`{"data":{"id":"key-id","type":"application_keys","attributes":{"last4":"alue","name":"k","created_at":"` + testAppKeyRawCreatedAt + `"}}}`))
			} else {
				_, _ = w.Write([]byte(`{"errors":["Forbidden"]}`))
			}
		default:
			t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	config := datadog.NewConfiguration()
	config.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	config.OperationServers = nil
	config.HTTPClient = server.Client()
	r := &serviceAccountApplicationKeyResource{
		Api:  datadogV2.NewServiceAccountsApi(datadog.NewAPIClient(config)),
		Auth: context.Background(),
	}
	var schemaResponse resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResponse)
	return r, schemaResponse
}

func TestServiceAccountApplicationKeyCreateKeepsStateWhenGetFails(t *testing.T) {
	ctx := context.Background()
	r, schemaResponse := newTestServiceAccountAppKeyResource(t, http.StatusForbidden)
	planned := serviceAccountApplicationKeyModel{
		ID:               types.StringUnknown(),
		ServiceAccountId: types.StringValue("sa-id"),
		Name:             types.StringValue("k"),
		Key:              types.StringUnknown(),
		CreatedAt:        types.StringUnknown(),
		Last4:            types.StringUnknown(),
		Scopes:           types.SetNull(types.StringType),
	}
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &planned); diags.HasError() {
		t.Fatalf("building plan: %v", diags.Errors())
	}
	response := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)

	if response.Diagnostics.HasError() {
		t.Fatalf("Create returned errors: %v", response.Diagnostics.Errors())
	}
	if len(response.Diagnostics.Warnings()) != 1 {
		t.Fatalf("got %d warnings, want 1", len(response.Diagnostics.Warnings()))
	}
	var got serviceAccountApplicationKeyModel
	if diags := response.State.Get(ctx, &got); diags.HasError() {
		t.Fatalf("reading state: %v", diags.Errors())
	}
	if got.ID.ValueString() != "key-id" || got.Key.ValueString() != "secret-value" || got.Last4.ValueString() != "alue" {
		t.Fatalf("POST result not saved: %+v", got)
	}
	if got.CreatedAt.IsUnknown() || got.CreatedAt.ValueString() != "2026-11-18T13:37:52Z" {
		t.Fatalf("created_at = %v, want create response time in RFC3339", got.CreatedAt)
	}
}

func TestServiceAccountApplicationKeyCreateUsesReadCreatedAt(t *testing.T) {
	ctx := context.Background()
	r, schemaResponse := newTestServiceAccountAppKeyResource(t, http.StatusOK)
	planned := serviceAccountApplicationKeyModel{
		ID: types.StringUnknown(), ServiceAccountId: types.StringValue("sa-id"), Name: types.StringValue("k"),
		Key: types.StringUnknown(), CreatedAt: types.StringUnknown(), Last4: types.StringUnknown(), Scopes: types.SetNull(types.StringType),
	}
	plan := tfsdk.Plan{Schema: schemaResponse.Schema}
	if diags := plan.Set(ctx, &planned); diags.HasError() {
		t.Fatalf("building plan: %v", diags.Errors())
	}
	response := resource.CreateResponse{State: tfsdk.State{Raw: plan.Raw, Schema: schemaResponse.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &response)
	if len(response.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", response.Diagnostics)
	}
	var got serviceAccountApplicationKeyModel
	response.State.Get(ctx, &got)
	if got.CreatedAt.ValueString() != testAppKeyRawCreatedAt || got.Key.ValueString() != "secret-value" {
		t.Fatalf("state = %+v", got)
	}
}

// State written by earlier provider versions has created_at in RFC3339 seconds.
// Read must keep the key, which GET does not return, and change only the computed created_at.
func TestServiceAccountApplicationKeyReadOldFormatState(t *testing.T) {
	ctx := context.Background()
	r, schemaResponse := newTestServiceAccountAppKeyResource(t, http.StatusOK)
	prior := serviceAccountApplicationKeyModel{
		ID: types.StringValue("key-id"), ServiceAccountId: types.StringValue("sa-id"), Name: types.StringValue("k"),
		Key: types.StringValue("secret-value"), CreatedAt: types.StringValue("2026-11-18T13:37:52Z"),
		Last4: types.StringValue("alue"), Scopes: types.SetNull(types.StringType),
	}
	state := tfsdk.State{Schema: schemaResponse.Schema}
	if diags := state.Set(ctx, &prior); diags.HasError() {
		t.Fatalf("building state: %v", diags.Errors())
	}
	response := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("Read errors: %v", response.Diagnostics.Errors())
	}
	var got serviceAccountApplicationKeyModel
	response.State.Get(ctx, &got)
	want := prior
	want.CreatedAt = types.StringValue(testAppKeyRawCreatedAt)
	if !got.Key.Equal(want.Key) || !got.Name.Equal(want.Name) || !got.Last4.Equal(want.Last4) ||
		!got.ID.Equal(want.ID) || !got.Scopes.Equal(want.Scopes) || !got.CreatedAt.Equal(want.CreatedAt) {
		t.Fatalf("state after Read = %+v, want %+v", got, want)
	}
}
