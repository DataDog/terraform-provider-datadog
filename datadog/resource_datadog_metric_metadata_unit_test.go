package datadog

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestUpdateMetricMetadataStateNullFields(t *testing.T) {
	var metadata datadogV1.MetricMetadata
	body := `{"type":"gauge","description":null,"short_name":null,"integration":null,"statsd_interval":null,"unit":null,"per_unit":null}`
	if err := json.Unmarshal([]byte(body), &metadata); err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, resourceDatadogMetricMetadata().SchemaMap(), map[string]interface{}{"metric": "custom.metric"})
	d.SetId("custom.metric")

	if diags := updateMetricMetadataState(d, &metadata); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := d.Get("type"); got != "gauge" {
		t.Errorf("type = %v, want gauge", got)
	}
	if got := d.Get("statsd_interval"); got != 0 {
		t.Errorf("statsd_interval = %v, want 0", got)
	}
}
