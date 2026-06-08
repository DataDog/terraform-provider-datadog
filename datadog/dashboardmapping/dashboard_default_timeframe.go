package dashboardmapping

// dashboardDefaultTimeframeLiveFields corresponds to the "live" variant of DashboardDefaultTimeframe.
var dashboardDefaultTimeframeLiveFields = []FieldSpec{
	{HCLKey: "unit", Type: TypeString, OmitEmpty: false, Required: true,
		ValidValues: []string{"minute", "hour", "day", "week", "month", "year"},
		Description: "Unit of the live timeframe span."},
	{HCLKey: "value", Type: TypeInt, OmitEmpty: false, Required: true,
		Description: "Value of the live timeframe span."},
}

// dashboardDefaultTimeframeFixedFields corresponds to the "fixed" variant of DashboardDefaultTimeframe.
var dashboardDefaultTimeframeFixedFields = []FieldSpec{
	{HCLKey: "from", Type: TypeInt, OmitEmpty: false, Required: true,
		Description: "Start time in milliseconds since epoch."},
	{HCLKey: "to", Type: TypeInt, OmitEmpty: false, Required: true,
		Description: "End time in milliseconds since epoch."},
}

// DashboardDefaultTimeframeField returns the TypeOneOf FieldSpec for default_timeframe.
// Used by both the v2 datadog_dashboard_v2 resource (via DashboardTopLevelFields) and the
// v1 datadog_dashboard resource, so both expose the same nested live/fixed block shape.
func DashboardDefaultTimeframeField() FieldSpec {
	return FieldSpec{
		HCLKey:        "default_timeframe",
		Type:          TypeOneOf,
		OmitEmpty:     true,
		NullOnClear:   true,
		Description:   "The default timeframe applied when opening the dashboard. Set to `null` to disable after it has been configured.",
		Discriminator: &OneOfDiscriminator{JSONKey: "type"},
		Children: []FieldSpec{
			{HCLKey: "live", Type: TypeBlock, OmitEmpty: true,
				Discriminator: &OneOfDiscriminator{Value: "live"},
				Description:   "A live timeframe applied when opening the dashboard.",
				Children:      dashboardDefaultTimeframeLiveFields},
			{HCLKey: "fixed", Type: TypeBlock, OmitEmpty: true,
				Discriminator: &OneOfDiscriminator{Value: "fixed"},
				Description:   "A fixed timeframe applied when opening the dashboard.",
				Children:      dashboardDefaultTimeframeFixedFields},
		},
	}
}
