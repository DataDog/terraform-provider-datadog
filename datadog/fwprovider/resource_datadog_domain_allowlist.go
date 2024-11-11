package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &domainAllowlistResource{}
	_ resource.ResourceWithImportState = &domainAllowlistResource{}
)

func NewDomainAllowlistResource() resource.Resource {
	return &domainAllowlistResource{}
}

type domainAllowlistResource struct {
	Api  *datadogV2.DomainAllowlistApi
	Auth context.Context
}

type domainAllowlistResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Domains []string     `tfsdk:"domains"`
}

func (r *domainAllowlistResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "domain_allowlist"
}

func (r *domainAllowlistResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides the Datadog Email Domain Allowlist resource. This can be used to manage the Datadog Email Domain Allowlist.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether the Email Domain Allowlist is enabled.",
				Required:    true,
			},
			"id": utils.ResourceIDAttribute(),
			"domains": schema.ListAttribute{
				Description: "The domains within the domain allowlist.",
				ElementType: types.StringType,
				Required:    true,
				Computed:    false,
			},
		},
	}
}

func (r *domainAllowlistResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetDomainAllowlistApiV2()
	r.Auth = providerData.Auth
}

func (r *domainAllowlistResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *domainAllowlistResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state domainAllowlistResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetDomainAllowlist(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, ""), "error getting team permission setting"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}
	domainAllowListData := resp.GetData()

	apiDomains, ok := domainAllowListData.Attributes.GetDomainsOk()
	priorEntries := state.Domains

	if !compareDomainEntries(priorEntries, *apiDomains) && ok && priorEntries != nil {
		state.Domains = *apiDomains
	}

	r.updateEnableState(ctx, &state, domainAllowListData.GetAttributes())
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)

}

func (r *domainAllowlistResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state domainAllowlistResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	domainAllowlistReq, _ := buildDomainAllowlistUpdateRequest(state)
	resp, httpResp, err := r.Api.PatchDomainAllowlist(r.Auth, *domainAllowlistReq)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, ""), "error creating domain allowlist"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

	domainAllowlistData := resp.GetData()

	state.ID = types.StringValue(domainAllowlistData.GetId())
	r.updateRequestState(ctx, &state, domainAllowlistData.Attributes)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *domainAllowlistResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state domainAllowlistResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	domainAllowlistReq, err := buildDomainAllowlistUpdateRequest(state)
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}
	resp, httpResp, err := r.Api.PatchDomainAllowlist(r.Auth, *domainAllowlistReq)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error updating domain allowlist"), ""))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

	domainAllowlistData := resp.GetData()
	r.updateRequestState(ctx, &state, domainAllowlistData.Attributes)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *domainAllowlistResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state domainAllowlistResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	domainAllowlistUpdateReq := datadogV2.NewDomainAllowlistRequestWithDefaults()
	domainAllowlistData := datadogV2.NewDomainAllowlist(datadogV2.DOMAINALLOWLISTTYPE_DOMAIN_ALLOWLIST)
	domainAllowlistAttributes := datadogV2.NewDomainAllowlistAttributesWithDefaults()
	domainAllowlistAttributes.SetEnabled(false)
	domainAllowlistAttributes.SetDomains([]string{})

	domainAllowlistData.SetAttributes(*domainAllowlistAttributes)
	domainAllowlistUpdateReq.SetData(*domainAllowlistData)

	resp, httpResp, err := r.Api.PatchDomainAllowlist(r.Auth, *domainAllowlistUpdateReq)

	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, ""), "error disabling and removing entries from domain allowlist"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

}

func (r *domainAllowlistResource) updateRequestState(ctx context.Context, state *domainAllowlistResourceModel, domainAllowlistAttrs *datadogV2.DomainAllowlistResponseDataAttributes) {
	if domainAllowlistAttrs != nil {
		if enabled, ok := domainAllowlistAttrs.GetEnabledOk(); ok && enabled != nil {
			state.Enabled = types.BoolValue(*enabled)
		}

		if domains, ok := domainAllowlistAttrs.GetDomainsOk(); ok && len(*domains) > 0 {
			state.Domains = domainAllowlistAttrs.GetDomains()
		}
	}
}

func (r *domainAllowlistResource) updateEnableState(ctx context.Context, state *domainAllowlistResourceModel, domainAllowlistAttrs datadogV2.DomainAllowlistResponseDataAttributes) {
	if enabled, ok := domainAllowlistAttrs.GetEnabledOk(); ok && enabled != nil {
		state.Enabled = types.BoolValue(*enabled)
	}
}

func buildDomainAllowlistUpdateRequest(state domainAllowlistResourceModel) (*datadogV2.DomainAllowlistRequest, error) {
	domainAllowlistRequest := datadogV2.NewDomainAllowlistRequestWithDefaults()
	domainAllowlistData := datadogV2.NewDomainAllowlist(datadogV2.DOMAINALLOWLISTTYPE_DOMAIN_ALLOWLIST)
	domainAllowlistAttributes := datadogV2.NewDomainAllowlistAttributesWithDefaults()

	enabled := state.Enabled
	domainAllowlistAttributes.SetEnabled(enabled.ValueBool())
	domains := state.Domains
	if domains != nil {
		domainAllowlistDomains := make([]string, len(domains))
		copy(domainAllowlistDomains, domains)
		domainAllowlistAttributes.SetDomains(domainAllowlistDomains)
	} else {
		domainAllowlistAttributes.SetDomains([]string{})
	}

	domainAllowlistData.SetAttributes(*domainAllowlistAttributes)
	domainAllowlistRequest.SetData(*domainAllowlistData)
	return domainAllowlistRequest, nil
}

func compareDomainEntries(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
