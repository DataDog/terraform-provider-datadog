package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &integrationAWSTagFilterResource{}
	_ resource.ResourceWithImportState = &integrationAWSTagFilterResource{}
)

type integrationAWSTagFilterResource struct {
	Api  *datadogV1.AWSIntegrationApi
	Auth context.Context
}

type integrationAWSTagFilterModel struct {
	ID           types.String `tfsdk:"id"`
	AccountID    types.String `tfsdk:"account_id"`
	Namespace    types.String `tfsdk:"namespace"`
	TagFilterStr types.String `tfsdk:"tag_filter_str"`
}

func NewIntegrationAWSTagFilterResource() resource.Resource {
	return &integrationAWSTagFilterResource{}
}

func (r *integrationAWSTagFilterResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAWSIntegrationApiV1()
	r.Auth = providerData.Auth
}

func (r *integrationAWSTagFilterResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_aws_tag_filter"
}

func (r *integrationAWSTagFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog AWS tag filter resource. This can be used to create and manage Datadog AWS tag filters.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				Required:    true,
				Description: "Your AWS Account ID without dashes.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				Required:    true,
				Description: "The namespace associated with the tag filter entry.",
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV1.NewAWSNamespaceFromValue),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tag_filter_str": schema.StringAttribute{
				Required:    true,
				Description: "The tag filter string.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationAWSTagFilterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationAWSTagFilterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationAWSTagFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagFilter, diags := r.getAWSTagFilter(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if tagFilter == nil {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(ctx, &state, tagFilter)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSTagFilterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationAWSTagFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	req := r.buildDatadogIntegrationAWSTagFilter(ctx, &state)
	_, _, err := r.Api.CreateAWSTagFilter(r.Auth, *req)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating aws tag filter"))
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s:%s", req.GetAccountId(), req.GetNamespace()))

	tagFilter, diags := r.getAWSTagFilter(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if tagFilter == nil {
		response.Diagnostics.AddError("error retrieving AWS tag filter", "")
		return
	}

	r.updateState(ctx, &state, tagFilter)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSTagFilterResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationAWSTagFilterModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	req := r.buildDatadogIntegrationAWSTagFilter(ctx, &state)
	_, _, err := r.Api.CreateAWSTagFilter(r.Auth, *req)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating aws tag filter"))
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%s:%s", req.GetAccountId(), req.GetNamespace()))

	tagFilter, diags := r.getAWSTagFilter(ctx, &state)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}
	if tagFilter == nil {
		response.Diagnostics.AddError("error retrieving AWS tag filter", "")
		return
	}

	r.updateState(ctx, &state, tagFilter)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationAWSTagFilterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationAWSTagFilterModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	IntegrationAWSMutex.Lock()
	defer IntegrationAWSMutex.Unlock()

	accountID, tfNamespace, err := utils.AccountAndNamespaceFromID(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error extracting account_id and namespace from id", err.Error())
		return
	}

	namespace := datadogV1.AWSNamespace(tfNamespace)
	deleteRequest := datadogV1.AWSTagFilterDeleteRequest{
		AccountId: &accountID,
		Namespace: &namespace,
	}

	_, _, err = r.Api.DeleteAWSTagFilter(r.Auth, deleteRequest)
	if err != nil {
		response.Diagnostics.AddError("error deleting aws tag filter", err.Error())
	}
}

func (r *integrationAWSTagFilterResource) getAWSTagFilter(ctx context.Context, state *integrationAWSTagFilterModel) (*datadogV1.AWSTagFilter, diag.Diagnostics) {
	var diags diag.Diagnostics

	accountID, tfNamespace, err := utils.AccountAndNamespaceFromID(state.ID.ValueString())
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, ""))
		return nil, diags
	}
	namespace := datadogV1.AWSNamespace(tfNamespace)

	tagFilters, _, err := r.Api.ListAWSTagFilters(r.Auth, accountID)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error listing aws tag filter."))
		return nil, diags
	}

	var filter *datadogV1.AWSTagFilter
	for _, tagFilter := range tagFilters.GetFilters() {
		if tagFilter.GetNamespace() == namespace {
			filter = &tagFilter
			break
		}
	}

	return filter, diags
}

func (r *integrationAWSTagFilterResource) updateState(ctx context.Context, state *integrationAWSTagFilterModel, tagFilter *datadogV1.AWSTagFilter) {
	state.Namespace = types.StringValue(string(tagFilter.GetNamespace()))
	state.TagFilterStr = types.StringValue(tagFilter.GetTagFilterStr())
}

func (r *integrationAWSTagFilterResource) buildDatadogIntegrationAWSTagFilter(ctx context.Context, state *integrationAWSTagFilterModel) *datadogV1.AWSTagFilterCreateRequest {
	filterRequest := datadogV1.NewAWSTagFilterCreateRequestWithDefaults()

	filterRequest.SetAccountId(state.AccountID.ValueString())

	namespace := datadogV1.AWSNamespace(state.Namespace.ValueString())
	filterRequest.SetNamespace(namespace)

	filterRequest.SetTagFilterStr(state.TagFilterStr.ValueString())

	return filterRequest
}
