package fwprovider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

type planTestProvider struct {
	r *serviceAccountApplicationKeyResource
}

func (p *planTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "datadog"
}
func (p *planTestProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}
func (p *planTestProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (p *planTestProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
func (p *planTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{func() resource.Resource { return &planTestResource{p.r} }}
}

// planTestResource keeps the real CRUD and plan logic and replaces Configure, which needs a FrameworkProvider.
type planTestResource struct {
	*serviceAccountApplicationKeyResource
}

func (r *planTestResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "datadog_service_account_application_key"
}
func (r *planTestResource) Configure(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse) {
}

// Runs refresh then PlanResourceChange through the protocol 6 server, as Terraform core does for
// "terraform plan" with unchanged configuration and state written by earlier provider versions.
func TestServiceAccountApplicationKeyPlanOldFormatStateIsNoOp(t *testing.T) {
	ctx := context.Background()
	r, _ := newTestServiceAccountAppKeyResource(t, http.StatusOK)
	server, err := providerserver.NewProtocol6WithError(&planTestProvider{r})()
	if err != nil {
		t.Fatal(err)
	}
	const typeName = "datadog_service_account_application_key"
	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	objType := schemaResp.ResourceSchemas[typeName].ValueType()
	if _, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{
		Config: mustDynamicValue(t, schemaResp.Provider.ValueType(), tftypes.NewValue(schemaResp.Provider.ValueType(), map[string]tftypes.Value{})),
	}); err != nil {
		t.Fatal(err)
	}

	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	nullStr := tftypes.NewValue(tftypes.String, nil)
	scopesType := tftypes.Set{ElementType: tftypes.String}
	build := func(created, key, last4, id tftypes.Value) tftypes.Value {
		return tftypes.NewValue(objType, map[string]tftypes.Value{
			"id": id, "service_account_id": str("sa-id"), "name": str("k"), "key": key,
			"created_at": created, "last4": last4, "scopes": tftypes.NewValue(scopesType, nil),
		})
	}
	oldState := build(str("2026-11-18T13:37:52Z"), str("secret-value"), str("alue"), str("key-id"))
	config := build(nullStr, nullStr, nullStr, nullStr)

	readResp, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: typeName, CurrentState: mustDynamicValue(t, objType, oldState)})
	if err != nil || len(readResp.Diagnostics) != 0 {
		t.Fatalf("ReadResource: %v %v", err, readResp.Diagnostics)
	}
	refreshed, err := readResp.NewState.Unmarshal(objType)
	if err != nil {
		t.Fatal(err)
	}

	// Terraform core builds the proposed state from config, with computed attributes taken from prior state.
	planResp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         typeName,
		PriorState:       readResp.NewState,
		ProposedNewState: readResp.NewState,
		Config:           mustDynamicValue(t, objType, config),
	})
	if err != nil || len(planResp.Diagnostics) != 0 {
		t.Fatalf("PlanResourceChange: %v %v", err, planResp.Diagnostics)
	}
	if len(planResp.RequiresReplace) != 0 {
		t.Fatalf("RequiresReplace = %v, want none", planResp.RequiresReplace)
	}
	planned, err := planResp.PlannedState.Unmarshal(objType)
	if err != nil {
		t.Fatal(err)
	}
	if !planned.Equal(refreshed) {
		t.Fatalf("planned state differs from refreshed state (resource action expected NoOp)\nplanned: %v\nprior:   %v", planned, refreshed)
	}
	var attrs map[string]tftypes.Value
	if err := planned.As(&attrs); err != nil {
		t.Fatal(err)
	}
	if !attrs["key"].Equal(str("secret-value")) {
		t.Fatalf("key = %v, want unchanged", attrs["key"])
	}
}

func mustDynamicValue(t *testing.T, typ tftypes.Type, v tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()
	dv, err := tfprotov6.NewDynamicValue(typ, v)
	if err != nil {
		t.Fatal(err)
	}
	return &dv
}
