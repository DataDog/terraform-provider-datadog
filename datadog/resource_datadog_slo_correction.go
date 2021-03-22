package datadog

import (
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

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
				ValidateFunc: validators.ValidateEnumValue(datadogV1.NewSLOCorrectionCategoryFromValue),
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
	createData := datadogV1.NewSLOCorrectionCreateData()
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
	updateData := datadogV1.NewSLOCorrectionUpdateData()
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
		return utils.TranslateClientError(err, "error creating SloCorrection")
	}
	sloCorrection := response.GetData()
	d.SetId(sloCorrection.GetId())

	return updateSLOCorrectionState(d, response.Data)
}

func updateSLOCorrectionState(d *schema.ResourceData, sloCorrectionData *datadogV1.SLOCorrection) error {
	if sloCorrectionAttributes, ok := sloCorrectionData.GetAttributesOk(); ok {
		if category, ok := sloCorrectionAttributes.GetCategoryOk(); ok {
			if err := d.Set("category", string(*category)); err != nil {
				return err
			}
		}
		if description, ok := sloCorrectionAttributes.GetDescriptionOk(); ok {
			if err := d.Set("description", *description); err != nil {
				return err
			}
		}
		if sloID, ok := sloCorrectionAttributes.GetSloIdOk(); ok {
			if err := d.Set("slo_id", *sloID); err != nil {
				return err
			}
		}
		if timezone, ok := sloCorrectionAttributes.GetTimezoneOk(); ok {
			if err := d.Set("timezone", *timezone); err != nil {
				return err
			}
		}
		if start, ok := sloCorrectionAttributes.GetStartOk(); ok {
			if err := d.Set("start", *start); err != nil {
				return err
			}
		}
		if end, ok := sloCorrectionAttributes.GetEndOk(); ok {
			if err := d.Set("end", *end); err != nil {
				return err
			}
		}
	}
	return nil
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
		return utils.TranslateClientError(err, "error reading SloCorrection")
	}
	return updateSLOCorrectionState(d, sloCorrectionGetResp.Data)
}

func resourceDatadogSloCorrectionUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject, err := buildDatadogSloCorrectionUpdate(d)
	id := d.Id()

	response, _, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.UpdateSLOCorrection(auth, id).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating SloCorrection")
	}

	return updateSLOCorrectionState(d, response.Data)
}

func resourceDatadogSloCorrectionDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id := d.Id()

	_, err = datadogClient.ServiceLevelObjectiveCorrectionsApi.DeleteSLOCorrection(auth, id).Execute()

	if err != nil {
		return utils.TranslateClientError(err, "error deleting SloCorrection")
	}

	return nil
}

func resourceDatadogSloCorrectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogSloCorrectionRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
