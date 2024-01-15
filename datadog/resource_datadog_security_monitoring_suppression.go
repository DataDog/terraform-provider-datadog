package datadog

import (
	"context"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogSecurityMonitoringSuppression() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Security Monitoring Suppression API resource. It can be used to create and manage Datadog security monitoring suppression rules.",
		CreateContext: resourceDatadogSecurityMonitoringSuppressionCreate,
		ReadContext:   resourceDatadogSecurityMonitoringSuppressionRead,
		UpdateContext: resourceDatadogSecurityMonitoringSuppressionUpdate,
		DeleteContext: resourceDatadogSecurityMonitoringSuppressionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: datadogSecurityMonitoringSuppressionSchema,
	}
}

func datadogSecurityMonitoringSuppressionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the suppression rule.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A description for the suppression rule.",
		},
		"enabled": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Whether the suppression rule is enabled.",
		},
		"expiration_date": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A RFC3339 timestamp giving an expiration date for the suppression rule. After this date, it won't suppress signals anymore.",
		},
		"rule_query": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The rule query of the suppression rule, with the same syntax as the search bar for detection rules.",
		},
		"suppression_query": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The suppression query of the suppression rule. If a signal matches this query, it is suppressed and is not triggered. Same syntax as the queries to search signals in the signal explorer.",
		},
	}
}

func resourceDatadogSecurityMonitoringSuppressionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	suppressionPayload, err := buildCreateSecurityMonitoringSuppressionPayload(d)

	if err != nil {
		return diag.FromErr(err)
	}

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().CreateSecurityMonitoringSuppression(auth, *suppressionPayload)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating security monitoring suppression")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateResourceDataFromSuppressionResponse(d, &response)
}

func resourceDatadogSecurityMonitoringSuppressionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	suppressionId := d.Id()

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringSuppression(auth, suppressionId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error fetching security monitoring suppression")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateResourceDataFromSuppressionResponse(d, &response)
}

func resourceDatadogSecurityMonitoringSuppressionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	suppressionId := d.Id()
	suppressionPayload, err := buildUpdateSecurityMonitoringSuppressionPayload(d)

	if err != nil {
		return diag.FromErr(err)
	}

	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().UpdateSecurityMonitoringSuppression(auth, suppressionId, *suppressionPayload)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring suppression")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	return updateResourceDataFromSuppressionResponse(d, &response)
}

func resourceDatadogSecurityMonitoringSuppressionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	suppressionId := d.Id()

	if httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().DeleteSecurityMonitoringSuppression(auth, suppressionId); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting security monitoring suppression")
	}

	return nil
}

func buildCreateSecurityMonitoringSuppressionPayload(d *schema.ResourceData) (*datadogV2.SecurityMonitoringSuppressionCreateRequest, error) {
	name, description, enabled, expirationDate, ruleQuery, suppressionQuery, err := extractSuppressionAttributesFromResource(d)

	if err != nil {
		return nil, err
	}

	attributes := datadogV2.NewSecurityMonitoringSuppressionCreateAttributes(enabled, name, ruleQuery, suppressionQuery)
	attributes.Description = description
	attributes.ExpirationDate = expirationDate

	data := datadogV2.NewSecurityMonitoringSuppressionCreateData(*attributes, datadogV2.SECURITYMONITORINGSUPPRESSIONTYPE_SUPPRESSIONS)
	return datadogV2.NewSecurityMonitoringSuppressionCreateRequest(*data), nil
}

func buildUpdateSecurityMonitoringSuppressionPayload(d *schema.ResourceData) (*datadogV2.SecurityMonitoringSuppressionUpdateRequest, error) {
	name, description, enabled, expirationDate, ruleQuery, suppressionQuery, err := extractSuppressionAttributesFromResource(d)

	if err != nil {
		return nil, err
	}

	attributes := datadogV2.NewSecurityMonitoringSuppressionUpdateAttributes()
	attributes.SetName(name)
	attributes.Description = description
	attributes.SetEnabled(enabled)
	// Expiration date needs to be set via AdditionalProperties because it needs to be explicitly
	// set to null if it's not in the Terraform definition.
	// If omitted, the API leaves the expiration date unchanged instead of removing it
	// The ExpirationDate field of SecurityMonitoringSuppressionUpdateAttributes has the omitempty tag, so if it is nil,
	// it is omitted from the JSON payload.
	attributes.AdditionalProperties = map[string]interface{}{"expiration_date": expirationDate}
	attributes.SetRuleQuery(ruleQuery)
	attributes.SetSuppressionQuery(suppressionQuery)

	data := datadogV2.NewSecurityMonitoringSuppressionUpdateData(*attributes, datadogV2.SECURITYMONITORINGSUPPRESSIONTYPE_SUPPRESSIONS)
	return datadogV2.NewSecurityMonitoringSuppressionUpdateRequest(*data), nil
}

func extractSuppressionAttributesFromResource(d *schema.ResourceData) (string, *string, bool, *int64, string, string, error) {
	// Mandatory fields

	name := d.Get("name").(string)
	enabled := d.Get("enabled").(bool)
	ruleQuery := d.Get("rule_query").(string)
	suppressionQuery := d.Get("suppression_query").(string)

	// Optional fields

	var description *string
	var expirationDate *int64

	if tfDescription, ok := d.GetOk("description"); ok {
		strDescription := tfDescription.(string)
		description = &strDescription
	}

	if tfExpirationDate, ok := d.GetOk("expiration_date"); ok {
		expirationDateTime, err := time.Parse(time.RFC3339, tfExpirationDate.(string))

		if err != nil {
			return "", nil, false, nil, "", "", err
		}

		expirationDateTimestamp := expirationDateTime.UnixMilli()
		expirationDate = &expirationDateTimestamp

	}

	return name, description, enabled, expirationDate, ruleQuery, suppressionQuery, nil
}

func updateResourceDataFromSuppressionResponse(d *schema.ResourceData, res *datadogV2.SecurityMonitoringSuppressionResponse) diag.Diagnostics {
	d.SetId(res.Data.GetId())

	attributes := res.Data.Attributes

	var diags diag.Diagnostics

	if err := d.Set("name", attributes.GetName()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("description", attributes.GetDescription()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("enabled", attributes.GetEnabled()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if attributes.ExpirationDate != nil {
		responseExpirationDate := time.UnixMilli(*attributes.ExpirationDate).UTC()
		expirationDate := responseExpirationDate.Format(time.RFC3339)

		userExpirationDateStr, ok := d.GetOk("expiration_date")
		if ok {
			if userExpirationDate, err := time.Parse(time.RFC3339, userExpirationDateStr.(string)); err == nil {
				// The API only requires a millisecond timestamp, it does not care about timezones.
				// If the timestamp string written by the user has the same millisecond value as the one returned by the API,
				// we keep the user-defined one in the state.
				if userExpirationDate.UnixMilli() == responseExpirationDate.UnixMilli() {
					expirationDate = userExpirationDateStr.(string)
				}
			} else {
				diags = append(diags, diag.FromErr(err)...)
			}
		}

		if err := d.Set("expiration_date", expirationDate); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	} else {
		if err := d.Set("expiration_date", nil); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	if err := d.Set("rule_query", attributes.GetRuleQuery()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if err := d.Set("suppression_query", attributes.GetSuppressionQuery()); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}
