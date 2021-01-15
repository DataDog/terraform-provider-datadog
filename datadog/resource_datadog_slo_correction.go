package datadog

import (
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogSloCorrection() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for interacting with the slo_correction API",
		Create:      resourceDatadogSloCorrectionCreate,
		Read:        resourceDatadogSloCorrectionRead,
		Update:      resourceDatadogSloCorrectionUpdate,
		Delete:      resourceDatadogSloCorrectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogSloCorrectionImport,
		},
		Schema: map[string]*schema.Schema{
			"category": {
				Type:         schema.TypeString,
				ValidateFunc: validateEnumValue(datadogV1.NewSLOCorrectionCategoryFromValue),
				Required:     true,
				Description:  "Category the SLO correction belongs to",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the correction being made.",
			},
			"end": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Ending time of the correction in epoch seconds",
			},
			"slo_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the SLO that this correction will be applied to",
			},
			"start": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Starting time of the correction in epoch seconds",
			},
			"timezone": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Timezone of the timestamps provided",
			},
		},
	}
}

func buildDatadogSloCorrection(d *schema.ResourceData) (*datadogV1.SLOCorrectionCreateRequest, error) {
	k := NewResourceDataKey(d, "")
	result := datadogV1.NewSLOCorrectionCreateRequestWithDefaults()
	return result, nil
}

func resourceDatadogSloCorrectionCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject, err := buildDatadogSloCorrection(d)

	response, _, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.CreateSLOCorrection(auth).Body(*ddObject).Execute()
	if err != nil {
		return translateClientError(err, "error creating SloCorrection")
	}
	// FIXME: no property found that looks like an Id for model SloCorrection
	// you need to manually add code that would call `d.SetId(<the-actual-id>)` to store
	// the Id in the state properly

	return resourceDatadogSloCorrectionRead(d, meta)
}

func resourceDatadogSloCorrectionRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id := d.Id()

	resource, httpResp, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id).Execute()

	if err != nil {
		if httpResp.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error reading SloCorrection")
	}

	return nil
}

func resourceDatadogSloCorrectionUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject, err := buildDatadogSloCorrection(d)
	id := d.Id()

	_, _, err = datadogClient.ServiceLevelObjectiveCorrectionsApi.UpdateSLOCorrection(auth, id).Body(*ddObject).Execute()
	if err != nil {
		return translateClientError(err, "error creating SloCorrection")
	}

	return resourceDatadogSloCorrectionRead(d, meta)
}

func resourceDatadogSloCorrectionDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id := d.Id()

	_, err = datadogClient.ServiceLevelObjectiveCorrectionsApi.DeleteSLOCorrection(auth, id).Execute()

	if err != nil {
		return translateClientError(err, "error deleting SloCorrection")
	}

	return nil
}

func resourceDatadogSloCorrectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogSloCorrectionRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
