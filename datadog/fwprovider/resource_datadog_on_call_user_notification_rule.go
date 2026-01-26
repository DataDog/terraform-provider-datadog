package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &onCallUserNotificationRule{}
	_ resource.ResourceWithImportState = &onCallUserNotificationRule{}
)

type onCallUserNotificationRule struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

type onCallUserNotificationRuleModel struct {
	ID           types.String                     `tfsdk:"id"`
	UserID       types.String                     `tfsdk:"user_id"`
	ChannelID    types.String                     `tfsdk:"channel_id"`
	Category     types.String                     `tfsdk:"category"`
	DelayMinutes types.Int32                      `tfsdk:"delay_minutes"`
	Phone        *userNotificationRulePhoneConfig `tfsdk:"phone"`
}

type userNotificationRulePhoneConfig struct {
	Method types.String `tfsdk:"method"`
}

func NewOnCallUserNotificationRuleResource() resource.Resource {
	return &onCallUserNotificationRule{}
}

func (r *onCallUserNotificationRule) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetOnCallApiV2()
	r.Auth = providerData.Auth
}

func (r *onCallUserNotificationRule) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "on_call_user_notification_rule"
}

func (r *onCallUserNotificationRule) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog On-Call user notification rule resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the user to associate the notification rule with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"channel_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the notification channel to associate the notification rule with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"category": schema.StringAttribute{
				Required:    true,
				Description: "Notification category to associate the rule with.",
				Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewOnCallNotificationRuleCategoryFromValue)},
			},
			"delay_minutes": schema.Int32Attribute{
				Required:    true,
				Description: "Number of minutes to elapse before this rule is evaluated.  `0` indicates immediate evaluation.",
				Validators:  []validator.Int32{int32validator.AtLeast(0)},
			},
		},
		Blocks: map[string]schema.Block{
			"phone": schema.SingleNestedBlock{
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Optional:    true,
						Description: "Specifies the method in which a phone is used in a notification rule.",
						Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewOnCallPhoneNotificationRuleMethodFromValue)},
					},
				},
			},
		},
	}
}

func (r *onCallUserNotificationRule) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallUserNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	userID := plan.UserID.ValueString()
	channelID := plan.ChannelID.ValueString()

	body := r.makeCreateUserNotificationRuleRequestBody(channelID, &plan)

	resp, _, err := r.Api.CreateUserNotificationRule(r.Auth, userID, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating user notification rule"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state := r.makeUserNotificationRuleStateFromResponse(userID, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallUserNotificationRule) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state onCallUserNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	userID := state.UserID.ValueString()

	resp, httpResp, err := r.Api.GetUserNotificationRule(r.Auth, userID, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving user notification rule"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	newState := r.makeUserNotificationRuleStateFromResponse(userID, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *onCallUserNotificationRule) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallUserNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	userID := plan.UserID.ValueString()
	channelID := plan.ChannelID.ValueString()

	body := r.makeUpdateUserNotificationRuleRequestBody(id, channelID, &plan)

	resp, _, err := r.Api.UpdateUserNotificationRule(r.Auth, userID, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating user notification rule"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state := r.makeUserNotificationRuleStateFromResponse(userID, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallUserNotificationRule) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state onCallUserNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	userID := state.UserID.ValueString()

	httpResp, err := r.Api.DeleteUserNotificationRule(r.Auth, userID, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error removing user notification channel"))
		return
	}
}

func (r *onCallUserNotificationRule) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	idParts := strings.Split(request.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user_id,rule_id. Got: %q", request.ID),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_id"), idParts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}

func (r *onCallUserNotificationRule) makeCreateUserNotificationRuleRequestBody(channelID string, state *onCallUserNotificationRuleModel) *datadogV2.CreateOnCallNotificationRuleRequest {
	channelRelationshipData := datadogV2.NewOnCallNotificationRuleChannelRelationshipDataWithDefaults()
	channelRelationshipData.SetId(channelID)

	relationship := datadogV2.NewOnCallNotificationRuleChannelRelationshipWithDefaults()
	relationship.SetData(*channelRelationshipData)

	relationships := datadogV2.NewOnCallNotificationRuleRelationshipsWithDefaults()
	relationships.SetChannel(*relationship)

	attributes := datadogV2.NewOnCallNotificationRuleRequestAttributesWithDefaults()
	attributes.SetCategory(datadogV2.OnCallNotificationRuleCategory(state.Category.ValueString()))
	attributes.SetDelayMinutes(int64(state.DelayMinutes.ValueInt32()))

	if state.Phone != nil {
		phoneChannelSettings := datadogV2.NewOnCallPhoneNotificationRuleSettingsWithDefaults()
		phoneChannelSettings.SetMethod(datadogV2.OnCallPhoneNotificationRuleMethod(state.Phone.Method.ValueString()))

		attributes.SetChannelSettings(datadogV2.OnCallPhoneNotificationRuleSettingsAsOnCallNotificationRuleChannelSettings(phoneChannelSettings))
	}

	data := datadogV2.NewCreateOnCallNotificationRuleRequestDataWithDefaults()
	data.SetAttributes(*attributes)
	data.SetRelationships(*relationships)

	req := datadogV2.NewCreateOnCallNotificationRuleRequestWithDefaults()
	req.SetData(*data)

	return req
}

func (r *onCallUserNotificationRule) makeUpdateUserNotificationRuleRequestBody(id string, channelID string, state *onCallUserNotificationRuleModel) *datadogV2.UpdateOnCallNotificationRuleRequest {
	channelRelationshipData := datadogV2.NewOnCallNotificationRuleChannelRelationshipDataWithDefaults()
	channelRelationshipData.SetId(channelID)

	relationship := datadogV2.NewOnCallNotificationRuleChannelRelationshipWithDefaults()
	relationship.SetData(*channelRelationshipData)

	relationships := datadogV2.NewOnCallNotificationRuleRelationshipsWithDefaults()
	relationships.SetChannel(*relationship)

	attributes := datadogV2.NewUpdateOnCallNotificationRuleRequestAttributesWithDefaults()
	attributes.SetCategory(datadogV2.OnCallNotificationRuleCategory(state.Category.ValueString()))
	attributes.SetDelayMinutes(int64(state.DelayMinutes.ValueInt32()))

	if state.Phone != nil {
		phoneChannelSettings := datadogV2.NewOnCallPhoneNotificationRuleSettingsWithDefaults()
		phoneChannelSettings.SetMethod(datadogV2.OnCallPhoneNotificationRuleMethod(state.Phone.Method.ValueString()))

		attributes.SetChannelSettings(datadogV2.OnCallPhoneNotificationRuleSettingsAsOnCallNotificationRuleChannelSettings(phoneChannelSettings))
	}

	data := datadogV2.NewUpdateOnCallNotificationRuleRequestDataWithDefaults()
	data.SetAttributes(*attributes)
	data.SetRelationships(*relationships)
	data.SetId(id)

	req := datadogV2.NewUpdateOnCallNotificationRuleRequestWithDefaults()
	req.SetData(*data)

	return req
}

func (r *onCallUserNotificationRule) makeUserNotificationRuleStateFromResponse(userID string, rule *datadogV2.OnCallNotificationRule) *onCallUserNotificationRuleModel {
	data := rule.GetData()

	relationships := data.GetRelationships()
	attributes := data.GetAttributes()

	state := &onCallUserNotificationRuleModel{}
	state.ID = types.StringValue(data.GetId())
	state.UserID = types.StringValue(userID)
	state.ChannelID = types.StringValue(relationships.Channel.Data.GetId())
	state.Category = types.StringValue(string(attributes.GetCategory()))
	state.DelayMinutes = types.Int32Value(int32(attributes.GetDelayMinutes()))

	channelSettings := attributes.GetChannelSettings()

	switch cfg := channelSettings.GetActualInstance().(type) {
	case *datadogV2.OnCallPhoneNotificationRuleSettings:
		state.Phone = &userNotificationRulePhoneConfig{
			Method: types.StringValue(string(cfg.GetMethod())),
		}
	}

	return state
}
