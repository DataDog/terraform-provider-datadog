package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
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

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "Your AWS Account ID without dashes. If your account is a GovCloud or China account, specify the `access_key_id` here.",
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	req := buildDatadogIntegrationAwsTagFilter(d)
	if _, httpresp, err := datadogClientV1.AWSIntegrationApi.CreateAWSTagFilter(authV1, *req); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating aws tag filter")
	}

	d.SetId(fmt.Sprintf("%s:%s", req.GetAccountId(), req.GetNamespace()))
	return resourceDatadogIntegrationAwsTagFilterRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsTagFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	req := buildDatadogIntegrationAwsTagFilter(d)
	if _, httpresp, err := datadogClientV1.AWSIntegrationApi.CreateAWSTagFilter(authV1, *req); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating aws tag filter")
	}

	return resourceDatadogIntegrationAwsTagFilterRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsTagFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, tfNamespace, err := utils.AccountAndNamespaceFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := datadogV1.AWSNamespace(tfNamespace)

	resp, httpresp, err := datadogClientV1.AWSIntegrationApi.ListAWSTagFilters(authV1, accountID)
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, tfNamespace, err := utils.AccountAndNamespaceFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	namespace := datadogV1.AWSNamespace(tfNamespace)
	deleteRequest := datadogV1.AWSTagFilterDeleteRequest{
		AccountId: &accountID,
		Namespace: &namespace,
	}

	if _, httpresp, err := datadogClientV1.AWSIntegrationApi.DeleteAWSTagFilter(authV1, deleteRequest); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting aws tag filter")
	}

	return nil
}
