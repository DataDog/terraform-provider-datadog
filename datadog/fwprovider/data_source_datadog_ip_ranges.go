package fwprovider

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &ipRangesDataSource{}

func NewIPRangesDataSource() datasource.DataSource {
	return &ipRangesDataSource{}
}

type ipRangesDataSourceZoneModel struct {
	ID types.String `tfsdk:"id"`
	// v4
	AgentsIpv4               types.List `tfsdk:"agents_ipv4"`
	APIIpv4                  types.List `tfsdk:"api_ipv4"`
	APMIpv4                  types.List `tfsdk:"apm_ipv4"`
	LogsIpv4                 types.List `tfsdk:"logs_ipv4"`
	OrchestratorIpv4         types.List `tfsdk:"orchestrator_ipv4"`
	ProcessIpv4              types.List `tfsdk:"process_ipv4"`
	SyntheticsIpv4           types.List `tfsdk:"synthetics_ipv4"`
	SyntheticsIpv4ByLocation types.Map  `tfsdk:"synthetics_ipv4_by_location"`
	WebhooksIpv4             types.List `tfsdk:"webhooks_ipv4"`
	// v6
	AgentsIpv6               types.List `tfsdk:"agents_ipv6"`
	APIIpv6                  types.List `tfsdk:"api_ipv6"`
	APMIpv6                  types.List `tfsdk:"apm_ipv6"`
	LogsIpv6                 types.List `tfsdk:"logs_ipv6"`
	OrchestratorIpv6         types.List `tfsdk:"orchestrator_ipv6"`
	ProcessIpv6              types.List `tfsdk:"process_ipv6"`
	SyntheticsIpv6           types.List `tfsdk:"synthetics_ipv6"`
	SyntheticsIpv6ByLocation types.Map  `tfsdk:"synthetics_ipv6_by_location"`
	WebhooksIpv6             types.List `tfsdk:"webhooks_ipv6"`
}

type ipRangesDataSource struct {
	Api  *datadogV1.IPRangesApi
	Auth context.Context
}

func (d *ipRangesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetIPRangesApiV1()
	d.Auth = providerData.Auth
}

func (d *ipRangesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "ip_ranges"
}

func (d *ipRangesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about Datadog's IP addresses.",
		Attributes: map[string]schema.Attribute{
			// v4
			"agents_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Agent endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"api_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the API endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"apm_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the APM endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"logs_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Logs endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"orchestrator_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Orchestrator endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"process_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Process endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"synthetics_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"synthetics_ipv4_by_location": schema.MapAttribute{
				Description: "A map of IPv4 prefixes (string of concatenated IPs, delimited by ',') by location.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"webhooks_ipv4": schema.ListAttribute{
				Description: "An Array of IPv4 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			// v6
			"agents_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Agent endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"api_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the API endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"apm_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the APM endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"logs_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Logs endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"orchestrator_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Orchestrator endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"process_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Process endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"synthetics_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"synthetics_ipv6_by_location": schema.MapAttribute{
				Description: "A map of IPv6 prefixes (string of concatenated IPs, delimited by ',') by location.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"webhooks_ipv6": schema.ListAttribute{
				Description: "An Array of IPv6 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				Computed:    true,
				ElementType: types.StringType,
			},
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (d *ipRangesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	var state ipRangesDataSourceZoneModel

	ipAddresses, _, err := d.Api.GetIPRanges(d.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting IPRanges"))
		return
	}

	state.ID = types.StringValue("datadog-ip-ranges")

	// v4 and v6
	ipAddressesPtr := &ipAddresses
	agents := ipAddressesPtr.GetAgents()
	api := ipAddressesPtr.GetApi()
	apm := ipAddressesPtr.GetApm()
	logs := ipAddressesPtr.GetLogs()
	orchestrator := ipAddressesPtr.GetOrchestrator()
	process := ipAddressesPtr.GetProcess()
	synthetics := ipAddressesPtr.GetSynthetics()
	webhook := ipAddressesPtr.GetWebhooks()

	// Set model values from response
	// v4
	state.AgentsIpv4, _ = types.ListValueFrom(ctx, types.StringType, agents.GetPrefixesIpv4())
	state.APIIpv4, _ = types.ListValueFrom(ctx, types.StringType, api.GetPrefixesIpv4())
	state.APMIpv4, _ = types.ListValueFrom(ctx, types.StringType, apm.GetPrefixesIpv4())
	state.LogsIpv4, _ = types.ListValueFrom(ctx, types.StringType, logs.GetPrefixesIpv4())
	state.OrchestratorIpv4, _ = types.ListValueFrom(ctx, types.StringType, orchestrator.GetPrefixesIpv4())
	state.ProcessIpv4, _ = types.ListValueFrom(ctx, types.StringType, process.GetPrefixesIpv4())
	state.SyntheticsIpv4, _ = types.ListValueFrom(ctx, types.StringType, synthetics.GetPrefixesIpv4())
	state.WebhooksIpv4, _ = types.ListValueFrom(ctx, types.StringType, webhook.GetPrefixesIpv4())
	// v6
	state.AgentsIpv6, _ = types.ListValueFrom(ctx, types.StringType, agents.GetPrefixesIpv6())
	state.APIIpv6, _ = types.ListValueFrom(ctx, types.StringType, api.GetPrefixesIpv6())
	state.APMIpv6, _ = types.ListValueFrom(ctx, types.StringType, apm.GetPrefixesIpv6())
	state.LogsIpv6, _ = types.ListValueFrom(ctx, types.StringType, logs.GetPrefixesIpv6())
	state.OrchestratorIpv6, _ = types.ListValueFrom(ctx, types.StringType, orchestrator.GetPrefixesIpv6())
	state.ProcessIpv6, _ = types.ListValueFrom(ctx, types.StringType, process.GetPrefixesIpv6())
	state.SyntheticsIpv6, _ = types.ListValueFrom(ctx, types.StringType, synthetics.GetPrefixesIpv6())
	state.WebhooksIpv6, _ = types.ListValueFrom(ctx, types.StringType, webhook.GetPrefixesIpv6())

	ipv4PrefixesByLocationMap := make(map[string]string)
	ipv6PrefixesByLocationMap := make(map[string]string)

	ipv4PrefixesByLocation := synthetics.GetPrefixesIpv4ByLocation()
	ipv6PrefixesByLocation := synthetics.GetPrefixesIpv6ByLocation()

	for key, value := range ipv4PrefixesByLocation {
		ipv4PrefixesByLocationMap[key] = strings.Join(value, ",")
	}

	for key, value := range ipv6PrefixesByLocation {
		ipv6PrefixesByLocationMap[key] = strings.Join(value, ",")
	}

	syntheticsIpv4ByLocation, diags := types.MapValueFrom(ctx, types.StringType, ipv4PrefixesByLocationMap)
	state.SyntheticsIpv4ByLocation = syntheticsIpv4ByLocation
	response.Diagnostics.Append(diags...)

	SyntheticsIpv6ByLocation, diags := types.MapValueFrom(ctx, types.StringType, ipv6PrefixesByLocationMap)
	state.SyntheticsIpv6ByLocation = SyntheticsIpv6ByLocation
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
