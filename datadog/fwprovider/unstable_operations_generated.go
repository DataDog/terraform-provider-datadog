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
	"v2.CreateAnthropicUsageAndCostsIntegrationAccount",
	"v2.CreateChatGPTEnterpriseIntegrationAccount",
	"v2.CreateClaudeComplianceIntegrationAccount",
	"v2.CreateCoreWeaveCloudIntegrationAccount",
	"v2.CreateDatabricksIntegrationAccount",
	"v2.CreateElasticCloudIntegrationAccount",
	"v2.CreateGitHubCostsIntegrationAccount",
	"v2.CreateNebiusCloudIntegrationAccount",
	"v2.CreateOpenAICodexIntegrationAccount",
	"v2.CreateOpenAIModelInferenceIntegrationAccount",
	"v2.CreateSnowflakeIntegrationAccount",
	"v2.CreateTwilioIntegrationAccount",
	"v2.DeleteAnthropicUsageAndCostsIntegrationAccount",
	"v2.DeleteChatGPTEnterpriseIntegrationAccount",
	"v2.DeleteClaudeComplianceIntegrationAccount",
	"v2.DeleteCoreWeaveCloudIntegrationAccount",
	"v2.DeleteDatabricksIntegrationAccount",
	"v2.DeleteElasticCloudIntegrationAccount",
	"v2.DeleteGitHubCostsIntegrationAccount",
	"v2.DeleteNebiusCloudIntegrationAccount",
	"v2.DeleteOpenAICodexIntegrationAccount",
	"v2.DeleteOpenAIModelInferenceIntegrationAccount",
	"v2.DeleteSnowflakeIntegrationAccount",
	"v2.DeleteTwilioIntegrationAccount",
	"v2.GetAnthropicUsageAndCostsIntegrationAccount",
	"v2.GetChatGPTEnterpriseIntegrationAccount",
	"v2.GetClaudeComplianceIntegrationAccount",
	"v2.GetCoreWeaveCloudIntegrationAccount",
	"v2.GetDatabricksIntegrationAccount",
	"v2.GetElasticCloudIntegrationAccount",
	"v2.GetGitHubCostsIntegrationAccount",
	"v2.GetNebiusCloudIntegrationAccount",
	"v2.GetOpenAICodexIntegrationAccount",
	"v2.GetOpenAIModelInferenceIntegrationAccount",
	"v2.GetSnowflakeIntegrationAccount",
	"v2.GetTwilioIntegrationAccount",
	"v2.UpdateAnthropicUsageAndCostsIntegrationAccount",
	"v2.UpdateChatGPTEnterpriseIntegrationAccount",
	"v2.UpdateClaudeComplianceIntegrationAccount",
	"v2.UpdateCoreWeaveCloudIntegrationAccount",
	"v2.UpdateDatabricksIntegrationAccount",
	"v2.UpdateElasticCloudIntegrationAccount",
	"v2.UpdateGitHubCostsIntegrationAccount",
	"v2.UpdateNebiusCloudIntegrationAccount",
	"v2.UpdateOpenAICodexIntegrationAccount",
	"v2.UpdateOpenAIModelInferenceIntegrationAccount",
	"v2.UpdateSnowflakeIntegrationAccount",
	"v2.UpdateTwilioIntegrationAccount",
}
