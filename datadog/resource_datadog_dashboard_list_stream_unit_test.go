package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/stretchr/testify/assert"
)

// TestBuildTerraformListStreamWidgetRequests verifies that the list_stream read
// path flattens clustering_pattern_field_path and group_by from the API response
// back into Terraform state. Both fields have schema and write paths but were
// never mapped back on read, producing a perpetual diff.
// Regression test for https://github.com/DataDog/terraform-provider-datadog/issues/4149.
func TestBuildTerraformListStreamWidgetRequests(t *testing.T) {
	query := datadogV1.NewListStreamQuery(datadogV1.LISTSTREAMSOURCE_LOGS_PATTERN_STREAM, "service:foo")
	query.SetClusteringPatternFieldPath("message")
	query.SetGroupBy([]datadogV1.ListStreamGroupByItems{
		*datadogV1.NewListStreamGroupByItems("service"),
		*datadogV1.NewListStreamGroupByItems("status"),
	})

	column := datadogV1.NewListStreamColumn("details", datadogV1.LISTSTREAMCOLUMNWIDTH_AUTO)
	request := datadogV1.NewListStreamWidgetRequest(
		[]datadogV1.ListStreamColumn{*column},
		*query,
		datadogV1.LISTSTREAMRESPONSEFORMAT_EVENT_LIST,
	)

	result := buildTerraformListStreamWidgetRequests(&[]datadogV1.ListStreamWidgetRequest{*request})

	assert.Len(t, *result, 1)
	terraformQuery := (*result)[0]["query"].([]map[string]interface{})[0]

	// clustering_pattern_field_path must round-trip into state.
	clustering, ok := terraformQuery["clustering_pattern_field_path"].(*string)
	assert.True(t, ok, "clustering_pattern_field_path should be present in state")
	if assert.NotNil(t, clustering) {
		assert.Equal(t, "message", *clustering)
	}

	// group_by must round-trip into state as a list of {facet} maps.
	groupBy, ok := terraformQuery["group_by"].([]map[string]interface{})
	assert.True(t, ok, "group_by should be present in state")
	assert.Equal(t, []map[string]interface{}{
		{"facet": "service"},
		{"facet": "status"},
	}, groupBy)
}

// TestBuildTerraformListStreamWidgetRequestsOmitsUnsetFields ensures the new
// flatten blocks do not inject empty clustering_pattern_field_path/group_by keys
// for list stream sources that don't use them (e.g. event_stream), avoiding a
// spurious diff in the other direction.
func TestBuildTerraformListStreamWidgetRequestsOmitsUnsetFields(t *testing.T) {
	query := datadogV1.NewListStreamQuery(datadogV1.LISTSTREAMSOURCE_EVENT_STREAM, "example.metric")

	column := datadogV1.NewListStreamColumn("source", datadogV1.LISTSTREAMCOLUMNWIDTH_AUTO)
	request := datadogV1.NewListStreamWidgetRequest(
		[]datadogV1.ListStreamColumn{*column},
		*query,
		datadogV1.LISTSTREAMRESPONSEFORMAT_EVENT_LIST,
	)

	result := buildTerraformListStreamWidgetRequests(&[]datadogV1.ListStreamWidgetRequest{*request})

	assert.Len(t, *result, 1)
	terraformQuery := (*result)[0]["query"].([]map[string]interface{})[0]

	_, hasClustering := terraformQuery["clustering_pattern_field_path"]
	assert.False(t, hasClustering, "clustering_pattern_field_path should be absent when unset")
	_, hasGroupBy := terraformQuery["group_by"]
	assert.False(t, hasGroupBy, "group_by should be absent when unset")
}
