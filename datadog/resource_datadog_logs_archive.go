package datadog

import (
	"context"
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogLogsArchive() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Archive API resource, which is used to create and manage Datadog logs archives.",
		CreateContext: resourceDatadogLogsArchiveCreate,
		UpdateContext: resourceDatadogLogsArchiveUpdate,
		ReadContext:   resourceDatadogLogsArchiveRead,
		DeleteContext: resourceDatadogLogsArchiveDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name":  {Description: "Your archive name.", Type: schema.TypeString, Required: true},
				"query": {Description: "The archive query/filter. Logs matching this query are included in the archive.", Type: schema.TypeString, Required: true},
				"s3_archive": {
					Description: "Definition of an s3 archive.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bucket":     {Description: "Name of your s3 bucket.", Type: schema.TypeString, Required: true},
							"path":       {Description: "Path where the archive is stored.", Type: schema.TypeString, Optional: true},
							"account_id": {Description: "Your AWS account id.", Type: schema.TypeString, Required: true},
							"role_name":  {Description: "Your AWS role name", Type: schema.TypeString, Required: true},
						},
					},
				},
				"azure_archive": {
					Description: "Definition of an azure archive.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"container":       {Description: "The container where the archive is stored.", Type: schema.TypeString, Required: true},
							"client_id":       {Description: "Your client id.", Type: schema.TypeString, Required: true},
							"tenant_id":       {Description: "Your tenant id.", Type: schema.TypeString, Required: true},
							"storage_account": {Description: "The associated storage account.", Type: schema.TypeString, Required: true},
							"path":            {Description: "The path where the archive is stored.", Type: schema.TypeString, Optional: true},
						},
					},
				},
				"gcs_archive": {
					Description: "Definition of a GCS archive.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"bucket":       {Description: "Name of your GCS bucket.", Type: schema.TypeString, Required: true},
							"path":         {Description: "Path where the archive is stored.", Type: schema.TypeString, Optional: true},
							"client_email": {Description: "Your client email.", Type: schema.TypeString, Required: true},
							"project_id":   {Description: "Your project id.", Type: schema.TypeString, Required: true},
						},
					},
				},
				"rehydration_tags": {
					Description: "An array of tags to add to rehydrated logs from an archive.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"include_tags": {
					Description: "To store the tags in the archive, set the value `true`. If it is set to `false`, the tags will be dropped when the logs are sent to the archive.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
				"rehydration_max_scan_size_in_gb": {
					Description: "To limit the rehydration scan size for the archive, set a value in GB.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
			}
		},
	}
}

func resourceDatadogLogsArchiveCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ddArchive, err := buildDatadogArchiveCreateReq(d)
	if err != nil {
		return diag.FromErr(err)
	}
	createdArchive, httpResponse, err := apiInstances.GetLogsArchivesApiV2().CreateLogsArchive(auth, *ddArchive)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "failed to create logs archive using Datadog API")
	}
	if err := utils.CheckForUnparsed(createdArchive); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(*createdArchive.GetData().Id)
	return updateLogsArchiveState(d, &createdArchive)
}

func updateLogsArchiveState(d *schema.ResourceData, ddArchive *datadogV2.LogsArchive) diag.Diagnostics {
	if !ddArchive.HasData() {
		d.SetId("")
		return nil
	}
	if err := d.Set("name", ddArchive.Data.Attributes.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("query", ddArchive.Data.Attributes.Query); err != nil {
		return diag.FromErr(err)
	}
	archiveType, destination, err := buildDestination(ddArchive.Data.Attributes.Destination)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(fmt.Sprintf("%s_archive", archiveType), []interface{}{destination}); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("rehydration_tags", ddArchive.Data.Attributes.RehydrationTags); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("include_tags", ddArchive.Data.Attributes.IncludeTags); err != nil {
		return diag.FromErr(err)
	}

	rehydrationMaxSizeValue := ddArchive.Data.Attributes.RehydrationMaxScanSizeInGb.Get()
	if rehydrationMaxSizeValue != nil {
		if err = d.Set("rehydration_max_scan_size_in_gb", rehydrationMaxSizeValue); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDatadogLogsArchiveRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ddArchive, httpresp, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchive(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "failed to get logs archive using Datadog API")
	}
	if err := utils.CheckForUnparsed(ddArchive); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsArchiveState(d, &ddArchive)
}

func resourceDatadogLogsArchiveUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ddArchive, err := buildDatadogArchiveCreateReq(d)
	if err != nil {
		return diag.FromErr(err)
	}
	updatedArchive, httpResponse, err := apiInstances.GetLogsArchivesApiV2().UpdateLogsArchive(auth, d.Id(), *ddArchive)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs archive")
	}
	if err := utils.CheckForUnparsed(updatedArchive); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsArchiveState(d, &updatedArchive)
}

func resourceDatadogLogsArchiveDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpresp, err := apiInstances.GetLogsArchivesApiV2().DeleteLogsArchive(auth, d.Id()); err != nil {
		// API returns 404 when the specific archive id doesn't exist.
		if httpresp != nil && httpresp.StatusCode == 404 {
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting logs archive")
	}
	return nil
}

// Model to map
func buildDestination(archiveDestination datadogV2.NullableLogsArchiveDestination) (string, map[string]interface{}, error) {
	emptyDestination := map[string]interface{}{}
	if archiveDestination.IsSet() && archiveDestination.Get() != nil {
		destination := archiveDestination.Get().GetActualInstance()
		switch d := destination.(type) {
		case *datadogV2.LogsArchiveDestinationAzure:
			return "azure", buildAzureMap(*d), nil
		case *datadogV2.LogsArchiveDestinationGCS:
			return "gcs", buildGCSMap(*d), nil
		case *datadogV2.LogsArchiveDestinationS3:
			return "s3", buildS3Map(*d), nil
		}
	}
	return "", emptyDestination, fmt.Errorf("destination should be not null")
}

func buildAzureMap(destination datadogV2.LogsArchiveDestinationAzure) map[string]interface{} {
	result := make(map[string]interface{})
	integration := destination.GetIntegration()
	result["client_id"] = integration.GetClientId()
	result["tenant_id"] = integration.GetTenantId()
	result["container"] = destination.GetContainer()
	result["storage_account"] = destination.GetStorageAccount()
	result["path"] = destination.GetPath()
	return result
}

func buildGCSMap(destination datadogV2.LogsArchiveDestinationGCS) map[string]interface{} {
	result := make(map[string]interface{})
	integration := destination.GetIntegration()
	result["client_email"] = integration.GetClientEmail()
	result["project_id"] = integration.GetProjectId()
	result["bucket"] = destination.GetBucket()
	result["path"] = destination.GetPath()
	return result
}

func buildS3Map(destination datadogV2.LogsArchiveDestinationS3) map[string]interface{} {
	result := make(map[string]interface{})
	integration := destination.GetIntegration()
	result["account_id"] = integration.GetAccountId()
	result["role_name"] = integration.GetRoleName()
	result["bucket"] = destination.GetBucket()
	result["path"] = destination.GetPath()
	return result
}

// Map to model
func buildDatadogArchiveCreateReq(d *schema.ResourceData) (*datadogV2.LogsArchiveCreateRequest, error) {
	archive := datadogV2.NewLogsArchiveCreateRequest()
	destination, err := buildCreateReqDestination(d)
	if err != nil {
		return archive, err
	}
	attributes := datadogV2.NewLogsArchiveCreateRequestAttributes(
		*destination,
		d.Get("name").(string),
		d.Get("query").(string),
	)
	attributes.SetRehydrationTags(getRehydrationTags(d))
	attributes.SetIncludeTags(d.Get("include_tags").(bool))

	rehydrationMaxSizeValue, isRehydrationMaxSizeSet := d.GetOk("rehydration_max_scan_size_in_gb")
	if isRehydrationMaxSizeSet {
		attributes.SetRehydrationMaxScanSizeInGb(int64(rehydrationMaxSizeValue.(int)))
	} else {
		attributes.SetRehydrationMaxScanSizeInGbNil()
	}

	definition := datadogV2.NewLogsArchiveCreateRequestDefinitionWithDefaults()
	definition.SetAttributes(*attributes)
	archive.SetData(*definition)
	return archive, nil
}

func buildCreateReqDestination(d *schema.ResourceData) (*datadogV2.LogsArchiveCreateRequestDestination, error) {
	defDestinations := definedDestinations(d)
	if len(defDestinations) != 1 {
		return nil, fmt.Errorf("more than one or no destination type defined: %v", defDestinations)
	}
	archiveType := defDestinations[0]
	destinationMap := d.Get(archiveType)
	switch strings.TrimSuffix(archiveType, "_archive") {
	case string(datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE):
		destination, err := buildAzureDestination(destinationMap)
		if err != nil {
			return nil, err
		}
		result := datadogV2.LogsArchiveDestinationAzureAsLogsArchiveCreateRequestDestination(destination)
		return &result, nil
	case string(datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS):
		destination, err := buildGCSDestination(destinationMap)
		if err != nil {
			return nil, err
		}
		result := datadogV2.LogsArchiveDestinationGCSAsLogsArchiveCreateRequestDestination(destination)
		return &result, nil
	case string(datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3):
		destination, err := buildS3Destination(destinationMap)
		if err != nil {
			return nil, err
		}
		result := datadogV2.LogsArchiveDestinationS3AsLogsArchiveCreateRequestDestination(destination)
		return &result, nil
	default:
		return nil, fmt.Errorf("archive type '%s' doesn't exist", archiveType)
	}
}

func definedDestinations(d *schema.ResourceData) []string {
	var defined []string
	for _, destination := range []string{"azure_archive", "gcs_archive", "s3_archive"} {
		if _, ok := d.GetOk(destination); ok {
			defined = append(defined, destination)
		}
	}
	return defined
}

func buildAzureDestination(dest interface{}) (*datadogV2.LogsArchiveDestinationAzure, error) {
	d := dest.([]interface{})[0].(map[string]interface{})
	clientID, ok := d["client_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationAzure{}, fmt.Errorf("client_id is not defined")
	}
	tenantID, ok := d["tenant_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationAzure{}, fmt.Errorf("tenant_id is not defined")
	}
	integration := datadogV2.NewLogsArchiveIntegrationAzure(
		clientID.(string),
		tenantID.(string),
	)
	container, ok := d["container"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationAzure{}, fmt.Errorf("container is not defined")
	}
	storageAccount, ok := d["storage_account"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationAzure{}, fmt.Errorf("storage_account is not defined")
	}
	path, ok := d["path"]
	if !ok {
		path = ""
	}
	destination := datadogV2.NewLogsArchiveDestinationAzure(
		container.(string),
		*integration,
		storageAccount.(string),
		datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE,
	)
	destination.Path = datadog.PtrString(path.(string))
	return destination, nil
}

func buildGCSDestination(dest interface{}) (*datadogV2.LogsArchiveDestinationGCS, error) {
	d := dest.([]interface{})[0].(map[string]interface{})
	clientEmail, ok := d["client_email"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationGCS{}, fmt.Errorf("client_email is not defined")
	}
	projectID, ok := d["project_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationGCS{}, fmt.Errorf("project_id is not defined")
	}
	integration := datadogV2.NewLogsArchiveIntegrationGCS(
		clientEmail.(string),
		projectID.(string),
	)
	bucket, ok := d["bucket"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationGCS{}, fmt.Errorf("bucket is not defined")
	}
	path, ok := d["path"]
	if !ok {
		path = ""
	}
	destination := datadogV2.NewLogsArchiveDestinationGCS(
		bucket.(string),
		*integration,
		datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS,
	)
	destination.Path = datadog.PtrString(path.(string))
	return destination, nil
}

func buildS3Destination(dest interface{}) (*datadogV2.LogsArchiveDestinationS3, error) {
	d := dest.([]interface{})[0].(map[string]interface{})
	accountID, ok := d["account_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationS3{}, fmt.Errorf("account_id is not defined")
	}
	roleName, ok := d["role_name"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationS3{}, fmt.Errorf("role_name is not defined")
	}
	integration := datadogV2.NewLogsArchiveIntegrationS3(
		accountID.(string),
		roleName.(string),
	)
	bucket, ok := d["bucket"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationS3{}, fmt.Errorf("bucket is not defined")
	}
	path, ok := d["path"]
	if !ok {
		path = ""
	}
	destination := datadogV2.NewLogsArchiveDestinationS3(
		bucket.(string),
		*integration,
		datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3,
	)
	destination.Path = datadog.PtrString(path.(string))
	return destination, nil
}

func getRehydrationTags(d *schema.ResourceData) []string {
	tfList := d.Get("rehydration_tags").([]interface{})
	ddList := make([]string, len(tfList))
	for i, tfName := range tfList {
		ddList[i] = tfName.(string)
	}
	return ddList
}
