package fwprovider

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGcpUcConfigBillingAccountID(t *testing.T) {
	const remoteID = "AAAAAA_BBBBBB_CCCCCC"
	for _, tc := range []struct {
		name  string
		prior types.String
		want  string
	}{
		{"hyphenated", types.StringValue("AAAAAA-BBBBBB-CCCCCC"), "AAAAAA-BBBBBB-CCCCCC"},
		{"underscored", types.StringValue(remoteID), remoteID},
		{"changed account", types.StringValue("AAAAAA-BBBBBB-DDDDDD"), remoteID},
		{"import", types.StringNull(), remoteID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &gcpUcConfigResource{}
			state := gcpUcConfigModel{BillingAccountId: tc.prior}
			created := &datadogV2.GCPUsageCostConfig{}
			created.Attributes.SetAccountId(remoteID)
			r.updateStateFromResponseData(t.Context(), &state, created)
			if got := state.BillingAccountId.ValueString(); got != tc.want {
				t.Errorf("create returned %q, want %q", got, tc.want)
			}

			state = gcpUcConfigModel{BillingAccountId: tc.prior}
			attrs := &datadogV2.GcpUcConfigResponseDataAttributes{}
			attrs.SetAccountId(remoteID)
			read := &datadogV2.GcpUcConfigResponseData{Attributes: attrs}
			r.updateStateFromGcpUcConfigResponseData(t.Context(), &state, read)
			if got := state.BillingAccountId.ValueString(); got != tc.want {
				t.Errorf("read returned %q, want %q", got, tc.want)
			}
		})
	}
}
