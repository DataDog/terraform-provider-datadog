package fwprovider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

type ComplianceResourceEvaluationFilter struct {
	API  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type ResourceEvaluationFilterModel struct {
	CloudProvider types.String `tfsdk:"cloud_provider"`
	ID            types.String `tfsdk:"id"`
	Tags          types.List   `tfsdk:"tags"`
}

func NewResourceEvaluationFilter() resource.Resource {
	return &ComplianceResourceEvaluationFilter{}
}

var (
	_ resource.ResourceWithConfigure   = &ComplianceResourceEvaluationFilter{}
	_ resource.ResourceWithImportState = &ComplianceResourceEvaluationFilter{}
)

func (r *ComplianceResourceEvaluationFilter) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *ComplianceResourceEvaluationFilter) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "compliance_resource_evaluation_filter"
}

var tagFormatValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`^[^:]+:[^:]+$`),
	"each tag must be in the format 'key:value' (colon-separated)",
)

func toSliceString(list types.List) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make([]string, 0)

	if list.IsNull() || list.IsUnknown() {
		return result, nil
	}

	for _, elem := range list.Elements() {
		strVal, ok := elem.(types.String)
		if !ok {
			diags.AddError("Invalid element type creating tags list", fmt.Sprintf("Expected string in list but found %T", elem))
			continue
		}
		result = append(result, strVal.ValueString())
	}

	return result, diags
}

func (r *ComplianceResourceEvaluationFilter) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog ResourceEvaluationFilter resource. This can be used to create and manage a resource evaluation filter.",
		Attributes: map[string]schema.Attribute{
			"cloud_provider": schema.StringAttribute{
				Required:    true,
				Description: "The cloud provider of the filter's targeted resource. Only `aws`, `gcp` or `azure` are considered valid cloud providers.",
			},
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the of the filter's targeted resource. Different cloud providers target different resource IDs:\n  - `aws`: account id \n  - `gcp`: project id\n  - `azure`: subscription id",
			},
			"tags": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(tagFormatValidator),
				},
				Description: "List of tags to filter misconfiguration detections. Each entry should follow the format: \"key\":\"value\".",
			},
		},
	}
}

func (r *ComplianceResourceEvaluationFilter) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state ResourceEvaluationFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildUpdateResourceEvaluationFilterRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateResourceEvaluationFilters(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating resource evaluation filter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	attributes := resp.Data.GetAttributes()
	r.UpdateState(ctx, &state, &attributes)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func convertStringSliceToAttrValues(s []string) []attr.Value {
	out := make([]attr.Value, len(s))
	for i, v := range s {
		out[i] = types.StringValue(v)
	}
	return out
}

func (r *ComplianceResourceEvaluationFilter) UpdateState(_ context.Context, state *ResourceEvaluationFilterModel, attributes *datadogV2.ResourceFilterAttributes) {
	for p, accounts := range attributes.CloudProvider {
		for id, tagList := range accounts {
			tags := types.ListValueMust(types.StringType, convertStringSliceToAttrValues(tagList))
			state.CloudProvider = types.StringValue(p)
			state.ID = types.StringValue(id)
			state.Tags = tags
			break
		}
		break
	}
}

func (r *ComplianceResourceEvaluationFilter) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ResourceEvaluationFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if state.CloudProvider.IsNull() || state.CloudProvider.IsUnknown() {
		response.Diagnostics.AddError("Missing cloud_provider", "cloud_provider is required for lookup")
		return
	}

	provider := state.CloudProvider.ValueString()
	skipCache := true

	params := datadogV2.GetResourceEvaluationFiltersOptionalParameters{
		CloudProvider: &provider,
		AccountId:     state.ID.ValueStringPointer(),
		SkipCache:     &skipCache,
	}
	resp, _, err := r.API.GetResourceEvaluationFilters(r.Auth, params)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ComplianceResourceEvaluationFilter"))
		return
	}

	attributes := resp.Data.GetAttributes()
	r.UpdateState(ctx, &state, &attributes)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ComplianceResourceEvaluationFilter) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state ResourceEvaluationFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildUpdateResourceEvaluationFilterRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateResourceEvaluationFilters(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ComplianceResourceEvaluationFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	attributes := resp.Data.GetAttributes()
	r.UpdateState(ctx, &state, &attributes)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ComplianceResourceEvaluationFilter) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ResourceEvaluationFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	state.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	body, diags := r.buildUpdateResourceEvaluationFilterRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateResourceEvaluationFilters(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting ComplianceResourceEvaluationFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
}

func (r *ComplianceResourceEvaluationFilter) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import format",
			`Expected format: "cloud_provider:id" (e.g., "aws:123456789")`,
		)
		return
	}

	cloudProvider := parts[0]
	id := parts[1]

	resp.State.SetAttribute(ctx, path.Root("cloud_provider"), cloudProvider)
	resp.State.SetAttribute(ctx, path.Root("id"), id)
}

func (r *ComplianceResourceEvaluationFilter) buildUpdateResourceEvaluationFilterRequest(ctx context.Context, state *ResourceEvaluationFilterModel) (*datadogV2.UpdateResourceEvaluationFiltersRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	data := datadogV2.NewUpdateResourceEvaluationFiltersRequestDataWithDefaults()

	tagsList, tagDiags := toSliceString(state.Tags)
	diags.Append(tagDiags...)
	if tagDiags.HasError() {
		return nil, diags
	}

	if state.CloudProvider.IsNull() || state.CloudProvider.IsUnknown() {
		diags.AddError("Missing cloud_provider", "cloud_provider is required but was null or unknown")
		return nil, diags
	}
	if state.ID.IsNull() || state.ID.IsUnknown() {
		diags.AddError("Missing id", "id is required but was null or unknown")
		return nil, diags
	}

	attributes := datadogV2.ResourceFilterAttributes{
		CloudProvider: map[string]map[string][]string{
			state.CloudProvider.ValueString(): {
				state.ID.ValueString(): tagsList,
			},
		},
	}

	data.SetId(string(datadogV2.RESOURCEFILTERREQUESTTYPE_CSM_RESOURCE_FILTER))
	data.SetType(datadogV2.RESOURCEFILTERREQUESTTYPE_CSM_RESOURCE_FILTER)
	data.SetAttributes(attributes)

	req := datadogV2.NewUpdateResourceEvaluationFiltersRequestWithDefaults()
	req.SetData(*data)

	return req, diags
}
