package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &hostsDataSource{}

func NewHostsDataSource() datasource.DataSource {
	return &hostsDataSource{}
}

type HostListMetadataModel struct {
	AgentVersion    types.String `tfsdk:"agent_version"`
	CPU_cores       types.Int64  `tfsdk:"cpu_cores"`
	Gohai           types.String `tfsdk:"gohai"`
	Machine         types.String `tfsdk:"machine"`
	Platform        types.String `tfsdk:"platform"`
	Processor       types.String `tfsdk:"processor"`
	PythonV         types.String `tfsdk:"python_version"`
	Socket_FQDN     types.String `tfsdk:"socket_fqdn"`
	Socket_Hostname types.String `tfsdk:"socket_hostname"`
}

type HostListMetricsModel struct {
	CPU    types.Float64 `tfsdk:"cpu"`
	IOWait types.Float64 `tfsdk:"iowait"`
	Load   types.Float64 `tfsdk:"load"`
}

type HostListModel struct {
	Aliases          types.List   `tfsdk:"aliases"`
	Apps             types.List   `tfsdk:"apps"`
	AWSName          types.String `tfsdk:"aws_name"`
	HostName         types.String `tfsdk:"host_name"`
	ID               types.Int64  `tfsdk:"id"`
	IsMuted          types.Bool   `tfsdk:"is_muted"`
	LastReportedTime types.Int64  `tfsdk:"last_reported_time"`
	Meta             types.Object `tfsdk:"meta"`
	Metrics          types.Object `tfsdk:"metrics"`
	MuteTimeout      types.Int64  `tfsdk:"mute_timeout"`
	Name             types.String `tfsdk:"name"`
	Sources          types.List   `tfsdk:"sources"`
	Up               types.Bool   `tfsdk:"up"`
}

type HostsDataSourceModel struct {
	ID types.String `tfsdk:"id"`
	// Query Parameters
	Filter                types.String `tfsdk:"filter"`
	SortField             types.String `tfsdk:"sort_field"`
	SortDir               types.String `tfsdk:"sort_dir"`
	From                  types.Int64  `tfsdk:"from"`
	IncludeMutedHostsData types.Bool   `tfsdk:"include_muted_hosts_data"`
	IncludeHostsMetadata  types.Bool   `tfsdk:"include_hosts_metadata"`
	// Results
	HostList      []HostListModel `tfsdk:"host_list"`
	TotalMatching types.Int64     `tfsdk:"total_matching"`
	TotalReturned types.Int64     `tfsdk:"total_returned"`
}

type hostsDataSource struct {
	Api  *datadogV1.HostsApi
	Auth context.Context
}

func (d *hostsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"")
		return
	}

	d.Api = providerData.DatadogApiInstances.GetHostsApiV1()
	d.Auth = providerData.Auth
}

func (d *hostsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "hosts"
}

func (d *hostsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about your live hosts in Datadog.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"filter": schema.StringAttribute{
				Description: "String to filter search results.",
				Optional:    true,
			},
			"sort_field": schema.StringAttribute{
				Description: "Sort hosts by this field.",
				Optional:    true,
			},
			"sort_dir": schema.StringAttribute{
				Description: "Direction of sort.",
				Optional:    true,
			},
			"from": schema.Int64Attribute{
				Description: "Number of seconds since UNIX epoch from which you want to search your hosts.",
				Optional:    true,
			},
			"include_muted_hosts_data": schema.BoolAttribute{
				Description: "Include information on the muted status of hosts and when the mute expires.",
				Optional:    true,
			},
			"include_hosts_metadata": schema.BoolAttribute{
				Description: "Include additional metadata about the hosts (agent_version, machine, platform, processor, etc.).",
				Optional:    true,
			},
			// Datasource Results
			"host_list": schema.ListAttribute{
				Computed:    true,
				Description: "List of hosts.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":                 types.Int64Type,
						"aliases":            types.ListType{ElemType: types.StringType},
						"apps":               types.ListType{ElemType: types.StringType},
						"aws_name":           types.StringType,
						"host_name":          types.StringType,
						"is_muted":           types.BoolType,
						"last_reported_time": types.Int64Type,
						"mute_timeout":       types.Int64Type,
						"name":               types.StringType,
						"sources":            types.ListType{ElemType: types.StringType},
						"up":                 types.BoolType,
						"metrics": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"cpu":    types.Float64Type,
								"iowait": types.Float64Type,
								"load":   types.Float64Type,
							},
						},
						"meta": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"agent_version":   types.StringType,
								"cpu_cores":       types.Int64Type,
								"gohai":           types.StringType,
								"machine":         types.StringType,
								"platform":        types.StringType,
								"processor":       types.StringType,
								"python_version":  types.StringType,
								"socket_fqdn":     types.StringType,
								"socket_hostname": types.StringType,
							},
						},
					},
				},
			},
			"total_matching": schema.Int64Attribute{
				Description: "Number of host matching the query.",
				Computed:    true,
			},
			"total_returned": schema.Int64Attribute{
				Description: "Number of host returned.",
				Computed:    true,
			},
		},
		// Blocks: map[string]schema.Block{
		// 	"host_list": schema.ListNestedBlock{
		// 		NestedObject: schema.NestedBlockObject{
		// 			Attributes: map[string]schema.Attribute{
		// 				"id": schema.Int64Attribute{
		// 					Computed:    true,
		// 					Description: "The host ID.",
		// 				},
		// 				"aliases": schema.ListAttribute{
		// 					Computed:    true,
		// 					Description: "Host aliases collected by Datadog.",
		// 					ElementType: types.StringType,
		// 				},
		// 				"apps": schema.ListAttribute{
		// 					Computed:    true,
		// 					Description: "The Datadog integrations reporting metrics for the host.",
		// 					ElementType: types.StringType,
		// 				},
		// 				"aws_name": schema.StringAttribute{
		// 					Computed:    true,
		// 					Description: "AWS name of your host.",
		// 				},
		// 				"host_name": schema.StringAttribute{
		// 					Computed:    true,
		// 					Description: "The host name.",
		// 				},
		// 				"is_muted": schema.BoolAttribute{
		// 					Computed:    true,
		// 					Description: "Whether a host is muted.",
		// 				},
		// 				"last_reported_time": schema.Int64Attribute{
		// 					Computed:    true,
		// 					Description: "Last time the host reported a metric data point.",
		// 				},
		// 				"mute_timeout": schema.Int64Attribute{
		// 					Computed:    true,
		// 					Description: "Timeout of the mute applied to your host.",
		// 				},
		// 				"name": schema.StringAttribute{
		// 					Computed:    true,
		// 					Description: "The host name.",
		// 				},
		// 				"sources": schema.ListAttribute{
		// 					Computed:    true,
		// 					Description: "Source or cloud provider associated with your host.",
		// 					ElementType: types.StringType,
		// 				},
		// 				"up": schema.BoolAttribute{
		// 					Computed:    true,
		// 					Description: "Whether the expected metrics are received.",
		// 				},
		// 				"metrics": schema.ListAttribute{
		// 					Computed:    true,
		// 					Description: "Host Metrics collected.",
		// 					ElementType: types.ObjectType{
		// 						AttrTypes: map[string]attr.Type{
		// 							"cpu":    types.Float64Type,
		// 							"iowait": types.Float64Type,
		// 							"load":   types.Float64Type,
		// 						},
		// 					},
		// 				},
		// 				"meta": schema.ListAttribute{
		// 					Computed:    true,
		// 					Description: "Metadata associated with your host.",
		// 					ElementType: types.ObjectType{
		// 						AttrTypes: map[string]attr.Type{
		// 							"agent_version":   types.StringType,
		// 							"cpu_cores":       types.Int64Type,
		// 							"gohai":           types.StringType,
		// 							"machine":         types.StringType,
		// 							"platform":        types.StringType,
		// 							"processor":       types.StringType,
		// 							"python_version":  types.StringType,
		// 							"socket_fqdn":     types.StringType,
		// 							"socket_hostname": types.StringType,
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}
}

func (d *hostsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HostsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var parameters datadogV1.ListHostsOptionalParameters
	parameters.WithFilter(state.Filter.ValueString())
	parameters.WithCount(1000)
	parameters.WithIncludeHostsMetadata(state.IncludeHostsMetadata.ValueBool())
	parameters.WithIncludeMutedHostsData(state.IncludeMutedHostsData.ValueBool())
	if !state.SortField.IsNull() {
		parameters.WithSortField(state.SortField.ValueString())
	}
	if !state.SortDir.IsNull() {
		parameters.WithSortDir(state.SortDir.ValueString())
	}
	if !state.From.IsNull() {
		parameters.WithFrom(state.From.ValueInt64())
	}

	ddHostListResponse, _, err := d.Api.ListHosts(d.Auth, parameters)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting Hosts"))
		return
	}

	state.ID = types.StringValue("datadog-hosts")
	state.TotalMatching = basetypes.NewInt64Value(ddHostListResponse.GetTotalMatching())
	state.TotalReturned = basetypes.NewInt64Value(ddHostListResponse.GetTotalReturned())

	var hostList []HostListModel
	for _, val := range ddHostListResponse.HostList {
		var hostListEntry HostListModel
		hostListEntry.Aliases, _ = types.ListValueFrom(ctx, types.StringType, val.GetAliases())
		hostListEntry.Apps, _ = types.ListValueFrom(ctx, types.StringType, val.GetApps())
		hostListEntry.AWSName = basetypes.NewStringValue(val.GetAwsName())
		hostListEntry.HostName = basetypes.NewStringValue(val.GetHostName())
		hostListEntry.ID = basetypes.NewInt64Value(val.GetId())
		hostListEntry.IsMuted = basetypes.NewBoolValue(val.GetIsMuted())
		hostListEntry.LastReportedTime = basetypes.NewInt64Value(val.GetLastReportedTime())
		hostListEntry.MuteTimeout = basetypes.NewInt64Value(val.GetMuteTimeout())
		hostListEntry.Name = basetypes.NewStringValue(val.GetName())
		hostListEntry.Sources, _ = types.ListValueFrom(ctx, types.StringType, val.GetSources())
		hostListEntry.Up = basetypes.NewBoolValue(val.GetUp())
		// TODO add metadata and metrics
		if metrics, ok := val.GetMetricsOk(); ok {
			hostListEntry.Metrics, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
				"cpu":    types.Float64Type,
				"iowait": types.Float64Type,
				"load":   types.Float64Type,
			}, HostListMetricsModel{
				CPU:    basetypes.NewFloat64Value(metrics.GetCpu()),
				IOWait: basetypes.NewFloat64Value(metrics.GetIowait()),
				Load:   basetypes.NewFloat64Value(metrics.GetLoad()),
			})
		}
		if meta, ok := val.GetMetaOk(); ok {
			hostListEntry.Meta, _ = types.ObjectValueFrom(ctx, map[string]attr.Type{
				"agent_version":   types.StringType,
				"cpu_cores":       types.Int64Type,
				"gohai":           types.StringType,
				"machine":         types.StringType,
				"platform":        types.StringType,
				"processor":       types.StringType,
				"python_version":  types.StringType,
				"socket_fqdn":     types.StringType,
				"socket_hostname": types.StringType,
			}, HostListMetadataModel{
				AgentVersion:    basetypes.NewStringValue(meta.GetAgentVersion()),
				CPU_cores:       basetypes.NewInt64Value(meta.GetCpuCores()),
				Gohai:           basetypes.NewStringValue(meta.GetGohai()),
				Machine:         basetypes.NewStringValue(meta.GetMachine()),
				Platform:        basetypes.NewStringValue(meta.GetPlatform()),
				Processor:       basetypes.NewStringValue(meta.GetProcessor()),
				PythonV:         basetypes.NewStringValue(meta.GetPythonV()),
				Socket_FQDN:     basetypes.NewStringValue(meta.GetSocketFqdn()),
				Socket_Hostname: basetypes.NewStringValue(meta.GetSocketHostname()),
			})
		}
		hostList = append(hostList, hostListEntry)
	}
	state.HostList = hostList

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
