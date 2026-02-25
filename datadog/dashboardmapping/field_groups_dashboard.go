package dashboardmapping

// field_groups_dashboard.go — dashboard top-level field groups.
// These are NOT shared with any widget spec (see field_groups.go for widget groups).
// Descriptions sourced from OpenAPI components/schemas/ in dashboard.yaml.

// dashboardTemplateVariableFields corresponds to OpenAPI DashboardTemplateVariable.
var dashboardTemplateVariableFields = []FieldSpec{
	{HCLKey: "name", Required: true, Type: TypeString, OmitEmpty: false,
		Description: "The name of the variable."},
	{HCLKey: "prefix", OmitEmpty: true, Type: TypeString,
		Description: "The tag prefix associated with the variable. Only tags with this prefix appear in the variable dropdown."},
	{HCLKey: "available_values", OmitEmpty: true, Type: TypeStringList,
		Description: "The list of values that the template variable drop-down is be limited to"},
	{HCLKey: "default", OmitEmpty: true, Type: TypeString,
		Deprecated:  "Use `defaults` instead.",
		Description: "The default value for the template variable on dashboard load. Cannot be used in conjunction with `defaults`."},
	{HCLKey: "defaults", OmitEmpty: true, Type: TypeStringList,
		Description: "One or many default values for template variables on load. If more than one default is specified, they will be unioned together with `OR`. Cannot be used in conjunction with `default`."},
	{HCLKey: "type", OmitEmpty: true, Type: TypeString,
		Description: "The type of variable. This is to differentiate between filter variables (interpolated in query) and group by variables (interpolated into group by)."},
}

// dashboardTemplateVariablePresetValueFields corresponds to OpenAPI DashboardTemplateVariablePresetValue.
var dashboardTemplateVariablePresetValueFields = []FieldSpec{
	{HCLKey: "name", OmitEmpty: true, Type: TypeString,
		Description: "The name of the template variable"},
	{HCLKey: "values", OmitEmpty: true, Type: TypeStringList,
		Description: "One or many template variable values within the saved view, which will be unioned together using `OR` if more than one is specified. Cannot be used in conjunction with `value`."},
	{HCLKey: "value", OmitEmpty: true, Type: TypeString,
		Deprecated:  "Use `values` instead.",
		Description: "The value that should be assumed by the template variable in this preset. Cannot be used in conjunction with `values`."},
}

// dashboardTemplateVariablePresetFields corresponds to OpenAPI DashboardTemplateVariablePreset.
// HCL: "template_variable" (singular) → JSON: "template_variables" (plural).
var dashboardTemplateVariablePresetFields = []FieldSpec{
	{HCLKey: "name", OmitEmpty: true, Type: TypeString,
		Description: "The name of the preset."},
	{HCLKey: "template_variable", JSONKey: "template_variables",
		OmitEmpty: false, Type: TypeBlockList,
		Description: "The template variable names and assumed values under the given preset",
		Children:    dashboardTemplateVariablePresetValueFields},
}
