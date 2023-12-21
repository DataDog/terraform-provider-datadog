
package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
    "github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
		Description: "Provides a Datadog OpenapiApi resource. This can be used to create and manage Datadog openapi_api.",
		Attributes: map[string]schema.Attribute{
            "id": utils.ResourceIDAttribute(),
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
    resp, httpResp, err := r.Api.GetOpenAPI(r.Auth, id,)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OpenapiApi"))
		return
	}
    if err := utils.CheckForUnparsed(resp); err != nil {
        response.Diagnostics.AddError("response contains unparsedObject", err.Error())
        return
    }

    r.updateState(ctx, &state, &resp)

    // Save data into Terraform state
    response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}



func (r *openapiApiResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
    var state openapiApiModel
    response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
    if response.Diagnostics.HasError() {
        return
    }
    
    specFile := state.SpecFile.ValueString()
    
    body, diags := r.buildOpenapiApiRequestBody(ctx, &state)
    response.Diagnostics.Append(diags...)
    if response.Diagnostics.HasError() {
        return
    }

	resp, _, err := r.Api.CreateOpenAPI(r.Auth, specFile, *body, )
	if err != nil {
	    response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OpenapiApi"))
        return
	}
    if err := utils.CheckForUnparsed(resp); err != nil {
        response.Diagnostics.AddError("response contains unparsedObject", err.Error())
        return
    }
	r.updateState(ctx, &state, &resp)

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
    
    specFile := state.SpecFile.ValueString()
    
    body, diags := r.buildOpenapiApiUpdateRequestBody(ctx, &state)
    response.Diagnostics.Append(diags...)
    if response.Diagnostics.HasError() {
        return
    }
	

	resp, _, err := r.Api.UpdateOpenAPI(r.Auth, id, specFile, *body, )
	if err != nil {
	    response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OpenapiApi"))
        return
	}
    if err := utils.CheckForUnparsed(resp); err != nil {
        response.Diagnostics.AddError("response contains unparsedObject", err.Error())
        return
    }
	r.updateState(ctx, &state, &resp)

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

    httpResp, err := r.Api.DeleteOpenAPI(r.Auth, id,)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting openapi_api"))
		return
	}
}


func (r *openapiApiResource) updateState(ctx context.Context, state *openapiApiModel, resp *datadogV2.OpenAPIfile) {
    state.ID = types.StringValue(resp.GetId())
    
    
    if specFile, ok := resp.GetSpecFileOk(); ok {
    state.SpecFile = types.StringValue(*specFile)
    }
}




func (r *openapiApiResource) buildOpenapiApiRequestBody(ctx context.Context, state *openapiApiModel) (*datadogV2.string, diag.Diagnostics) {
    diags := diag.Diagnostics{}
    

	return req, diags
}



func (r *openapiApiResource) buildOpenapiApiUpdateRequestBody(ctx context.Context, state *openapiApiModel) (*datadogV2.string, diag.Diagnostics) {
    diags := diag.Diagnostics{}
    

	return req, diags
}
