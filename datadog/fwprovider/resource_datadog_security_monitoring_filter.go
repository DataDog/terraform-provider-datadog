package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

const securityFilterType = "security_filters"

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringFilterResource{}
	_ resource.ResourceWithImportState = &securityMonitoringFilterResource{}
)

type securityMonitoringFilterResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

func NewSecurityMonitoringFilterResource() resource.Resource {
	return &securityMonitoringFilterResource{}
}

func (r *securityMonitoringFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_filter"
}

func (r *securityMonitoringFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{}
}

func (r *securityMonitoringFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *securityMonitoringFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_filter Create is not yet implemented")
}

func (r *securityMonitoringFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_filter Read is not yet implemented")
}

func (r *securityMonitoringFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_filter Update is not yet implemented")
}

func (r *securityMonitoringFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_filter Delete is not yet implemented")
}
