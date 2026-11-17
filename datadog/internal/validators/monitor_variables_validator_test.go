package validators

import "testing"

func TestValidateMonitorVariablesForType(t *testing.T) {
	tests := []struct {
		monitorType string
		dataSources []string
		metricQuery bool
		wantErr     bool
	}{
		{"query alert", []string{"metrics"}, false, true}, // (A)
		{"cost alert", nil, true, true},                   // (B)
		{"query alert", nil, true, false},                 // metric_query on a metric monitor
		{"cost alert", []string{"cloud_cost", "datadog_usage", "metrics"}, false, false},
	}
	for _, tt := range tests {
		if err := ValidateMonitorVariablesForType(tt.monitorType, tt.dataSources, tt.metricQuery); (err != nil) != tt.wantErr {
			t.Errorf("%s %v metric_query=%v: got err %v", tt.monitorType, tt.dataSources, tt.metricQuery, err)
		}
	}
}
