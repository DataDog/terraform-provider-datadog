package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogSloCorrection() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with the slo_correction API.",
		CreateContext: resourceDatadogSloCorrectionCreate,
		ReadContext:   resourceDatadogSloCorrectionRead,
		UpdateContext: resourceDatadogSloCorrectionUpdate,
		DeleteContext: resourceDatadogSloCorrectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"category": {
				Type:             schema.TypeString,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOCorrectionCategoryFromValue),
				Required:         true,
				Description:      "Category the SLO correction belongs to.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the correction being made.",
			},
			"end": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Ending time of the correction in epoch seconds.",
			},
			"slo_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the SLO that this correction will be applied to.",
			},
			"start": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Starting time of the correction in epoch seconds.",
			},
			"timezone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The timezone to display in the UI for the correction times (defaults to \"UTC\")",
			},
		},
	}
}

func buildDatadogSloCorrection(d *schema.ResourceData) *datadogV1.SLOCorrectionCreateRequest {
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
	return result
}

func buildDatadogSloCorrectionUpdate(d *schema.ResourceData) *datadogV1.SLOCorrectionUpdateRequest {
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
	return result
}

func resourceDatadogSloCorrectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject := buildDatadogSloCorrection(d)

	response, httpResponse, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.CreateSLOCorrection(auth, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating SloCorrection")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}
	sloCorrection := response.GetData()
	d.SetId(sloCorrection.GetId())

	return updateSLOCorrectionState(d, response.Data)
}

func updateSLOCorrectionState(d *schema.ResourceData, sloCorrectionData *datadogV1.SLOCorrection) diag.Diagnostics {
	if sloCorrectionAttributes, ok := sloCorrectionData.GetAttributesOk(); ok {
		if category, ok := sloCorrectionAttributes.GetCategoryOk(); ok {
			if err := d.Set("category", string(*category)); err != nil {
				return diag.FromErr(err)
			}
		}
		if description, ok := sloCorrectionAttributes.GetDescriptionOk(); ok {
			if err := d.Set("description", *description); err != nil {
				return diag.FromErr(err)
			}
		}
		if sloID, ok := sloCorrectionAttributes.GetSloIdOk(); ok {
			if err := d.Set("slo_id", *sloID); err != nil {
				return diag.FromErr(err)
			}
		}
		if timezone, ok := sloCorrectionAttributes.GetTimezoneOk(); ok {
			if err := d.Set("timezone", *timezone); err != nil {
				return diag.FromErr(err)
			}
		}
		if start, ok := sloCorrectionAttributes.GetStartOk(); ok {
			if err := d.Set("start", *start); err != nil {
				return diag.FromErr(err)
			}
		}
		if end, ok := sloCorrectionAttributes.GetEndOk(); ok {
			if err := d.Set("end", *end); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceDatadogSloCorrectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id := d.Id()

	sloCorrectionGetResp, httpResponse, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.GetSLOCorrection(auth, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			// this condition takes on the job of the deprecated Exists handlers
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error reading SloCorrection")
	}
	if err := utils.CheckForUnparsed(sloCorrectionGetResp); err != nil {
		return diag.FromErr(err)
	}
	return updateSLOCorrectionState(d, sloCorrectionGetResp.Data)
}

func resourceDatadogSloCorrectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject := buildDatadogSloCorrectionUpdate(d)
	id := d.Id()

	response, httpResponse, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.UpdateSLOCorrection(auth, id, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating SloCorrection")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateSLOCorrectionState(d, response.Data)
}

func resourceDatadogSloCorrectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1
	var err error

	id := d.Id()

	httpResponse, err := datadogClient.ServiceLevelObjectiveCorrectionsApi.DeleteSLOCorrection(auth, id)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting SloCorrection")
	}

	return nil
}
