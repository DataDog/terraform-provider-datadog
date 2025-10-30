package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &syntheticsPrivateLocationResource{}
	_ resource.ResourceWithImportState = &syntheticsPrivateLocationResource{}
)

type syntheticsPrivateLocationResource struct {
	Api  *datadogV1.SyntheticsApi
	Auth context.Context
}

type syntheticsPrivateLocationModel struct {
	Id                          types.String                             `tfsdk:"id"`
	RestrictionPolicyResourceId types.String                             `tfsdk:"restriction_policy_resource_id"`
	Config                      types.String                             `tfsdk:"config"`
	Description                 types.String                             `tfsdk:"description"`
	Metadata                    []syntheticsPrivateLocationMetadataModel `tfsdk:"metadata"`
	Name                        types.String                             `tfsdk:"name"`
	Tags                        types.List                               `tfsdk:"tags"`
	ApiKey                      types.String                             `tfsdk:"api_key"`
}

type syntheticsPrivateLocationMetadataModel struct {
	RestrictedRoles types.Set `tfsdk:"restricted_roles"`
}

func NewSyntheticsPrivateLocationResource() resource.Resource {
	return &syntheticsPrivateLocationResource{}
}

func (r *syntheticsPrivateLocationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetSyntheticsApiV1()
	r.Auth = providerData.Auth
}

func (r *syntheticsPrivateLocationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "synthetics_private_location"
}

func (r *syntheticsPrivateLocationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog synthetics private location resource. This can be used to create and manage Datadog synthetics private locations.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Synthetics private location name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Description of the private location.",
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A list of tags to associate with your synthetics private location.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"config": schema.StringAttribute{
				Description: "Configuration skeleton for the private location. See installation instructions of the private location on how to use this configuration.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_key": schema.StringAttribute{
				Description: "API key used to generate the private location configuration.",
				Optional:    true,
				Sensitive:   true,
			},
			"id": utils.ResourceIDAttribute(),
			"restriction_policy_resource_id": schema.StringAttribute{
				Description: "Resource ID to use when setting restrictions with a `datadog_restriction_policy` resource.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"metadata": schema.ListNestedBlock{
				Description: "The private location metadata",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"restricted_roles": schema.SetAttribute{
							Description:        "A set of role identifiers pulled from the Roles API to restrict read and write access. **Deprecated.** This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
							DeprecationMessage: "This field is no longer supported by the Datadog API. Please use `datadog_restriction_policy` instead.",
							Optional:           true,
							ElementType:        types.StringType,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *syntheticsPrivateLocationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *syntheticsPrivateLocationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	resp, httpResp, err := r.Api.GetPrivateLocation(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving SyntheticsPrivateLocation"))
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

func (r *syntheticsPrivateLocationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildSyntheticsPrivateLocationRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreatePrivateLocation(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating SyntheticsPrivateLocation"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, resp.PrivateLocation)

	// set the config that is only returned when creating the private location
	conf, err := json.Marshal(resp.GetConfig())
	if err != nil {
		response.Diagnostics.AddError(
			"Error marshaling config to JSON",
			fmt.Sprintf("Could not marshal config: %s", err.Error()),
		)
	}

	confStr := string(conf)
	apiKey := state.ApiKey.ValueString()
	fmt.Println("before if api key")
	if apiKey != "" {
		fmt.Println("api key is not empty")
		// Remove the closing brace, append the new field, and add the brace back
		if len(confStr) > 1 && confStr[len(confStr)-1] == '}' {
			fmt.Println("confStr is not empty")
			// If config is just "{}", avoid the comma
			if confStr == "{}" {
				confStr = fmt.Sprintf(`{"datadogApiKey":"%s"}`, apiKey)
				fmt.Println("confStr is empty, new value is ", confStr)
			} else {
				confStr = confStr[:len(confStr)-1] + fmt.Sprintf(`,"datadogApiKey":"%s"}`, apiKey)
				fmt.Println("confStr is not empty, new value is ", confStr)
			}
		}
	}

	state.Config = types.StringValue(confStr)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *syntheticsPrivateLocationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	body, diags := r.buildSyntheticsPrivateLocationRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdatePrivateLocation(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating SyntheticsPrivateLocation"))
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

func (r *syntheticsPrivateLocationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state syntheticsPrivateLocationModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()

	httpResp, err := r.Api.DeletePrivateLocation(r.Auth, id)
	if err != nil {
		if httpResp == nil || httpResp.StatusCode != 404 {
			// The resource is assumed to still exist, and all prior state is preserved.
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting SyntheticsPrivateLocation"))
			return
		}
	}
}

func (r *syntheticsPrivateLocationResource) updateState(ctx context.Context, state *syntheticsPrivateLocationModel, resp *datadogV1.SyntheticsPrivateLocation) {
	state.Id = types.StringValue(resp.GetId())

	state.Description = types.StringValue(resp.GetDescription())
	state.Name = types.StringValue(resp.GetName())
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, resp.Tags)
	// Convert the private location ID to the format expected by restriction policies
	// The format should be: synthetics-private-location:pl:xxx (keep the pl: prefix)
	privateLocationId := resp.GetId()
	restrictionPolicyId := fmt.Sprintf("synthetics-private-location:%s", privateLocationId)
	state.RestrictionPolicyResourceId = types.StringValue(restrictionPolicyId)

	if metadata, ok := resp.GetMetadataOk(); ok {
		if restrictedRoles, ok := metadata.GetRestrictedRolesOk(); ok {
			localMetadata := syntheticsPrivateLocationMetadataModel{}
			localMetadata.RestrictedRoles, _ = types.SetValueFrom(ctx, types.StringType, *restrictedRoles)
			state.Metadata = []syntheticsPrivateLocationMetadataModel{localMetadata}
		}
	}
}

func (r *syntheticsPrivateLocationResource) buildSyntheticsPrivateLocationRequestBody(ctx context.Context, state *syntheticsPrivateLocationModel) (*datadogV1.SyntheticsPrivateLocation, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	syntheticsPrivateLocation := datadogV1.NewSyntheticsPrivateLocationWithDefaults()

	syntheticsPrivateLocation.SetName(state.Name.ValueString())

	if !state.Description.IsNull() {
		syntheticsPrivateLocation.SetDescription(state.Description.ValueString())
	}

	metadata := datadogV1.SyntheticsPrivateLocationMetadata{}
	if state.Metadata != nil {
		if len(state.Metadata) == 1 && !state.Metadata[0].RestrictedRoles.IsNull() {
			var restrictedRoles []string
			diags.Append(state.Metadata[0].RestrictedRoles.ElementsAs(ctx, &restrictedRoles, false)...)
			metadata.SetRestrictedRoles(restrictedRoles)
		}
	}
	syntheticsPrivateLocation.SetMetadata(metadata)

	tags := make([]string, 0)
	if !state.Tags.IsNull() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	}
	syntheticsPrivateLocation.SetTags(tags)

	return syntheticsPrivateLocation, diags
}
