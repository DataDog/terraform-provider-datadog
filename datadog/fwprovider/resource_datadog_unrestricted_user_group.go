package fwprovider

import (
	"context"
	"regexp"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &unrestrictedUserGroupResource{}
	_ resource.ResourceWithImportState = &unrestrictedUserGroupResource{}
)

var unrestrictedUserGroupPrincipalRegexp = regexp.MustCompile(`^(role|team|user):[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type unrestrictedUserGroupResource struct {
	API  *datadogV2.RestrictionPoliciesApi
	Auth context.Context
}

type unrestrictedUserGroupModel struct {
	ID          types.String `tfsdk:"id"`
	PrincipalID types.String `tfsdk:"principal_id"`
	Products    types.Set    `tfsdk:"products"`
}

func NewUnrestrictedUserGroupResource() resource.Resource {
	return &unrestrictedUserGroupResource{}
}

func (r *unrestrictedUserGroupResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetRestrictionPoliciesApiV2()
	r.Auth = providerData.Auth
}

func (r *unrestrictedUserGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "unrestricted_user_group"
}

func (r *unrestrictedUserGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog unrestricted user group resource. This can be used to grant a role, team, or user unrestricted access to data and asset products. The resource manages the principal's complete list of products: products not listed are removed.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"principal_id": schema.StringAttribute{
				Required:    true,
				Description: "The principal identifier, formatted as `type:uuid`. Supported types are `role`, `team`, and `user`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(unrestrictedUserGroupPrincipalRegexp, "must be formatted as `role:<uuid>`, `team:<uuid>`, or `user:<uuid>`"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"products": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The data products or asset resource types for which the principal has unrestricted access, such as `logs`, `apm`, `dashboard`, or `monitor`.",
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 50),
					setvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
				},
			},
		},
	}
}

func (r *unrestrictedUserGroupResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *unrestrictedUserGroupResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state unrestrictedUserGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// There is no endpoint to get a single principal's group, so the group is found in the list.
	resp, _, err := r.API.ListUnrestrictedUserGroups(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing unrestricted user groups"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	group := findUnrestrictedUserGroup(resp.GetData(), state.ID.ValueString())
	if group == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(updateUnrestrictedUserGroupState(ctx, &state, group)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *unrestrictedUserGroupResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state unrestrictedUserGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.upsert(ctx, &state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *unrestrictedUserGroupResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state unrestrictedUserGroupModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.upsert(ctx, &state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *unrestrictedUserGroupResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state unrestrictedUserGroupModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// The API returns success when the group does not exist.
	httpResp, err := r.API.DeleteUnrestrictedUserGroup(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting unrestricted user group"))
	}
}

// upsert replaces the principal's complete list of products, which serves both create and update.
func (r *unrestrictedUserGroupResource) upsert(ctx context.Context, state *unrestrictedUserGroupModel, diags *diag.Diagnostics) {
	principalID := state.PrincipalID.ValueString()
	body, d := buildUnrestrictedUserGroupRequest(ctx, state)
	diags.Append(d...)
	if diags.HasError() {
		return
	}

	resp, _, err := r.API.CreateOrUpdateUnrestrictedUserGroup(r.Auth, principalID, *body)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error saving unrestricted user group"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("response contains unparsedObject", err.Error())
		return
	}

	group := resp.GetData()
	diags.Append(updateUnrestrictedUserGroupState(ctx, state, &group)...)
}

func findUnrestrictedUserGroup(groups []datadogV2.UnrestrictedUserGroup, principalID string) *datadogV2.UnrestrictedUserGroup {
	for i := range groups {
		if groups[i].GetId() == principalID {
			return &groups[i]
		}
	}
	return nil
}

func updateUnrestrictedUserGroupState(ctx context.Context, state *unrestrictedUserGroupModel, group *datadogV2.UnrestrictedUserGroup) diag.Diagnostics {
	state.ID = types.StringValue(group.GetId())
	state.PrincipalID = types.StringValue(group.GetId())

	attributes := group.GetAttributes()
	products := []string{}
	for _, product := range attributes.GetProducts() {
		products = append(products, product.GetName())
	}
	productSet, diags := types.SetValueFrom(ctx, types.StringType, products)
	state.Products = productSet
	return diags
}

func buildUnrestrictedUserGroupRequest(ctx context.Context, state *unrestrictedUserGroupModel) (*datadogV2.UnrestrictedUserGroupRequest, diag.Diagnostics) {
	var names []string
	diags := state.Products.ElementsAs(ctx, &names, false)
	if diags.HasError() {
		return nil, diags
	}
	// Sort so the request body is deterministic for cassette replay.
	sort.Strings(names)

	products := make([]datadogV2.UnrestrictedUserGroupProduct, 0, len(names))
	for _, name := range names {
		products = append(products, *datadogV2.NewUnrestrictedUserGroupProduct(name))
	}

	attributes := datadogV2.NewUnrestrictedUserGroupAttributes(products)
	data := datadogV2.NewUnrestrictedUserGroup(*attributes, state.PrincipalID.ValueString(), datadogV2.UNRESTRICTEDUSERGROUPTYPE_UNRESTRICTED_USER_GROUPS)
	return datadogV2.NewUnrestrictedUserGroupRequest(*data), diags
}
