package datadog

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func accountAndRoleFromID(id string) (string, string) {
	result := strings.SplitN(id, "_", 2)
	return result[0], result[1]
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
	client := meta.(*datadog.Client)

	integrations, err := client.GetIntegrationAWS()
	if err != nil {
		return false, err
	}
	accountID, roleName := accountAndRoleFromID(d.Id())
	for _, integration := range *integrations {
		if integration.GetAccountID() == accountID && integration.GetRoleName() == roleName {
			return true, nil
		}
	}
	return false, nil
}

func resourceDatadogIntegrationAwsPrepareCreateRequest(d *schema.ResourceData, accountID string, roleName string) datadog.IntegrationAWSAccount {

	iaws := datadog.IntegrationAWSAccount{
		AccountID: datadog.String(accountID),
		RoleName:  datadog.String(roleName),
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
	iaws.FilterTags = filterTags
	iaws.HostTags = hostTags
	iaws.AccountSpecificNamespaceRules = accountSpecificNamespaceRules
	return iaws
}

func resourceDatadogIntegrationAwsCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] debugging logger")
	client := meta.(*datadog.Client)

	accountID := d.Get("account_id").(string)
	roleName := d.Get("role_name").(string)

	iaws := resourceDatadogIntegrationAwsPrepareCreateRequest(d, accountID, roleName)
	responce, err := client.CreateIntegrationAWS(&iaws)

	if err != nil {
		return fmt.Errorf("error creating a Amazon Web Services integration: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%s_%s", accountID, roleName))
	d.Set("external_id", responce.ExternalID)

	return resourceDatadogIntegrationAwsRead(d, meta)
}

func resourceDatadogIntegrationAwsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	accountID, roleName := accountAndRoleFromID(d.Id())

	integrations, err := client.GetIntegrationAWS()
	if err != nil {
		return err
	}
	for _, integration := range *integrations {
		if integration.GetAccountID() == accountID && integration.GetRoleName() == roleName {
			d.Set("account_id", integration.GetAccountID())
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

	client := meta.(*datadog.Client)

	accountID, roleName := accountAndRoleFromID(d.Id())

	iaws := resourceDatadogIntegrationAwsPrepareCreateRequest(d, accountID, roleName)

	_, err := client.CreateIntegrationAWS(&iaws)
	if err != nil {
		return fmt.Errorf("error updating a Amazon Web Services integration: %s", err.Error())
	}

	return resourceDatadogIntegrationAwsRead(d, meta)
}

func resourceDatadogIntegrationAwsDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)
	accountID, roleName := accountAndRoleFromID(d.Id())

	if err := client.DeleteIntegrationAWS(
		&datadog.IntegrationAWSAccountDeleteRequest{
			AccountID: datadog.String(accountID),
			RoleName:  datadog.String(roleName),
		},
	); err != nil {
		return fmt.Errorf("error deleting a Amazon Web Services integration: %s", err.Error())
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
