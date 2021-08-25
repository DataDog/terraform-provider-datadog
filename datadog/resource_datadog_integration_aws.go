package datadog

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var integrationAwsMutex = sync.Mutex{}
var accountAndRoleNameIDRegex = regexp.MustCompile("[\\d]+:.*")

func resourceDatadogIntegrationAws() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.\n\n",
		CreateContext: resourceDatadogIntegrationAwsCreate,
		ReadContext:   resourceDatadogIntegrationAwsRead,
		UpdateContext: resourceDatadogIntegrationAwsUpdate,
		DeleteContext: resourceDatadogIntegrationAwsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatadogIntegrationAwsImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description:   "Your AWS Account ID without dashes.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"access_key_id", "secret_access_key"},
			},
			"role_name": {
				Description:   "Your Datadog role delegation name.",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"access_key_id", "secret_access_key"},
			},
			"filter_tags": {
				Description: "Array of EC2 tags (in the form `key:value`) defines a filter that Datadog uses when collecting metrics from EC2. Wildcards, such as `?` (for single characters) and `*` (for multiple characters) can also be used. Only hosts that match one of the defined tags will be imported into Datadog. The rest will be ignored. Host matching a given tag can also be excluded by adding `!` before the tag. e.x. `env:production,instance-type:c1.*,!region:us-east-1`.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"host_tags": {
				Description: "Array of tags (in the form `key:value`) to add to all hosts and metrics reporting through this integration.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"account_specific_namespace_rules": {
				Description: "Enables or disables metric collection for specific AWS namespaces for this AWS account only. A list of namespaces can be found at the [available namespace rules API endpoint](https://docs.datadoghq.com/api/v1/aws-integration/#list-namespace-rules).",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        schema.TypeBool,
			},
			"excluded_regions": {
				Description: "An array of AWS regions to exclude from metrics collection.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"external_id": {
				Description: "AWS External ID. **NOTE** This provider will not be able to detect changes made to the `external_id` field from outside Terraform.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"access_key_id": {
				Description:   "Your AWS access key ID. Only required if your AWS account is a GovCloud or China account.",
				Type:          schema.TypeString,
				ConflictsWith: []string{"account_id", "role_name"},
				Optional:      true,
			},
			"secret_access_key": {
				Description:   "Your AWS secret access key. Only required if your AWS account is a GovCloud or China account.",
				Type:          schema.TypeString,
				Sensitive:     true,
				ConflictsWith: []string{"account_id", "role_name"},
				Optional:      true,
			},
		},
	}
}

func buildDatadogIntegrationAwsStruct(d *schema.ResourceData) *datadogV1.AWSAccount {
	iaws := datadogV1.NewAWSAccount()

	if v, ok := d.GetOk("account_id"); ok {
		iaws.SetAccountId(v.(string))
	}
	if v, ok := d.GetOk("role_name"); ok {
		iaws.SetRoleName(v.(string))
	}
	if v, ok := d.GetOk("access_key_id"); ok {
		iaws.SetAccessKeyId(v.(string))
	}
	if v, ok := d.GetOk("secret_access_key"); ok {
		iaws.SetSecretAccessKey(v.(string))
	}

	filterTags := make([]string, 0)
	if attr, ok := d.GetOk("filter_tags"); ok {
		for _, s := range attr.([]interface{}) {
			filterTags = append(filterTags, s.(string))
		}
	}
	iaws.SetFilterTags(filterTags)

	hostTags := make([]string, 0)
	if attr, ok := d.GetOk("host_tags"); ok {
		for _, s := range attr.([]interface{}) {
			hostTags = append(hostTags, s.(string))
		}
	}
	iaws.SetHostTags(hostTags)

	accountSpecificNamespaceRules := make(map[string]bool)
	if attr, ok := d.GetOk("account_specific_namespace_rules"); ok {
		// TODO: this is not very defensive, test if we can fail on non bool input
		for k, v := range attr.(map[string]interface{}) {
			accountSpecificNamespaceRules[k] = v.(bool)
		}
	}
	iaws.SetAccountSpecificNamespaceRules(accountSpecificNamespaceRules)

	excludedRegions := make([]string, 0)
	if attr, ok := d.GetOk("excluded_regions"); ok {
		for _, s := range attr.([]interface{}) {
			excludedRegions = append(excludedRegions, s.(string))
		}
	}
	iaws.SetExcludedRegions(excludedRegions)

	return iaws
}

func resourceDatadogIntegrationAwsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	iaws := buildDatadogIntegrationAwsStruct(d)
	response, httpresp, err := datadogClientV1.AWSIntegrationApi.CreateAWSAccount(authV1, *iaws)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating AWS integration")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("access_key_id"); ok {
		d.SetId(v.(string))
	} else {
		accountID := d.Get("account_id").(string)
		roleName := d.Get("role_name").(string)
		d.SetId(fmt.Sprintf("%s:%s", accountID, roleName))
	}

	d.Set("external_id", response.ExternalId)

	return resourceDatadogIntegrationAwsRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	var accountID, roleName, accessKeyID string
	var err error
	if accountAndRoleNameIDRegex.MatchString(d.Id()) {
		accountID, roleName, err = utils.AccountAndRoleFromID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		accessKeyID = d.Id()
	}

	integrations, httpresp, err := datadogClientV1.AWSIntegrationApi.ListAWSAccounts(authV1)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 400 {
			// API returns 400 if integration is not installed
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting AWS integration")
	}
	if err := utils.CheckForUnparsed(integrations); err != nil {
		return diag.FromErr(err)
	}

	for _, integration := range integrations.GetAccounts() {
		if (accountID != "" && integration.GetAccountId() == accountID && integration.GetRoleName() == roleName) ||
			(accessKeyID != "" && integration.GetAccessKeyId() == accessKeyID) {
			d.Set("account_id", integration.GetAccountId())
			d.Set("role_name", integration.GetRoleName())
			d.Set("access_key_id", integration.GetAccessKeyId())
			d.Set("filter_tags", integration.GetFilterTags())
			d.Set("host_tags", integration.GetHostTags())
			d.Set("account_specific_namespace_rules", integration.GetAccountSpecificNamespaceRules())
			d.Set("excluded_regions", integration.GetExcludedRegions())
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceDatadogIntegrationAwsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	iaws := buildDatadogIntegrationAwsStruct(d)

	if !accountAndRoleNameIDRegex.MatchString(d.Id()) {
		_, httpresp, err := datadogClientV1.AWSIntegrationApi.UpdateAWSAccount(authV1, *iaws,
			*datadogV1.NewUpdateAWSAccountOptionalParameters().
				WithAccessKeyId(d.Id()),
		)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpresp, "error updating AWS integration")
		}

		d.SetId(iaws.GetAccessKeyId())
		return resourceDatadogIntegrationAwsRead(ctx, d, meta)

	}

	existingAccountID, existingRoleName, err := utils.AccountAndRoleFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, httpresp, err := datadogClientV1.AWSIntegrationApi.UpdateAWSAccount(authV1, *iaws,
		*datadogV1.NewUpdateAWSAccountOptionalParameters().
			WithAccountId(existingAccountID).
			WithRoleName(existingRoleName),
	)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating AWS integration")
	}
	d.SetId(fmt.Sprintf("%s:%s", iaws.GetAccountId(), iaws.GetRoleName()))
	return resourceDatadogIntegrationAwsRead(ctx, d, meta)
}

func buildDatadogIntegrationAwsDeleteStruct(d *schema.ResourceData) *datadogV1.AWSAccountDeleteRequest {
	awsDeleteRequest := datadogV1.NewAWSAccountDeleteRequest()

	if v, ok := d.GetOk("account_id"); ok {
		awsDeleteRequest.SetAccountId(v.(string))
	}
	if v, ok := d.GetOk("role_name"); ok {
		awsDeleteRequest.SetRoleName(v.(string))
	}
	if v, ok := d.GetOk("access_key_id"); ok {
		awsDeleteRequest.SetAccessKeyId(v.(string))
	}

	return awsDeleteRequest
}

func resourceDatadogIntegrationAwsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	iaws := buildDatadogIntegrationAwsDeleteStruct(d)

	_, httpresp, err := datadogClientV1.AWSIntegrationApi.DeleteAWSAccount(authV1, *iaws)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting AWS integration")
	}

	return nil
}

func resourceDatadogIntegrationAwsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if diagErr := resourceDatadogIntegrationAwsRead(ctx, d, meta); diagErr != nil {
		return nil, fmt.Errorf(diagErr[0].Summary)
	}
	d.Set("external_id", os.Getenv("EXTERNAL_ID"))
	return []*schema.ResourceData{d}, nil
}
