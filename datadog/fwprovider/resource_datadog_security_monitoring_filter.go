package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
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

type securityMonitoringFilterResourceModel struct {
	ID               types.String           `tfsdk:"id"`
	Name             types.String           `tfsdk:"name"`
	Version          types.Int64            `tfsdk:"version"`
	Query            types.String           `tfsdk:"query"`
	IsEnabled        types.Bool             `tfsdk:"is_enabled"`
	FilteredDataType types.String           `tfsdk:"filtered_data_type"`
	ExclusionFilter  []exclusionFilterModel `tfsdk:"exclusion_filter"`
}

type exclusionFilterModel struct {
	Name  types.String `tfsdk:"name"`
	Query types.String `tfsdk:"query"`
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
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource for security filters.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the security filter.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The version of the security filter.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"query": schema.StringAttribute{
				Required:    true,
				Description: "The query of the security filter.",
			},
			"is_enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the security filter is enabled.",
			},
			"filtered_data_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("logs"),
				Description: "The filtered data type.",
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewSecurityFilterFilteredDataTypeFromValue),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"exclusion_filter": schema.ListNestedBlock{
				Description: "Exclusion filters to exclude some logs from the security filter.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Exclusion filter name.",
						},
						"query": schema.StringAttribute{
							Required:    true,
							Description: "Exclusion filter query. Logs that match this query are excluded from the security filter.",
						},
					},
				},
			},
		},
	}
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
