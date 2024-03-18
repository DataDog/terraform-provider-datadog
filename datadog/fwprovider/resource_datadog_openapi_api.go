package fwprovider

import (
	"context"
	"io"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	frameworkSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/google/uuid"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &openapiApiResource{}
	_ resource.ResourceWithImportState = &openapiApiResource{}
)

type openapiApiResource struct {
	Api  *datadogV2.APIManagementApi
	Auth context.Context
}

type openapiApiModel struct {
	ID   types.String `tfsdk:"id"`
	Spec types.String `tfsdk:"spec"`
}

func NewOpenapiApiResource() resource.Resource {
	return &openapiApiResource{}
}

func (r *openapiApiResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetAPIManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *openapiApiResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "openapi_api"
}

func (r *openapiApiResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog OpenAPI resource. This can be used to synchronize Datadog's [API catalog](https://docs.datadoghq.com/api_catalog/) with an [OpenAPI](https://www.openapis.org/) specifications file.",
		Attributes: map[string]schema.Attribute{
			"id": frameworkSchema.StringAttribute{
				Description: "The API ID of this resource in Datadog.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"spec": schema.StringAttribute{
				Description: "The textual content of the OpenAPI specification. Use [`file()`](https://developer.hashicorp.com/terraform/language/functions/file) in order to reference another file in the repository (see exmaple).",
				Required:    true,
			},
		},
	}
}

func (r *openapiApiResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *openapiApiResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state openapiApiModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	uuid, _ := uuid.Parse(id)
	resp, httpResp, err := r.Api.GetOpenAPI(r.Auth, uuid)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OpenapiApi"))
		return
	}
	specData, err := io.ReadAll(resp)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading spec"))
		return
	}
	state.Spec = types.StringValue(string(specData))

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *openapiApiResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state openapiApiModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	specFile := state.Spec.ValueString()

	var bodyReader io.Reader
	bodyReader = strings.NewReader(specFile)
	params := datadogV2.NewCreateOpenAPIOptionalParameters().WithOpenapiSpecFile(bodyReader)
	resp, _, err := r.Api.CreateOpenAPI(r.Auth, *params)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OpenapiApi"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	respData := resp.GetData()
	state.ID = types.StringValue(respData.GetId().String())

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *openapiApiResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state openapiApiModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	uuid, _ := uuid.Parse(id)

	specFile := state.Spec.ValueString()

	var bodyReader io.Reader
	bodyReader = strings.NewReader(specFile)
	params := datadogV2.NewUpdateOpenAPIOptionalParameters().WithOpenapiSpecFile(bodyReader)

	resp, _, err := r.Api.UpdateOpenAPI(r.Auth, uuid, *params)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OpenapiApi"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	respData := resp.GetData()
	state.ID = types.StringValue(respData.GetId().String())

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *openapiApiResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state openapiApiModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	uuid, _ := uuid.Parse(id)

	httpResp, err := r.Api.DeleteOpenAPI(r.Auth, uuid)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting openapi_api"))
		return
	}
}
