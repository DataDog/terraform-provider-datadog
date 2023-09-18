package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationAwsTagFilter() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog AWS tag filter resource. This can be used to create and manage Datadog AWS tag filters.",
		CreateContext: resourceDatadogIntegrationAwsTagFilterCreate,
		UpdateContext: resourceDatadogIntegrationAwsTagFilterUpdate,
		ReadContext:   resourceDatadogIntegrationAwsTagFilterRead,
		DeleteContext: resourceDatadogIntegrationAwsTagFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"account_id": {
					Description: "Your AWS Account ID without dashes.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"namespace": {
					Description:      "The namespace associated with the tag filter entry.",
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewAWSNamespaceFromValue),
				},
				"tag_filter_str": {
					Description: "The tag filter string.",
					Type:        schema.TypeString,
					Required:    true,
				},
			}
		},
	}
}

func buildDatadogIntegrationAwsTagFilter(d *schema.ResourceData) *datadogV1.AWSTagFilterCreateRequest {
	filterRequest := datadogV1.NewAWSTagFilterCreateRequestWithDefaults()
	if v, ok := d.GetOk("account_id"); ok {
		filterRequest.SetAccountId(v.(string))
	}
	if v, ok := d.GetOk("namespace"); ok {
		namespace := datadogV1.AWSNamespace(v.(string))
		filterRequest.SetNamespace(namespace)
	}
	if v, ok := d.GetOk("tag_filter_str"); ok {
		filterRequest.SetTagFilterStr(v.(string))
	}

	return filterRequest
}

func resourceDatadogIntegrationAwsTagFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	req := buildDatadogIntegrationAwsTagFilter(d)
	if _, httpresp, err := apiInstances.GetAWSIntegrationApiV1().CreateAWSTagFilter(auth, *req); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating aws tag filter")
	}

	d.SetId(fmt.Sprintf("%s:%s", req.GetAccountId(), req.GetNamespace()))
	return resourceDatadogIntegrationAwsTagFilterRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsTagFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	req := buildDatadogIntegrationAwsTagFilter(d)
	if _, httpresp, err := apiInstances.GetAWSIntegrationApiV1().CreateAWSTagFilter(auth, *req); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating aws tag filter")
	}

	return resourceDatadogIntegrationAwsTagFilterRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsTagFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountID, tfNamespace, err := utils.AccountAndNamespaceFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := datadogV1.AWSNamespace(tfNamespace)

	resp, httpresp, err := apiInstances.GetAWSIntegrationApiV1().ListAWSTagFilters(auth, accountID)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error listing aws tag filter")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	for _, ns := range resp.GetFilters() {
		if ns.GetNamespace() == namespace {
			d.Set("account_id", accountID)
			d.Set("namespace", ns.GetNamespace())
			d.Set("tag_filter_str", ns.GetTagFilterStr())
			return nil
		}
	}

	// Set ID to an empty string if namespace is not found.
	// This allows Terraform to destroy the resource in state.
	d.SetId("")
	return nil
}

func resourceDatadogIntegrationAwsTagFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID, tfNamespace, err := utils.AccountAndNamespaceFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := datadogV1.AWSNamespace(tfNamespace)
	deleteRequest := datadogV1.AWSTagFilterDeleteRequest{
		AccountId: &accountID,
		Namespace: &namespace,
	}

	if _, httpresp, err := apiInstances.GetAWSIntegrationApiV1().DeleteAWSTagFilter(auth, deleteRequest); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting aws tag filter")
	}

	return nil
}
