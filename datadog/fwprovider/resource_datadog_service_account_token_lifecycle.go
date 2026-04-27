package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.Resource               = &serviceAccountTokenLifecycleResource{}
	_ resource.ResourceWithConfigure  = &serviceAccountTokenLifecycleResource{}
	_ resource.ResourceWithModifyPlan = &serviceAccountTokenLifecycleResource{}
)

func NewServiceAccountTokenLifecycleResource() resource.Resource {
	return &serviceAccountTokenLifecycleResource{}
}

type serviceAccountTokenLifecycleResource struct {
	Api  *datadogV2.ServiceAccountsApi
	Auth context.Context
}

type serviceAccountTokenLifecycleModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	KeyID            types.String `tfsdk:"key_id"`
	KeyIDVersion     types.String `tfsdk:"key_id_version"`
	RetainCount      types.Int64  `tfsdk:"retain_count"`
	ActiveKeyID      types.String `tfsdk:"active_key_id"`
	PreviousKeys     types.List   `tfsdk:"previous_keys"`
}

func (r *serviceAccountTokenLifecycleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "service_account_token_lifecycle"
}

func (r *serviceAccountTokenLifecycleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Tracks the lifecycle of a Datadog Service Account Access Token created by an `ephemeral.datadog_service_account_token` resource. Records the active token UUID and previous UUIDs (for graceful cutover during rotation), and revokes them when pruned by `retain_count` or when the resource is destroyed.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"service_account_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the service account that owns the tracked tokens.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_id": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "UUID of the SAT to track. Typically wired from `ephemeral.datadog_service_account_token.<name>.id`. Write-only because the source is an ephemeral attribute; pair with `key_id_version` so the framework can detect changes.",
			},
			"key_id_version": schema.StringAttribute{
				Required:    true,
				Description: "Non-ephemeral version trigger paired with `key_id`. Change this whenever you want the lifecycle to re-evaluate `key_id` (in particular: on rotation). Typical wiring: a string that incorporates `rotation_trigger` (e.g. the constructed SAT name or the rotation_trigger value itself).",
			},
			"retain_count": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Description: "Number of previous SATs to keep alive (for graceful cutover) before revoking. Defaults to 1 (immediate revocation on rotation).",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"active_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the currently active SAT.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"previous_keys": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "UUIDs of previously active SATs, most-recent first, kept alive per `retain_count`.",
			},
		},
	}
}

func (r *serviceAccountTokenLifecycleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data Type", fmt.Sprintf("Expected *FrameworkProvider, got %T", req.ProviderData))
		return
	}
	r.Api = providerData.DatadogApiInstances.GetServiceAccountsApiV2()
	r.Auth = providerData.Auth
}

// ModifyPlan marks active_key_id and previous_keys as unknown when key_id_version
// changes, so the planned values reflect "will be a new UUID after apply" rather than
// the stale state values. Without this, Update would mutate fields the plan said
// wouldn't change, and Terraform raises "Provider produced inconsistent result".
func (r *serviceAccountTokenLifecycleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Skip on create (no prior state) and destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state, plan serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.KeyIDVersion.Equal(plan.KeyIDVersion) {
		// Version changed → key_id will be re-read in Update and active/previous will shift.
		plan.ActiveKeyID = types.StringUnknown()
		plan.PreviousKeys = types.ListUnknown(types.StringType)
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

func (r *serviceAccountTokenLifecycleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cfg serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.ServiceAccountID
	plan.ActiveKeyID = cfg.KeyID
	plan.PreviousKeys = types.ListValueMust(types.StringType, []attr.Value{})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceAccountTokenLifecycleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountID := state.ServiceAccountID.ValueString()

	// Verify active key still exists.
	if !state.ActiveKeyID.IsNull() && state.ActiveKeyID.ValueString() != "" {
		_, httpResp, err := r.Api.GetServiceAccountAccessToken(r.Auth, serviceAccountID, state.ActiveKeyID.ValueString())
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				state.ActiveKeyID = types.StringNull()
			} else {
				resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading active SAT"))
				return
			}
		}
	}

	// Filter previous_keys for ones that still exist.
	var prev []string
	resp.Diagnostics.Append(state.PreviousKeys.ElementsAs(ctx, &prev, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	live := make([]string, 0, len(prev))
	for _, id := range prev {
		_, httpResp, err := r.Api.GetServiceAccountAccessToken(r.Auth, serviceAccountID, id)
		if err != nil {
			if httpResp != nil && httpResp.StatusCode == 404 {
				continue
			}
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading previous SAT"))
			return
		}
		live = append(live, id)
	}
	listVal, diags := types.ListValueFrom(ctx, types.StringType, live)
	resp.Diagnostics.Append(diags...)
	state.PreviousKeys = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serviceAccountTokenLifecycleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cfg serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newKeyID := cfg.KeyID.ValueString()
	oldActive := state.ActiveKeyID.ValueString()

	if newKeyID != oldActive && oldActive != "" {
		// Rotation or replacement: prepend old active to previous_keys.
		var prev []string
		resp.Diagnostics.Append(state.PreviousKeys.ElementsAs(ctx, &prev, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		newPrev := append([]string{oldActive}, prev...)

		// Prune to retain_count.
		retain := int(plan.RetainCount.ValueInt64())
		if retain < 1 {
			retain = 1
		}
		if len(newPrev) > retain {
			toRevoke := newPrev[retain:]
			newPrev = newPrev[:retain]
			for _, id := range toRevoke {
				_, err := r.Api.RevokeServiceAccountAccessToken(r.Auth, state.ServiceAccountID.ValueString(), id)
				if err != nil {
					resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error revoking pruned SAT %q", id)))
					return
				}
			}
		}

		listVal, diags := types.ListValueFrom(ctx, types.StringType, newPrev)
		resp.Diagnostics.Append(diags...)
		plan.PreviousKeys = listVal
	} else {
		plan.PreviousKeys = state.PreviousKeys
	}

	plan.ActiveKeyID = types.StringValue(newKeyID)
	plan.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serviceAccountTokenLifecycleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceAccountTokenLifecycleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountID := state.ServiceAccountID.ValueString()

	// Revoke active.
	if !state.ActiveKeyID.IsNull() && state.ActiveKeyID.ValueString() != "" {
		_, err := r.Api.RevokeServiceAccountAccessToken(r.Auth, serviceAccountID, state.ActiveKeyID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error revoking active SAT"))
			return
		}
	}

	// Revoke previous_keys.
	var prev []string
	resp.Diagnostics.Append(state.PreviousKeys.ElementsAs(ctx, &prev, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, id := range prev {
		_, err := r.Api.RevokeServiceAccountAccessToken(r.Auth, serviceAccountID, id)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error revoking previous SAT %q", id)))
			return
		}
	}
}
