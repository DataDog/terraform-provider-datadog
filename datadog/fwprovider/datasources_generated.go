package fwprovider

import "github.com/hashicorp/terraform-plugin-framework/datasource"

// generatedDatasources holds the data sources produced by the generator-v2 emit
// pipeline. Keeping them in their own slice (and file) means regenerating does
// not churn framework_provider.go, and the generated set stays reviewable in one
// place. FrameworkProvider.DataSources registers this slice alongside the
// hand-written Datasources.
//
// It is intentionally empty until a generated data source is promoted to the
// live provider; add its constructor here when it is.
var generatedDatasources = []func() datasource.DataSource{}
