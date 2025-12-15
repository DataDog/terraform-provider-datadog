package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const (
	createSyncConfigPath = "/api/unstable/cloudinventoryservice/createsyncconfig"
	getSyncStatusPath    = "/api/unstable/cloudinventoryservice/syncstatus"
)

// CloudProvider represents the cloud provider type
type cloudProvider string

const (
	cloudProviderAWS   cloudProvider = "aws"
	cloudProviderGCP   cloudProvider = "gcp"
	cloudProviderAzure cloudProvider = "azure"
)

// API request/response types - JSON:API format
type syncConfigRequestWrapper struct {
	Data syncConfigRequestData `json:"data"`
}

type syncConfigRequestData struct {
	Type       string                      `json:"type"`
	ID         string                      `json:"id"`
	Attributes syncConfigRequestAttributes `json:"attributes"`
}

type syncConfigRequestAttributes struct {
	AWS   *awsSyncConfigRequest   `json:"aws,omitempty"`
	Azure *azureSyncConfigRequest `json:"azure,omitempty"`
	GCP   *gcpSyncConfigRequest   `json:"gcp,omitempty"`
}

type awsSyncConfigRequest struct {
	AWSAccountID            string `json:"aws_account_id"`
	DestinationBucketName   string `json:"destination_bucket_name"`
	DestinationBucketRegion string `json:"destination_bucket_region"`
	DestinationPrefix       string `json:"destination_prefix,omitempty"`
}

type azureSyncConfigRequest struct {
	ClientID       string `json:"client_id"`
	TenantID       string `json:"tenant_id"`
	SubscriptionID string `json:"subscription_id"`
	ResourceGroup  string `json:"resource_group"`
	StorageAccount string `json:"storage_account"`
	Container      string `json:"container"`
}

type gcpSyncConfigRequest struct {
	ProjectID             string `json:"project_id"`
	DestinationBucketName string `json:"destination_bucket_name"`
	SourceBucketName      string `json:"source_bucket_name"`
	ServiceAccountEmail   string `json:"service_account_email"`
}

type syncConfigResponse struct {
	ID string `json:"id"`
}

type cloudInventorySyncStatusResponse struct {
	SyncStatuses []syncLevelStatus `json:"sync_statuses"`
}

type syncLevelStatus struct {
	// AWS fields
	AWSBucketName string `json:"aws_bucket_name,omitempty"`
	AWSAccountID  string `json:"aws_account_id,omitempty"`
	AWSRegion     string `json:"aws_region,omitempty"`

	// Azure fields
	AzureStorageAccountName string `json:"azure_storage_account_name,omitempty"`
	AzureContainerName      string `json:"azure_container_name,omitempty"`
	AzureClientID           string `json:"azure_client_id,omitempty"`
	AzureTenantID           string `json:"azure_tenant_id,omitempty"`

	// GCP fields
	GCPBucketName          string `json:"gcp_bucket_name,omitempty"`
	GCPProjectID           string `json:"gcp_project_id,omitempty"`
	GCPServiceAccountEmail string `json:"gcp_service_account_email,omitempty"`

	// Common fields
	Prefix    string `json:"prefix,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

var (
	_ resource.ResourceWithConfigure   = &cloudInventorySyncConfigResource{}
	_ resource.ResourceWithImportState = &cloudInventorySyncConfigResource{}
)

type cloudInventorySyncConfigResource struct {
	Api  *datadog.APIClient
	Auth context.Context
}

type cloudInventorySyncConfigModel struct {
	ID            types.String `tfsdk:"id"`
	CloudProvider types.String `tfsdk:"cloud_provider"`

	// AWS fields
	AWSAccountID            types.String `tfsdk:"aws_account_id"`
	DestinationBucketName   types.String `tfsdk:"destination_bucket_name"`
	DestinationBucketRegion types.String `tfsdk:"destination_bucket_region"`
	DestinationPrefix       types.String `tfsdk:"destination_prefix"`

	// Azure fields
	AzureClientID       types.String `tfsdk:"azure_client_id"`
	AzureTenantID       types.String `tfsdk:"azure_tenant_id"`
	AzureSubscriptionID types.String `tfsdk:"azure_subscription_id"`
	AzureResourceGroup  types.String `tfsdk:"azure_resource_group"`
	AzureStorageAccount types.String `tfsdk:"azure_storage_account"`
	AzureContainer      types.String `tfsdk:"azure_container"`

	// GCP fields
	GCPProjectID           types.String `tfsdk:"gcp_project_id"`
	GCPDestinationBucket   types.String `tfsdk:"gcp_destination_bucket_name"`
	GCPSourceBucket        types.String `tfsdk:"gcp_source_bucket_name"`
	GCPServiceAccountEmail types.String `tfsdk:"gcp_service_account_email"`
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
		Description: "Provides a Datadog Cloud Inventory Sync Config resource. This can be used to create cloud inventory sync configurations for AWS, Azure, or GCP.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"cloud_provider": schema.StringAttribute{
				Description: "The cloud provider type. Valid values are `aws`, `azure`, `gcp`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "azure", "gcp"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// AWS attributes
			"aws_account_id": schema.StringAttribute{
				Description: "AWS Account ID of the account holding the bucket. Required when cloud_provider is `aws`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_bucket_name": schema.StringAttribute{
				Description: "Name of the bucket holding the inventory files. Required when cloud_provider is `aws`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_bucket_region": schema.StringAttribute{
				Description: "AWS Region of the bucket holding the inventory files. Required when cloud_provider is `aws`.",
				Optional:    true,
			},
			"destination_prefix": schema.StringAttribute{
				Description: "Name of the prefix holding the inventory files.",
				Optional:    true,
			},

			// Azure attributes
			"azure_client_id": schema.StringAttribute{
				Description: "Azure Client ID. Required when cloud_provider is `azure`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"azure_tenant_id": schema.StringAttribute{
				Description: "Azure Tenant ID. Required when cloud_provider is `azure`.",
				Optional:    true,
			},
			"azure_subscription_id": schema.StringAttribute{
				Description: "Azure Subscription ID. Required when cloud_provider is `azure`.",
				Optional:    true,
			},
			"azure_resource_group": schema.StringAttribute{
				Description: "Azure Resource Group. Required when cloud_provider is `azure`.",
				Optional:    true,
			},
			"azure_storage_account": schema.StringAttribute{
				Description: "Azure Storage Account name. Required when cloud_provider is `azure`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"azure_container": schema.StringAttribute{
				Description: "Azure Container name. Required when cloud_provider is `azure`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// GCP attributes
			"gcp_project_id": schema.StringAttribute{
				Description: "GCP Project ID of the project holding the bucket. Required when cloud_provider is `gcp`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"gcp_destination_bucket_name": schema.StringAttribute{
				Description: "Name of the GCP bucket holding the inventory files. Required when cloud_provider is `gcp`.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"gcp_source_bucket_name": schema.StringAttribute{
				Description: "Name of the GCP bucket the inventory report is generated for. Required when cloud_provider is `gcp`.",
				Optional:    true,
			},
			"gcp_service_account_email": schema.StringAttribute{
				Description: "Service account email used for reading the bucket. Required when cloud_provider is `gcp`.",
				Optional:    true,
			},
		},
	}
}

func (r *cloudInventorySyncConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *cloudInventorySyncConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	provider := cloudProvider(state.CloudProvider.ValueString())
	identifiers := r.getIdentifiers(&state)

	syncStatus, httpResp, err := r.findSyncStatus(provider, identifiers)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving cloud inventory sync config"))
		return
	}

	if syncStatus == nil {
		// Config not found in the list
		response.State.RemoveResource(ctx)
		return
	}

	r.updateStateFromStatus(ctx, &state, syncStatus)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *cloudInventorySyncConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Validate required fields based on provider
	if diags := r.validateProviderFields(&state); diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	body := r.buildRequestBody(&state)

	respBytes, _, err := utils.SendRequest(r.Auth, r.Api, "PUT", createSyncConfigPath, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating cloud inventory sync config"))
		return
	}

	// Set the ID based on provider and identifiers
	state.ID = types.StringValue(r.generateID(&state))

	// Try to parse response if not empty
	if len(respBytes) > 0 {
		var resp syncConfigResponse
		if err := json.Unmarshal(respBytes, &resp); err == nil && resp.ID != "" {
			state.ID = types.StringValue(resp.ID)
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *cloudInventorySyncConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state cloudInventorySyncConfigModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Since create is idempotent, we can just call create again
	body := r.buildRequestBody(&state)

	_, _, err := utils.SendRequest(r.Auth, r.Api, "PUT", createSyncConfigPath, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating cloud inventory sync config"))
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *cloudInventorySyncConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddWarning(
		"Resource cannot be deleted",
		"Cloud Inventory Sync Config cannot be deleted from Datadog and is only removed from Terraform state. The sync configuration will continue to exist in Datadog.",
	)
}

func (r *cloudInventorySyncConfigResource) validateProviderFields(state *cloudInventorySyncConfigModel) diag.Diagnostics {
	var diags diag.Diagnostics
	provider := state.CloudProvider.ValueString()

	switch provider {
	case "aws":
		if state.AWSAccountID.IsNull() || state.AWSAccountID.ValueString() == "" {
			diags.AddAttributeError(path.Root("aws_account_id"), "Missing required field", "aws_account_id is required when cloud_provider is 'aws'")
		}
		if state.DestinationBucketName.IsNull() || state.DestinationBucketName.ValueString() == "" {
			diags.AddAttributeError(path.Root("destination_bucket_name"), "Missing required field", "destination_bucket_name is required when cloud_provider is 'aws'")
		}
		if state.DestinationBucketRegion.IsNull() || state.DestinationBucketRegion.ValueString() == "" {
			diags.AddAttributeError(path.Root("destination_bucket_region"), "Missing required field", "destination_bucket_region is required when cloud_provider is 'aws'")
		}
	case "azure":
		if state.AzureClientID.IsNull() || state.AzureClientID.ValueString() == "" {
			diags.AddAttributeError(path.Root("azure_client_id"), "Missing required field", "azure_client_id is required when cloud_provider is 'azure'")
		}
		if state.AzureTenantID.IsNull() || state.AzureTenantID.ValueString() == "" {
			diags.AddAttributeError(path.Root("azure_tenant_id"), "Missing required field", "azure_tenant_id is required when cloud_provider is 'azure'")
		}
		if state.AzureSubscriptionID.IsNull() || state.AzureSubscriptionID.ValueString() == "" {
			diags.AddAttributeError(path.Root("azure_subscription_id"), "Missing required field", "azure_subscription_id is required when cloud_provider is 'azure'")
		}
		if state.AzureResourceGroup.IsNull() || state.AzureResourceGroup.ValueString() == "" {
			diags.AddAttributeError(path.Root("azure_resource_group"), "Missing required field", "azure_resource_group is required when cloud_provider is 'azure'")
		}
		if state.AzureStorageAccount.IsNull() || state.AzureStorageAccount.ValueString() == "" {
			diags.AddAttributeError(path.Root("azure_storage_account"), "Missing required field", "azure_storage_account is required when cloud_provider is 'azure'")
		}
		if state.AzureContainer.IsNull() || state.AzureContainer.ValueString() == "" {
			diags.AddAttributeError(path.Root("azure_container"), "Missing required field", "azure_container is required when cloud_provider is 'azure'")
		}
	case "gcp":
		if state.GCPProjectID.IsNull() || state.GCPProjectID.ValueString() == "" {
			diags.AddAttributeError(path.Root("gcp_project_id"), "Missing required field", "gcp_project_id is required when cloud_provider is 'gcp'")
		}
		if state.GCPDestinationBucket.IsNull() || state.GCPDestinationBucket.ValueString() == "" {
			diags.AddAttributeError(path.Root("gcp_destination_bucket_name"), "Missing required field", "gcp_destination_bucket_name is required when cloud_provider is 'gcp'")
		}
		if state.GCPSourceBucket.IsNull() || state.GCPSourceBucket.ValueString() == "" {
			diags.AddAttributeError(path.Root("gcp_source_bucket_name"), "Missing required field", "gcp_source_bucket_name is required when cloud_provider is 'gcp'")
		}
		if state.GCPServiceAccountEmail.IsNull() || state.GCPServiceAccountEmail.ValueString() == "" {
			diags.AddAttributeError(path.Root("gcp_service_account_email"), "Missing required field", "gcp_service_account_email is required when cloud_provider is 'gcp'")
		}
	}

	return diags
}

func (r *cloudInventorySyncConfigResource) buildRequestBody(state *cloudInventorySyncConfigModel) *syncConfigRequestWrapper {
	provider := cloudProvider(state.CloudProvider.ValueString())
	attributes := syncConfigRequestAttributes{}

	switch provider {
	case cloudProviderAWS:
		attributes.AWS = &awsSyncConfigRequest{
			AWSAccountID:            state.AWSAccountID.ValueString(),
			DestinationBucketName:   state.DestinationBucketName.ValueString(),
			DestinationBucketRegion: state.DestinationBucketRegion.ValueString(),
			DestinationPrefix:       state.DestinationPrefix.ValueString(),
		}
	case cloudProviderAzure:
		attributes.Azure = &azureSyncConfigRequest{
			ClientID:       state.AzureClientID.ValueString(),
			TenantID:       state.AzureTenantID.ValueString(),
			SubscriptionID: state.AzureSubscriptionID.ValueString(),
			ResourceGroup:  state.AzureResourceGroup.ValueString(),
			StorageAccount: state.AzureStorageAccount.ValueString(),
			Container:      state.AzureContainer.ValueString(),
		}
	case cloudProviderGCP:
		attributes.GCP = &gcpSyncConfigRequest{
			ProjectID:             state.GCPProjectID.ValueString(),
			DestinationBucketName: state.GCPDestinationBucket.ValueString(),
			SourceBucketName:      state.GCPSourceBucket.ValueString(),
			ServiceAccountEmail:   state.GCPServiceAccountEmail.ValueString(),
		}
	}

	return &syncConfigRequestWrapper{
		Data: syncConfigRequestData{
			Type:       "cloud_provider",
			ID:         string(provider),
			Attributes: attributes,
		},
	}
}

func (r *cloudInventorySyncConfigResource) getIdentifiers(state *cloudInventorySyncConfigModel) map[string]string {
	provider := state.CloudProvider.ValueString()
	identifiers := make(map[string]string)

	switch provider {
	case "aws":
		identifiers["aws_account_id"] = state.AWSAccountID.ValueString()
		identifiers["destination_bucket_name"] = state.DestinationBucketName.ValueString()
	case "azure":
		identifiers["client_id"] = state.AzureClientID.ValueString()
		identifiers["storage_account"] = state.AzureStorageAccount.ValueString()
		identifiers["container"] = state.AzureContainer.ValueString()
	case "gcp":
		identifiers["project_id"] = state.GCPProjectID.ValueString()
		identifiers["destination_bucket_name"] = state.GCPDestinationBucket.ValueString()
	}

	return identifiers
}

func (r *cloudInventorySyncConfigResource) generateID(state *cloudInventorySyncConfigModel) string {
	provider := state.CloudProvider.ValueString()

	switch provider {
	case "aws":
		return fmt.Sprintf("aws:%s:%s", state.AWSAccountID.ValueString(), state.DestinationBucketName.ValueString())
	case "azure":
		return fmt.Sprintf("azure:%s:%s:%s", state.AzureClientID.ValueString(), state.AzureStorageAccount.ValueString(), state.AzureContainer.ValueString())
	case "gcp":
		return fmt.Sprintf("gcp:%s:%s", state.GCPProjectID.ValueString(), state.GCPDestinationBucket.ValueString())
	}

	return ""
}

func (r *cloudInventorySyncConfigResource) updateStateFromStatus(ctx context.Context, state *cloudInventorySyncConfigModel, status *syncLevelStatus) {
	provider := state.CloudProvider.ValueString()

	switch provider {
	case "aws":
		if status.AWSAccountID != "" {
			state.AWSAccountID = types.StringValue(status.AWSAccountID)
		}
		if status.AWSBucketName != "" {
			state.DestinationBucketName = types.StringValue(status.AWSBucketName)
		}
		if status.AWSRegion != "" {
			state.DestinationBucketRegion = types.StringValue(status.AWSRegion)
		}
		if status.Prefix != "" {
			state.DestinationPrefix = types.StringValue(status.Prefix)
		}
	case "azure":
		if status.AzureClientID != "" {
			state.AzureClientID = types.StringValue(status.AzureClientID)
		}
		if status.AzureTenantID != "" {
			state.AzureTenantID = types.StringValue(status.AzureTenantID)
		}
		if status.AzureStorageAccountName != "" {
			state.AzureStorageAccount = types.StringValue(status.AzureStorageAccountName)
		}
		if status.AzureContainerName != "" {
			state.AzureContainer = types.StringValue(status.AzureContainerName)
		}
	case "gcp":
		if status.GCPProjectID != "" {
			state.GCPProjectID = types.StringValue(status.GCPProjectID)
		}
		if status.GCPBucketName != "" {
			state.GCPDestinationBucket = types.StringValue(status.GCPBucketName)
		}
		if status.GCPServiceAccountEmail != "" {
			state.GCPServiceAccountEmail = types.StringValue(status.GCPServiceAccountEmail)
		}
	}
}

// getSyncStatuses retrieves all sync statuses
func (r *cloudInventorySyncConfigResource) getSyncStatuses() (*cloudInventorySyncStatusResponse, error) {
	respBytes, _, err := utils.SendRequest(r.Auth, r.Api, "GET", getSyncStatusPath, nil)
	if err != nil {
		return nil, err
	}

	var response cloudInventorySyncStatusResponse
	if err := json.Unmarshal(respBytes, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &response, nil
}

// findSyncStatus finds a specific sync status by matching identifiers
func (r *cloudInventorySyncConfigResource) findSyncStatus(provider cloudProvider, identifiers map[string]string) (*syncLevelStatus, *struct{ StatusCode int }, error) {
	response, err := r.getSyncStatuses()
	if err != nil {
		return nil, nil, err
	}

	for _, status := range response.SyncStatuses {
		if matchesSyncStatus(&status, provider, identifiers) {
			return &status, nil, nil
		}
	}

	return nil, nil, nil
}

// matchesSyncStatus checks if a syncLevelStatus matches the given provider and identifiers
func matchesSyncStatus(status *syncLevelStatus, provider cloudProvider, identifiers map[string]string) bool {
	switch provider {
	case cloudProviderAWS:
		return status.AWSAccountID == identifiers["aws_account_id"] &&
			status.AWSBucketName == identifiers["destination_bucket_name"]
	case cloudProviderAzure:
		return status.AzureClientID == identifiers["client_id"] &&
			status.AzureStorageAccountName == identifiers["storage_account"] &&
			status.AzureContainerName == identifiers["container"]
	case cloudProviderGCP:
		return status.GCPProjectID == identifiers["project_id"] &&
			status.GCPBucketName == identifiers["destination_bucket_name"]
	}
	return false
}
