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
						"path":            {Type: schema.TypeString, Optional: true},
						"region":          {Type: schema.TypeString, Optional: true}, //FIXME: should it be removed because it is set by mcnulty ?
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
	return "", emptyDestination, fmt.Errorf("Destination is not set. ") //FIXME: what to do if the destination is not returned
}

func buildAzureMap(destination datadogV2.LogsArchiveDestinationAzure) (map[string]interface{}) {
	result := make(map[string]interface{})
	result["client_id"] = destination.Integration.ClientId
	result["tenant_id"] = destination.Integration.TenantId
	result["container"] = destination.Container
	result["storage_account"] = destination.StorageAccount
	result["region"] = destination.Region
	result["path"] = destination.Path
	return result
}

func buildGCSMap(destination datadogV2.LogsArchiveDestinationGCS) (map[string]interface{}) {
	result := make(map[string]interface{})
	return result
}

func buildS3Map(destination datadogV2.LogsArchiveDestinationS3) (map[string]interface{}) {
	result := make(map[string]interface{})
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
	if buildFunction, exists := buildCreateReqDestinationByTypeFunctions[archiveType]; exists {
		return buildFunction(d.Get(archiveType).(map[string]interface{}))
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

var buildCreateReqDestinationByTypeFunctions = map[string]func(map[string]interface{}) (datadogV2.LogsArchiveCreateRequestDestination, error){
	string(datadogV2.LOGSARCHIVEDESTINATIONAZURETYPE_AZURE): func(d map[string]interface{}) (datadogV2.LogsArchiveCreateRequestDestination, error) {
		if destination, err := buildAzureDestination(d); err != nil {
			return datadogV2.LogsArchiveCreateRequestDestination{}, nil
		} else {
			return datadogV2.LogsArchiveDestinationAzureAsLogsArchiveCreateRequestDestination(destination), nil
		}
	},
	string(datadogV2.LOGSARCHIVEDESTINATIONGCSTYPE_GCS): func(d map[string]interface{}) (datadogV2.LogsArchiveCreateRequestDestination, error) {
		if destination, err := buildGCSDestination(d); err != nil {
			return datadogV2.LogsArchiveCreateRequestDestination{}, nil
		} else {
			return datadogV2.LogsArchiveDestinationGCSAsLogsArchiveCreateRequestDestination(destination), nil
		}
	},
	string(datadogV2.LOGSARCHIVEDESTINATIONS3TYPE_S3): func(d map[string]interface{}) (datadogV2.LogsArchiveCreateRequestDestination, error) {
		if destination, err := buildS3Destination(d); err != nil {
			return datadogV2.LogsArchiveCreateRequestDestination{}, nil
		} else {
			return datadogV2.LogsArchiveDestinationS3AsLogsArchiveCreateRequestDestination(destination), nil
		}
	},
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
		region, ok := d["region"]
		if !ok {
			region = ""
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
			Region:         datadogV2.PtrString(region.(string)),
			Path:           datadogV2.PtrString(path.(string)),
		}
		return destination, nil
}

func buildGCSDestination(d map[string]interface{}) (*datadogV2.LogsArchiveDestinationGCS, error) {
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
	return destination, nil
}

func buildS3Destination(d map[string]interface{}) (*datadogV2.LogsArchiveDestinationS3, error) {
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
	return destination, nil
}
