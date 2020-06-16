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
			State: schema.ImportStatePassthrough, //FIXME: hein ?
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

	ddArchive, err := buildDatadogArchive(d)
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
	return nil
}

func resourceDatadogLogsArchiveUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceDatadogLogsArchiveRead(d, meta)
}

func resourceDatadogLogsArchiveDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogLogsArchiveExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2
	ddArchive, httpresp, err := datadogClientV2.LogsArchivesApi.GetLogsArchive(authV2, d.Id()).Execute()
	if err != nil {
		// API returns 404 when the specific archive id doesn't exist.
		if httpresp.StatusCode == 404 {
			return false, nil
		}
		return false, translateClientError(err, "error checking if logs archive exists")
	}
	return ddArchive.HasData(), nil
}

func buildDatadogArchive(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequest, error) {
	archive := datadogV2.NewLogsArchiveCreateRequest()
	destination, err := buildDestination(d)
	if err != nil {
		return *archive, err
	}
	attributes := datadogV2.NewLogsArchiveCreateRequestAttributes(
		destination,
		"name",  //FIXME
		"query", //FIXME
	)
	definition := datadogV2.NewLogsArchiveCreateRequestDefinitionWithDefaults()
	definition.SetAttributes(*attributes)
	archive.SetData(*definition)
	return *archive, nil
}

func buildDestination(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error) {
	emptyDestination := datadogV2.LogsArchiveCreateRequestDestination{}
	defDestinations := definedDestinations(d)
	if len(defDestinations) != 1 {
		return emptyDestination, fmt.Errorf("Invalid archive definition: (Defined destinantions:%v)", defDestinations)
	}
	archiveType := defDestinations[0]
	return buildDestinationByType(archiveType, d)
}

func buildDestinationByType(archiveType string, d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error) {
	emptyDestination := datadogV2.LogsArchiveCreateRequestDestination{}
	if buildFunction, exists := buildDestinationByTypeFunctions[archiveType]; exists {
		return buildFunction(d)
	}
	return emptyDestination, fmt.Errorf("Archive type '%s' doesn't exist", archiveType)
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

var buildDestinationByTypeFunctions = map[string]func(*schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error){
	string(datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE): func(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error) {
		integration := datadogV2.LogsArchiveIntegrationAzure{
			ClientId: "clientId", //FIXME
			TenantId: "tenantId", //FIXME
		}
		destination := &datadogV2.LogsArchiveDestinationAzure{
			Container:      "container", //FIXME
			Integration:    integration,
			StorageAccount: "storageAccount", //FIXME
			Type:           datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE,
		}
		return datadogV2.LogsArchiveDestinationAzureAsLogsArchiveCreateRequestDestination(destination), nil
	},
	string(datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS): func(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error) {
		integration := datadogV2.LogsArchiveIntegrationGCS{
			ClientEmail: "clientEmail", //FIXME
			ProjectId:   "projectId",   //FIXME
		}
		destination := &datadogV2.LogsArchiveDestinationGCS{
			Bucket:      "bucket", //FIXME
			Integration: integration,
			Path:        datadogV2.PtrString("path"), //FIXME
			Type:        datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS,
		}
		return datadogV2.LogsArchiveDestinationGCSAsLogsArchiveCreateRequestDestination(destination), nil
	},
	string(datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3): func(d *schema.ResourceData) (datadogV2.LogsArchiveCreateRequestDestination, error) {
		integration := datadogV2.LogsArchiveIntegrationS3{
			AccountId: "accountId", //FIXME
			RoleName:  "roleName",  //FIXME
		}
		destination := &datadogV2.LogsArchiveDestinationS3{
			Bucket:      "bucket", //FIXME
			Integration: integration,
			Path:        datadogV2.PtrString("path"), //FIXME
			Type:        datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3,
		}
		return datadogV2.LogsArchiveDestinationS3AsLogsArchiveCreateRequestDestination(destination), nil
	},
}
