package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const resourceType = "monitor-notification-rule"

var (
	_ resource.ResourceWithConfigure   = &MonitorNotificationRuleResource{}
	_ resource.ResourceWithImportState = &MonitorNotificationRuleResource{}
)

type MonitorNotificationRuleResource struct {
	Api  *datadogV2.MonitorsApi
	Auth context.Context
}

type MonitorNotificationRuleModel struct {
	ID                                           types.String                                  `tfsdk:"id"`
	Name                                         types.String                                  `tfsdk:"name"`
	Recipients                                   types.Set                                     `tfsdk:"recipients"`
	MonitorNotificationRuleFilter                *MonitorNotificationRuleFilter                `tfsdk:"filter"`
	MonitorNotificationRuleConditionalRecipients *MonitorNotificationRuleConditionalRecipients `tfsdk:"conditional_recipients"`
}

type MonitorNotificationRuleFilter struct {
	Scope types.String `tfsdk:"scope"`
	Tags  types.Set    `tfsdk:"tags"`
}

type MonitorNotificationRuleConditionalRecipientsCondition struct {
	Scope      types.String `tfsdk:"scope"`
	Recipients types.Set    `tfsdk:"recipients"`
}

type MonitorNotificationRuleConditionalRecipients struct {
	MonitorNotificationRuleConditionalRecipientsConditions []MonitorNotificationRuleConditionalRecipientsCondition `tfsdk:"conditions"`
	FallbackRecipients                                     types.Set                                               `tfsdk:"fallback_recipients"`
}

func NewMonitorNotificationRuleResource() resource.Resource {
	return &MonitorNotificationRuleResource{}
}

func (r *MonitorNotificationRuleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			frameworkPath.MatchRoot("filter").AtName("tags"),
			frameworkPath.MatchRoot("filter").AtName("scope"),
		),
		resourcevalidator.ExactlyOneOf(
			frameworkPath.MatchRoot("recipients"),
			frameworkPath.MatchRoot("conditional_recipients"),
		),
	}
}

func (r *MonitorNotificationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMonitorsApiV2()
	r.Auth = providerData.Auth
}

func (r *MonitorNotificationRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "monitor_notification_rule"
}

func (r *MonitorNotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog MonitorNotificationRule resource.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the monitor notification rule.",
			},
			"recipients": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "List of recipients to notify. Cannot be used with `conditional_recipients`.",
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Description: "Specifies the matching criteria for monitor notifications.",
				Attributes: map[string]schema.Attribute{
					"tags": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "A list of tag key:value pairs (e.g. team:product). All tags must match (AND semantics).",
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
					"scope": schema.StringAttribute{
						Description: "A scope expression composed of `key:value` pairs (such as `env:prod`) with boolean operators (AND, OR, NOT) and parentheses for grouping.",
						Optional:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
			"conditional_recipients": schema.SingleNestedBlock{
				Description: "Use conditional recipients to define different recipients for different situations. Cannot be used with `recipients`.",
				Attributes: map[string]schema.Attribute{
					"fallback_recipients": schema.SetAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "If none of the `conditions` applied, `fallback_recipients` will get notified.",
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"conditions": schema.ListNestedBlock{
						Description: "Conditions of the notification rule.",
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"scope": schema.StringAttribute{
									Required:    true,
									Description: "Defines the condition under which the recipients are notified. Supported formats: Monitor status condition using `transition_type:<status>` (for example `transition_type:is_alert`) or a single tag `key:value pair` (for example `env:prod`).",
								},
								"recipients": schema.SetAttribute{
									ElementType: types.StringType,
									Description: "A list of recipients to notify. Uses the same format as the monitor message field. Must not start with an '@'.",
									Required:    true,
									Validators: []validator.Set{
										setvalidator.SizeAtLeast(1),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *MonitorNotificationRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

func (r *MonitorNotificationRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state MonitorNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetMonitorNotificationRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving MonitorNotificationRule"))
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *MonitorNotificationRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state MonitorNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	createRequest, diags := r.buildMonitorNotificationRuleCreateRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateMonitorNotificationRule(r.Auth, *createRequest)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating MonitorNotificationRule"))
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *MonitorNotificationRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state MonitorNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	updateRequest, diags := r.buildMonitorNotificationRuleUpdateRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateMonitorNotificationRule(r.Auth, id, *updateRequest)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating MonitorNotificationRule"))
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *MonitorNotificationRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state MonitorNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteMonitorNotificationRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting MonitorNotificationRule"))
		return
	}
}

func (r *MonitorNotificationRuleResource) updateState(ctx context.Context, state *MonitorNotificationRuleModel, resp *datadogV2.MonitorNotificationRuleResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	state.Name = types.StringValue(attributes.GetName())
	state.Recipients = fwutils.ToTerraformSetString(ctx, attributes.GetRecipientsOk)

	if filter, ok := attributes.GetFilterOk(); ok && filter != nil {
		state.MonitorNotificationRuleFilter = &MonitorNotificationRuleFilter{
			Scope: fwutils.ToTerraformStr(filter.MonitorNotificationRuleFilterScope.GetScopeOk()),
			Tags:  fwutils.ToTerraformSetString(ctx, filter.MonitorNotificationRuleFilterTags.GetTagsOk),
		}
	}
	r.updateConditionalRecipientsState(ctx, state, attributes)
}

func (r *MonitorNotificationRuleResource) updateConditionalRecipientsState(ctx context.Context, state *MonitorNotificationRuleModel, attributes datadogV2.MonitorNotificationRuleResponseAttributes) {
	conditionalRecipients, ok := attributes.GetConditionalRecipientsOk()
	if !ok || conditionalRecipients == nil {
		return
	}

	conditionsPtr, ok := conditionalRecipients.GetConditionsOk()
	if !ok || conditionsPtr == nil {
		// In practice, will never hit this scenario
		return
	}

	conditions := *conditionsPtr
	conditionsState := make([]MonitorNotificationRuleConditionalRecipientsCondition, 0, len(conditions))
	for _, condition := range conditions {
		conditionState := MonitorNotificationRuleConditionalRecipientsCondition{
			Scope:      fwutils.ToTerraformStr(condition.GetScopeOk()),
			Recipients: fwutils.ToTerraformSetString(ctx, condition.GetRecipientsOk),
		}
		conditionsState = append(conditionsState, conditionState)
	}

	state.MonitorNotificationRuleConditionalRecipients = &MonitorNotificationRuleConditionalRecipients{
		MonitorNotificationRuleConditionalRecipientsConditions: conditionsState,
		FallbackRecipients: fwutils.ToTerraformSetString(ctx, conditionalRecipients.GetFallbackRecipientsOk),
	}
}

func (r *MonitorNotificationRuleResource) buildRequestAttributes(ctx context.Context, state *MonitorNotificationRuleModel) (*datadogV2.MonitorNotificationRuleAttributes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewMonitorNotificationRuleAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())

	notificationRuleFilter := datadogV2.MonitorNotificationRuleFilter{}
	if !state.MonitorNotificationRuleFilter.Scope.IsNull() {
		scopeFilter := datadogV2.MonitorNotificationRuleFilterScope{}
		fwutils.SetOptString(state.MonitorNotificationRuleFilter.Scope, scopeFilter.SetScope)
		notificationRuleFilter.MonitorNotificationRuleFilterScope = &scopeFilter
	} else if !state.MonitorNotificationRuleFilter.Tags.IsNull() {
		tagsFilter := datadogV2.MonitorNotificationRuleFilterTags{}
		fwutils.SetOptStringList(state.MonitorNotificationRuleFilter.Tags, tagsFilter.SetTags, ctx)
		notificationRuleFilter.MonitorNotificationRuleFilterTags = &tagsFilter
	}
	attributes.SetFilter(notificationRuleFilter)

	fwutils.SetOptStringList(state.Recipients, attributes.SetRecipients, ctx)

	if conditionalRecipientsStruct := r.buildConditionalRecipientsRequest(ctx, state.MonitorNotificationRuleConditionalRecipients); conditionalRecipientsStruct != nil {
		attributes.SetConditionalRecipients(*conditionalRecipientsStruct)
	}

	return attributes, diags
}

func (r *MonitorNotificationRuleResource) buildMonitorNotificationRuleCreateRequest(ctx context.Context, state *MonitorNotificationRuleModel) (*datadogV2.MonitorNotificationRuleCreateRequest, diag.Diagnostics) {
	attributes, diags := r.buildRequestAttributes(ctx, state)

	data := datadogV2.NewMonitorNotificationRuleCreateRequestDataWithDefaults()
	data.SetType(resourceType)
	data.SetAttributes(*attributes)

	req := datadogV2.NewMonitorNotificationRuleCreateRequestWithDefaults()
	req.SetData(*data)
	return req, diags
}

func (r *MonitorNotificationRuleResource) buildMonitorNotificationRuleUpdateRequest(ctx context.Context, state *MonitorNotificationRuleModel) (*datadogV2.MonitorNotificationRuleUpdateRequest, diag.Diagnostics) {
	attributes, diags := r.buildRequestAttributes(ctx, state)

	data := datadogV2.NewMonitorNotificationRuleUpdateRequestDataWithDefaults()
	data.SetId(state.ID.ValueString())
	data.SetType(resourceType)
	data.SetAttributes(*attributes)

	req := datadogV2.NewMonitorNotificationRuleUpdateRequestWithDefaults()
	req.SetData(*data)
	return req, diags
}

func (r *MonitorNotificationRuleResource) buildConditionalRecipientsRequest(ctx context.Context, conditionalRecipients *MonitorNotificationRuleConditionalRecipients) *datadogV2.MonitorNotificationRuleConditionalRecipients {
	if conditionalRecipients == nil {
		return nil
	}
	conditionalRecipientsReq := datadogV2.MonitorNotificationRuleConditionalRecipients{}
	conditionsReq := []datadogV2.MonitorNotificationRuleCondition{}
	for _, condition := range conditionalRecipients.MonitorNotificationRuleConditionalRecipientsConditions {
		conditionReq := datadogV2.MonitorNotificationRuleCondition{}
		fwutils.SetOptStringList(condition.Recipients, conditionReq.SetRecipients, ctx)
		fwutils.SetOptString(condition.Scope, conditionReq.SetScope)
		conditionsReq = append(conditionsReq, conditionReq)
	}
	conditionalRecipientsReq.SetConditions(conditionsReq)
	fwutils.SetOptStringList(conditionalRecipients.FallbackRecipients, conditionalRecipientsReq.SetFallbackRecipients, ctx)
	return &conditionalRecipientsReq
}
