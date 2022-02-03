package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogAuthnMapping() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog SAML Role Mappings resource. This can be used to create and manage Datadog SAML Role Mappings.",
		CreateContext: resourceDatadogAuthnMappingCreate,
		ReadContext:   resourceDatadogAuthnMappingRead,
		UpdateContext: resourceDatadogAuthnMappingUpdate,
		DeleteContext: resourceDatadogAuthnMappingDelete,

		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Identity provider key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"value": {
				Description: "Identity provider value.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"role": {
				Description: "The role to assign for key:value mapping.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceDatadogAuthnMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDatadogAuthnMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDatadogAuthnMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceDatadogAuthnMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
