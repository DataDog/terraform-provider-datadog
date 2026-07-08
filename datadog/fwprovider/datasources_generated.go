package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/datasource"

// generatedDatasources holds the data sources produced by the generator-v2 emit
// pipeline. tfgen owns this file: every generate run rewrites it from the set of
// data sources it produced, keeping the generated registrations in one
// reviewable place without churning framework_provider.go. Do not edit by hand.
//
// FrameworkProvider.DataSources registers this slice alongside the hand-written
// Datasources.
var generatedDatasources = []func() datasource.DataSource{
	NewDatadogAuthnMappingDataSource,
}
