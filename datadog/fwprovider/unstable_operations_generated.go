package fwprovider

// generatedUnstableOperations lists the x-unstable operations the generated
// artifacts call. The pinned SDK defaults every one of them to disabled, so an
// artifact naming an operation absent from this slice compiles and registers
// and then fails every call at runtime. tfgen owns this file: each run merges
// the operations it produced into the existing set (union, sorted) so a scoped
// --include run never drops entries another artifact needs. Do not edit by hand.
//
// EnableGeneratedUnstableOperations enables each of these on the client config.
var generatedUnstableOperations = []string{
	"v2.CreateElasticCloudIntegrationAccount",
	"v2.CreateTwilioIntegrationAccount",
	"v2.DeleteElasticCloudIntegrationAccount",
	"v2.DeleteTwilioIntegrationAccount",
	"v2.GetElasticCloudIntegrationAccount",
	"v2.GetTwilioIntegrationAccount",
	"v2.UpdateElasticCloudIntegrationAccount",
	"v2.UpdateTwilioIntegrationAccount",
}
