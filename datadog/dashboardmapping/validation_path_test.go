package dashboardmapping

import "testing"

func TestWidgetDefinitionHCLKey(t *testing.T) {
	if got := WidgetDefinitionHCLKey("timeseries"); got != "timeseries_definition" {
		t.Fatalf("expected timeseries_definition, got %q", got)
	}
}

func TestWidgetErrorPathToHCL(t *testing.T) {
	tests := []struct {
		name       string
		widgetType string
		jsonPath   string
		want       string
	}{
		{
			name:       "legacy request query",
			widgetType: "timeseries",
			jsonPath:   "requests.0.q",
			want:       "request.0.q",
		},
		{
			name:       "bracket indexes",
			widgetType: "timeseries",
			jsonPath:   "requests[0].q",
			want:       "request.0.q",
		},
		{
			name:       "backend list representation",
			widgetType: "timeseries",
			jsonPath:   "['requests', 0, 'q']",
			want:       "request.0.q",
		},
		{
			name:       "formula and function query",
			widgetType: "timeseries",
			jsonPath:   "requests.0.queries.1.query",
			want:       "request.0.query.1.query",
		},
		{
			name:       "unknown widget preserves backend path",
			widgetType: "future_widget",
			jsonPath:   "requests.0.q",
			want:       "requests.0.q",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := WidgetErrorPathToHCL(test.widgetType, test.jsonPath); got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}
