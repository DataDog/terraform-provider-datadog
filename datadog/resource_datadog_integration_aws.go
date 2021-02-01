package datadog

import (
	"fmt"
	"os"
	"strings"
	"sync"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var integrationAwsMutex = sync.Mutex{}

func accountAndRoleFromID(id string) (string, string, error) {
	result := strings.SplitN(id, ":", 2)
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account ID and Role name from an Amazon Web Services integration id: %s", id)
	}
	return result[0], result[1], nil
}

func resourceDatadogIntegrationAws() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog - Amazon Web Services integration resource. This can be used to create and manage Datadog - Amazon Web Services integration.\n\n",
		Create:      resourceDatadogIntegrationAwsCreate,
		Read:        resourceDatadogIntegrationAwsRead,
		Update:      resourceDatadogIntegrationAwsUpdate,
		Delete:      resourceDatadogIntegrationAwsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationAwsImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "Your AWS Account ID without dashes.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"role_name": {
				Description: "Your Datadog role delegation name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"filter_tags": {
				Description: "Array of EC2 tags (in the form key:value) defines a filter that Datadog use when collecting metrics from EC2. Wildcards, such as `?` (for single characters) and `*` (for multiple characters) can also be used.\n\nOnly hosts that match one of the defined tags will be imported into Datadog. The rest will be ignored. Host matching a given tag can also be excluded by adding `!` before the tag.\n\ne.x. `env:production,instance-type:c1.*,!region:us-east-1`.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"host_tags": {
				Description: "Array of tags (in the form key:value) to add to all hosts and metrics reporting through this integration.",
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
		},
	}
}

func buildDatadogIntegrationAwsStruct(d *schema.ResourceData, accountID string, roleName string) *datadogV1.AWSAccount {
	iaws := datadogV1.NewAWSAccount()
	iaws.SetAccountId(accountID)
	iaws.SetRoleName(roleName)

	filterTags := make([]string, 0)
	if attr, ok := d.GetOk("filter_tags"); ok {
		for _, s := range attr.([]interface{}) {
			filterTags = append(filterTags, s.(string))
		}
		iaws.SetFilterTags(filterTags)
	}

	hostTags := make([]string, 0)
	if attr, ok := d.GetOk("host_tags"); ok {
		for _, s := range attr.([]interface{}) {
			hostTags = append(hostTags, s.(string))
		}
		iaws.SetHostTags(hostTags)
	}

	accountSpecificNamespaceRules := make(map[string]bool)
	if attr, ok := d.GetOk("account_specific_namespace_rules"); ok {
		// TODO: this is not very defensive, test if we can fail on non bool input
		for k, v := range attr.(map[string]interface{}) {
			accountSpecificNamespaceRules[k] = v.(bool)
		}
		iaws.SetAccountSpecificNamespaceRules(accountSpecificNamespaceRules)
	}

	excludedRegions := make([]string, 0)
	if attr, ok := d.GetOk("excluded_regions"); ok {
		for _, s := range attr.([]interface{}) {
			excludedRegions = append(excludedRegions, s.(string))
		}
		iaws.SetExcludedRegions(excludedRegions)
	}

	return iaws
}

func resourceDatadogIntegrationAwsCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID := d.Get("account_id").(string)
	roleName := d.Get("role_name").(string)

	iaws := buildDatadogIntegrationAwsStruct(d, accountID, roleName)
	response, _, err := datadogClientV1.AWSIntegrationApi.CreateAWSAccount(authV1).Body(*iaws).Execute()

	if err != nil {
		return translateClientError(err, "error creating AWS integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, roleName))
	d.Set("external_id", response.ExternalId)

	return resourceDatadogIntegrationAwsRead(d, meta)
}

func resourceDatadogIntegrationAwsRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, roleName, err := accountAndRoleFromID(d.Id())
	if err != nil {
		return err
	}

	integrations, _, err := datadogClientV1.AWSIntegrationApi.ListAWSAccounts(authV1).Execute()
	if err != nil {
		return translateClientError(err, "error getting AWS integration")
	}

	for _, integration := range integrations.GetAccounts() {
		if integration.GetAccountId() == accountID && integration.GetRoleName() == roleName {
			d.Set("account_id", integration.GetAccountId())
			d.Set("role_name", integration.GetRoleName())
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

func resourceDatadogIntegrationAwsUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	existingAccountID, existingRoleName, err := accountAndRoleFromID(d.Id())
	if err != nil {
		return err
	}
	newAccountID := d.Get("account_id").(string)
	newRoleName := d.Get("role_name").(string)

	iaws := buildDatadogIntegrationAwsStruct(d, newAccountID, newRoleName)
	_, _, err = datadogClientV1.AWSIntegrationApi.UpdateAWSAccount(authV1).
		Body(*iaws).AccountId(existingAccountID).RoleName(existingRoleName).Execute()
	if err != nil {
		return translateClientError(err, "error updating AWS integration")
	}
	d.SetId(fmt.Sprintf("%s:%s", iaws.GetAccountId(), iaws.GetRoleName()))
	return resourceDatadogIntegrationAwsRead(d, meta)
}

func resourceDatadogIntegrationAwsDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID, roleName, err := accountAndRoleFromID(d.Id())
	if err != nil {
		return err
	}
	iaws := buildDatadogIntegrationAwsStruct(d, accountID, roleName)

	_, _, err = datadogClientV1.AWSIntegrationApi.DeleteAWSAccount(authV1).Body(*iaws).Execute()
	if err != nil {
		return translateClientError(err, "error deleting AWS integration")
	}

	return nil
}

func resourceDatadogIntegrationAwsImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationAwsRead(d, meta); err != nil {
		return nil, err
	}
	d.Set("external_id", os.Getenv("EXTERNAL_ID"))
	return []*schema.ResourceData{d}, nil
}
