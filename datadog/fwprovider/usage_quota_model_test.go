package fwprovider

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestUsageQuotaLimitValidation(t *testing.T) {
	require.False(t, validUsageQuotaLimit(math.NaN()))
	for _, tc := range []struct {
		name  string
		value types.Float64
		valid bool
	}{
		{"zero", types.Float64Value(0), true}, {"whole", types.Float64Value(100), true},
		{"negative", types.Float64Value(-1), false}, {"fraction", types.Float64Value(99.5), false},
		{"overflow", types.Float64Value(math.Exp2(63)), false}, {"infinity", types.Float64Value(math.Inf(1)), false},
		{"unknown", types.Float64Unknown(), true}, {"null", types.Float64Null(), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var resp validator.Float64Response
			usageQuotaLimitValidator{}.ValidateFloat64(context.Background(), validator.Float64Request{Path: path.Root("usage_limit"), ConfigValue: tc.value}, &resp)
			require.Equal(t, tc.valid, !resp.Diagnostics.HasError())
		})
	}
}

func TestUsageQuotaImportID(t *testing.T) {
	for _, id := range []string{"", "ai_credits", "/opaque", "ai_credits/", "ai_credits/id/extra", " ai_credits/id", "ai_credits/id\n"} {
		_, _, err := parseUsageQuotaImportID(id)
		require.Error(t, err, id)
	}
	namespace, id, err := parseUsageQuotaImportID("future_namespace/opaque-ID_123")
	require.NoError(t, err)
	require.Equal(t, "future_namespace", namespace)
	require.Equal(t, "opaque-ID_123", id)
}

func quotaTestPlan() usageQuotaModel {
	return usageQuotaModel{ID: types.StringValue("ai_credits/opaque"), QuotaID: types.StringValue("opaque"), QuotaNamespace: types.StringValue("ai_credits"), Scope: types.MapNull(types.StringType), UsageLimit: types.Float64Value(100), Enforced: types.BoolValue(true), OrgPublicID: types.StringValue("org")}
}

func TestUsageQuotaStatePreservesLegacyFraction(t *testing.T) {
	ctx := context.Background()
	model := quotaTestPlan()
	model.UsageLimit = types.Float64Value(99.5)
	state := tfsdk.State{Schema: usageQuotaSchema()}
	require.False(t, state.Set(ctx, &model).HasError())
	var result usageQuotaModel
	require.False(t, state.Get(ctx, &result).HasError())
	require.Equal(t, 99.5, result.UsageLimit.ValueFloat64())
	require.True(t, result.Scope.IsNull())
}

func TestUsageQuotaScopeValidation(t *testing.T) {
	validators := usageQuotaSchema().Attributes["scope"].(schema.MapAttribute).Validators
	for _, tc := range []struct {
		name  string
		value types.Map
		valid bool
	}{
		{"org", types.MapNull(types.StringType), true},
		{"empty", types.MapValueMust(types.StringType, map[string]attr.Value{}), false},
		{"unknown_scope", types.MapUnknown(types.StringType), true},
		{"unknown_role", types.MapValueMust(types.StringType, map[string]attr.Value{"role_uuid": types.StringUnknown()}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var resp validator.MapResponse
			for _, v := range validators {
				v.ValidateMap(context.Background(), validator.MapRequest{Path: path.Root("scope"), ConfigValue: tc.value}, &resp)
			}
			require.Equal(t, tc.valid, !resp.Diagnostics.HasError())
		})
	}
}

func TestUsageQuotaCreatePayload(t *testing.T) {
	for _, scoped := range []bool{false, true} {
		model := quotaTestPlan()
		model.UsageLimit = types.Float64Value(0)
		model.Enforced = types.BoolValue(false)
		if scoped {
			model.Scope = types.MapValueMust(types.StringType, map[string]attr.Value{"user_handle": types.StringValue("Exact.Case@example.com")})
		}
		body, diags := buildUsageQuotaCreate(context.Background(), model)
		require.False(t, diags.HasError())
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		expected := `{"data":[{"type":"quotas","attributes":{"usage_limit":0,"enforced":false}}]}`
		if scoped {
			expected = `{"data":[{"type":"quotas","attributes":{"usage_limit":0,"enforced":false,"scope":{"user_handle":"Exact.Case@example.com"}}}]}`
		}
		require.JSONEq(t, expected, string(raw))
	}
	model := quotaTestPlan()
	model.Scope = types.MapValueMust(types.StringType, map[string]attr.Value{"role_uuid": types.StringUnknown()})
	_, diags := buildUsageQuotaCreate(context.Background(), model)
	require.True(t, diags.HasError())
	model = quotaTestPlan()
	model.UsageLimit = types.Float64Value(99.5)
	_, diags = buildUsageQuotaCreate(context.Background(), model)
	require.True(t, diags.HasError())
}

func TestUsageQuotaUpdatePayload(t *testing.T) {
	for _, tc := range []struct {
		name     string
		limit    float64
		enforced bool
		changed  bool
		attrs    string
	}{
		{"no change", 100, true, false, `{}`},
		{"disable only", 100, false, true, `{"enforced":false}`},
		{"limit only", 200, true, true, `{"usage_limit":200}`},
		{"both", 200, false, true, `{"usage_limit":200,"enforced":false}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := quotaTestPlan()
			plan := state
			plan.UsageLimit = types.Float64Value(tc.limit)
			plan.Enforced = types.BoolValue(tc.enforced)
			body, changed, diags := buildUsageQuotaUpdate(plan, state)
			require.False(t, diags.HasError())
			require.Equal(t, tc.changed, changed)
			raw, err := json.Marshal(body)
			require.NoError(t, err)
			require.JSONEq(t, `{"data":{"id":"opaque","type":"quotas","attributes":`+tc.attrs+`}}`, string(raw))
		})
	}
	state := quotaTestPlan()
	state.UsageLimit = types.Float64Value(99.5)
	plan := state
	plan.Enforced = types.BoolValue(false)
	body, _, diags := buildUsageQuotaUpdate(plan, state)
	require.False(t, diags.HasError())
	require.False(t, body.Data.Attributes.HasUsageLimit())
	plan.UsageLimit = types.Float64Value(100)
	body, _, diags = buildUsageQuotaUpdate(plan, state)
	require.False(t, diags.HasError())
	require.Equal(t, int64(100), body.Data.Attributes.GetUsageLimit())
}
