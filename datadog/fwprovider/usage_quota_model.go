package fwprovider

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type usageQuotaModel struct {
	ID             types.String  `tfsdk:"id"`
	QuotaID        types.String  `tfsdk:"quota_id"`
	QuotaNamespace types.String  `tfsdk:"quota_namespace"`
	Scope          types.Map     `tfsdk:"scope"`
	UsageLimit     types.Float64 `tfsdk:"usage_limit"`
	Enforced       types.Bool    `tfsdk:"enforced"`
	OrgPublicID    types.String  `tfsdk:"org_public_id"`
}

func usageQuotaSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages one usage quota in the authenticated organization's account. Existing quotas must be imported, and each quota must be managed by only one Terraform resource. The usage quotas API must be enabled for the organization. Pending limits and descendant organizations are not managed. Disabling enforcement clears any pending limit; other updates preserve it.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true, Description: "Terraform identifier in <quota_namespace>/<quota_id> format.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"quota_id":        schema.StringAttribute{Computed: true, Description: "Opaque API quota identifier. IDs may contain encoded scope values, including user handles; they are not encrypted.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"quota_namespace": schema.StringAttribute{Required: true, Description: "Product-specific namespace, such as ai_credits. Changing this value replaces the quota.", Validators: []validator.String{stringvalidator.RegexMatches(regexp.MustCompile(`^[^/\s]+$`), "must be nonempty and contain no whitespace or slash")}, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"scope":           schema.MapAttribute{Optional: true, ElementType: types.StringType, Description: "Namespace-specific scope. Omit for an organization-wide quota. The API currently accepts one key, such as user_handle, role_uuid, or team_uuid; use * for a default scope. Changing scope replaces the quota. Prefer stable scope values to avoid replacements caused by unknown planned values.", Validators: []validator.Map{mapvalidator.SizeAtLeast(1)}, PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()}},
			"usage_limit":     schema.Float64Attribute{Required: true, Description: "Non-negative whole usage units. For ai_credits, units are credits per UTC calendar month. Imported legacy fractional values are preserved in state but require a whole value in configuration.", Validators: []validator.Float64{usageQuotaLimitValidator{}}},
			"enforced":        schema.BoolAttribute{Required: true, Description: "Whether the quota blocks usage. Setting false clears pending limits."},
			"org_public_id":   schema.StringAttribute{Computed: true, Description: "Public ID of the organization owning the quota.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}

type usageQuotaLimitValidator struct{}

func (usageQuotaLimitValidator) Description(context.Context) string {
	return "must be a finite, non-negative whole number below 2^63"
}
func (v usageQuotaLimitValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v usageQuotaLimitValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !validUsageQuotaLimit(req.ConfigValue.ValueFloat64()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid usage quota limit", v.Description(ctx))
	}
}
func validUsageQuotaLimit(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v < math.Exp2(63) && math.Trunc(v) == v
}

func parseUsageQuotaImportID(id string) (string, string, error) {
	namespace, quotaID, ok := strings.Cut(id, "/")
	if !ok || namespace == "" || quotaID == "" || strings.Contains(quotaID, "/") || strings.ContainsAny(id, " \t\r\n") {
		return "", "", fmt.Errorf("expected <quota_namespace>/<opaque-quota-id> with two nonempty components and no whitespace")
	}
	return namespace, quotaID, nil
}

// buildUsageQuotaCreate creates only the payload. Its caller MUST set
// create_only=true and include_descendants=false on the generated API operation.
func buildUsageQuotaCreate(ctx context.Context, plan usageQuotaModel) (datadogV2.UsageQuotasCreateRequest, diag.Diagnostics) {
	var diags diag.Diagnostics
	if plan.UsageLimit.IsNull() || plan.UsageLimit.IsUnknown() || !validUsageQuotaLimit(plan.UsageLimit.ValueFloat64()) || plan.Enforced.IsNull() || plan.Enforced.IsUnknown() {
		diags.AddError("Invalid usage quota plan", "Creating a quota requires a known whole usage_limit and enforced value.")
		return datadogV2.UsageQuotasCreateRequest{}, diags
	}
	attrs := datadogV2.NewUsageQuotaCreateAttributes()
	attrs.SetUsageLimit(int64(plan.UsageLimit.ValueFloat64()))
	attrs.SetEnforced(plan.Enforced.ValueBool())
	if !plan.Scope.IsNull() {
		var scope map[string]string
		diags.Append(plan.Scope.ElementsAs(ctx, &scope, false)...)
		if diags.HasError() {
			return datadogV2.UsageQuotasCreateRequest{}, diags
		}
		attrs.SetScope(scope)
	}
	return datadogV2.UsageQuotasCreateRequest{Data: []datadogV2.UsageQuotaCreateData{{Type: datadogV2.UsageQuotaType("quotas"), Attributes: *attrs}}}, diags
}

// An enforced-only PATCH must omit usage_limit: resending it unnecessarily
// invokes the API's current-usage validation. Never include pending_usage_limit.
func buildUsageQuotaUpdate(plan, state usageQuotaModel) (datadogV2.UsageQuotaUpdateRequest, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	attrs := datadogV2.NewUsageQuotaUpdateAttributes()
	changed := false
	if !plan.UsageLimit.Equal(state.UsageLimit) {
		if plan.UsageLimit.IsNull() || plan.UsageLimit.IsUnknown() || !validUsageQuotaLimit(plan.UsageLimit.ValueFloat64()) {
			diags.AddError("Invalid usage quota plan", "Updating a quota requires a known whole usage_limit.")
			return datadogV2.UsageQuotaUpdateRequest{}, false, diags
		}
		attrs.SetUsageLimit(int64(plan.UsageLimit.ValueFloat64()))
		changed = true
	}
	if !plan.Enforced.Equal(state.Enforced) {
		if plan.Enforced.IsNull() || plan.Enforced.IsUnknown() {
			diags.AddError("Invalid usage quota plan", "Updating a quota requires a known enforced value.")
			return datadogV2.UsageQuotaUpdateRequest{}, false, diags
		}
		attrs.SetEnforced(plan.Enforced.ValueBool())
		changed = true
	}
	return datadogV2.UsageQuotaUpdateRequest{Data: datadogV2.UsageQuotaUpdateData{Id: state.QuotaID.ValueString(), Type: datadogV2.UsageQuotaType("quotas"), Attributes: *attrs}}, changed, diags
}
