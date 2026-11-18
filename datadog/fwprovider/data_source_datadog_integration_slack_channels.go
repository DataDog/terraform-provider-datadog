package fwprovider

import (
	"context"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogIntegrationSlackChannelsDataSource{}
)

type datadogIntegrationSlackChannelsDataSource struct {
	Api  *datadogV1.SlackIntegrationApi
	Auth context.Context
}

type datadogIntegrationSlackChannelsDataSourceModel struct {
	ID          types.String                   `tfsdk:"id"`
	AccountName types.String                   `tfsdk:"account_name"`
	Channels    []integrationSlackChannelModel `tfsdk:"channels"`
}

type integrationSlackChannelModel struct {
	Name    types.String                         `tfsdk:"name"`
	Display *integrationSlackChannelDisplayModel `tfsdk:"display"`
}

type integrationSlackChannelDisplayModel struct {
	Message     types.Bool `tfsdk:"message"`
	Notified    types.Bool `tfsdk:"notified"`
	Snapshot    types.Bool `tfsdk:"snapshot"`
	Tags        types.Bool `tfsdk:"tags"`
	MuteButtons types.Bool `tfsdk:"mute_buttons"`
}

func NewDatadogIntegrationSlackChannelsDataSource() datasource.DataSource {
	return &datadogIntegrationSlackChannelsDataSource{}
}

func (d *datadogIntegrationSlackChannelsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetSlackIntegrationApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogIntegrationSlackChannelsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "integration_slack_channels"
}

func (d *datadogIntegrationSlackChannelsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to list the Slack channels configured for a Datadog Slack integration account.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"account_name": schema.StringAttribute{
				Required:    true,
				Description: "Slack account name.",
			},
			"channels": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of Slack channels configured for the account.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Slack channel name.",
						},
						"display": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Configuration options for what is shown in an alert event message.",
							Attributes: map[string]schema.Attribute{
								"message": schema.BoolAttribute{
									Computed:    true,
									Description: "Show the main body of the alert event.",
								},
								"notified": schema.BoolAttribute{
									Computed:    true,
									Description: "Show the list of @-handles in the alert event.",
								},
								"snapshot": schema.BoolAttribute{
									Computed:    true,
									Description: "Show the alert event's snapshot image.",
								},
								"tags": schema.BoolAttribute{
									Computed:    true,
									Description: "Show the scopes on which the monitor alerted.",
								},
								"mute_buttons": schema.BoolAttribute{
									Computed:    true,
									Description: "Show interactive buttons to mute the alerting monitor.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *datadogIntegrationSlackChannelsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogIntegrationSlackChannelsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	accountName := state.AccountName.ValueString()

	ddResp, httpResp, err := d.Api.GetSlackIntegrationChannels(d.Auth, accountName)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error listing Slack integration channels"), ""))
		return
	}

	sort.Slice(ddResp, func(i, j int) bool { return ddResp[i].GetName() < ddResp[j].GetName() })
	state.Channels = make([]integrationSlackChannelModel, 0, len(ddResp))
	for _, channel := range ddResp {
		if err := utils.CheckForUnparsed(channel); err != nil {
			response.Diagnostics.AddError("response contains unparsedObject", err.Error())
			return
		}
		channelModel := integrationSlackChannelModel{
			Name: types.StringValue(channel.GetName()),
		}
		// The API defaults every display flag to true, so a missing value must stay null rather than read as false.
		if display, ok := channel.GetDisplayOk(); ok {
			channelModel.Display = &integrationSlackChannelDisplayModel{
				Message:     types.BoolPointerValue(display.Message),
				Notified:    types.BoolPointerValue(display.Notified),
				Snapshot:    types.BoolPointerValue(display.Snapshot),
				Tags:        types.BoolPointerValue(display.Tags),
				MuteButtons: types.BoolPointerValue(display.MuteButtons),
			}
		}
		state.Channels = append(state.Channels, channelModel)
	}
	state.ID = types.StringValue(accountName)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
