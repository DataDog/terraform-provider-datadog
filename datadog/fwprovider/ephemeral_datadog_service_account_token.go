package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ ephemeral.EphemeralResource              = &serviceAccountTokenEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &serviceAccountTokenEphemeralResource{}
)

func NewServiceAccountTokenEphemeralResource() ephemeral.EphemeralResource {
	return &serviceAccountTokenEphemeralResource{}
}

type serviceAccountTokenEphemeralResource struct {
	Api  *datadogV2.ServiceAccountsApi
	Auth context.Context
}

type serviceAccountTokenEphemeralModel struct {
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	Prefix           types.String `tfsdk:"prefix"`
	Anchor           types.String `tfsdk:"anchor"`
	RotationTrigger  types.String `tfsdk:"rotation_trigger"`
	Scopes           types.Set    `tfsdk:"scopes"`
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Key              types.String `tfsdk:"key"`
}

func (r *serviceAccountTokenEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "service_account_token"
}

func (r *serviceAccountTokenEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Ephemeral resource that creates or finds a Datadog Service Account Access Token (SAT) without storing the secret in Terraform state. The provider constructs a unique server-side name from `prefix`, `anchor`, and `rotation_trigger`.",
		Attributes: map[string]schema.Attribute{
			"service_account_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the service account that owns this token.",
			},
			"prefix": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable prefix for the SAT. Forms the leading part of the constructed server-side name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"anchor": schema.StringAttribute{
				Required:    true,
				Description: "Stable cross-apply identity suffix for the SAT name. Typically wired from `terraform_data.<name>.id` so the value is randomly generated, persists across applies, and rotates only when the upstream `terraform_data.triggers_replace` fires. Can also be supplied as a static user-controlled string.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"rotation_trigger": schema.StringAttribute{
				Required:    true,
				Description: "Rotation control suffix for the SAT name. Bumping this value (e.g., \"1\" → \"2\", or wiring from `time_rotating.<name>.rotation_rfc3339`) produces a new SAT on the next apply. Should typically share its source with the value driving the upstream `terraform_data.triggers_replace`.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"scopes": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Authorization scopes for the SAT.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the SAT.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Server-side name of the SAT (`${prefix}-${anchor}-${rotation_trigger}`).",
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The token value. Populated only when a new SAT was created during this Open; null when an existing SAT was reused.",
			},
		},
	}
}

func (r *serviceAccountTokenEphemeralResource) Configure(_ context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *serviceAccountTokenEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var cfg serviceAccountTokenEphemeralModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceAccountID := cfg.ServiceAccountID.ValueString()
	actualName := constructSATName(cfg.Prefix.ValueString(), cfg.Anchor.ValueString(), cfg.RotationTrigger.ValueString())
	cfg.Name = types.StringValue(actualName)

	scopes := setToStrings(cfg.Scopes)

	// Look up by name. Filter is substring; we filter to exact match below.
	listResp, _, err := r.Api.ListServiceAccountAccessTokens(
		r.Auth,
		serviceAccountID,
		*datadogV2.NewListServiceAccountAccessTokensOptionalParameters().WithFilter(actualName),
	)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing service account tokens"))
		return
	}

	var matches []datadogV2.PersonalAccessToken
	for _, t := range listResp.GetData() {
		attrs := t.GetAttributes()
		if attrs.HasName() && attrs.GetName() == actualName {
			matches = append(matches, t)
		}
	}

	switch len(matches) {
	case 0:
		// Not found — create.
		full, _, err := r.Api.CreateServiceAccountAccessToken(
			r.Auth,
			serviceAccountID,
			buildSATCreateRequest(actualName, scopes),
		)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating service account token"))
			return
		}
		data := full.GetData()
		attrs := data.GetAttributes()
		cfg.ID = types.StringValue(data.GetId())
		cfg.Key = types.StringValue(attrs.GetKey())

	case 1:
		// Found — reuse. Reconcile scopes if drifted.
		existing := matches[0]
		existingAttrs := existing.GetAttributes()

		if !scopesEqual(existingAttrs.GetScopes(), scopes) {
			_, _, err := r.Api.UpdateServiceAccountAccessToken(
				r.Auth,
				serviceAccountID,
				existing.GetId(),
				buildPATUpdateRequest(existing.GetId(), scopes),
			)
			if err != nil {
				resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating service account token scopes"))
				return
			}
		}

		cfg.ID = types.StringValue(existing.GetId())
		cfg.Key = types.StringNull()

	default:
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.GetId())
		}
		resp.Diagnostics.AddError(
			"Multiple service account tokens matched",
			fmt.Sprintf(
				"Found %d tokens for service account %q with name %q. "+
					"Terraform cannot determine which token to manage. "+
					"Resolve by revoking the duplicate(s) or changing prefix/anchor/rotation_trigger.\n"+
					"Matched UUIDs: %v",
				len(matches), serviceAccountID, actualName, ids,
			),
		)
		return
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &cfg)...)
}

// constructSATName builds the server-side SAT name from the user-facing prefix, anchor, and
// rotation_trigger inputs. Kept as a separate function so other parts of the provider can
// compute the same value.
func constructSATName(prefix, anchor, rotationTrigger string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, anchor, rotationTrigger)
}

func buildSATCreateRequest(name string, scopes []string) datadogV2.ServiceAccountAccessTokenCreateRequest {
	attrs := datadogV2.NewServiceAccountAccessTokenCreateAttributes(name, scopes)
	data := datadogV2.NewServiceAccountAccessTokenCreateData(*attrs, datadogV2.PERSONALACCESSTOKENSTYPE_PERSONAL_ACCESS_TOKENS)
	return *datadogV2.NewServiceAccountAccessTokenCreateRequest(*data)
}

func buildPATUpdateRequest(id string, scopes []string) datadogV2.PersonalAccessTokenUpdateRequest {
	attrs := datadogV2.NewPersonalAccessTokenUpdateAttributes()
	attrs.SetScopes(scopes)
	data := datadogV2.NewPersonalAccessTokenUpdateData(*attrs, id, datadogV2.PERSONALACCESSTOKENSTYPE_PERSONAL_ACCESS_TOKENS)
	return *datadogV2.NewPersonalAccessTokenUpdateRequest(*data)
}

func setToStrings(s types.Set) []string {
	out := make([]string, 0, len(s.Elements()))
	for _, e := range s.Elements() {
		if v, ok := e.(types.String); ok {
			out = append(out, v.ValueString())
		}
	}
	return out
}

func scopesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]struct{}, len(a))
	for _, x := range a {
		seen[x] = struct{}{}
	}
	for _, x := range b {
		if _, ok := seen[x]; !ok {
			return false
		}
	}
	return true
}
