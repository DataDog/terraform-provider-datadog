package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGcpUcConfigKeepsConfiguredBillingAccountID(t *testing.T) {
	attrs := datadogV2.NewGCPUsageCostConfigAttributesWithDefaults()
	attrs.SetAccountId("03B59F_8AA142_1DBB1A")
	data := datadogV2.NewGCPUsageCostConfigWithDefaults()
	data.SetId("1")
	data.SetAttributes(*attrs)

	cases := []struct {
		name    string
		current types.String
		want    string
	}{
		{"hyphenated config is preserved", types.StringValue("03B59F-8AA142-1DBB1A"), "03B59F-8AA142-1DBB1A"},
		{"underscore config is unchanged", types.StringValue("03B59F_8AA142_1DBB1A"), "03B59F_8AA142_1DBB1A"},
		{"import takes the API value", types.StringNull(), "03B59F_8AA142_1DBB1A"},
		{"different account takes the API value", types.StringValue("AAAAAA-BBBBBB-CCCCCC"), "03B59F_8AA142_1DBB1A"},
	}
	r := &gcpUcConfigResource{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &gcpUcConfigModel{BillingAccountId: tc.current}
			r.updateStateFromResponseData(context.Background(), state, data)
			if got := state.BillingAccountId.ValueString(); got != tc.want {
				t.Fatalf("billing_account_id = %q, want %q", got, tc.want)
			}
		})
	}
}
