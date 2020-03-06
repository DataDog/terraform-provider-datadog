package datadog

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		Create: resourceDatadogIntegrationAwsCreate,
		Read:   resourceDatadogIntegrationAwsRead,
		Update: resourceDatadogIntegrationAwsUpdate,
		Delete: resourceDatadogIntegrationAwsDelete,
		Exists: resourceDatadogIntegrationAwsExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationAwsImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // waits for update API call support
			},
			"filter_tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true, // waits for update API call support
			},
			"host_tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true, // waits for update API call support
			},
			"account_specific_namespace_rules": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeBool,
				ForceNew: true, // waits for update API call support
			},
			"external_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceDatadogIntegrationAwsExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrations, _, err := client.AWSIntegrationApi.GetAllAWSAccounts(auth).Execute()
	if err != nil {
		return false, err
	}
	accountID, roleName, err := accountAndRoleFromID(d.Id())
	if err != nil {
		return false, err
	}
	for _, integration := range integrations.GetAccounts() {
		if integration.GetAccountId() == accountID && integration.GetRoleName() == roleName {
			return true, nil
		}
	}
	return false, nil
}

func resourceDatadogIntegrationAwsPrepareCreateRequest(d *schema.ResourceData, accountID string, roleName string) datadog.AWSAccount {

	iaws := datadog.AWSAccount{
		AccountId: datadog.PtrString(accountID),
		RoleName:  datadog.PtrString(roleName),
	}

	filterTags := []string{}

	if attr, ok := d.GetOk("filter_tags"); ok {
		for _, s := range attr.([]interface{}) {
			filterTags = append(filterTags, s.(string))
		}
	}

	hostTags := []string{}

	if attr, ok := d.GetOk("host_tags"); ok {
		for _, s := range attr.([]interface{}) {
			hostTags = append(hostTags, s.(string))
		}
	}

	accountSpecificNamespaceRules := make(map[string]bool)

	if attr, ok := d.GetOk("account_specific_namespace_rules"); ok {
		// TODO: this is not very defensive, test if we can fail on non bool input
		for k, v := range attr.(map[string]interface{}) {
			accountSpecificNamespaceRules[k] = v.(bool)
		}
	}
	iaws.FilterTags = &filterTags
	iaws.HostTags = &hostTags
	iaws.AccountSpecificNamespaceRules = &accountSpecificNamespaceRules
	return iaws
}

func resourceDatadogIntegrationAwsCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID := d.Get("account_id").(string)
	roleName := d.Get("role_name").(string)

	iaws := resourceDatadogIntegrationAwsPrepareCreateRequest(d, accountID, roleName)
	response, _, err := client.AWSIntegrationApi.CreateAWSAccount(auth).Body(iaws).Execute()

	if err != nil {
		return translateClientError(err, "error creating a Amazon Web Services integration")
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, roleName))
	d.Set("external_id", response.GetExternalId())

	return resourceDatadogIntegrationAwsRead(d, meta)
}

func resourceDatadogIntegrationAwsRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	accountID, roleName, err := accountAndRoleFromID(d.Id())

	if err != nil {
		return err
	}

	integrations, _, err := client.AWSIntegrationApi.GetAllAWSAccounts(auth).Execute()
	if err != nil {
		return err
	}
	for _, integration := range integrations.GetAccounts() {
		if integration.GetAccountId() == accountID && integration.GetRoleName() == roleName {
			d.Set("account_id", integration.GetAccountId())
			d.Set("role_name", integration.GetRoleName())
			d.Set("filter_tags", integration.FilterTags)
			d.Set("host_tags", integration.HostTags)
			d.Set("account_specific_namespace_rules", integration.AccountSpecificNamespaceRules)
			return nil
		}
	}
	return fmt.Errorf("error getting a Amazon Web Services integration: account_id=%s, role_name=%s", accountID, roleName)
}

func resourceDatadogIntegrationAwsUpdate(d *schema.ResourceData, meta interface{}) error {
	// Unfortunately the PUT operation for updating the AWS configuration is not available at the moment.
	// However this feature is one we have in our backlog. I don't know if it's scheduled for delivery short-term,
	// however I will follow-up after reviewing with product management.
	// ©

	// UpdateIntegrationAWS function:
	// func (client *Client) UpdateIntegrationAWS(awsAccount *IntegrationAWSAccount) (*IntegrationAWSAccountCreateResponse, error) {
	// 	var out IntegrationAWSAccountCreateResponse
	// 	if err := client.doJsonRequest("PUT", "/v1/integration/aws", awsAccount, &out); err != nil {
	// 		return nil, err
	// 	}
	// 	return &out, nil
	// }

	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID, roleName, err := accountAndRoleFromID(d.Id())
	if err != nil {
		return err
	}

	iaws := resourceDatadogIntegrationAwsPrepareCreateRequest(d, accountID, roleName)

	_, _, err = client.AWSIntegrationApi.UpdateAWSAccount(auth).Body(iaws).Execute()
	if err != nil {
		return translateClientError(err, "error updating a Amazon Web Services integration")
	}

	return resourceDatadogIntegrationAwsRead(d, meta)
}

func resourceDatadogIntegrationAwsDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID, roleName, err := accountAndRoleFromID(d.Id())
	if err != nil {
		return err
	}

	if _, _, err := client.AWSIntegrationApi.DeleteAWSAccount(auth).Body(
		datadog.AWSAccount{
			AccountId: datadog.PtrString(accountID),
			RoleName:  datadog.PtrString(roleName),
		},
	).Execute(); err != nil {
		return translateClientError(err, "error deleting a Amazon Web Services integration")
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
