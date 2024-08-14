package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/planmodifiers"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &ipAllowListResource{}
	_ resource.ResourceWithImportState = &ipAllowListResource{}
)

func NewIpAllowListResource() resource.Resource {
	return &ipAllowListResource{}
}

type ipAllowListResource struct {
	Api  *datadogV2.IPAllowlistApi
	Auth context.Context
}

type ipAllowListResourceModel struct {
	ID      types.String        `tfsdk:"id"`
	Enabled types.Bool          `tfsdk:"enabled"`
	Entry   []*ipAllowListEntry `tfsdk:"entry"`
}

type ipAllowListEntry struct {
	CidrBlock types.String `tfsdk:"cidr_block"`
	Note      types.String `tfsdk:"note"`
}

func (r *ipAllowListResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "ip_allowlist"
}

func (r *ipAllowListResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides the Datadog IP allowlist resource. This can be used to manage the Datadog IP allowlist",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Description: "Whether the IP Allowlist is enabled.",
				Required:    true,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"entry": schema.SetNestedBlock{
				Description: "Set of objects containing an IP address or range of IP addresses in the allowlist and an accompanying note.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cidr_block": schema.StringAttribute{
							Required:    true,
							Description: "IP address or range of addresses.",
							Validators:  []validator.String{validators.CidrIpValidator()},
							PlanModifiers: []planmodifier.String{
								planmodifiers.NormalizeIP(),
							},
						},
						"note": schema.StringAttribute{
							Optional:    true,
							Description: "Note accompanying IP address.",
						},
					},
				},
			},
		},
	}
}

func (r *ipAllowListResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetIPAllowlistApiV2()
	r.Auth = providerData.Auth
}

func (r *ipAllowListResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *ipAllowListResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ipAllowListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetIPAllowlist(r.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, ""), "error getting team permission setting"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}
	ipAllowListData := resp.GetData()

	apiEntries, ok := ipAllowListData.Attributes.GetEntriesOk()
	priorEntries := state.Entry

	if !compareEntries(priorEntries, *apiEntries) && ok && priorEntries != nil {
		r.updateIPAllowlistEntriesState(ctx, &state, apiEntries)
	}

	r.updateEnableState(ctx, &state, ipAllowListData.Attributes)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)

}

func (r *ipAllowListResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state ipAllowListResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	ipAllowlistReq, _ := buildIPAllowlistUpdateRequest(state)
	resp, httpResp, err := r.Api.UpdateIPAllowlist(r.Auth, *ipAllowlistReq)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, ""), "error updating IP allowlist"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

	ipAllowlistData := resp.GetData()

	state.ID = types.StringValue(ipAllowlistData.GetId())
	r.updateState(ctx, &state, ipAllowlistData.Attributes)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ipAllowListResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state ipAllowListResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	ipAllowlistReq, err := buildIPAllowlistUpdateRequest(state)
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}
	resp, httpResp, err := r.Api.UpdateIPAllowlist(r.Auth, *ipAllowlistReq)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, " error updating IP allowlist"), ""))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

	ipAllowlistData := resp.GetData()
	r.updateState(ctx, &state, ipAllowlistData.Attributes)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ipAllowListResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ipAllowListResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	ipAllowlistUpdateReq := datadogV2.NewIPAllowlistUpdateRequestWithDefaults()
	ipAllowlistData := datadogV2.NewIPAllowlistDataWithDefaults()
	ipAllowlistAttributes := datadogV2.NewIPAllowlistAttributesWithDefaults()
	ipAllowlistAttributes.SetEnabled(false)
	ipAllowlistAttributes.SetEntries([]datadogV2.IPAllowlistEntry{})

	ipAllowlistData.SetAttributes(*ipAllowlistAttributes)
	ipAllowlistUpdateReq.SetData(*ipAllowlistData)

	resp, httpResp, err := r.Api.UpdateIPAllowlist(r.Auth, *ipAllowlistUpdateReq)

	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, ""), "error disabling and removing entries from IP allowlist"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("", err.Error())
		return
	}

}

func (r *ipAllowListResource) updateState(ctx context.Context, state *ipAllowListResourceModel, ipAllowlistAttrs *datadogV2.IPAllowlistAttributes) {
	if ipAllowlistAttrs != nil {
		if enabled, ok := ipAllowlistAttrs.GetEnabledOk(); ok && enabled != nil {
			state.Enabled = types.BoolValue(*enabled)
		}

		if entries, ok := ipAllowlistAttrs.GetEntriesOk(); ok && len(*entries) > 0 {
			r.updateIPAllowlistEntriesState(ctx, state, entries)
		}
	}
}

func (r *ipAllowListResource) updateEnableState(ctx context.Context, state *ipAllowListResourceModel, ipAllowlistAttrs *datadogV2.IPAllowlistAttributes) {
	if ipAllowlistAttrs != nil {
		if enabled, ok := ipAllowlistAttrs.GetEnabledOk(); ok && enabled != nil {
			state.Enabled = types.BoolValue(*enabled)
		}
	}
}

func (r *ipAllowListResource) updateIPAllowlistEntriesState(ctx context.Context, state *ipAllowListResourceModel, ipAllowlistEntries *[]datadogV2.IPAllowlistEntry) {
	var entries []*ipAllowListEntry
	for _, ipAllowlistEntry := range *ipAllowlistEntries {
		ipAllowlistEntryData := ipAllowlistEntry.GetData()
		ipAllowlistEntryAttributes := ipAllowlistEntryData.GetAttributes()
		cidrBlock, okCidr := ipAllowlistEntryAttributes.GetCidrBlockOk()
		note, okNote := ipAllowlistEntryAttributes.GetNoteOk()
		if okCidr && okNote {
			entry := &ipAllowListEntry{
				CidrBlock: types.StringValue(*cidrBlock),
				Note:      types.StringValue(*note),
			}
			entries = append(entries, entry)
		}
	}
	state.Entry = entries
}

func buildIPAllowlistUpdateRequest(state ipAllowListResourceModel) (*datadogV2.IPAllowlistUpdateRequest, error) {
	ipAllowlistUpdateRequest := datadogV2.NewIPAllowlistUpdateRequestWithDefaults()
	ipAllowlistData := datadogV2.NewIPAllowlistDataWithDefaults()
	ipAllowlistAttributes := datadogV2.NewIPAllowlistAttributesWithDefaults()

	enabled := state.Enabled
	ipAllowlistAttributes.SetEnabled(enabled.ValueBool())
	entries := state.Entry
	if entries != nil {
		ipAllowlistEntries := make([]datadogV2.IPAllowlistEntry, len(entries))
		for i, entry := range entries {
			ipAllowlistEntry := datadogV2.NewIPAllowlistEntryWithDefaults()
			ipAllowlistEntryData := datadogV2.NewIPAllowlistEntryDataWithDefaults()
			ipAllowlistEntryAttributes := datadogV2.NewIPAllowlistEntryAttributesWithDefaults()
			ipAllowlistEntryAttributes.SetCidrBlock(entry.CidrBlock.ValueString())
			ipAllowlistEntryAttributes.SetNote(entry.Note.ValueString())
			ipAllowlistEntryData.SetAttributes(*ipAllowlistEntryAttributes)
			ipAllowlistEntry.SetData(*ipAllowlistEntryData)
			ipAllowlistEntries[i] = *ipAllowlistEntry
		}
		ipAllowlistAttributes.SetEntries(ipAllowlistEntries)
	} else {
		ipAllowlistAttributes.SetEntries([]datadogV2.IPAllowlistEntry{})
	}

	ipAllowlistData.SetAttributes(*ipAllowlistAttributes)
	ipAllowlistUpdateRequest.SetData(*ipAllowlistData)
	return ipAllowlistUpdateRequest, nil
}

func compareEntries(slice1 []*ipAllowListEntry, slice2 []datadogV2.IPAllowlistEntry) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i].CidrBlock.ValueString() != slice2[i].GetData().Attributes.GetCidrBlock() || slice1[i].Note.ValueString() != slice2[i].GetData().Attributes.GetNote() {
			return false
		}
	}
	return true
}
