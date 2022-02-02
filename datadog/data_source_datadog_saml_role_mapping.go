package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogSamlRoleMapping() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing SAML Role Mapping for use in other resources.",
		ReadContext: dataSourceDatadogSamlRoleMappingRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The SAML Role Mapping ID.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func dataSourceDatadogSamlRoleMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
