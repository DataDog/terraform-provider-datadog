package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringDefaultRuleResource{}
	_ resource.ResourceWithImportState = &securityMonitoringDefaultRuleResource{}
)

type securityMonitoringDefaultRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

func NewSecurityMonitoringDefaultRuleResource() resource.Resource {
	return &securityMonitoringDefaultRuleResource{}
}

func (r *securityMonitoringDefaultRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringDefaultRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_default_rule"
}

func (r *securityMonitoringDefaultRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource for default rules. It can only be imported, you can't create a default rule.",
		Attributes:  map[string]schema.Attribute{},
	}
}

func (r *securityMonitoringDefaultRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError(
		"Default rule cannot be created",
		"cannot create a default rule, please import it first before making changes",
	)
}

func (r *securityMonitoringDefaultRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	response.Diagnostics.AddError("not implemented", "Read is not implemented yet for the framework default rule resource")
}

func (r *securityMonitoringDefaultRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("not implemented", "Update is not implemented yet for the framework default rule resource")
}

func (r *securityMonitoringDefaultRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// no-op
}

func (r *securityMonitoringDefaultRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
