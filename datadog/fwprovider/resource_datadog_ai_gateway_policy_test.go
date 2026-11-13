package fwprovider

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func policyTarget(provider, model string) aiGatewayPolicyTarget {
	target := aiGatewayPolicyTarget{Provider: types.StringNull(), Model: types.StringValue(model)}
	if provider != "" {
		target.Provider = types.StringValue(provider)
	}
	return target
}

func TestBuildAiGatewayPolicyRequestAttributes(t *testing.T) {
	state := aiGatewayPolicyModel{
		Inbound:  types.StringValue("coding-agent"),
		UserID:   types.StringNull(),
		Priority: types.Int64Value(3),
		Tiers: []aiGatewayPolicyTier{
			{Targets: []aiGatewayPolicyTarget{policyTarget("prod-us", "gpt-5.2-codex")}},
			{Targets: []aiGatewayPolicyTarget{policyTarget("", "claude-sonnet-5*"), policyTarget("*", "gpt-5*")}},
		},
	}

	body, err := json.Marshal(buildAiGatewayPolicyRequestAttributes(&state))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"inbound":"coding-agent","priority":3,"tiers":[` +
		`{"priority":1,"targets":[{"model":"gpt-5.2-codex","provider":"prod-us"}]},` +
		`{"priority":2,"targets":[{"model":"claude-sonnet-5*"},{"model":"gpt-5*","provider":"*"}]}]}`
	if string(body) != want {
		t.Errorf("request body\n got: %s\nwant: %s", body, want)
	}
}

// policyTier describes one tier of a mocked response: its targets as {provider, model} pairs and the saved row ID of
// its first target. Targets in the tier get consecutive IDs from firstID.
type policyTier struct {
	firstID int64
	targets [][2]string
}

func policyResponse(inbound string, tiers ...policyTier) datadogV2.AiGatewayPolicyResource {
	apiTiers := make([]datadogV2.AiGatewayPolicyTier, 0, len(tiers))
	for _, tier := range tiers {
		targets := make([]datadogV2.AiGatewayPolicyTarget, 0, len(tier.targets))
		for i, pm := range tier.targets {
			target := datadogV2.NewAiGatewayPolicyTarget(tier.firstID+int64(i), 3)
			if pm[0] != "" {
				target.Provider = *datadog.NewNullableString(datadog.PtrString(pm[0]))
			}
			target.SetModel(pm[1])
			targets = append(targets, *target)
		}
		apiTiers = append(apiTiers, *datadogV2.NewAiGatewayPolicyTier(targets))
	}
	return *datadogV2.NewAiGatewayPolicyResource(*datadogV2.NewAiGatewayPolicyAttributes(3, apiTiers), inbound, datadogV2.AIGATEWAYPOLICYTYPE_ROUTING_DOCUMENTS)
}

func TestUpdateAiGatewayPolicyState(t *testing.T) {
	codex := policyTier{targets: [][2]string{{"prod-us", "gpt-5.2-codex"}}}
	fallback := policyTier{targets: [][2]string{{"*", "gpt-5*"}, {"", "claude-sonnet-5*"}}}
	at := func(tier policyTier, firstID int64) policyTier { tier.firstID = firstID; return tier }

	t.Run("tiers are ordered by saved target ID, not response order", func(t *testing.T) {
		// Written as [codex, fallback]: codex got ID 10, fallback 11 and 12. The response lists them reversed.
		for _, state := range []aiGatewayPolicyModel{
			{Inbound: types.StringValue("coding-agent"), UserID: types.StringNull()}, // import: no prior tiers
			{Inbound: types.StringValue("coding-agent"), UserID: types.StringNull(), Tiers: []aiGatewayPolicyTier{
				{Targets: []aiGatewayPolicyTarget{policyTarget("prod-us", "gpt-5.2-codex")}},
			}},
		} {
			updateAiGatewayPolicyState(&state, policyResponse("coding-agent", at(fallback, 11), at(codex, 10)))
			if len(state.Tiers) != 2 || state.Tiers[0].Targets[0].Model.ValueString() != "gpt-5.2-codex" {
				t.Fatalf("tiers = %+v, want codex first", state.Tiers)
			}
			if !state.Tiers[1].Targets[1].Provider.IsNull() {
				t.Errorf("a target without a provider must read back as null")
			}
			if state.ID.ValueString() != "coding-agent" || state.Priority.ValueInt64() != 3 {
				t.Errorf("id = %q, priority = %d", state.ID.ValueString(), state.Priority.ValueInt64())
			}
		}
	})

	t.Run("a reorder made outside Terraform shows in state", func(t *testing.T) {
		state := aiGatewayPolicyModel{Inbound: types.StringValue("coding-agent"), UserID: types.StringNull(), Tiers: []aiGatewayPolicyTier{
			{Targets: []aiGatewayPolicyTarget{policyTarget("prod-us", "gpt-5.2-codex")}},
			{Targets: []aiGatewayPolicyTarget{policyTarget("*", "gpt-5*"), policyTarget("", "claude-sonnet-5*")}},
		}}
		// Rewritten elsewhere as [fallback, codex].
		updateAiGatewayPolicyState(&state, policyResponse("coding-agent", at(codex, 22), at(fallback, 20)))
		if got := state.Tiers[0].Targets[0].Model.ValueString(); got != "gpt-5*" {
			t.Errorf("first tier model = %q, want the fallback tier the server now holds first", got)
		}
	})

	t.Run("a user-scoped policy gets a user-scoped ID", func(t *testing.T) {
		const userID = "55555555-5555-5555-5555-555555555555"
		state := aiGatewayPolicyModel{Inbound: types.StringValue("coding-agent"), UserID: types.StringValue(userID)}
		updateAiGatewayPolicyState(&state, policyResponse("coding-agent", at(codex, 1)))
		if got := state.ID.ValueString(); got != userID+":coding-agent" {
			t.Errorf("id = %q", got)
		}
	})
}

func TestParseAiGatewayPolicyID(t *testing.T) {
	const userID = "55555555-5555-5555-5555-555555555555"
	tests := []struct {
		id, wantUser, wantInbound string
	}{
		{id: "coding-agent", wantInbound: "coding-agent"},
		{id: "gpt-5*", wantInbound: "gpt-5*"},
		{id: userID + ":coding-agent", wantUser: userID, wantInbound: "coding-agent"},
		{id: "team:coding", wantInbound: "team:coding"},
	}
	for _, tt := range tests {
		user, inbound := parseAiGatewayPolicyID(tt.id)
		gotUser := ""
		if user != nil {
			gotUser = user.String()
		}
		if gotUser != tt.wantUser || inbound != tt.wantInbound {
			t.Errorf("parse(%q) = (%q, %q), want (%q, %q)", tt.id, gotUser, inbound, tt.wantUser, tt.wantInbound)
		}
	}
}
