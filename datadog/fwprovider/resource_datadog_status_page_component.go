package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.ResourceWithConfigure      = &statusPageComponentResource{}
	_ resource.ResourceWithImportState    = &statusPageComponentResource{}
	_ resource.ResourceWithValidateConfig = &statusPageComponentResource{}
)

func NewStatusPageComponentResource() resource.Resource {
	return &statusPageComponentResource{}
}

// statusPageChildModel is a group's child component. Child IDs are intentionally not
// exposed here (computed IDs on growable nested blocks break plan consistency); use
// the datadog_status_page_components data source to discover child IDs.
type statusPageChildModel struct {
	Name     types.String `tfsdk:"name"`
	Position types.Int64  `tfsdk:"position"`
}

type statusPageComponentResourceModel struct {
	ID           types.String           `tfsdk:"id"`
	StatusPageID types.String           `tfsdk:"status_page_id"`
	Name         types.String           `tfsdk:"name"`
	Type         types.String           `tfsdk:"type"`
	Position     types.Int64            `tfsdk:"position"`
	Components   []statusPageChildModel `tfsdk:"components"`
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
		Description: "Provides a Datadog Status Page Component resource. A `component` is a single status item; a `group` bundles child components declared via `components` blocks. Groups must declare at least one child (the API rejects empty groups).",
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
				Description:   "The ordering position of this item on the page (0-based). Positions must stay within bounds across all top-level items.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"components": schema.ListNestedBlock{
				Description: "Child components of a `group`. Required (and only valid) when `type` is `group`; child names must be unique within the group.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the child component.",
						},
						"position": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The ordering position of the child within the group (0-based).",
						},
					},
				},
			},
		},
	}
}

func (r *statusPageComponentResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var cfg statusPageComponentResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &cfg)...)
	if response.Diagnostics.HasError() {
		return
	}
	if cfg.Type.ValueString() == "group" && len(cfg.Components) == 0 {
		response.Diagnostics.AddAttributeError(frameworkPath.Root("components"), "group requires children",
			"A component of type `group` must declare at least one `components` block.")
	}
	if cfg.Type.ValueString() == "component" && len(cfg.Components) > 0 {
		response.Diagnostics.AddAttributeError(frameworkPath.Root("components"), "component cannot have children",
			"`components` blocks are only valid when `type` is `group`.")
	}
}

func (r *statusPageComponentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan statusPageComponentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, err := uuid.Parse(plan.StatusPageID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid status_page_id"))
		return
	}

	attrs := datadogV2.NewCreateComponentRequestDataAttributesWithDefaults()
	attrs.SetName(plan.Name.ValueString())
	attrs.SetType(datadogV2.CreateComponentRequestDataAttributesType(plan.Type.ValueString()))
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		attrs.SetPosition(plan.Position.ValueInt64())
	}
	if plan.Type.ValueString() == "group" {
		children := make([]datadogV2.CreateComponentRequestDataAttributesComponentsItems, 0, len(plan.Components))
		for i, c := range plan.Components {
			item := datadogV2.NewCreateComponentRequestDataAttributesComponentsItemsWithDefaults()
			item.SetName(c.Name.ValueString())
			item.SetType(datadogV2.STATUSPAGESCOMPONENTGROUPATTRIBUTESCOMPONENTSITEMSTYPE_COMPONENT)
			item.SetPosition(childPosition(c, int64(i)))
			children = append(children, *item)
		}
		attrs.SetComponents(children)
	}
	data := datadogV2.NewCreateComponentRequestData(*attrs, datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS)
	body := datadogV2.NewCreateComponentRequest()
	body.SetData(*data)

	resp, _, err := r.Api.CreateComponent(r.Auth, pid, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating status page component"))
		return
	}
	r.updateStateFromResponse(&plan, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
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
	r.updateStateFromResponse(&state, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *statusPageComponentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state statusPageComponentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	pid, cid, err := parseComponentIDs(plan.StatusPageID.ValueString(), state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "invalid component ids"))
		return
	}

	// Patch the top-level item's name/position.
	attrs := datadogV2.NewPatchComponentRequestDataAttributesWithDefaults()
	attrs.SetName(plan.Name.ValueString())
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		attrs.SetPosition(plan.Position.ValueInt64())
	}
	patchData := datadogV2.NewPatchComponentRequestData(*attrs, cid, datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS)
	patchBody := datadogV2.NewPatchComponentRequest()
	patchBody.SetData(*patchData)
	if _, _, err := r.Api.UpdateComponent(r.Auth, pid, cid, *patchBody); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating status page component"))
		return
	}

	// Reconcile group children (keyed by name) against the live component tree.
	if plan.Type.ValueString() == "group" {
		if diag := r.reconcileChildren(pid, cid, &plan); diag != nil {
			response.Diagnostics.Append(diag)
			return
		}
	}

	// Re-read to capture server-assigned positions.
	resp, _, err := r.Api.GetComponent(r.Auth, pid, cid)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error re-reading status page component"))
		return
	}
	r.updateStateFromResponse(&plan, resp.GetData())
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
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

// reconcileChildren brings a group's children into line with the plan, keyed by name.
// It fetches the live children (for their IDs), then renames/repositions existing ones,
// creates new ones (via the group relationship), and deletes removed ones.
func (r *statusPageComponentResource) reconcileChildren(pid, groupID uuid.UUID, plan *statusPageComponentResourceModel) frameworkDiag.Diagnostic {
	resp, _, err := r.Api.GetComponent(r.Auth, pid, groupID)
	if err != nil {
		return utils.FrameworkErrorDiag(err, "error listing current group children")
	}
	curData := resp.GetData()
	curAttrs := curData.GetAttributes()
	current := map[string]datadogV2.StatusPagesComponentDataAttributesComponentsItems{}
	for _, c := range curAttrs.GetComponents() {
		current[c.GetName()] = c
	}

	planNames := map[string]bool{}
	for i := range plan.Components {
		c := &plan.Components[i]
		name := c.Name.ValueString()
		planNames[name] = true
		pos := childPosition(*c, int64(i))
		if cur, ok := current[name]; ok {
			if cur.GetPosition() != pos {
				a := datadogV2.NewPatchComponentRequestDataAttributesWithDefaults()
				a.SetName(name)
				a.SetPosition(pos)
				d := datadogV2.NewPatchComponentRequestData(*a, cur.GetId(), datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS)
				b := datadogV2.NewPatchComponentRequest()
				b.SetData(*d)
				if _, _, err := r.Api.UpdateComponent(r.Auth, pid, cur.GetId(), *b); err != nil {
					return utils.FrameworkErrorDiag(err, "error updating child component")
				}
			}
			continue
		}
		a := datadogV2.NewCreateComponentRequestDataAttributesWithDefaults()
		a.SetName(name)
		a.SetType(datadogV2.CREATECOMPONENTREQUESTDATAATTRIBUTESTYPE_COMPONENT)
		a.SetPosition(pos)
		d := datadogV2.NewCreateComponentRequestData(*a, datadogV2.STATUSPAGESCOMPONENTGROUPTYPE_COMPONENTS)
		gd := datadogV2.NewCreateComponentRequestDataRelationshipsGroupDataWithDefaults()
		gd.SetId(groupID)
		g := datadogV2.NewCreateComponentRequestDataRelationshipsGroupWithDefaults()
		g.SetData(*gd)
		rel := datadogV2.NewCreateComponentRequestDataRelationshipsWithDefaults()
		rel.SetGroup(*g)
		d.SetRelationships(*rel)
		b := datadogV2.NewCreateComponentRequest()
		b.SetData(*d)
		if _, _, err := r.Api.CreateComponent(r.Auth, pid, *b); err != nil {
			return utils.FrameworkErrorDiag(err, "error creating child component")
		}
	}
	for name, cur := range current {
		if planNames[name] {
			continue
		}
		if _, err := r.Api.DeleteComponent(r.Auth, pid, cur.GetId()); err != nil {
			return utils.FrameworkErrorDiag(err, "error deleting child component")
		}
	}
	return nil
}

// updateStateFromResponse populates state from an API component. Children are ordered
// to match the incoming plan's order (by name) when present, else by API order.
func (r *statusPageComponentResource) updateStateFromResponse(state *statusPageComponentResourceModel, data datadogV2.StatusPagesComponentData) {
	attrs := data.GetAttributes()
	state.ID = types.StringValue(data.GetId().String())
	state.Name = types.StringValue(attrs.GetName())
	state.Type = types.StringValue(string(attrs.GetType()))
	state.Position = types.Int64Value(attrs.GetPosition())

	respByName := map[string]datadogV2.StatusPagesComponentDataAttributesComponentsItems{}
	respOrder := attrs.GetComponents()
	for _, c := range respOrder {
		respByName[c.GetName()] = c
	}

	if len(state.Components) > 0 {
		out := make([]statusPageChildModel, 0, len(state.Components))
		for _, c := range state.Components {
			if rc, ok := respByName[c.Name.ValueString()]; ok {
				out = append(out, statusPageChildModel{
					Name:     types.StringValue(rc.GetName()),
					Position: types.Int64Value(rc.GetPosition()),
				})
			}
		}
		state.Components = out
		return
	}

	if len(respOrder) == 0 {
		state.Components = nil
		return
	}
	out := make([]statusPageChildModel, 0, len(respOrder))
	for _, c := range respOrder {
		out = append(out, statusPageChildModel{
			Name:     types.StringValue(c.GetName()),
			Position: types.Int64Value(c.GetPosition()),
		})
	}
	state.Components = out
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

func childPosition(c statusPageChildModel, fallback int64) int64 {
	if !c.Position.IsNull() && !c.Position.IsUnknown() {
		return c.Position.ValueInt64()
	}
	return fallback
}
