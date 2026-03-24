package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &onCallUserNotificationChannel{}
	_ resource.ResourceWithImportState = &onCallUserNotificationChannel{}
)

type onCallUserNotificationChannel struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

type onCallUserNotificationChannelModel struct {
	ID     types.String                        `tfsdk:"id"`
	UserID types.String                        `tfsdk:"user_id"`
	Email  *userNotificationChannelEmailConfig `tfsdk:"email"`
	Phone  *userNotificationChannelPhoneConfig `tfsdk:"phone"`
}

type userNotificationChannelEmailConfig struct {
	Address types.String   `tfsdk:"address"`
	Formats []types.String `tfsdk:"formats"`
}

type userNotificationChannelPhoneConfig struct {
	Number types.String `tfsdk:"number"`
}

func NewOnCallUserNotificationChannelResource() resource.Resource {
	return &onCallUserNotificationChannel{}
}

func (c *onCallUserNotificationChannel) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	c.Api = providerData.DatadogApiInstances.GetOnCallApiV2()
	c.Auth = providerData.Auth
}

func (c *onCallUserNotificationChannel) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "on_call_user_notification_channel"
}

func (c *onCallUserNotificationChannel) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	// NOTE: The user notification channel resource does not support updates.
	// To make this work in the TF provider, all attributes below must have a `RequiresReplace` plan modifier
	response.Schema = schema.Schema{
		Description: "Provides a Datadog On-Call user notification channel resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"user_id": schema.StringAttribute{
				Computed:    false,
				Required:    true,
				Description: "ID of the user to associate the notification channel with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"phone": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"number": schema.StringAttribute{
						Optional:    true,
						Description: "The E-164 formatted phone number (for example, +3371234567)",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"email": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Optional:    true,
						Description: "The e-mail address to be notified",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"formats": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "Preferred content formats for notifications",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

func (c *onCallUserNotificationChannel) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallUserNotificationChannelModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(plan.Validate()...)
	if response.Diagnostics.HasError() {
		return
	}

	userID := plan.UserID.ValueString()
	body := c.makeUserNotificationChannelRequestBody(&plan)

	resp, _, err := c.Api.CreateUserNotificationChannel(c.Auth, userID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating user notification channel"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state, err := c.makeUserNotificationChannelStateFromResponse(userID, &resp)
	if err != nil {
		response.Diagnostics.AddError("response object was invalid", err.Error())
	}
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (c *onCallUserNotificationChannel) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state onCallUserNotificationChannelModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	userID := state.UserID.ValueString()

	resp, httpResp, err := c.Api.GetUserNotificationChannel(c.Auth, userID, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving user notification channel"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	newState, err := c.makeUserNotificationChannelStateFromResponse(userID, &resp)
	if err != nil {
		response.Diagnostics.AddError("response object was invalid", err.Error())
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (c *onCallUserNotificationChannel) Update(_ context.Context, _ resource.UpdateRequest, response *resource.UpdateResponse) {
	// This is a no-op as channels cannot currently be updated.
	// This should not be called as all attributes are marked as require-replace
	response.Diagnostics.AddError(
		"updates are not supported",
		"updates to the user notification channel are not supported.  they must be deleted and recreated",
	)
}

func (c *onCallUserNotificationChannel) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state onCallUserNotificationChannelModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	userID := state.UserID.ValueString()

	httpResp, err := c.Api.DeleteUserNotificationChannel(c.Auth, userID, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error removing user notification channel"))
		return
	}
}

func (c *onCallUserNotificationChannel) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user_id,channel_id. Got: %q", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_id"), idParts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func (c *onCallUserNotificationChannel) makeUserNotificationChannelRequestBody(state *onCallUserNotificationChannelModel) *datadogV2.CreateUserNotificationChannelRequest {
	attributes := datadogV2.NewCreateNotificationChannelAttributesWithDefaults()

	if phone := state.Phone; phone != nil {
		phoneConfig := datadogV2.NewCreatePhoneNotificationChannelConfigWithDefaults()
		phoneConfig.SetNumber(phone.Number.ValueString())

		attributes.SetConfig(datadogV2.CreatePhoneNotificationChannelConfigAsCreateNotificationChannelConfig(phoneConfig))
	}

	if email := state.Email; email != nil {
		formats := make([]datadogV2.NotificationChannelEmailFormatType, len(email.Formats))
		for i := range email.Formats {
			formats[i] = datadogV2.NotificationChannelEmailFormatType(email.Formats[i].ValueString())
		}

		emailConfig := datadogV2.NewCreateEmailNotificationChannelConfigWithDefaults()
		emailConfig.SetAddress(email.Address.ValueString())
		emailConfig.SetFormats(formats)

		attributes.SetConfig(datadogV2.CreateEmailNotificationChannelConfigAsCreateNotificationChannelConfig(emailConfig))
	}

	data := datadogV2.NewCreateNotificationChannelDataWithDefaults()
	data.SetAttributes(*attributes)

	req := datadogV2.NewCreateUserNotificationChannelRequestWithDefaults()
	req.SetData(*data)

	return req
}

func (c *onCallUserNotificationChannel) makeUserNotificationChannelStateFromResponse(userID string, channel *datadogV2.NotificationChannel) (*onCallUserNotificationChannelModel, error) {
	channelData, ok := channel.GetDataOk()
	if !ok {
		return nil, fmt.Errorf("missing data in response object")
	}

	state := &onCallUserNotificationChannelModel{}
	state.ID = types.StringValue(channelData.GetId())
	state.UserID = types.StringValue(userID)

	attributes := channelData.GetAttributes()

	channelConfig, ok := attributes.GetConfigOk()
	if !ok {
		return nil, fmt.Errorf("missing config in channel data")
	}

	switch cfg := channelConfig.GetActualInstance().(type) {
	case *datadogV2.NotificationChannelPhoneConfig:
		state.Phone = &userNotificationChannelPhoneConfig{
			Number: types.StringValue(cfg.Number),
		}
	case *datadogV2.NotificationChannelEmailConfig:
		formats := make([]types.String, len(cfg.Formats))
		for i := range cfg.Formats {
			formats[i] = types.StringValue(string(cfg.Formats[i]))
		}

		state.Email = &userNotificationChannelEmailConfig{
			Address: types.StringValue(cfg.Address),
			Formats: formats,
		}
	case *datadogV2.NotificationChannelPushConfig:
		return nil, fmt.Errorf("push channels are not supported")
	}

	return state, nil
}

func (m *onCallUserNotificationChannelModel) Validate() diag.Diagnostics {
	diags := diag.Diagnostics{}

	configPath := path.Root("config")
	if m.Phone == nil && m.Email == nil {
		diags.AddAttributeError(configPath, "missing configuration", "config must specify one of email or phone")
	}

	if m.Phone != nil {
		if m.Phone.Number.IsNull() {
			phonePath := configPath.AtName("phone")
			diags.AddAttributeError(phonePath, "missing number", "number is required")
		}
	}

	if m.Email != nil {
		emailPath := configPath.AtName("email")

		if m.Email.Address.IsNull() {
			diags.AddAttributeError(emailPath, "missing address", "address is required")
		}

		if len(m.Email.Formats) == 0 {
			diags.AddAttributeError(emailPath, "missing formats", "formats are required")
		}

		for _, v := range m.Email.Formats {
			_, err := datadogV2.NewNotificationChannelEmailFormatTypeFromValue(v.ValueString())
			if err != nil {
				diags.AddAttributeError(emailPath, "invalid format value", err.Error())
			}
		}
	}

	return diags
}
