package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &serviceAccountResource{}
	_ resource.ResourceWithImportState = &serviceAccountResource{}
)

type serviceAccountResource struct {
	Api  *datadogV2.ServiceAccountsApi
	Auth context.Context
}
type serviceAccountResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Email    types.String `tfsdk:"email"`
	Name     types.String `tfsdk:"name"`
	Roles    types.Set    `tfsdk:"roles"`
}

func NewServiceAccountResource() resource.Resource {
	return &serviceAccountResource{}
}

func (*serviceAccountResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "service_account"
}

func (r *serviceAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceAccountsApiV2()
	r.Auth = providerData.Auth
}

func (r *serviceAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog `service_account_application_key` resource. This can be used to create and manage Datadog service account application keys.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name for the service account.",
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the service account is disabled.",
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"email": schema.StringAttribute{
				Description: "Email of the associated user.",
				Required:    true,
			},
			"roles": schema.SetAttribute{
				Description: "A list a role IDs to assign to the service account.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (*serviceAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (*serviceAccountResource) Create(context.Context, resource.CreateRequest, *resource.CreateResponse) {
	panic("unimplemented")
}

func (*serviceAccountResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	panic("unimplemented")
}

func (*serviceAccountResource) Read(context.Context, resource.ReadRequest, *resource.ReadResponse) {
	panic("unimplemented")
}

func (*serviceAccountResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("unimplemented")
}
