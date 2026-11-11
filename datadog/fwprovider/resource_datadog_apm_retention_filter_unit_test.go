package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestApmRetentionFilterRateState(t *testing.T) {
	text := types.StringValue
	for _, tc := range []struct {
		name                string
		span, trace         types.String
		apiSpan, apiTrace   float64
		hasTrace            bool
		wantSpan, wantTrace types.String
	}{
		{"independent precision", text("0.1"), text("0.01"), 0.1, 0.01, true, text("0.1"), text("0.01")},
		{"configured spelling", text("0.50"), text("1e-2"), 0.5, 0.01, true, text("0.50"), text("1e-2")},
		{"remote changes", text("0.1"), text("0.2"), 0.12, 0.23, true, text("0.12"), text("0.23")},
		{"import", types.StringNull(), types.StringNull(), 0.5, 0.01, true, text("0.5"), text("0.01")},
		{"omitted trace", text("1.0"), types.StringNull(), 1, 0, false, text("1.0"), types.StringNull()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := ApmRetentionFilterModel{Rate: tc.span, TraceRate: tc.trace}
			r := ApmRetentionFilterResource{}
			r.updateState(context.Background(), &state, "id", "test", tc.apiSpan, tc.hasTrace, tc.apiTrace, "*", true, "spans-sampling-processor")
			if !state.Rate.Equal(tc.wantSpan) || !state.TraceRate.Equal(tc.wantTrace) {
				t.Fatalf("got span %s, trace %s; want span %s, trace %s", state.Rate, state.TraceRate, tc.wantSpan, tc.wantTrace)
			}
		})
	}
}
