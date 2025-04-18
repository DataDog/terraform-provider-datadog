package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	ID                            types.String                   `tfsdk:"id"`
	Name                          types.String                   `tfsdk:"name"`
	Recipients                    types.Set                      `tfsdk:"recipients"`
	MonitorNotificationRuleFilter *MonitorNotificationRuleFilter `tfsdk:"filter"`
}

type MonitorNotificationRuleFilter struct {
	Tags types.Set `tfsdk:"tags"`
}

func NewMonitorNotificationRuleResource() resource.Resource {
	return &MonitorNotificationRuleResource{}
}

func (r *MonitorNotificationRuleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			frameworkPath.MatchRoot("filter").AtName("tags"),
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
				Required:    true,
				ElementType: types.StringType,
				Description: "List of recipients to notify.",
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"tags": schema.SetAttribute{
						Required:    true,
						ElementType: types.StringType,
						Description: "Tags that all target monitors must match.",
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
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
	state.Recipients, _ = types.SetValueFrom(ctx, types.StringType, attributes.GetRecipients())

	if filter := attributes.GetFilter(); filter.MonitorNotificationRuleFilterTags != nil {
		tags, _ := types.SetValueFrom(ctx, types.StringType, filter.MonitorNotificationRuleFilterTags.GetTags())
		state.MonitorNotificationRuleFilter = &MonitorNotificationRuleFilter{
			Tags: tags,
		}
	}
}

func buildRequestAttributes(ctx context.Context, state *MonitorNotificationRuleModel) (*datadogV2.MonitorNotificationRuleAttributes, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewMonitorNotificationRuleAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())

	var recipients []string
	diags.Append(state.Recipients.ElementsAs(ctx, &recipients, false)...)
	attributes.SetRecipients(recipients)

	var tags []string
	diags.Append(state.MonitorNotificationRuleFilter.Tags.ElementsAs(ctx, &tags, false)...)
	filterTags := datadogV2.NewMonitorNotificationRuleFilterTags(tags)
	attributes.SetFilter(datadogV2.MonitorNotificationRuleFilter{MonitorNotificationRuleFilterTags: filterTags})

	return attributes, diags
}

func (r *MonitorNotificationRuleResource) buildMonitorNotificationRuleCreateRequest(ctx context.Context, state *MonitorNotificationRuleModel) (*datadogV2.MonitorNotificationRuleCreateRequest, diag.Diagnostics) {
	attributes, diags := buildRequestAttributes(ctx, state)

	data := datadogV2.NewMonitorNotificationRuleCreateRequestDataWithDefaults()
	data.SetType(resourceType)
	data.SetAttributes(*attributes)

	req := datadogV2.NewMonitorNotificationRuleCreateRequestWithDefaults()
	req.SetData(*data)
	return req, diags
}

func (r *MonitorNotificationRuleResource) buildMonitorNotificationRuleUpdateRequest(ctx context.Context, state *MonitorNotificationRuleModel) (*datadogV2.MonitorNotificationRuleUpdateRequest, diag.Diagnostics) {
	attributes, diags := buildRequestAttributes(ctx, state)

	data := datadogV2.NewMonitorNotificationRuleUpdateRequestDataWithDefaults()
	data.SetId(state.ID.ValueString())
	data.SetType(resourceType)
	data.SetAttributes(*attributes)

	req := datadogV2.NewMonitorNotificationRuleUpdateRequestWithDefaults()
	req.SetData(*data)
	return req, diags
}
