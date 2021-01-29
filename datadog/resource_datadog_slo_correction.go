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
				Optional:    true,
				Description: "The timezone to display in the UI for the correction times (defaults to \"UTC\")",
			},
		},
	}
}

func buildDatadogSloCorrection(d *schema.ResourceData) (*datadogV1.SLOCorrectionCreateRequest, error) {
	result := datadogV1.NewSLOCorrectionCreateRequestWithDefaults()
	// `type` is hardcoded to 'correction' in Data
	// only need to set `attributes` here
	createData := datadogV1.NewSLOCorrectionCreateRequestDataWithDefaults()
	attributes := datadogV1.NewSLOCorrectionCreateRequestAttributesWithDefaults()
	correctionCategory := datadogV1.SLOCorrectionCategory(d.Get("category").(string))
	attributes.SetCategory(correctionCategory)
	attributes.SetStart(int64(d.Get("start").(int)))
	attributes.SetEnd(int64(d.Get("end").(int)))
	attributes.SetSloId(d.Get("slo_id").(string))

	if timezone, ok := d.GetOk("timezone"); ok {
		attributes.SetTimezone(timezone.(string))
	}

	if description, ok := d.GetOk("description"); ok {
		attributes.SetDescription(description.(string))
	}
	createData.SetAttributes(*attributes)
	result.SetData(*createData)
	return result, nil
}

func buildDatadogSloCorrectionUpdate(d *schema.ResourceData) (*datadogV1.SLOCorrectionUpdateRequest, error) {
	result := datadogV1.NewSLOCorrectionUpdateRequestWithDefaults()
	updateData := datadogV1.NewSLOCorrectionUpdateRequestDataWithDefaults()
	attributes := datadogV1.NewSLOCorrectionUpdateRequestAttributesWithDefaults()
	if description, ok := d.GetOk("description"); ok {
		attributes.SetDescription(description.(string))
	}
	if timezone, ok := d.GetOk("timezone"); ok {
		attributes.SetTimezone(timezone.(string))
	}
	if start, ok := d.GetOk("start"); ok {
		attributes.SetStart(int64(start.(int)))
	}
	if end, ok := d.GetOk("end"); ok {
		attributes.SetEnd(int64(end.(int)))
	}
	if category, ok := d.GetOk("category"); ok {
		attributes.SetCategory(datadogV1.SLOCorrectionCategory(category.(string)))
	}
	updateData.SetAttributes(*attributes)
	result.SetData(*updateData)
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
	sloCorrection := response.GetData()
	d.SetId(sloCorrection.GetId())

	return resourceDatadogSloCorrectionRead(d, meta)
}

func resourceDatadogSloCorrectionRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id := d.Id()

	sloCorrectionGetResp, httpResp, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id).Execute()
	if err != nil {
		if httpResp.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error reading SloCorrection")
	}
	sloCorrectionGetData := sloCorrectionGetResp.GetData()
	if sloCorrectionAttributes, ok := sloCorrectionGetData.GetAttributesOk(); ok {
		if category, ok := sloCorrectionAttributes.GetCategoryOk(); ok {
			d.Set("category", string(*category))
		}
		if description, ok := sloCorrectionAttributes.GetDescriptionOk(); ok {
			d.Set("description", *description)
		}
		if sloID, ok := sloCorrectionAttributes.GetSloIdOk(); ok {
			d.Set("slo_id", *sloID)
		}
		if timezone, ok := sloCorrectionAttributes.GetTimezoneOk(); ok {
			d.Set("timezone", *timezone)
		}
		if start, ok := sloCorrectionAttributes.GetStartOk(); ok {
			d.Set("start", *start)
		}
		if end, ok := sloCorrectionAttributes.GetEndOk(); ok {
			d.Set("end", *end)
		}
	}
	return nil
}

func resourceDatadogSloCorrectionUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject, err := buildDatadogSloCorrectionUpdate(d)
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
