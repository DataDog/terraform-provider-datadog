package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.ResourceWithConfigure      = &statusPageComponentResource{}
	_ resource.ResourceWithImportState    = &statusPageComponentResource{}
	_ resource.ResourceWithValidateConfig = &statusPageComponentResource{}
)

type statusPageComponentResource struct {
	Api  *datadogV2.StatusPagesApi
	Auth context.Context
}

type statusPageComponentModel struct {
	ID         types.String                  `tfsdk:"id"`
	PageID     types.String                  `tfsdk:"page_id"`
	Name       types.String                  `tfsdk:"name"`
	Type       types.String                  `tfsdk:"type"`
	Position   types.Int64                   `tfsdk:"position"`
	Status     types.String                  `tfsdk:"status"`
	Components []statusPageSubComponentModel `tfsdk:"components"`
	CreatedAt  types.String                  `tfsdk:"created_at"`
	ModifiedAt types.String                  `tfsdk:"modified_at"`
}

type statusPageSubComponentModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Position types.Int64  `tfsdk:"position"`
	Status   types.String `tfsdk:"status"`
}

func NewStatusPageComponentResource() resource.Resource {
	return &statusPageComponentResource{}
}

func (r *statusPageComponentResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "status_page_component"
}

func (r *statusPageComponentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog status page component resource. This can be used to create and manage components on a status page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the component.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"page_id": schema.StringAttribute{
				Description: "The ID of the status page this component belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the component.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of the component. Valid values are: component, group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"position": schema.Int64Attribute{
				Description: "The position of the component on the status page. Must be between `0` and the current number of existing components on the page, inclusive (i.e. it can append one past the current highest position, but cannot skip ahead further or be negative). A `position` value that depends on a sibling component being created first requires an explicit `depends_on` on that sibling to guarantee creation order.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the component.",
				Computed:    true,
			},
			"components": schema.ListNestedAttribute{
				Description: "The sub-components of a component of type `group`. Sub-components can only be set at creation time; changing this list requires replacing the component.",
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The ID of the sub-component.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the sub-component.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the sub-component. Valid value is: component.",
							Required:    true,
						},
						"position": schema.Int64Attribute{
							Description: "The position of the sub-component within the group.",
							Required:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the sub-component.",
							Computed:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp of when the component was created.",
				Computed:    true,
			},
			"modified_at": schema.StringAttribute{
				Description: "Timestamp of when the component was last modified.",
				Computed:    true,
			},
		},
	}
}

func (r *statusPageComponentResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetStatusPagesApiV2()
	r.Auth = providerData.Auth
}

func (r *statusPageComponentResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var cfg statusPageComponentModel
	response.Diagnostics.Append(request.Config.Get(ctx, &cfg)...)
	if response.Diagnostics.HasError() {
		return
	}

	switch cfg.Type.ValueString() {
	case "group":
		if len(cfg.Components) == 0 {
			response.Diagnostics.AddAttributeError(path.Root("components"), "group requires children",
				"A component of type `group` must declare at least one `components` block.")
		}
	case "component":
		if len(cfg.Components) > 0 {
			response.Diagnostics.AddAttributeError(path.Root("components"), "component cannot have children",
				"`components` blocks are only valid when `type` is `group`.")
		}
	}

	for i, sub := range cfg.Components {
		if subType := sub.Type.ValueString(); subType != "" && subType != "component" {
			response.Diagnostics.AddAttributeError(path.Root("components").AtListIndex(i).AtName("type"),
				"invalid sub-component type",
				fmt.Sprintf("Sub-components must have `type = \"component\"`, got %q. Nested groups are not supported.", subType))
		}
	}
}

func buildStatusPageComponentSubComponents(components []statusPageSubComponentModel) []datadogV2.CreateComponentRequestDataAttributesComponentsItems {
	if components == nil {
		return nil
	}
	result := make([]datadogV2.CreateComponentRequestDataAttributesComponentsItems, len(components))
	for i, sub := range components {
		result[i] = datadogV2.CreateComponentRequestDataAttributesComponentsItems{
			Name:     sub.Name.ValueString(),
			Position: sub.Position.ValueInt64(),
			Type:     datadogV2.StatusPagesComponentGroupAttributesComponentsItemsType(sub.Type.ValueString()),
		}
	}
	return result
}

func (r *statusPageComponentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan statusPageComponentModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(plan.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	attributes := datadogV2.CreateComponentRequestDataAttributes{
		Name:       plan.Name.ValueString(),
		Position:   plan.Position.ValueInt64(),
		Type:       datadogV2.CreateComponentRequestDataAttributesType(plan.Type.ValueString()),
		Components: buildStatusPageComponentSubComponents(plan.Components),
	}

	body := datadogV2.CreateComponentRequest{
		Data: &datadogV2.CreateComponentRequestData{
			Type:       datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS,
			Attributes: attributes,
		},
	}

	resp, httpResp, err := r.Api.CreateComponent(r.Auth, pageID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error creating status page component",
			fmt.Sprintf("Could not create status page component, unexpected error: %s. HTTP Response: %v", err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 201 {
		response.Diagnostics.AddError(
			"Error creating status page component",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	var state statusPageComponentModel
	state.PageID = plan.PageID
	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageComponentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state statusPageComponentModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	componentID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page component ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := r.Api.GetComponent(r.Auth, pageID, componentID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError(
			"Error reading status page component",
			"Could not read status page component ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	r.updateStateFromResponse(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageComponentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan statusPageComponentModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(plan.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	componentID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page component ID: "+err.Error(),
		)
		return
	}

	name := plan.Name.ValueString()
	position := plan.Position.ValueInt64()
	attributes := &datadogV2.PatchComponentRequestDataAttributes{
		Name:     &name,
		Position: &position,
	}

	body := datadogV2.PatchComponentRequest{
		Data: &datadogV2.PatchComponentRequestData{
			Id:         componentID,
			Type:       datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS,
			Attributes: *attributes,
		},
	}

	resp, httpResp, err := r.Api.UpdateComponent(r.Auth, pageID, componentID, body)
	if err != nil {
		response.Diagnostics.AddError(
			"Error updating status page component",
			fmt.Sprintf("Could not update status page component ID %s, unexpected error: %s. HTTP Response: %v", plan.ID.ValueString(), err.Error(), httpResp),
		)
		return
	}
	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error updating status page component",
			fmt.Sprintf("Received HTTP status %d. Response body: %v", httpResp.StatusCode, httpResp),
		)
		return
	}

	r.updateStateFromResponse(&plan, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *statusPageComponentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state statusPageComponentModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	pageID, err := uuid.Parse(state.PageID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing page ID",
			"Could not parse status page ID: "+err.Error(),
		)
		return
	}

	componentID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse status page component ID: "+err.Error(),
		)
		return
	}

	httpResp, err := r.Api.DeleteComponent(r.Auth, pageID, componentID)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.AddError(
			"Error deleting status page component",
			"Could not delete status page component ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func (r *statusPageComponentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("unexpected import format", "expected '<page_id>:<component_id>'")
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("page_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
}

func (r *statusPageComponentResource) updateStateFromResponse(state *statusPageComponentModel, resp *datadogV2.StatusPagesComponent) {
	data := resp.GetData()

	if id, ok := data.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(id.String())
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if statusPage, ok := relationships.GetStatusPageOk(); ok && statusPage != nil {
			pageData := statusPage.GetData()
			state.PageID = types.StringValue(pageData.GetId().String())
		}
	}

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		if name, ok := attributes.GetNameOk(); ok && name != nil {
			state.Name = types.StringValue(*name)
		}
		state.Type = types.StringValue(string(attributes.GetType()))
		if position, ok := attributes.GetPositionOk(); ok && position != nil {
			state.Position = types.Int64Value(*position)
		}
		if status, ok := attributes.GetStatusOk(); ok && status != nil {
			state.Status = types.StringValue(string(*status))
		}
		if createdAt, ok := attributes.GetCreatedAtOk(); ok && createdAt != nil {
			state.CreatedAt = types.StringValue(createdAt.Format("2006-01-02T15:04:05Z"))
		}
		if modifiedAt, ok := attributes.GetModifiedAtOk(); ok && modifiedAt != nil {
			state.ModifiedAt = types.StringValue(modifiedAt.Format("2006-01-02T15:04:05Z"))
		}
		if components, ok := attributes.GetComponentsOk(); ok && components != nil && len(*components) > 0 {
			state.Components = make([]statusPageSubComponentModel, len(*components))
			for i, sub := range *components {
				subModel := statusPageSubComponentModel{}
				if sub.Id != nil {
					subModel.ID = types.StringValue(sub.Id.String())
				}
				if sub.Name != nil {
					subModel.Name = types.StringValue(*sub.Name)
				}
				if sub.Type != nil {
					subModel.Type = types.StringValue(string(*sub.Type))
				}
				if sub.Position != nil {
					subModel.Position = types.Int64Value(*sub.Position)
				}
				if sub.Status != nil {
					subModel.Status = types.StringValue(string(*sub.Status))
				}
				state.Components[i] = subModel
			}
		} else {
			state.Components = nil
		}
	}
}
