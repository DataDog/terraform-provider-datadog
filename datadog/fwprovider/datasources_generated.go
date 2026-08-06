package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/datasource"

// generatedDatasources holds the data sources produced by the generator-v2 emit
// pipeline. tfgen owns this file: each run merges the constructors it produced
// into the existing set (union, sorted) so a scoped --include run never drops
// data sources it did not regenerate; reconcile prunes entries whose annotation
// is gone. Do not edit by hand.
//
// FrameworkProvider.DataSources registers this slice alongside the hand-written
// Datasources.
var generatedDatasources = []func() datasource.DataSource{
	NewDatadogApplicationKeysDataSource,
}
