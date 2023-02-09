package datadog

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &IPRangesDataSource{}

func NewIPRangesDataSource() datasource.DataSource {
	return &IPRangesDataSource{}
}

type iPRangesDataSourceZoneModel struct {
	ID types.String `tfsdk:"id"`
	// v4
	AgentsIpv4               types.List `tfsdk:"agents_ipv4"`
	APIIpv4                  types.List `tfsdk:"api_ipv4"`
	APMIpv4                  types.List `tfsdk:"apm_ipv4"`
	LogsIpv4                 types.List `tfsdk:"logs_ipv4"`
	ProcessIpv4              types.List `tfsdk:"process_ipv4"`
	SyntheticsIpv4           types.List `tfsdk:"synthetics_ipv4"`
	SyntheticsIpv4ByLocation types.Map  `tfsdk:"synthetics_ipv4_by_location"`
	WebhooksIpv4             types.List `tfsdk:"webhooks_ipv4"`
	// v6
	AgentsIpv6               types.List `tfsdk:"agents_ipv6"`
	APIIpv6                  types.List `tfsdk:"api_ipv6"`
	APMIpv6                  types.List `tfsdk:"apm_ipv6"`
	LogsIpv6                 types.List `tfsdk:"logs_ipv6"`
	ProcessIpv6              types.List `tfsdk:"process_ipv6"`
	SyntheticsIpv6           types.List `tfsdk:"synthetics_ipv6"`
	SyntheticsIpv6ByLocation types.Map  `tfsdk:"synthetics_ipv6_by_location"`
	WebhooksIpv6             types.List `tfsdk:"webhooks_ipv6"`
}

type IPRangesDataSource struct {
	Api  *datadogV1.IPRangesApi
	Auth context.Context
}

func (d *IPRangesDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*datadogFrameworkProvider)

	d.Api = providerData.DatadogApiInstances.GetIPRangesApiV1()
	d.Auth = providerData.Auth

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"")
		return
	}
	d.Api = providerData.DatadogApiInstances.GetIPRangesApiV1()
	d.Auth = providerData.Auth
}

func (d *IPRangesDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "ip_ranges"
}

func (d *IPRangesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about Datadog's IP addresses.",
		Attributes: map[string]schema.Attribute{
			// v4
			"agents_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the Agent endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the Agent endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"api_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the API endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the API endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"apm_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the APM endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the APM endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"logs_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the Logs endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the Logs endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"process_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the Process endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the Process endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"synthetics_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"synthetics_ipv4_by_location": schema.MapAttribute{
				Description:         "A map of IPv4 prefixes (string of concatenated IPs, delimited by ',') by location.",
				MarkdownDescription: "A map of IPv4 prefixes (string of concatenated IPs, delimited by ',') by location.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"webhooks_ipv4": schema.ListAttribute{
				Description:         "An Array of IPv4 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				MarkdownDescription: "An Array of IPv4 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			// v6
			"agents_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the Agent endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the Agent endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"api_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the API endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the API endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"apm_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the APM endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the APM endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"logs_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the Logs endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the Logs endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"process_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the Process endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the Process endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"synthetics_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the Synthetics endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"synthetics_ipv6_by_location": schema.MapAttribute{
				Description:         "A map of IPv6 prefixes (string of concatenated IPs, delimited by ',') by location.",
				MarkdownDescription: "A map of IPv6 prefixes (string of concatenated IPs, delimited by ',') by location.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"webhooks_ipv6": schema.ListAttribute{
				Description:         "An Array of IPv6 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				MarkdownDescription: "An Array of IPv6 addresses in CIDR format specifying the A records for the Webhooks endpoint.",
				Computed:            true,
				ElementType:         types.StringType,
			},

			"id": schema.StringAttribute{
				Description:         "Data source ID.",
				MarkdownDescription: "Data source ID.",
				Computed:            true,
			},
		},
	}
}

func (d *IPRangesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data iPRangesDataSourceZoneModel

	ipAddresses, _, err := d.Api.GetIPRanges(d.Auth)
	if err != nil {
		response.Diagnostics.AddError("error getting IPRanges", err.Error())
		return
	}
	if err := utils.CheckForUnparsed(ipAddresses); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data.ID = types.StringValue("datadog-ip-ranges")

	// v4 and v6
	ipAddressesPtr := &ipAddresses
	agents := ipAddressesPtr.GetAgents()
	api := ipAddressesPtr.GetApi()
	apm := ipAddressesPtr.GetApm()
	logs := ipAddressesPtr.GetLogs()
	process := ipAddressesPtr.GetProcess()
	synthetics := ipAddressesPtr.GetSynthetics()
	webhook := ipAddressesPtr.GetWebhooks()

	// Set model values from response
	// v4
	data.AgentsIpv4, _ = types.ListValueFrom(ctx, types.StringType, agents.GetPrefixesIpv4())
	data.APIIpv4, _ = types.ListValueFrom(ctx, types.StringType, api.GetPrefixesIpv4())
	data.APMIpv4, _ = types.ListValueFrom(ctx, types.StringType, apm.GetPrefixesIpv4())
	data.LogsIpv4, _ = types.ListValueFrom(ctx, types.StringType, logs.GetPrefixesIpv4())
	data.ProcessIpv4, _ = types.ListValueFrom(ctx, types.StringType, process.GetPrefixesIpv4())
	data.SyntheticsIpv4, _ = types.ListValueFrom(ctx, types.StringType, synthetics.GetPrefixesIpv4())
	data.WebhooksIpv4, _ = types.ListValueFrom(ctx, types.StringType, webhook.GetPrefixesIpv4())
	// v6
	data.AgentsIpv6, _ = types.ListValueFrom(ctx, types.StringType, agents.GetPrefixesIpv4())
	data.APIIpv6, _ = types.ListValueFrom(ctx, types.StringType, api.GetPrefixesIpv4())
	data.APMIpv6, _ = types.ListValueFrom(ctx, types.StringType, apm.GetPrefixesIpv4())
	data.LogsIpv6, _ = types.ListValueFrom(ctx, types.StringType, logs.GetPrefixesIpv4())
	data.ProcessIpv6, _ = types.ListValueFrom(ctx, types.StringType, process.GetPrefixesIpv4())
	data.SyntheticsIpv6, _ = types.ListValueFrom(ctx, types.StringType, synthetics.GetPrefixesIpv4())
	data.WebhooksIpv6, _ = types.ListValueFrom(ctx, types.StringType, webhook.GetPrefixesIpv4())

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

	syntheticsIpv4ByLocationState, _ := types.MapValueFrom(ctx, types.StringType, ipv4PrefixesByLocationMap)
	data.SyntheticsIpv4ByLocation, _ = types.MapValueFrom(ctx, types.StringType, syntheticsIpv4ByLocationState)

	syntheticsIpv6ByLocationState, _ := types.MapValueFrom(ctx, types.StringType, ipv4PrefixesByLocationMap)
	data.SyntheticsIpv6ByLocation, _ = types.MapValueFrom(ctx, types.StringType, syntheticsIpv6ByLocationState)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}
