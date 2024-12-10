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
		DeprecationMessage: "**This resource is deprecated - use the `datadog_integration_aws_account` resource instead**: https://registry.terraform.io/providers/DataDog/datadog/latest/docs/resources/integration_aws_account",
		Description:        "Provides a Datadog AWS tag filter resource. This can be used to create and manage Datadog AWS tag filters.",
		CreateContext:      resourceDatadogIntegrationAwsTagFilterCreate,
		UpdateContext:      resourceDatadogIntegrationAwsTagFilterUpdate,
		ReadContext:        resourceDatadogIntegrationAwsTagFilterRead,
		DeleteContext:      resourceDatadogIntegrationAwsTagFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"account_id": {
					Description: "Your AWS Account ID without dashes.",
					Type:        schema.TypeString,
					Required:    true,
					// TODO: When backend is ready, add validation back.
					// ValidateDiagFunc: validators.ValidateAWSAccountID,
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
	utils.IntegrationAwsMutex.Lock()
	defer utils.IntegrationAwsMutex.Unlock()

	req := buildDatadogIntegrationAwsTagFilter(d)
	if _, httpresp, err := apiInstances.GetAWSIntegrationApiV1().CreateAWSTagFilter(auth, *req); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating aws tag filter")
	}

	d.SetId(fmt.Sprintf("%s:%s", req.GetAccountId(), req.GetNamespace()))
	readDiag := resourceDatadogIntegrationAwsTagFilterRead(ctx, d, meta)
	if !readDiag.HasError() && d.Id() == "" {
		return diag.FromErr(fmt.Errorf("aws integration tag filter resource for account id `%s` with namespace `%s` not found after creation", req.GetAccountId(), req.GetNamespace()))
	}
	return readDiag
}

func resourceDatadogIntegrationAwsTagFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	utils.IntegrationAwsMutex.Lock()
	defer utils.IntegrationAwsMutex.Unlock()

	req := buildDatadogIntegrationAwsTagFilter(d)
	if _, httpresp, err := apiInstances.GetAWSIntegrationApiV1().CreateAWSTagFilter(auth, *req); err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating aws tag filter")
	}

	readDiag := resourceDatadogIntegrationAwsTagFilterRead(ctx, d, meta)
	if !readDiag.HasError() && d.Id() == "" {
		return diag.FromErr(fmt.Errorf("aws integration tag filter resource for account id `%s` with namespace `%s` not found after creation", req.GetAccountId(), req.GetNamespace()))
	}
	return readDiag
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
	utils.IntegrationAwsMutex.Lock()
	defer utils.IntegrationAwsMutex.Unlock()

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
