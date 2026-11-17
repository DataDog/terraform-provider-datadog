package validators

import (
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

// ValidateMonitorVariablesForType checks that formula variables match the monitor type:
// metrics variables belong in `metric_query` on non-cost monitors, and cost monitors
// need `cloud_cost_query` (which requires an aggregator).
func ValidateMonitorVariablesForType(monitorType string, cloudCostDataSources []string, hasMetricQuery bool) error {
	if monitorType == string(datadogV1.MONITORTYPE_COST_ALERT) {
		if hasMetricQuery {
			return fmt.Errorf("`variables.metric_query` is not supported on %q monitors; use `variables.cloud_cost_query` (which requires an aggregator) instead", monitorType)
		}
		return nil
	}
	for _, ds := range cloudCostDataSources {
		if ds == "metrics" {
			return fmt.Errorf("`variables.cloud_cost_query` with `data_source = \"metrics\"` is only supported on \"cost alert\" monitors; use `variables.metric_query` for %q monitors instead and drop `aggregator`", monitorType)
		}
	}
	return nil
}
