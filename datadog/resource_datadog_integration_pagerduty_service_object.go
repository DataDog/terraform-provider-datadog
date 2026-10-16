package datadog

import (
	"context"
	"fmt"
	"os"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const maskedSecret = "*****"

func resourceDatadogIntegrationPagerdutySO() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides access to individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.",
		CreateContext: resourceDatadogIntegrationPagerdutySOCreate,
		ReadContext:   resourceDatadogIntegrationPagerdutySORead,
		UpdateContext: resourceDatadogIntegrationPagerdutySOUpdate,
		DeleteContext: resourceDatadogIntegrationPagerdutySODelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatadogIntegrationPagerdutySOImport,
		},
		ValidateRawResourceConfigFuncs: []schema.ValidateRawResourceConfigFunc{
			validation.PreferWriteOnlyAttribute(cty.GetAttrPath("service_key"), cty.GetAttrPath("service_key_wo")),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"service_name": {
					Description: "Your Service name in PagerDuty.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"service_key": {
					Description:  "Your Service name associated service key in PagerDuty. This key may also be referred to as an Integration Key or Routing Key in the Pagerduty Integration [documentation](https://www.pagerduty.com/docs/guides/datadog-integration-guide/), UI, and within the [Pagerduty Provider for Terraform](https://registry.terraform.io/providers/PagerDuty/pagerduty/latest/docs/resources/service_integration#integration_key) Note: Since the Datadog API never returns service keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html) to have it destroyed and recreated. Exactly one of `service_key` or `service_key_wo` must be set.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					ExactlyOneOf: []string{"service_key", "service_key_wo"},
				},
				"service_key_wo": {
					Description:  "Write-only service key for the PagerDuty service. The value is not stored in the Terraform state. Exactly one of `service_key` or `service_key_wo` must be set. Must be used with `service_key_wo_version`.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					WriteOnly:    true,
					ExactlyOneOf: []string{"service_key", "service_key_wo"},
					RequiredWith: []string{"service_key_wo_version"},
				},
				"service_key_wo_version": {
					Description:  "Version for `service_key_wo` rotation. Changing this triggers an update.",
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringIsNotEmpty,
					RequiredWith: []string{"service_key_wo"},
				},
			}
		},
	}
}

// getIntegrationPagerdutyServiceKey returns the service key from the write-only
// attribute when it is set, and from the plaintext attribute otherwise.
func getIntegrationPagerdutyServiceKey(d *schema.ResourceData) (string, diag.Diagnostics) {
	serviceKeyWo, diags := d.GetRawConfigAt(cty.GetAttrPath("service_key_wo"))
	if diags.HasError() {
		return "", diags
	}
	if serviceKeyWo.IsKnown() && !serviceKeyWo.IsNull() {
		return serviceKeyWo.AsString(), nil
	}
	return d.Get("service_key").(string), nil
}

func buildIntegrationPagerdutySO(d *schema.ResourceData, serviceKey string) *datadogV1.PagerDutyService {
	so := &datadogV1.PagerDutyService{}
	if v, ok := d.GetOk("service_name"); ok {
		so.SetServiceName(v.(string))
	}
	if serviceKey != "" {
		so.SetServiceKey(serviceKey)
	}

	return so
}

func buildIntegrationPagerdutyServiceKey(serviceKey string) *datadogV1.PagerDutyServiceKey {
	pdServiceKey := &datadogV1.PagerDutyServiceKey{}
	if serviceKey != "" {
		pdServiceKey.SetServiceKey(serviceKey)
	}

	return pdServiceKey
}

func resourceDatadogIntegrationPagerdutySOCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	serviceKey, diags := getIntegrationPagerdutyServiceKey(d)
	if diags.HasError() {
		return diags
	}
	so := buildIntegrationPagerdutySO(d, serviceKey)
	if _, httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().CreatePagerDutyIntegrationService(auth, *so); err != nil {
		// TODO: warn user that PD integration must be enabled to be able to create service objects
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating PagerDuty integration service")
	}
	d.SetId(so.GetServiceName())

	return resourceDatadogIntegrationPagerdutySORead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutySORead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	so, httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().GetPagerDutyIntegrationService(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting PagerDuty integration service")
	}
	if err := utils.CheckForUnparsed(so); err != nil {
		return diag.FromErr(err)
	}

	d.Set("service_name", so.GetServiceName())
	// Only update service_key if not set on d - the API endpoints never return
	// the keys, so this is how we recognize new values.
	// When service_key_wo is used, service_key must stay empty to match the configuration.
	_, usesServiceKey := d.GetOk("service_key")
	_, usesServiceKeyWo := d.GetOk("service_key_wo_version")
	if !usesServiceKey && !usesServiceKeyWo {
		d.Set("service_key", maskedSecret)
	}

	return nil
}

func resourceDatadogIntegrationPagerdutySOUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	serviceKey, diags := getIntegrationPagerdutyServiceKey(d)
	if diags.HasError() {
		return diags
	}
	pdServiceKey := buildIntegrationPagerdutyServiceKey(serviceKey)
	if httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().UpdatePagerDutyIntegrationService(auth, d.Id(), *pdServiceKey); err != nil {
		// Keep the prior state so the next apply retries. Otherwise the new
		// service_key_wo_version is saved and a failed rotation is never resent.
		d.Partial(true)
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating PagerDuty integration service")
	}

	return resourceDatadogIntegrationPagerdutySORead(ctx, d, meta)
}

func resourceDatadogIntegrationPagerdutySODelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationPdMutex.Lock()
	defer integrationPdMutex.Unlock()

	if httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().DeletePagerDutyIntegrationService(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting PagerDuty integration service")
	}

	return nil
}

func resourceDatadogIntegrationPagerdutySOImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	originalId := d.Id()
	if diagErr := resourceDatadogIntegrationPagerdutySORead(ctx, d, meta); diagErr != nil {
		return nil, fmt.Errorf("%s", diagErr[0].Summary)
	}

	// We can assume resource was not found for import when `id` is set to nil in the read step
	if d.Id() == "" {
		return nil, fmt.Errorf("error importing pagerduty service object. Resource with id `%s` does not exist", originalId)
	}

	d.Set("service_key", os.Getenv("SERVICE_KEY"))
	return []*schema.ResourceData{d}, nil
}
