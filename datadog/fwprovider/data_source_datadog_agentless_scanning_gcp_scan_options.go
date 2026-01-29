package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &agentlessScanningGcpScanOptionsDataSource{}
)

type GcpScanOptionsModel struct {
	GcpProjectId     types.String `tfsdk:"gcp_project_id"`
	VulnContainersOs types.Bool   `tfsdk:"vuln_containers_os"`
	VulnHostOs       types.Bool   `tfsdk:"vuln_host_os"`
}

type agentlessScanningGcpScanOptionsDataSourceModel struct {
	ID             types.String           `tfsdk:"id"`
	GcpScanOptions []*GcpScanOptionsModel `tfsdk:"gcp_scan_options"`
}

type agentlessScanningGcpScanOptionsDataSource struct {
	Api  *datadogV2.AgentlessScanningApi
	Auth context.Context
}

func NewAgentlessScanningGcpScanOptionsDataSource() datasource.DataSource {
	return &agentlessScanningGcpScanOptionsDataSource{}
}

func (d *agentlessScanningGcpScanOptionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAgentlessScanningApiV2()
	d.Auth = providerData.Auth
}

func (d *agentlessScanningGcpScanOptionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "agentless_scanning_gcp_scan_options"
}

func (d *agentlessScanningGcpScanOptionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve all GCP scan options for Datadog Agentless Scanning.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"gcp_scan_options": schema.ListAttribute{
				Computed:    true,
				Description: "List of GCP scan options.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"gcp_project_id":     types.StringType,
						"vuln_containers_os": types.BoolType,
						"vuln_host_os":       types.BoolType,
					},
				},
			},
		},
	}
}

func (d *agentlessScanningGcpScanOptionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state agentlessScanningGcpScanOptionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, _, err := d.Api.ListGcpScanOptions(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing GCP scan options"))
		return
	}

	d.updateState(&state, ddResp.GetData())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *agentlessScanningGcpScanOptionsDataSource) updateState(state *agentlessScanningGcpScanOptionsDataSourceModel, scanOptionsData []datadogV2.GcpScanOptionsData) {
	var gcpScanOptions []*GcpScanOptionsModel
	for _, scanOption := range scanOptionsData {
		attributes := scanOption.GetAttributes()
		opt := GcpScanOptionsModel{
			GcpProjectId:     types.StringValue(scanOption.GetId()),
			VulnContainersOs: types.BoolValue(attributes.GetVulnContainersOs()),
			VulnHostOs:       types.BoolValue(attributes.GetVulnHostOs()),
		}
		gcpScanOptions = append(gcpScanOptions, &opt)
	}

	state.ID = types.StringValue(utils.ConvertToSha256("agentless_scanning_gcp_scan_options"))
	state.GcpScanOptions = gcpScanOptions
}
