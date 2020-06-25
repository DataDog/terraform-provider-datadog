package datadog

import (
	"fmt"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogLogsArchive() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsArchiveCreate,
		Update: resourceDatadogLogsArchiveUpdate,
		Read:   resourceDatadogLogsArchiveRead,
		Delete: resourceDatadogLogsArchiveDelete,
		Exists: resourceDatadogLogsArchiveExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name":  {Type: schema.TypeString, Required: true},
			"query": {Type: schema.TypeString, Required: true},
			"s3": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket":       {Type: schema.TypeString, Required: true},
						"path":         {Type: schema.TypeString, Required: true},
						"client_email": {Type: schema.TypeString, Required: true},
						"project_id":   {Type: schema.TypeString, Required: true},
						"account_id":   {Type: schema.TypeString, Required: true},
						"role_name":    {Type: schema.TypeString, Required: true},
					},
				},
			},
			"azure": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container":       {Type: schema.TypeString, Required: true},
						"client_id":       {Type: schema.TypeString, Required: true},
						"tenant_id":       {Type: schema.TypeString, Required: true},
						"storage_account": {Type: schema.TypeString, Required: true},
						"path":            {Type: schema.TypeString, Optional: true},
					},
				},
			},
			"gcs": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket":       {Type: schema.TypeString, Required: true},
						"path":         {Type: schema.TypeString, Required: true},
						"client_email": {Type: schema.TypeString, Required: true},
						"project_id":   {Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	}
}

func resourceDatadogLogsArchiveCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ddArchive, err := buildDatadogArchiveCreateReq(d)
	if err != nil {
		return err
	}
	createdArchive, _, err := datadogClientV2.LogsArchivesApi.CreateLogsArchive(authV2).Body(ddArchive).Execute()
	if err != nil {
		return translateClientError(err, "failed to create logs archive using Datadog API")
	}
	d.SetId(*createdArchive.GetData().Id)
	return resourceDatadogLogsArchiveRead(d, meta)
}

func resourceDatadogLogsArchiveRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ddArchive, _, err := datadogClientV2.LogsArchivesApi.GetLogsArchive(authV2, d.Id()).Execute()
	if err != nil {
		return translateClientError(err, "failed to get logs archive using Datadog API")
	}
	if err = d.Set("name", ddArchive.Data.Attributes.Name); err != nil {
		return err
	}
	if err = d.Set("query", ddArchive.Data.Attributes.Query); err != nil {
		return err
	}
	archiveType, destination, err := buildDestination(ddArchive.Data.Attributes.Destination)
	if err != nil {
		return err
	}
	if err := d.Set(archiveType, destination); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsArchiveUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	ddArchive, err := buildDatadogArchiveCreateReq(d)
	if err != nil {
		return err
	}
	if _, _, err := datadogClientV2.LogsArchivesApi.UpdateLogsArchive(authV2, d.Id()).Body(ddArchive).Execute(); err != nil {
		return translateClientError(err, "error updating logs archive")
	}
	return resourceDatadogLogsArchiveRead(d, meta)
}

func resourceDatadogLogsArchiveDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if httpresp, err := datadogClientV2.LogsArchivesApi.DeleteLogsArchive(authV2, d.Id()).Execute(); err != nil {
		// API returns 404 when the specific archive id doesn't exist.
		if httpresp != nil && httpresp.StatusCode == 404 {
			return nil
		}
		return translateClientError(err, "error deleting logs archive")
	}
	return nil
}

func resourceDatadogLogsArchiveExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	ddArchive, httpresp, err := datadogClientV2.LogsArchivesApi.GetLogsArchive(authV2, d.Id()).Execute()
	if err != nil {
		// API returns 404 when the specific archive id doesn't exist.
		if httpresp != nil && httpresp.StatusCode == 404 {
			return false, nil
		}
		return false, translateClientError(err, "error checking if logs archive exists")
	}
	return ddArchive.HasData(), nil
}

//Model to map
func buildDestination(archiveDestination datadogV2.NullableLogsArchiveDestination) (string, map[string]interface{}, error) {
	emptyDestination := map[string]interface{}{}
	if archiveDestination.IsSet() {
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
	return "", emptyDestination, fmt.Errorf("Destination should be not null.")
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

//Map to model
func buildDatadogArchiveCreateReq(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequest, error) {
	archive := datadogV2.NewLogsArchiveCreateRequest()
	destination, err := buildCreateReqDestination(d)
	if err != nil {
		return *archive, err
	}
	attributes := datadogV2.NewLogsArchiveCreateRequestAttributes(
		destination,
		d.Get("name").(string),
		d.Get("query").(string),
	)
	definition := datadogV2.NewLogsArchiveCreateRequestDefinitionWithDefaults()
	definition.SetAttributes(*attributes)
	archive.SetData(*definition)
	return *archive, nil
}

func buildCreateReqDestination(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error) {
	emptyDestination := datadogV2.LogsArchiveCreateRequestDestination{}
	defDestinations := definedDestinations(d)
	if len(defDestinations) != 1 {
		return emptyDestination, fmt.Errorf("More than one type defined: %v", defDestinations)
	}
	archiveType := defDestinations[0]
	destinationMap := d.Get(archiveType).(map[string]interface{})
	switch archiveType {
	case string(datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE):
		if destination, err := buildAzureDestination(destinationMap); err != nil {
			return datadogV2.LogsArchiveCreateRequestDestination{}, err
		} else {
			return datadogV2.LogsArchiveDestinationAzureAsLogsArchiveCreateRequestDestination(destination), nil
		}
	case string(datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS):
		if destination, err := buildGCSDestination(destinationMap); err != nil {
			return datadogV2.LogsArchiveCreateRequestDestination{}, err
		} else {
			return datadogV2.LogsArchiveDestinationGCSAsLogsArchiveCreateRequestDestination(destination), nil
		}
	case string(datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3):
		if destination, err := buildS3Destination(destinationMap); err != nil {
			return datadogV2.LogsArchiveCreateRequestDestination{}, err
		} else {
			return datadogV2.LogsArchiveDestinationS3AsLogsArchiveCreateRequestDestination(destination), nil
		}
	default:
		return datadogV2.LogsArchiveCreateRequestDestination{}, fmt.Errorf("Archive type '%s' doesn't exist", archiveType)
	}
}

func definedDestinations(d *schema.ResourceData) []string {
	defined := []string{}
	for _, destination := range []string{"azure", "gcs", "s3"} {
		if _, ok := d.GetOk(destination); ok {
			defined = append(defined, destination)
		}
	}
	return defined
}

func buildAzureDestination(d map[string]interface{}) (*datadogV2.LogsArchiveDestinationAzure, error) {
	clientId, ok := d["client_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationAzure{}, fmt.Errorf("client_id is not defined")
	}
	tenantId, ok := d["tenant_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationAzure{}, fmt.Errorf("tenant_id is not defined")
	}
	integration := datadogV2.LogsArchiveIntegrationAzure{
		ClientId: clientId.(string),
		TenantId: tenantId.(string),
	}
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
	destination := &datadogV2.LogsArchiveDestinationAzure{
		Container:      container.(string),
		Integration:    integration,
		StorageAccount: storageAccount.(string),
		Type:           datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE,
		Path:           datadogV2.PtrString(path.(string)),
	}
	return destination, nil
}

func buildGCSDestination(d map[string]interface{}) (*datadogV2.LogsArchiveDestinationGCS, error) {
	clientEmail, ok := d["client_email"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationGCS{}, fmt.Errorf("client_email is not defined")
	}
	projectId, ok := d["project_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationGCS{}, fmt.Errorf("project_id is not defined")
	}
	integration := datadogV2.LogsArchiveIntegrationGCS{
		ClientEmail: clientEmail.(string),
		ProjectId:   projectId.(string),
	}
	bucket, ok := d["bucket"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationGCS{}, fmt.Errorf("bucket is not defined")
	}
	path, ok := d["path"]
	if !ok {
		path = ""
	}
	destination := &datadogV2.LogsArchiveDestinationGCS{
		Bucket:      bucket.(string),
		Integration: integration,
		Path:        datadogV2.PtrString(path.(string)),
		Type:        datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS,
	}
	return destination, nil
}

func buildS3Destination(d map[string]interface{}) (*datadogV2.LogsArchiveDestinationS3, error) {
	accountId, ok := d["account_id"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationS3{}, fmt.Errorf("account_id is not defined")
	}
	roleName, ok := d["role_name"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationS3{}, fmt.Errorf("role_name is not defined")
	}
	integration := datadogV2.LogsArchiveIntegrationS3{
		AccountId: accountId.(string),
		RoleName:  roleName.(string),
	}
	bucket, ok := d["bucket"]
	if !ok {
		return &datadogV2.LogsArchiveDestinationS3{}, fmt.Errorf("bucket is not defined")
	}
	path, ok := d["path"]
	if !ok {
		path = ""
	}
	destination := &datadogV2.LogsArchiveDestinationS3{
		Bucket:      bucket.(string),
		Integration: integration,
		Path:        datadogV2.PtrString(path.(string)),
		Type:        datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3,
	}
	return destination, nil
}
