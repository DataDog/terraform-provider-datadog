package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringRuleResource{}
	_ resource.ResourceWithImportState = &securityMonitoringRuleResource{}
)

type securityMonitoringRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

func NewSecurityMonitoringRuleResource() resource.Resource {
	return &securityMonitoringRuleResource{}
}

func (r *securityMonitoringRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_rule"
}

func (r *securityMonitoringRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule, use `datadog_security_monitoring_default_rule` instead.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *securityMonitoringRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Create is not yet implemented")
}

func (r *securityMonitoringRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Read is not yet implemented")
}

func (r *securityMonitoringRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Update is not yet implemented")
}

func (r *securityMonitoringRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Delete is not yet implemented")
}

func (r *securityMonitoringRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule ImportState is not yet implemented")
}
