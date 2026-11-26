package fwprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
)

func TestRetryOnCSMThreatsLocked(t *testing.T) {
	errAPI := errors.New("api error")

	tests := []struct {
		name      string
		responses []int
		wantCalls int
		wantErr   bool
	}{
		{name: "success first try", responses: []int{http.StatusOK}, wantCalls: 1},
		{name: "locked then success", responses: []int{http.StatusLocked, http.StatusLocked, http.StatusOK}, wantCalls: 3},
		{name: "non-locked error is not retried", responses: []int{http.StatusBadRequest}, wantCalls: 1, wantErr: true},
		{name: "nil response is not retried", responses: []int{0}, wantCalls: 1, wantErr: true},
		{name: "locked until timeout", responses: []int{http.StatusLocked}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			err := retryOnCSMThreatsLocked(context.Background(), 2*time.Second, func() (*http.Response, error) {
				status := tt.responses[min(calls, len(tt.responses)-1)]
				calls++
				switch status {
				case http.StatusOK:
					return &http.Response{StatusCode: status}, nil
				case 0:
					return nil, errAPI
				default:
					return &http.Response{StatusCode: status}, errAPI
				}
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantCalls > 0 && calls != tt.wantCalls {
				t.Fatalf("got %d calls, want %d", calls, tt.wantCalls)
			}
			if tt.wantCalls == 0 && calls < 2 {
				t.Fatalf("got %d calls, want retries until timeout", calls)
			}
		})
	}
}

func TestCSMThreatsAgentRuleValidateActions(t *testing.T) {
	tests := []struct {
		name    string
		action  ActionModel
		wantErr string
	}{
		{name: "empty action", action: ActionModel{}, wantErr: "set, hash, kill or network_filter"},
		{name: "kill without signal", action: ActionModel{Kill: &KillActionModel{Signal: types.StringNull(), Scope: types.StringValue("process")}}, wantErr: "'signal' is required"},
		{name: "kill with empty signal", action: ActionModel{Kill: &KillActionModel{Signal: types.StringValue("")}}, wantErr: "'signal' is required"},
		{name: "kill with signal", action: ActionModel{Kill: &KillActionModel{Signal: types.StringValue("SIGKILL")}}},
		{name: "network_filter without filter", action: ActionModel{NetworkFilter: &NetworkFilterActionModel{Filter: types.StringNull(), Policy: types.StringValue("drop")}}, wantErr: "'filter' is required"},
		{name: "network_filter with filter", action: ActionModel{NetworkFilter: &NetworkFilterActionModel{Filter: types.StringValue("tcp port 4444")}}},
	}

	r := &csmThreatsAgentRuleResource{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := r.validateActions(context.Background(), []ActionModel{tt.action})
			if tt.wantErr == "" {
				if diags.HasError() {
					t.Fatalf("unexpected error: %v", diags)
				}
				return
			}
			if !diags.HasError() {
				t.Fatalf("expected error containing %q, got none", tt.wantErr)
			}
			if detail := diags.Errors()[0].Detail(); !strings.Contains(detail, tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, detail)
			}
		})
	}
}

func TestCSMThreatsAgentRuleKillActionRoundTrip(t *testing.T) {
	in := &KillActionModel{
		Signal:                    types.StringValue("SIGTERM"),
		Scope:                     types.StringValue("container"),
		DisableContainerDisarmer:  types.BoolValue(true),
		DisableExecutableDisarmer: types.BoolUnknown(),
	}

	built := buildKillAction(in)
	if built.DisableExecutableDisarmer != nil {
		t.Fatalf("unknown value must not be sent, got %v", *built.DisableExecutableDisarmer)
	}

	got := flattenKillAction(built)
	want := KillActionModel{
		Signal:                    types.StringValue("SIGTERM"),
		Scope:                     types.StringValue("container"),
		DisableContainerDisarmer:  types.BoolValue(true),
		DisableExecutableDisarmer: types.BoolValue(false),
	}
	if *got != want {
		t.Fatalf("got %+v, want %+v", *got, want)
	}

	empty := flattenKillAction(buildKillAction(&KillActionModel{Signal: types.StringValue("SIGKILL"), Scope: types.StringNull()}))
	if !empty.Scope.IsNull() {
		t.Fatalf("omitted scope must flatten to null, got %v", empty.Scope)
	}
}

func TestCSMThreatsAgentRuleNetworkFilterActionRoundTrip(t *testing.T) {
	in := &NetworkFilterActionModel{
		Filter: types.StringValue("tcp port 4444"),
		Policy: types.StringValue("allow"),
		Scope:  types.StringNull(),
	}

	built := buildNetworkFilterAction(in)
	if built.Scope != nil {
		t.Fatalf("null scope must not be sent, got %q", *built.Scope)
	}

	got := flattenNetworkFilterAction(built)
	if *got != *in {
		t.Fatalf("got %+v, want %+v", *got, *in)
	}
}

const csmThreatsKillAndNetworkFilterActionsJSON = `[
	{"kill":{"signal":"SIGKILL","scope":"process","disable_container_disarmer":true}},
	{"network_filter":{"filter":"tcp port 4444","policy":"drop","scope":"cgroup"}}
]`

func csmThreatsAgentRuleTestModel() *csmThreatsAgentRuleModel {
	return &csmThreatsAgentRuleModel{
		Id:          types.StringValue("rule-1"),
		PolicyId:    types.StringValue("policy-1"),
		Name:        types.StringValue("my_rule"),
		Description: types.StringNull(),
		Enabled:     types.BoolValue(true),
		Expression:  customtypes.TrimSpaceStringValue{StringValue: types.StringValue("exec.file.name == \"nc\"")},
		ProductTags: types.SetNull(types.StringType),
		Silent:      types.BoolValue(false),
		Actions: []ActionModel{
			{Kill: &KillActionModel{
				Signal:                    types.StringValue("SIGKILL"),
				Scope:                     types.StringValue("process"),
				DisableContainerDisarmer:  types.BoolValue(true),
				DisableExecutableDisarmer: types.BoolUnknown(),
			}},
			{NetworkFilter: &NetworkFilterActionModel{
				Filter: types.StringValue("tcp port 4444"),
				Policy: types.StringValue("drop"),
				Scope:  types.StringValue("cgroup"),
			}},
		},
	}
}

func TestCSMThreatsAgentRuleCreatePayloadSendsKillAndNetworkFilter(t *testing.T) {
	r := &csmThreatsAgentRuleResource{}
	payload, err := r.buildCreateCSMThreatsAgentRulePayload(csmThreatsAgentRuleTestModel())
	require.NoError(t, err)

	actions, err := json.Marshal(payload.Data.Attributes.Actions)
	require.NoError(t, err)
	require.JSONEq(t, csmThreatsKillAndNetworkFilterActionsJSON, string(actions))
}

func TestCSMThreatsAgentRuleUpdatePayloadSendsKillAndNetworkFilter(t *testing.T) {
	r := &csmThreatsAgentRuleResource{}
	payload, err := r.buildUpdateCSMThreatsAgentRulePayload(csmThreatsAgentRuleTestModel())
	require.NoError(t, err)

	actions, err := json.Marshal(payload.Data.Attributes.Actions)
	require.NoError(t, err)
	require.JSONEq(t, csmThreatsKillAndNetworkFilterActionsJSON, string(actions))
}

func TestCSMThreatsAgentRulesDataSourceReadsKillAndNetworkFilter(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/api/v2/remote_config/products/cws/agent_rules", request.URL.Path)
		assert.Equal(t, "policy-1", request.URL.Query().Get("policy_id"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":[{"id":"rule-1","type":"agent_rule","attributes":{"name":"my_rule","description":"","enabled":true,"expression":"exec.file.name == \"nc\"","product_tags":[],"actions":%s}}]}`, csmThreatsKillAndNetworkFilterActionsJSON)
	}))
	defer server.Close()

	clientConfig := datadog.NewConfiguration()
	clientConfig.Servers = datadog.ServerConfigurations{{URL: server.URL}}
	clientConfig.OperationServers = nil
	clientConfig.HTTPClient = server.Client()
	d := &csmThreatsAgentRulesDataSource{api: datadogV2.NewCSMThreatsApi(datadog.NewAPIClient(clientConfig)), auth: ctx}

	var schemaResponse datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	configPlan := tfsdk.Plan{Schema: schemaResponse.Schema}
	require.False(t, configPlan.Set(ctx, &csmThreatsAgentRulesDataSourceModel{
		PolicyId:      types.StringValue("policy-1"),
		Id:            types.StringNull(),
		AgentRulesIds: types.ListNull(types.StringType),
	}).HasError())

	response := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Raw: configPlan.Raw, Schema: schemaResponse.Schema}}, &response)
	require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)

	var state csmThreatsAgentRulesDataSourceModel
	require.False(t, response.State.Get(ctx, &state).HasError())
	require.Len(t, state.AgentRules, 1)
	actions := state.AgentRules[0].Actions
	require.Len(t, actions, 2)
	require.Equal(t, KillActionModel{
		Signal:                    types.StringValue("SIGKILL"),
		Scope:                     types.StringValue("process"),
		DisableContainerDisarmer:  types.BoolValue(true),
		DisableExecutableDisarmer: types.BoolValue(false),
	}, *actions[0].Kill)
	require.Equal(t, NetworkFilterActionModel{
		Filter: types.StringValue("tcp port 4444"),
		Policy: types.StringValue("drop"),
		Scope:  types.StringValue("cgroup"),
	}, *actions[1].NetworkFilter)
}
