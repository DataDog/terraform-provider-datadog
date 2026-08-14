package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &statusPageComponentResource{}
	_ resource.ResourceWithImportState = &statusPageComponentResource{}
)

func NewStatusPageComponentResource() resource.Resource {
	return &statusPageComponentResource{}
}

type statusPageComponentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	StatusPageID types.String `tfsdk:"status_page_id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Position     types.Int64  `tfsdk:"position"`
	GroupID      types.String `tfsdk:"group_id"`
}

type statusPageComponentResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

func (r *statusPageComponentResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		return
	}
	r.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	r.Auth = providerData.Auth
}

func (r *statusPageComponentResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "status_page_component"
}

func (r *statusPageComponentResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Status Page Component resource. This can be used to create and manage a component or component group on a `datadog_status_page`.",
		Attributes: map[string]schema.Attribute{
			"status_page_id": schema.StringAttribute{
				Required:      true,
				Description:   "The ID of the status page this component belongs to.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the component or group.",
			},
			"type": schema.StringAttribute{
				Required:      true,
				Description:   "The component type. Valid values are `component`, `group`.",
				Validators:    []validator.String{stringvalidator.OneOf("component", "group")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"position": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Description:   "The ordering position of the component within its page or group.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"group_id": schema.StringAttribute{
				Optional:      true,
				Description:   "The ID of the parent group (a component of type `group`). Omit for top-level components and for groups.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *statusPageComponentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state statusPageComponentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, err := uuid.Parse(state.StatusPageID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status_page_id"))
		return
	}

	attrs := datadogV2.NewCreateComponentRequestDataAttributesWithDefaults()
	attrs.SetName(state.Name.ValueString())
	attrs.SetType(datadogV2.CreateComponentRequestDataAttributesType(state.Type.ValueString()))
	if !state.Position.IsNull() && !state.Position.IsUnknown() {
		attrs.SetPosition(state.Position.ValueInt64())
	}
	data := datadogV2.NewCreateComponentRequestData(*attrs, datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS)
	if !state.GroupID.IsNull() {
		gid, err := uuid.Parse(state.GroupID.ValueString())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid group_id"))
			return
		}
		groupData := datadogV2.NewCreateComponentRequestDataRelationshipsGroupDataWithDefaults()
		groupData.SetId(gid)
		group := datadogV2.NewCreateComponentRequestDataRelationshipsGroupWithDefaults()
		group.SetData(*groupData)
		rels := datadogV2.NewCreateComponentRequestDataRelationshipsWithDefaults()
		rels.SetGroup(*group)
		data.SetRelationships(*rels)
	}
	body := datadogV2.NewCreateComponentRequest()
	body.SetData(*data)

	resp, _, err := r.Api.CreateComponent(r.Auth, pid, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating status page component"))
		return
	}
	r.updateState(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageComponentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageComponentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, cid, err := parseComponentIDs(state.StatusPageID.ValueString(), state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid component ids"))
		return
	}
	resp, httpResp, err := r.Api.GetComponent(r.Auth, pid, cid)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving status page component"))
		return
	}
	r.updateState(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageComponentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state statusPageComponentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, cid, err := parseComponentIDs(state.StatusPageID.ValueString(), state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid component ids"))
		return
	}
	attrs := datadogV2.NewPatchComponentRequestDataAttributesWithDefaults()
	attrs.SetName(state.Name.ValueString())
	if !state.Position.IsNull() && !state.Position.IsUnknown() {
		attrs.SetPosition(state.Position.ValueInt64())
	}
	data := datadogV2.NewPatchComponentRequestData(*attrs, cid, datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS)
	body := datadogV2.NewPatchComponentRequest()
	body.SetData(*data)

	resp, _, err := r.Api.UpdateComponent(r.Auth, pid, cid, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating status page component"))
		return
	}
	r.updateState(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageComponentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageComponentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, cid, err := parseComponentIDs(state.StatusPageID.ValueString(), state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid component ids"))
		return
	}
	if _, err := r.Api.DeleteComponent(r.Auth, pid, cid); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting status page component"))
	}
}

func (r *statusPageComponentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.SplitN(request.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		response.Diagnostics.AddError("invalid import id", fmt.Sprintf("expected <status_page_id>:<component_id>, got %q", request.ID))
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("status_page_id"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("id"), parts[1])...)
}

func (r *statusPageComponentResource) updateState(state *statusPageComponentResourceModel, data datadogV2.StatusPagesComponentData) {
	attrs := data.GetAttributes()
	state.ID = types.StringValue(data.GetId().String())
	state.Name = types.StringValue(attrs.GetName())
	state.Type = types.StringValue(string(attrs.GetType()))
	state.Position = types.Int64Value(attrs.GetPosition())
	if rels, ok := data.GetRelationshipsOk(); ok {
		if grp, ok := rels.GetGroupOk(); ok {
			if gd, ok := grp.GetDataOk(); ok {
				state.GroupID = types.StringValue(gd.GetId().String())
			}
		}
	}
}

func parseComponentIDs(pageID, componentID string) (uuid.UUID, uuid.UUID, error) {
	pid, err := uuid.Parse(pageID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid status_page_id %q: %w", pageID, err)
	}
	cid, err := uuid.Parse(componentID)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("invalid component id %q: %w", componentID, err)
	}
	return pid, cid, nil
}
