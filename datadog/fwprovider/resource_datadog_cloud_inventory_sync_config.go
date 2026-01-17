package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &cloudInventorySyncConfigResource{}
	_ resource.ResourceWithImportState = &cloudInventorySyncConfigResource{}
)

// API paths
const (
	syncConfigsPath    = "/api/unstable/cloudinventoryservice/syncconfigs"
	syncConfigByIDPath = "/api/unstable/cloudinventoryservice/syncconfigs/%s"
)

type cloudInventorySyncConfigResource struct {
	Api  *datadog.APIClient
	Auth context.Context
}

type cloudInventorySyncConfigModel struct {
	ID            types.String `tfsdk:"id"`
	CloudProvider types.String `tfsdk:"cloud_provider"`
	Aws           *awsModel    `tfsdk:"aws"`
	Azure         *azureModel  `tfsdk:"azure"`
	Gcp           *gcpModel    `tfsdk:"gcp"`
}

type awsModel struct {
	AwsAccountId            types.String `tfsdk:"aws_account_id"`
	DestinationBucketName   types.String `tfsdk:"destination_bucket_name"`
	DestinationBucketRegion types.String `tfsdk:"destination_bucket_region"`
	DestinationPrefix       types.String `tfsdk:"destination_prefix"`
}

type azureModel struct {
	ClientId       types.String `tfsdk:"client_id"`
	TenantId       types.String `tfsdk:"tenant_id"`
	SubscriptionId types.String `tfsdk:"subscription_id"`
	ResourceGroup  types.String `tfsdk:"resource_group"`
	StorageAccount types.String `tfsdk:"storage_account"`
	Container      types.String `tfsdk:"container"`
}

type gcpModel struct {
	DestinationBucketName types.String `tfsdk:"destination_bucket_name"`
	ProjectId             types.String `tfsdk:"project_id"`
	ServiceAccountEmail   types.String `tfsdk:"service_account_email"`
	SourceBucketName      types.String `tfsdk:"source_bucket_name"`
}

// API request/response types - JSON:API format
type syncConfigRequest struct {
	Data *syncConfigRequestData `json:"data"`
}

type syncConfigRequestData struct {
	Type       string                       `json:"type"`
	ID         string                       `json:"id"`
	Attributes *syncConfigRequestAttributes `json:"attributes,omitempty"`
}

type syncConfigRequestAttributes struct {
	Aws   *awsRequestAttributes   `json:"aws,omitempty"`
	Azure *azureRequestAttributes `json:"azure,omitempty"`
	Gcp   *gcpRequestAttributes   `json:"gcp,omitempty"`
}

type awsRequestAttributes struct {
	AwsAccountId            string `json:"aws_account_id"`
	DestinationBucketName   string `json:"destination_bucket_name"`
	DestinationBucketRegion string `json:"destination_bucket_region"`
	DestinationPrefix       string `json:"destination_prefix,omitempty"`
}

type azureRequestAttributes struct {
	ClientId       string `json:"client_id"`
	TenantId       string `json:"tenant_id"`
	SubscriptionId string `json:"subscription_id"`
	ResourceGroup  string `json:"resource_group"`
	StorageAccount string `json:"storage_account"`
	Container      string `json:"container"`
}

type gcpRequestAttributes struct {
	ProjectId             string `json:"project_id"`
	DestinationBucketName string `json:"destination_bucket_name"`
	SourceBucketName      string `json:"source_bucket_name"`
	ServiceAccountEmail   string `json:"service_account_email"`
}

type syncConfigResponse struct {
	Data *syncConfigResponseData `json:"data"`
}

type syncConfigResponseData struct {
	Type       string                        `json:"type"`
	ID         string                        `json:"id"`
	Attributes *syncConfigResponseAttributes `json:"attributes,omitempty"`
}

type syncConfigResponseAttributes struct {
	CloudProvider string `json:"cloud_provider,omitempty"`
	// AWS response fields
	awsResponseAttributes
	// Azure response fields
	azureResponseAttributes
	// GCP response fields
	gcpResponseAttributes
	// Common fields
	Prefix    string `json:"prefix,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

type awsResponseAttributes struct {
	AwsAccountId  string `json:"aws_account_id,omitempty"`
	AwsBucketName string `json:"aws_bucket_name,omitempty"`
	AwsRegion     string `json:"aws_region,omitempty"`
}

type azureResponseAttributes struct {
	AzureClientId           string `json:"azure_client_id,omitempty"`
	AzureTenantId           string `json:"azure_tenant_id,omitempty"`
	AzureStorageAccountName string `json:"azure_storage_account_name,omitempty"`
	AzureContainerName      string `json:"azure_container_name,omitempty"`
}

type gcpResponseAttributes struct {
	GcpProjectId           string `json:"gcp_project_id,omitempty"`
	GcpBucketName          string `json:"gcp_bucket_name,omitempty"`
	GcpServiceAccountEmail string `json:"gcp_service_account_email,omitempty"`
}

func NewCloudInventorySyncConfigResource() resource.Resource {
	return &cloudInventorySyncConfigResource{}
}

func (r *cloudInventorySyncConfigResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.HttpClient
	r.Auth = providerData.Auth
}

func (r *cloudInventorySyncConfigResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "cloud_inventory_sync_config"
}

func (r *cloudInventorySyncConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog CloudInventorySyncConfig resource. This can be used to create and manage Datadog cloud_inventory_sync_config.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"cloud_provider": schema.StringAttribute{
				Description: "The cloud provider type. Valid values are `aws`, `azure`, `gcp`.",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"aws": schema.SingleNestedBlock{
				Description: "AWS-specific configuration. Required when cloud_provider is `aws`.",
				Attributes: map[string]schema.Attribute{
					"aws_account_id": schema.StringAttribute{
						Optional:    true,
						Description: "AWS Account ID of the account holding the bucket.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"destination_bucket_name": schema.StringAttribute{
						Optional:    true,
						Description: "Name of the S3 bucket holding the inventory files.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"destination_bucket_region": schema.StringAttribute{
						Optional:    true,
						Description: "AWS Region of the bucket holding the inventory files.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"destination_prefix": schema.StringAttribute{
						Optional:    true,
						Description: "Prefix path within the bucket for inventory files.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"azure": schema.SingleNestedBlock{
				Description: "Azure-specific configuration. Required when cloud_provider is `azure`.",
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Client ID.",
					},
					"tenant_id": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Tenant ID.",
					},
					"subscription_id": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Subscription ID.",
					},
					"resource_group": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Resource Group name.",
					},
					"storage_account": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Storage Account name.",
					},
					"container": schema.StringAttribute{
						Optional:    true,
						Description: "Azure Storage Container name.",
					},
				},
			},
			"gcp": schema.SingleNestedBlock{
				Description: "GCP-specific configuration. Required when cloud_provider is `gcp`.",
				Attributes: map[string]schema.Attribute{
					"project_id": schema.StringAttribute{
						Optional:    true,
						Description: "GCP Project ID of the project holding the bucket.",
					},
					"destination_bucket_name": schema.StringAttribute{
						Optional:    true,
						Description: "Name of the GCS bucket holding the inventory files.",
					},
					"source_bucket_name": schema.StringAttribute{
						Optional:    true,
						Description: "Name of the source bucket the inventory report is generated for.",
					},
					"service_account_email": schema.StringAttribute{
						Optional:    true,
						Description: "Service account email used for reading the bucket.",
					},
				},
			},
		},
	}
}

func (r *cloudInventorySyncConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *cloudInventorySyncConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	path := fmt.Sprintf(syncConfigByIDPath, id)

	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodGet, path, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving CloudInventorySyncConfig"))
		return
	}

	var resp syncConfigResponse
	if len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			response.Diagnostics.AddError("error unmarshalling response", err.Error())
			return
		}
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *cloudInventorySyncConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildCloudInventorySyncConfigRequestBody(ctx, &state)

	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodPut, syncConfigsPath, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating CloudInventorySyncConfig"))
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		response.Diagnostics.AddError("error creating CloudInventorySyncConfig", fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode))
		return
	}

	var resp syncConfigResponse
	if len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			response.Diagnostics.AddError("error parsing CloudInventorySyncConfig response", err.Error())
			return
		}
		r.updateState(ctx, &state, &resp)
	}

	if state.ID.IsNull() || state.ID.ValueString() == "" {
		response.Diagnostics.AddError("error creating CloudInventorySyncConfig", "no ID returned from API")
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *cloudInventorySyncConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildCloudInventorySyncConfigRequestBody(ctx, &state)

	// Uses same upsert endpoint as Create
	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodPut, syncConfigsPath, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating CloudInventorySyncConfig"))
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		response.Diagnostics.AddError("error updating CloudInventorySyncConfig", fmt.Sprintf("unexpected status code: %d", httpResp.StatusCode))
		return
	}

	var resp syncConfigResponse
	if len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, &resp); err == nil {
			r.updateState(ctx, &state, &resp)
		}
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *cloudInventorySyncConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	path := fmt.Sprintf(syncConfigByIDPath, id)

	_, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodDelete, path, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting cloud_inventory_sync_config"))
		return
	}
}

func (r *cloudInventorySyncConfigResource) updateState(ctx context.Context, state *cloudInventorySyncConfigModel, resp *syncConfigResponse) {
	if resp == nil || resp.Data == nil {
		return
	}

	if resp.Data.ID != "" {
		state.ID = types.StringValue(resp.Data.ID)
	}

	if resp.Data.Attributes == nil {
		return
	}

	attrs := resp.Data.Attributes

	if attrs.CloudProvider != "" {
		state.CloudProvider = types.StringValue(attrs.CloudProvider)
	}

	// Update AWS fields
	if attrs.AwsAccountId != "" || attrs.AwsBucketName != "" || attrs.AwsRegion != "" {
		if state.Aws == nil {
			state.Aws = &awsModel{}
		}
		if attrs.AwsAccountId != "" {
			state.Aws.AwsAccountId = types.StringValue(attrs.AwsAccountId)
		}
		if attrs.AwsBucketName != "" {
			state.Aws.DestinationBucketName = types.StringValue(attrs.AwsBucketName)
		}
		if attrs.AwsRegion != "" {
			state.Aws.DestinationBucketRegion = types.StringValue(attrs.AwsRegion)
		}
		if attrs.Prefix != "" {
			state.Aws.DestinationPrefix = types.StringValue(attrs.Prefix)
		}
	}

	// Update Azure fields
	if attrs.AzureClientId != "" || attrs.AzureStorageAccountName != "" {
		if state.Azure == nil {
			state.Azure = &azureModel{}
		}
		if attrs.AzureClientId != "" {
			state.Azure.ClientId = types.StringValue(attrs.AzureClientId)
		}
		if attrs.AzureTenantId != "" {
			state.Azure.TenantId = types.StringValue(attrs.AzureTenantId)
		}
		if attrs.AzureStorageAccountName != "" {
			state.Azure.StorageAccount = types.StringValue(attrs.AzureStorageAccountName)
		}
		if attrs.AzureContainerName != "" {
			state.Azure.Container = types.StringValue(attrs.AzureContainerName)
		}
	}

	// Update GCP fields
	if attrs.GcpProjectId != "" || attrs.GcpBucketName != "" {
		if state.Gcp == nil {
			state.Gcp = &gcpModel{}
		}
		if attrs.GcpProjectId != "" {
			state.Gcp.ProjectId = types.StringValue(attrs.GcpProjectId)
		}
		if attrs.GcpBucketName != "" {
			state.Gcp.DestinationBucketName = types.StringValue(attrs.GcpBucketName)
		}
		if attrs.GcpServiceAccountEmail != "" {
			state.Gcp.ServiceAccountEmail = types.StringValue(attrs.GcpServiceAccountEmail)
		}
	}
}

func (r *cloudInventorySyncConfigResource) buildCloudInventorySyncConfigRequestBody(ctx context.Context, state *cloudInventorySyncConfigModel) *syncConfigRequest {
	cloudProvider := state.CloudProvider.ValueString()
	attributes := &syncConfigRequestAttributes{}

	if state.Aws != nil {
		attributes.Aws = &awsRequestAttributes{}
		if !state.Aws.AwsAccountId.IsNull() {
			attributes.Aws.AwsAccountId = state.Aws.AwsAccountId.ValueString()
		}
		if !state.Aws.DestinationBucketName.IsNull() {
			attributes.Aws.DestinationBucketName = state.Aws.DestinationBucketName.ValueString()
		}
		if !state.Aws.DestinationBucketRegion.IsNull() {
			attributes.Aws.DestinationBucketRegion = state.Aws.DestinationBucketRegion.ValueString()
		}
		if !state.Aws.DestinationPrefix.IsNull() {
			attributes.Aws.DestinationPrefix = state.Aws.DestinationPrefix.ValueString()
		}
	}

	if state.Azure != nil {
		attributes.Azure = &azureRequestAttributes{}
		if !state.Azure.ClientId.IsNull() {
			attributes.Azure.ClientId = state.Azure.ClientId.ValueString()
		}
		if !state.Azure.TenantId.IsNull() {
			attributes.Azure.TenantId = state.Azure.TenantId.ValueString()
		}
		if !state.Azure.SubscriptionId.IsNull() {
			attributes.Azure.SubscriptionId = state.Azure.SubscriptionId.ValueString()
		}
		if !state.Azure.ResourceGroup.IsNull() {
			attributes.Azure.ResourceGroup = state.Azure.ResourceGroup.ValueString()
		}
		if !state.Azure.StorageAccount.IsNull() {
			attributes.Azure.StorageAccount = state.Azure.StorageAccount.ValueString()
		}
		if !state.Azure.Container.IsNull() {
			attributes.Azure.Container = state.Azure.Container.ValueString()
		}
	}

	if state.Gcp != nil {
		attributes.Gcp = &gcpRequestAttributes{}
		if !state.Gcp.ProjectId.IsNull() {
			attributes.Gcp.ProjectId = state.Gcp.ProjectId.ValueString()
		}
		if !state.Gcp.DestinationBucketName.IsNull() {
			attributes.Gcp.DestinationBucketName = state.Gcp.DestinationBucketName.ValueString()
		}
		if !state.Gcp.SourceBucketName.IsNull() {
			attributes.Gcp.SourceBucketName = state.Gcp.SourceBucketName.ValueString()
		}
		if !state.Gcp.ServiceAccountEmail.IsNull() {
			attributes.Gcp.ServiceAccountEmail = state.Gcp.ServiceAccountEmail.ValueString()
		}
	}

	return &syncConfigRequest{
		Data: &syncConfigRequestData{
			Type:       "cloud_provider",
			ID:         cloudProvider,
			Attributes: attributes,
		},
	}
}
