package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogSamlRoleMappings() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about multiple existing SAML Role Mappings for use in other resources.",
		ReadContext: dataSourceDatadogSamlRoleMappingRead,
		Schema:      map[string]*schema.Schema{},
	}
}

func dataSourceDatadogSamlRoleMappingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
