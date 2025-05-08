package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceEvaluationFilter struct {
	API  *datadogV2.SecurityMonitoringApi
	Auth context.Context
}

type ResourceEvaluationFilterModel struct {
	CloudProvider types.String `tfsdk:"cloud_provider"`
	ID            types.String `tfsdk:"id"`
	Tags          types.Set    `tfsdk:"tags"`
}

func NewResourceEvaluationFilter() resource.Resource {
	return &ResourceEvaluationFilter{}
}

func (r *ResourceEvaluationFilter) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.API = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.Auth = providerData.Auth
}

func (r *ResourceEvaluationFilter) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "resource_evaluation_filter"
}

var tagFormatValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`^[^:]+:[^:]+$`),
	"each tag must be in the format 'key:value' (colon-separated)",
)

func toSliceString(set types.Set) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	result := make([]string, 0)

	if set.IsNull() || set.IsUnknown() {
		return result, nil
	}

	for _, elem := range set.Elements() {
		strVal, ok := elem.(types.String)
		if !ok {
			diags.AddError("Invalid element type", "Expected string in set but found a different type")
			continue
		}
		result = append(result, strVal.ValueString())
	}

	return result, diags
}

func (r *ResourceEvaluationFilter) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage a single resource evaluation filter.",
		Attributes: map[string]schema.Attribute{
			"cloud_provider": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Required: true,
			},
			"tags": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(tagFormatValidator),
				},
			},
		},
	}
}

func (r *ResourceEvaluationFilter) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func convertStringSliceToAttrValues(s []string) []attr.Value {
	out := make([]attr.Value, len(s))
	for i, v := range s {
		out[i] = types.StringValue(v)
	}
	return out
}

func (r *ResourceEvaluationFilter) UpdateState(_ context.Context, state *ResourceEvaluationFilterModel, attributes *datadogV2.ResourceFilterAttributes) {
	// since we are handling a response after an update/read request, the cloud provider map will have at most one key
	// and the map of each cloud provider will also have at most one key
	for p, accounts := range attributes.CloudProvider {
		for id, tagList := range accounts {
			tags := types.SetValueMust(types.StringType, convertStringSliceToAttrValues(tagList))
			state.CloudProvider = types.StringValue(p)
			state.ID = types.StringValue(id)
			state.Tags = tags
			break
		}
		break
	}
}

func (r *ResourceEvaluationFilter) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state ResourceEvaluationFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if state.CloudProvider.IsNull() || state.CloudProvider.IsUnknown() {
		response.Diagnostics.AddError("Missing cloud_provider", "cloud_provider is required for lookup")
		return
	}
	provider, err := datadogV2.NewResourceFilterProviderEnumFromValue(state.CloudProvider.ValueString())

	params := datadogV2.GetResourceEvaluationFiltersOptionalParameters{
		CloudProvider: provider,
		AccountId:     state.ID.ValueStringPointer(),
	}

	resp, _, err := r.API.GetResourceEvaluationFilters(r.Auth, params)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving ResourceEvaluationFilter"))
		return
	}

	attributes := resp.Data.GetAttributes()
	r.UpdateState(ctx, &state, &attributes)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ResourceEvaluationFilter) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
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
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating ResourceEvaluationFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	attributes := resp.Data.GetAttributes()
	r.UpdateState(ctx, &state, &attributes)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *ResourceEvaluationFilter) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state ResourceEvaluationFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// empty tags
	state.Tags = types.SetValueMust(types.StringType, []attr.Value{})

	// create body as normal with empty tags
	body, diags := r.buildUpdateResourceEvaluationFilterRequest(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.API.UpdateResourceEvaluationFilters(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting ResourceEvaluationFilter"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	attributes := resp.Data.GetAttributes()
	r.UpdateState(ctx, &state, &attributes)
}

func (r *ResourceEvaluationFilter) buildUpdateResourceEvaluationFilterRequest(ctx context.Context, state *ResourceEvaluationFilterModel) (*datadogV2.UpdateResourceEvaluationFiltersRequest, diag.Diagnostics) {
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

	bytes, _ := json.MarshalIndent(attributes, "", "  ")
	fmt.Println(string(bytes))

	req := datadogV2.NewUpdateResourceEvaluationFiltersRequestWithDefaults()
	req.SetData(*data)

	return req, diags
}
