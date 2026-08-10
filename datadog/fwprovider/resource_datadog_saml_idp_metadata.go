package fwprovider

import (
	"context"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure = &samlIdpMetadataResource{}
)

type samlIdpMetadataResource struct {
	Api  *datadogV2.OrganizationsApi
	Auth context.Context
}

type samlIdpMetadataModel struct {
	ID          types.String `tfsdk:"id"`
	IdpMetadata types.String `tfsdk:"idp_metadata"`
	EntityId    types.String `tfsdk:"entity_id"`
	SsoUrl      types.String `tfsdk:"sso_url"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
}

func NewSamlIdpMetadataResource() resource.Resource {
	return &samlIdpMetadataResource{}
}

func (r *samlIdpMetadataResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetOrganizationsApiV2()
	r.Auth = providerData.Auth
}

func (r *samlIdpMetadataResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "saml_idp_metadata"
}

func (r *samlIdpMetadataResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog SAML IdP metadata resource. This can be used to upload or replace the identity provider (IdP) metadata used for the organization's SAML login configuration. An organization has at most one SAML configuration. Note: the uploaded metadata XML cannot be read back from the Datadog API, so changes made outside of Terraform to the metadata content are not detected. Destroying this resource removes it from Terraform state but does not remove the IdP metadata from Datadog.",
		Attributes: map[string]schema.Attribute{
			"idp_metadata": schema.StringAttribute{
				Required:    true,
				Description: "The content of the IdP metadata XML file, for example loaded with the `file()` function. A leading UTF-8 byte order mark (BOM) is stripped automatically before upload.",
			},
			"entity_id": schema.StringAttribute{
				Computed:    true,
				Description: "The IdP entity ID of the SAML configuration.",
			},
			"sso_url": schema.StringAttribute{
				Computed:    true,
				Description: "The single sign-on (SSO) URL of the SAML configuration.",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp (RFC3339) at which the IdP certificate of the SAML configuration expires.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *samlIdpMetadataResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state samlIdpMetadataModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetSAMLConfiguration(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting SAML configuration"))
		return
	}

	r.updateState(&state, resp.Data)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *samlIdpMetadataResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state samlIdpMetadataModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.uploadIdpMetadata(&state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *samlIdpMetadataResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state samlIdpMetadataModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.uploadIdpMetadata(&state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *samlIdpMetadataResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddWarning(
		"SAML IdP metadata cannot be deleted",
		"The Datadog API does not support deleting uploaded IdP metadata. The resource is only removed from Terraform state; the SAML configuration remains in Datadog.",
	)
}

// uploadIdpMetadata uploads the metadata and refreshes the state from the
// resulting SAML configuration.
func (r *samlIdpMetadataResource) uploadIdpMetadata(state *samlIdpMetadataModel, diags *diag.Diagnostics) {
	// Some IdPs (for example Azure AD) prepend a UTF-8 BOM which the Datadog API rejects.
	content := strings.TrimPrefix(state.IdpMetadata.ValueString(), "\uFEFF")

	_, err := r.Api.UploadIdPMetadata(r.Auth, *datadogV2.NewUploadIdPMetadataOptionalParameters().WithIdpFile(strings.NewReader(content)))
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error uploading IdP metadata"))
		return
	}

	resp, _, err := r.Api.ListSAMLConfigurations(r.Auth)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing SAML configurations after uploading IdP metadata"))
		return
	}
	if len(resp.Data) == 0 {
		diags.AddError("error retrieving SAML configuration after uploading IdP metadata", "the IdP metadata upload succeeded but no SAML configuration was returned by the Datadog API")
		return
	}

	// An organization has at most one SAML configuration.
	r.updateState(state, resp.Data[0])
}

func (r *samlIdpMetadataResource) updateState(state *samlIdpMetadataModel, resp datadogV2.SAMLConfiguration) {
	state.ID = types.StringValue(resp.GetId())

	state.EntityId = types.StringNull()
	state.SsoUrl = types.StringNull()
	state.ExpiresAt = types.StringNull()

	attributes := resp.GetAttributes()
	if entityId, ok := attributes.GetEntityIdOk(); ok && entityId != nil {
		state.EntityId = types.StringValue(*entityId)
	}
	if ssoUrl, ok := attributes.GetSsoUrlOk(); ok && ssoUrl != nil {
		state.SsoUrl = types.StringValue(*ssoUrl)
	}
	if expiresAt, ok := attributes.GetExpiresAtOk(); ok && expiresAt != nil {
		state.ExpiresAt = types.StringValue(expiresAt.Format(time.RFC3339))
	}
}
